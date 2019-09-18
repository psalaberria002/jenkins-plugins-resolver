package meta

import (
	"testing"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/golang/protobuf/proto"
)

func TestReadMetadata(t *testing.T) {
	testCases := []struct {
		file string
		pm   *api.PluginMetadata
	}{
		{"testdata/foo.meta", &api.PluginMetadata{
			FullName: "Foo",
			Plugin:   &api.Plugin{Name: "foo", Version: "1.0"},
		}},
		{"testdata/foobar.meta", &api.PluginMetadata{
			FullName: "Foo",
			Plugin:   &api.Plugin{Name: "foo", Version: "1.0"},
			Dependencies: []*api.Plugin{
				&api.Plugin{Name: "bar", Version: "1.0"},
			},
		}},
	}

	for _, tc := range testCases {
		pm, err := ReadMetadata(tc.file)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		got := proto.MarshalTextString(pm)
		want := proto.MarshalTextString(tc.pm)
		if got != want {
			t.Errorf("wanted: %s, got: %s", want, got)
		}
	}
}
