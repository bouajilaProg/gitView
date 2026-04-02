package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
)

func TestGraphFixtureComplete(t *testing.T) {
	repoPath := os.Getenv("GITVIEW_FIXTURE_REPO")
	if repoPath == "" {
		t.Skip("GITVIEW_FIXTURE_REPO not set")
	}

	if !filepath.IsAbs(repoPath) {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("get working directory: %v", err)
		}
		repoPath = filepath.Join(filepath.Dir(wd), repoPath)
	}

	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		t.Fatalf("resolve fixture repo path: %v", err)
	}

	repo, err := git.PlainOpen(absPath)
	if err != nil {
		t.Fatalf("open fixture repo: %v", err)
	}

	graph, err := buildGraph(repo)
	if err != nil {
		t.Fatalf("build graph: %v", err)
	}

	// Expected commits (all commits in the fixture)
	expectedCommits := map[string]bool{
		"22bb15f": true, "66c7498": true, "0f04381": true, "73b99f6": true,
		"f7ad22d": true, "02ea37f": true, "8787add": true, "02374a6": true,
		"5af199e": true, "b97bc69": true, "f59a463": true, "3143930": true,
		"7ff6f8a": true, "9865cc1": true, "d502ee3": true, "0d3e23d": true,
		"1eda422": true, "6e179e6": true, "5bdb262": true, "cafe641": true,
		"818ef01": true, "6f71bf6": true, "6b291b7": true, "d0bc942": true,
		"50d0fea": true, "df5c73a": true, "dc79c02": true, "5e890fe": true,
		"738377b": true, "a25aff8": true, "89959ea": true, "6e2929a": true,
		"eb576c0": true, "4e0092c": true, "ab459d4": true, "900fdd0": true,
		"92e5425": true, "6595ae0": true, "f71b25a": true, "7ee7e61": true,
		"2d6f811": true, "53b4baf": true, "5585396": true, "c4fca7f": true,
		"64c93a5": true, "88f4632": true, "c2c6fe8": true, "f2de2b4": true,
		"9f16fdd": true, "db3e4c7": true, "146656f": true, "d1d8e35": true,
		"471d5e5": true, "4f6f011": true, "7aedaa5": true, "084c34d": true,
		"8245a62": true, "89e88ff": true, "9dddd0b": true, "f6a74b3": true,
		"c739130": true, "901f92b": true, "1eb49da": true, "91d946f": true,
		"8ae86f1": true, "0ca4b43": true, "810dccf": true, "647b270": true,
		"1424da8": true, "5bac58a": true, "3c6a936": true, "839dbb1": true,
		"2631423": true, "52dfa26": true, "7f937cc": true, "c7490fa": true,
		"a4eba94": true, "b88fa8c": true, "f1b8074": true, "d65e229": true,
		"e52d94a": true, "f275a7b": true, "56c29d3": true, "be8d2c3": true,
		"238f15e": true, "92b2dc5": true, "8eb7ee6": true, "0613371": true,
		"8c7ab96": true,
	}

	// Check node count
	if len(graph.Nodes) != len(expectedCommits) {
		t.Fatalf("unexpected node count: got %d want %d", len(graph.Nodes), len(expectedCommits))
	}

	// Check all expected commits exist
	nodeMap := make(map[string]Node)
	for _, node := range graph.Nodes {
		nodeMap[node.ID] = node
		if !expectedCommits[node.ID] {
			t.Errorf("unexpected commit in graph: %s", node.ID)
		}
	}

	for id := range expectedCommits {
		if _, exists := nodeMap[id]; !exists {
			t.Errorf("missing commit in graph: %s", id)
		}
	}

	// Check that main branch commits are on lane 0
	mainCommits := []string{
		"22bb15f", "66c7498", "eb576c0", "4e0092c", "ab459d4", "900fdd0",
		"92e5425", "6595ae0", "5585396", "c4fca7f", "9f16fdd", "db3e4c7",
		"7aedaa5", "084c34d", "c739130", "901f92b", "1eb49da", "91d946f",
		"8ae86f1", "0ca4b43", "1424da8", "5bac58a", "2631423", "52dfa26",
		"a4eba94", "b88fa8c", "e52d94a", "f275a7b", "238f15e", "92b2dc5",
		"8c7ab96",
	}

	for _, id := range mainCommits {
		if node, exists := nodeMap[id]; exists {
			if node.Lane != 0 {
				t.Errorf("main commit %s should be on lane 0, got lane %d", id, node.Lane)
			}
		}
	}

	// Check that lanes exist and have proper structure
	if len(graph.Lanes) == 0 {
		t.Error("graph should have at least one lane")
	}

	// Check lane 0 is main
	foundMain := false
	for _, lane := range graph.Lanes {
		if lane.Index == 0 && lane.Name == "main" {
			foundMain = true
			break
		}
	}
	if !foundMain {
		t.Error("lane 0 should be named 'main'")
	}

	// Check edges exist
	if len(graph.Edges) == 0 {
		t.Error("graph should have edges")
	}

	// Verify edge connectivity - each edge should reference valid nodes
	for _, edge := range graph.Edges {
		if _, exists := nodeMap[edge.Source]; !exists {
			t.Errorf("edge source %s not found in nodes", edge.Source)
		}
		if _, exists := nodeMap[edge.Target]; !exists {
			t.Errorf("edge target %s not found in nodes", edge.Target)
		}
	}

	// Check that branch commits are NOT on lane 0
	branchCommits := []string{
		// epic-A
		"0f04381", "73b99f6", "5af199e", "b97bc69", "f59a463", "3143930",
		"7ff6f8a", "9865cc1", "d502ee3", "0d3e23d",
		// epic-B
		"8787add", "02374a6", "1eda422", "6e179e6", "5bdb262", "cafe641",
		"818ef01", "6f71bf6", "6b291b7", "d0bc942",
		// epic-C
		"f7ad22d", "02ea37f", "50d0fea", "df5c73a", "dc79c02", "5e890fe",
		"738377b", "a25aff8", "89959ea", "6e2929a",
		// l2 branches
		"f71b25a", "7ee7e61", "2d6f811", "53b4baf",
		"64c93a5", "88f4632", "c2c6fe8", "f2de2b4",
		"146656f", "d1d8e35", "471d5e5", "4f6f011",
		"8245a62", "89e88ff", "9dddd0b", "f6a74b3",
		// l1 branches
		"810dccf", "647b270", "3c6a936", "839dbb1",
		"7f937cc", "c7490fa", "f1b8074", "d65e229",
		"56c29d3", "be8d2c3", "8eb7ee6", "0613371",
	}

	for _, id := range branchCommits {
		if node, exists := nodeMap[id]; exists {
			if node.Lane == 0 {
				t.Errorf("branch commit %s should NOT be on lane 0, got lane %d", id, node.Lane)
			}
		}
	}
}
