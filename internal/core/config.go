package core

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	RepoURL     string `yaml:"repo_url"`
	RepoPath    string `yaml:"repo_path"`
	Branch      string `yaml:"branch"`
	Concurrency int    `yaml:"concurrency"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.RepoPath == "" {
		return nil, fmt.Errorf("repo_path is required in config")
	}

	if cfg.Branch == "" {
		cfg.Branch = detectCurrentBranch(cfg.RepoPath)
		if cfg.Branch == "" {
			cfg.Branch = "main"
		}
	}

	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 3
	}

	return &cfg, nil
}
