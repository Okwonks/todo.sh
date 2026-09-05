package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Okwonks/go-todo/internal/model"
	"github.com/Okwonks/go-todo/internal/tui/constants"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	statsMint  = lipgloss.Color("#6fe3ac")
	statsCoral = lipgloss.Color("#e8899a")
	statsAmber = lipgloss.Color("#e8b96f")
	statsBlue  = lipgloss.Color("#6fa8e8")
	statsCyan  = lipgloss.Color("#6fd4e8")
	statsMuted = lipgloss.Color("240")
	statsFg    = lipgloss.Color("255")

	statsCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(statsMuted).
			Padding(1, 2)

	statsLabel = lipgloss.NewStyle().Foreground(statsMuted)
)

// Shared dashboard sizing. Every row targets the same outer width
// (72 = three 22-wide stat cards + their borders) so the chart and
// bottom row line up with the top row instead of shrink-wrapping to
// their own content.
const (
	statCardWidth     = 22 // outer 24
	chartWidth        = 70 // outer 72
	priorityCardWidth = 32 // outer 34
	activityCardWidth = 34 // outer 36; 34+36+gap(2) = 72
	bottomCardHeight  = 5  // label + blank + 3 data rows, forced equal
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

	style := lipgloss.NewStyle().Bold(true).Margin(1)
	title := style.Render("Todo.sh | [stats]")

	topRow := lipgloss.
		JoinHorizontal(
			lipgloss.Top,
			renderStatCard("total tasks", fmt.Sprintf("%d", data.total), statsFg, ""),
			renderStatCard("completed", fmt.Sprintf("%d", data.completed), statsMint, fmt.Sprintf("%d%%", percentage(data.completed, data.total))),
			renderStatCard("overdue", fmt.Sprintf("%d", data.overdue), statsCoral, ""),
		)

	chart := statsCard.
	  Width(chartWidth).
	  Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				statsLabel.Render("completed / day - last 14 days"),
				"",
				renderCompletionGraph(data.daily, 6, statsMint),
			),
		)

	high, med, low := data.high, data.med, data.low
	maxPriority := high
	if med > maxPriority {
		maxPriority = med
	}
	if low > maxPriority {
		maxPriority = low
	}

	priorityCard := statsCard.
	  Width(priorityCardWidth).
		Height(bottomCardHeight).
	  Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				statsLabel.Render("by priority"),
				"",
				priorityRow("high", high, maxPriority, statsCoral),
				priorityRow("med", med, maxPriority, statsAmber),
				priorityRow("low", low, maxPriority, statsBlue),
			),
		)

	activityCard := statsCard.
	  Width(activityCardWidth).
		Height(bottomCardHeight).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				statsLabel.Render("recent activity"),
				"",
				renderActivity(data.recent, activityCardWidth),
			),
		)

	bottom := lipgloss.JoinHorizontal(lipgloss.Top, priorityCard, "  ", activityCard)

	help := constants.HelpStyle("[q] quit • [r] refresh • [tab] back to list")

	return lipgloss.NewStyle().
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				topRow,
				chart,
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

	return int(b.Sub(a) / (24 * time.Hour))
}

func renderStatCard(label, value string, color lipgloss.Color, suffix string) string {
	valueLine := lipgloss.NewStyle().Bold(true).Foreground(color).Render(value)
	if suffix != "" {
		valueLine += " " + statsLabel.Render(suffix)
	}

	return statsCard.
	  Width(statCardWidth).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				statsLabel.Render(label),
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

func renderCompletionGraph(values []int, maxHeight int, color lipgloss.Color) string {
	if len(values) == 0 {
		return ""
	}

	max := max(values...)

	barStyle := lipgloss.NewStyle().Foreground(color)
	cols := make([]string, len(values))
	for i, v := range values {
		h := int(float64(maxHeight) * float64(v) / float64(max))
		if v > 0 && h == 0 {
			h = 1
		}
		var b strings.Builder
		for row := maxHeight; row > 0; row-- {
			if row <= h {
				b.WriteString(barStyle.Render("███"))
			} else {
				b.WriteString("  ")
			}
			b.WriteString("\n")
		}
		cols[i] = strings.TrimRight(b.String(), "\n")
	}
	return lipgloss.JoinHorizontal(lipgloss.Bottom, cols...)
}

func priorityRow(label string, count, maxValue int, color lipgloss.Color) string {
	width := 18
	if maxValue == 0 {
		maxValue = 1
	}
	filled := int(float64(width) * float64(count) / float64(maxValue))
	if filled > width {
		filled = width
	}
	bar := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)) +
	  lipgloss.NewStyle().Foreground(statsMuted).Render(strings.Repeat("░", width - filled))
	return fmt.Sprintf("%-5s %s %d", label, bar, count)
}

func max(values ...int) int {
	result := 1

	for _, value := range values {
		if value > result {
			result = value
		}
	}

	return result
}

func renderActivity(tasks []model.Todo, cardWidth int) string {
	if len(tasks) == 0 {
		return statsLabel.Render("no recent activity")
	}

	const (
		timeWidth   = 6 // "15:04 "
		statusWidth = 8 // longest status ("started"/"blocked"/"pending") + 1 space
		padding     = 4 // card's own left+right padding
	)

	titleBudget := cardWidth - padding - timeWidth - statusWidth;
	if titleBudget < 4 {
		titleBudget = 4
	}

	var lines []string
	for _, task := range tasks {
		t := activityTime(task)

		statusStyle := lipgloss.NewStyle().Foreground(statsCyan)
		if task.Completed {
			statusStyle = lipgloss.NewStyle().Foreground(statsMint)
		}
		title := task.Description
		if len(title) > titleBudget {
			cut := titleBudget - 3
			if cut < 1 {
				cut = 1
			}
			title = title[:cut] + "..."
		}
		lines = append(lines, fmt.Sprintf("%s %s %s",
		  statsLabel.Render(t.Format("15:04")),
			statusStyle.Render(string(task.Status)),
			title,
		))
	}
	return strings.Join(lines, "\n")
}
