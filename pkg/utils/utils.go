package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	jsonnet "github.com/google/go-jsonnet"
	version "github.com/hashicorp/go-version"
	"github.com/juju/errors"
)

// FileExists will test if a file exists
func FileExists(filename string) (bool, error) {
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.Trace(err)
	}
	return true, nil
}

// UnmarshalJSON unmarshals a JSON file into a protocol buffer
func UnmarshalJSON(filename string, pb proto.Message) error {
	r, err := os.Open(filename)
	if err != nil {
		return errors.Trace(err)
	}
	defer r.Close()

	return jsonpb.Unmarshal(r, pb)
}

// UnmarshalJsonnet unmarshals a JSON file into a protocol buffer
func UnmarshalJsonnet(filename string, pb proto.Message) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Trace(err)
	}
	json, err := jsonnet.MakeVM().EvaluateSnippet(filename, string(data))
	if err != nil {
		return errors.Trace(err)
	}
	return jsonpb.UnmarshalString(json, pb)
}

// UnmarshalYAML unmarshals a YAML file into a protocol buffer
func UnmarshalYAML(filename string, pb proto.Message) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Trace(err)
	}
	// protobuf does not support YAML yet so we are transforming
	// the YAML bits to JSON in order to unmarshal with jsonpb
	if err := yaml.Unmarshal(data, pb); err != nil {
		return errors.Trace(err)
	}
	jsb, err := json.Marshal(pb)
	if err != nil {
		return errors.Trace(err)
	}
	return jsonpb.UnmarshalString(string(jsb), pb)
}

// MarshalJSON marshals a protocol buffer into a JSON file.
func MarshalJSON(filename string, pb proto.Message) error {
	f, err := os.Create(filename)
	if err != nil {
		return errors.Trace(err)
	}
	defer f.Close()

	m := &jsonpb.Marshaler{Indent: "  "}
	return m.Marshal(f, pb)
}

// UnmarshalFile unmarshals a file into a protocol buffer
func UnmarshalFile(filename string, pb proto.Message) error {
	var unmarshal func(string, proto.Message) error
	switch filepath.Ext(filename) {
	case ".json":
		unmarshal = UnmarshalJSON
	case ".jsonnet":
		unmarshal = UnmarshalJsonnet
	case ".yaml", ".yml":
		unmarshal = UnmarshalYAML
	}
	if unmarshal == nil {
		return errors.Errorf("unsupported input file type: %s\n", filename)
	}
	if err := unmarshal(filename, pb); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// VersionLower returns whether i version is lower than j version
func VersionLower(i string, j string) (bool, error) {
	vj, err := version.NewVersion(j)
	if err != nil {
		return false, errors.Errorf("Error parsing version %s: %s", j, err)
	}

	if i == "" && j != "" {
		return true, nil
	}

	vi, err := version.NewVersion(i)
	if err != nil {
		return false, errors.Errorf("Error parsing version %s: %s", i, err)
	}

	return vi.LessThan(vj), nil
}
