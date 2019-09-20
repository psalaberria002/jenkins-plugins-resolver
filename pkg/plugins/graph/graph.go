package graph

import (
	"context"
	"time"

	api "github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/crypto"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/meta"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/utils"
	"github.com/juju/errors"
	"github.com/mmikulicic/multierror"
)

const (
	timeoutMin = 5
)

// WriteGraph will write a graph into a file
func WriteGraph(g *api.Graph, graphPath string) error {
	return utils.MarshalJSON(graphPath, g)
}

// ReadGraph reads a graph from the store
func ReadGraph(graphPath string) (*api.Graph, error) {
	g := &api.Graph{}
	return g, utils.UnmarshalJSON(graphPath, g)
}

// NewNode will return a graph node for the given plugin
func NewNode(p *api.Plugin, workingDir string) (*api.Graph_Node, error) {
	node := api.Graph_Node{
		Plugin: p,
	}
	dependencies := []*api.Graph_Node{}
	optionalDependencies := []*api.Graph_Node{}

	metaPath := meta.GetMetaPath(p, workingDir)
	pm, err := meta.ReadMetadata(metaPath)
	if err != nil {
		return nil, errors.Trace(err)
	}
	for _, dep := range pm.Dependencies {
		nodeDep, err := NewNode(dep, workingDir)
		if err != nil {
			return nil, errors.Trace(err)
		}
		dependencies = append(dependencies, nodeDep)
	}
	node.Dependencies = dependencies

	for _, dep := range pm.OptionalDependencies {
		nodeDep, err := NewNode(dep, workingDir)
		if err != nil {
			return nil, errors.Trace(err)
		}
		optionalDependencies = append(optionalDependencies, nodeDep)
	}
	node.OptionalDependencies = optionalDependencies

	return &node, nil
}

// FetchGraph computes the graph for a list of plugins or read it from the store
func FetchGraph(plugins *api.PluginsRequest, inputFile string, workingDir string, maxWorkers int) (*api.Graph, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutMin*time.Minute)
	defer cancel()

	hash, err := crypto.SHA256(inputFile)
	if err != nil {
		return nil, errors.Trace(err)
	}

	graphPath := GetGraphPath(hash, workingDir)
	cached, err := utils.FileExists(graphPath)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if cached {
		return ReadGraph(graphPath)
	}

	if err := fetch(ctx, plugins, workingDir, maxWorkers); err != nil {
		return nil, errors.Trace(err)
	}

	// Iterate the provided list of plugins first to initialize the map
	var errs error
	var nodes []*api.Graph_Node
	for _, p := range plugins.Plugins {
		node, err := NewNode(p, workingDir)
		if err != nil {
			errs = multierror.Append(errs, errors.Trace(err))
			continue
		}
		nodes = append(nodes, node)
	}
	if errs != nil {
		return nil, errors.Trace(errs)
	}

	g := api.Graph{
		Nodes: nodes,
	}
	if err := WriteGraph(&g, graphPath); err != nil {
		return nil, errors.Trace(err)
	}

	return &g, nil
}

func fetch(ctx context.Context, plugins *api.PluginsRequest, workingDir string, maxWorkers int) error {
	var errs error

	// Iterate the provided list of plugins first to fetch the metadata from upstream
	if err := meta.RunWorkersPoll(plugins, workingDir, maxWorkers); err != nil {
		return errors.Trace(err)
	}

	// Read metadata to fetch the metadata for them
	for _, p := range plugins.Plugins {
		metaPath := meta.GetMetaPath(p, workingDir)
		pm, err := meta.ReadMetadata(metaPath)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}

		// Iterate dependencies
		depPluginsRequest := api.PluginsRequest{
			Plugins: pm.Dependencies,
		}
		if err := fetch(ctx, &depPluginsRequest, workingDir, maxWorkers); err != nil {
			errs = multierror.Append(errs, err)
			continue
		}

		// Iterate optional dependencies
		optDepPluginsRequest := api.PluginsRequest{
			Plugins: pm.OptionalDependencies,
		}
		if err := fetch(ctx, &optDepPluginsRequest, workingDir, maxWorkers); err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
	}
	return errs
}
