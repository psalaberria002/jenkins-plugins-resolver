//go:generate protoc --proto_path=. --proto_path=../vendor --go_out=. pluginsapi.proto

// Package api provides APIs for plugins relate formats
package api

import "fmt"

// Filename returns the filename string for a plugin
func (p *Plugin) Filename() string {
	return fmt.Sprintf("%s-%s", p.Name, p.Version)
}

// Identifier returns the identifier string for a plugin
func (p *Plugin) Identifier() string {
	return fmt.Sprintf("%s:%s", p.Name, p.Version)
}

// GetPluginsRegistry returns a PluginsRegistry structure from
// a Project one.
func (p *Project) GetPluginsRegistry() *PluginsRegistry {
	plugins := []*Plugin{}
	for name, version := range p.Dependencies {
		plugins = append(plugins, &Plugin{
			Name:    name,
			Version: version,
		})
	}
	return &PluginsRegistry{
		Plugins: plugins,
	}
}

// ByName implements the sort.Interface interface for sorting list of plugins
type ByName []*Plugin

func (pl ByName) Len() int {
	return len(pl)
}

func (pl ByName) Less(i, j int) bool {
	return pl[i].Name < pl[j].Name
}

func (pl ByName) Swap(i, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}
