package utils

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicallyWriteFile(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "test.txt")

	require.NoError(t, AtomicallyWriteFile(filename, []byte("hello"), 0600))

	data, err := os.ReadFile(filename)
	require.NoError(t, err)

	assert.Equal(t, []byte("hello"), data)
}
