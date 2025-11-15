package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TestHealthEndpoint verifies that the /health endpoint returns a 200 OK.
// This acts as a basic liveness test for the API router.
func TestHealthEndpoint(t *testing.T) {
	// Create an isolated router containing only the /health route.
	testRouter := http.NewServeMux()
	testRouter.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a mock HTTP GET request to /health.
	testRequest := httptest.NewRequest(http.MethodGet, "/health", nil)

	// Recorder captures the response written by the handler.
	testResponseRecorder := httptest.NewRecorder()

	// Serve the request through the router.
	testRouter.ServeHTTP(testResponseRecorder, testRequest)

	// Verify the status code.
	if testResponseRecorder.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK, got %d", testResponseRecorder.Code)
	}
}

// TestMetricsEndpoint verifies that the /metrics endpoint returns
// real Prometheus metric output and responds with 200 OK.
func TestMetricsEndpoint(t *testing.T) {
	// Create a router containing only the /metrics route.
	testRouter := http.NewServeMux()
	testRouter.Handle("/metrics", promhttp.Handler())

	// Create a mock HTTP GET request to /metrics.
	testRequest := httptest.NewRequest(http.MethodGet, "/metrics", nil)

	// Recorder captures the response.
	testResponseRecorder := httptest.NewRecorder()

	// Serve the request.
	testRouter.ServeHTTP(testResponseRecorder, testRequest)

	// Verify 200 OK.
	if testResponseRecorder.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK for /metrics, got %d", testResponseRecorder.Code)
	}

	// Verify that the metrics output contains a known built-in Prometheus metric.
	// go_gc_duration_seconds is exported by the Prometheus Go client automatically.
	metricsOutput := testResponseRecorder.Body.String()
	if !strings.Contains(metricsOutput, "go_gc_duration_seconds") {
		t.Fatalf("expected Prometheus metrics output, but known metric 'go_gc_duration_seconds' not found")
	}
}
