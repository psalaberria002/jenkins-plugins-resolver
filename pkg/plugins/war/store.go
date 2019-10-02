package war

import (
	"fmt"
	"path/filepath"
)

// GetStorePath returns the path to the store
func GetStorePath(workingDir string) string {
	return filepath.Join(workingDir, "war")
}

// GetWarPath returns the path to the war in the store
func GetWarPath(jm *JenkinsManifest, workingDir string) string {
	return filepath.Join(GetStorePath(workingDir), fmt.Sprintf("jenkins-%s.war", jm.Version))
}
