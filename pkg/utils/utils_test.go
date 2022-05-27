package utils

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/utils/testdata/example"
	"github.com/golang/protobuf/proto"
)

func TestFileExists(t *testing.T) {
	testCases := []struct {
		file string
		want bool
	}{
		{"testdata/foo.txt", true},
		{"testdata/bar.txt", false},
	}
	for _, tc := range testCases {
		got, err := FileExists(tc.file)
		if err != nil {
			t.Fatalf("%+v\n", err)
		}
		if got != tc.want {
			t.Errorf("%s was expected to exist (%v) but existed (%v)\n", tc.file, tc.want, got)
		}
	}
}

func TestUnmarshalJSON(t *testing.T) {
	testCases := []struct {
		file string
		want string
	}{
		{"testdata/unmarshal.json", proto.MarshalTextString(&example.Test{Foo: 123, Bar: "string"})},
		{"testdata/unmarshal.yml", proto.MarshalTextString(&example.Test{Foo: 123, Bar: "string"})},
		{"testdata/unmarshal.jsonnet", proto.MarshalTextString(&example.Test{Foo: 123, Bar: "string"})},
	}
	for _, tc := range testCases {
		msg := &example.Test{}
		switch filepath.Ext(tc.file) {
		case ".json":
			if err := UnmarshalJSON(tc.file, msg); err != nil {
				t.Fatalf("%+v\n", err)
			}
		case ".jsonnet":
			if err := UnmarshalJsonnet(tc.file, msg); err != nil {
				t.Fatalf("%+v\n", err)
			}
		case ".yml":
			if err := UnmarshalYAML(tc.file, msg); err != nil {
				t.Fatalf("%+v\n", err)
			}
		default:
			t.Fatalf("unsupported input file type: %s\n.", tc.file)
		}

		got := proto.MarshalTextString(msg)
		if got != tc.want {
			t.Errorf("got: %q, wanted: %q\n", got, tc.want)
		}
	}
}

func TestMarshalJSON(t *testing.T) {
	testCases := []struct {
		msg  proto.Message
		want string
	}{
		{&example.Test{Foo: 123, Bar: "string"}, `{
  "foo": 123,
  "bar": "string"
}`},
	}
	for _, tc := range testCases {
		// Create temp file for marshaling the message to
		file, err := ioutil.TempFile("", "marshal")
		if err != nil {
			log.Fatal(err)
		}
		defer os.Remove(file.Name())

		if err := MarshalJSON(file.Name(), tc.msg); err != nil {
			t.Fatalf("%+v\n", err)
		}

		// Read marshaled data
		r, err := os.Open(file.Name())
		if err != nil {
			t.Fatalf("%+v\n", err)
		}
		defer r.Close()
		data, err := ioutil.ReadAll(r)
		if err != nil {
			t.Fatalf("%+v\n", err)
		}
		got := string(data)

		if got != tc.want {
			t.Errorf("got: %q, wanted: %q\n", got, tc.want)
		}
	}
}

func TestContinuousDeliveryVersionLower(t *testing.T) {
	testCases := []struct {
		vi   string
		vj   string
		want bool
	}{
		{"1108.v57edf648f5d4", "2648.va9433432b33c", true},
		{"3108.v93ed", "2648.vbc92", false},
		{"3108.v93ed", "1.3108.v93ed", true},
		{"1.3108.v93ed", "3108.v93ed", false},
		{"4.6.3108.v93ed", "4.7108.v93ed", false},
		{"4.7108.v93ed", "4.6.3108.v93ed", true},
		{"5.4.7108.v93ed", "5.4.7107.v93ed", false},
		{"1.2.2.5.4.7108.v93ed", "1.2.1.5.4.7107.v93ed", false},
	}
	for _, tc := range testCases {
		got, err := continuousDeliveryVersionLower(tc.vi, tc.vj)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		msg := "greater"
		if tc.want {
			msg = "lower"
		}
		if got != tc.want {
			t.Errorf("%s should be %s than %s but it is not properly detected\n", tc.vi, msg, tc.vj)
		}
	}
}

func TestVersionLower(t *testing.T) {
	testCases := []struct {
		vi   string
		vj   string
		want bool
	}{
		// Test standard versions
		{"1.0.0", "1.0.0", false},
		{"1.0.0", "2.0.0", true},
		{"1.0.0", "1.1.0", true},
		{"1.0.0", "1.0.1", true},
		{"1.0.0", "1.0.0.1", true},
		// Test exceptions
		{"", "60.v1747b0eb632a", true},
		{"1108.v57edf648f5d4", "2648.va9433432b33c", true},
		{"3108.v93ed", "2648.vbc92", false},
		{"3108.v93ed", "1.3108.v93ed", true},
		{"1.3108.v93ed", "3108.v93ed", false},
		{"4.6.3108.v93ed", "4.7108.v93ed", false},
		{"4.7108.v93ed", "4.6.3108.v93ed", true},
		{"5.4.7108.v93ed", "5.4.7107.v93ed", false},
		{"1.2.2.5.4.7108.v93ed", "1.2.1.5.4.7107.v93ed", false},
	}
	for _, tc := range testCases {
		got, err := VersionLower(tc.vi, tc.vj)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		msg := "greater"
		if tc.want {
			msg = "lower"
		}
		if got != tc.want {
			t.Errorf("%s should be %s than %s but it is not properly detected\n", tc.vi, msg, tc.vj)
		}
	}
}
