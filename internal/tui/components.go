package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// truncateVis truncates a string to visMax visible characters,
// stripping trailing ANSI sequences if needed.
func truncateVis(s string, visMax int) string {
	w := lipgloss.Width(s)
	if w <= visMax {
		return s
	}
	runes := []rune(s)
	for i := len(runes); i > 0; i-- {
		if lipgloss.Width(string(runes[:i])) <= visMax {
			return string(runes[:i])
		}
	}
	return ""
}

// ─── Box rendering ───────────────────────────────────────────────────────────

// titledBox renders a bordered box with an optional title.
// borderColor determines the visual weight: use cBorderHi for primary panels,
// cBorder for secondary.
func titledBox(title string, lines []string, width int, borderColor lipgloss.Color) string {
	var b strings.Builder
	b.WriteString(boxTop(title, width, borderColor))
	b.WriteString("\n")
	for _, l := range lines {
		b.WriteString(boxLine(l, width, borderColor))
		b.WriteString("\n")
	}
	b.WriteString(boxBottom(width, borderColor))
	return b.String()
}

// titledBoxDouble renders a double-line bordered box for primary/focused panels.
func titledBoxDouble(title string, lines []string, width int, borderColor lipgloss.Color) string {
	var b strings.Builder
	b.WriteString(boxTopDouble(title, width, borderColor))
	b.WriteString("\n")
	for _, l := range lines {
		b.WriteString(boxLineDouble(l, width, borderColor))
		b.WriteString("\n")
	}
	b.WriteString(boxBottomDouble(width, borderColor))
	return b.String()
}

// ─── Single-line borders (secondary panels) ──────────────────────────────────

func boxTop(title string, width int, c lipgloss.Color) string {
	bs := lipgloss.NewStyle().Foreground(c)
	ts := lipgloss.NewStyle().Foreground(c).Bold(true)
	left := bs.Render("╭─")
	mid := ""
	if title != "" {
		mid = bs.Render(" ") + ts.Render(title) + bs.Render(" ")
	}
	used := lipgloss.Width(left) + lipgloss.Width(mid) + 1
	dash := width - used
	if dash < 0 {
		dash = 0
	}
	return left + mid + bs.Render(strings.Repeat("─", dash)+"╮")
}

func boxBottom(width int, c lipgloss.Color) string {
	bs := lipgloss.NewStyle().Foreground(c)
	dash := width - 2
	if dash < 0 {
		dash = 0
	}
	return bs.Render("╰" + strings.Repeat("─", dash) + "╯")
}

func boxLine(content string, width int, c lipgloss.Color) string {
	bs := lipgloss.NewStyle().Foreground(c)
	inner := width - 4
	if inner < 0 {
		inner = 0
	}
	cw := lipgloss.Width(content)
	if cw > inner {
		content = truncateVis(content, inner)
		cw = lipgloss.Width(content)
	}
	pad := inner - cw
	if pad < 0 {
		pad = 0
	}
	return bs.Render("│") + " " + content + strings.Repeat(" ", pad) + " " + bs.Render("│")
}

// ─── Double-line borders (primary / focused panels) ──────────────────────────

func boxTopDouble(title string, width int, c lipgloss.Color) string {
	bs := lipgloss.NewStyle().Foreground(c)
	ts := lipgloss.NewStyle().Foreground(c).Bold(true)
	left := bs.Render("╔═")
	mid := ""
	if title != "" {
		mid = bs.Render(" ") + ts.Render(title) + bs.Render(" ")
	}
	used := lipgloss.Width(left) + lipgloss.Width(mid) + 1
	dash := width - used
	if dash < 0 {
		dash = 0
	}
	return left + mid + bs.Render(strings.Repeat("═", dash)+"╗")
}

func boxBottomDouble(width int, c lipgloss.Color) string {
	bs := lipgloss.NewStyle().Foreground(c)
	dash := width - 2
	if dash < 0 {
		dash = 0
	}
	return bs.Render("╚" + strings.Repeat("═", dash) + "╝")
}

func boxLineDouble(content string, width int, c lipgloss.Color) string {
	bs := lipgloss.NewStyle().Foreground(c)
	inner := width - 4
	if inner < 0 {
		inner = 0
	}
	cw := lipgloss.Width(content)
	if cw > inner {
		content = truncateVis(content, inner)
		cw = lipgloss.Width(content)
	}
	pad := inner - cw
	if pad < 0 {
		pad = 0
	}
	return bs.Render("║") + " " + content + strings.Repeat(" ", pad) + " " + bs.Render("║")
}
