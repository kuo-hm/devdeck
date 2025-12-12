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
	EnvFile     string       `yaml:"env_file,omitempty" json:"env_file,omitempty"`
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

	// Resolve relative paths and load .env files
	configDir := filepath.Dir(path)
	for i := range config.Tasks {
		task := &config.Tasks[i]

		// Resolve Directory
		if task.Directory != "" && !filepath.IsAbs(task.Directory) {
			task.Directory = filepath.Join(configDir, task.Directory)
		}

		// Process EnvFile
		if task.EnvFile != "" {
			envPath := task.EnvFile
			if !filepath.IsAbs(envPath) {
				envPath = filepath.Join(configDir, envPath)
			}

			fileEnv, err := parseEnvFile(envPath)
			if err == nil {
				// Prepend file envs so manual envs override them
				task.Env = append(fileEnv, task.Env...)
			} else {
				// Log warning? For now just ignore or print to stdout
				fmt.Printf("Warning: failed to load env_file %s: %v\n", envPath, err)
			}
		}
	}

	return &config, nil
}

func parseEnvFile(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var envs []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Basic KEY=VAL parsing
		if strings.Contains(line, "=") {
			envs = append(envs, line)
		}
	}
	return envs, nil
}
