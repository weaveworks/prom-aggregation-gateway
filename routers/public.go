package routers

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	promMetrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	mGin "github.com/slok/go-http-metrics/middleware/gin"
	"github.com/zapier/prom-aggregation-gateway/metrics"
)

type ApiRouterConfig struct {
	CorsDomain   string
	Accounts     []string
	authAccounts gin.Accounts
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

func setupAPIRouter(cfg ApiRouterConfig, agg *metrics.Aggregate, promConfig promMetrics.Config) *gin.Engine {
	corsConfig := cors.Config{}
	if cfg.CorsDomain != "*" {
		corsConfig.AllowOrigins = []string{cfg.CorsDomain}
	} else {
		corsConfig.AllowAllOrigins = true
	}
	cfg.authAccounts = processAuthConfig(cfg.Accounts)

	metricsMiddleware := middleware.New(middleware.Config{
		Recorder: promMetrics.NewRecorder(promConfig),
	})

	r := gin.New()
	r.RedirectTrailingSlash = false

	// add metric middleware for NoRoute handler
	r.NoRoute(mGin.Handler("noRoute", metricsMiddleware))

	neededHandlers := []gin.HandlerFunc{}

	if len(cfg.Accounts) > 0 {
		neededHandlers = append(neededHandlers, gin.BasicAuth(cfg.authAccounts))
	}

	r.GET("/metrics",
		mGin.Handler("getMetrics", metricsMiddleware),
		cors.New(corsConfig),
		agg.HandleRender,
	)

	insertMethods := []func(string, ...gin.HandlerFunc) gin.IRoutes{r.POST, r.PUT}
	insertPaths := []string{"/metrics", "/metrics/*labels"}
	for _, method := range insertMethods {
		for _, path := range insertPaths {
			method(path, createHandlers("postMetrics", metricsMiddleware, neededHandlers, agg.HandleInsert)...)
		}
	}

	return r
}
