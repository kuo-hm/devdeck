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

	// Determine border colors based on focus
	listBorderColor := lipgloss.Color("63") // Default dim purple
	if m.focusedPane == FocusList {
		listBorderColor = lipgloss.Color("205") // Pink for focus
	}

	logBorderColor := lipgloss.Color("63")
	if m.focusedPane == FocusLog {
		logBorderColor = lipgloss.Color("205")
	}

	secondaryBorderColor := lipgloss.Color("63")
	if m.focusedPane == FocusSecondary {
		secondaryBorderColor = lipgloss.Color("205")
	}

	// Determine list width (e.g., 30% of screen, min 30)
	listWidth := m.viewport.Width / 2 // Viewport is half width? No, viewport width is calculated as msg.Width / 2 in Update.
	// Wait, m.viewport.Width is already half the screen width roughly?
	// In Update: m.viewport = viewport.New(msg.Width/2, ...)
	// So m.viewport.Width is 50%.

	// Let's use a fixed 30% of total width for the List?
	// We don't have msg.Width here in View(), we rely on m.viewport.Width which is half.
	// Actually, let's just make it wider for now, e.g. 45.
	// Or better, let's use a dynamic size if we can infer it.
	// We can't easily infer total width from just m.viewport.Width unless we assume it's exactly 50%.

	// Let's just set it to 40 for now, simpler and covers most cases.
	listWidth = 40

	// Render the list
	taskList := lipgloss.NewStyle().
		MarginTop(5).
		Width(listWidth).
		Height(m.viewport.Height-2). // +2 for border overhead
		Border(lipgloss.NormalBorder()).
		BorderForeground(listBorderColor).
		Padding(1, 2).
		Render(tasksView.String())

	// Render the logs
	var logPane string

	// Width for logs needs to be remaining space?
	// Currently viewport is set to half width.
	// We should probably rely on the layout engine (JoinHorizontal) or update Viewport width in Update.
	// For now, increasing list width might push logs?
	// View() just renders strings. JoinHorizontal joins them.
	// If List is 40 and Log is 50% of screen... on a small screen (80 cols), 40 + 40 = 80. Tight.
	// On 100 cols: 40 + 50 = 90. Fine.

	if m.pinnedIndex >= 0 {
		// Split View
		pinnedView := lipgloss.NewStyle().
			Width(m.secondaryViewport.Width).
			Height(m.secondaryViewport.Height-2). // +2 for border
			Border(lipgloss.NormalBorder()).
			BorderForeground(secondaryBorderColor).
			Padding(0, 1).
			Render(m.secondaryViewport.View())

		currentView := lipgloss.NewStyle().
			Width(m.viewport.Width).
			Height(m.viewport.Height-2). // +2 for border
			Border(lipgloss.NormalBorder()).
			BorderForeground(logBorderColor).
			Padding(0, 1).
			Render(m.viewport.View())

		logPane = lipgloss.JoinVertical(lipgloss.Left, pinnedView, currentView)

		// Adjust task list height to match total height
		totalHeight := m.secondaryViewport.Height + m.viewport.Height + 4 // +4 for two borders
		taskList = lipgloss.NewStyle().
			Width(listWidth).
			Height(totalHeight).
			Border(lipgloss.NormalBorder()).
			BorderForeground(listBorderColor).
			Padding(1, 2).
			Render(tasksView.String())

	} else {
		// Single View
		logPane = lipgloss.NewStyle().
			MarginTop(5).
			Width(m.viewport.Width).
			Height(m.viewport.Height-2). // +2 for border
			Border(lipgloss.NormalBorder()).
			BorderForeground(logBorderColor).
			Padding(0, 1).
			Render(m.viewport.View())
	}

	// Main view is list + logs
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, taskList, logPane)

	if m.inputMode != InputNone {
		inputView := lipgloss.NewStyle().
			Width(60).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			Render(m.textInput.View())

		// Place input view at bottom, or overlay?
		// Simpler to join vertical at the bottom for now, effectively reducing other heights?
		// Or just append it. If we append, it might push screen up.
		// Let's just return it at bottom.
		return lipgloss.JoinVertical(lipgloss.Left, mainView, inputView)
	}

	// Wrap everything in a container with padding
	return lipgloss.NewStyle().
		PaddingTop(50).
		Render(mainView)
}
