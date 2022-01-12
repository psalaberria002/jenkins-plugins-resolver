package utils

import (
	"encoding/json"
	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/google/go-jsonnet"
	"github.com/hashicorp/go-version"
	"github.com/juju/errors"
	"github.com/mkmik/multierror"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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
	var errs error

	lower, err := versionLower(i, j)
	if err == nil {
		return lower, nil
	}
	errs = multierror.Append(errs, err)

	for _, e := range exceptionExpressions {
		lower, err = versionLowerException(i, j, e)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
		return lower, nil
	}

	return false, errs
}

func versionLower(i string, j string) (bool, error) {
	vj, err := version.NewVersion(j)
	if err != nil {
		return false, errors.Errorf("unable to parse version %s: %s", j, err)
	}

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

// exceptionExpression contains a compiled regular expression and a function to test whether a version
// matching it is lower than another vesrion.
type exceptionExpression struct {
	re *regexp.Regexp
	fn versionComparator
}

// The exceptions to manage
var exceptionExpressions []*exceptionExpression

// ExceptionRegexpsRaw are the raw regular expressions that we know are exceptions to standard version formats.
var ExceptionRegexpsRaw = map[string]versionComparator{
	// Exception found at https://plugins.jenkins.io/workflow-cps/#releases
	// Example: 2648.va9433432b33c
	`([0-9]+)\.v([a-z0-9]+)`: func(i, j []string) (bool, error) {
		xi, err := strconv.Atoi(i[1])
		if err != nil {
			return false, errors.Errorf("malformed version: %s in %v is not an integer", i[1], i)
		}
		xj, err := strconv.Atoi(j[1])
		if err != nil {
			return false, errors.Errorf("malformed version: %s in %v is not an integer", i[1], i)
		}
		return xi < xj, nil
	},
}

func init() {
	for raw, fn := range ExceptionRegexpsRaw {
		re := regexp.MustCompile(raw)
		exceptionExpressions = append(exceptionExpressions, &exceptionExpression{re: re, fn: fn})
	}
}

func versionLowerException(i string, j string, exp *exceptionExpression) (bool, error) {
	im := exp.re.FindStringSubmatch(i)
	if im == nil {
		return false, errors.Errorf("unable to parse version (exception) %s: It does not match %s", i, exp.re.String())
	}

	ij := exp.re.FindStringSubmatch(j)
	if ij == nil {
		return false, errors.Errorf("unable to parse version (exception) %s: It does not match %s", j, exp.re.String())
	}

	return exp.fn(im, ij)
}
