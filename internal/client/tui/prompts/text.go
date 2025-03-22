package prompts

import (
	"fmt"
	"strconv"

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

func AskInt(prompt string, placeholder string) (int, error) {
	finalModel, err := run(components.NewStringInputModel(prompt, placeholder, true))
	if err != nil {
		return 0, fmt.Errorf("error running prompt: %w", err)
	}

	value := finalModel.Value()
	valueInt, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("error converting value to int: %w", err)
	}

	return valueInt, nil
}
