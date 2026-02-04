package tui

import (
	"github.com/Okwonks/go-todo/internal/client"
	tea "github.com/charmbracelet/bubbletea"
)

func Start() error {
	c := client.NewClient("http://localhost:8080")
	m := InitRoot(c)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		return err
	}
	return nil
}
