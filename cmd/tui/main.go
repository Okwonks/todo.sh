package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Okwonks/go-todo/internal/client"
	dbmodel "github.com/Okwonks/go-todo/internal/model"
	"github.com/charmbracelet/bubbles/table"
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

	columns := []table.Column{}

	t := table.New(
		table.WithColumns(columns),
		table.WithHeight(10),
	)

	return model{
		state: stateList,
		table: t,
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
		case "q":
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
			rows = append(rows, table.Row{fmt.Sprint(t.ID, t.Description, t.CreatedAt)})
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
			t := &dbmodel.Todo{
				Description: m.taskInput.description,
				DueDate:     m.taskInput.dueDate,
				Priority:    m.taskInput.priority,
			}

			_, err := m.client.CreateTodo(t)
			if err != nil {
				m.err = err
			}
			
			m.state = stateList
			return m, fetchTodos(m.client)
		}

		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.state == stateList {
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
	description string
	dueDate     *time.Time 
	priority    int
	submitted   bool
	cursor      int
}

func newTaskInputModel() taskInputModel {
	return taskInputModel{}
}

func (m taskInputModel) Init() tea.Cmd {
	return nil
}

func (m taskInputModel) Update(msg tea.Msg) (taskInputModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
	  switch msg.String() {
		case "tab":
		  m.cursor = (m.cursor + 1) % 2
		case "enter":
		  m.submitted = true
			return m, nil
		case "esc":
		  m.submitted = false
			return m, nil
		case "backspace":
		  // TODO: change this to use []inputs, making setting the values easier
		  if m.cursor == 0 && len(m.description) > 0 {
				m.description = m.description[:len(m.description)-1]
			}
		  if m.cursor == 1 && m.priority > 0 {
				m.priority -= 1
			}
		default:
		  if msg.Type == tea.KeyRunes {
				if m.cursor == 0 {
					m.description += msg.String()
				} else  {
					m.priority = min(m.priority + 1, 5)
				}
			}
		}
	}
	return m, nil
}

func (m taskInputModel) View() string {
	return fmt.Sprintf(
		"Create Task\n\nDescription:  %s\nPriority:   %d\n\n[enter] submit • [esc] cancel",
		m.description, m.priority,
	)
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Println("Error running TUI:", err)
		os.Exit(1)
	}
}
