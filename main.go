package main

import (
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

//go:embed assets/index.html
var staticFiles embed.FS

// Node represents a commit node in the graph
type Node struct {
	ID      string   `json:"id"`
	Message string   `json:"message"`
	Author  string   `json:"author"`
	Date    string   `json:"date"`
	Files   []string `json:"files"`
	Lane    int      `json:"lane"`
}

// Edge represents a connection between commits
type Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// Graph represents the complete commit graph
type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// CommitData holds intermediate commit information
type CommitData struct {
	Hash      string
	Message   string
	Author    string
	Date      time.Time
	Parents   []string
	Files     []string
	Lane      int
	Processed bool
}

func main() {
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/api/graph", handleGraph)

	log.Println("Git-Canvas-Go running on http://localhost:6060")
	log.Fatal(http.ListenAndServe(":6060", nil))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	content, err := staticFiles.ReadFile("assets/index.html")
	if err != nil {
		http.Error(w, "Failed to read index.html", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(content)
}

func handleGraph(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	repo, err := git.PlainOpen(".")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to open repository: " + err.Error()})
		return
	}

	graph, err := buildGraph(repo)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to build graph: " + err.Error()})
		return
	}

	json.NewEncoder(w).Encode(graph)
}

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

	// Sort commits by date (newest first) for topological ordering
	sort.Slice(commitOrder, func(i, j int) bool {
		return commits[commitOrder[i]].Date.After(commits[commitOrder[j]].Date)
	})

	// Perform topological sort and assign lanes
	sortedOrder := topologicalSort(commits, commitOrder)
	assignLanes(commits, sortedOrder)

	// Build the graph structure
	graph := &Graph{
		Nodes: make([]Node, 0, len(sortedOrder)),
		Edges: make([]Edge, 0),
	}

	for _, hash := range sortedOrder {
		c := commits[hash]
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

	return graph, nil
}

func getChangedFiles(c *object.Commit) []string {
	var files []string

	// Get parent commit for diff
	parent, err := c.Parent(0)
	if err != nil {
		// No parent (initial commit), get all files from tree
		tree, err := c.Tree()
		if err != nil {
			return files
		}
		tree.Files().ForEach(func(f *object.File) error {
			files = append(files, f.Name)
			return nil
		})
		return files
	}

	// Get diff between parent and current commit
	parentTree, err := parent.Tree()
	if err != nil {
		return files
	}

	currentTree, err := c.Tree()
	if err != nil {
		return files
	}

	changes, err := parentTree.Diff(currentTree)
	if err != nil {
		return files
	}

	for _, change := range changes {
		name := change.To.Name
		if name == "" {
			name = change.From.Name
		}
		files = append(files, name)
	}

	return files
}

// topologicalSort performs Kahn's algorithm for topological ordering
func topologicalSort(commits map[string]*CommitData, initialOrder []string) []string {
	// Build in-degree map (children count for each commit)
	inDegree := make(map[string]int)
	children := make(map[string][]string)

	for hash := range commits {
		inDegree[hash] = 0
	}

	// Build reverse graph (child -> parent becomes parent -> child for processing)
	for hash, commit := range commits {
		for _, parent := range commit.Parents {
			if _, exists := commits[parent]; exists {
				children[parent] = append(children[parent], hash)
				inDegree[hash]++
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
		return commits[queue[i]].Date.After(commits[queue[j]].Date)
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
						return commits[queue[i]].Date.After(commits[queue[j]].Date)
					})
				}
			}
		}
	}

	return sorted
}

// assignLanes assigns lane indices to commits for visual layout
func assignLanes(commits map[string]*CommitData, sortedOrder []string) {
	// Track which lanes are currently occupied and by which branch
	activeLanes := make(map[int]string) // lane -> last commit hash using it
	commitLanes := make(map[string]int) // commit hash -> lane
	nextLane := 0

	for _, hash := range sortedOrder {
		commit := commits[hash]
		assignedLane := -1

		// Check if any parent is in the sorted order and has a lane
		for _, parentHash := range commit.Parents {
			if parentLane, exists := commitLanes[parentHash]; exists {
				// If this commit continues from a single parent, try to use its lane
				if len(commit.Parents) == 1 {
					assignedLane = parentLane
					break
				}
			}
		}

		// If no lane inherited, find the first available lane
		if assignedLane == -1 {
			// Look for an available lane
			for lane := 0; lane < nextLane; lane++ {
				lastCommit, occupied := activeLanes[lane]
				if !occupied {
					assignedLane = lane
					break
				}
				// Check if the lane's last commit is a parent of this commit
				for _, parentHash := range commit.Parents {
					if lastCommit == parentHash {
						assignedLane = lane
						break
					}
				}
				if assignedLane != -1 {
					break
				}
			}
		}

		// If still no lane, create a new one
		if assignedLane == -1 {
			assignedLane = nextLane
			nextLane++
		}

		commit.Lane = assignedLane
		commitLanes[hash] = assignedLane
		activeLanes[assignedLane] = hash

		// For merge commits, mark the secondary parent lanes as potentially available
		if len(commit.Parents) > 1 {
			for i, parentHash := range commit.Parents {
				if i > 0 {
					if parentLane, exists := commitLanes[parentHash]; exists {
						// The branch being merged in can free up its lane
						delete(activeLanes, parentLane)
					}
				}
			}
		}
	}
}
