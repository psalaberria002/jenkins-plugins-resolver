package jpi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/utils"
)

func TestEnsureStorePathExists(t *testing.T) {
	// Create folder first to get an unique name
	folder := filepath.Join(os.TempDir(), "test")
	if err := os.RemoveAll(folder); err != nil {
		t.Fatalf("%+v", err)
	}
	if ok, err := utils.FileExists(folder); err != nil {
		t.Fatalf("%+v", err)
	} else if ok {
		t.Errorf("%s should not exist before ensuring it exists", folder)
	}

	if err := EnsureStorePathExists(folder); err != nil {
		t.Fatalf("%+v", err)
	}

	if ok, err := utils.FileExists(folder); err != nil {
		t.Fatalf("%+v", err)
	} else if !ok {
		t.Errorf("%s should exist but it does not", folder)
	}
}
