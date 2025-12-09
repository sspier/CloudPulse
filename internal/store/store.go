package store

import (
	"context"
	"github.com/sspier/cloudpulse/internal/model"
)

// Store defines the interface for persisting targets and results
type Store interface {
	AddTarget(ctx context.Context, name, url string) (model.Target, error)
	ListTargets(ctx context.Context) ([]model.Target, error)
	AddResult(ctx context.Context, result model.Result) error
	LatestResults(ctx context.Context) ([]model.Result, error)
	ResultsForTarget(ctx context.Context, targetID string) ([]model.Result, error)
}
