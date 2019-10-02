package common

import (
	"log"
)

// Incompatibility represents a versioning issue between a lock file and a plugin registry
type Incompatibility struct {
	Cause      string
	Requesters []string
}

// Incompatibilities maps a plugin to a target incompatibility
type Incompatibilities map[string]Incompatibility

// Print prints a map of incompatibilities
func (incs Incompatibilities) Print() {
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
