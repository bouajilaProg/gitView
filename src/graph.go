package main

import (
	"bufio"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func buildGraph(repo *git.Repository) (*Graph, error) {
	commits := make(map[string]*CommitData)
	var commitOrder []string

	// Get HEAD reference
	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	// Get commit iterator
	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash(), All: true})
	if err != nil {
		return nil, err
	}

	// Collect all commits
	err = commitIter.ForEach(func(c *object.Commit) error {
		hash := c.Hash.String()

		// Get parent hashes
		var parents []string
		for _, p := range c.ParentHashes {
			parents = append(parents, p.String())
		}

		// Get files changed in this commit
		files := getChangedFiles(c)

		commits[hash] = &CommitData{
			Hash:    hash,
			Message: c.Message,
			Author:  c.Author.Name,
			Date:    c.Author.When,
			Parents: parents,
			Files:   files,
			Lane:    -1,
		}
		commitOrder = append(commitOrder, hash)

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Sort commits by date (newest first) for deterministic ordering
	sort.Slice(commitOrder, func(i, j int) bool {
		left := commits[commitOrder[i]].Date
		right := commits[commitOrder[j]].Date
		if left.Equal(right) {
			return commitOrder[i] < commitOrder[j]
		}
		return left.After(right)
	})

	// Perform topological sort and assign lanes
	sortedOrder := topologicalSort(commits, commitOrder)
	laneTypes := assignLanes(commits, sortedOrder)

	// Map lane names from branch tips
	branchLaneNames := make(map[int][]string)
	headName := ""
	if ref.Name().IsBranch() {
		headName = ref.Name().Short()
	}

	// Get branch creation times for filtering
	branchCreationTimes := getBranchCreationTimes(repo)

	commitRefs := make(map[string][]string)
	branchIter, err := repo.Branches()
	if err == nil {
		_ = branchIter.ForEach(func(branchRef *plumbing.Reference) error {
			branchName := branchRef.Name().Short()
			branchHash := branchRef.Hash().String()

			commitRefs[branchHash] = append(commitRefs[branchHash], branchName)

			if commit, exists := commits[branchHash]; exists {
				lane := commit.Lane
				names := branchLaneNames[lane]
				if !containsString(names, branchName) {
					branchLaneNames[lane] = append(names, branchName)
				}
			}
			return nil
		})
	}

	// Filter lane branch names: remove branches created after the oldest commit in that lane
	filterLaneBranchNames(commits, sortedOrder, branchLaneNames, branchCreationTimes)

	mergedBranchByCommit := mergedBranchNamesFromCommits(commits)
	if reflogMerged := mergedBranchNamesFromReflog(); len(reflogMerged) > 0 {
		for hash, name := range reflogMerged {
			if _, exists := mergedBranchByCommit[hash]; !exists {
				mergedBranchByCommit[hash] = name
			}
		}
	}

	applyMergedBranchFallbacks(commits, mergedBranchByCommit, commitRefs, branchLaneNames)
	predictedBranches := predictMergedBranchRefs(commits, mergedBranchByCommit)

	// Build the graph structure
	graph := &Graph{
		Nodes: make([]Node, 0, len(sortedOrder)),
		Edges: make([]Edge, 0),
		Lanes: make([]Lane, 0),
	}

	maxLane := -1

	for _, hash := range sortedOrder {
		c := commits[hash]
		if c.Lane > maxLane {
			maxLane = c.Lane
		}
		node := Node{
			ID:              c.Hash[:7],
			Hash:            c.Hash,
			Message:         c.Message,
			Author:          c.Author,
			Date:            c.Date.Format(time.RFC3339),
			Files:           c.Files,
			Lane:            c.Lane,
			Refs:            commitRefs[c.Hash],
			PredictedBranch: predictedBranches[c.Hash],
		}
		graph.Nodes = append(graph.Nodes, node)

		// Add edges to parents
		// Skip edges from dedicated (unmerged) lanes to different lanes
		// unless this commit has multiple parents (is a merge commit)
		for index, parent := range c.Parents {
			if parentCommit, exists := commits[parent]; exists {
				// If on a dedicated lane and parent is on a different lane,
				// only draw edge if this is a merge commit (has multiple parents)
				if laneTypes[c.Lane] == "dedicated" && parentCommit.Lane != c.Lane {
					if len(c.Parents) == 1 {
						// Skip edge - unmerged branch pointing back to main
						continue
					}
				}

				colorLane := c.Lane
				if index > 0 {
					colorLane = parentCommit.Lane
				}
				graph.Edges = append(graph.Edges, Edge{
					Source: c.Hash[:7],
					Target: parent[:7],
					Color:  colorLane,
				})
			}
		}
	}

	if maxLane >= 0 {
		for lane := 0; lane <= maxLane; lane++ {
			names := append([]string{}, branchLaneNames[lane]...)
			if len(names) > 0 {
				names = orderBranchNames(names, headName)
			}
			laneName := strings.Join(names, ", ")
			lType := "dedicated"
			if t, ok := laneTypes[lane]; ok && t != "" {
				lType = t
			}
			graph.Lanes = append(graph.Lanes, Lane{Index: lane, Name: laneName, Type: lType})
		}
	}

	return graph, nil
}

func containsString(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func orderBranchNames(names []string, headName string) []string {
	unique := make([]string, 0, len(names))
	for _, name := range names {
		if !containsString(unique, name) {
			unique = append(unique, name)
		}
	}
	if headName == "" || !containsString(unique, headName) {
		sort.Strings(unique)
		return unique
	}
	remaining := make([]string, 0, len(unique)-1)
	for _, name := range unique {
		if name != headName {
			remaining = append(remaining, name)
		}
	}
	sort.Strings(remaining)
	return append([]string{headName}, remaining...)
}

func mergedBranchNamesFromCommits(commits map[string]*CommitData) map[string]string {
	result := make(map[string]string)
	for hash, commit := range commits {
		if len(commit.Parents) < 2 {
			continue
		}
		name := parseMergedBranchName(commit.Message)
		if name == "" {
			continue
		}
		result[hash] = name
	}
	return result
}

func mergedBranchNamesFromReflog() map[string]string {
	result := make(map[string]string)
	file, err := os.Open(".git/logs/HEAD")
	if err != nil {
		return result
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) < 2 {
			continue
		}
		metaFields := strings.Fields(parts[0])
		if len(metaFields) < 2 {
			continue
		}
		newHash := metaFields[1]
		message := parts[1]
		name := parseMergedBranchName(message)
		if name == "" {
			continue
		}
		result[newHash] = name
	}

	return result
}

// getBranchCreationTimes returns a map of branch name -> creation time.
// It reads from branch-specific reflogs to find when each branch was created.
// For main/master branches or branches created via clone, we don't set a creation
// time, which means they won't be filtered out (treated as "always existed").
func getBranchCreationTimes(repo *git.Repository) map[string]time.Time {
	result := make(map[string]time.Time)

	branchIter, err := repo.Branches()
	if err != nil {
		return result
	}

	_ = branchIter.ForEach(func(branchRef *plumbing.Reference) error {
		branchName := branchRef.Name().Short()

		// main/master are the default branches - treat them as "always existed"
		if branchName == "main" || branchName == "master" {
			// Don't add to result - they will be kept by filterLaneBranchNames
			return nil
		}

		// Try to get creation time from branch-specific reflog
		reflogPath := ".git/logs/refs/heads/" + branchName
		if creationTime, isRealCreation := getBranchCreationFromReflog(reflogPath); isRealCreation {
			result[branchName] = creationTime
			return nil
		}

		// If we couldn't determine creation time, don't add to result (keep the branch)
		return nil
	})

	return result
}

// getBranchCreationFromReflog reads a branch reflog and returns the timestamp
// of the first entry if it represents a real branch creation (not a clone/fetch).
// A real branch creation has:
// - First hash is all zeros (0000000...)
// - Message contains "branch" or "checkout" but not "clone"
func getBranchCreationFromReflog(reflogPath string) (time.Time, bool) {
	file, err := os.Open(reflogPath)
	if err != nil {
		return time.Time{}, false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		// Reflog format: <old-hash> <new-hash> <author> <email> <timestamp> <timezone>\t<message>
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) < 2 {
			return time.Time{}, false
		}

		metaFields := strings.Fields(parts[0])
		message := strings.ToLower(parts[1])

		// Check if this is a real branch creation (first hash is zeros)
		// and not a clone operation
		if len(metaFields) < 2 {
			return time.Time{}, false
		}

		firstHash := metaFields[0]
		isNewBranch := strings.HasPrefix(firstHash, "0000000")
		isClone := strings.Contains(message, "clone")

		// Only consider it a branch creation if it starts from zeros and isn't a clone
		if !isNewBranch || isClone {
			return time.Time{}, false
		}

		// Find the timestamp (Unix timestamp before timezone)
		for i := len(metaFields) - 1; i >= 0; i-- {
			// Timezone is like +0000 or -0500
			if len(metaFields[i]) == 5 && (metaFields[i][0] == '+' || metaFields[i][0] == '-') {
				if i > 0 {
					var timestamp int64
					if parseUnixTimestamp(metaFields[i-1], &timestamp) {
						return time.Unix(timestamp, 0), true
					}
				}
			}
		}
	}
	return time.Time{}, false
}

