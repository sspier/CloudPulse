package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	httpRouter := http.NewServeMux()

	httpRouter.HandleFunc("/health", healthHandler)
	httpRouter.HandleFunc("/targets", targetsHandler)
	httpRouter.HandleFunc("/results", resultsHandler)
	httpRouter.HandleFunc("/results/{id}", resultsForTargetHandler)

	go func() {
		ticker := time.NewTicker(30 * time.Second) // adjust interval as needed
		defer ticker.Stop()

		for range ticker.C {
			targets := targetStore.ListTargets()
			if len(targets) == 0 {
				continue
			}

			log.Printf("scheduler: running checks for %d targets\n", len(targets))

			for _, t := range targets {
				// Run each check in its own goroutine so slow targets don't block others.
				go runCheck(t)
			}
		}
	}()

	server := &http.Server{
		Addr:              ":8080",
		Handler:           httpRouter,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Println("CloudPulse API listening on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
