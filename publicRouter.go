package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	mGin "github.com/slok/go-http-metrics/middleware/gin"
)

type apiRouterConfig struct {
	corsDomain string
	accounts   gin.Accounts
}

func createHandlers(endpointName string,
	metricMiddleware middleware.Middleware,
	neededHandlers []gin.HandlerFunc,
	handlers ...gin.HandlerFunc) []gin.HandlerFunc {
	h := []gin.HandlerFunc{
		mGin.Handler(endpointName, metricMiddleware),
	}
	h = append(h, neededHandlers...)
	return append(h, handlers...)
}

func setupAPIRouter(cfg apiRouterConfig, agg *aggregate, promConfig metrics.Config) *gin.Engine {
	corsConfig := cors.Config{}
	if cfg.corsDomain != "*" {
		corsConfig.AllowOrigins = []string{cfg.corsDomain}
	} else {
		corsConfig.AllowAllOrigins = true
	}

	metricsMiddleware := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(promConfig),
	})

	r := gin.New()
	r.RedirectTrailingSlash = false

	// add metric middleware for NoRoute handler
	r.NoRoute(mGin.Handler("noRoute", metricsMiddleware))

	neededHandlers := []gin.HandlerFunc{}

	if len(cfg.accounts) > 0 {
		neededHandlers = append(neededHandlers, gin.BasicAuth(cfg.accounts))
	}

	r.GET("/metrics",
		mGin.Handler("getMetrics", metricsMiddleware),
		cors.New(corsConfig),
		agg.handleRender,
	)

	insertMethods := []func(string, ...gin.HandlerFunc) gin.IRoutes{r.POST, r.PUT}
	insertPaths := []string{"/metrics", "/metrics/*labels"}
	for _, method := range insertMethods {
		for _, path := range insertPaths {
			method(path, createHandlers("postMetrics", metricsMiddleware, neededHandlers, agg.handleInsert)...)
		}
	}

	return r
}
