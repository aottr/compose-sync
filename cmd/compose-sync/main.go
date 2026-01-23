package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/aottr/compose-sync/internal/core"
	"github.com/aottr/compose-sync/internal/stacks"
	"github.com/aottr/compose-sync/internal/version"
)

func main() {
	configPath := flag.String("config", "config.yml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("compose-sync version %s\n", version.Version)
		fmt.Printf("Commit: %s\n", version.Commit)
		fmt.Printf("Build date: %s\n", version.BuildDate)
		os.Exit(0)
	}

	cfg, err := core.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	lockPath := filepath.Join(cfg.RepoPath, ".compose-sync.lock")
	releaseLock, err := core.AcquireLock(lockPath)
	if err != nil {
		log.Fatalf("Failed to acquire lock: %v", err)
	}
	defer func() {
		if err := releaseLock(); err != nil {
			log.Printf("Warning: failed to release lock: %v", err)
		}
	}()

	currentHost, err := core.DetectHost()
	if err != nil {
		log.Fatalf("Failed to detect host: %v", err)
	}
	fmt.Printf("Detected host: %s\n", currentHost)

	fmt.Printf("Pulling git repository (branch: %s)...\n", cfg.Branch)
	changedFiles, err := core.PullAndDetectChanges(cfg.RepoPath, cfg.Branch)
	if err != nil {
		log.Fatalf("Failed to pull or detect changes: %v", err)
	}

	if len(changedFiles) == 0 {
		fmt.Println("No changes detected.")
		return
	}

	changedStacks, err := stacks.DetectChangedStacks(changedFiles)
	if err != nil {
		log.Fatalf("Failed to detect changed stacks: %v", err)
	}

	assignedStacks, err := core.GetAssignedStacks(cfg.RepoPath, currentHost)
	if err != nil {
		log.Fatalf("Failed to get assigned stacks: %v", err)
	}
	fmt.Printf("Stacks assigned to this host: %v\n", assignedStacks)

	fmt.Printf("Changed stacks: %v\n", changedStacks)

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

	stacks.DeployStacks(cfg.RepoPath, stacksToDeploy, cfg.Concurrency)
}
