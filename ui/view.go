package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kuo-hm/devdeck/config"
)

func getThemeColor(theme *config.Theme, key, fallback string) lipgloss.Color {
	if theme == nil {
		return lipgloss.Color(fallback)
	}
	switch key {
	case "primary":
		if theme.Primary != "" {
			return lipgloss.Color(theme.Primary)
		}
	case "secondary":
		if theme.Secondary != "" {
			return lipgloss.Color(theme.Secondary)
		}
	case "border":
		if theme.Border != "" {
			return lipgloss.Color(theme.Border)
		}
	case "text":
		if theme.Text != "" {
			return lipgloss.Color(theme.Text)
		}
	}
	return lipgloss.Color(fallback)
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	// Define colors
	primary := getThemeColor(m.theme, "primary", "205")         // Pink
	secondary := getThemeColor(m.theme, "secondary", "#7D56F4") // Purple
	border := getThemeColor(m.theme, "border", "63")            // Dim Purple
	text := getThemeColor(m.theme, "text", "240")               // Grey

	focusedStyle := lipgloss.NewStyle().Foreground(primary)
	normalStyle := lipgloss.NewStyle().Foreground(text)
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(secondary).
		Padding(0, 1)

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

		pin := "  "
		if m.pinnedIndex == i {
			pin = "📌"
		}

		line := fmt.Sprintf("%s %s %s %s", cursor, pin, status, proc.Config.Name)

		if proc.Status == "Running" {
			mb := float64(proc.MemUsage) / 1024 / 1024
			line += fmt.Sprintf(" (%.0f%% CPU, %.0f MB)", proc.CPUUsage, mb)
		}

		if proc.Err != nil {
			line += fmt.Sprintf(" (Err: %v)", proc.Err)
		}

		if m.cursor == i {
			tasksView.WriteString(focusedStyle.Render(line) + "\n")
		} else {
			tasksView.WriteString(normalStyle.Render(line) + "\n")
		}
	}

	tasksView.WriteString("\n'r': restart\n's': split view\n'i': input\n'/': search\n'?': help\n'q': quit\n")

	// Determine border colors based on focus
	listBorderColor := border
	if m.focusedPane == FocusList {
		listBorderColor = primary
	}

	logBorderColor := border
	if m.focusedPane == FocusLog {
		logBorderColor = primary
	}

	secondaryBorderColor := border
	if m.focusedPane == FocusSecondary {
		secondaryBorderColor = primary
	}

	// Layout Constants (Must match WindowSizeMsg in Update)
	const listWidthC = 35

	// Available height for panels (Screen - Header - Footer overhead)
	// We use m.viewport.Height as the guide for Log Pane height.
	// Task list should match total height.

	// Render the list
	// We want the list to fill the available height. Use m.height as base.
	// Safety check for m.height being 0 at start
	safeHeight := m.height - 4
	if safeHeight < 0 {
		safeHeight = 0
	}

	taskList := lipgloss.NewStyle().
		Width(listWidthC).
		Height(safeHeight).
		Border(lipgloss.NormalBorder()).
		BorderForeground(listBorderColor).
		Padding(0, 1). // Reduced padding to save space?
		Render(tasksView.String())

	// Render the logs
	var logPane string

	// The log viewports are already resized in Update() to the correct width/height.
	// We just need to wrap them in a border.

	if m.pinnedIndex >= 0 {
		// Split View
		pinnedName := m.processes[m.pinnedIndex].Config.Name

		// Manually create a title if BorderTitle isn't available/reliable?
		// Let's try standard BorderTitle.
		// Manual Pin Header
		titleStyle := lipgloss.NewStyle().
			Foreground(secondary).
			Bold(true).
			Padding(0, 0, 0, 1) // Left padding for text indent

		title := titleStyle.Render("📌 " + pinnedName)

		// Join title + content vertically.
		// Note: model.go reduces viewport height by 1 extra line (3 total) to fit this title.
		content := lipgloss.JoinVertical(lipgloss.Left, title, m.secondaryViewport.View())

		pinnedView := lipgloss.NewStyle().
			Width(m.secondaryViewport.Width).
			Height(m.secondaryViewport.Height+2).
			Border(lipgloss.NormalBorder()).
			BorderForeground(secondaryBorderColor).
			Padding(0, 0). // Padding inside border handled by content/title structure?
			// Viewport has no padding. Title has padding.
			// But keeping 0,1 padding matches the other log pane style?
			// If we put title inside, 'Padding(0,1)' applies to the wrapper.
			// Let's keep it consistent.
			Padding(0, 1).
			Render(content)

		currentView := lipgloss.NewStyle().
			Width(m.viewport.Width).
			Height(m.viewport.Height+1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(logBorderColor).
			Padding(0, 1).
			Render(m.viewport.View())

		logPane = lipgloss.JoinVertical(lipgloss.Left, pinnedView, currentView)
	} else {
		// Single View
		logPane = lipgloss.NewStyle().
			Width(m.viewport.Width).
			// Height(m.viewport.Height).
			Border(lipgloss.NormalBorder()).
			BorderForeground(logBorderColor).
			Padding(0, 1).
			Render(m.viewport.View())
	}

	// Status Bar
	statusBarStyle := lipgloss.NewStyle().
		Width(m.width).
		Background(lipgloss.Color("#333333")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1)

	statusText := fmt.Sprintf("CPU: %.1f%% | MEM: %.1f%%", m.cpuUsage, m.memUsage)
	statusBar := statusBarStyle.Render(statusText)

	// Combine List + LogPane
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, taskList, logPane)

	// Create help view if visible (MODAL)
	if m.helpVisible {
		helpBox := lipgloss.NewStyle().
			Width(60).
			Height(20).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			Align(lipgloss.Center).
			Render(
				titleStyle.Render("Help") + "\n\n" +
					"Navigation\n" +
					"  ↑/k, ↓/j   : Move cursor\n" +
					"  Tab        : Switch focus\n\n" +
					"Actions\n" +
					"  Enter      : Select / Input\n" +
					"  r          : Restart process\n" +
					"  s          : Split/Pin view\n" +
					"  i          : Interact (Stdin)\n" +
					"  /          : Search logs\n\n" +
					"General\n" +
					"  ?          : Close Help\n" +
					"  q/Esc      : Quit / Back",
			)

		// Center modal? For now just return it as full view or overlay.
		// Previous code just returned it with padding.
		return lipgloss.NewStyle().
			Padding(2).
			Render(helpBox)
	}

	// Helper for Input Mode (shows input box)
	if m.inputMode != InputNone {
		inputView := lipgloss.NewStyle().
			Width(60).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primary).
			Render(m.textInput.View())

		// Insert input between main view and status bar?
		// Or append to main view.
		mainView = lipgloss.JoinVertical(lipgloss.Left, mainView, inputView)
	}

	// Append Status Bar at the very bottom
	// finalView := lipgloss.JoinVertical(lipgloss.Left, mainView, statusBar) // Redundant

	// Wrap everything in a container with padding
	// User previously set PaddingTop(50), resetting to 1 for usability.
	// Actually padding 1 around the whole specific view might shift the status bar.
	// Status bar should be full width.
	// So maybe padding applies to mainView BEFORE status bar?

	// Let's optimize: MainView Padding 1, StatusBar no padding (except internal).

	mainViewPadded := lipgloss.NewStyle().Padding(1).Render(mainView)

	// Re-compose final
	return lipgloss.JoinVertical(lipgloss.Left, mainViewPadded, statusBar)
}
