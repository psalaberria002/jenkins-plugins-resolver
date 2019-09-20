package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/downloader/jenkinsdownloader"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/graph"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/jpi"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/meta"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/utils"
	"github.com/golang/protobuf/jsonpb"
	"github.com/juju/errors"
	"github.com/mmikulicic/multierror"
	"go.nami.run/gotools/version"
)

const (
	maxWorkers = 10
)

var (
	workingDir = flag.String("working-dir", filepath.Join(os.Getenv("HOME"), ".jenkins"), "plugins working dir")
	inputFile  = flag.String("input", "plugins.json", "input file.")
	outputFile = flag.String("output", "", "output file. If not provided, it will default to <input>.lock")
	optional   = flag.Bool("optional", false, "add optional dependencies to the output. It will allow plugins to run with all the expected features.")
	showGraph  = flag.Bool("show-graph", false, "show whole dependencies graph in JSON")
)

func versionLower(i string, j string) (bool, error) {
	vj, err := version.New(j)
	if err != nil {
		return false, errors.Errorf("Error parsing version %s: %s", j, err)
	}

	if i == "" && j != "" {
		return true, nil
	}

	vi, err := version.New(i)
	if err != nil {
		return false, errors.Errorf("Error parsing version %s: %s", i, err)
	}

	return vi.Less(vj), nil
}

type pluginsMap map[string]string

func updatePluginsMap(pm pluginsMap, p *api.Plugin) error {
	if ok, err := versionLower(pm[p.Name], p.Version); err != nil {
		return errors.Trace(err)
	} else if !ok {
		return nil
	}
	pm[p.Name] = p.Version
	return nil
}

// NewPluginsRequest iterates a map of plugins and cretes a new plugins request
func NewPluginsRequest(pm pluginsMap) *api.PluginsRequest {
	names := make([]string, 0, len(pm))
	for name := range pm {
		names = append(names, name)
	}
	sort.Strings(names)

	plugins := []*api.Plugin{}
	for _, name := range names {
		p := api.Plugin{
			Name:    name,
			Version: pm[name],
		}
		plugins = append(plugins, &p)
	}
	return &api.PluginsRequest{
		Plugins: plugins,
	}
}

// This method will resolve all the required dependencies and update the plugins
// map with any newer (and new) found plugin version.
//
// It will iterate over the graph nodes by accessing dependencies recursively.
// If we want to resolve optional dependencies too (-optional flag) we will also
// iterate over the optional dependencies, as we can consider them regular ones.
func resolveNodeDependencies(pm pluginsMap, n *api.Graph_Node) error {
	if err := updatePluginsMap(pm, n.Plugin); err != nil {
		return errors.Trace(err)
	}
	for _, nd := range n.Dependencies {
		if err := resolveNodeDependencies(pm, nd); err != nil {
			return errors.Trace(err)
		}
	}
	if *optional {
		for _, nd := range n.OptionalDependencies {
			if err := resolveNodeDependencies(pm, nd); err != nil {
				return errors.Trace(err)
			}
		}
	}
	return nil
}

// This method will resolve all the "required optional dependencies" and update the plugins
// map with any newer (and new) found plugin version.
//
// Required optional dependencies are those already added dependencies (regular dependencies).
// Optional dependencies are not optional anymore if another plugin depends directly on it.
//
// Example:
//  - a:1.0 has b:2.0 as optional dep
//  - c:1.0 has b:1.0 as dep
//
// Therefore, Jenkins will complain about b:1.0 being older than required (b is
// installed now and, therefore, required to meet the requirements of a:1.0)
//
// It seems this is not a documented behavior:
// https://wiki.jenkins.io/display/JENKINS/Dependencies+among+plugins
func resolveNodeOptionalDependencies(pm pluginsMap, n *api.Graph_Node) error {
	// If we don't want optional dependencies to be included in the output,
	// we will only process those optional dependencies that have been already
	// added to the map (they are real dependencies for another plugin)
	if !*optional && pm[n.Plugin.Name] == "" {
		return nil
	}
	if err := updatePluginsMap(pm, n.Plugin); err != nil {
		return errors.Trace(err)
	}
	for _, nd := range n.OptionalDependencies {
		if err := resolveNodeOptionalDependencies(pm, nd); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func resolve(pr *api.PluginsRequest) (*api.PluginsRequest, error) {
	downloader := jenkinsdownloader.NewDownloader()
	g, err := graph.FetchGraph(pr, downloader, *inputFile, *workingDir, maxWorkers)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if *showGraph {
		m := jsonpb.Marshaler{Indent: "  "}
		if err := m.Marshal(os.Stdout, g); err != nil {
			return nil, errors.Trace(err)
		}
	}

	// Auxiliar mapping to set plugins by name
	pluginsMap := make(map[string]string)

	// We need to resolve nodes dependencies first because they might include
	// plugins that are optional dependencies for others.
	var errs error
	for _, n := range g.Nodes {
		if err := resolveNodeDependencies(pluginsMap, n); err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
	}
	for _, n := range g.Nodes {
		if err := resolveNodeOptionalDependencies(pluginsMap, n); err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
	}
	if errs != nil {
		return nil, errors.Trace(errs)
	}

	npr, err := NewPluginsRequest(pluginsMap), nil
	if err != nil {
		return nil, errors.Trace(err)
	}

	if err := graph.FindIncompatibilities(g, pr, npr); err != nil {
		return nil, errors.Trace(err)
	}

	return npr, nil
}

func run() error {
	plugins := &api.PluginsRequest{}
	if err := utils.UnmarshalJSON(*inputFile, plugins); err != nil {
		return errors.Trace(err)
	}

	pr, err := resolve(plugins)
	if err != nil {
		return errors.Trace(err)
	}

	if err := utils.MarshalJSON(*outputFile, pr); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func init() {
	flag.Parse()
	if ok, err := utils.FileExists(*inputFile); err != nil {
		flag.Usage()
		log.Fatalf("%+v", err)
	} else if !ok {
		flag.Usage()
		log.Fatalf("%s does not exist", *inputFile)
	}
	if *outputFile == "" {
		*outputFile = *inputFile + ".lock"
	}

	// Ensure working paths exist
	if err := os.MkdirAll(*workingDir, 0777); err != nil {
		log.Fatalf("%+v", err)
	}
	if err := graph.EnsureStorePathExists(*workingDir); err != nil {
		log.Fatalf("%+v", err)
	}
	if err := jpi.EnsureStorePathExists(*workingDir); err != nil {
		log.Fatalf("%+v", err)
	}
	if err := meta.EnsureStorePathExists(*workingDir); err != nil {
		log.Fatalf("%+v", err)
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("%+v", err)
	}
}
