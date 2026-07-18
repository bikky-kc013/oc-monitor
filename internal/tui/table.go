package tui

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/bikky/oc-monitor/internal/db"
)

type model struct {
	d *db.DB

	width, height int
	view          int
	cursor        int

	sessions  []db.Session
	live      []db.Session
	agents    []db.AgentRow
	projects  []db.ProjectRow
	daily     []db.PeriodRow
	hourly    []db.PeriodRow
	results   []db.SearchResult
	activity  map[string]db.Activity
	query     string
	searching bool
	sortCtx   bool

	todayCost     float64
	todayTokens   int64
	loadErr       string
	inputTokens   map[string]int64

	detail            *db.Session
	parts             []db.Part
	detailInputTokens int64
	detailMaxTokens   int

	vp        viewport.Model
	tbl       table.Model
	prevView  int
	tblCols   int
	tblInnerW int
	tblRows   int

	burn        []float64
	tick        int
	prevLiveIDs map[string]bool
	flushIDs    map[string]bool // sessions that just appeared in live — render with highlight
}

func New(d *db.DB) *model {
	tbl := table.New(
		table.WithHeight(12),
		table.WithFocused(true),
		table.WithStyles(table.Styles{
			Header:   lipgloss.NewStyle().Foreground(cToken).Bold(true).Padding(0, 1),
			Cell:     lipgloss.NewStyle().Padding(0, 1).Foreground(cText),
			Selected: lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(cAccent).Bold(true).Padding(0, 1),
		}),
	)
	vp := viewport.New(80, 10)
	return &model{d: d, tbl: tbl, vp: vp, activity: map[string]db.Activity{}, prevLiveIDs: map[string]bool{}}
}

func (m *model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m *model) syncTable() {
	max := m.maxCur()
	if m.cursor > max {
		m.cursor = max
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	w := m.tblInnerW
	if w <= 0 {
		w = 76
	}

	var cols []table.Column
	var rows []table.Row

	switch m.view {
	case vLive:
		// Detect newly appeared sessions for flash highlight
		newFlush := map[string]bool{}
		for _, s := range m.live {
			if !m.prevLiveIDs[s.ID] {
				newFlush[s.ID] = true
			}
		}
		m.flushIDs = newFlush
		cols = sessCols(w)
		rows = sessRows(m.live, m.d, m.inputTokens, m.flushIDs, cols[6].Width)
	case vSessions:
		s := m.sessions
		if m.sortCtx {
			sortByCtx(s, m.d, m.inputTokens)
		}
		cols = sessCols(w)
		rows = sessRows(s, m.d, m.inputTokens, nil, cols[6].Width)
	case vAgents, vModels:
		cols = agentCols(w)
		rows = agentRows(m.agents)
	case vDaily:
		cols = periodCols(w)
		rows = periodRows(m.daily, cols[3].Width)
	case vHourly:
		cols = periodCols(w)
		rows = periodRows(m.hourly, cols[3].Width)
	case vProjects:
		cols = projCols(w)
		rows = projRows(m.projects)
	case vSearch:
		cols = searchCols(w)
		rows = searchRows(m.results)
	}

	// Rebuild table if view or width changed
	viewCols := len(cols)
	if m.tblCols != w || m.prevView != int(m.view) || len(m.tbl.Columns()) != viewCols {
		m.prevView = int(m.view)
		m.tbl = table.New(
			table.WithHeight(12),
			table.WithFocused(true),
			table.WithStyles(table.Styles{
				Header:   lipgloss.NewStyle().Foreground(cToken).Bold(true).Padding(0, 1),
				Cell:     lipgloss.NewStyle().Padding(0, 1).Foreground(cText),
				Selected: lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(cAccent).Bold(true).Padding(0, 1),
			}),
		)
		m.tbl.SetCursor(m.cursor)
		m.tbl.SetColumns(cols)
		m.tblCols = w
	}

	m.tbl.SetRows(rows)
	m.tblRows = len(rows)
	if m.cursor > len(rows)-1 {
		m.cursor = 0
	}

	// Update prevLiveIDs for next tick's flash detection
	m.prevLiveIDs = map[string]bool{}
	for _, s := range m.live {
		m.prevLiveIDs[s.ID] = true
	}
}
