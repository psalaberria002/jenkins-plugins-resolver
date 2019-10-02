package jar

import (
	"testing"
)

func TestExtractManifest(t *testing.T) {
	testCases := []struct {
		file string
		want string
	}{
		{"testdata/foo.war", "hello world!\r\n"},
		{"testdata/foo.jpi", "hello world!\r\n"},
	}

	for _, tc := range testCases {
		got, err := ExtractManifest(tc.file)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if got != tc.want {
			t.Errorf("wanted: %q, got: %q", tc.want, got)
		}
	}
}
