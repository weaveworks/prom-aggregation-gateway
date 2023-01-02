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
	r.RedirectTrailingSlash = false

	r.GET("/metrics",
		mGin.Handler("getMetrics", metricsMiddleware),
		cors.New(corsConfig),
		agg.handleRender)

	insertHandler := mGin.Handler("postMetrics", metricsMiddleware)
	insertMethods := []func(string, ...gin.HandlerFunc) gin.IRoutes{r.POST, r.PUT}
	insertPaths := []string{"/metrics", "/metrics/*labels"}
	for _, method := range insertMethods {
		for _, path := range insertPaths {
			method(path, insertHandler, agg.handleInsert)
		}
	}

	return r
}
