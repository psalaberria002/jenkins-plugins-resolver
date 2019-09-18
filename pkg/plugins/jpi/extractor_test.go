package jpi

import (
	"os"
	"testing"
)

func TestExtractManifest(t *testing.T) {
	testCases := []struct {
		file string
		want string
	}{
		{"testdata/foo.zip", "hello world!\r\n"},
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

func TestGetFileMimeType(t *testing.T) {
	testCases := []struct {
		file string
		want string
	}{
		{"testdata/jpis/credentials-2.2.0.jpi", "application/zip"},
		{"testdata/foo.zip", "application/zip"},
	}

	for _, tc := range testCases {
		r, err := os.Open(tc.file)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		defer r.Close()

		got, err := getFileMimeType(r)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if got != tc.want {
			t.Errorf("wanted: %s, got: %s", tc.want, got)
		}
	}
}
