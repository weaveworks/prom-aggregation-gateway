package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	mGin "github.com/slok/go-http-metrics/middleware/gin"
)

func handleHealthCheck(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, `{"alive": true}`)
}

func main() {
	listen := flag.String("listen", ":80", "Address and port to listen on.")
	metricsListen := flag.String("metricsListen", ":8888", "Address and port serve the metrics endpoint on")
	cors := flag.String("cors", "*", "The 'Access-Control-Allow-Origin' value to be returned.")
	flag.Parse()

	metricsMiddleware := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGTERM, syscall.SIGINT)

	a := newAggregate()
	r := gin.Default()

	r.GET("/healthy", handleHealthCheck)
	r.GET("/ready", handleHealthCheck)
	r.GET("/metrics", mGin.Handler("metrics", metricsMiddleware), a.handler)
	r.POST("/metrics/:job", mGin.Handler("/metrics/:job", metricsMiddleware), func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", *cors)
		// TODO: job work just place holder for now
		// job := c.Param("job")
		if err := a.parseAndMerge(c.Request.Body); err != nil {
			log.Println(err)
			http.Error(c.Writer, err.Error(), http.StatusBadRequest)
			return
		}
	})

	// Serve endpoint
	go func() {
		log.Printf("server listening at %s", *listen)
		if err := r.Run(*listen); err != nil {
			log.Panicf("error while serving: %s", err)
		}
	}()

	// Serve metric endpoint
	go func() {
		log.Printf("metrics listening at %s", *metricsListen)
		if err := http.ListenAndServe(*metricsListen, promhttp.Handler()); err != nil {
			log.Panicf("error while serving metrics: %s", err)
		}
	}()

	// Block until a interrupt or term signal is sent
	<-sigChannel
}
