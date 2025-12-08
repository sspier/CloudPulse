package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sspier/cloudpulse/internal/model"
)

// InMemoryStore keeps all targets and probe results in memory
type InMemoryStore struct {
	// rwMutex is a read-write mutex that protects concurrent access to the store
	// it allows multiple readers or a single writer
	rwMutex        sync.RWMutex
	sequenceNumber int64 // sequence number for generating unique IDs
	targets        map[string]model.Target
	results        map[string][]model.Result
}

// NewInMemoryStore sets up empty maps so the store is ready to use
func NewInMemoryStore() *InMemoryStore {
	// create the store
	return &InMemoryStore{
		targets: make(map[string]model.Target),
		results: make(map[string][]model.Result),
	}
}

// AddTarget registers a new target and returns it
func (inMemoryStore *InMemoryStore) AddTarget(ctx context.Context, name, url string) (model.Target, error) {
	// store is protected by a mutex, so we need to lock it
	inMemoryStore.rwMutex.Lock()
	// unlock when we're done
	defer inMemoryStore.rwMutex.Unlock()

	// increment the sequence number
	inMemoryStore.sequenceNumber++

	// generate a unique ID for the target using the current time and sequence number
	uniqueId := fmt.Sprintf(
		"%s-%d",
		time.Now().UTC().Format("20060102150405.000000000"),
		inMemoryStore.sequenceNumber,
	)

	// create the target
	target := model.Target{
		ID:   uniqueId,
		Name: name,
		URL:  url,
	}

	// add the target to the store
	inMemoryStore.targets[uniqueId] = target
	return target, nil
}

// ListTargets returns all targets as a slice
func (inMemoryStore *InMemoryStore) ListTargets(ctx context.Context) ([]model.Target, error) {
	inMemoryStore.rwMutex.RLock()
	defer inMemoryStore.rwMutex.RUnlock()

	targets := make([]model.Target, 0, len(inMemoryStore.targets))
	for _, target := range inMemoryStore.targets {
		targets = append(targets, target)
	}
	return targets, nil
}

// AddResult appends a new probe result for a given target
func (inMemoryStore *InMemoryStore) AddResult(ctx context.Context, result model.Result) error {
	inMemoryStore.rwMutex.Lock()
	defer inMemoryStore.rwMutex.Unlock()

	inMemoryStore.results[result.TargetID] = append(inMemoryStore.results[result.TargetID], result)
	return nil
}

// LatestResults returns the most recent result for each target
func (inMemoryStore *InMemoryStore) LatestResults(ctx context.Context) ([]model.Result, error) {
	inMemoryStore.rwMutex.RLock()
	defer inMemoryStore.rwMutex.RUnlock()

	latest := make([]model.Result, 0, len(inMemoryStore.targets))

	for id := range inMemoryStore.targets {
		// get the latest result for this target
		resultsForTarget := inMemoryStore.results[id]
		if len(resultsForTarget) == 0 {
			// no results for this target
			continue
		}
		latestResult := resultsForTarget[len(resultsForTarget)-1]
		latest = append(latest, latestResult)
	}

	return latest, nil
}

// ResultsForTarget returns the full probe history for a single target
func (inMemoryStore *InMemoryStore) ResultsForTarget(ctx context.Context, id string) ([]model.Result, error) {
	// inMemoryStore is protected by a mutex, so we need to lock it
	inMemoryStore.rwMutex.RLock()
	// unlock when we're done
	defer inMemoryStore.rwMutex.RUnlock()

	// if the target exists, return its results
	// otherwise return an empty slice
	if resultsForTarget, ok := inMemoryStore.results[id]; ok {
		// make a copy of the resultsCopy to avoid race conditions
		resultsCopy := make([]model.Result, len(resultsForTarget))
		// copy the results into the new slice
		copy(resultsCopy, resultsForTarget)
		// return the copy
		return resultsCopy, nil
	}

	// if the target doesn't exist, return an empty slice
	return []model.Result{}, nil
}
