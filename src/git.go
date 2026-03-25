package main

import (
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
)

func getChangedFiles(c *object.Commit) []FileStat {
	var files []FileStat

	// Get parent commit for diff
	parent, err := c.Parent(0)
	if err != nil {
		// No parent (initial commit), get all files from tree
		tree, err := c.Tree()
		if err != nil {
			return files
		}
		tree.Files().ForEach(func(f *object.File) error {
			files = append(files, FileStat{Name: f.Name, Status: "A"})
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
		action, err := change.Action()
		if err != nil {
			continue
		}
		status := "M"
		switch action {
		case merkletrie.Insert:
			status = "A"
		case merkletrie.Delete:
			status = "D"
		case merkletrie.Modify:
			status = "M"
		}
		name := change.To.Name
		if name == "" {
			name = change.From.Name
		}
		files = append(files, FileStat{Name: name, Status: status})
	}

	return files
}
