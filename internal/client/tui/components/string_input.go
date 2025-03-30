// Package components provides reusable UI components for the TUI client.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// StringInputModel represents a text input field with optional secret mode.
type StringInputModel struct {
	textInput textinput.Model
	canceled  bool
}

// NewStringInputModel creates a new StringInputModel with the specified prompt, placeholder, and secret mode.
// If secret is true, the input will be masked.
func NewStringInputModel(prompt string, placeholder string, secret bool) StringInputModel {
	textInput := textinput.New()

	textInput.Prompt = fmt.Sprintf("%s ", prompt)
	textInput.Placeholder = placeholder
	if secret {
		textInput.EchoMode = textinput.EchoPassword
	}

	textInput.Focus()

	return StringInputModel{
		textInput: textInput,
	}
}

// Canceled returns true if the input was canceled by the user (e.g., by pressing Ctrl+C or Esc).
func (m StringInputModel) Canceled() bool {
	return m.canceled
}

// Init initializes the StringInputModel and returns the initial command to start blinking the cursor.
func (m StringInputModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update processes incoming messages and updates the state of the StringInputModel.
// It handles user input and determines whether to quit or continue.
func (m StringInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.canceled = true
			return m, tea.Quit
		case tea.KeyEnter:
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// Value returns the trimmed value of the text input.
func (m StringInputModel) Value() string {
	return strings.TrimSpace(m.textInput.Value())
}

// View renders the current state of the text input as a string.
func (m StringInputModel) View() string {
	return m.textInput.View()
}
