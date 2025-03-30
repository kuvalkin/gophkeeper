package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// TextareaInputModel represents a multi-line text input field.
type TextareaInputModel struct {
	area     textarea.Model
	prompt   string
	canceled bool
}

// NewTextareaInputModel creates a new TextareaInputModel with the specified prompt and placeholder.
func NewTextareaInputModel(prompt string, placeholder string) TextareaInputModel {
	area := textarea.New()

	area.Placeholder = placeholder
	area.MaxHeight = -1
	area.CharLimit = -1
	area.SetHeight(10)

	area.Focus()

	return TextareaInputModel{
		prompt: prompt,
		area:   area,
	}
}

// Canceled returns true if the input was canceled by the user (e.g., by pressing Ctrl+C or Esc).
func (m TextareaInputModel) Canceled() bool {
	return m.canceled
}

// Init initializes the TextareaInputModel and returns the initial command to start blinking the cursor.
func (m TextareaInputModel) Init() tea.Cmd {
	return textarea.Blink
}

// Update processes incoming messages and updates the state of the TextareaInputModel.
// It handles user input and determines whether to quit or continue.
func (m TextareaInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.canceled = true
			return m, tea.Quit
		}

		if keyMsg.String() == "ctrl+s" {
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.area, cmd = m.area.Update(msg)
	return m, cmd
}

// Value returns the trimmed value of the textarea input.
func (m TextareaInputModel) Value() string {
	return strings.TrimSpace(m.area.Value())
}

// View renders the current state of the textarea input as a string.
func (m TextareaInputModel) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s\n\n",
		m.prompt,
		m.area.View(),
		"(ctrl+c or esc to quit, ctrl+s to submit)",
	)
}
