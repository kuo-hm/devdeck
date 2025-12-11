package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kuo-hm/devdeck/config"
	"github.com/kuo-hm/devdeck/ui"
)

func main() {
	cfg, err := config.LoadConfig("devdeck.yaml")
	if err != nil {
		fmt.Printf("there's been an error: %v", err)
		os.Exit(1)
	}

	p := tea.NewProgram(ui.InitialModel(cfg), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error: %v", err)
		os.Exit(1)
	}
}
