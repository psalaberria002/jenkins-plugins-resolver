package crypto

import (
	"crypto/sha256"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/juju/errors"
)

// SHA256 will return the sha256 sum of the provided protocol buffer
func SHA256(pb proto.Message) (string, error) {
	var b proto.Buffer
	b.SetDeterministic(true)
	err := b.Marshal(pb)
	if err != nil {
		return "", errors.Trace(err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(b.Bytes())), nil
}
