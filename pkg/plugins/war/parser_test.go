package war

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/bitnami-labs/jenkins-plugins-resolver/api"
	"github.com/bitnami-labs/jenkins-plugins-resolver/pkg/plugins/jar"
	"github.com/golang/protobuf/jsonpb"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		warfile string
		jm      *JenkinsManifest
		want    string
	}{
		{
			warfile: "testdata/jenkins.foo.war",
			jm: &JenkinsManifest{
				Version: "2.176.3",
				// The war file contains more plugins, but we are providing a manifest
				// to have a shorter test.
				PluginsFiles: []string{
					"WEB-INF/detached-plugins/command-launcher.hpi",
					"WEB-INF/detached-plugins/cvs.hpi",
					"WEB-INF/detached-plugins/junit.hpi",
				},
			},
			want: `{
 "version": "2.176.3",
 "plugins": [
  {
   "fullName": "Command Agent Launcher Plugin",
   "plugin": {
    "name": "command-launcher",
    "version": "1.0"
   },
   "dependencies": [
    {
     "name": "script-security",
     "version": "1.18.1"
    }
   ]
  },
  {
   "fullName": "Jenkins CVS Plug-in",
   "plugin": {
    "name": "cvs",
    "version": "2.11"
   }
  },
  {
   "fullName": "JUnit Plugin",
   "plugin": {
    "name": "junit",
    "version": "1.6"
   }
  }
 ]
}`,
		},
	}

	for _, tc := range testCases {
		jk, err := tc.jm.Parse(tc.warfile)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		sort.Sort(api.ByPluginMetadataName(jk.Plugins))

		m := jsonpb.Marshaler{Indent: " "}
		got, err := m.MarshalToString(jk)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if got != tc.want {
			t.Errorf("got: %q, wanted: %q", got, tc.want)
		}
	}
}

func TestParseManifest(t *testing.T) {
	testCases := []struct {
		warfile string
		want    *JenkinsManifest
	}{
		{
			warfile: "testdata/jenkins.foo.war",
			want: &JenkinsManifest{
				Version: "2.176.3",
				// We are sorting the plugins list to ensure reproducibility.
				PluginsFiles: []string{
					"WEB-INF/detached-plugins/ant.hpi",
					"WEB-INF/detached-plugins/antisamy-markup-formatter.hpi",
					"WEB-INF/detached-plugins/command-launcher.hpi",
					"WEB-INF/detached-plugins/credentials.hpi",
					"WEB-INF/detached-plugins/cvs.hpi",
					"WEB-INF/detached-plugins/display-url-api.hpi",
					"WEB-INF/detached-plugins/external-monitor-job.hpi",
					"WEB-INF/detached-plugins/javadoc.hpi",
					"WEB-INF/detached-plugins/jdk-tool.hpi",
					"WEB-INF/detached-plugins/junit.hpi",
					"WEB-INF/detached-plugins/ldap.hpi",
					"WEB-INF/detached-plugins/mailer.hpi",
					"WEB-INF/detached-plugins/matrix-auth.hpi",
					"WEB-INF/detached-plugins/matrix-project.hpi",
					"WEB-INF/detached-plugins/script-security.hpi",
					"WEB-INF/detached-plugins/ssh-credentials.hpi",
					"WEB-INF/detached-plugins/ssh-slaves.hpi",
					"WEB-INF/detached-plugins/translation.hpi",
					"WEB-INF/detached-plugins/windows-slaves.hpi",
				},
			},
		},
	}

	for _, tc := range testCases {
		manifest, err := jar.ExtractManifest(tc.warfile)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		jm, err := ParseManifest(manifest)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		sort.Strings(jm.PluginsFiles)

		got, err := json.Marshal(jm)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		want, err := json.Marshal(tc.want)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if string(got) != string(want) {
			t.Errorf("got: %+v, wanted: %+v", string(got), string(want))
		}
	}
}
