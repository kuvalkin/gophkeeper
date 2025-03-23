package prompts

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/kuvalkin/gophkeeper/internal/client/tui/components"
)

func AskString(ctx context.Context, prompt string, placeholder string) (string, error) {
	finalModel, err := run(ctx, components.NewStringInputModel(prompt, placeholder, false), false)
	if err != nil {
		return "", fmt.Errorf("error running prompt: %w", err)
	}

	return finalModel.Value(), nil
}

func AskPassword(ctx context.Context, prompt string, placeholder string) (string, error) {
	finalModel, err := run(ctx, components.NewStringInputModel(prompt, placeholder, true), false)
	if err != nil {
		return "", fmt.Errorf("error running prompt: %w", err)
	}

	return finalModel.Value(), nil
}

func AskInt(ctx context.Context, prompt string, placeholder string) (int, error) {
	finalModel, err := run(ctx, components.NewStringInputModel(prompt, placeholder, true), false)
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

func AskText(ctx context.Context, prompt string, placeholder string) (string, error) {
	finalModel, err := run(ctx, components.NewTextareaInputModel(prompt, placeholder), true)
	if err != nil {
		return "", fmt.Errorf("error running prompt: %w", err)
	}

	return finalModel.Value(), nil
}

func Confirm(ctx context.Context, prompt string) bool {
	confirm, err := AskString(ctx, fmt.Sprintf("%s [Y/n]", prompt), "y")
	if err != nil {
		return false
	}

	if confirm == "" || strings.ToLower(confirm) == "y" || strings.ToLower(confirm) == "yes" {
		return true
	}

	return false
}
