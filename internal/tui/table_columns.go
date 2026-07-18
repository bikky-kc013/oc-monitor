package tui

import (
	"github.com/charmbracelet/bubbles/table"
)

func budget(innerWidth, numCols int) int {
	b := innerWidth - numCols*2
	if b < numCols*4 {
		b = numCols * 4
	}
	return b
}

func colWidths(total int, weights []int, mins []int) []int {
	sum := 0
	for _, w := range weights {
		sum += w
	}
	if sum == 0 {
		sum = 1
	}
	out := make([]int, len(weights))
	used := 0
	for i, w := range weights {
		cw := total * w / sum
		if cw < mins[i] {
			cw = mins[i]
		}
		out[i] = cw
		used += cw
	}
	if n := len(out); n > 0 {
		diff := total - used
		out[n-1] += diff
		if out[n-1] < mins[n-1] {
			out[n-1] = mins[n-1]
		}
	}
	return out
}

func sessCols(innerWidth int) []table.Column {
	titles := []string{"ID", "Title", "Agent", "Model", "Cost", "Tokens", "Context", "Updated"}
	weights := []int{6, 20, 8, 12, 8, 6, 18, 6}
	mins := []int{6, 10, 6, 8, 7, 5, 15, 6}
	w := colWidths(budget(innerWidth, len(titles)), weights, mins)
	cols := make([]table.Column, len(titles))
	for i, t := range titles {
		cols[i] = table.Column{Title: t, Width: w[i]}
	}
	return cols
}

func agentCols(innerWidth int) []table.Column {
	titles := []string{"Agent", "Model", "Sessions", "Cost", "Tokens", "Last Active"}
	weights := []int{16, 26, 12, 14, 14, 18}
	mins := []int{8, 12, 8, 8, 7, 8}
	w := colWidths(budget(innerWidth, len(titles)), weights, mins)
	cols := make([]table.Column, len(titles))
	for i, t := range titles {
		cols[i] = table.Column{Title: t, Width: w[i]}
	}
	return cols
}

func projCols(innerWidth int) []table.Column {
	titles := []string{"Project", "Sessions", "Cost", "Tokens", "Agent", "Model", "Last Active"}
	weights := []int{20, 10, 12, 12, 12, 18, 16}
	mins := []int{8, 8, 8, 7, 6, 8, 8}
	w := colWidths(budget(innerWidth, len(titles)), weights, mins)
	cols := make([]table.Column, len(titles))
	for i, t := range titles {
		cols[i] = table.Column{Title: t, Width: w[i]}
	}
	return cols
}

func periodCols(innerWidth int) []table.Column {
	titles := []string{"Period", "Cost", "Tokens", "Bar"}
	weights := []int{22, 16, 16, 46}
	mins := []int{10, 8, 6, 12}
	w := colWidths(budget(innerWidth, len(titles)), weights, mins)
	cols := make([]table.Column, len(titles))
	for i, t := range titles {
		cols[i] = table.Column{Title: t, Width: w[i]}
	}
	return cols
}

func searchCols(innerWidth int) []table.Column {
	titles := []string{"Session", "Type", "Match"}
	weights := []int{28, 12, 60}
	mins := []int{10, 6, 16}
	w := colWidths(budget(innerWidth, len(titles)), weights, mins)
	cols := make([]table.Column, len(titles))
	for i, t := range titles {
		cols[i] = table.Column{Title: t, Width: w[i]}
	}
	return cols
}
