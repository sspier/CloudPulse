package main

import (
	"context"
	"log"
	"sync"

	"github.com/sspier/cloudpulse/internal/model"
	"github.com/sspier/cloudpulse/internal/probe"
	"github.com/sspier/cloudpulse/internal/store"
)

type Handler struct {
	// used to persist targets and results
	store store.Store
}

// HandleRequest is the entry point for the handler
// it is called by the AWS Lambda runtime
func (handler *Handler) HandleRequest(ctx context.Context) (string, error) {
	// list all listOfTargets
	listOfTargets, err := handler.store.ListTargets(ctx)

	// if there are no targets, return an error
	if err != nil {
		log.Printf("failed to list targets: %v", err)
		return "", err
	}
	if len(listOfTargets) == 0 {
		return "no targets to probe", nil
	}

	log.Printf("starting probes for %d targets", len(listOfTargets))

	// wait group to wait for all probes to complete
	// - basically a counter + latch that blocks until all registered goroutines signal they are done
	// - used to coordinate concurrent tasks in Go
	var waitGroup sync.WaitGroup

	// run probes for each target
	for _, target := range listOfTargets {
		// add to wait group
		waitGroup.Add(1)
		// run probe in a goroutine
		go func(target model.Target) {
			// when done, decrement wait group
			defer waitGroup.Done()
			// run probe and store result
			result := probe.Check(ctx, target)
			if err := handler.store.AddResult(ctx, result); err != nil {
				log.Printf("failed to store result for %s: %v", target.ID, err)
			}
		}(target)
	}

	// wait for all probes to complete
	waitGroup.Wait()
	return "probes completed", nil
}
