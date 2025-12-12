package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kuo-hm/devdeck/config"
	"github.com/kuo-hm/devdeck/process"
)

// Focus represents the active pane
type Focus int

const (
	FocusList Focus = iota
	FocusLog
	FocusSecondary
)

// InputMode determines what the input bar is used for
type InputMode int

const (
	InputNone InputMode = iota
	InputProcess
	InputSearch
)

// Model represents the state of the UI.
type Model struct {
	processes         []*process.Process
	cursor            int
	ready             bool
	viewport          viewport.Model
	secondaryViewport viewport.Model
	pinnedIndex       int
	focusedPane       Focus
	textInput         textinput.Model
	inputMode         InputMode
	searchQuery       string
	matches           []int // Line numbers of search matches
	matchIndex        int   // Current match index (in matches array)
	helpVisible       bool
	theme             *config.Theme
	width             int
	height            int
}

// InitialModel creates the initial state from the configuration.
func InitialModel(cfg *config.Config) Model {
	processes := make([]*process.Process, len(cfg.Tasks))
	for i, task := range cfg.Tasks {
		processes[i] = process.NewProcess(task)
	}

	ti := textinput.New()
	ti.Placeholder = "Type input..."
	ti.CharLimit = 156
	ti.Width = 30

	return Model{
		processes:   processes,
		cursor:      0,
		pinnedIndex: -1,
		focusedPane: FocusList,
		textInput:   ti,
		inputMode:   InputNone,
		searchQuery: "",
		matches:     []int{},
		matchIndex:  -1,
		theme:       cfg.Theme,
	}
}

// Init starts all processes and the activity listener loop.
func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, proc := range m.processes {
		if err := proc.Start(); err != nil {
			// Handle start error
		}
		cmds = append(cmds, waitForActivity(proc.Config.Name, proc.Output))
	}
	return tea.Batch(cmds...)
}

func waitForActivity(name string, output chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-output
		if !ok {
			return nil
		}
		return LogMsg{ProcessName: name, Content: line}
	}
}

