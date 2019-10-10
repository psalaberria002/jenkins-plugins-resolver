package graph

import (
	"fmt"
	"log"
	"strings"

	api "github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/requesters"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/utils"
	"github.com/juju/errors"
	"github.com/mkmik/multierror"
)

// Incompatibility represents a versioning issue between a lock file and a plugin registry
type Incompatibility struct {
	Plugin     *api.Plugin
	Cause      string
	Requesters []string
}

// Incompatibilities maps a plugin to a target incompatibility
type Incompatibilities []Incompatibility

// Print prints a map of incompatibilities
func (incs Incompatibilities) Print() {
	for ni, inc := range incs {
		incReqSep := "├"
		incSep := "│"
		if ni == len(incs)-1 {
			incReqSep = "└"
			incSep = " "
		}
		log.Printf("  %s── Requester: %s (%s)\n", incReqSep, inc.Plugin.Identifier(), inc.Plugin.Requester)
		log.Printf("  %s   Cause: %s\n", incSep, inc.Cause)

		for nr, req := range inc.Requesters {
			if nr == len(inc.Requesters)-1 {
				log.Printf("  %s     └── %s\n", incSep, req)
			} else {
				log.Printf("  %s     ├── %s\n", incSep, req)
			}
		}
		log.Printf("  %s\n", incSep)
	}
}

// FindIncompatibilities walks through a graph and check if there are missmatches between a
// list of requested plugins and the locked plugin versions.
func FindIncompatibilities(plugins []*api.Plugin, lockedPlugins []*api.Plugin, g *api.Graph) (Incompatibilities, error) {
	var errs error
	incompatibilities := Incompatibilities{}
	for _, ip := range plugins {
		var p *api.Plugin
		for _, op := range lockedPlugins {
			if ip.Name == op.Name {
				p = op
				break
			}
		}
		if p == nil {
			errs = multierror.Append(errs, errors.Errorf("unable to find %s in the locked file", ip.Identifier()))
			continue
		}
		lower, err := utils.VersionLower(ip.Version, p.Version)
		if err != nil {
			errs = multierror.Append(errs, errors.Trace(err))
			continue
		}

		if lower {
			// If the plugin requester is the war file, we can skip it
			// as our project file is supposed to request higher versions
			// of plugins
			if ip.Requester == requesters.WAR {
				continue
			}
			reqs := []string{}
			for _, n := range g.Nodes {
				reqs = findPluginRequesters(n, p, reqs, []string{})
			}
			inc := Incompatibility{
				Plugin:     ip,
				Cause:      "Some plugins require a newer version.",
				Requesters: reqs,
			}
			incompatibilities = append(incompatibilities, inc)
		}
	}
	return incompatibilities, errs
}

func findPluginRequesters(n *api.Graph_Node, p *api.Plugin, r []string, aux []string) []string {
	// We set a fancy header in the first iteration
	if len(aux) == 0 {
		aux = append(aux, fmt.Sprintf("%s (%s)", n.Plugin.Identifier(), n.Plugin.Requester))
	} else {
		aux = append(aux, n.Plugin.Identifier())
	}

	// If a match is found, we can populate the plugin tree
	// and append the tree to the list of requesters
	if p.Identifier() == n.Plugin.Identifier() {
		tree := strings.Join(aux, " > ")
		r = append(r, tree)
	}

	// We will find requesters in both, direct dependencies
	// and optional dependencies
	for _, nd := range n.Dependencies {
		r = findPluginRequesters(nd, p, r, aux)
	}
	for _, nd := range n.OptionalDependencies {
		r = findPluginRequesters(nd, p, r, aux)
	}
	return r
}
