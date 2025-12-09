package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sspier/cloudpulse/internal/model"
	"github.com/sspier/cloudpulse/internal/probe"
	"github.com/sspier/cloudpulse/internal/store"
)

// targetStore is the global store instance.
// In a real app, we might inject this dependency.
var targetStore store.Store = NewInMemoryStore()

// runCheck performs a single probe of the target url and records the result
// this is called both when a target is created and by the background scheduler
func runCheck(t model.Target) {
	// Use the shared probe logic
	result := probe.Check(context.Background(), t)

	// store the probe result so it can be retrieved via GET /results and GET /results/{id}
	targetStore.AddResult(context.Background(), result)
}

// simple health endpoint used by load balancers and humans
func healthHandler(responseWriter http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(responseWriter, "ok")
}

// targetsHandler handles target creation and listing
// GET returns all targets
// POST registers a new target and kicks off an immediate probe
func targetsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	responseWriter.Header().Set("Content-Type", "application/json")

	switch request.Method {

	case http.MethodGet:
		// pull all targets from the store
		targets, err := targetStore.ListTargets(request.Context())
		if err != nil {
			http.Error(responseWriter, "internal error", http.StatusInternalServerError)
			return
		}

		responseWriter.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(responseWriter).Encode(targets); err != nil {
			log.Println("error encoding targets:", err)
		}

	case http.MethodPost:
		// small inline struct for decoding POST body
		var payload struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		}

		// reject invalid json bodies or missing fields
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			http.Error(responseWriter, "invalid JSON body", http.StatusBadRequest)
			return
		}

		// create the target
		created, err := targetStore.AddTarget(request.Context(), payload.Name, payload.URL)
		if err != nil {
			http.Error(responseWriter, "failed to create target", http.StatusInternalServerError)
			return
		}

		// run an immediate uptime check in the background
		go runCheck(created)

		responseWriter.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(responseWriter).Encode(created); err != nil {
			log.Println("error encoding created target:", err)
		}
	}
}

// resultsHandler returns the most recent probe result for each target
// this is used for dashboards where you want an at-a-glance view
func resultsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		responseWriter.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")

	results, err := targetStore.LatestResults(request.Context())
	if err != nil {
		http.Error(responseWriter, "internal error", http.StatusInternalServerError)
		return
	}

	if results == nil {
		// always return [] instead of null for better json ergonomics
		results = []model.Result{}
	}

	responseWriter.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(responseWriter).Encode(results); err != nil {
		log.Println("error encoding results:", err)
	}
}

// resultsForTargetHandler returns the full probe history for a given target
// useful for charts, graphs, or debugging uptime issues
func resultsForTargetHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		responseWriter.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// extract target id from URL path
	id := request.PathValue("id")
	if id == "" {
		http.Error(responseWriter, "target ID required", http.StatusBadRequest)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")

	// fetch the full list of results for this specific target
	results, err := targetStore.ResultsForTarget(request.Context(), id)
	if err != nil {
		http.Error(responseWriter, "internal error", http.StatusInternalServerError)
		return
	}

	if results == nil {
		results = []model.Result{}
	}

	responseWriter.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(responseWriter).Encode(results); err != nil {
		log.Println("error encoding results:", err)
	}
}
