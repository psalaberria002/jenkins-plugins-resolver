package utils

import (
	"os"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
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
