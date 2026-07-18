package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bikky/oc-monitor/internal/db"
	"github.com/charmbracelet/lipgloss"
)

// ─── Main view dispatcher ────────────────────────────────────────────────────

func (m *model) View() string {
	w := m.contentWidth()

	// Check if data has loaded yet
	if m.tblRows == 0 && len(m.sessions) == 0 && len(m.live) == 0 && m.loadErr == "" {
		return m.viewLoading(w)
	}

	var b strings.Builder
	b.WriteString(m.renderHeader(w))
	b.WriteString("\n")
	b.WriteString(m.renderTabs())
	b.WriteString("\n")

	if m.loadErr != "" {
		b.WriteString(titledBoxDouble("error", []string{sErr.Render(m.loadErr)}, w, cErr))
		b.WriteString("\n")
	} else if m.view == vDetail {
		b.WriteString(titledBoxDouble(m.detailTitle(), strings.Split(m.renderDetail(), "\n"), w, cAccent))
	} else if m.view == vLive && w >= 140 {
		// Two-column layout for wide terminals
		b.WriteString(m.renderLiveWide(w))
	} else if (m.view == vDaily || m.view == vHourly) && w >= 100 {
		// Graph + table layout for period views
		b.WriteString(m.renderPeriodView(w))
	} else if m.view == vAgents || m.view == vModels {
		// Agents/Models with cost breakdown chart
		b.WriteString(m.renderAgentsView(w))
	} else {
		// Standard single-panel layout
		b.WriteString(titledBox(viewLabels[m.view], strings.Split(m.tbl.View(), "\n"), w, cBorder))
	}

	if m.searching {
		b.WriteString("\n  " + sBold.Render("search:") + " " + m.query + "█")
	}

	b.WriteString("\n" + m.renderHelp())
	return b.String()
}

// ─── Loading view (before data arrives) ─────────────────────────────────────

func (m *model) viewLoading(w int) string {
	var b strings.Builder
	b.WriteString(m.renderHeader(w))
	b.WriteString("\n")
	b.WriteString(m.renderTabs())
	b.WriteString("\n")

	lines := []string{
		"",
		emptyStateAnimated(m.tick),
		"",
	}
	b.WriteString(titledBox("", lines, w, cBorder))
	b.WriteString("\n")
	b.WriteString("\n" + m.renderHelp())
	return b.String()
}

// ─── Wide Live view (two-column) ────────────────────────────────────────────

func (m *model) renderLiveWide(w int) string {
	leftW := w*45/100
	rightW := w - leftW - 3 // gap

	// Left: Now panel + burn chart
	var leftLines []string
	leftLines = append(leftLines, sBold.Render("⚡ Active Agents"))
	leftLines = append(leftLines, "")

	if len(m.live) == 0 {
		leftLines = append(leftLines, sDim.Render("  no active agents"))
	} else {
		limit := len(m.live)
		if limit > 4 {
			limit = 4
		}
		for _, s := range m.live[:limit] {
			agent := s.Agent
			if agent == "" {
				agent = "agent"
			}
			model := s.ModelID
			if model == "" {
				model = "—"
			}
			title := s.Title
			if len(title) > 20 {
				title = title[:17] + "..."
			}
			status, statusStyle := activityStatus(m.activity[s.ID])

			leftLines = append(leftLines, fmt.Sprintf("  %s %s  %s",
				sDot.Render("●"), sBold.Render(agent), statusStyle.Render(status)))
			leftLines = append(leftLines, fmt.Sprintf("     %s  %s  %s",
				sDim.Render(model), sDim.Render(title),
				sCost.Render(fmt.Sprintf("$%.4f", s.Cost))))
		}
	}

	leftLines = append(leftLines, "")
	leftLines = append(leftLines, sBold.Render("🔥 Burn Rate"))
	leftLines = append(leftLines, "")
	if len(m.burn) > 0 {
		leftInnerW := leftW - 4
		if leftInnerW < 10 {
			leftInnerW = 10
		}
		chart := lineChart(m.burn, leftInnerW, 6)
		for _, l := range strings.Split(chart, "\n") {
			leftLines = append(leftLines, "  "+padRight(l, leftInnerW))
		}
	} else {
		leftLines = append(leftLines, sDim.Render("  no data"))
	}

	left := titledBoxDouble("now", leftLines, leftW, cOK)

	// Right: session table
	rightContent := strings.Split(m.tbl.View(), "\n")
	right := titledBox(viewLabels[m.view], rightContent, rightW, cBorder)

	// Join side by side
	return lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
}

