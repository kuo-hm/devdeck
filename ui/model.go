package ui

import (
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

// Model represents the state of the UI.
type Model struct {
	processes         []*process.Process
	cursor            int
	ready             bool
	viewport          viewport.Model
	secondaryViewport viewport.Model
	pinnedIndex       int
	focusedPane       Focus
}

// InitialModel creates the initial state from the configuration.
func InitialModel(cfg *config.Config) Model {
	processes := make([]*process.Process, len(cfg.Tasks))
	for i, task := range cfg.Tasks {
		processes[i] = process.NewProcess(task)
	}

	return Model{
		processes:   processes,
		cursor:      0,
		pinnedIndex: -1,
		focusedPane: FocusList,
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
		case "tab":
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

		case "ctrl+c", "q":
			for _, p := range m.processes {
				_ = p.Stop()
			}
			return m, tea.Quit
		case "up", "k":
			if m.focusedPane == FocusList {
				if m.cursor > 0 {
					m.cursor--
					m.viewport.SetContent(m.processes[m.cursor].LogBuffer)
					m.viewport.GotoBottom()
				}
			}
		case "down", "j":
			if m.focusedPane == FocusList {
				if m.cursor < len(m.processes)-1 {
					m.cursor++
					m.viewport.SetContent(m.processes[m.cursor].LogBuffer)
					m.viewport.GotoBottom()
				}
			}
		case "r":
			proc := m.processes[m.cursor]
			_ = proc.Restart()
			proc.LogBuffer += "\n--- RESTARTED ---\n"

			atBottom := m.viewport.AtBottom()
			m.viewport.SetContent(proc.LogBuffer)
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
		case "s":
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

	case LogMsg:
		proc := m.processes[msg.TaskIndex]
		proc.LogBuffer += msg.Content + "\n"

		if msg.TaskIndex == m.cursor {
			atBottom := m.viewport.AtBottom()
			m.viewport.SetContent(proc.LogBuffer)
			if atBottom {
				m.viewport.GotoBottom()
			}
		}

		if msg.TaskIndex == m.pinnedIndex {
			atBottom := m.secondaryViewport.AtBottom()
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

	return m, tea.Batch(cmds...)
}
