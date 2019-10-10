package graph

import (
	"testing"

	api "github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/golang/protobuf/jsonpb"
)

func TestFindIncompatibilities(t *testing.T) {
	testCases := []struct {
		graph             string
		plugins           string
		incompatibilities bool
	}{
		// This is a regular test with no incompatibilities
		{
			graph: `{
				"nodes": [{
					"plugin": {
						"name": "google-login",
						"version": "1.4"
					},
					"dependencies": [{
						"plugin": {
							"name": "mailer",
							"version": "1.6"
						}
					}]
				}]
			}`,
			plugins: `{
				"plugins": [{
					"name": "google-login",
					"version": "1.4"
				}]
			}`,
			incompatibilities: false,
		},
		// This test will check that there is an incompatibility when we require
		// a plugin which it is also required by some dependency and our requested
		// version is older than that.
		// NOTE: Some deps were removed in order to simplify the scenario being tested here.
		{
			graph: `{
				"nodes": [{
					"plugin": {
						"name": "google-login",
						"version": "1.4"
					},
					"dependencies": [{
						"plugin": {
							"name": "mailer",
							"version": "1.6"
						}
					}]
				}, {
					"plugin": {
						"name": "mailer",
						"version": "1.1"
					}
				}]
			}`,
			plugins: `{
				"plugins": [{
					"name": "google-login",
					"version": "1.4"
				}, {
					"name": "mailer",
					"version": "1.1"
				}]
			}`,
			incompatibilities: true,
		},
		// This test will check that there is an incompatibility when we require
		// a plugin which it is also an optional dep in some dependency and our requested
		// version is older than that.
		{
			graph: `{
				"nodes": [{
					"plugin": {
						"name": "copyartifact",
						"version": "1.42.1"
					},
					"optionalDependencies": [{
						"plugin": {
							"name": "maven-plugin",
							"version": "3.1.2"
						}
					}]
				}, {
					"plugin": {
						"name": "maven-plugin",
						"version": "1.466"
					}
				}]
			}`,
			plugins: `{
				"plugins": [{
					"name": "copyartifact",
					"version": "1.42.1"
				}, {
					"name": "maven-plugin",
					"version": "1.466"
				}]
			}`,
			incompatibilities: true,
		},
		// This test will check that there is no incompatibilities when we require
		// a plugin which it is also an optional dep in some dependency but our requested
		// version is newer or equal than that.
		// NOTE: Some deps were removed in order to simplify the scenario being tested here.
		{
			graph: `{
				"nodes": [{
					"plugin": {
						"name": "copyartifact",
						"version": "1.42.1"
					},
					"optionalDependencies": [{
						"plugin": {
							"name": "maven-plugin",
							"version": "3.1.2"
						}
					}]
				}, {
					"plugin": {
						"name": "maven-plugin",
						"version": "3.1.2"
					}
				}]
			}`,
			plugins: `{
				"plugins": [{
					"name": "copyartifact",
					"version": "1.42.1"
				}, {
					"name": "maven-plugin",
					"version": "3.1.2"
				}]
			}`,
			incompatibilities: false,
		},
	}

	for _, tc := range testCases {
		g := &api.Graph{}
		if err := jsonpb.UnmarshalString(tc.graph, g); err != nil {
			t.Fatalf("%+v", err)
		}
		pr := &api.PluginsRegistry{}
		if err := jsonpb.UnmarshalString(tc.plugins, pr); err != nil {
			t.Fatalf("%+v", err)
		}

		// We should pass all the tests regardless if we capture optional
		// dependencies in the lock or not.
		for _, optional := range []bool{true, false} {
			lock, err := LockPlugins(g, optional)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			incs, err := FindIncompatibilities(pr.Plugins, lock.Plugins, g)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			if !tc.incompatibilities && len(incs) > 0 {
				incs.Print()
				t.Errorf("not expected to find incompatibilities but it did (optional: %v)", optional)
			}
			if tc.incompatibilities && len(incs) == 0 {
				t.Errorf("expected to find incompatibilities but it did not (optional: %v)", optional)
			}
		}
	}
}
