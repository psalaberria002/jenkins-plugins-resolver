package jpi

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/utils"
	"github.com/juju/errors"
	"github.com/mmikulicic/multierror"
)

const (
	updatesURL = "https://updates.jenkins.io/download"
	timeoutMin = 2
)

// This way we can override the url for testing
func getDownloadURL(p *api.Plugin) string {
	return fmt.Sprintf("%s/plugins/%s/%s/%s.hpi", updatesURL, p.Name, p.Version, p.Name)
}

func download(ctx context.Context, url string, w io.Writer) error {
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

// FetchPlugin will download the requested plugin in the provided path
func FetchPlugin(p *api.Plugin, workingDir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutMin*time.Minute)
	defer cancel()

	pluginPath := GetPluginPath(p, workingDir)
	cached, err := utils.FileExists(pluginPath)
	if err != nil {
		return errors.Trace(err)
	}

	if cached {
		return nil
	}

	w, err := os.Create(pluginPath)
	if err != nil {
		return errors.Trace(err)
	}
	defer w.Close()

	url := getDownloadURL(p)
	if err := download(ctx, url, w); err != nil {
		return errors.Annotatef(err, "unable to download %q", url)
	}

	return nil
}

func worker(id int, jobs <-chan *downloadRequest, results chan<- error) {
	for dr := range jobs {
		log.Printf("#%2d> downloading %s...\n", id, dr.Plugin.Identifier())
		err := FetchPlugin(dr.Plugin, dr.WorkingDir)
		results <- err
	}
}

type downloadRequest struct {
	Plugin     *api.Plugin
	WorkingDir string
}

func newDownloadRequest(p *api.Plugin, path string) *downloadRequest {
	return &downloadRequest{
		Plugin:     p,
		WorkingDir: path,
	}
}

// RunWorkersPoll will start a poll of workers to download the provided plugins list
func RunWorkersPoll(psr *api.PluginsRequest, workingDir string, maxNumWorkers int) error {
	numPlugins := len(psr.Plugins)
	jobs := make(chan *downloadRequest, numPlugins)
	results := make(chan error, numPlugins)

	// Setup workers to download plugins concurrently
	for workerID := 0; workerID <= maxNumWorkers; workerID++ {
		go worker(workerID, jobs, results)
	}

	for _, p := range psr.Plugins {
		jobs <- newDownloadRequest(p, workingDir)
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