// parseUnixTimestamp parses a string as Unix timestamp
func parseUnixTimestamp(s string, result *int64) bool {
	var ts int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
		ts = ts*10 + int64(c-'0')
	}
	*result = ts
	return true
}

// filterLaneBranchNames filters branch names per lane so only branches
// that were created before or at the same time as the oldest commit in that lane are shown.
func filterLaneBranchNames(
	commits map[string]*CommitData,
	sortedOrder []string,
	branchLaneNames map[int][]string,
	branchCreationTimes map[string]time.Time,
) {
	// Find the oldest commit date in each lane
	oldestCommitInLane := make(map[int]time.Time)
	for _, hash := range sortedOrder {
		commit, exists := commits[hash]
		if !exists {
			continue
		}
		lane := commit.Lane
		if oldest, exists := oldestCommitInLane[lane]; !exists || commit.Date.Before(oldest) {
			oldestCommitInLane[lane] = commit.Date
		}
	}

	// Filter branch names for each lane
	for lane, branchNames := range branchLaneNames {
		oldestCommit, hasOldest := oldestCommitInLane[lane]
		if !hasOldest {
			continue
		}

		filtered := make([]string, 0, len(branchNames))
		for _, branchName := range branchNames {
			branchCreation, hasBranchTime := branchCreationTimes[branchName]
			// Keep branch if:
			// 1. We don't know when it was created (keep to be safe)
			// 2. It was created before or at the same time as the oldest commit in the lane
			if !hasBranchTime || !branchCreation.After(oldestCommit) {
				filtered = append(filtered, branchName)
			}
		}
		branchLaneNames[lane] = filtered
	}
}

