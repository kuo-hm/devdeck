package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type HealthCheck struct {
	Type     string `yaml:"type" json:"type"`         // "tcp" or "http"
	Target   string `yaml:"target" json:"target"`     // "localhost:8080" or "http://localhost..."
	Interval int    `yaml:"interval" json:"interval"` // ms
	Timeout  int    `yaml:"timeout" json:"timeout"`   // ms
}

type Task struct {
	Name        string       `yaml:"name" json:"name"`
	Command     string       `yaml:"command" json:"command"`
	Directory   string       `yaml:"directory,omitempty" json:"directory,omitempty"`
	Env         []string     `yaml:"env,omitempty" json:"env,omitempty"`
	HealthCheck *HealthCheck `yaml:"health_check,omitempty" json:"health_check,omitempty"`
	DependsOn   []string     `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	Groups      []string     `yaml:"groups,omitempty" json:"groups,omitempty"`
}

type Theme struct {
	Primary   string `yaml:"primary" json:"primary"`
	Secondary string `yaml:"secondary" json:"secondary"`
	Border    string `yaml:"border" json:"border"`
	Text      string `yaml:"text" json:"text"`
}

type Config struct {
	Tasks []Task `yaml:"tasks" json:"tasks"`
	Theme *Theme `yaml:"theme,omitempty" json:"theme,omitempty"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".json":
		if err := json.Unmarshal(file, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config file: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(file, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config file: %w", err)
		}
	default:
		// Try YAML by default if unknown extension, or error?
		// Let's fallback to YAML for backward compatibility or strictness?
		// Given the task, let's assume strictness or default to YAML.
		// I'll try YAML as default fallback to match previous behavior if someone passed a file without extension (unlikely but safe).
		if err := yaml.Unmarshal(file, &config); err != nil {
			// If YAML fails, try JSON? No, that's messy. Let's stick to extension check.
			// Actually, returning an error for unknown extension is cleaner.
			return nil, fmt.Errorf("unsupported config file extension: %s", ext)
		}
	}

	// Resolve relative paths
	configDir := filepath.Dir(path)
	for i := range config.Tasks {
		if config.Tasks[i].Directory != "" && !filepath.IsAbs(config.Tasks[i].Directory) {
			config.Tasks[i].Directory = filepath.Join(configDir, config.Tasks[i].Directory)
		}
	}

	return &config, nil
}