// ─── Period view (graph + table) ────────────────────────────────────────────

func (m *model) renderPeriodView(w int) string {
	var b strings.Builder

	var data []float64
	var labels []string
	var periods []db.PeriodRow

	switch m.view {
	case vDaily:
		periods = m.daily
	case vHourly:
		periods = m.hourly
	}

	for _, p := range periods {
		data = append(data, p.Cost)
		labels = append(labels, p.Label)
	}

	// Available width for content lines inside a titledBox(w):
	// box uses 4 chars (border + padding), so inner content = w - 4
	innerW := w - 4
	if innerW < 10 {
		innerW = 10
	}

	if len(data) > 0 {
		chart := lineChart(data, innerW, 8)
		chartLines := strings.Split(chart, "\n")

		// Summary line — pad to innerW so it aligns with chart lines
		totalCost := 0.0
		maxCost := 0.0
		for _, v := range data {
			totalCost += v
			if v > maxCost {
				maxCost = v
			}
		}
		// Build summary line with proper width
		summaryParts := []seg{
			{text: "total: " + fmtCost(totalCost), style: sCost},
			{text: "   peak: " + fmtCost(maxCost), style: sWarn},
			{text: "   avg: " + fmtCost(totalCost/float64(max(len(data), 1))), style: sDim},
		}
		summary := buildLine(summaryParts, innerW)
		chartLines = append([]string{summary, ""}, chartLines...)

		b.WriteString(titledBox("trend", chartLines, w, cChart))
		b.WriteString("\n")
	}

	if m.view == vDaily && len(labels) >= 7 {
		heatW := innerW
		if heatW > 60 {
			heatW = 60
		}
		heatmap := calendarHeatmap(labels, data, heatW)
		if heatmap != "" {
			heatLines := strings.Split(heatmap, "\n")
			b.WriteString(titledBox("activity", heatLines, w, cBorder))
			b.WriteString("\n")
		}
	}

	tableLines := strings.Split(m.tbl.View(), "\n")
	b.WriteString(titledBox(viewLabels[m.view], tableLines, w, cBorder))

	return b.String()
}

// ─── Agents/Models view with cost breakdown ──────────────────────────────────

func (m *model) renderAgentsView(w int) string {
	var b strings.Builder

	innerW := w - 4
	if innerW < 10 {
		innerW = 10
	}

	// Build cost breakdown data from agent rows
	if len(m.agents) > 0 {
		items := make([]barItem, 0, min(len(m.agents), 10))
		for _, a := range m.agents {
			label := a.Agent
			if label == "" {
				label = "unknown"
			}
			items = append(items, barItem{Label: label, Value: a.Cost})
		}

		chart := barChart(items, innerW)
		if chart != "" {
			chartLines := strings.Split(chart, "\n")
			b.WriteString(titledBox("cost by agent", chartLines, w, cBorder))
			b.WriteString("\n")
		}
	}

	// Table
	b.WriteString(titledBox(viewLabels[m.view], strings.Split(m.tbl.View(), "\n"), w, cBorder))

	return b.String()
}

// ─── Header ──────────────────────────────────────────────────────────────────

