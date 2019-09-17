package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/jpi"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/utils"
	"github.com/juju/errors"
)

const (
	maxWorkers = 10
)

var (
	workingDir = flag.String("working-dir", filepath.Join(os.Getenv("HOME"), ".jenkins"), "plugins working dir")
	inputFile  = flag.String("input", "plugins.json.lock", "input file. You can use the output of jpresolver")
)

func run() error {
	plugins := &api.PluginsRequest{}
	if err := utils.UnmarshalJSON(*inputFile, plugins); err != nil {
		return errors.Trace(err)
	}

	return jpi.RunWorkersPoll(plugins, *workingDir, maxWorkers)
}

func init() {
	flag.Parse()
	if *inputFile == "" {
		flag.Usage()
		log.Fatalf("undefined input file")
	}
	// Ensure working paths exist
	if err := os.MkdirAll(*workingDir, 0777); err != nil {
		log.Fatalf("%+v", err)
	}
	if err := jpi.EnsureStorePathExists(*workingDir); err != nil {
		log.Fatalf("%+v", err)
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("%+v", err)
	}

	log.Printf("done!")
}
