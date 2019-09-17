package meta

import (
	"log"

	api "github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/jpi"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/utils"
	"github.com/juju/errors"
	"github.com/mmikulicic/multierror"
)

// WriteMetadata will write the plugin metadata into a file
func WriteMetadata(pm *api.PluginMetadata, metaPath string) error {
	return utils.MarshalJSON(metaPath, pm)
}

// ReadMetadata will read the plugin metadata from a file
func ReadMetadata(metaPath string) (*api.PluginMetadata, error) {
	pm := &api.PluginMetadata{}
	return pm, utils.UnmarshalJSON(metaPath, pm)
}

// Print provides a handy way of showing the plugin metadata
func Print(pm *api.PluginMetadata) {
	if pm.Dependencies != nil && len(pm.Dependencies) > 0 {
		log.Printf("%s depends on %q", pm.Plugin.Identifier(), pm.Dependencies)
		return
	}
	log.Printf("%s has no dependencies", pm.Plugin.Identifier())
}

// FetchMetadata will fetch the metadata for the requested plugin
func FetchMetadata(p *api.Plugin, workingDir string) (*api.PluginMetadata, error) {
	metaPath := GetMetaPath(p, workingDir)
	cached, err := utils.FileExists(metaPath)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if cached {
		return ReadMetadata(metaPath)
	}

	if err := jpi.FetchPlugin(p, workingDir); err != nil {
		return nil, errors.Trace(err)
	}

	jpiFile := jpi.GetPluginPath(p, workingDir)
	manifest, err := jpi.ExtractManifest(jpiFile)
	if err != nil {
		return nil, errors.Trace(err)
	}

	pm, err := jpi.ParseManifest(manifest)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if err := WriteMetadata(pm, metaPath); err != nil {
		return nil, errors.Trace(err)
	}

	return nil, nil
}

func worker(id int, jobs <-chan *metadataRequest, results chan<- error) {
	for mr := range jobs {
		log.Printf("#%2d> fetching %s metadata...\n", id, mr.Plugin.Identifier())
		_, err := FetchMetadata(mr.Plugin, mr.WorkingDir)
		results <- err
	}
}

type metadataRequest struct {
	Plugin     *api.Plugin
	WorkingDir string
}

func newMetadataRequest(p *api.Plugin, path string) *metadataRequest {
	return &metadataRequest{
		Plugin:     p,
		WorkingDir: path,
	}
}

// RunWorkersPoll will start a poll of workers to generate the metadata for the provided plugins list
func RunWorkersPoll(psr *api.PluginsRequest, workingDir string, maxNumWorkers int) error {
	numPlugins := len(psr.Plugins)
	jobs := make(chan *metadataRequest, numPlugins)
	results := make(chan error, numPlugins)

	// Setup workers to download plugins concurrently
	for workerID := 0; workerID <= maxNumWorkers; workerID++ {
		go worker(workerID, jobs, results)
	}

	for _, p := range psr.Plugins {
		jobs <- newMetadataRequest(p, workingDir)
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
