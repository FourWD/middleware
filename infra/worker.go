package infra

import "context"

// Worker is a background goroutine managed by the App lifecycle.
// Run must honor ctx cancellation and return when ctx is done.
// Return value context.Canceled is treated as normal shutdown and is not logged as an error.
type Worker struct {
	Name string
	Run  func(ctx context.Context) error
}
