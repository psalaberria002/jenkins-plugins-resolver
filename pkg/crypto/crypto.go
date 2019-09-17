package crypto

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/juju/errors"
)

// SHA256 will return the sha256 sum of the provided filename content
func SHA256(filename string) (string, error) {
	hasher := sha256.New()
	r, err := os.Open(filename)
	if err != nil {
		return "", errors.Trace(err)
	}
	defer r.Close()
	if _, err := io.Copy(hasher, r); err != nil {
		return "", errors.Trace(err)
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
