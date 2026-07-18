package tui

import (
	"time"

	"github.com/bikky/oc-monitor/internal/db"
	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

type updateMsg struct {
	sessions    []db.Session
	live        []db.Session
	agents      []db.AgentRow
	projects    []db.ProjectRow
	daily       []db.PeriodRow
	hourly      []db.PeriodRow
	activity    map[string]db.Activity
	todayCost   float64
	todayTokens int64
	burn        []float64
	err         string
}

type detailMsg struct {
	session     *db.Session
	parts       []db.Part
	inputTokens int64
	err         string
}

type tokenMsg struct {
	data map[string]int64
}

type registryLoadedMsg struct{}
