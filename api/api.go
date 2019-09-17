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