// Update handles incoming messages and updates the model.
// Update handles incoming messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case ConfigChangedMsg:
		newCfg := msg
		m.theme = newCfg.Theme
		var newProcs []*process.Process

		// Map existing processes by name for easy lookup
		existing := make(map[string]*process.Process)
		for _, p := range m.processes {
			existing[p.Config.Name] = p
		}

		for _, task := range newCfg.Tasks {
			if proc, ok := existing[task.Name]; ok {
				// Process exists, check if config changed
				// Simple comparison: check command, dir, env
				// We can use reflect.DeepEqual if we import reflect, or manual check.
				// Manual check for key fields is safer/faster.
				changed := proc.Config.Command != task.Command ||
					proc.Config.Directory != task.Directory ||
					len(proc.Config.Env) != len(task.Env) // Superficial env check

				if !changed {
					// Check Env deeply
					for i, e := range proc.Config.Env {
						if e != task.Env[i] {
							changed = true
							break
						}
					}
				}

				if changed {
					// Config changed, restart with new config
					_ = proc.Stop()
					// Create new process instance to ensure clean state
					newProc := process.NewProcess(task)
					_ = newProc.Start()
					newProcs = append(newProcs, newProc)
					// We need to re-hook the activity listener?
					// Yes, Init() called Start() and waitForActivity.
					// We need to spawn waitForActivity for the new process.
					cmds = append(cmds, waitForActivity(newProc.Config.Name, newProc.Output))
				} else {
					// Keep existing process
					newProcs = append(newProcs, proc)
					// We need to ensure the index in waitForActivity matches?
					// waitForActivity captures 'index'. If index changes (reorder), LogMsg will have old index.
					// Fix: LogMsg should probably identify by Name, or we just accept that reordering might break log routing temporarily?
					// Actually, waitForActivity is a goroutine. If we reorder, index 0 might become index 1.
					// The existing goroutine for that process will send LogMsg{TaskIndex: 0}.
					// But if we moved it to index 1, m.processes[0] is now something else.
					// So LogMsg will append logs to the WRONG process!
					// CRITICAL ISSUE.

					// To fix hot reload reordering, LogMsg should use a pointer to the process or Name?
					// Using pointer is unsafe in tea.Msg? No, it's fine.
					// Or just use Name.
				}
			} else {
				// New process
				newProc := process.NewProcess(task)
				if err := newProc.Start(); err == nil {
					newProcs = append(newProcs, newProc)
					cmds = append(cmds, waitForActivity(newProc.Config.Name, newProc.Output))
				}
			}
		}

		// Stop processes that are removed (in existing but not in newProcs)
		// Wait, newProcs contains the new list.
		// We iterated newCfg.Tasks.
		// Any process in 'existing' that was NOT reused is effectively removed?
		// No, we reused instances. If we reused, it's in newProcs.
		// If we created new, the old one is abandoned. We must stop it.

		// Let's track used names
		usedNames := make(map[string]bool)
		for _, p := range newProcs {
			usedNames[p.Config.Name] = true
		}

		for name, p := range existing {
			if !usedNames[name] {
				_ = p.Stop()
			}
		}

		m.processes = newProcs
		// Adjust cursor if out of bounds
		if m.cursor >= len(m.processes) {
			m.cursor = len(m.processes) - 1
			if m.cursor < 0 {
				m.cursor = 0
			}
		}

		// Re-render viewport
		if len(m.processes) > 0 {
			m.viewport.SetContent(m.processes[m.cursor].LogBuffer)
		} else {
			m.viewport.SetContent("")
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Layout definition:
		// Task List: 35 chars wide (fixed enough for names)
		// Logs: Remainder
		// Height: Full height - header(2) - borders(2) - footer(1)?

		listWidth := 35
		logWidth := msg.Width - listWidth - 4 // -4 for borders/padding overhead (approx)
		if logWidth < 10 {
			logWidth = 10
		}

		// Calculate available height for content
		// Title (2) + Footer/Help hint (2) + Borders (2) = 6 lines overhead?
		// Let's be safe with -4 for viewports, View() handles outer containers.
		// View() adds borders.

		availableHeight := msg.Height - 4 // General safe area

		if !m.ready {
			m.viewport = viewport.New(logWidth, availableHeight)
			m.secondaryViewport = viewport.New(logWidth, availableHeight)
			m.ready = true
		} else {
			m.viewport.Width = logWidth
			m.secondaryViewport.Width = logWidth

			if m.pinnedIndex >= 0 {
				// Split view: vertical half
				halfHeight := availableHeight / 2
				remainderHeight := availableHeight - halfHeight

				// Top pane has title: -3 (2 border, 1 title)
				vpHeightTop := halfHeight - 3
				if vpHeightTop < 0 {
					vpHeightTop = 0
				}

				vpHeightBot := remainderHeight - 2
				if vpHeightBot < 0 {
					vpHeightBot = 0
				}

				m.secondaryViewport.Height = vpHeightTop
				m.viewport.Height = vpHeightBot
			} else {
				// Single View
				h := availableHeight
				if h < 0 {
					h = 0
				}
				m.viewport.Height = h
				m.secondaryViewport.Height = 0
			}
		}

	case tea.KeyMsg:
		// Toggle Help
		if msg.String() == "?" {
			m.helpVisible = !m.helpVisible
			return m, nil
		}

		// If help is visible, ignore other keys except Esc/q to close
		if m.helpVisible {
			if msg.String() == "esc" || msg.String() == "q" {
				m.helpVisible = false
			}
			return m, nil
		}

		switch msg.String() {
		case "i":
			if m.inputMode == InputNone {
				m.inputMode = InputProcess
				m.textInput.Placeholder = "Type input..."
				m.textInput.Focus()
				return m, textinput.Blink
			}

		case "/":
			if m.inputMode == InputNone {
				m.inputMode = InputSearch
				m.textInput.Placeholder = "Search logs..."
				m.textInput.Focus()
				return m, textinput.Blink
			}

		case "enter":
			if m.inputMode == InputProcess {
				val := m.textInput.Value()
				// Send input to the currently selected process
				proc := m.processes[m.cursor]
				_ = proc.SendInput(val)

				// Reset input
				m.textInput.SetValue("")
				m.textInput.Blur()
				m.inputMode = InputNone
			} else if m.inputMode == InputSearch {
				val := m.textInput.Value()
				m.searchQuery = val

				// Update viewport with filtered content
				proc := m.processes[m.cursor]
				content, matches := highlightLogs(proc.LogBuffer, m.searchQuery)
				m.viewport.SetContent(content)
				m.matches = matches
				m.viewport.GotoBottom()

				// Reset input
				m.textInput.SetValue("")
				m.textInput.Blur()
				m.inputMode = InputNone
			}

		case "esc":
			if m.inputMode != InputNone {
				m.textInput.SetValue("")
				m.textInput.Blur()
				m.inputMode = InputNone
			} else if m.searchQuery != "" {
				// Clear search query
				m.searchQuery = ""
				proc := m.processes[m.cursor]
				m.viewport.SetContent(proc.LogBuffer)
				m.viewport.GotoBottom()
			}

		case "tab":
			if m.inputMode == InputNone { // Only tab if not typing
				if m.pinnedIndex != -1 {
					// Cycle 3 panes if split
					m.focusedPane = (m.focusedPane + 1) % 3
				} else {
					// Cycle 2 panes if single
					m.focusedPane = (m.focusedPane + 1) % 2
					if m.focusedPane == FocusSecondary {
						m.focusedPane = FocusList // Skip secondary if not visible
					}
				}
			}

		case "ctrl+c", "q":
			if m.inputMode == InputNone {
				for _, p := range m.processes {
					_ = p.Stop()
				}
				return m, tea.Quit
			}

		case "up", "k":
			if m.inputMode == InputNone && m.focusedPane == FocusList {
				if m.cursor > 0 {
					m.cursor--
					content, _ := highlightLogs(m.processes[m.cursor].LogBuffer, m.searchQuery)
					m.viewport.SetContent(content)
					m.viewport.GotoBottom()
				}
			}
		case "down", "j":
			if m.inputMode == InputNone && m.focusedPane == FocusList {
				if m.cursor < len(m.processes)-1 {
					m.cursor++
					content, _ := highlightLogs(m.processes[m.cursor].LogBuffer, m.searchQuery)
					m.viewport.SetContent(content)
					m.viewport.GotoBottom()
				}
			}
		case "r":
			if m.inputMode == InputNone {
				proc := m.processes[m.cursor]
				_ = proc.Restart()
				proc.LogBuffer += "\n--- RESTARTED ---\n"

				atBottom := m.viewport.AtBottom()
				content, matches := highlightLogs(proc.LogBuffer, m.searchQuery)
				m.viewport.SetContent(content)
				m.matches = matches
				if atBottom {
					m.viewport.GotoBottom()
				}

				if m.cursor == m.pinnedIndex {
					atBottomSec := m.secondaryViewport.AtBottom()
					m.secondaryViewport.SetContent(proc.LogBuffer)
					if atBottomSec {
						m.secondaryViewport.GotoBottom()
					}
				}
			}
		case "s":
			if m.inputMode == InputNone {
				if m.pinnedIndex == -1 {
					// Enable Split View
					m.pinnedIndex = m.cursor
					m.secondaryViewport.SetContent(m.processes[m.pinnedIndex].LogBuffer)

					// Resize viewports for split
					if m.height > 0 {
						availableHeight := m.height - 4
						halfHeight := availableHeight / 2               // Top pane
						remainderHeight := availableHeight - halfHeight // Bottom pane gets remainder

						// Top pane has a Title line now ("📌 Name")
						// So subtract 2 (border) + 1 (title) = 3
						vpHeightTop := halfHeight - 3
						if vpHeightTop < 0 {
							vpHeightTop = 0
						}

						vpHeightBot := remainderHeight - 2
						if vpHeightBot < 0 {
							vpHeightBot = 0
						}

						m.secondaryViewport.Height = vpHeightTop
						m.viewport.Height = vpHeightBot

						m.viewport.GotoBottom() // Re-scroll user viewport too
						m.secondaryViewport.GotoBottom()
					}
				} else {
					// Disable Split View
					m.pinnedIndex = -1
					m.focusedPane = FocusList // Reset focus if on secondary

					// Resize viewport for single view
					if m.height > 0 {
						availableHeight := m.height - 4
						h := availableHeight
						if h < 0 {
							h = 0
						}
						m.viewport.Height = h
						m.secondaryViewport.Height = 0
						m.viewport.GotoBottom()
					}
				}
			}
		}

	case LogMsg:
		// Find process by name
		var proc *process.Process
		var index int = -1
		for i, p := range m.processes {
			if p.Config.Name == msg.ProcessName {
				proc = p
				index = i
				break
			}
		}

		if proc == nil {
			// Process might have been removed during hot reload
			return m, nil
		}

		proc.LogBuffer += msg.Content + "\n"

		if index == m.cursor {
			atBottom := m.viewport.AtBottom()
			// Apply filter if search query exists
			content, matches := highlightLogs(proc.LogBuffer, m.searchQuery)
			m.viewport.SetContent(content)
			m.matches = matches
			if atBottom {
				m.viewport.GotoBottom()
			}
		}

		if index == m.pinnedIndex {
			atBottom := m.secondaryViewport.AtBottom()
			// Pinned view always shows full logs (design choice for now)
			m.secondaryViewport.SetContent(proc.LogBuffer)
			if atBottom {
				m.secondaryViewport.GotoBottom()
			}
		}

		cmds = append(cmds, waitForActivity(msg.ProcessName, proc.Output))
	}

	if m.focusedPane == FocusLog {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.focusedPane == FocusSecondary {
		m.secondaryViewport, cmd = m.secondaryViewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.inputMode != InputNone {
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func highlightLogs(buffer string, query string) (string, []int) {
	if query == "" {
		return buffer, []int{}
	}

	lines := strings.Split(buffer, "\n")
	var matches []int
	var highlighted strings.Builder

	// Prepare highlight style (classic yellow background, black text)
	hlStyle := lipgloss.NewStyle().Background(lipgloss.Color("#FFFF00")).Foreground(lipgloss.Color("#000000"))

	lowQuery := strings.ToLower(query)

	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), lowQuery) {
			matches = append(matches, i)

			var sb strings.Builder
			currentLower := strings.ToLower(line)
			currentOriginal := line

			for {
				idx := strings.Index(currentLower, lowQuery)
				if idx == -1 {
					sb.WriteString(currentOriginal)
					break
				}

				sb.WriteString(currentOriginal[:idx])
				sb.WriteString(hlStyle.Render(currentOriginal[idx : idx+len(query)]))

				currentLower = currentLower[idx+len(query):]
				currentOriginal = currentOriginal[idx+len(query):]
			}
			highlighted.WriteString(sb.String() + "\n")
		} else {
			highlighted.WriteString(line + "\n")
		}
	}
	return highlighted.String(), matches
}
