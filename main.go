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

//go:embed src/index.html
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
	content, err := staticFiles.ReadFile("src/index.html")
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
