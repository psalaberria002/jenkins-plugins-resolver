package jpi

import (
	"fmt"
	"path/filepath"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
)

// GetStorePath returns the path to the store
func GetStorePath(workingDir string) string {
	return filepath.Join(workingDir, "jpi")
}

// GetPluginPath returns the path to the plugin in the store
func GetPluginPath(p *api.Plugin, workingDir string) string {
	return filepath.Join(GetStorePath(workingDir), fmt.Sprintf("%s.jpi", p.Filename()))
}
