package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
)

func main() {
	configPath := flag.String("config", "config.yml", "Path to configuration file")
	dryRun := flag.Bool("dry-run", false, "Show what would be deployed without actually deploying")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Acquire file lock to prevent concurrent sync jobs
	lockPath := filepath.Join(cfg.RepoPath, ".compose-sync.lock")
	releaseLock, err := acquireLock(lockPath)
	if err != nil {
		log.Fatalf("Failed to acquire lock: %v", err)
	}
	defer func() {
		if err := releaseLock(); err != nil {
			log.Printf("Warning: failed to release lock: %v", err)
		}
	}()

	currentHost, err := detectHost()
	if err != nil {
		log.Fatalf("Failed to detect host: %v", err)
	}
	fmt.Printf("Detected host: %s\n", currentHost)

	fmt.Println("Pulling git repository...")
	changedStacks, err := pullAndDetectChanges(cfg.RepoPath)
	if err != nil {
		log.Fatalf("Failed to pull or detect changes: %v", err)
	}

	assignedStacks, err := getAssignedStacks(cfg.RepoPath, currentHost)
	if err != nil {
		log.Fatalf("Failed to get assigned stacks: %v", err)
	}
	fmt.Printf("Stacks assigned to this host: %v\n", assignedStacks)

	if len(changedStacks) == 0 {
		fmt.Println("No changes detected.")
		return
	}

	fmt.Printf("Changed stacks: %v\n", changedStacks)

	// Filter to only stacks assigned to this host
	stacksToDeploy := []string{}
	for _, stack := range changedStacks {
		if slices.Contains(assignedStacks, stack) {
			stacksToDeploy = append(stacksToDeploy, stack)
		}
	}

	if len(stacksToDeploy) == 0 {
		fmt.Println("No changed stacks are assigned to this host.")
		return
	}

	fmt.Printf("Stacks to deploy: %v\n", stacksToDeploy)

	if *dryRun {
		fmt.Println("DRY RUN: Would deploy the following stacks:")
		for _, stack := range stacksToDeploy {
			fmt.Printf("  - %s\n", stack)
		}
		return
	}

	// Deploy each stack
	for _, stack := range stacksToDeploy {
		composePath := filepath.Join(cfg.RepoPath, "stacks", stack, "compose.yml")
		if _, err := os.Stat(composePath); os.IsNotExist(err) {
			composePath = filepath.Join(cfg.RepoPath, "stacks", stack, "compose.yaml")
		}
		fmt.Printf("Deploying stack: %s\n", stack)
		if err := deployStack(composePath); err != nil {
			log.Printf("Failed to deploy stack %s: %v", stack, err)
			continue
		}
		fmt.Printf("Successfully deployed stack: %s\n", stack)
	}
}
