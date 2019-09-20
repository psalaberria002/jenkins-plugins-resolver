package common

import (
	"context"
	"io"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
)

// Downloader wraps a common interface for downloader implementations
type Downloader interface {
	Download(context.Context, *api.Plugin, io.Writer) error
	GetDownloadURL(*api.Plugin) string
}
