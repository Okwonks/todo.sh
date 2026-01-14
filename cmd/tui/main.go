package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Okwonks/go-todo/internal/client"
	dbmodel "github.com/Okwonks/go-todo/internal/model"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type modelState int

const (
	stateList modelState = iota
	stateCreate
)

type model struct {
	state     modelState
	table     table.Model
	taskInput taskInputModel
	client    *client.Client
	tasks     []dbmodel.Todo
	err       error
}

func initialModel() model {
	c := client.NewClient("http://localhost:8080")

	columns := []table.Column{
		{Title: "ID",          Width: 5},
		{Title: "CreatedAt",   Width: 20},
		{Title: "Description", Width: 30},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithHeight(10),
	)

	return model{
		state:  stateList,
		table:  t,
		client: c,
	}
}

type listTodosMsg []dbmodel.Todo
type errorMsg error

func fetchTodos(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		todos, err := c.List()
		if err != nil {
			return errorMsg(err)
		}
		return listTodosMsg(todos)
	}
}

func (m model) Init() tea.Cmd {
	return fetchTodos(m.client)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
	  switch msg.String() {
		case "ctrl+c", "q":
		  return m, tea.Quit
		case "r":
		  return m, fetchTodos(m.client)
		case "n":
		  m.state = stateCreate
			m.taskInput = newTaskInputModel()
		  return m, nil
		}

	case listTodosMsg:
	  m.tasks = msg
		rows := make([]table.Row, 0, len(msg))

	  for _, t := range msg {
			rows = append(rows, table.Row{fmt.Sprint(t.ID), t.Description, t.CreatedAt.String()})
		}

	  m.table.SetRows(rows)
	  return m, nil

	case errorMsg:
	  m.err = msg
		return m, nil
	}

	if m.state == stateCreate {
		newTask, cmd := m.taskInput.Update(msg)
		m.taskInput = newTask

		if m.taskInput.submitted {
			p, err := strconv.Atoi(m.taskInput.inputs[1].Value())
			if err != nil {
				m.err = err
				return m, fetchTodos(m.client)
			}

			t := &dbmodel.Todo{
				Description: m.taskInput.inputs[0].Value(),
				Priority:    p,
			}

			_, err = m.client.CreateTodo(t)
			if err != nil {
				m.err = err
			}
			
			m.state = stateList
			return m, fetchTodos(m.client)
		}

		if !m.taskInput.submitted && m.taskInput.showList {
			m.state = stateList
			return m, fetchTodos(m.client)
		}

		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.state == stateCreate {
		return m.taskInput.View()
	}

	style := lipgloss.NewStyle().Bold(true).Margin(1)
	title := style.Render("Task")

	errBlock := ""
	if m.err != nil {
		errBlock = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Render("Error: " + m.err.Error()) + "\n"
	}

	help := "\n[q] quit • [r] refresh • [n] new task"

	return fmt.Sprintf("%s\n%s\n%s", title, errBlock, m.table.View()) + help
}

type taskInputModel struct {
	submitted   bool
	inputs      []textinput.Model
	focusIndex  int
	cursorMode  cursor.Mode
	showList    bool
}

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle
	noStyle             = lipgloss.NewStyle()
	// helpStyle           = blurredStyle
	// cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

func newTaskInputModel() taskInputModel {
	m := taskInputModel{
		inputs: make([]textinput.Model, 2),
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

func (m taskInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m taskInputModel) Update(msg tea.Msg) (taskInputModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
	  switch msg.String() {
		case "esc":
		  m.submitted = false
			m.focusIndex = 0
			m.showList = true
			return m, nil
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.inputs) {
				m.submitted = true
				return m, nil
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

func (m *taskInputModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m taskInputModel) View() string {
	var b strings.Builder

	style := lipgloss.NewStyle().Bold(true).Margin(1)
	title := style.Render("Task")
	b.WriteString(title)
	b.WriteRune('\n')

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


	help := "\n[q] quit • [esc] go back"

	return b.String() + help
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Println("Error running TUI:", err)
		os.Exit(1)
	}
}
