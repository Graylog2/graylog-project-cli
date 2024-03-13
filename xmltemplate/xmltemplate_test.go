package xmltemplate

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"text/template"
)

func TestVersionTemplateFuncs2(t *testing.T) {
	tests := []struct {
		serverVersion string
		function      string
		version       string
		rendered      bool
	}{
		{"1.2.3", "versionGt", "1.2.3", false},
		{"1.2.3", "versionGt", "1.3.0", false},
		{"1.2.3", "versionGt", "1.2.0", true},
		{"6.0.0", "versionGt", "6.0.0-SNAPSHOT", true},
		{"6.0.0-rc.1", "versionGt", "6.0.0-SNAPSHOT", true},

		{"1.2.3", "versionGte", "1.2.3", true},
		{"1.2.3", "versionGte", "1.2.0", true},
		{"1.2.3", "versionGte", "1.3.0", false},
		{"6.0.0-SNAPSHOT", "versionGte", "6.0.0", false},
		{"6.0.0-alpha.1", "versionGte", "6.0.0-SNAPSHOT", true},
		{"6.0.0-SNAPSHOT", "versionGte", "6.0.0-SNAPSHOT", true},
		{"6.0.0-beta.1", "versionGte", "6.0.0-alpha.2", true},
		{"6.0.0", "versionGte", "6.0.0-alpha.2", true},
		{"6.0.0", "versionGte", "6.0.0-SNAPSHOT", true},

		{"1.2.3", "versionLt", "1.2.3", false},
		{"1.2.3", "versionLt", "1.3.0", true},
		{"1.2.3", "versionLt", "1.2.0", false},
		{"6.0.0-SNAPSHOT", "versionLt", "6.0.0-alpha.1", true},

		{"1.2.3", "versionLte", "1.2.3", true},
		{"1.2.3", "versionLte", "1.2.0", false},
		{"1.2.3", "versionLte", "1.3.0", true},
	}
	for _, test := range tests {
		testName := fmt.Sprintf("%s-%s-%s-%t", test.serverVersion, test.function, test.version, test.rendered)

		t.Run(testName, func(t *testing.T) {
			serverVersion, err := version.NewVersion(test.serverVersion)
			require.Nil(t, err)

			tmpl, err := template.New(testName).Funcs(versionTemplateFuncs(serverVersion)).
				Parse(fmt.Sprintf(`{{ if %s "%s" }}RENDERED{{ end }}`, test.function, test.version))
			require.Nil(t, err)

			var out bytes.Buffer

			require.Nil(t, tmpl.Execute(&out, nil))

			if test.rendered {
				assert.Contains(t, out.String(), "RENDERED")
			} else {
				assert.NotContains(t, out.String(), "RENDERED")
			}
		})
	}
}
