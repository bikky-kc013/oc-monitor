package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)


func fmtCost(v float64) string {
	if v >= 1.0 {
		return fmt.Sprintf("$%.2f", v)
	}
	if v >= 0.01 {
		return fmt.Sprintf("$%.3f", v)
	}
	return fmt.Sprintf("$%.4f", v)
}

func padRight(s string, visTarget int) string {
	w := lipgloss.Width(s)
	if w < visTarget {
		return s + strings.Repeat(" ", visTarget-w)
	}
	return s
}


type seg struct {
	text  string
	style lipgloss.Style
}

func buildLine(parts []seg, w int) string {
	total := 0
	for _, p := range parts {
		total += lipgloss.Width(p.text)
	}
	pad := 0
	if total < w {
		pad = w - total
	}
	var b strings.Builder
	for i, p := range parts {
		t := p.text
		if i == len(parts)-1 && pad > 0 {
			t += strings.Repeat(" ", pad)
		}
		b.WriteString(p.style.Render(t))
	}
	return b.String()
}


func lineChart(data []float64, w, h int) string {
	if len(data) == 0 || w < 8 || h < 3 {
		return ""
	}
	if h < 3 {
		h = 3
	}

	dMin, dMax := data[0], data[0]
	for _, v := range data[1:] {
		if v < dMin {
			dMin = v
		}
		if v > dMax {
			dMax = v
		}
	}
	margin := (dMax - dMin) * 0.05
	if margin == 0 {
		margin = dMax * 0.05
	}
	if margin == 0 {
		margin = 1
	}
	dMin -= margin
	dMax += margin
	rng := dMax - dMin

	chartW := w - 9
	if chartW < 4 {
		chartW = 4
	}

	points := make([]float64, chartW)
	if len(data) <= chartW {
		copy(points, data)
		for i := len(data); i < chartW; i++ {
			points[i] = data[len(data)-1]
		}
	} else {
		step := float64(len(data)-1) / float64(chartW-1)
		for i := 0; i < chartW; i++ {
			idx := int(float64(i) * step)
			if idx >= len(data) {
				idx = len(data) - 1
			}
			points[i] = data[idx]
		}
	}

	pxRow := make([]int, chartW)
	for i, v := range points {
		norm := (v - dMin) / rng
		row := int(norm * float64(h-1))
		if row < 0 {
			row = 0
		}
		if row >= h {
			row = h - 1
		}
		pxRow[i] = row
	}

	yVals := make([]string, h)
	for r := 0; r < h; r++ {
		var raw string
		switch {
		case r == h-1:
			raw = fmtCost(dMax)
		case r == 0:
			raw = fmtCost(dMin)
		case r == h/2:
			raw = fmtCost((dMin + dMax) / 2)
		default:
			raw = ""
		}
		yVals[r] = padRight(raw, 8)
	}

	var out []string
	for r := h - 1; r >= 0; r-- {
		cells := make([]rune, chartW)
		for c := 0; c < chartW; c++ {
			if pxRow[c] == r {
				cells[c] = '█'
			} else {
				cells[c] = ' '
			}
		}
		parts := []seg{
			{text: yVals[r], style: sDim},
			{text: " ", style: lipgloss.NewStyle()},
			{text: string(cells), style: sChart},
		}
		out = append(out, buildLine(parts, w))
	}

	return strings.Join(out, "\n")
}


type barItem struct {
	Label string
	Value float64
}

