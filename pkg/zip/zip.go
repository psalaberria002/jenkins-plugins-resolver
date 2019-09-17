package zip

import (
	"archive/zip"
	"io"
	"os"

	"github.com/juju/errors"
)

// OpenFile finds a file `src` from the provided reader `w` for a zip file and returns a reader to it
func OpenFile(r *os.File, src string) (io.ReadCloser, error) {
	stat, err := r.Stat()
	if err != nil {
		return nil, errors.Trace(err)
	}

	zr, err := zip.NewReader(r, stat.Size())
	if err != nil {
		return nil, errors.Trace(err)
	}

	// we will get out of the loop when EOF
	for _, f := range zr.File {
		if f.Name != src {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		return rc, nil
	}

	return nil, errors.Errorf("unable to find %s in the zip file", src)
}
