package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

var (
	version = "0.0.0"
)

func main() {
	var apiListen, lifecycleListen, corsDomain string
	var showVersion bool

	flag.StringVar(&apiListen, "apiListen", ":80", "Listen for API requests on this host/port.")
	flag.StringVar(&lifecycleListen, "lifecycleListen", ":8888", "Listen for lifecycle requests (health, metrics) on this host/port")
	flag.StringVar(&corsDomain, "cors", "*", "The 'Access-Control-Allow-Origin' value to be returned.")
	flag.BoolVar(&showVersion, "version", false, "Display the version")
	flag.Parse()

	if showVersion {
		handleVersionCommand()
		return
	}

	runServers(corsDomain, apiListen, lifecycleListen)
}

func runServers(corsDomain string, apiListen string, lifecycleListen string) {
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGTERM, syscall.SIGINT)

	agg := newAggregate()

	apiRouter := setupAPIRouter(corsDomain, agg)
	go runServer("api", apiRouter, apiListen)

	lifecycleRouter := setupLifecycleRouter()
	go runServer("lifecycle", lifecycleRouter, lifecycleListen)

	// Block until an interrupt or term signal is sent
	<-sigChannel
}

func runServer(label string, r *gin.Engine, listen string) {
	log.Printf("%s server listening at %s", label, listen)
	if err := r.Run(listen); err != nil {
		log.Panicf("error while serving %s: %v", label, err)
	}
}

func handleVersionCommand() {
	fmt.Printf("%s\n", version)
}
