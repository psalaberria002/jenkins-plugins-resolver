package testdownloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/downloader/httpdownloader"
)

// Downloader wrapps the official Jenkins downloads URL around a PluginsFetcher
type Downloader struct {
	URL        string
	Plugins    []*api.Plugin
	FileServer *httptest.Server
	MuxServer  *httptest.Server
}

// NewDownloader will return a new fetcher
func NewDownloader(dir string, plugins []*api.Plugin) *Downloader {

	d := &Downloader{
		Plugins:    plugins,
		FileServer: newFileServer(),
	}
	s := d.newMuxServer()
	d.MuxServer = s
	d.URL = s.URL

	return d
}

func getDownloadURLPath(p *api.Plugin) string {
	return fmt.Sprintf("/plugins/%s/%s/%s.hpi", p.Name, p.Version, p.Name)
}

// GetDownloadURL prints the URL for the given plugin
func (d *Downloader) GetDownloadURL(p *api.Plugin) string {
	return d.URL + getDownloadURLPath(p)
}

// Download will fetch a plugin from the official Jenkins download page and will
// write it to the provided writer.
func (d *Downloader) Download(ctx context.Context, p *api.Plugin, w io.Writer) error {
	url := d.GetDownloadURL(p)
	return httpdownloader.Download(ctx, url, w)
}

func getMockDownloadURL(serverURL string, p *api.Plugin) string {
	return fmt.Sprintf("%s/%s.jpi", serverURL, p.Filename())
}

// This mock will serve any file found under testdata/jpis
func newFileServer() *httptest.Server {
	handler := http.FileServer(http.Dir("testdata/jpis"))
	return httptest.NewServer(handler)
}

// This mock will simulate that there is a file in the expected URL schema
// by redirecting the request to the FileServer mock.
func (d *Downloader) newMuxServer() *httptest.Server {
	mux := http.NewServeMux()

	// Register each of the plugins requests
	for _, p := range d.Plugins {
		// New redirect handler to the FileServer mock
		// ie; http://fsmock:50502/credentials-2.2.0.jpi
		fsmockURL := fmt.Sprintf("%s/%s.jpi", d.FileServer.URL, p.Filename())
		handler := http.RedirectHandler(fsmockURL, http.StatusMovedPermanently)
		// New mux from the cannonical URL to the redirect handler
		// Example:
		//
		// http://muxmock:50503/plugins/credentials/2.2.0/credentials.hpi -> http://fsmock:50502/credentials-2.2.0.jpi
		//
		muxerURL := getDownloadURLPath(p)
		mux.Handle(muxerURL, handler)
	}
	return httptest.NewServer(mux)
}
