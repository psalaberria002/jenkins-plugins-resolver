package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
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

func walkNode(pm pluginsMap, n *api.Graph_Node) error {
	if err := updatePluginsMap(pm, n.Plugin); err != nil {
		return errors.Trace(err)
	}
	for _, nd := range n.Dependencies {
		if err := walkNode(pm, nd); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func resolve(pr *api.PluginsRequest) (*api.PluginsRequest, error) {
	g, err := graph.FetchGraph(pr, *inputFile, *workingDir, maxWorkers)
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

	var errs error
	for _, n := range g.Nodes {
		if err := walkNode(pluginsMap, n); err != nil {
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
