package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/bikky/oc-monitor/internal/db"
	"github.com/charmbracelet/lipgloss"
)

// ─── Activity status ─────────────────────────────────────────────────────────

func activityStatus(a db.Activity) (string, lipgloss.Style) {
	if a.SessionID == "" {
		return "idle", sDim
	}
	switch a.Type {
	case "tool":
		tool := a.Tool
		if tool == "" {
			tool = "tool"
		}
		if a.HasOutput {
			return "✓ ran " + tool, sOK
		}
		elapsed := time.Since(a.Created).Round(time.Second)
		return fmt.Sprintf("⚙ running %s (%s)", tool, elapsed), sWarn
	case "text":
		return "writing response…", sCyan()
	case "reasoning":
		return "thinking…", sDim
	default:
		return "idle", sDim
	}
}

func sCyan() lipgloss.Style { return lipgloss.NewStyle().Foreground(cToken) }

// ─── Time formatting ─────────────────────────────────────────────────────────

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return t.Format("Jan 02")
	}
}

// ─── Number formatting ───────────────────────────────────────────────────────

func fmtNumber(n int64) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

// ─── Context / usage bars ────────────────────────────────────────────────────

func ctxBar(pct float64, totalW int) string {
	// Reserve 5 chars for " 0%" percentage text
	pctW := 5
	barW := totalW - pctW - 1
	if barW < 1 {
		barW = 1
	}
	filled := int(pct / 100 * float64(barW))
	if filled > barW {
		filled = barW
	}
	if filled < 0 {
		filled = 0
	}
	empty := barW - filled
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	pctStr := fmt.Sprintf(" %3.0f%%", pct)
	style := sBar
	switch {
	case pct > 85:
		style = sBarC
	case pct > 60:
		style = sBarW
	}
	return style.Render(bar) + sDim.Render(pctStr)
}

func costBar(val, max float64, totalW int) string {
	if max <= 0 {
		return ""
	}
	pctW := 5
	barW := totalW - pctW - 1
	if barW < 1 {
		barW = 1
	}
	filled := int(val / max * float64(barW))
	if filled > barW {
		filled = barW
	}
	if filled < 1 && val > 0 {
		filled = 1
	}
	bar := sBar.Render(strings.Repeat("█", filled) + strings.Repeat("░", barW-filled))
	pct := sDim.Render(fmt.Sprintf(" %3.0f%%", val/max*100))
	return bar + pct
}

// ─── Sparkline ───────────────────────────────────────────────────────────────

func sparkline(vals []float64, w int) string {
	blocks := []rune("▁▂▃▄▅▆▇█")
	min, max := vals[0], vals[0]
	for _, v := range vals[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	r := max - min
	if r == 0 {
		r = 1
	}
	var out []rune
	step := 1
	if len(vals) > w {
		step = len(vals) / w
	}
	for i := 0; i < len(vals) && len(out) < w; i += step {
		norm := (vals[i] - min) / r
		idx := int(norm * 7)
		if idx > 7 {
			idx = 7
		}
		out = append(out, blocks[idx])
	}
	return sWarn.Render(string(out))
}

// ─── Loading / empty states ──────────────────────────────────────────────────

// loadingSpinner returns an animated spinner string for the given tick.
func loadingSpinner(tick int) string {
	frame := spinFrames[tick%len(spinFrames)]
	return sDim.Render(frame + " loading…")
}

// emptyState returns a styled "no data" message with an optional suggestion.
func emptyState(msg string) string {
	if msg == "" {
		msg = "no data"
	}
	return sDim.Render(strings.Repeat(" ", 8) + "─ " + msg + " ─")
}

// emptyStateAnimated returns a spinning loading animation for use during initial load.
func emptyStateAnimated(tick int) string {
	frame := spinFrames[tick%len(spinFrames)]
	return sDim.Render(strings.Repeat(" ", 4) + frame + " waiting for data…")
}
