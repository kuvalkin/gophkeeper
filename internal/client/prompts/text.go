package prompts

import (
	"fmt"

	"github.com/kuvalkin/gophkeeper/internal/client/tui/components"
)

func AskString(prompt string, placeholder string) (string, error) {
	finalModel, err := run(components.NewStringInputModel(prompt, placeholder, false))
	if err != nil {
		return "", fmt.Errorf("error running prompt: %w", err)
	}

	return finalModel.Value(), nil
}

func AskPassword(prompt string, placeholder string) (string, error) {
	finalModel, err := run(components.NewStringInputModel(prompt, placeholder, true))
	if err != nil {
		return "", fmt.Errorf("error running prompt: %w", err)
	}

	return finalModel.Value(), nil
}
