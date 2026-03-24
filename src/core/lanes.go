package main

import "sort"

// assignLanes assigns lane indices to commits for visual layout
// Processing from newest (HEAD) to oldest, children inherit lane to parents
func assignLanes(commits map[string]*CommitData, sortedOrder []string) {
	// Build child map: parent -> list of children
	childrenOf := make(map[string][]string)
	for hash, commit := range commits {
		for _, parentHash := range commit.Parents {
			if _, exists := commits[parentHash]; exists {
				childrenOf[parentHash] = append(childrenOf[parentHash], hash)
			}
		}
	}

	for parentHash, children := range childrenOf {
		sort.Slice(children, func(i, j int) bool {
			left := commits[children[i]].Date
			right := commits[children[j]].Date
			if left.Equal(right) {
				return children[i] < children[j]
			}
			return left.After(right)
		})
		childrenOf[parentHash] = children
	}

	// Track lane assignments
	commitLanes := make(map[string]int) // commit hash -> lane
	laneAvailable := make(map[int]bool) // lane -> is it free for reuse
	nextLane := 0

	// Process commits from newest to oldest (topological order)
	for _, hash := range sortedOrder {
		commit := commits[hash]
		assignedLane := -1

		// Check if any child already assigned a lane to us
		children := childrenOf[hash]
		for _, childHash := range children {
			if childLane, exists := commitLanes[childHash]; exists {
				childCommit := commits[childHash]
				// If the child has only one parent (this commit), inherit its lane
				if len(childCommit.Parents) == 1 {
					assignedLane = childLane
					break
				}
				// If we are the first parent of a merge commit, inherit its lane
				if len(childCommit.Parents) > 1 && childCommit.Parents[0] == hash {
					assignedLane = childLane
					break
				}
			}
		}

		// If no lane inherited, look for a free lane or create new one
		if assignedLane == -1 {
			// Try to find a recently freed lane
			for lane := 0; lane < nextLane; lane++ {
				if laneAvailable[lane] {
					assignedLane = lane
					laneAvailable[lane] = false
					break
				}
			}
		}

		// Still no lane? Create a new one
		if assignedLane == -1 {
			assignedLane = nextLane
			nextLane++
		}

		commit.Lane = assignedLane
		commitLanes[hash] = assignedLane

		// If this commit is being merged (it's the non-first parent of a merge),
		// its lane can be freed after this point
		for _, childHash := range children {
			childCommit := commits[childHash]
			if len(childCommit.Parents) > 1 {
				for i, parentHash := range childCommit.Parents {
					if i > 0 && parentHash == hash {
						// This branch was merged, lane can be reused
						laneAvailable[assignedLane] = true
					}
				}
			}
		}
	}
}
