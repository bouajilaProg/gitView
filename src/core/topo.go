package main

import "sort"

// topologicalSort performs Kahn's algorithm for topological ordering
func topologicalSort(commits map[string]*CommitData, initialOrder []string) []string {
	// Build in-degree map (number of children for each commit)
	inDegree := make(map[string]int)

	for hash := range commits {
		inDegree[hash] = 0
	}

	// Build child->parent edges (so children are processed before parents)
	for _, commit := range commits {
		for _, parent := range commit.Parents {
			if _, exists := commits[parent]; exists {
				inDegree[parent]++
			}
		}
	}

	// Start with commits that have no children (leaf commits / HEAD)
	var queue []string
	for hash := range commits {
		if inDegree[hash] == 0 {
			queue = append(queue, hash)
		}
	}

	// Sort queue by date for deterministic ordering
	sort.Slice(queue, func(i, j int) bool {
		left := commits[queue[i]].Date
		right := commits[queue[j]].Date
		if left.Equal(right) {
			return queue[i] < queue[j]
		}
		return left.After(right)
	})

	var sorted []string
	for len(queue) > 0 {
		// Pop first element
		current := queue[0]
		queue = queue[1:]
		sorted = append(sorted, current)

		// Process parents (reduce their in-degree)
		for _, parent := range commits[current].Parents {
			if _, exists := commits[parent]; exists {
				inDegree[parent]--
				if inDegree[parent] == 0 {
					queue = append(queue, parent)
					// Re-sort queue by date
					sort.Slice(queue, func(i, j int) bool {
						left := commits[queue[i]].Date
						right := commits[queue[j]].Date
						if left.Equal(right) {
							return queue[i] < queue[j]
						}
						return left.After(right)
					})
				}
			}
		}
	}

	return sorted
}
