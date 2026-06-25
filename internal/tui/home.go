package tui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Okwonks/go-todo/internal/client"
	"github.com/Okwonks/go-todo/internal/model"
	"github.com/Okwonks/go-todo/internal/tui/constants"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

type mode int

const (
	view mode = iota
	create
	edit
)

type mainModel struct {
	mode     mode
	client   *client.Client
	table    table.Model
	taskForm FormModel
	tasks    []model.Todo
	err      error
	width    int
	height   int
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

	f := InitTask(c, &model.Todo{})

	m := mainModel{mode: view, table: t, client: c, taskForm: f}
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update table height dynamically
		// Reserve space for title, error, helper text 
		headerHeight := 3
		footerHeight := 3
		errorHeight := 0
		if m.err != nil {
			errorHeight = 2
		}

		availableHeight := max(m.height - headerHeight - footerHeight - errorHeight, 5)
		m.table.SetHeight(availableHeight)

		// Propagate to form model
		m.taskForm.width = m.width
		m.taskForm.height = m.height

		return m, nil
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
	case CreateTask, EditTask:
		return m, nil
	case BackToRoot:
		m.mode = view
		return m, fetchTodos(m.client)
	case tea.KeyMsg:
		if m.mode == view {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "ctrl+r", "r":
				return m, fetchTodos(m.client)
			case "n":
				newTask := &model.Todo{}
				m.taskForm = InitTask(m.client, newTask)
				m.mode = create
				return m, m.taskForm.Init()
			case "e", " ":
				selectedRow := m.table.SelectedRow()
				if len(selectedRow) > 0 {
					idStr := selectedRow[0]
					id, err := strconv.ParseInt(idStr, 10, 64)
					if err == nil {
						var selectedTask *model.Todo
						for _, t := range m.tasks {
							if t.ID == id {
								selectedTask = &t
								break
							}
						}
						if selectedTask != nil {
							m.taskForm = InitTask(m.client, selectedTask)
							m.mode = edit
							return m, m.taskForm.Init()
						}
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	if m.mode == create || m.mode == edit {
		m.taskForm, cmd = m.taskForm.Update(msg)
	}

	if m.mode == view {
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m mainModel) View() string {
	style := lipgloss.NewStyle().Bold(true).Margin(1)
	title := style.Render("Todo.sh")

	const breakpoint = 120

	var content string

	switch {
	case m.width < breakpoint:
		content = m.renderStackedContent()
	default:
		content = m.renderWideContent()
	}

	errBlock := ""
	if m.err != nil {
		errBlock = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Render("Error: " + m.err.Error()) + "\n"
	}

	help := constants.HelpStyle("[q] quit • [r] refresh • [n] new task • [e] edit")

	bgView := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		content,
		errBlock,
		help,
	)

	if m.mode == create || m.mode == edit {
		form := m.taskForm.View()
		return drawModalOverlay(bgView, form, m.width, m.height)
	}

	return bgView
}

func (m mainModel) renderStackedContent() string {
	table := m.table.View()
	stats := fmt.Sprintf("Tasks: %d", len(m.tasks))
	return lipgloss.JoinVertical(lipgloss.Left, table, stats)
} 

func (m mainModel) renderWideContent() string {
	tableWidth := int(float64(m.width) * 0.4)
	tableStyle := lipgloss.NewStyle().Width(tableWidth)
	leftColumn := tableStyle.Render(m.table.View())

	sidebarWidth := int(float64(m.width - tableWidth - 2) * 0.2)
	sidebarStyle := lipgloss.NewStyle().
	  Width(sidebarWidth).
		Border(lipgloss.NormalBorder()).
		Padding(1)

	rightColumn := sidebarStyle.Render(m.renderSidebar())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)
}

func (m mainModel) renderSidebar() string {
	totalTasks := len(m.tasks)
	completedTasks := 0
	highPriority := 0

	for _, task := range m.tasks {
		if task.Completed {
			completedTasks += 1
		}

		if task.Priority >= 3 {
			highPriority += 1
		}
	}

	stats := fmt.Sprintf(
		"Stats\n\n" +
		"Total Tasks: %d\n" +
		"Completed: %d\n" +
		"High Priority: %d\n",
		totalTasks,
		completedTasks,
		highPriority,
	)

	return stats
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripAnsi(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func drawModalOverlay(bg string, modal string, width, height int) string {
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Dark gray / dimmed

	// Split background and modal into lines
	bgLines := strings.Split(bg, "\n")
	modalLines := strings.Split(modal, "\n")

	// Prepare background lines: strip ANSI, pad/truncate to screen size, and dim them
	processedBg := make([]string, height)
	for i := range height {
		var lineText string
		if i < len(bgLines) {
			lineText = stripAnsi(bgLines[i])
		}
		// Pad or truncate to match width
		visualLen := runewidth.StringWidth(lineText)
		if visualLen < width {
			lineText = lineText + strings.Repeat(" ", width-visualLen)
		} else if visualLen > width {
			lineText = runewidth.Truncate(lineText, width, "")
		}
		processedBg[i] = lineText
	}

	// Calculate vertical starting position for the modal
	modalHeight := len(modalLines)
	startY := max((height - modalHeight) / 2, 0)

	// Calculate maximum visual width of the modal
	modalWidth := 0
	for _, mLine := range modalLines {
		w := lipgloss.Width(mLine)
		if w > modalWidth {
			modalWidth = w
		}
	}

	startX := max((width - modalWidth) / 2, 0)

	// Build the final screen lines
	outputLines := make([]string, height)
	for y := range height {
		bgLine := processedBg[y]

		// Check if this vertical line contains the modal
		if y >= startY && y < startY+modalHeight {
			modalLine := modalLines[y-startY]
			visualModLen := lipgloss.Width(modalLine)

			// Slice the plain background text safely
			leftPart := runewidth.Truncate(bgLine, startX, "")
			remainder := runewidth.Truncate(bgLine, startX, "")
			rightPart := runewidth.Truncate(remainder,  visualModLen, "")

			// Render with dimmed background parts and colored modal in the middle
			outputLines[y] = dimStyle.Render(leftPart) + modalLine + dimStyle.Render(rightPart)
		} else {
			// No modal on this line, just render the dimmed background line
			outputLines[y] = dimStyle.Render(bgLine)
		}
	}

	return strings.Join(outputLines, "\n")
}
