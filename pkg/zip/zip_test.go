package zip

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestOpenFile(t *testing.T) {
	testCases := []struct {
		zip  string
		file string
		want string
	}{
		// File in root
		{"testdata/test.zip", "test/foo.txt", "hello world!\n"},
		// File in folder
		{"testdata/test.zip", "foo.txt", "hello world!\n"},
	}
	for _, tc := range testCases {
		r, err := os.Open(tc.zip)
		if err != nil {
			t.Fatalf("unable to open file %q: %+v\n", tc.zip, err)
		}
		defer r.Close()

		rc, err := OpenFile(r, tc.file)
		if err != nil {
			t.Errorf("unable to open %s file from %s: %+v", tc.file, tc.zip, err)
		}
		defer rc.Close()

		data, err := ioutil.ReadAll(rc)
		if err != nil {
			t.Errorf("unable to read data from %s: %+v", tc.file, err)
		}

		if string(data) != tc.want {
			t.Errorf("wanted: %q, got: %q\n", tc.want, string(data))
		}
	}
}
