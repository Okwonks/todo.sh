package main

import (
	"fmt"
	"os"

	"github.com/Okwonks/go-todo/internal/client"
	"github.com/Okwonks/go-todo/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	c := client.NewClient("http://localhost:8080")
	m := tui.InitRoot(c)

	 _, err := tea.NewProgram(m, tea.WithAltScreen()).Run()

	if err != nil {
		fmt.Println("Error running TUI:", err)
		os.Exit(1)
	}

	fmt.Print("\033[2J\033[H")
}
