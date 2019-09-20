package jenkinsdownloader

import (
	"context"
	"fmt"
	"io"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/downloader/httpdownloader"
)

// Downloader wrapps the official Jenkins downloads URL around a PluginsFetcher
type Downloader struct {
	URL string
}

const (
	jenkinsURL = "https://updates.jenkins.io/download"
)

// NewDownloader will return a new fetcher
func NewDownloader() *Downloader {
	return &Downloader{
		URL: jenkinsURL,
	}
}

// GetDownloadURL prints the URL for the given plugin
func (d *Downloader) GetDownloadURL(p *api.Plugin) string {
	return fmt.Sprintf("%s/plugins/%s/%s/%s.hpi", d.URL, p.Name, p.Version, p.Name)
}

// Download will fetch a plugin from the official Jenkins download page and will
// write it to the provided writer.
func (d *Downloader) Download(ctx context.Context, p *api.Plugin, w io.Writer) error {
	url := d.GetDownloadURL(p)
	return httpdownloader.Download(ctx, url, w)
}
