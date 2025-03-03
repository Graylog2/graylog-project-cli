package gh_test

import (
	"github.com/Graylog2/graylog-project-cli/gh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestParsePullDependencies(t *testing.T) {
	writeString := func(t *testing.T, writer io.StringWriter, value string) {
		_, err := writer.WriteString(value)
		require.Nil(t, err)
	}

	tmpdir := t.TempDir()
	filename := filepath.Join(tmpdir, "body.txt")

	f, err := os.Create(filename)
	require.Nil(t, err)

	writeString(t, f, "/prdGraylog2/graylog2-server#0\n") // Non-matching line
	writeString(t, f, "/prd Graylog2/graylog2-server#1\n")
	writeString(t, f, "Something else\n\n")
	writeString(t, f, "/jpd https://github.com/Graylog2/graylog2-server/pull/100\n")
	writeString(t, f, "/jenkins-pr-deps Graylog2/graylog2-server#200\n")

	require.Nil(t, f.Close())

	file, err := os.Open(filename)
	require.Nil(t, err)

	deps, err := gh.ParsePullDependencies(file)
	require.Nil(t, err)

	assert.Exactly(t, []string{
		"Graylog2/graylog2-server#1",
		"https://github.com/Graylog2/graylog2-server/pull/100",
		"Graylog2/graylog2-server#200",
	}, deps)
}
