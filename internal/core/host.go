package core

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func DetectHost() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}
	return hostname, nil
}

type inventory struct {
	Hosts map[string][]string `yaml:"hosts"`
}

func GetAssignedStacks(repoPath, hostname string) ([]string, error) {
	inventoryFile := filepath.Join(repoPath, "inventory.yml")

	data, err := os.ReadFile(inventoryFile)
	if err != nil {
		return []string{}, fmt.Errorf("failed to read inventory file %s: %w", inventoryFile, err)
	}

	var inv inventory
	if err := yaml.Unmarshal(data, &inv); err != nil {
		return []string{}, fmt.Errorf("failed to parse inventory file %s: %w", inventoryFile, err)
	}

	if inv.Hosts == nil {
		return []string{}, fmt.Errorf("inventory file %s has no 'hosts' key", inventoryFile)
	}

	stacks, exists := inv.Hosts[hostname]
	if !exists {
		return []string{}, nil
	}

	if stacks == nil {
		return []string{}, nil
	}

	return stacks, nil
}
