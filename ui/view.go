package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	normalStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	titleStyle   = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)
)

func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	var tasksView strings.Builder
	tasksView.WriteString(titleStyle.Render("DevDeck") + "\n\n")

	for i, proc := range m.processes {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		status := "🔴"
		if proc.Status == "Running" {
			status = "🟢"
		}

		line := fmt.Sprintf("%s %s %s", cursor, status, proc.Config.Name)
		if proc.Err != nil {
			line += fmt.Sprintf(" (Err: %v)", proc.Err)
		}

		if m.cursor == i {
			tasksView.WriteString(focusedStyle.Render(line) + "\n")
		} else {
			tasksView.WriteString(normalStyle.Render(line) + "\n")
		}
	}

	tasksView.WriteString("\n'r': restart\n's': split view\n'q': quit\n")

	// Render the list
	taskList := lipgloss.NewStyle().
		Width(30).
		Height(m.viewport.Height). // Note: Helper needed if heights differ? For now use active viewport height
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Render(tasksView.String())

	// Render the logs
	var logPane string

	if m.pinnedIndex >= 0 {
		// Split View
		pinnedView := lipgloss.NewStyle().
			Width(m.secondaryViewport.Width).
			Height(m.secondaryViewport.Height).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("205")). // Pink border for pinned
			Padding(0, 1).
			Render(m.secondaryViewport.View())

		currentView := lipgloss.NewStyle().
			Width(m.viewport.Width).
			Height(m.viewport.Height).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1).
			Render(m.viewport.View())

		logPane = lipgloss.JoinVertical(lipgloss.Left, pinnedView, currentView)

		// Adjust task list height to match total height
		// A bit hacky: recalculate border height
		totalHeight := m.secondaryViewport.Height + m.viewport.Height + 4 // + borders
		taskList = lipgloss.NewStyle().
			Width(30).
			Height(totalHeight).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			Render(tasksView.String())

	} else {
		// Single View
		logPane = lipgloss.NewStyle().
			Width(m.viewport.Width).
			Height(m.viewport.Height).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1).
			Render(m.viewport.View())
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, taskList, logPane)
}
