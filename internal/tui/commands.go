package tui

import (
	"time"

	"github.com/bikky/oc-monitor/internal/db"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) loadAll() tea.Cmd {
	return func() tea.Msg {
		d := m.d
		sessions, _ := d.Sessions(500)
		live, lerr := d.ActiveSessions(15 * time.Second)
		agents, _ := d.AgentStats()
		projects, _ := d.ProjectRollups()
		daily, _ := d.DailySpend(30)
		hourly, _ := d.HourlySpend(48)

		var activity map[string]db.Activity
		if len(live) > 0 {
			ids := make([]string, len(live))
			for i, s := range live {
				ids[i] = s.ID
			}
			activity, _ = d.LatestActivity(ids)
		} else {
			activity = map[string]db.Activity{}
		}

		startOfDay := time.Now().Truncate(24 * time.Hour)
		todayCost := d.TotalSpendSince(startOfDay)
		todayTokens := int64(0)
		for _, p := range daily {
			if p.Label == time.Now().Format("2006-01-02") {
				todayTokens = p.Tokens
			}
		}

		s5m, _ := d.ActiveSessions(5 * time.Minute)
		now := time.Now()
		burn := make([]float64, 5)
		for i := 0; i < 5; i++ {
			start := now.Add(-time.Duration(i+1) * time.Minute)
			end := now.Add(-time.Duration(i) * time.Minute)
			var cost float64
			for _, s := range s5m {
				if !s.Updated.Before(start) && s.Updated.Before(end) {
					cost += s.Cost
				}
			}
			burn[4-i] = cost
		}

		return updateMsg{
			sessions:    sessions,
			live:        live,
			agents:      agents,
			projects:    projects,
			daily:       daily,
			hourly:      hourly,
			activity:    activity,
			todayCost:   todayCost,
			todayTokens: todayTokens,
			burn:        burn,
			err:         func() string { if lerr != nil { return lerr.Error() }; return "" }(),
		}
	}
}

func (m *model) fetchTokens() tea.Cmd {
	d := m.d
	ids := make([]string, len(m.sessions))
	for i, s := range m.sessions {
		ids[i] = s.ID
	}
	for _, s := range m.live {
		found := false
		for _, id := range ids {
			if id == s.ID {
				found = true
				break
			}
		}
		if !found {
			ids = append(ids, s.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return func() tea.Msg {
		return tokenMsg{data: d.BatchInputTokens(ids)}
	}
}

func (m *model) regFetchCmd() tea.Cmd {
	return func() tea.Msg {
		r := m.d.Registry()
		if r == nil {
			return registryLoadedMsg{}
		}
		_ = r.Fetch()
		return registryLoadedMsg{}
	}
}
