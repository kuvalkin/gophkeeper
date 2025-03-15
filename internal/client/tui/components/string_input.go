package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type StringInputModel struct {
	textInput textinput.Model
	canceled  bool
}

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

func (m StringInputModel) Canceled() bool {
	return m.canceled
}

func (m StringInputModel) Init() tea.Cmd {
	return textinput.Blink
}

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

func (m StringInputModel) Value() string {
	return strings.TrimSpace(m.textInput.Value())
}

func (m StringInputModel) View() string {
	return m.textInput.View()
}
