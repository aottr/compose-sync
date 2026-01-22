package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"gopkg.in/yaml.v3"
)

type config struct {
	RepoURL  string `yaml:"repo_url"`
	RepoPath string `yaml:"repo_path"`
}

func loadConfig(path string) (*config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.RepoPath == "" {
		return nil, fmt.Errorf("repo_path is required in config")
	}

	return &cfg, nil
}

func acquireLock(lockPath string) (func() error, error) {
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		return nil, errors.New("another compose-sync is already running")
	}

	return func() error {
		if err := syscall.Flock(int(f.Fd()), syscall.LOCK_UN); err != nil {
			f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		return os.Remove(lockPath)
	}, nil
}

func detectHost() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}
	return hostname, nil
}

type inventory struct {
	Hosts map[string][]string `yaml:"hosts"`
}

func getAssignedStacks(repoPath, hostname string) ([]string, error) {
	inventoryFile := filepath.Join(repoPath, "inventory.yml")

	data, err := os.ReadFile(inventoryFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read inventory file %s: %w", inventoryFile, err)
	}

	var inv inventory
	if err := yaml.Unmarshal(data, &inv); err != nil {
		return nil, fmt.Errorf("failed to parse inventory file %s: %w", inventoryFile, err)
	}

	stacks, exists := inv.Hosts[hostname]
	if !exists {
		return []string{}, nil // Host not in inventory, no stacks assigned
	}

	return stacks, nil
}

func pullAndDetectChanges(repoPath string) ([]string, error) {
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("repository path does not exist: %s", repoPath)
	}

	if _, err := os.Stat(fmt.Sprintf("%s/.git", repoPath)); os.IsNotExist(err) {
		return nil, fmt.Errorf("path is not a git repository: %s", repoPath)
	}

	prevHead, err := getGitHead(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous HEAD: %w", err)
	}

	if err := gitPull(repoPath); err != nil {
		return nil, fmt.Errorf("failed to pull: %w", err)
	}

	newHead, err := getGitHead(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get new HEAD: %w", err)
	}

	// If HEAD didnt change, no changes were pulled
	if prevHead == newHead {
		return []string{}, nil
	}

	changedStacks, err := findChangedStacks(repoPath, prevHead, newHead)
	if err != nil {
		return nil, fmt.Errorf("failed to find changed stacks: %w", err)
	}

	return changedStacks, nil
}

func gitPull(repoPath string) error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %s, %w", string(output), err)
	}
	return nil
}

func getGitHead(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func findChangedStacks(repoPath, oldCommit, newCommit string) ([]string, error) {
	// Get list of changed files between the two commits
	cmd := exec.Command("git", "diff", "--name-only", oldCommit, newCommit)
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}

	// Parse changed files and get stack names
	changedFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	stackSet := make(map[string]bool)

	for _, file := range changedFiles {
		if file == "" {
			continue
		}

		// Check if file is in the stacks directory
		// Format: stacks/<stack-name>/compose.yml
		if strings.HasPrefix(file, "stacks/") {
			parts := strings.Split(file, "/")
			if len(parts) >= 2 {
				stackName := parts[1]
				// filter out unknown files
				if len(parts) >= 3 && (parts[2] == "compose.yml" || parts[2] == "compose.yaml") {
					stackSet[stackName] = true
				}
			}
		}
	}

	stacks := make([]string, 0, len(stackSet))
	for stack := range stackSet {
		stacks = append(stacks, stack)
	}

	return stacks, nil
}

func deployStack(composePath string) error {
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return fmt.Errorf("compose file does not exist: %s", composePath)
	}

	composeDir := filepath.Dir(composePath)

	cmd := exec.Command("docker", "compose", "-f", composePath, "up", "-d")
	cmd.Dir = composeDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose up failed: %s, %w", string(output), err)
	}

	return nil
}
