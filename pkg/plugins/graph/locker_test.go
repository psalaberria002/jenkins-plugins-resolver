package graph

import (
	"testing"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

func TestLockPlugins(t *testing.T) {
	testCases := []struct {
		graph    string
		lock     string
		optional bool
	}{
		// Regular scenario; direct requirements get added
		{
			`{
				"nodes": [{
					"plugin": {
						"name": "foo",
						"version": "1.0"
					}
				}]
			}`, `{
				"plugins": [{
					"name": "foo",
					"version": "1.0"
				}]
			}`, false,
		},
		// Depedency scenario; dependencies get added
		{
			`{
				"nodes": [{
					"plugin": {
						"name": "foo",
						"version": "1.0"
					},
					"dependencies": [{
						"plugin": {
							"name": "bar",
							"version": "2.0"
						}
					}]
				}, {
					"plugin": {
					  "name": "bar",
					  "version": "3.0"
					}
				}]
			}`, `{
				"plugins": [{
					"name": "bar",
					"version": "3.0"
				}, {
					"name": "foo",
					"version": "1.0"
				}]
			}`, false,
		},
		// Optional dependencies scenario; optional dependencies do not
		// get added if we don't ask for it (optional flag disabled)
		{
			`{
				"nodes": [{
					"plugin": {
						"name": "foo",
						"version": "1.0"
					},
					"optional_dependencies": [{
						"plugin": {
							"name": "bar",
							"version": "2.0"
						}
					}]
				}]
			}`, `{
				"plugins": [{
					"name": "foo",
					"version": "1.0"
				}]
			}`, false,
		},
		// Optional dependencies scenario; optional dependencies get
		// added if we ask for it (optional flag enabled)
		{
			`{
				"nodes": [{
					"plugin": {
						"name": "foo",
						"version": "1.0"
					},
					"optional_dependencies": [{
						"plugin": {
							"name": "bar",
							"version": "2.0"
						}
					}]
				}]
			}`, `{
				"plugins": [{
					"name": "bar",
					"version": "2.0"
				}, {
					"name": "foo",
					"version": "1.0"
				}]
			}`, true,
		},
		// Optional dependencies scenario; locked versions get
		// updated to the newest version even if the latest version
		// is required by an optional dependency, but there is a dependency
		// that requires the plugin.
		{
			`{
				"nodes": [{
					"plugin": {
						"name": "foo",
						"version": "1.0"
					},
					"dependencies": [{
						"plugin": {
							"name": "faa",
							"version": "1.0"
						},
						"dependencies": [{
							"plugin": {
								"name": "bar",
								"version": "1.0"
							}
						}]
					}],
					"optional_dependencies": [{
						"plugin": {
							"name": "bar",
							"version": "2.0"
						}
					}]
				}]
			}`, `{
				"plugins": [{
					"name": "bar",
					"version": "2.0"
				}, {
					"name": "faa",
					"version": "1.0"
				}, {
					"name": "foo",
					"version": "1.0"
				}]
			}`, false,
		},
		// Optional dependencies scenario; locked versions get
		// updated to the newest version even if the latest version
		// is required by an optional dependency, but there is a direct
		// requirement that requires the plugin.
		{
			`{
				"nodes": [{
					"plugin": {
						"name": "foo",
						"version": "1.0"
					},
					"optional_dependencies": [{
						"plugin": {
							"name": "bar",
							"version": "2.0"
						}
					}]
				}, {
					"plugin": {
						"name": "faa",
						"version": "1.0"
					},
					"dependencies": [{
						"plugin": {
							"name": "bar",
							"version": "1.0"
						}
					}]
				}]
			}`, `{
				"plugins": [{
					"name": "bar",
					"version": "2.0"
				}, {
					"name": "faa",
					"version": "1.0"
				}, {
					"name": "foo",
					"version": "1.0"
				}]
			}`, false,
		},
	}

	for _, tc := range testCases {
		graph := &api.Graph{}
		if err := jsonpb.UnmarshalString(tc.graph, graph); err != nil {
			t.Fatalf("%+v", err)
		}
		wantLock := &api.PluginsRegistry{}
		if err := jsonpb.UnmarshalString(tc.lock, wantLock); err != nil {
			t.Fatalf("%+v", err)
		}

		lock, err := LockPlugins(graph, tc.optional)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		want := proto.MarshalTextString(wantLock)
		got := proto.MarshalTextString(lock)
		if got != want {
			t.Errorf("wanted: %q, got: %q", want, got)
		}
	}
}
