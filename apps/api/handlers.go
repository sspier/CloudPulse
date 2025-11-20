package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var targetStore = NewInMemoryStore()

var httpClient = &http.Client{
	Timeout: 5 * time.Second,
}

// runCheck performs a single probe of the target URL and records the result.
func runCheck(t Target) {
	resp, err := httpClient.Get(t.URL)
	status := "down"
	httpStatus := 0

	if err == nil {
		httpStatus = resp.StatusCode
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			status = "up"
		}
		resp.Body.Close()
	}

	targetStore.AddResult(Result{
		TargetID:   t.ID,
		Status:     status,
		HTTPStatus: httpStatus,
	})
}

// healthHandler: used by load balancers, Kubernetes, or humans
func healthHandler(responseWriter http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(responseWriter, "ok")
}

// targetsHandler: handles GET and POST /targets
func targetsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	responseWriter.Header().Set("Content-Type", "application/json")

	switch request.Method {
	case http.MethodGet:
		targets := targetStore.ListTargets()

		responseWriter.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(responseWriter).Encode(targets); err != nil {
			log.Println("error encoding targets:", err)
		}

	case http.MethodPost:
		var payload struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		}

		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			http.Error(responseWriter, "invalid JSON body", http.StatusBadRequest)
			return
		}

		created := targetStore.AddTarget(payload.Name, payload.URL)

		// Kick off an immediate check in the background
		go runCheck(created)

		responseWriter.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(responseWriter).Encode(created); err != nil {
			log.Println("error encoding created target:", err)
		}
	}
}

// resultsHandler: returns the latest result for each target
func resultsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		responseWriter.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")

	results := targetStore.LatestResults()
	if results == nil {
		// Encode as [] instead of null
		results = []Result{}
	}

	responseWriter.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(responseWriter).Encode(results); err != nil {
		log.Println("error encoding results:", err)
	}
}

// resultsForTargetHandler returns all results for a specific target ID.
func resultsForTargetHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		responseWriter.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := request.PathValue("id")
	if id == "" {
		http.Error(responseWriter, "target ID required", http.StatusBadRequest)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")

	results := targetStore.ResultsForTarget(id)
	if results == nil {
		results = []Result{}
	}

	responseWriter.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(responseWriter).Encode(results); err != nil {
		log.Println("error encoding results:", err)
	}
}
