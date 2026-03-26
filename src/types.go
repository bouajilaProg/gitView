package main

import "time"

// FileStat represents a changed file and its status
type FileStat struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Node represents a commit node in the graph
type Node struct {
	ID              string     `json:"id"`
	Hash            string     `json:"hash"`
	Message         string     `json:"message"`
	Author          string     `json:"author"`
	Date            string     `json:"date"`
	Files           []FileStat `json:"files"`
	Lane            int        `json:"lane"`
	Refs            []string   `json:"refs"`
	PredictedBranch string     `json:"predictedBranch"`
}

// Edge represents a connection between commits
type Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Color  int    `json:"colorLane"`
}

// Graph represents the complete commit graph
type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
	Lanes []Lane `json:"lanes"`
}

// Lane represents a branch lane label
type Lane struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
}

// CommitData holds intermediate commit information
type CommitData struct {
	Hash      string
	Message   string
	Author    string
	Date      time.Time
	Parents   []string
	Files     []FileStat
	Lane      int
	Processed bool
}
