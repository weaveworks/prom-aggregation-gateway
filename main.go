package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
)

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}

func main() {
	listen := flag.String("listen", ":80", "Address and port to listen on.")
	metricsListen := flag.String("metricsListen", ":8888", "Address and port serve the metrics endpoint on")
	cors := flag.String("cors", "*", "The 'Access-Control-Allow-Origin' value to be returned.")
	pushPath := flag.String("push-path", "/metrics", "HTTP path to accept pushed metrics.")
	flag.Parse()

	metricsMiddleware := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGTERM, syscall.SIGINT)

	a := newAggregate()
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", a.handler)
	mux.HandleFunc("/healthy", handleHealthCheck)
	mux.HandleFunc("/ready", handleHealthCheck)
	mux.HandleFunc(*pushPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", *cors)
		if err := a.parseAndMerge(r.Body); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	})
	h := std.Handler("", metricsMiddleware, mux)

	// Serve endpoint
	go func() {
		log.Printf("server listening at %s", *listen)
		if err := http.ListenAndServe(*listen, h); err != nil {
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
