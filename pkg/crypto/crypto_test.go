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
			//   "bar": "string"
			//   "foo": 123,
			// }%
			// $ sha256sum test
			// 3135baa3a6668734085a13e7d66c36a7d3ae892c5cfb6f1476ceb3c682d46427  test
			want: "3135baa3a6668734085a13e7d66c36a7d3ae892c5cfb6f1476ceb3c682d46427",
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
