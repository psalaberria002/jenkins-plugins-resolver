package common

import (
	"os"

	"github.com/juju/errors"
)

// EnsureStorePathExists creates the store directory for a given path generator
// If the store directory exists, does not thing.
func EnsureStorePathExists(workingDir string, pathGen func(string) string) error {
	if err := os.MkdirAll(pathGen(workingDir), 0777); err != nil {
		return errors.Errorf("unable to create the store path: %+v", err)
	}
	return nil
}
