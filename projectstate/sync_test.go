package projectstate

import (
	"os"
	"path/filepath"
	"testing"

	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteJavaVersion(t *testing.T) {
	tempDir := t.TempDir()
	// GetCwdE resolves symlinks, so resolve the temp dir too to compute the expected path.
	resolvedDir, err := filepath.EvalSymlinks(tempDir)
	require.NoError(t, err)

	t.Chdir(tempDir)

	require.NoError(t, writeJavaVersion(p.Project{JVMVersion: 17}))

	content, err := os.ReadFile(filepath.Join(resolvedDir, ".java-version"))
	require.NoError(t, err)
	assert.Equal(t, "17\n", string(content))

	info, err := os.Stat(filepath.Join(resolvedDir, ".java-version"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())
}

func TestWriteJavaVersionErrorsWhenUnset(t *testing.T) {
	tempDir := t.TempDir()
	resolvedDir, err := filepath.EvalSymlinks(tempDir)
	require.NoError(t, err)

	t.Chdir(tempDir)

	// JVMVersion defaults to 0 when a manifest omits "jvm_version".
	require.Error(t, writeJavaVersion(p.Project{}))

	// No .java-version file should have been written.
	_, err = os.Stat(filepath.Join(resolvedDir, ".java-version"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestWriteJavaVersionOverwritesExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	resolvedDir, err := filepath.EvalSymlinks(tempDir)
	require.NoError(t, err)

	filename := filepath.Join(resolvedDir, ".java-version")
	require.NoError(t, os.WriteFile(filename, []byte("11\n"), 0o644))

	t.Chdir(tempDir)

	require.NoError(t, writeJavaVersion(p.Project{JVMVersion: 21}))

	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, "21\n", string(content))
}
