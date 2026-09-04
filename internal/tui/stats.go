package tui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Okwonks/go-todo/internal/model"
	"github.com/Okwonks/go-todo/internal/tui/constants"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type StatsModel struct {
	tasks  []model.Todo
	width  int
	height int
}

func InitStats() StatsModel {
	return StatsModel{}
}

func (m StatsModel) SetTasks(tasks []model.Todo) StatsModel {
	m.tasks = tasks
	return m
}

func (m StatsModel) Init() tea.Cmd {
	return nil
}

func (m StatsModel) View() string {
	if len(m.tasks)== 0 {
		return lipgloss.NewStyle().
			Padding(2).
			Render("No tasks available")
	}

	data := m.calculate()

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("252")).
		Render("Todo.sh - stats")

	help := constants.HelpStyle("[q] quit • [r] refresh • [tab] switch view")

	header := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
	)

	cards := lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderStatCard(
			"total tasks",
			strconv.Itoa(data.total),
			"",
		),
		renderStatCard(
			"completed",
			strconv.Itoa(data.completed),
			fmt.Sprintf("(%d%%)", percentage(data.completed, data.total)),
		),
		renderStatCard(
			"overdue",
			strconv.Itoa(data.overdue),
			"",
		),
	)

	graph := renderCompletionGraph(data.daily)

	priority := renderPriority(
		data.high,
		data.med,
		data.low,
	)

	activity := renderActivity(data.recent)

	bottom := lipgloss.JoinHorizontal(
		lipgloss.Top,
		priority,
		activity,
	)

	return lipgloss.NewStyle().
		Padding(1, 2).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				header,
				cards,
				graph,
				bottom,
				help,
			),
		)
}

func (m StatsModel) Update(msg tea.Msg) (StatsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

type statsData struct {
	total     int
	completed int
	overdue   int

	high int
	med  int
	low  int

	daily []int

	recent []model.Todo
}

func (m StatsModel) calculate() statsData {
	now := time.Now()

	res := statsData{total: len(m.tasks), daily: make([]int, 14)}

	for _, task := range m.tasks {
		if task.Completed {
			res.completed++
		}

		switch {
		case task.Priority <= 2:
			res.high++
		case task.Priority == 3:
		  res.med++
		default:
			res.low++
		}

		if task.DueDate != nil && !task.Completed && task.DueDate.Before(now) {
			res.overdue++
		}

		if task.Completed && task.UpdatedAt != nil {
			daysAgo := daysBetween(*task.UpdatedAt, now)

			if daysAgo >= 0 && daysAgo < 14 {
				res.daily[13 - daysAgo]++
			}
		}
	}

	res.recent = append(res.recent, m.tasks...)

	sort.
		Slice(res.recent, func (i, j int) bool {
			return activityTime(res.recent[i]).
				After(activityTime(res.recent[j]))
		})

	if len(res.recent) > 3 {
		res.recent = res.recent[:3]
	}

	return res
}

func activityTime(t model.Todo) time.Time {
	if t.UpdatedAt != nil {
		return *t.UpdatedAt
	}

	if t.CreatedAt != nil {
		return *t.CreatedAt
	}

	return time.Time{}
}

func daysBetween(a, b time.Time) int {
	a = a.Truncate(24 * time.Hour)
	b = b.Truncate(24 * time.Hour)

	return int(b.Sub(a) / 24 * time.Hour)
}

func renderStatCard(title, value, secondary string) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	valueStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("252")).
		MarginTop(1)
	
	secondaryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("254"))

	valueLine := valueStyle.Render(value)

	if secondary != "" {
		valueLine += " " + secondaryStyle.Render(secondary)
	}

	return lipgloss.NewStyle().
		Width(24).
		Height(5).
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		MarginRight(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				titleStyle.Render(title),
				valueLine,
			),
		)
}

func percentage(value, total int) int {
	if total == 0 {
		return 0
	}

	return (value * 100) / total
}

func renderCompletionGraph(values []int) string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Render("completed / day — last 14 days")

	maxValue := 0

	for _, value := range values {
		if value > maxValue {
			maxValue = value
		}
	}

	const barHeight = 6

	var lines []string

	for row := barHeight; row > 0; row-- {
		var line strings.Builder

		for _, value := range values {
			level := 0

			if maxValue > 0 {
				level = (value * barHeight) / maxValue
			}

			if level >= row {
				line.WriteString("███")
			} else {
				line.WriteString("  ")
			}
		}

		lines = append(lines, line.String())
	}

	graph := strings.Join(lines, "\n")

	return lipgloss.NewStyle().
		Width(70).
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				graph,
			),
		)
}

func renderPriority(high, med, low int) string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Render("by priority")

	maxValue := max(high, med, low)

	rows := []string{
		priorityRow("high", high, maxValue),
		priorityRow("med", med, maxValue),
		priorityRow("low", low, maxValue),
	}

	return lipgloss.NewStyle().
		Width(40).
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				strings.Join(rows, "\n"),
			),
		)
}

func priorityRow(label string, value, maxValue int) string {
	const width = 12

	filled := 0

	if maxValue > 0 {
		filled = (value * width) / maxValue
	}

	empty := width - filled

	bar := strings.Repeat("█", filled) +
		strings.Repeat("░", empty)

	return fmt.Sprintf(
		"%-5s %s %d",
		label,
		bar,
		value,
	)
}

func max(values ...int) int {
	result := 0

	for _, value := range values {
		if value > result {
			result = value
		}
	}

	return result
}

func renderActivity(tasks []model.Todo) string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Render("recent activity")

	var rows []string

	for _, task := range tasks {
		t := activityTime(task)

		action := "new"

		if task.Completed {
			action = "done"
		}

		rows = append(
			rows,
			fmt.Sprintf(
				"%s  %s  %s",
				t.Format("15:04"),
				action,
				task.Description,
			),
		)
	}

	return lipgloss.NewStyle().
		Width(45).
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				strings.Join(rows, "\n"),
			),
		)
}
