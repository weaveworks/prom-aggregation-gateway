package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zapier/prom-aggregation-gateway/config"
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
	Name      string `json:"name"`
	IsAlive   bool   `json:"alive"`
	Version   string `json:"version"`
	CommitSHA string `json:"commitSHA"`
}

func handleHealthCheck(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, HealthResponse{
		Name:      config.Name,
		Version:   config.Version,
		CommitSHA: config.CommitSHA,
		IsAlive:   true,
	})
}
