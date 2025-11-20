package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	// create the http router that wires paths to handler functions
	httpRouter := http.NewServeMux()

	// register API endpoints for health checks, targets, and results
	httpRouter.HandleFunc("/health", healthHandler)
	httpRouter.HandleFunc("/targets", targetsHandler)
	httpRouter.HandleFunc("/results", resultsHandler)
	// returns full probe history for a specific target
	httpRouter.HandleFunc("/results/{id}", resultsForTargetHandler)

	// background scheduler for recurring uptime checks
	// this runs forever in its own goroutine, waking up on a ticker interval
	// on each tick, it grabs all known targets and runs a probe for each
	// probes run in separate goroutines so slow targets don’t block the rest
	go func() {
		ticker := time.NewTicker(30 * time.Second) // check interval
		defer ticker.Stop()

		for range ticker.C {
			targets := targetStore.ListTargets()
			if len(targets) == 0 {
				// no targets to probe yet
				continue
			}

			log.Printf("scheduler: running checks for %d targets\n", len(targets))

			for _, t := range targets {
				go runCheck(t) // run each check asynchronously
			}
		}
	}()

	// configure the http server with sensible timeouts
	// read header timeout helps protect against slowloris-style clients
	// idle timeout prevents connections from lingering unnecessarily
	server := &http.Server{
		Addr:              ":8080",
		Handler:           httpRouter,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Println("CloudPulse API listening on :8080")

	// start serving and crash loudly if something unexpected happens
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
