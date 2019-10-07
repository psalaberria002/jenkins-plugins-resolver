package crypto

import (
	"crypto/sha256"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/juju/errors"
)

// SHA256 will return the sha256 sum of the provided protocol buffer
func SHA256(pb proto.Message) (string, error) {
	h := sha256.New()
	m := jsonpb.Marshaler{Indent: "  "}
	err := m.Marshal(h, pb)
	return fmt.Sprintf("%x", h.Sum(nil)), errors.Trace(err)
}
