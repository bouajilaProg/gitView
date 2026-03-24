package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"

	"github.com/go-git/go-git/v5"
)

//go:embed viewer/*
var staticFiles embed.FS

func registerRoutes() {
	webFS, err := fs.Sub(staticFiles, "viewer")
	if err != nil {
		log.Fatal("Failed to load web assets: ", err)
	}

	http.Handle("/", http.FileServer(http.FS(webFS)))
	http.HandleFunc("/api/graph", handleGraph)
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
