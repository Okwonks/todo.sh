package tui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Okwonks/go-todo/internal/client"
	"github.com/Okwonks/go-todo/internal/model"
	"github.com/Okwonks/go-todo/internal/tui/constants"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mode int

const (
	view mode = iota
	create
)

type mainModel struct {
	mode        mode
	client      *client.Client
	table       table.Model
	newTaskForm FormModel
	tasks       []model.Todo
	err         error
}

func InitRoot(c *client.Client) tea.Model {
	columns := []table.Column{
		{Title: "ID",          Width:  4},
		{Title: "Description", Width: 40},
		{Title: "P",           Width:  2},
		{Title: "Due Date",    Width: 30},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true)
	s.Selected = s.Selected.
		Background(lipgloss.Color("229")).
		Bold(false)
	t.SetStyles(s)

	f := InitNewTask(c)

	m := mainModel{mode: view, table: t, client: c, newTaskForm: f} 
	return m
}

type listTodosMsg []model.Todo
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

func formatDueDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func (m mainModel) Init() tea.Cmd {
  return fetchTodos(m.client)
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case listTodosMsg:
	  m.tasks = msg
		rows := make([]table.Row, 0, len(msg))

	  for _, t := range msg {
			dueDate := formatDueDate(t.DueDate) 
			rows = append(rows, table.Row{fmt.Sprint(t.ID), t.Description, strconv.Itoa(t.Priority), dueDate})
		}

	  m.table.SetRows(rows)
	  return m, nil
	case errorMsg:
	  m.err = msg
		return m, nil
	case CreateTask:
	  return m, nil
	case BackToRoot:
		if msg.NewTask != nil {
			t := msg.NewTask
			rows := m.table.Rows()
			dueDate := formatDueDate(t.DueDate) 
			rows = append(rows,  table.Row{fmt.Sprint(t.ID), t.Description, strconv.Itoa(t.Priority), dueDate})
			m.table.SetRows(rows)
		}
	  m.mode = view
		return m, nil
	case tea.KeyMsg:
	  if m.mode == view {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "ctrl+r", "r":
				return m, fetchTodos(m.client)
			case "n":
				m.mode = create

				return m, func() tea.Msg { return CreateTask{} }
			}
		}
	}

	var cmd tea.Cmd
	if m.mode == create {
		m.newTaskForm, cmd = m.newTaskForm.Update(msg)
	}

	if m.mode == view {
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m mainModel) View() string {
	style := lipgloss.NewStyle().Bold(true).Margin(1)
	title := style.Render("Todo.sh")

	if m.mode == create {
		return fmt.Sprintf("%s\n%s", title, m.newTaskForm.View())
	}

	errBlock := ""
	if m.err != nil {
		errBlock = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Render("Error: " + m.err.Error()) + "\n"
	}

	help := constants.HelpStyle("\n\n[q] quit • [r] refresh • [n] new task")

	return fmt.Sprintf("%s\n%s\n%s", title, errBlock, m.table.View()) + help
}
