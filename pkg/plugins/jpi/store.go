package jpi

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/juju/errors"
)

// GetStorePath returns the path to the store
func GetStorePath(workingDir string) string {
	return filepath.Join(workingDir, "jpi")
}

// GetPluginPath returns the path to the plugin in the store
func GetPluginPath(p *api.Plugin, workingDir string) string {
	return filepath.Join(GetStorePath(workingDir), fmt.Sprintf("%s.jpi", p.Filename()))
}

// EnsureStorePathExists will create the store path directory if it does not exist
func EnsureStorePathExists(workingDir string) error {
	if err := os.MkdirAll(GetStorePath(workingDir), 0777); err != nil {
		return errors.Errorf("unable to create the store path: %+v", err)
	}
	return nil
}
