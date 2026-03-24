package main

import "github.com/go-git/go-git/v5/plumbing/object"

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
