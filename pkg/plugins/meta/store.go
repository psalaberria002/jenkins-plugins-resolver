package meta

import (
	"fmt"
	"path/filepath"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
)

// GetStorePath returns the path to the store
func GetStorePath(workingDir string) string {
	return filepath.Join(workingDir, "meta")
}

// GetMetaPath returns the path to the plugin metadata in the store
func GetMetaPath(p *api.Plugin, workingDir string) string {
	return filepath.Join(GetStorePath(workingDir), fmt.Sprintf("%s.meta", p.Filename()))
}
