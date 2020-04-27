package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/common"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/downloader/jenkinsdownloader"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/graph"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/jpi"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/meta"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/war"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/utils"
	"github.com/golang/protobuf/jsonpb"
	"github.com/juju/errors"
	"github.com/mkmik/multierror"
)

const (
	maxWorkers = 10
)

var (
	gitCommit = "UNKNOWN"

	inputFile  = flag.String("input", "plugins.json", "input file (.json, .jsonnet. .yaml or .yml)")
	warFile    = flag.String("war", "", "jenkins war file")
	optional   = flag.Bool("optional", false, "add optional dependencies to the output. It will allow plugins to run with all the expected features.")
	showGraph  = flag.Bool("show-graph", false, "show whole dependencies graph in JSON")
	workingDir = flag.String("working-dir", filepath.Join(os.Getenv("HOME"), ".jpr"), "plugins working dir")
)

// mergePlugins returns the merge of two slices of plugins with some caveats:
//
// The requested plugins are provided by the user and the bundled plugins are Jenkins dependencies.
// Therefore, we will include all the requested versions that meet the Jenkins requirements. Otherwise,
// it will fail.
//
// Any additional bundled plugin (not present in the requested plugin list) will be included too.
func mergePlugins(requestedPlugins []*api.Plugin, bundledPlugins []*api.Plugin) ([]*api.Plugin, error) {
	var errs error

	plugins := []*api.Plugin{}
	for _, requestedPlugin := range requestedPlugins {
		// A requested plugin is compatible if the bundled plugin list does not
		// include the same plugin with a higher version request.
		compatible := func(rp *api.Plugin) error {
			for _, bp := range bundledPlugins {
				if rp.Name == bp.Name {
					lower, err := utils.VersionLower(bp.Version, rp.Version)
					if err != nil {
						return errors.Trace(err)
					}
					if !lower && rp.Version != bp.Version {
						return errors.Errorf("found bundled plugin %s: it is higher than the requested %s plugin, they are incompatible. Please bump your requested plugin version.", bp.Identifier(), rp.Identifier())
					}
				}
			}
			return nil
		}

		if err := compatible(requestedPlugin); err != nil {
			errs = multierror.Append(errs, errors.Trace(err))
			continue
		}
		plugins = append(plugins, requestedPlugin)
	}
	if errs != nil {
		return plugins, errors.Trace(errs)
	}

	// Add remaining plugins
	for _, bundledPlugin := range bundledPlugins {
		found := false
		for _, p := range plugins {
			if bundledPlugin.Name == p.Name {
				found = true
				break
			}
		}
		if !found {
			log.Printf("added bundled plugin %s.\n", bundledPlugin.Identifier())
			plugins = append(plugins, bundledPlugin)
		}
	}

	return plugins, nil
}

func lockPlugins(requestedPlugins []*api.Plugin, bundledPlugins []*api.Plugin) (*api.PluginsRegistry, error) {
	plugins, err := mergePlugins(requestedPlugins, bundledPlugins)
	if err != nil {
		return nil, errors.Trace(err)
	}

	downloader := jenkinsdownloader.NewDownloader()
	g, err := graph.FetchGraph(plugins, downloader, *workingDir, maxWorkers)
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

	incs, err := graph.FindIncompatibilities(plugins, lock.Plugins, g)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if len(incs) > 0 {
		log.Printf(" There were found some incompatibilities:\n")
		incs.Print()
	}

	return lock, nil
}

func readInput() (*api.Project, error) {
	project := &api.Project{}
	if err := utils.UnmarshalFile(*inputFile, project); err != nil {
		return nil, errors.Trace(err)
	}
	return project, nil
}

func writeOutput(pr *api.PluginsRegistry) error {
	outputFile := fmt.Sprintf("%s-lock.%s", strings.TrimSuffix(*inputFile, filepath.Ext(*inputFile)), "json")
	sort.Sort(api.ByName(pr.Plugins))
	return utils.MarshalJSON(outputFile, pr)
}

func run() error {
	if err := validateFlags(); err != nil {
		flag.Usage()
		return errors.Trace(err)
	}

	project, err := readInput()
	if err != nil {
		return errors.Trace(err)
	}
	plugins := project.GetPluginsRegistry()

	jkpr := &api.PluginsRegistry{}
	if *warFile != "" {
		jk, err := war.Read(*warFile, *workingDir)
		if err != nil {
			return errors.Trace(err)
		}
		jkpr = war.NewPluginsRegistry(jk)
	}

	lock, err := lockPlugins(plugins.Plugins, jkpr.Plugins)
	if err != nil {
		return errors.Trace(err)
	}

	return writeOutput(lock)
}

func validateFlags() error {
	if ok, err := utils.FileExists(*inputFile); err != nil {
		return errors.Trace(err)
	} else if !ok {
		return errors.Errorf("%s does not exist", *inputFile)
	}

	// Ensure working paths exist
	var errs error
	for _, fn := range []func(string) string{
		graph.GetStorePath,
		jpi.GetStorePath,
		meta.GetStorePath,
		war.GetStorePath,
	} {
		if err := common.EnsureStorePathExists(*workingDir, fn); err != nil {
			errs = multierror.Append(errs, errors.Trace(err))
			continue
		}
	}
	return errs
}

func main() {
	flag.Parse()

	log.Printf("Version commit: %s\n", gitCommit)
	if err := run(); err != nil {
		log.Fatalf("%+v", err)
	}
}