func barChart(items []barItem, w int) string {
	if len(items) == 0 || w < 12 {
		return ""
	}

	sorted := make([]barItem, len(items))
	copy(sorted, items)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Value > sorted[j].Value })

	if len(sorted) > 12 {
		sorted = sorted[:12]
	}

	maxVal := sorted[0].Value
	if maxVal <= 0 {
		maxVal = 1
	}

	maxLabel := 0
	for _, it := range sorted {
		if len(it.Label) > maxLabel {
			maxLabel = len(it.Label)
		}
	}
	if maxLabel > 16 {
		maxLabel = 16
	}

	valW := 14
	barW := w - 2 - maxLabel - 1 - 1 - valW
	if barW < 4 {
		barW = 4
	}

	total := 0.0
	for _, it := range sorted {
		total += it.Value
	}

	var out []string
	for _, it := range sorted {
		filled := int(it.Value / maxVal * float64(barW))
		if filled < 1 && it.Value > 0 {
			filled = 1
		}
		pct := it.Value / total * 100

		label := it.Label
		if len(label) > maxLabel {
			label = label[:maxLabel-1] + "…"
		}
		labelPad := padRight(label, maxLabel)

		bar := strings.Repeat("█", filled) + strings.Repeat("░", barW-filled)

		valStr := fmt.Sprintf(" %s %5.1f%%", fmtCost(it.Value), pct)

		parts := []seg{
			{text: "  ", style: lipgloss.NewStyle()},
			{text: labelPad, style: sDim},
			{text: " ", style: lipgloss.NewStyle()},
			{text: bar, style: sBar},
			{text: valStr, style: sDim},
		}
		out = append(out, buildLine(parts, w))
	}

	return strings.Join(out, "\n")
}


func calendarHeatmap(labels []string, values []float64, w int) string {
	if len(labels) == 0 || w < 14 {
		return ""
	}

	type dayVal struct {
		t   time.Time
		val float64
	}
	days := make([]dayVal, 0, len(labels))
	for i, l := range labels {
		t, err := time.Parse("2006-01-02", l)
		if err != nil {
			continue
		}
		v := 0.0
		if i < len(values) {
			v = values[i]
		}
		days = append(days, dayVal{t: t, val: v})
	}
	if len(days) == 0 {
		return ""
	}

	minV, maxV := days[0].val, days[0].val
	for _, d := range days[1:] {
		if d.val < minV {
			minV = d.val
		}
		if d.val > maxV {
			maxV = d.val
		}
	}

	valMap := make(map[string]float64)
	for _, d := range days {
		valMap[d.t.Format("2006-01-02")] = d.val
	}

	latest := days[len(days)-1].t
	startDate := latest.AddDate(0, 0, -7*19)
	for startDate.Weekday() != time.Monday {
		startDate = startDate.AddDate(0, 0, -1)
	}

	weekCount := int(latest.Sub(startDate).Hours()/24/7) + 1
	maxWeeks := (w - 4) / 3
	if maxWeeks < 1 {
		maxWeeks = 1
	}
	if weekCount > maxWeeks {
		weekCount = maxWeeks
	}

	colors := []lipgloss.Style{sDim, sBar, sBarW, sErr}

	grid := make([][]string, 7)
	for d := 0; d < 7; d++ {
		grid[d] = make([]string, weekCount)
		for ww := 0; ww < weekCount; ww++ {
			grid[d][ww] = "██"
		}
	}

	dayLabels := []string{"Mon", "Wed", "Fri", "Sun"}
	for ww := 0; ww < weekCount; ww++ {
		weekStart := startDate.AddDate(0, 0, ww*7)
		for d := 0; d < 7; d++ {
			date := weekStart.AddDate(0, 0, d)
			if date.After(latest) {
				continue
			}
			dateStr := date.Format("2006-01-02")
			val, ok := valMap[dateStr]
			if !ok {
				continue
			}
			level := 0
			if maxV > minV {
				norm := (val - minV) / (maxV - minV)
				if norm > 0.66 {
					level = 3
				} else if norm > 0.33 {
					level = 2
				} else if val > 0 {
					level = 1
				}
			} else if val > 0 {
				level = 1
			}
			grid[d][ww] = colors[level].Render("██")
		}
	}

	var out []string
	for d := 0; d < 7; d++ {
		var dayLabel string
		if d == 0 || d == 2 || d == 4 || d == 6 {
			dayLabel = dayLabels[d/2]
		} else {
			dayLabel = "   "
		}
		cells := ""
		for ww, cell := range grid[d] {
			if ww > 0 {
				cells += " "
			}
			cells += cell
		}
		row := dayLabel + "  " + cells
		out = append(out, padRight(row, w))
	}

	return strings.Join(out, "\n")
}
