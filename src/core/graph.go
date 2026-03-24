package main

import (
	"sort"
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
	branchLaneNames := make(map[int]string)
	headName := ""
	if ref.Name().IsBranch() {
		headName = ref.Name().Short()
	}

	branchIter, err := repo.Branches()
	if err == nil {
		_ = branchIter.ForEach(func(branchRef *plumbing.Reference) error {
			branchName := branchRef.Name().Short()
			branchHash := branchRef.Hash().String()
			if commit, exists := commits[branchHash]; exists {
				lane := commit.Lane
				if branchName == headName {
					branchLaneNames[lane] = branchName
				} else if _, exists := branchLaneNames[lane]; !exists {
					branchLaneNames[lane] = branchName
				}
			}
			return nil
		})
	}

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
			Message: c.Message,
			Author:  c.Author,
			Date:    c.Date.Format(time.RFC3339),
			Files:   c.Files,
			Lane:    c.Lane,
		}
		graph.Nodes = append(graph.Nodes, node)

		// Add edges to parents
		for _, parent := range c.Parents {
			if _, exists := commits[parent]; exists {
				graph.Edges = append(graph.Edges, Edge{
					Source: c.Hash[:7],
					Target: parent[:7],
				})
			}
		}
	}

	if maxLane >= 0 {
		for lane := 0; lane <= maxLane; lane++ {
			laneName := branchLaneNames[lane]
			graph.Lanes = append(graph.Lanes, Lane{Index: lane, Name: laneName})
		}
	}

	return graph, nil
}
