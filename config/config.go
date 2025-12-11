package config

import (
    "fmt"
    "os"

    "gopkg.in/yaml.v3"
)

type Task struct {
    Name      string   `yaml:"name"`
    Command   string   `yaml:"command"`
    Directory string   `yaml:"directory,omitempty"`
    Env       []string `yaml:"env,omitempty"`
}

type Config struct {
    Tasks []Task `yaml:"tasks"`
}

func LoadConfig(path string) (*Config, error) {
    file, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }

    var config Config
    if err := yaml.Unmarshal(file, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config file: %w", err)
    }

    return &config, nil
}
