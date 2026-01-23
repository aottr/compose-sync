package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func PullAndDetectChanges(repoPath, branch string) ([]string, error) {
	if err := ensureGitRepository(repoPath); err != nil {
		return nil, err
	}

	prevHead, err := getGitHead(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous HEAD: %w", err)
	}

	if err := gitFetchPull(repoPath, branch); err != nil {
		return nil, fmt.Errorf("failed to pull: %w", err)
	}

	newHead, err := getGitHead(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get new HEAD: %w", err)
	}
	if prevHead == newHead {
		return []string{}, nil
	}

	output, err := gitCommand(repoPath, "diff", "--name-only", prevHead, newHead)
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}
	changedFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	return changedFiles, nil
}

/**
 * Git helper functions
 */

func gitCommand(repoPath string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	return cmd.Output()
}

func detectCurrentBranch(repoPath string) string {
	output, err := gitCommand(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func ensureGitRepository(repoPath string) error {
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("repository path does not exist: %s", repoPath)
	}

	if _, err := os.Stat(fmt.Sprintf("%s/.git", repoPath)); os.IsNotExist(err) {
		return fmt.Errorf("path is not a git repository: %s", repoPath)
	}
	return nil
}

func gitFetchPull(repoPath, branch string) error {
	if _, err := gitCommand(repoPath, "fetch", "origin", branch); err != nil {
		return fmt.Errorf("git fetch failed: %w", err)
	}

	currentBranch, _ := gitCommand(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if strings.TrimSpace(string(currentBranch)) != branch {
		if _, err := gitCommand(repoPath, "checkout", branch); err != nil {
			return fmt.Errorf("git checkout failed: %w", err)
		}
	}

	cmd := exec.Command("git", "pull", "origin", branch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %s, %w", string(output), err)
	}
	return nil
}

func getGitHead(repoPath string) (string, error) {
	output, err := gitCommand(repoPath, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
