package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sspier/cloudpulse/internal/store"
)

func main() {
	resultsTable := os.Getenv("TABLE_NAME_RESULTS")
	targetsTable := os.Getenv("TABLE_NAME_TARGETS")

	// CLOUD MODE: if the results table and the target table are set, we assume cloud mode and use DynamoDB
	if resultsTable != "" && targetsTable != "" {
		log.Println("initializing dynamodb store")
		awsRegion := os.Getenv("AWS_REGION")
		if awsRegion == "" {
			awsRegion = "us-east-1"
		}

		// initialize the store with DynamoDB using the region, targets table, and results table as well as the top level context
		db, err := store.NewDynamoDBStore(context.Background(), awsRegion, targetsTable, resultsTable)
		if err != nil {
			log.Fatalf("failed to init dynamodb store: %v", err)
		}
		targetStore = db
		// LOCAL MODE: if the results table and the target table are not set, we assume local mode and use in-memory store
	} else {
		log.Println("initializing in-memory store") // targetStore is already init to NewInMemoryStore by default in handlers.go
	}

	// create the http router that wires paths to handler functions
	httpRouter := http.NewServeMux()

	// register API endpoints for health checks, targets, and results
	httpRouter.HandleFunc("/health", healthHandler)
	httpRouter.HandleFunc("/targets", targetsHandler)
	httpRouter.HandleFunc("/results", resultsHandler)
	// returns full probe history for a specific target
	httpRouter.HandleFunc("/results/{id}", resultsForTargetHandler)

	// background scheduler for recurring uptime checks
	// we only run this if explicit configuration says so, OR if we are in local mode.
	// in cloud mode, the runner service handles this
	// for simplicity:
	// - if using in-memory, run scheduler
	// - if using DynamoDB, assume runner handles it
	if resultsTable == "" {
		go func() {
			// NewTicker creates a channel that sends a timestamp every X interval
			// here: every 30 seconds, we wake up and run probes
			//
			// a timeTicker does NOT block; it simply emits events on timeTicker.C
			timeTicker := time.NewTicker(30 * time.Second)
			defer timeTicker.Stop()

			// this loop runs forever while the process is alive
			// each iteration executes when timeTicker.C receives a "tick"
			for range timeTicker.C {

				// use a fresh context per cycle
				// context.Context or 'ctx' in Go is used for control flow, specifically around:
				// - timeouts/deadlines: "stop this operation if it takes longer than 5 seconds"
				// - cancellation: "the user hit Ctrl+C or closed their browser, so stop processing this request immediately to save resources"
				// - request scoped values: carrying trace IDs or authentication data throughout the request chain
				ctx := context.Background()

				// pull the latest set of targets from the store
				targets, err := targetStore.ListTargets(ctx)
				if err != nil {
					log.Printf("scheduler: failed to list targets: %v\n", err)
					continue
				}

				// if no targets exist yet, nothing to do this tick
				if len(targets) == 0 {
					continue
				}

				log.Printf("scheduler: running checks for %d targets\n", len(targets))

				// kick off one probe per target
				// each check runs asynchronously in its own goroutine
				// so targets don't block each other
				//
				// the scheduler loop does not wait for check completion
				// in the DynamoDB-backed version, results will be written to the table
				for _, target := range targets {
					go runCheck(target)
				}
			}
		}()
	} else {
		// when a resultsTable IS configured, we assume the environment is AWS/cloud
		// in those environments we don't run a background loop inside the API
		// instead the runner executing via EventBridge performs checks on a schedule
		log.Println("scheduler disabled (assuming external runner)")
	}

	// configure the http httpServer with sensible timeouts
	// read header timeout helps protect against slowloris-style clients
	// idle timeout prevents connections from lingering
	httpServer := &http.Server{
		Addr:              ":8080",
		Handler:           httpRouter,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Println("CloudPulse API listening on :8080")

	// start serving and crash loudly if something unexpected happens
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
