package graph

import (
	"fmt"
	"log"
	"strings"

	api "github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/utils"
	"github.com/juju/errors"
	"github.com/mmikulicic/multierror"
)

// Incompatibility represents a versioning issue between a lock file and a plugin registry
type Incompatibility struct {
	Cause      string
	Requesters []string
}

// Incompatibilities maps a plugin to a target incompatibility
type Incompatibilities map[string]Incompatibility

// FindIncompatibilities walks through a graph if there are missmatches between a plugin registry and its
// locked version
func FindIncompatibilities(pr *api.PluginsRegistry, lock *api.PluginsRegistry, g *api.Graph) (Incompatibilities, error) {
	var errs error
	incompatibilities := make(Incompatibilities)
	for _, ip := range pr.Plugins {
		var p *api.Plugin
		for _, op := range lock.Plugins {
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
			reqs := []string{}
			for _, n := range g.Nodes {
				reqs = findPluginRequesters(n, p, reqs, []string{})
			}
			incompatibilities[ip.Identifier()] = Incompatibility{
				Cause:      fmt.Sprintf("a newer version was locked: %s", p.Identifier()),
				Requesters: reqs,
			}
		}
	}
	return incompatibilities, errs
}

func findPluginRequesters(n *api.Graph_Node, p *api.Plugin, r []string, aux []string) []string {
	if p.Version == n.Plugin.Version {
		aux := append(aux, n.Plugin.Identifier())
		tree := strings.Join(aux, " > ")
		r = append(r, tree)
	} else {
		aux = append(aux, n.Plugin.Identifier())
	}
	for _, nd := range n.Dependencies {
		r = findPluginRequesters(nd, p, r, aux)
	}
	for _, nd := range n.OptionalDependencies {
		r = findPluginRequesters(nd, p, r, aux)
	}
	return r
}

// PrintIncompatibilities prints a map of incompatibilities
func (incs Incompatibilities) PrintIncompatibilities() {
	for id, inc := range incs {
		log.Printf("  ├── %s (%s):\n", id, inc.Cause)
		for nr, req := range inc.Requesters {
			if nr == len(inc.Requesters)-1 {
				log.Printf("  │   └── %s\n", req)
			} else {
				log.Printf("  │   ├── %s\n", req)
			}
		}
	}
}
