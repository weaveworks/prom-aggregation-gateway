package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	mGin "github.com/slok/go-http-metrics/middleware/gin"
)

func setupAPIRouter(corsDomain string, agg *aggregate, promConfig metrics.Config) *gin.Engine {
	corsConfig := cors.Config{}
	if corsDomain != "*" {
		corsConfig.AllowOrigins = []string{corsDomain}
	} else {
		corsConfig.AllowAllOrigins = true
	}

	metricsMiddleware := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(promConfig),
	})

	r := gin.New()

	r.GET("/metrics",
		mGin.Handler("metrics", metricsMiddleware),
		cors.New(corsConfig),
		agg.handleRender)
	r.POST("/metrics/job/:job",
		mGin.Handler("/metrics/job", metricsMiddleware),
		agg.handleInsert)

	return r
}
