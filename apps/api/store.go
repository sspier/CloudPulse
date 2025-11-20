package main

import (
	"fmt"
	"sync"
	"time"
)

// inMemoryStore keeps all targets and probe results in memory
// used only for local dev; cloud version will swap this out for dynamodb or another persistent store
type InMemoryStore struct {
	mu      sync.RWMutex        // protects both targets and results
	seq     int64               // simple sequence to make ids unique
	targets map[string]Target   // targetID -> target
	results map[string][]Result // targetID -> list of probe results
}

// newInMemoryStore sets up empty maps so the store is ready to use
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		targets: make(map[string]Target),
		results: make(map[string][]Result),
	}
}

// addTarget registers a new target and returns it
// ids are timestamp-based plus a local sequence to avoid collisions
func (store *InMemoryStore) AddTarget(name, url string) Target {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.seq++

	id := fmt.Sprintf(
		"%s-%d",
		time.Now().UTC().Format("20060102150405.000000000"),
		store.seq,
	)

	target := Target{
		ID:   id,
		Name: name,
		URL:  url,
	}

	store.targets[id] = target
	return target
}

// listTargets returns all targets as a slice
// slice is rebuilt each call to avoid exposing internal maps
func (store *InMemoryStore) ListTargets() []Target {
	store.mu.RLock()
	defer store.mu.RUnlock()

	targets := make([]Target, 0, len(store.targets))
	for _, target := range store.targets {
		targets = append(targets, target)
	}
	return targets
}

// addResult appends a new probe result for a given target
// results are stored as a growing slice; in the cloud version we’ll cap or paginate this
func (store *InMemoryStore) AddResult(result Result) {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.results[result.TargetID] = append(store.results[result.TargetID], result)
}

// latestResults returns the most recent result for each target
// used by GET /results to show a snapshot of overall health
func (store *InMemoryStore) LatestResults() []Result {
	store.mu.RLock()
	defer store.mu.RUnlock()

	latest := make([]Result, 0, len(store.targets))

	for id := range store.targets {
		resultsForTarget := store.results[id]
		if len(resultsForTarget) == 0 {
			// target exists but has no checks yet
			continue
		}

		// grab the most recent result for this target
		latestResult := resultsForTarget[len(resultsForTarget)-1]
		latest = append(latest, latestResult)
	}

	return latest
}

// resultsForTarget returns the full probe history for a single target
// caller gets a copy so they can’t mutate internal state
func (store *InMemoryStore) ResultsForTarget(id string) []Result {
	store.mu.RLock()
	defer store.mu.RUnlock()

	if r, ok := store.results[id]; ok {
		out := make([]Result, len(r))
		copy(out, r)
		return out
	}

	// always return [] instead of nil to align with json expectations
	return []Result{}
}
