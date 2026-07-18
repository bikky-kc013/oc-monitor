package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) key(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.searching {
		return m.searchKey(msg)
	}
	if m.view == vDetail {
		return m.detailKey(msg)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "tab", "right":
		m.view = (m.view + 1) % len(viewLabels)
		m.cursor = 0
		m.syncTable()
	case "shift+tab", "left":
		m.view = (m.view - 1 + len(viewLabels)) % len(viewLabels)
		m.cursor = 0
		m.syncTable()
	case "j", "k", "down", "up":
		m.tbl, _ = m.tbl.Update(msg)
		m.cursor = m.tbl.Cursor()
	case "pgdown", "ctrl+f":
		m.tbl, _ = m.tbl.Update(msg)
		m.cursor = m.tbl.Cursor()
	case "pgup", "ctrl+b":
		m.tbl, _ = m.tbl.Update(msg)
		m.cursor = m.tbl.Cursor()
	case "g", "G":
		m.tbl, _ = m.tbl.Update(msg)
		m.cursor = m.tbl.Cursor()
	case "/":
		m.searching = true
		m.query = ""
	case "a":
		if len(m.live) > 0 {
			m.view = vLive
			m.cursor = 0
			m.syncTable()
		}
	case "x":
		m.sortCtx = !m.sortCtx
		m.syncTable()
	case "e":
		m.exportJSON()
	case "r":
		return m, m.loadAll()
	case "enter":
		if cmd := m.enterDetail(); cmd != nil {
			return m, cmd
		}
	}
	return m, nil
}

func (m *model) detailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "backspace":
		m.view = vSessions
		m.cursor = 0
		m.syncTable()
	case "j", "down":
		m.vp.LineDown(1)
	case "k", "up":
		m.vp.LineUp(1)
	case "pgdown", "ctrl+f":
		m.vp.HalfViewDown()
	case "pgup", "ctrl+b":
		m.vp.HalfViewUp()
	case "g":
		m.vp.GotoTop()
	case "G":
		m.vp.GotoBottom()
	case "tab":
		m.view = (m.view + 1) % len(viewLabels)
		m.cursor = 0
		m.syncTable()
	case "shift+tab":
		m.view = (m.view - 1 + len(viewLabels)) % len(viewLabels)
		m.cursor = 0
		m.syncTable()
	}
	return m, nil
}

func (m *model) searchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searching = false
	case "enter":
		m.searching = false
		if m.query != "" {
			m.results, _ = m.d.Search(m.query, 100)
			m.view = vSearch
			m.cursor = 0
			m.syncTable()
		}
	case "backspace":
		if len(m.query) > 0 {
			m.query = m.query[:len(m.query)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.query += msg.String()
		}
	}
	return m, nil
}

func (m *model) enterDetail() tea.Cmd {
	c := m.cursor
	d := m.d
	open := func(id string) tea.Cmd {
		return func() tea.Msg {
			s, err := d.GetSession(id)
			if err != nil {
				return detailMsg{err: err.Error()}
			}
			parts, err := d.Parts(s.ID, 500)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			inTok := d.LatestInputTokens(s.ID)
			return detailMsg{
				session:     s,
				parts:       parts,
				inputTokens: inTok,
				err:         errStr,
			}
		}
	}
	switch m.view {
	case vLive:
		if c >= 0 && c < len(m.live) {
			return open(m.live[c].ID)
		}
	case vSessions:
		if c >= 0 && c < len(m.sessions) {
			return open(m.sessions[c].ID)
		}
	case vProjects:
		if c >= 0 && c < len(m.projects) {
			p := m.projects[c]
			m.sessions, _ = m.d.SessionsByDir(p.FullPath, 200)
			m.view = vSessions
			m.cursor = 0
			m.syncTable()
		}
	case vSearch:
		if c >= 0 && c < len(m.results) {
			return open(m.results[c].SessionID)
		}
	}
	return nil
}


func (m *model) exportJSON() {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "oc-monitor-exports")
	os.MkdirAll(dir, 0o755)
	ts := time.Now().Format("20060102-150405")
	view := strings.ToLower(viewLabels[m.view])
	name := filepath.Join(dir, fmt.Sprintf("oc-monitor-%s-%s.json", view, ts))

	var data any
	switch m.view {
	case vLive:
		data = m.live
	case vSessions:
		data = m.sessions
	case vAgents, vModels:
		data = m.agents
	case vDaily:
		data = m.daily
	case vHourly:
		data = m.hourly
	case vProjects:
		data = m.projects
	case vSearch:
		data = m.results
	default:
		data = m.sessions
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(name, b, 0o644)
}

func (m *model) maxCur() int {
	switch m.view {
	case vLive:
		return len(m.live) - 1
	case vSessions:
		return len(m.sessions) - 1
	case vAgents, vModels:
		return len(m.agents) - 1
	case vDaily:
		return len(m.daily) - 1
	case vHourly:
		return len(m.hourly) - 1
	case vProjects:
		return len(m.projects) - 1
	case vSearch:
		return len(m.results) - 1
	}
	return 0
}
