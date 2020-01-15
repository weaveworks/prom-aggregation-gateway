package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/weaveworks/prom-aggregation-gateway/aggate"
)

func main() {
	listen := flag.String("listen", ":80", "Address and port to listen on.")
	cors := flag.String("cors", "*", "The 'Access-Control-Allow-Origin' value to be returned.")
	apiEndpoint := flag.String("api-endpoint", "/api/ui/metrics", "Endpoint for retrieving push metrics from clients")
	labelQueryParam := flag.String("label-query-param", "", "Append labels to metrics from query parameters <label-query-param>=<label-key>:<label-value>")
	flag.Parse()

	a := aggate.NewAggate()
	http.HandleFunc("/metrics", a.Handler)
	http.HandleFunc(*apiEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", *cors)
		if err := a.ParseAndMerge(r.Body, r.URL.Query(), *labelQueryParam); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	})
	log.Fatal(http.ListenAndServe(*listen, nil))
}
