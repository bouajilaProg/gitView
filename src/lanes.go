package main

import "sort"

// assignLanes assigns lane indices to commits for visual layout
// Rules:
// 1. Main branch (first parents) stays on lane 0
// 2. Each merged branch segment gets its own unique lane (no reuse)
// 3. Active (unmerged) branches get dedicated lanes
func assignLanes(commits map[string]*CommitData, sortedOrder []string) map[int]string {
	// Build child map: parent -> list of children
	childrenOf := make(map[string][]string)
	for hash, commit := range commits {
		for _, parentHash := range commit.Parents {
			if _, exists := commits[parentHash]; exists {
				childrenOf[parentHash] = append(childrenOf[parentHash], hash)
			}
		}
	}

	// Sort children by date (newest first) for consistent lane inheritance
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

	// Identify commits that are second+ parents of merge commits
	// These are the "tips" of merged branches and should start new lanes
	isMergeTip := make(map[string]bool)
	for _, commit := range commits {
		if len(commit.Parents) > 1 {
			for i := 1; i < len(commit.Parents); i++ {
				isMergeTip[commit.Parents[i]] = true
			}
		}
	}

	// Track lane assignments
	commitLanes := make(map[string]int) // commit hash -> lane
	laneTypes := make(map[int]string)   // lane -> "dedicated" or "merged"
	nextLane := 0

	// Determine if a commit is on an unmerged branch (has no children = branch tip)
	hasNoChildren := make(map[string]bool)
	for _, hash := range sortedOrder {
		if len(childrenOf[hash]) == 0 {
			hasNoChildren[hash] = true
		}
	}

	// Process commits from newest to oldest (topological order)
	for _, hash := range sortedOrder {
		commit := commits[hash]
		assignedLane := -1

		// Determine if this is an unmerged branch tip
		isUnmergedTip := hasNoChildren[hash]

		// If this commit is a merge tip (second parent of a merge), it starts a new branch segment
		// Always get a NEW lane - no reuse
		if isMergeTip[hash] {
			assignedLane = nextLane
			nextLane++
			laneTypes[assignedLane] = "merged"
		} else {
			// Check if any child already assigned a lane to us (lane inheritance)
			children := childrenOf[hash]
			for _, childHash := range children {
				if childLane, exists := commitLanes[childHash]; exists {
					childCommit := commits[childHash]
					// If the child has only one parent, inherit its lane
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

			// Still no lane? Create a new one
			if assignedLane == -1 {
				assignedLane = nextLane
				nextLane++
				if isUnmergedTip {
					laneTypes[assignedLane] = "dedicated"
				} else {
					laneTypes[assignedLane] = "merged"
				}
			}
		}

		commit.Lane = assignedLane
		commitLanes[hash] = assignedLane
	}

	return laneTypes
}
