package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
	"github.com/kuo-hm/devdeck/config"
	"github.com/kuo-hm/devdeck/ui"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			f, _ := os.OpenFile("devdeck-crash.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if f != nil {
				timestamp := time.Now().Format(time.RFC3339)
				f.WriteString(fmt.Sprintf("[%s] PANIC: %v\n%s\n", timestamp, r, string(debug.Stack())))
				f.Close()
			}
			fmt.Println("\n\nDevDeck crashed! 💥")
			fmt.Println("Error details saved to devdeck-crash.log")
			os.Exit(1)
		}
	}()

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
