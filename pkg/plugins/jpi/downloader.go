package jpi

import (
	"context"
	"log"
	"time"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/downloader/common"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/utils"
	"github.com/google/renameio"
	"github.com/juju/errors"
	"github.com/mkmik/multierror"
)

var (
	updatesURL = "https://updates.jenkins.io/download"
)

const (
	timeoutMin = 2
)

// FetchPlugin will download the requested plugin in the provided path
func FetchPlugin(p *api.Plugin, d common.Downloader, workingDir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutMin*time.Minute)
	defer cancel()

	pluginPath := GetPluginPath(p, workingDir)
	if cached, err := utils.FileExists(pluginPath); err != nil {
		return errors.Trace(err)
	} else if cached {
		return nil
	}

	log.Printf("> downloading %s plugin...\n", p.Identifier())

	t, err := renameio.TempFile("", pluginPath)
	if err != nil {
		return err
	}
	defer t.Cleanup()

	if err := d.Download(ctx, p, t); err != nil {
		return errors.Annotatef(err, "unable to download %q", d.GetDownloadURL(p))
	}

	return t.CloseAtomicallyReplace()
}

func worker(id int, jobs <-chan *downloadRequest, results chan<- error) {
	for dr := range jobs {
		err := FetchPlugin(dr.Plugin, dr.Downloader, dr.WorkingDir)
		results <- err
	}
}

type downloadRequest struct {
	Plugin     *api.Plugin
	WorkingDir string
	Downloader common.Downloader
}

func newDownloadRequest(p *api.Plugin, d common.Downloader, path string) *downloadRequest {
	return &downloadRequest{
		Downloader: d,
		Plugin:     p,
		WorkingDir: path,
	}
}

// RunWorkersPoll will start a poll of workers to download the provided plugins list
func RunWorkersPoll(psr *api.PluginsRegistry, d common.Downloader, workingDir string, maxNumWorkers int) error {
	numPlugins := len(psr.Plugins)
	jobs := make(chan *downloadRequest, numPlugins)
	results := make(chan error, numPlugins)

	// Setup workers to download plugins concurrently
	for workerID := 0; workerID <= maxNumWorkers; workerID++ {
		go worker(workerID, jobs, results)
	}

	for _, p := range psr.Plugins {
		jobs <- newDownloadRequest(p, d, workingDir)
	}
	close(jobs)

	var errs error
	for n := 0; n < numPlugins; n++ {
		if err := <-results; err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs
}
