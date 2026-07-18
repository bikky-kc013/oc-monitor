package tui

import "github.com/charmbracelet/lipgloss"

const (
	vLive = iota
	vSessions
	vAgents
	vModels
	vDaily
	vHourly
	vProjects
	vSearch
	vDetail
)

var viewLabels = []string{"Live", "Sessions", "Agents", "Models", "Daily", "Hourly", "Projects", "Search", "Detail"}

const (
	cCost = lipgloss.Color("220")

	cToken = lipgloss.Color("51")

	cOK = lipgloss.Color("42")

	cWarn = lipgloss.Color("220")
	cErr = lipgloss.Color("203")

	cAccent = lipgloss.Color("213")

	cText = lipgloss.Color("252")
	cDim  = lipgloss.Color("243")
	cMute = lipgloss.Color("238")

	cBorder = lipgloss.Color("60")
	cBorderHi = lipgloss.Color("99")

	cChart   = lipgloss.Color("42")
	cChartHi = lipgloss.Color("51")
)

var (
	sTitle = lipgloss.NewStyle().Foreground(cAccent).Bold(true)
	sTab   = lipgloss.NewStyle().Foreground(cDim).Padding(0, 1)
	sTabSel = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(cAccent).Padding(0, 1).Bold(true)
	sBold  = lipgloss.NewStyle().Foreground(cText).Bold(true)
	sDim   = lipgloss.NewStyle().Foreground(cDim)
	sMute  = lipgloss.NewStyle().Foreground(cMute)
	sHelp  = lipgloss.NewStyle().Foreground(cDim)

	sCost = lipgloss.NewStyle().Foreground(cCost).Bold(true)
	sTok  = lipgloss.NewStyle().Foreground(cToken)

	sOK  = lipgloss.NewStyle().Foreground(cOK)
	sWarn = lipgloss.NewStyle().Foreground(cWarn)
	sErr  = lipgloss.NewStyle().Foreground(cErr).Bold(true)
	sDot  = lipgloss.NewStyle().Foreground(cOK).Bold(true)

	sBar    = lipgloss.NewStyle().Foreground(cOK)
	sBarW   = lipgloss.NewStyle().Foreground(cWarn)
	sBarC   = lipgloss.NewStyle().Foreground(cErr)
	sChart  = lipgloss.NewStyle().Foreground(cChart)
	sChartHi = lipgloss.NewStyle().Foreground(cChartHi)
)

var spinFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
