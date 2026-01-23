package stacks

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

func DeployStacks(repoPath string, stacks []string, concurrency int) {
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, stack := range stacks {
		wg.Add(1)
		go func(stack string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			composePath := filepath.Join(repoPath, "stacks", stack, "compose.yml")
			if _, err := os.Stat(composePath); os.IsNotExist(err) {
				composePath = filepath.Join(repoPath, "stacks", stack, "compose.yaml")
			}

			fmt.Printf("Deploying stack: %s\n", stack)
			if err := deployStack(composePath); err != nil {
				log.Printf("Failed to deploy stack %s: %v", stack, err)
				return
			}
			fmt.Printf("Successfully deployed stack: %s\n", stack)
		}(stack)
	}
	wg.Wait()
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
