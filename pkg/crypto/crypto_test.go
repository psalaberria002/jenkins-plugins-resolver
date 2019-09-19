package crypto

import (
	"testing"
)

func TestSHA256(t *testing.T) {
	testCases := []struct {
		file string
		want string
	}{
		{"testdata/foo.txt", "ecf701f727d9e2d77c4aa49ac6fbbcc997278aca010bddeeb961c10cf54d435a"},
	}
	for _, tc := range testCases {
		got, err := SHA256(tc.file)
		if err != nil {
			t.Fatalf("%+v\n", err)
		}

		if got != tc.want {
			t.Errorf("wanted: %q, got: %q\n", tc.want, got)
		}
	}
}
