package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// entry point for the CloudPulse API server
func main() {
	// hello()
	// HTTP request router
	httpRouter := http.NewServeMux()

	// health endpoint
	httpRouter.HandleFunc("/health", healthHandler)

	// Prometheus metrics
	httpRouter.Handle("/metrics", promhttp.Handler())

	// determine port API should listen on
	listenAddress := ":" + getEnv("PORT", "8080")

	// configure HTTP server with timeouts and logging
	apiServer := &http.Server{
		Addr:              listenAddress,
		Handler:           logRequests(httpRouter), // request-logging
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// start server in a background goroutine
	go func() {
		log.Printf("CloudPulse API listening on %s", listenAddress)
		if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server encountered an error: %v", err)
		}
	}()

	// ---- shutdown handling ----

	// channel to capture OS interrupts
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)

	// block until shutdown signal is received
	<-shutdownSignal
	log.Println("Shutdown signal received, stopping CloudPulse API...")

	// allow in-flight requests up to 5 seconds to finish before forcing shutdown
	shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := apiServer.Shutdown(shutdownContext); err != nil {
		log.Printf("Graceful shutdown failed: %v", err)
	}

	log.Println("CloudPulse API stopped")
}

func healthHandler(responseWriter http.ResponseWriter, request *http.Request) {
	responseWriter.WriteHeader(http.StatusOK)
	fmt.Fprint(responseWriter, "ok")
}

// returns value of environment variable if present, otherwise returns provided default value
func getEnv(envKey, defaultValue string) string {
	if value := os.Getenv(envKey); value != "" {
		return value
	}
	return defaultValue
}

// logs each request method, path, and duration
func logRequests(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		startTime := time.Now()

		nextHandler.ServeHTTP(responseWriter, request)

		duration := time.Since(startTime).Truncate(time.Microsecond)
		log.Printf("%s %s %s", request.Method, request.URL.Path, duration)
	})
}
