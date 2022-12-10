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
	apiListen := flag.String("apiListen", ":80", "Listen for API requests on this host/port.")
	lifecycleListen := flag.String("lifecycleListen", ":8888", "Listen for lifecycle requests (health, metrics) on this host/port")
	corsDomain := flag.String("cors", "*", "The 'Access-Control-Allow-Origin' value to be returned.")
	showVersion := flag.Bool("version", false, "Display the version")
	flag.Parse()

	if showVersion != nil && *showVersion {
		handleVersionCommand()
		return
	}

	runServers(corsDomain, apiListen, lifecycleListen)
}

func runServers(corsDomain *string, apiListen *string, lifecycleListen *string) {
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGTERM, syscall.SIGINT)

	runServer := func(label string, r *gin.Engine, listen string) {
		log.Printf("%s server listening at %s", label, listen)
		if err := r.Run(listen); err != nil {
			log.Panicf("error while serving %s: %v", label, err)
		}
	}

	agg := newAggregate()

	if apiListen != nil {
		apiRouter := setupAPIRouter(corsDomain, agg)
		go runServer("api", apiRouter, *apiListen)
	}

	if lifecycleListen != nil {
		lifecycleRouter := setupLifecycleRouter()
		go runServer("lifecycle", lifecycleRouter, *lifecycleListen)
	}

	// Block until an interrupt or term signal is sent
	<-sigChannel
}

func handleVersionCommand() {
	fmt.Printf("%s\n", version)
}
