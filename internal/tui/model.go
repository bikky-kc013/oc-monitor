package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resize()
		cmds := []tea.Cmd{m.loadAll(), tickCmd()}
		if m.d.Registry() != nil && m.d.Registry().Stale() {
			cmds = append(cmds, m.regFetchCmd())
		}
		return m, tea.Batch(cmds...)

	case tickMsg:
		m.tick++
		return m, tickCmd()

	case updateMsg:
		m.sessions = msg.sessions
		m.live = msg.live
		m.agents = msg.agents
		m.projects = msg.projects
		m.daily = msg.daily
		m.hourly = msg.hourly
		m.activity = msg.activity
		m.todayCost = msg.todayCost
		m.todayTokens = msg.todayTokens
		m.burn = msg.burn
		m.loadErr = msg.err
		m.cursor = 0
		m.syncTable()
		return m, m.fetchTokens()

	case tokenMsg:
		m.inputTokens = msg.data
		return m, nil

	case detailMsg:
		if msg.err != "" {
			m.loadErr = msg.err
			return m, nil
		}
		m.detail = msg.session
		m.parts = msg.parts
		m.detailInputTokens = msg.inputTokens
		m.detailMaxTokens = m.d.MaxInputTokens(msg.session.ModelID)
		m.view = vDetail
		m.vp.GotoTop()
		return m, nil

	case registryLoadedMsg:
		return m, nil

	case tea.KeyMsg:
		return m.key(msg)
	}
	return m, nil
}

func (m *model) resize() {
	w := m.contentWidth()
	m.tblInnerW = w - 4
	if m.tblInnerW < 20 {
		m.tblInnerW = 20
	}
	m.tbl.SetWidth(m.tblInnerW)
	h := m.height - 14
	if h < 3 {
		h = 3
	}
	m.tbl.SetHeight(h)
	m.vp.Width = m.tblInnerW
	m.vp.Height = m.height - 16
	m.syncTable()
}

func (m *model) contentWidth() int {
	w := m.width - 2
	if w < 40 {
		w = 76
	}
	return w
}
