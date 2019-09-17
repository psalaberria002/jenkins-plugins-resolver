package jpi

import (
	"io/ioutil"
	"net/http"
	"os"

	zipper "github.com/bitnami-labs/jenkins-plugins-resolver/pkg/zip"
	"github.com/juju/errors"
)

const (
	manifestPath = "META-INF/MANIFEST.MF"
)

// ExtractManifest takes a jpi and returns a reader to its manifest content
func ExtractManifest(jpi string) (string, error) {
	r, err := os.Open(jpi)
	if err != nil {
		return "", errors.Trace(err)
	}
	defer r.Close()

	if mimetype, err := getFileMimeType(r); err != nil {
		return "", errors.Trace(err)
	} else if mimetype != "application/zip" {
		return "", errors.Errorf("%s is not a valid jpi file", jpi)
	}

	rc, err := zipper.OpenFile(r, manifestPath)
	if err != nil {
		return "", errors.Trace(err)
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return "", errors.Trace(err)
	}

	return string(data), nil
}

// getFileMimeType guesses the mimetype of a file from reading its first 512 bytes
func getFileMimeType(f *os.File) (string, error) {
	text := make([]byte, 512)
	// We need to capture the number of bytes ReadAt returns and slice text to that
	// length. Otherwise, text would contain extra data that is not in the file to
	// DetectContentType.
	nbr, err := f.ReadAt(text, 0)
	if err != nil {
		return "", nil
	}
	return http.DetectContentType(text[:nbr]), nil
}
