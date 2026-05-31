package collector

import "context"

// Collector is the interface every metric collector must implement.
// Run executes a full collection cycle and returns an error only when
// the entire run has failed (every target failed for every metric).
// Partial failures are logged internally and do not surface as errors.
type Collector interface {
	Run(ctx context.Context) error
}
