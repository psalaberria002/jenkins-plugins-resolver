package crypto

import (
	"testing"

	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/utils/testdata/example"
	"github.com/golang/protobuf/proto"
)

func TestSHA256(t *testing.T) {
	testCases := []struct {
		pb   proto.Message
		want string
	}{
		{
			pb: &example.Test{Foo: 123, Bar: "string"},
			// $ cat -t test
			// {
			//   "foo": 123,
			//   "bar": "string"
			// }%
			// $ sha256sum test
			// 9748e3d228813b8d2f305a08de6ee38165e797a683714a6eb0dcd79f0c99b06d  test
			want: "9748e3d228813b8d2f305a08de6ee38165e797a683714a6eb0dcd79f0c99b06d",
		},
	}
	for _, tc := range testCases {
		got, err := SHA256(tc.pb)
		if err != nil {
			t.Fatalf("%+v\n", err)
		}

		if got != tc.want {
			t.Errorf("wanted: %q, got: %q\n", tc.want, got)
		}
	}
}
