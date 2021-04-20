package jpi

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/common"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/downloader/testdownloader"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/utils"
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
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.plugin.Name), func(t *testing.T) {
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
		})
	}
}

func TestFetchPlugin(t *testing.T) {
	workingDir := filepath.Join(os.TempDir(), ".jenkins")
	defer os.RemoveAll(workingDir)
	if err := common.EnsureStorePathExists(workingDir, GetStorePath); err != nil {
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
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.plugin.Name), func(t *testing.T) {
			err := FetchPlugin(tc.plugin, d, workingDir)

			if tc.shouldFetch {
				if err != nil {
					t.Errorf("expected to fetch the plugin %s but it could not: %+v", tc.plugin, err)
				}
				return
			}

			if err == nil {
				t.Errorf("not expected to fetch the plugin %s but it could", tc.plugin)
			}

			pluginPath := GetPluginPath(tc.plugin, workingDir)
			if ok, err := utils.FileExists(pluginPath); err != nil {
				t.Fatal(err)
			} else if ok {
				t.Errorf("not expected to download the plugin %q from %q but %q file exists", tc.plugin, d.GetDownloadURL(tc.plugin), pluginPath)
			}
		})
	}
}

func TestRunWorkersPoll(t *testing.T) {
	workingDir := filepath.Join(os.TempDir(), ".jenkins")
	defer os.RemoveAll(workingDir)
	if err := common.EnsureStorePathExists(workingDir, GetStorePath); err != nil {
		t.Fatalf("%+v", err)
	}

	d := testdownloader.NewDownloader("testdata/jpis", testPlugins)
	defer d.FileServer.Close()
	defer d.MuxServer.Close()

	maxWorkers := 2

	testCases := []struct {
		plugins   *api.PluginsRegistry
		shouldEnd bool
	}{
		{&api.PluginsRegistry{Plugins: []*api.Plugin{
			&api.Plugin{Name: "credentials", Version: "2.2.0"},
			&api.Plugin{Name: "structs", Version: "1.7"},
		}}, true},
		{&api.PluginsRegistry{Plugins: []*api.Plugin{
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
