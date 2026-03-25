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
	assignLanes(commits, sortedOrder)

	// Map lane names from branch tips
	branchLaneNames := make(map[int][]string)
	headName := ""
	if ref.Name().IsBranch() {
		headName = ref.Name().Short()
	}

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

	mergedBranchByCommit := mergedBranchNamesFromCommits(commits)
	if reflogMerged := mergedBranchNamesFromReflog(); len(reflogMerged) > 0 {
		for hash, name := range reflogMerged {
			if _, exists := mergedBranchByCommit[hash]; !exists {
				mergedBranchByCommit[hash] = name
			}
		}
	}

	applyMergedBranchFallbacks(commits, mergedBranchByCommit, commitRefs, branchLaneNames)

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
			ID:      c.Hash[:7],
			Hash:    c.Hash,
			Message: c.Message,
			Author:  c.Author,
			Date:    c.Date.Format(time.RFC3339),
			Files:   c.Files,
			Lane:    c.Lane,
			Refs:    commitRefs[c.Hash],
		}
		graph.Nodes = append(graph.Nodes, node)

		// Add edges to parents
		for index, parent := range c.Parents {
			if parentCommit, exists := commits[parent]; exists {
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
			graph.Lanes = append(graph.Lanes, Lane{Index: lane, Name: laneName})
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
		name := strings.TrimSpace(matches[1])
		if name != "" {
			return name
		}
	}
	return ""
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
			if len(commitRefs[mergeHash]) == 0 {
				commitRefs[mergeHash] = []string{branchName}
				lane := mergeCommit.Lane
				if lane >= 0 && !containsString(branchLaneNames[lane], branchName) {
					branchLaneNames[lane] = append(branchLaneNames[lane], branchName)
				}
			}
			continue
		}

		mergedParent := mergeCommit.Parents[1]
		if len(commitRefs[mergedParent]) == 0 {
			commitRefs[mergedParent] = []string{branchName}
			if commit, ok := commits[mergedParent]; ok {
				lane := commit.Lane
				if lane >= 0 && !containsString(branchLaneNames[lane], branchName) {
					branchLaneNames[lane] = append(branchLaneNames[lane], branchName)
				}
			}
		}
	}
}
