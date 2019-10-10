package war

import (
	"log"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/jar"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/requesters"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/utils"
	"github.com/juju/errors"
)

// WriteWar writes a Jenkins struct into a file
func WriteWar(g *api.Jenkins, warPath string) error {
	return utils.MarshalJSON(warPath, g)
}

// ReadWar reads a Jenkins struct from a file
func ReadWar(warPath string) (*api.Jenkins, error) {
	jk := &api.Jenkins{}
	return jk, utils.UnmarshalJSON(warPath, jk)
}

// Read reads a Jenkins struct from a file.
// If it does not exist, it will parse it and write the processed data to the store.
// If it does exist, it will read it from the store directly.
func Read(warfile string, workingDir string) (*api.Jenkins, error) {
	manifest, err := jar.ExtractManifest(warfile)
	if err != nil {
		return nil, errors.Trace(err)
	}

	jm, err := ParseManifest(manifest)
	if err != nil {
		return nil, errors.Trace(err)
	}

	warPath := GetWarPath(jm, workingDir)
	cached, err := utils.FileExists(warPath)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if cached {
		return ReadWar(warPath)
	}

	jenkins, err := jm.Parse(warfile)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if err := WriteWar(jenkins, warPath); err != nil {
		return nil, errors.Trace(err)
	}

	return jenkins, nil
}

// AddMissings walks through a jenkins input to check if there are missing plugins in a plugin registry and add them
func AddMissings(jkpr *api.PluginsRegistry, pr *api.PluginsRegistry) {
	for _, warPlugin := range jkpr.GetPlugins() {
		found := false
		for _, plugin := range pr.Plugins {
			if plugin.GetName() == warPlugin.GetName() {
				found = true
				break
			}
		}
		if !found {
			log.Printf("adding Jenkins detached plugin to the plugins registry: %s\n", warPlugin.Identifier())
			pr.Plugins = append(pr.Plugins, warPlugin)
		}
	}
}

// NewPluginsRegistry returns a PluginRegistry struct from a Jenkins struct.
func NewPluginsRegistry(jk *api.Jenkins) *api.PluginsRegistry {
	jkpr := &api.PluginsRegistry{}
	for _, warPluginMetadata := range jk.GetPlugins() {
		p := warPluginMetadata.GetPlugin()
		jkpr.Plugins = append(jkpr.Plugins, &api.Plugin{
			Name:      p.Name,
			Version:   p.Version,
			Requester: requesters.WAR,
		})
	}
	return jkpr
}
