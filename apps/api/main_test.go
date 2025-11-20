package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// testHealthEndpoint checks that /health responds with 200
// this acts as a simple liveness check for the router setup
func TestHealthEndpoint(t *testing.T) {

	// create a small router that exposes only /health for this test
	testRouter := http.NewServeMux()
	testRouter.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// build a fake GET request to /health
	testRequest := httptest.NewRequest(http.MethodGet, "/health", nil)

	// httptest recorder captures whatever the handler writes
	testResponseRecorder := httptest.NewRecorder()

	// execute the request through the router
	testRouter.ServeHTTP(testResponseRecorder, testRequest)

	// expect 200 OK for a healthy endpoint
	if testResponseRecorder.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK, got %d", testResponseRecorder.Code)
	}
}

// testMetricsEndpoint verifies that /metrics returns valid Prometheus output
// promhttp exports the Go runtime metrics automatically, so we can assert content
func TestMetricsEndpoint(t *testing.T) {

	// create a router exposing only /metrics
	testRouter := http.NewServeMux()
	testRouter.Handle("/metrics", promhttp.Handler())

	// build a fake GET request to /metrics
	testRequest := httptest.NewRequest(http.MethodGet, "/metrics", nil)

	// record the response written by the handler
	testResponseRecorder := httptest.NewRecorder()

	// run the request
	testRouter.ServeHTTP(testResponseRecorder, testRequest)

	// expect 200 OK from Prometheus handler
	if testResponseRecorder.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK for /metrics, got %d", testResponseRecorder.Code)
	}

	// verify that at least one known built-in metric exists
	// go_gc_duration_seconds is exported by default by the Prometheus client
	metricsOutput := testResponseRecorder.Body.String()
	if !strings.Contains(metricsOutput, "go_gc_duration_seconds") {
		t.Fatalf("expected Prometheus metrics output, metric 'go_gc_duration_seconds' not found")
	}
}

// TestTargetsPOST creates a new target via POST /targets and verifies the response
func TestTargetsPOST(t *testing.T) {

	// reset global store so this test starts clean
	targetStore = NewInMemoryStore()

	router := http.NewServeMux()
	router.HandleFunc("/targets", targetsHandler)

	body := `{"name": "Example", "url": "https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/targets", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected HTTP 201 Created, got %d", rr.Code)
	}

	var target Target
	if err := json.NewDecoder(rr.Body).Decode(&target); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if target.Name != "Example" {
		t.Fatalf("expected target name 'Example', got %q", target.Name)
	}
	if target.URL != "https://example.com" {
		t.Fatalf("expected target url 'https://example.com', got %q", target.URL)
	}
	if target.ID == "" {
		t.Fatalf("expected non-empty target id")
	}
}

// TestTargetsGET returns all targets via GET /targets
func TestTargetsGET(t *testing.T) {

	// reset store and seed a couple of targets
	targetStore = NewInMemoryStore()
	targetStore.AddTarget("Example A", "https://a.example.com")
	targetStore.AddTarget("Example B", "https://b.example.com")

	router := http.NewServeMux()
	router.HandleFunc("/targets", targetsHandler)

	req := httptest.NewRequest(http.MethodGet, "/targets", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK, got %d", rr.Code)
	}

	var targets []Target
	if err := json.NewDecoder(rr.Body).Decode(&targets); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
}

// TestResultsLatest verifies GET /results returns the most recent result per target
func TestResultsLatest(t *testing.T) {

	// reset store and seed a target + multiple results
	targetStore = NewInMemoryStore()
	tgt := targetStore.AddTarget("Example", "https://example.com")

	// older result
	targetStore.AddResult(Result{
		TargetID:   tgt.ID,
		Status:     "down",
		HTTPStatus: 500,
	})

	// latest result
	targetStore.AddResult(Result{
		TargetID:   tgt.ID,
		Status:     "up",
		HTTPStatus: 200,
	})

	router := http.NewServeMux()
	router.HandleFunc("/results", resultsHandler)

	req := httptest.NewRequest(http.MethodGet, "/results", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK, got %d", rr.Code)
	}

	var results []Result
	if err := json.NewDecoder(rr.Body).Decode(&results); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 latest result, got %d", len(results))
	}

	if results[0].Status != "up" || results[0].HTTPStatus != 200 {
		t.Fatalf("expected latest result to be up/200, got status=%q httpStatus=%d", results[0].Status, results[0].HTTPStatus)
	}
}

// TestResultsHistory verifies GET /results/{id} returns full history for a target
func TestResultsHistory(t *testing.T) {

	// reset store and seed a target + multiple results
	targetStore = NewInMemoryStore()
	tgt := targetStore.AddTarget("Example", "https://example.com")

	targetStore.AddResult(Result{
		TargetID:   tgt.ID,
		Status:     "down",
		HTTPStatus: 500,
	})
	targetStore.AddResult(Result{
		TargetID:   tgt.ID,
		Status:     "up",
		HTTPStatus: 200,
	})

	router := http.NewServeMux()
	router.HandleFunc("/results/{id}", resultsForTargetHandler)

	// build a request with the target id in the path
	req := httptest.NewRequest(http.MethodGet, "/results/"+tgt.ID, nil)

	// go 1.22 style mux patterns require the mux to set path values
	// httptest doesn’t do that automatically, so we set the path value manually
	req.SetPathValue("id", tgt.ID)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK, got %d", rr.Code)
	}

	var results []Result
	if err := json.NewDecoder(rr.Body).Decode(&results); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results in history, got %d", len(results))
	}

	// first result should be the older one; second should be the newer one
	if results[0].Status != "down" || results[0].HTTPStatus != 500 {
		t.Fatalf("expected first result down/500, got status=%q httpStatus=%d", results[0].Status, results[0].HTTPStatus)
	}
	if results[1].Status != "up" || results[1].HTTPStatus != 200 {
		t.Fatalf("expected second result up/200, got status=%q httpStatus=%d", results[1].Status, results[1].HTTPStatus)
	}
}
