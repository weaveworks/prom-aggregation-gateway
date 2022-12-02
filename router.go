package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	mGin "github.com/slok/go-http-metrics/middleware/gin"
)

func strPtr(s string) *string {
	return &s
}

func setupRouter(cors *string) *gin.Engine {
	r := gin.New()

	metricsMiddleware := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})

	r.GET("/healthy", handleHealthCheck)
	r.GET("/ready", handleHealthCheck)
	r.GET("/metrics", mGin.Handler("metrics", metricsMiddleware), aggregateHandler)
	r.POST("/metrics/job/:job", mGin.Handler("/metrics/job/", metricsMiddleware), func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", *cors)
		// TODO: job work just place holder for now
		// job := c.Param("job")
		if err := Aggregate.parseAndMerge(c.Request.Body); err != nil {
			log.Println(err)
			http.Error(c.Writer, err.Error(), http.StatusBadRequest)
			return
		}
	})

	return r
}
