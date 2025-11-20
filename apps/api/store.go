package main

import (
	"sync"
	"time"
)

// InMemoryStore holds all targets and probe results in memory.
// This is for local dev only; in the cloud version this will become DynamoDB.
type InMemoryStore struct {
	mu      sync.RWMutex
	targets map[string]Target
	results map[string][]Result
}

// NewInMemoryStore creates a new empty store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		targets: make(map[string]Target),
		results: make(map[string][]Result),
	}
}

// AddTarget inserts a new target and returns it.
func (store *InMemoryStore) AddTarget(name, url string) Target {
	store.mu.Lock()
	defer store.mu.Unlock()

	id := time.Now().UTC().Format("20060102150405.000000000")

	target := Target{
		ID:   id,
		Name: name,
		URL:  url,
	}

	store.targets[id] = target
	return target
}

// ListTargets returns all known targets.
func (store *InMemoryStore) ListTargets() []Target {
	store.mu.RLock()
	defer store.mu.RUnlock()

	targets := make([]Target, 0, len(store.targets))
	for _, target := range store.targets {
		targets = append(targets, target)
	}
	return targets
}

// AddResult appends a probe result for a target.
func (store *InMemoryStore) AddResult(result Result) {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.results[result.TargetID] = append(store.results[result.TargetID], result)
}

// LatestResults returns the most recent result for each target.
func (store *InMemoryStore) LatestResults() []Result {
	store.mu.RLock()
	defer store.mu.RUnlock()

	latest := make([]Result, 0, len(store.targets))

	for id := range store.targets {
		resultsForTarget := store.results[id]
		if len(resultsForTarget) == 0 {
			continue
		}
		latestResult := resultsForTarget[len(resultsForTarget)-1]
		latest = append(latest, latestResult)
	}

	return latest
}

// ResultsForTarget returns the full history for a specific target.
func (store *InMemoryStore) ResultsForTarget(id string) []Result {
	store.mu.RLock()
	defer store.mu.RUnlock()

	if r, ok := store.results[id]; ok {
		// return a copy so nobody mutates internal storage
		out := make([]Result, len(r))
		copy(out, r)
		return out
	}

	return []Result{} // empty array, not null
}
