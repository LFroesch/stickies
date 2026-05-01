package main

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary = lipgloss.Color("#5AF78E")
	colorAccent  = lipgloss.Color("#57C7FF")
	colorWarn    = lipgloss.Color("#FF6AC1")
	colorError   = lipgloss.Color("#FF5C57")
	colorDim     = lipgloss.Color("#606060")
	colorText    = lipgloss.Color("#EEEEEE")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	activePageStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Underline(true)

	inactivePageStyle = lipgloss.NewStyle().
				Foreground(colorDim)

	keyStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	dimTextStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	errorTextStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	successTextStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true)

	warnTextStyle = lipgloss.NewStyle().
			Foreground(colorWarn).
			Bold(true)

	statusMsgStyle = lipgloss.NewStyle().
			Foreground(colorAccent)

	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDim).
			Padding(0, 1)

	panelActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorAccent).
				Padding(0, 1)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true)

	pinnedMarkStyle = lipgloss.NewStyle().
			Foreground(colorWarn).
			Bold(true)

	tagChipStyle = lipgloss.NewStyle().
			Foreground(colorAccent)

	dateHeaderStyle = lipgloss.NewStyle().
			Foreground(colorWarn).
			Bold(true)

	todayMarkStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	textStyle = lipgloss.NewStyle().
			Foreground(colorText)
)
