package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var targetStore = NewInMemoryStore()

// shared http client with a timeout so probes don’t hang forever
var httpClient = &http.Client{
	Timeout: 5 * time.Second,
}

// runCheck performs a single probe of the target url and records the result
// this is called both when a target is created and by the background scheduler
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

	// store the probe result so it can be retrieved via GET /results and GET /results/{id}
	targetStore.AddResult(Result{
		TargetID:   t.ID,
		Status:     status,
		HTTPStatus: httpStatus,
	})
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
		targets := targetStore.ListTargets()

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
		created := targetStore.AddTarget(payload.Name, payload.URL)

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

	results := targetStore.LatestResults()
	if results == nil {
		// always return [] instead of null for better json ergonomics
		results = []Result{}
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
	results := targetStore.ResultsForTarget(id)
	if results == nil {
		results = []Result{}
	}

	responseWriter.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(responseWriter).Encode(results); err != nil {
		log.Println("error encoding results:", err)
	}
}
