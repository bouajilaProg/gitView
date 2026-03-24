package main

import (
	"container/heap"
)

// PriorityQueue implements heap.Interface and holds Commits
type PriorityQueue struct {
	hashes  []string
	commits map[string]*CommitData
}

func (pq PriorityQueue) Len() int { return len(pq.hashes) }
func (pq PriorityQueue) Less(i, j int) bool {
	// Same sorting logic: Newer date first, then lexicographical hash
	iDate, jDate := pq.commits[pq.hashes[i]].Date, pq.commits[pq.hashes[j]].Date
	if iDate.Equal(jDate) {
		return pq.hashes[i] < pq.hashes[j]
	}
	return iDate.After(jDate)
}
func (pq PriorityQueue) Swap(i, j int)       { pq.hashes[i], pq.hashes[j] = pq.hashes[j], pq.hashes[i] }
func (pq *PriorityQueue) Push(x interface{}) { pq.hashes = append(pq.hashes, x.(string)) }
func (pq *PriorityQueue) Pop() interface{} {
	old := pq.hashes
	n := len(old)
	item := old[n-1]
	pq.hashes = old[0 : n-1]
	return item
}

func topologicalSortOptimized(commits map[string]*CommitData) []string {
	inDegree := make(map[string]int)
	for _, commit := range commits {
		for _, parent := range commit.Parents {
			if _, exists := commits[parent]; exists {
				inDegree[parent]++
			}
		}
	}

	// Initialize the Priority Queue with "leaf" commits
	pq := &PriorityQueue{commits: commits}
	heap.Init(pq)
	for hash := range commits {
		if inDegree[hash] == 0 {
			heap.Push(pq, hash)
		}
	}

	var sorted []string
	for pq.Len() > 0 {
		// Pop the "best" commit based on date (O(log n) instead of O(n log n))
		current := heap.Pop(pq).(string)
		sorted = append(sorted, current)

		for _, parent := range commits[current].Parents {
			if _, exists := commits[parent]; exists {
				inDegree[parent]--
				if inDegree[parent] == 0 {
					heap.Push(pq, parent) // Auto-sorts on insert
				}
			}
		}
	}
	return sorted
}

// topologicalSort keeps the legacy signature for callers.
func topologicalSort(commits map[string]*CommitData, initialOrder []string) []string {
	return topologicalSortOptimized(commits)
}
