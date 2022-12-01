package main

import (
	"flag"
	"io"
	"log"
	"net/http"
)

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}

func main() {
	listen := flag.String("listen", ":80", "Address and port to listen on.")
	cors := flag.String("cors", "*", "The 'Access-Control-Allow-Origin' value to be returned.")
	pushPath := flag.String("push-path", "/metrics/", "HTTP path to accept pushed metrics.")
	flag.Parse()

	a := newAggregate()
	http.HandleFunc("/metrics", a.handler)
	http.HandleFunc("/-/healthy", handleHealthCheck)
	http.HandleFunc("/-/ready", handleHealthCheck)
	http.HandleFunc(*pushPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", *cors)
		if err := a.parseAndMerge(r.Body); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	})
	log.Fatal(http.ListenAndServe(*listen, nil))
}
