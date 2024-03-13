package gh

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSplitRepoString(t *testing.T) {
	t.Run("Graylog2/graylog-project-cli", func(t *testing.T) {
		owner, repo, err := SplitRepoString("Graylog2/graylog-project-cli")
		require.Nil(t, err)
		assert.Equal(t, "Graylog2", owner)
		assert.Equal(t, "graylog-project-cli", repo)
	})

	t.Run("Graylog2/graylog-project-cli.git", func(t *testing.T) {
		owner, repo, err := SplitRepoString("Graylog2/graylog-project-cli.git")
		require.Nil(t, err)
		assert.Equal(t, "Graylog2", owner)
		assert.Equal(t, "graylog-project-cli", repo)
	})

	t.Run("<empty>", func(t *testing.T) {
		_, _, err := SplitRepoString("")
		require.NotNil(t, err)
	})

	t.Run("Graylog2", func(t *testing.T) {
		_, _, err := SplitRepoString("Graylog2")
		require.NotNil(t, err)
	})

	t.Run("Graylog2/foo/bar", func(t *testing.T) {
		_, _, err := SplitRepoString("Graylog2/foo/bar")
		require.NotNil(t, err)
	})
}