func parseMergedBranchName(message string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)merge pull request #\d+ from [^/]+/([^\s]+)`),
		regexp.MustCompile(`(?i)merge branch ['\"]([^'\"]+)['\"]`),
		regexp.MustCompile(`(?i)merge remote-tracking branch ['\"]([^'\"]+)['\"]`),
		regexp.MustCompile(`(?i)^merge\s+([^:]+)`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(strings.TrimSpace(message))
		if len(matches) < 2 {
			continue
		}
		name := sanitizeMergedBranchName(matches[1])
		if name != "" {
			return name
		}
	}
	return ""
}

func sanitizeMergedBranchName(value string) string {
	name := strings.TrimSpace(value)
	name = strings.TrimPrefix(name, "refs/heads/")
	name = strings.TrimPrefix(name, "refs/remotes/")
	name = strings.TrimPrefix(name, "origin/")

	upper := strings.ToUpper(name)
	if upper == "HEAD" || strings.HasPrefix(upper, "HEAD ") || strings.HasPrefix(upper, "HEAD-") || strings.HasPrefix(upper, "HEAD@") {
		return ""
	}

	return name
}

func applyMergedBranchFallbacks(commits map[string]*CommitData, mergedBranches map[string]string, commitRefs map[string][]string, branchLaneNames map[int][]string) {
	if len(mergedBranches) == 0 {
		return
	}

	for mergeHash, branchName := range mergedBranches {
		mergeCommit, exists := commits[mergeHash]
		if !exists {
			continue
		}
		if len(mergeCommit.Parents) < 2 {
			continue
		}

		// Only assign the branch name to the merged parent (pre-merge commit)
		// Don't add to branchLaneNames to avoid showing on all commits in that lane
		mergedParent := mergeCommit.Parents[1]
		if len(commitRefs[mergedParent]) == 0 {
			commitRefs[mergedParent] = []string{branchName}
		}
	}
}

func predictMergedBranchRefs(commits map[string]*CommitData, mergedBranches map[string]string) map[string]string {
	result := make(map[string]string)
	if len(mergedBranches) == 0 {
		return result
	}

	for mergeHash, branchName := range mergedBranches {
		mergeCommit, exists := commits[mergeHash]
		if !exists || len(mergeCommit.Parents) < 2 {
			continue
		}

		mainParent := mergeCommit.Parents[0]
		mergedParent := mergeCommit.Parents[1]
		mergeBase := findMergeBase(commits, mainParent, mergedParent)
		if mergeBase == "" {
			continue
		}

		visited := make(map[string]bool)
		stack := []string{mergedParent}
		for len(stack) > 0 {
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if visited[current] {
				continue
			}
			visited[current] = true

			if current == mergeBase {
				continue
			}

			commit, ok := commits[current]
			if !ok {
				continue
			}

			if _, exists := result[current]; !exists {
				result[current] = branchName
			}

			for _, parent := range commit.Parents {
				if !visited[parent] {
					stack = append(stack, parent)
				}
			}
		}
	}

	return result
}

func findMergeBase(commits map[string]*CommitData, mainParent string, mergedParent string) string {
	if mainParent == "" || mergedParent == "" {
		return ""
	}

	mainAncestors := collectAncestors(commits, mainParent)
	if len(mainAncestors) == 0 {
		return ""
	}

	visited := make(map[string]bool)
	queue := []string{mergedParent}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if visited[current] {
			continue
		}
		visited[current] = true

		if _, ok := mainAncestors[current]; ok {
			return current
		}

		commit, ok := commits[current]
		if !ok {
			continue
		}
		for _, parent := range commit.Parents {
			if !visited[parent] {
				queue = append(queue, parent)
			}
		}
	}

	return ""
}

func collectAncestors(commits map[string]*CommitData, start string) map[string]struct{} {
	ancestors := make(map[string]struct{})
	if start == "" {
		return ancestors
	}

	stack := []string{start}
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if _, seen := ancestors[current]; seen {
			continue
		}
		ancestors[current] = struct{}{}
		commit, ok := commits[current]
		if !ok {
			continue
		}
		for _, parent := range commit.Parents {
			if _, seen := ancestors[parent]; !seen {
				stack = append(stack, parent)
			}
		}
	}

	return ancestors
}
