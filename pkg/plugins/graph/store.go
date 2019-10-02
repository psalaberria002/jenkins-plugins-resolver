package graph

import (
	"fmt"
	"path/filepath"
)

// GetStorePath returns the path to the store
func GetStorePath(workingDir string) string {
	return filepath.Join(workingDir, "graph")
}

// GetGraphPath returns the path to the plugin graph in the store
func GetGraphPath(hash string, workingDir string) string {
	// We are returning the SHA to name the graph files (.jenkins/graphs/<sha>.json)
	// We use this system as cache, as we are dealing with static plugins which dependencies
	// will remain the same over time (thankfully). Therefore, by using the hash of the input
	// file for the graph filename we grant reproducibility.
	return filepath.Join(GetStorePath(workingDir), fmt.Sprintf("%s.graph", hash))
}
