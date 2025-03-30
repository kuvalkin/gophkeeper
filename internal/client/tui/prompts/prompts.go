// Package prompts provides utilities for creating and handling terminal-based user prompts.
package prompts

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/kuvalkin/gophkeeper/internal/client/tui/components"
)

// Prompter defines an interface for various types of user prompts.
type Prompter interface {
	// AskString prompts the user for a string input.
	AskString(ctx context.Context, prompt string, placeholder string) (string, error)
	// AskPassword prompts the user for a password input (hidden).
	AskPassword(ctx context.Context, prompt string, placeholder string) (string, error)
	// AskInt prompts the user for an integer input.
	AskInt(ctx context.Context, prompt string, placeholder string) (int, error)
	// AskText prompts the user for multi-line text input.
	AskText(ctx context.Context, prompt string, placeholder string) (string, error)
	// Confirm prompts the user for a yes/no confirmation.
	Confirm(ctx context.Context, prompt string) bool
}

// TerminalPrompter is an implementation of the Prompter interface for terminal-based prompts.
type TerminalPrompter struct{}

// NewTerminalPrompter creates a new instance of TerminalPrompter.
func NewTerminalPrompter() *TerminalPrompter {
	return &TerminalPrompter{}
}

// AskString prompts the user for a string input.
func (p *TerminalPrompter) AskString(ctx context.Context, prompt string, placeholder string) (string, error) {
	finalModel, err := run(ctx, components.NewStringInputModel(prompt, placeholder, false), false)
	if err != nil {
		return "", fmt.Errorf("error running prompt: %w", err)
	}

	return finalModel.Value(), nil
}

// AskPassword prompts the user for a password input (hidden).
func (p *TerminalPrompter) AskPassword(ctx context.Context, prompt string, placeholder string) (string, error) {
	finalModel, err := run(ctx, components.NewStringInputModel(prompt, placeholder, true), false)
	if err != nil {
		return "", fmt.Errorf("error running prompt: %w", err)
	}

	return finalModel.Value(), nil
}

// AskInt prompts the user for an integer input.
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

// AskText prompts the user for multi-line text input.
func (p *TerminalPrompter) AskText(ctx context.Context, prompt string, placeholder string) (string, error) {
	finalModel, err := run(ctx, components.NewTextareaInputModel(prompt, placeholder), true)
	if err != nil {
		return "", fmt.Errorf("error running prompt: %w", err)
	}

	return finalModel.Value(), nil
}

// Confirm prompts the user for a yes/no confirmation.
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
