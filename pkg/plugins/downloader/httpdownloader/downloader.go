package httpdownloader

import (
	"context"
	"io"
	"net/http"

	"github.com/juju/errors"
)

// Download will fetch a plugin from a HTTP/HTTPS endpoint and will
// write it to the provided writer.
func Download(ctx context.Context, url string, w io.Writer) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return errors.Trace(err)
	}
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return errors.Trace(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("got status %q", resp.Status)
	}

	_, err = io.Copy(w, resp.Body)
	return err
}
