package prompts

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type cancellable interface {
	Canceled() bool
}

var ErrCanceled = errors.New("cancelled")

func run[T tea.Model](initModel T) (T, error) {
	finalModel, err := tea.NewProgram(initModel).Run()

	if err != nil {
		return finalModel.(T), fmt.Errorf("run error: %w", err)
	}

	if finalModelCancellabel, ok := finalModel.(cancellable); ok {
		if finalModelCancellabel.Canceled() {
			return finalModel.(T), ErrCanceled
		}
	}

	return finalModel.(T), nil
}
