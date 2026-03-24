package tui

import (
	"fmt"

	"github.com/Okwonks/go-todo/internal/client"
	tea "github.com/charmbracelet/bubbletea"
)

func Start() error {
	c := client.NewClient("http://localhost:8080")
	m := InitRoot(c)

	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()

	fmt.Print("\033[2J\033[H")

	return err
}