func (m *model) renderHeader(w int) string {
	spin := spinFrames[m.tick%len(spinFrames)]
	dbName := m.d.Path
	if home, err := os.UserHomeDir(); err == nil {
		dbName = strings.Replace(dbName, home, "~", 1)
	}
	line1 := fmt.Sprintf("%s %s   %s   %s",
		sTitle.Render("oc-monitor"),
		sDim.Render(dbName),
		sOK.Render(spin+" live"),
		sDim.Render(time.Now().Format("15:04:05")))

	activeCount := len(m.live)
	activeLabel := sDim.Render(fmt.Sprintf("%d active", activeCount))
	if activeCount > 0 {
		activeLabel = sOK.Render(fmt.Sprintf("● %d active", activeCount))
	}
	line2 := fmt.Sprintf("today  %s   %s   %d sessions   %s",
		sCost.Render(fmt.Sprintf("$%.4f", m.todayCost)),
		sTok.Render(fmtNumber(m.todayTokens)+" tok"),
		len(m.sessions),
		activeLabel)

	if len(m.burn) > 0 {
		line2 += "   burn " + sparkline(m.burn, 12)
	}

	return titledBoxDouble("", []string{line1, line2}, w, cAccent)
}

// ─── Tabs ────────────────────────────────────────────────────────────────────

func (m *model) renderTabs() string {
	var parts []string
	for i, name := range viewLabels {
		if i == m.view {
			parts = append(parts, sTabSel.Render(name))
		} else {
			parts = append(parts, sTab.Render(name))
		}
	}
	return " " + strings.Join(parts, "")
}

// ─── Help ────────────────────────────────────────────────────────────────────

func (m *model) renderHelp() string {
	if m.view == vDetail {
		return sHelp.Render("  j/k:scroll  g/G:top/bottom  pgup/pgdn  esc:back  tab:switch  q:quit")
	}
	return sHelp.Render("  ←/→/tab:switch  j/k:nav  enter:open  /:search  x:sort-ctx  e:export  r:refresh  a:live  q:quit")
}

// ─── Detail view ─────────────────────────────────────────────────────────────

func (m *model) detailTitle() string {
	if m.detail == nil {
		return "detail"
	}
	t := m.detail.Title
	if len(t) > 50 {
		t = t[:47] + "..."
	}
	return t
}

func (m *model) renderDetail() string {
	if m.detail == nil {
		return sDim.Render("no session selected")
	}
	s := m.detail
	maxTok := m.detailMaxTokens
	inTok := m.detailInputTokens

	var lines []string
	lines = append(lines, fmt.Sprintf("%s   %s", sDim.Render(s.Agent+" / "+s.ModelID), sDim.Render(s.Directory)))

	if maxTok > 0 {
		pct := float64(inTok) / float64(maxTok) * 100
		lines = append(lines, fmt.Sprintf("%s  %s / %s tokens",
			ctxBar(pct, 20), fmtNumber(inTok), fmtNumber(int64(maxTok))))
	}

	lines = append(lines, fmt.Sprintf("Cost: %s   In: %s   Out: %s",
		sCost.Render(fmt.Sprintf("$%.4f", s.Cost)),
		sTok.Render(fmtNumber(s.TokensIn)),
		sTok.Render(fmtNumber(s.TokensOut))))
	lines = append(lines, "")

	limit := 200
	if len(m.parts) < limit {
		limit = len(m.parts)
	}
	for i := len(m.parts) - 1; i >= len(m.parts)-limit && i >= 0; i-- {
		p := m.parts[i]
		ts := sDim.Render(p.Created.Format("15:04:05"))
		switch p.Type {
		case "text":
			txt := p.Text
			if len(txt) > 180 {
				txt = txt[:180] + "..."
			}
			lines = append(lines, fmt.Sprintf("%s  %s", ts, txt))
		case "tool":
			icon := sWarn.Render("⚙")
			if p.HasOutput {
				icon = sOK.Render("✓")
			}
			tool := p.Tool
			if tool == "" {
				tool = "unknown"
			}
			lines = append(lines, fmt.Sprintf("%s  %s %s", ts, icon, tool))
		case "reasoning":
			txt := p.Text
			if len(txt) > 80 {
				txt = txt[:80] + "..."
			}
			lines = append(lines, fmt.Sprintf("%s  %s", ts, sDim.Render(txt)))
		case "step-start", "step-finish":
		default:
			lines = append(lines, fmt.Sprintf("%s  %s", ts, sDim.Render("["+p.Type+"]")))
		}
	}

	m.vp.SetContent(strings.Join(lines, "\n"))
	return m.vp.View()
}
