package jpi

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/downloader/testdownloader"
)

var (
	testPlugins = []*api.Plugin{
		&api.Plugin{Name: "credentials", Version: "2.2.0"},
		&api.Plugin{Name: "structs", Version: "1.7"},
	}
)

func TestDownload(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutMin*time.Minute)
	defer cancel()

	d := testdownloader.NewDownloader("testdata/jpis", testPlugins)
	defer d.FileServer.Close()
	defer d.MuxServer.Close()

	testCases := []struct {
		plugin         *api.Plugin
		shouldDownload bool
	}{
		{&api.Plugin{Name: "credentials", Version: "2.2.0"}, true},
		{&api.Plugin{Name: "structs", Version: "1.7"}, true},
		{&api.Plugin{Name: "foo", Version: "1.0"}, false},
	}
	for _, tc := range testCases {
		// Create temp file for storing the "downloaded" jpi
		file, err := ioutil.TempFile("", "plugin")
		if err != nil {
			t.Fatalf("%+v", err)
		}
		defer os.Remove(file.Name())

		w, err := os.OpenFile(file.Name(), os.O_RDWR, 0666)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		err = d.Download(ctx, tc.plugin, w)
		if tc.shouldDownload && err != nil {
			t.Errorf("expected to download the plugin %s from %s but it could not: %+v", tc.plugin, d.GetDownloadURL(tc.plugin), err)
		}
		if !tc.shouldDownload && err == nil {
			t.Errorf("not expected to download the plugin %s from %s but it could", tc.plugin, d.GetDownloadURL(tc.plugin))
		}
	}
}

func TestFetchPlugin(t *testing.T) {
	workingDir := filepath.Join(os.TempDir(), ".jenkins")
	defer os.RemoveAll(workingDir)
	if err := EnsureStorePathExists(workingDir); err != nil {
		t.Fatalf("%+v", err)
	}

	d := testdownloader.NewDownloader("testdata/jpis", testPlugins)
	defer d.FileServer.Close()
	defer d.MuxServer.Close()

	testCases := []struct {
		plugin      *api.Plugin
		shouldFetch bool
	}{
		{&api.Plugin{Name: "credentials", Version: "2.2.0"}, true},
		{&api.Plugin{Name: "structs", Version: "1.7"}, true},
		{&api.Plugin{Name: "foo", Version: "1.0"}, false},
	}
	for _, tc := range testCases {
		err := FetchPlugin(tc.plugin, d, workingDir)
		if tc.shouldFetch && err != nil {
			t.Errorf("expected to fetch the plugin %s but it could not: %+v", tc.plugin, err)
		}
		if !tc.shouldFetch && err == nil {
			t.Errorf("not expected to fetch the plugin %s but it could", tc.plugin)
		}
	}
}

func TestRunWorkersPoll(t *testing.T) {
	workingDir := filepath.Join(os.TempDir(), ".jenkins")
	defer os.RemoveAll(workingDir)
	if err := EnsureStorePathExists(workingDir); err != nil {
		t.Fatalf("%+v", err)
	}

	d := testdownloader.NewDownloader("testdata/jpis", testPlugins)
	defer d.FileServer.Close()
	defer d.MuxServer.Close()

	maxWorkers := 2

	testCases := []struct {
		plugins   *api.PluginsRequest
		shouldEnd bool
	}{
		{&api.PluginsRequest{Plugins: []*api.Plugin{
			&api.Plugin{Name: "credentials", Version: "2.2.0"},
			&api.Plugin{Name: "structs", Version: "1.7"},
		}}, true},
		{&api.PluginsRequest{Plugins: []*api.Plugin{
			&api.Plugin{Name: "credentials", Version: "2.2.0"},
			&api.Plugin{Name: "structs", Version: "1.7"},
			&api.Plugin{Name: "foo", Version: "1.0"},
		}}, false},
	}
	for _, tc := range testCases {
		err := RunWorkersPoll(tc.plugins, d, workingDir, maxWorkers)
		if tc.shouldEnd && err != nil {
			t.Errorf("expected to download all the plugins (%q) but it could not: %+v", tc.plugins, err)
		}
		if !tc.shouldEnd && err == nil {
			t.Errorf("not expected to download all the plugins (%q) but it could", tc.plugins)
		}
	}
}
