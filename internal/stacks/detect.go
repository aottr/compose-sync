package stacks

import (
	"strings"
)

func DetectChangedStacks(changedFiles []string) ([]string, error) {
	stackSet := make(map[string]bool)

	for _, file := range changedFiles {
		if file == "" {
			continue
		}

		if strings.HasPrefix(file, "stacks/") {
			parts := strings.Split(file, "/")
			if len(parts) >= 2 {
				stackName := parts[1]
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
