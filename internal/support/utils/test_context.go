// Package utils provides utility functions and helpers for common operations.
package utils

import (
	"context"
	"testing"
)

// TestContext creates a context with a deadline derived from the testing.T instance.
// If no deadline is set on the testing.T, it returns a cancellable background context.
func TestContext(t *testing.T) (context.Context, context.CancelFunc) {
	deadline, ok := t.Deadline()
	if !ok {
		return context.WithCancel(context.Background())
	}

	return context.WithDeadline(context.Background(), deadline)
}
