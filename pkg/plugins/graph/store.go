package graph

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/juju/errors"
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

// EnsureStorePathExists will create the store path directory if it does not exist
func EnsureStorePathExists(workingDir string) error {
	if err := os.MkdirAll(GetStorePath(workingDir), 0777); err != nil {
		return errors.Errorf("unable to create the store path: %+v", err)
	}
	return nil
}
