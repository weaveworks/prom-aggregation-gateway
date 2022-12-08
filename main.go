package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var version = "0.0.0"

func handleHealthCheck(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, `{"alive": true}`)
}

func main() {
	var (
		showVersion = false
	)
	listen := flag.String("listen", ":80", "Address and port to listen on.")
	metricsListen := flag.String("metricsListen", ":8888", "Address and port serve the metrics endpoint on")
	cors := flag.String("cors", "*", "The 'Access-Control-Allow-Origin' value to be returned.")
	flag.BoolVar(&showVersion, "version", false, "Display the version")
	flag.Parse()

	if showVersion {
		fmt.Printf("%s\n", version)
		return
	}

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGTERM, syscall.SIGINT)

	rc := &RouterConfig{
		AllowedCORS: cors,
	}
	r := setupRouter(rc)

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
