package main

import "sort"

// assignLanes assigns lane indices to commits for visual layout
// Rules:
// 1. Main branch (first parents) stays on lane 0
// 2. Each merged branch segment gets its own unique lane (no reuse)
// 3. Active (unmerged) branches get dedicated lanes
// 4. When a branch ends at a commit, no new branch from that commit can reuse its lane
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

	// Track lanes that are "blocked" at each commit (lanes ending there)
	// When a branch ends at commit X with lane L, no other branch from X should use L
	blockedLanesAt := make(map[string]map[int]bool)

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
			children := childrenOf[hash]

			// Get blocked lanes at this commit (lanes from children that end here)
			blocked := blockedLanesAt[hash]

			// PRIORITY 1: If we are the first parent of a merge commit, inherit its lane
			// This maintains main branch continuity through merges
			for _, childHash := range children {
				if childLane, exists := commitLanes[childHash]; exists {
					childCommit := commits[childHash]
					if len(childCommit.Parents) > 1 && childCommit.Parents[0] == hash {
						assignedLane = childLane
						break
					}
				}
			}

			// PRIORITY 2: If we are the first parent of ANY child (including single-parent)
			// This handles linear chains
			if assignedLane == -1 {
				for _, childHash := range children {
					if childLane, exists := commitLanes[childHash]; exists {
						childCommit := commits[childHash]
						if len(childCommit.Parents) > 0 && childCommit.Parents[0] == hash {
							// Don't inherit if this lane is blocked (another branch ends here)
							if blocked != nil && blocked[childLane] {
								continue
							}
							assignedLane = childLane
							break
						}
					}
				}
			}

			// PRIORITY 3: Inherit from any single-parent child (fallback)
			if assignedLane == -1 {
				for _, childHash := range children {
					if childLane, exists := commitLanes[childHash]; exists {
						childCommit := commits[childHash]
						if len(childCommit.Parents) == 1 {
							// Don't inherit if this lane is blocked
							if blocked != nil && blocked[childLane] {
								continue
							}
							assignedLane = childLane
							break
						}
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

		// Mark lanes as blocked at parent commits when this is not their first parent
		// This prevents color reuse when multiple branches fork from the same commit
		for i, parentHash := range commit.Parents {
			if i > 0 {
				// This commit is a non-first child of its parent
				// Block this lane at the parent so siblings don't reuse it
				if blockedLanesAt[parentHash] == nil {
					blockedLanesAt[parentHash] = make(map[int]bool)
				}
				blockedLanesAt[parentHash][assignedLane] = true
			}
		}
	}

	return laneTypes
}
