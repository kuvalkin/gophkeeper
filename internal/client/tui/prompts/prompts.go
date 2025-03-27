package prompts

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/kuvalkin/gophkeeper/internal/client/tui/components"
)

type Prompter interface {
	AskString(ctx context.Context, prompt string, placeholder string) (string, error)
	AskPassword(ctx context.Context, prompt string, placeholder string) (string, error)
	AskInt(ctx context.Context, prompt string, placeholder string) (int, error)
	AskText(ctx context.Context, prompt string, placeholder string) (string, error)
	Confirm(ctx context.Context, prompt string) bool
}

type TerminalPrompter struct{}

func NewTerminalPrompter() *TerminalPrompter {
	return &TerminalPrompter{}
}

func (p *TerminalPrompter) AskString(ctx context.Context, prompt string, placeholder string) (string, error) {
	finalModel, err := run(ctx, components.NewStringInputModel(prompt, placeholder, false), false)
	if err != nil {
		return "", fmt.Errorf("error running prompt: %w", err)
	}

	return finalModel.Value(), nil
}

func (p *TerminalPrompter) AskPassword(ctx context.Context, prompt string, placeholder string) (string, error) {
	finalModel, err := run(ctx, components.NewStringInputModel(prompt, placeholder, true), false)
	if err != nil {
		return "", fmt.Errorf("error running prompt: %w", err)
	}

	return finalModel.Value(), nil
}

func (p *TerminalPrompter) AskInt(ctx context.Context, prompt string, placeholder string) (int, error) {
	val, err := p.AskString(ctx, prompt, placeholder)
	if err != nil {
		return 0, err
	}

	valueInt, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("error converting value to int: %w", err)
	}

	return valueInt, nil
}

func (p *TerminalPrompter) AskText(ctx context.Context, prompt string, placeholder string) (string, error) {
	finalModel, err := run(ctx, components.NewTextareaInputModel(prompt, placeholder), true)
	if err != nil {
		return "", fmt.Errorf("error running prompt: %w", err)
	}

	return finalModel.Value(), nil
}

func (p *TerminalPrompter) Confirm(ctx context.Context, prompt string) bool {
	confirm, err := p.AskString(ctx, fmt.Sprintf("%s [Y/n]", prompt), "y")
	if err != nil {
		return false
	}

	if confirm == "" || strings.ToLower(confirm) == "y" || strings.ToLower(confirm) == "yes" {
		return true
	}

	return false
}
