package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/downloader/jenkinsdownloader"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/graph"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/jpi"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/meta"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/utils"
	"github.com/golang/protobuf/jsonpb"
	"github.com/juju/errors"
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

func resolve(pr *api.PluginsRegistry) (*api.PluginsRegistry, error) {
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

	lock, err := graph.LockPlugins(g, *optional)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if err := graph.FindIncompatibilities(pr, lock, g); err != nil {
		return nil, errors.Trace(err)
	}

	return lock, nil
}

func run() error {
	plugins := &api.PluginsRegistry{}
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
