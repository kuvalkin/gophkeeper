package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type TextareaInputModel struct {
	area     textarea.Model
	prompt   string
	canceled bool
}

func NewTextareaInputModel(prompt string, placeholder string) TextareaInputModel {
	area := textarea.New()

	area.Placeholder = placeholder

	area.Focus()

	return TextareaInputModel{
		prompt: prompt,
		area:   area,
	}
}

func (m TextareaInputModel) Canceled() bool {
	return m.canceled
}

func (m TextareaInputModel) Init() tea.Cmd {
	return textarea.Blink
}

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

func (m TextareaInputModel) Value() string {
	return strings.TrimSpace(m.area.Value())
}

func (m TextareaInputModel) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s\n\n",
		m.prompt,
		m.area.View(),
		"(ctrl+c or esc to quit, ctrl+s to submit)",
	)
}
