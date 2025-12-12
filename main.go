package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kuo-hm/devdeck/config"
	"github.com/kuo-hm/devdeck/ui"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "devdeck.yaml", "Path to configuration file")
	flag.StringVar(&configPath, "c", "devdeck.yaml", "Path to configuration file (shorthand)")
	flag.Parse()

	cfg, err := config.LoadConfig(configPath)
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
