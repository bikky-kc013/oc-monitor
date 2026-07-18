package main

import (
	"fmt"
	"os"

	"github.com/bikky/oc-monitor/internal/db"
	"github.com/bikky/oc-monitor/internal/registry"
	"github.com/bikky/oc-monitor/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if os.Getenv("DEBUG") != "" {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
	}

	database, err := db.Open()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer database.Close()

	reg := registry.NewClient()
	_ = reg.LoadFromDisk()
	database.SetRegistry(reg)

	m := tui.New(database)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
