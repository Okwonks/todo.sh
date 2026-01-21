package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Okwonks/go-todo/internal/client"
	"github.com/Okwonks/go-todo/internal/model"
	"github.com/Okwonks/go-todo/internal/tui/constants"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CreateTask struct {}

type BackToRoot struct {
	NewTask *model.Todo
}

type FormModel struct {
	submitted   bool
	inputs      []textinput.Model
	focusIndex  int
	cursorMode  cursor.Mode
	client      *client.Client
	err         error
}

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle
	noStyle             = lipgloss.NewStyle()

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

func InitNewTask(c *client.Client) FormModel {
	m := FormModel{
		inputs: make([]textinput.Model, 2),
		client: c,
	}

	var ti textinput.Model
	for i := range m.inputs {
		ti = textinput.New()
		ti.Cursor.Style = cursorStyle
		ti.CharLimit = 200
		ti.Width = 200

		switch i {
		case 0:
		  ti.Placeholder = "Task"
			ti.Focus()
		  ti.PromptStyle = focusedStyle
			ti.TextStyle = focusedStyle
		case 1:
		  ti.Placeholder = "Priority"
			ti.CharLimit = 1
			ti.Validate = func(s string) error {
				if s == "" {
					return nil
				}
				for _, r := range s {
					if r < '1' || r > '5' {
						return fmt.Errorf("only digits between 1 and 5 allowed")
					}
				}
				return nil
			}
		}
		m.inputs[i] = ti
	}

	return m
}

func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
	  switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return BackToRoot{} }
		case "ctrl+c":
		  return m, tea.Quit
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.inputs) {
				p, err := strconv.Atoi(m.inputs[1].Value())
				if err != nil {
					m.err = err
					return m, nil
				}

				t := &model.Todo{
					Description: m.inputs[0].Value(),
					Priority:    p,
				}

				ct, err := m.client.CreateTodo(t)
				if err != nil {
					m.err = err
				}

				for i := range m.inputs {
					m.inputs[i].Reset()
					if i == 0 {
						m.inputs[i].Focus()
						m.inputs[i].PromptStyle = focusedStyle
						m.inputs[i].TextStyle = focusedStyle
					}
				}
				m.focusIndex = 0

				return m, func() tea.Msg {
					return BackToRoot{NewTask: ct}
				}
			}

		  if s == "up" || s == "shift+tab" {
				m.focusIndex--;
			} else {
				m.focusIndex++;
			}

		  if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
		  }

		  cmds := make([]tea.Cmd, len(m.inputs))
		  for i := 0; i < len(m.inputs); i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

		  return m, tea.Batch(cmds...)
		}
	}

	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *FormModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m FormModel) View() string {
	var b strings.Builder

	for i:= range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	help := constants.HelpStyle("\n[ctrl+c] quit • [esc] go back")

	return b.String() + help
}
