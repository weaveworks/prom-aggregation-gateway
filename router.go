package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	mGin "github.com/slok/go-http-metrics/middleware/gin"
)

type RouterConfig struct {
	MetricsMiddleware *middleware.Middleware
	AllowedCORS       *string
	Aggregate         *aggregate
}

func (rc *RouterConfig) setupDefault() {
	if rc.MetricsMiddleware == nil {
		m := newMetricMiddleware(&metrics.Config{})
		rc.MetricsMiddleware = &m
	}

	if rc.AllowedCORS == nil {
		rc.AllowedCORS = strPtr("*")
	}

	if rc.Aggregate == nil {
		rc.Aggregate = newAggregate()
	}
}

func DefaultRouterConfig() RouterConfig {
	rc := RouterConfig{}
	rc.setupDefault()
	return rc
}

func newMetricMiddleware(cfg *metrics.Config) middleware.Middleware {
	if cfg == nil {
		cfg = &metrics.Config{}
	}
	return middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(*cfg),
	})
}

func setupRouter(cfg *RouterConfig) *gin.Engine {

	r := gin.New()

	cfg.setupDefault()

	r.GET("/healthy", handleHealthCheck)
	r.GET("/ready", handleHealthCheck)
	r.GET("/metrics", mGin.Handler("metrics", *cfg.MetricsMiddleware), cfg.Aggregate.handler)
	r.POST("/metrics/job/:job", mGin.Handler("/metrics/job", *cfg.MetricsMiddleware), func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", *cfg.AllowedCORS)
		job := c.Param("job")
		// TODO: add logic to verify correct format of job label
		if job == "" {
			err := fmt.Errorf("must send in a valid job name, sent: %s", job)
			log.Println(err)
			http.Error(c.Writer, err.Error(), http.StatusBadRequest)
			return
		}
		if err := cfg.Aggregate.parseAndMerge(c.Request.Body, job); err != nil {
			log.Println(err)
			http.Error(c.Writer, err.Error(), http.StatusBadRequest)
			return
		}
	})

	return r
}
