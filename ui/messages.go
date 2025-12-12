package ui

import (
	"github.com/kuo-hm/devdeck/config"
)

type LogMsg struct {
	ProcessName string
	Content     string
}

type ProcessFinishedMsg struct {
	ProcessName string
	Err         error
}

type ConfigChangedMsg *config.Config
