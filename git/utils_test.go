package git

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetRemoteUrl(t *testing.T) {
	repo := t.TempDir()

	require.Nil(t, Exec("init", repo))

	t.Run("WithoutOrigin", func(t *testing.T) {
		_, err := GetRemoteUrl(repo, "origin")
		assert.NotNil(t, err)
	})

	t.Run("WithoutOrigin", func(t *testing.T) {
		require.Nil(t, ExecInPath(repo, "remote", "add", "origin", "git@github.com:test/test.git"))
		url, err := GetRemoteUrl(repo, "origin")
		require.Nil(t, err)
		assert.Equal(t, "git@github.com:test/test.git", url)
	})
}
