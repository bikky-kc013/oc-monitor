package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/bikky/oc-monitor/internal/db"
)

func sessRows(ss []db.Session, d *db.DB, inputTokens map[string]int64, flushIDs map[string]bool, ctxColW int) []table.Row {
	out := make([]table.Row, 0, len(ss))
	for _, s := range ss {
		title := s.Title
		if len(title) > 24 {
			title = title[:21] + "..."
		}
		agent := s.Agent
		if agent == "" {
			agent = "—"
		}
		model := s.ModelID
		if len(model) > 14 {
			model = model[:11] + "..."
		}
		maxTok := d.MaxInputTokens(s.ModelID)
		inTok := inputTokens[s.ID]
		pct := 0.0
		if maxTok > 0 {
			pct = float64(inTok) / float64(maxTok) * 100
		}
		id := s.ID
		if len(id) > 10 {
			id = id[:10]
		}

		// Flash newly appeared sessions with a dot prefix
		idCell := id
		if flushIDs != nil && flushIDs[s.ID] {
			idCell = "●" + id[:min(len(id), 9)]
		}

		out = append(out, table.Row{
			idCell, title, agent, model,
			sCost.Render(fmt.Sprintf("$%.4f", s.Cost)),
			sTok.Render(fmt.Sprintf("%dk", (s.TokensIn+s.TokensOut)/1000)),
			ctxBar(pct, ctxColW),
			timeAgo(s.Updated),
		})
	}
	return out
}

func agentRows(aa []db.AgentRow) []table.Row {
	out := make([]table.Row, 0, len(aa))
	for _, a := range aa {
		m := a.Model
		if len(m) > 20 {
			m = m[:17] + "..."
		}
		out = append(out, table.Row{
			a.Agent, m, fmt.Sprintf("%d", a.Count),
			sCost.Render(fmt.Sprintf("$%.4f", a.Cost)),
			sTok.Render(fmt.Sprintf("%dk", a.Tokens/1000)),
			timeAgo(a.LastActive),
		})
	}
	return out
}

func projRows(pp []db.ProjectRow) []table.Row {
	out := make([]table.Row, 0, len(pp))
	for _, p := range pp {
		out = append(out, table.Row{
			p.Dir, fmt.Sprintf("%d", p.Count),
			sCost.Render(fmt.Sprintf("$%.4f", p.Cost)),
			sTok.Render(fmt.Sprintf("%dk", p.Tokens/1000)),
			p.Agent, p.Model, timeAgo(p.LastActive),
		})
	}
	return out
}

func periodRows(pp []db.PeriodRow, barWidth int) []table.Row {
	if barWidth < 6 {
		barWidth = 6
	}
	maxCost := 0.0
	for _, p := range pp {
		if p.Cost > maxCost {
			maxCost = p.Cost
		}
	}
	out := make([]table.Row, 0, len(pp))
	for _, p := range pp {
		bar := costBar(p.Cost, maxCost, barWidth)
		out = append(out, table.Row{
			p.Label,
			sCost.Render(fmt.Sprintf("$%.4f", p.Cost)),
			sTok.Render(fmt.Sprintf("%dk", p.Tokens/1000)),
			bar,
		})
	}
	return out
}

func searchRows(rr []db.SearchResult) []table.Row {
	out := make([]table.Row, 0, len(rr))
	for _, r := range rr {
		title := r.SessionTitle
		if len(title) > 24 {
			title = title[:21] + "..."
		}
		text := r.Text
		if len(text) > 48 {
			text = text[:45] + "..."
		}
		out = append(out, table.Row{title, r.PartType, text})
	}
	return out
}

func sortByCtx(ss []db.Session, d *db.DB, inputTokens map[string]int64) {
	pct := func(s db.Session) float64 {
		mx := d.MaxInputTokens(s.ModelID)
		if mx == 0 {
			return 0
		}
		return float64(inputTokens[s.ID]) / float64(mx)
	}
	for i := 0; i < len(ss); i++ {
		for j := i + 1; j < len(ss); j++ {
			if pct(ss[i]) < pct(ss[j]) {
				ss[i], ss[j] = ss[j], ss[i]
			}
		}
	}
}
