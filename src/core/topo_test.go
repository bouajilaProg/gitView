package main

import (
	"reflect"
	"testing"
	"time"
)

func TestTopologicalSortLinearOrder(t *testing.T) {
	commits := map[string]*CommitData{
		"A": {Hash: "A", Date: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
		"B": {Hash: "B", Parents: []string{"A"}, Date: time.Date(2023, 1, 2, 10, 0, 0, 0, time.UTC)},
		"C": {Hash: "C", Parents: []string{"B"}, Date: time.Date(2023, 1, 3, 10, 0, 0, 0, time.UTC)},
	}

	got := topologicalSortOptimized(commits)
	want := []string{"C", "B", "A"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected order: got %v want %v", got, want)
	}
}

func TestTopologicalSortBranchTieBreak(t *testing.T) {
	sameDate := time.Date(2023, 1, 2, 10, 0, 0, 0, time.UTC)

	commits := map[string]*CommitData{
		"a": {Hash: "a", Date: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
		"b": {Hash: "b", Parents: []string{"a"}, Date: sameDate},
		"c": {Hash: "c", Parents: []string{"a"}, Date: sameDate},
		"d": {Hash: "d", Parents: []string{"b", "c"}, Date: time.Date(2023, 1, 3, 10, 0, 0, 0, time.UTC)},
	}

	got := topologicalSortOptimized(commits)
	want := []string{"d", "b", "c", "a"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected order: got %v want %v", got, want)
	}
}
