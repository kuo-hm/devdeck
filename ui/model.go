package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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
	}
}

// Init starts all processes and the activity listener loop.
func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	for i, proc := range m.processes {
		if err := proc.Start(); err != nil {
			// Handle start error
		}
		cmds = append(cmds, waitForActivity(i, proc.Output))
	}
	return tea.Batch(cmds...)
}

func waitForActivity(index int, output chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-output
		if !ok {
			return nil
		}
		return LogMsg{TaskIndex: index, Content: line}
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
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width/2, msg.Height-6)
			m.secondaryViewport = viewport.New(msg.Width/2, msg.Height-6)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width / 2
			m.secondaryViewport.Width = msg.Width / 2

			if m.pinnedIndex >= 0 {
				availableHeight := msg.Height - 4
				h := availableHeight / 2
				if h < 0 {
					h = 0
				}
				m.viewport.Height = h
				m.secondaryViewport.Height = h
			} else {
				availableHeight := msg.Height - 2
				h := availableHeight
				if h < 0 {
					h = 0
				}
				m.viewport.Height = h
				m.secondaryViewport.Height = 0
			}
		}

	case tea.KeyMsg:
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
				content := filterLogs(proc.LogBuffer, m.searchQuery)
				m.viewport.SetContent(content)
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
					content := filterLogs(m.processes[m.cursor].LogBuffer, m.searchQuery)
					m.viewport.SetContent(content)
					m.viewport.GotoBottom()
				}
			}
		case "down", "j":
			if m.inputMode == InputNone && m.focusedPane == FocusList {
				if m.cursor < len(m.processes)-1 {
					m.cursor++
					content := filterLogs(m.processes[m.cursor].LogBuffer, m.searchQuery)
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
				content := filterLogs(proc.LogBuffer, m.searchQuery)
				m.viewport.SetContent(content)
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
					m.pinnedIndex = m.cursor
					m.secondaryViewport.SetContent(m.processes[m.pinnedIndex].LogBuffer)
					m.secondaryViewport.GotoBottom()

					h := (m.viewport.Height - 2) / 2
					if h < 0 {
						h = 0
					}
					m.viewport.Height = h
					m.secondaryViewport.Height = h - 1
				} else {
					m.pinnedIndex = -1
					m.viewport.Height = (m.viewport.Height * 2) + 2
					if m.focusedPane == FocusSecondary {
						m.focusedPane = FocusList
					}
				}
			}
		}

	case LogMsg:
		proc := m.processes[msg.TaskIndex]
		proc.LogBuffer += msg.Content + "\n"

		if msg.TaskIndex == m.cursor {
			atBottom := m.viewport.AtBottom()
			// Apply filter if search query exists
			content := filterLogs(proc.LogBuffer, m.searchQuery)
			m.viewport.SetContent(content)
			if atBottom {
				m.viewport.GotoBottom()
			}
		}

		if msg.TaskIndex == m.pinnedIndex {
			atBottom := m.secondaryViewport.AtBottom()
			// Pinned view always shows full logs (design choice for now)
			m.secondaryViewport.SetContent(proc.LogBuffer)
			if atBottom {
				m.secondaryViewport.GotoBottom()
			}
		}

		cmds = append(cmds, waitForActivity(msg.TaskIndex, proc.Output))
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

func filterLogs(buffer string, query string) string {
	if query == "" {
		return buffer
	}
	var filtered strings.Builder
	lines := strings.Split(buffer, "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(query)) {
			filtered.WriteString(line + "\n")
		}
	}
	return filtered.String()
}
