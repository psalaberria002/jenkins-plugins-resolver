package jpi

import (
	"regexp"
	"strings"

	api "github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/juju/errors"
)

var (
	versionRegex      = regexp.MustCompile(`Plugin-Version:\s*([^\r]+)\r?\n`)
	shortNameRegex    = regexp.MustCompile(`Short-Name:\s*([^\r]+)\r?\n`)
	fullNameRegex     = regexp.MustCompile(`Long-Name:\s*([^\r]+)\r?\n`)
	dependenciesRegex = regexp.MustCompile(`Plugin-Dependencies:\s*([^\r]+)\r?\n`)
	dependencyRegex   = regexp.MustCompile(`([^:]+):([^;]+)(;(.*))?`)
)

// ParseManifest will parse a MANIFEST.MF file into a proper struct
func ParseManifest(manifest string) (*api.PluginMetadata, error) {
	var err error

	// These manifest have a specific syntax. We can treat a carriage,
	// new line and one space as part of the same line.
	manifest = strings.ReplaceAll(manifest, "\r\n ", "")

	version, err := findMatch(versionRegex, manifest)
	if err != nil {
		return nil, errors.Trace(err)
	}
	shortName, err := findMatch(shortNameRegex, manifest)
	if err != nil {
		return nil, errors.Trace(err)
	}
	fullName, err := findMatch(fullNameRegex, manifest)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var dependencies []*api.Plugin
	var optionalDependencies []*api.Plugin
	// NOTE: We are missing the error check in purpose: Plugins with
	// 		 no deps will miss the Plugin-Dependencies field.
	dependenciesStr, _ := findMatch(dependenciesRegex, manifest)
	if dependenciesStr != "" {
		dependencies, optionalDependencies, err = NewDependencies(dependenciesStr)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	pm := api.PluginMetadata{
		FullName: fullName,
		Plugin: &api.Plugin{
			Name:    shortName,
			Version: version,
		},
		Dependencies:         dependencies,
		OptionalDependencies: optionalDependencies,
	}

	return &pm, nil
}

func findMatch(r *regexp.Regexp, manifest string) (string, error) {
	matches := r.FindStringSubmatch(manifest)
	if len(matches) == 0 {
		return "", errors.Errorf("cannot match %q in %q", r, manifest)
	}
	return matches[1], nil
}

// NewDependencies will parse a dependencies string to return a list of dependencies
func NewDependencies(dependenciesStr string) ([]*api.Plugin, []*api.Plugin, error) {
	var dependencies []*api.Plugin
	var optionalDependencies []*api.Plugin
	for _, depStr := range strings.Split(dependenciesStr, ",") {
		depMatches := dependencyRegex.FindStringSubmatch(depStr)
		if len(depMatches) == 0 {
			return nil, nil, errors.Errorf("canonot match %q in %q", dependencyRegex, depStr)
		}
		dependency := api.Plugin{
			Name:    depMatches[1],
			Version: depMatches[2],
		}

		optional := depMatches[4] == "resolution:=optional"
		if optional {
			optionalDependencies = append(optionalDependencies, &dependency)
		} else {
			dependencies = append(dependencies, &dependency)
		}
	}

	return dependencies, optionalDependencies, nil
}
