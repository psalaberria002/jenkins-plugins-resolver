package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/google/go-jsonnet"
	"github.com/hashicorp/go-version"
	"github.com/juju/errors"
	"github.com/mkmik/multierror"
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

func continuousDeliveryVersionLower(vi string, vj string) (bool, error) {
	re := regexp.MustCompile(`^([0-9.]+)\.v[a-z0-9_]+$`)
	matchi := re.FindStringSubmatch(vi)
	if matchi == nil {
		return false, errors.Errorf("unable to parse version %q: It does not match %s", vi, re.String())
	}

	matchj := re.FindStringSubmatch(vj)
	if matchj == nil {
		return false, errors.Errorf("unable to parse version %q: It does not match %s", vj, re.String())
	}

	viSplitByDots := strings.Split(matchi[1], ".")
	vjSplitByDots := strings.Split(matchj[1], ".")
	if len(viSplitByDots) == len(vjSplitByDots) {
		return versionLower(matchi[1], matchj[1])
	} else if len(viSplitByDots) < len(vjSplitByDots) {
		return true, nil
	} else {
		return false, nil
	}
}

// VersionLower returns whether i version is lower than j version
func VersionLower(i string, j string) (bool, error) {

	// compare differently if continuousDeliveryVersioning
	re := regexp.MustCompile(`^([0-9.]+)\.v[a-z0-9_]+$`)
	matchi := re.FindStringSubmatch(i)
	matchj := re.FindStringSubmatch(j)

	// if both use same versioning, compare.
	// continuousDeliveryVersioning is always newer than semantic (need prove?)
	if matchi != nil && matchj != nil {
		return continuousDeliveryVersionLower(i, j)
	} else if matchi != nil {
		return false, nil
	} else if matchj != nil {
		return true, nil
	}

	// both versions not matching regex. probably semantic (check aswell?)
	var errs error

	lower, err := versionLower(i, j)
	if err == nil {
		return lower, nil
	}
	errs = multierror.Append(errs, err)

	return false, errs
}

func versionLower(i string, j string) (bool, error) {
	vj, err := version.NewVersion(j)
	if err != nil {
		return false, errors.Errorf("unable to parse version %s: %s", j, err)
	}

	// When comparing bundled plugins to requested plugins,
	// the bundled plugin version can be empty
	if i == "" && j != "" {
		return true, nil
	}

	vi, err := version.NewVersion(i)
	if err != nil {
		return false, errors.Errorf("unable to parse version %s: %s", i, err)
	}

	return vi.LessThan(vj), nil
}

type versionComparator func(i, j []string) (bool, error)

func init() {
}
