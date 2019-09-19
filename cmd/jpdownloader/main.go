package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/jpi"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/utils"
	"github.com/juju/errors"
	"github.com/mmikulicic/multierror"
)

const (
	maxWorkers = 10
)

var (
	workingDir = flag.String("working-dir", filepath.Join(os.Getenv("HOME"), ".jenkins"), "plugins working directory, default to the HOME/.jenkins directory")
	outputDir  = flag.String("output-dir", filepath.Join(os.Getenv("JENKINS_HOME"), "plugins"), "output directory, default to the JENKINS_HOME/plugins directory")
	inputFile  = flag.String("input", "plugins.json.lock", "input file. You can use the output of jpresolver")
)

// This function will copy the downloaded plugins (from the working directory
// used to as fs cache) to the output directory. It will rename them to match
// the jenkins requirements (.pinned suffix and no version in the filename).
func copyPlugins(plugins *api.PluginsRequest) error {
	var errs error
	for _, p := range plugins.Plugins {
		src := jpi.GetPluginPath(p, *workingDir)
		r, err := os.Open(src)
		if err != nil {
			errs = multierror.Append(errs, errors.Trace(err))
			continue
		}
		defer r.Close()

		dst := filepath.Join(*outputDir, fmt.Sprintf("%s.jpi.pinned", p.Name))
		w, err := os.Create(dst)
		if err != nil {
			errs = multierror.Append(errs, errors.Trace(err))
			continue
		}
		defer w.Close()

		_, err = io.Copy(w, r)
		if err != nil {
			errs = multierror.Append(errs, errors.Trace(err))
			continue
		}
	}
	return errs
}

func run() error {
	plugins := &api.PluginsRequest{}
	if err := utils.UnmarshalJSON(*inputFile, plugins); err != nil {
		return errors.Trace(err)
	}

	if err := jpi.RunWorkersPoll(plugins, *workingDir, maxWorkers); err != nil {
		return errors.Trace(err)
	}

	return copyPlugins(plugins)
}

func init() {
	flag.Parse()
	if *inputFile == "" {
		flag.Usage()
		log.Fatalf("undefined input file")
	}
	if ok, err := utils.FileExists(*outputDir); err != nil {
		log.Fatalf("%+v", err)
	} else if !ok {
		flag.Usage()
		log.Fatalf("the output directory does not exist")
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
