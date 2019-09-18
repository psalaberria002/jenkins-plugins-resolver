package jpi

import (
	"testing"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/golang/protobuf/proto"
)

func TestParseManifest(t *testing.T) {
	testCases := []struct {
		manifest string
		pm       *api.PluginMetadata
	}{
		{"Plugin-Version: 1.2.3\r\nShort-Name: foo\r\nLong-Name: Foo\r\nPlugin-Dependencies: bar:1.0,bar:1.1,bar:1.2;resolution:=optional\r\n",
			&api.PluginMetadata{
				Plugin:   &api.Plugin{Name: "foo", Version: "1.2.3"},
				FullName: "Foo",
				Dependencies: []*api.Plugin{
					&api.Plugin{Name: "bar", Version: "1.0"},
					&api.Plugin{Name: "bar", Version: "1.1"},
				},
				OptionalDependencies: []*api.Plugin{
					&api.Plugin{Name: "bar", Version: "1.2"},
				},
			},
		},
		{"Plugin-Version: 1.2.3\r\nShort-Name: foo\r\nLong-Name: Foo\r\n",
			&api.PluginMetadata{
				Plugin:   &api.Plugin{Name: "foo", Version: "1.2.3"},
				FullName: "Foo",
			},
		},
	}

	for _, tc := range testCases {
		pm, err := ParseManifest(tc.manifest)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		got := proto.MarshalTextString(pm)
		want := proto.MarshalTextString(tc.pm)
		if got != want {
			t.Errorf("wanted: %q, got: %q", want, got)
		}
	}
}

func TestNewDependencies(t *testing.T) {
	testCases := []struct {
		dependenciesStr      string
		dependencies         []*api.Plugin
		optionalDependencies []*api.Plugin
	}{
		{
			"bar:1.0,bar:1.1,bar:1.2;resolution:=optional",
			[]*api.Plugin{
				&api.Plugin{Name: "bar", Version: "1.0"},
				&api.Plugin{Name: "bar", Version: "1.1"},
			},
			[]*api.Plugin{
				&api.Plugin{Name: "bar", Version: "1.2"},
			},
		},
		{
			"bar:1.0,bar:1.1",
			[]*api.Plugin{
				&api.Plugin{Name: "bar", Version: "1.0"},
				&api.Plugin{Name: "bar", Version: "1.1"},
			},
			[]*api.Plugin{},
		},
		{
			"bar:1.0;resolution:=optional,bar:1.1;resolution:=optional",
			[]*api.Plugin{},
			[]*api.Plugin{
				&api.Plugin{Name: "bar", Version: "1.0"},
				&api.Plugin{Name: "bar", Version: "1.1"},
			},
		},
	}

	for _, tc := range testCases {
		deps, optDeps, err := NewDependencies(tc.dependenciesStr)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if len(tc.dependencies) != len(deps) {
			t.Errorf("wanted: %d deps, got: %d deps", len(tc.dependencies), len(deps))
		}
		if len(tc.optionalDependencies) != len(optDeps) {
			t.Errorf("wanted: %d deps, got: %d deps", len(tc.optionalDependencies), len(optDeps))
		}
		// Use api.PluginMetadata for comparisions
		got := proto.MarshalTextString(&api.PluginMetadata{Dependencies: deps, OptionalDependencies: optDeps})
		want := proto.MarshalTextString(&api.PluginMetadata{Dependencies: tc.dependencies, OptionalDependencies: tc.optionalDependencies})
		if got != want {
			t.Errorf("wanted: %s, got: %s", want, got)
		}
	}
}
