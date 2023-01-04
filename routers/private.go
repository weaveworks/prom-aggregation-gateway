package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func setupLifecycleRouter(promRegistry *prometheus.Registry) *gin.Engine {
	r := gin.New()

	metricsHandler := promhttp.InstrumentMetricHandler(
		promRegistry,
		promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{}),
	)

	r.GET("/healthy", handleHealthCheck)
	r.GET("/ready", handleHealthCheck)
	r.GET("/metrics", convertHandler(metricsHandler))

	return r
}

func convertHandler(h http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

type HealthResponse struct {
	IsAlive bool `json:"alive"`
}

func handleHealthCheck(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, HealthResponse{IsAlive: true})
}
