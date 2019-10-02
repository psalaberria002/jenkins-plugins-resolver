package jar

import (
	zipper "github.com/bitnami-labs/jenkins-plugins-resolver/pkg/zip"
	"github.com/juju/errors"
)

const (
	manifestPath = "META-INF/MANIFEST.MF"
)

// ExtractManifest returns a reader to the manifest of a jar file.
func ExtractManifest(jarfile string) (string, error) {
	data, err := zipper.ExtractFile(jarfile, manifestPath)
	if err != nil {
		return "", errors.Trace(err)
	}

	return string(data), nil
}
