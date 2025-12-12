package main

import (
	"flag"
	"fmt"
	"os"

	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
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

	// Watch for config changes
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					// Reload config
					newCfg, err := config.LoadConfig(configPath)
					if err == nil {
						p.Send(ui.ConfigChangedMsg(newCfg))
					}
				}
			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()

	if err := watcher.Add(configPath); err != nil {
		log.Fatal(err)
	}
	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error: %v", err)
		os.Exit(1)
	}
}
