package prompts

import (
	"context"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type cancellable interface {
	Canceled() bool
}

// ErrCanceled is returned when a prompt is canceled by the user.
var ErrCanceled = errors.New("cancelled")

func run[T tea.Model](ctx context.Context, initModel T, altScreen bool) (T, error) {
	options := []tea.ProgramOption{
		tea.WithContext(ctx),
	}
	if altScreen {
		options = append(options, tea.WithAltScreen())
	}

	finalModel, err := tea.NewProgram(initModel, options...).Run()

	if err != nil {
		return finalModel.(T), fmt.Errorf("run error: %w", err)
	}

	if finalModelCancellable, ok := finalModel.(cancellable); ok {
		if finalModelCancellable.Canceled() {
			return finalModel.(T), ErrCanceled
		}
	}

	return finalModel.(T), nil
}
