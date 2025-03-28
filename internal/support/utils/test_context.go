package utils

import (
	"context"
	"testing"
)

func TestContext(t *testing.T) (context.Context, context.CancelFunc) {
	deadline, ok := t.Deadline()
	if !ok {
		return context.WithCancel(context.Background())
	}

	return context.WithDeadline(context.Background(), deadline)
}
