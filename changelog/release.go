package changelog

import (
	"fmt"
	"github.com/Graylog2/graylog-project-cli/git"
	"github.com/Graylog2/graylog-project-cli/logger"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/utils"
	"os"
	"path/filepath"
	"regexp"
)

const gitkeepContent = "# Keep the directory in Git"

func Release(project p.Project, versionPattern *regexp.Regexp) error {
	return p.ForEachSelectedModuleE(project, func(module p.Module) error {
		err := ReleaseInPath(module.Path, module.Revision, versionPattern)
		if err != nil {
			return fmt.Errorf("couldn't release changelog in path %s: %w", module.Path, err)
		}
		return nil
	})
}

func ReleaseInPath(path string, version string, versionPattern *regexp.Regexp) error {
	if !versionPattern.MatchString(version) {
		return fmt.Errorf("invalid release version: %s (pattern: %s)", version, SemverVersionPattern)
	}

	return utils.InDirectoryE(path, func() error {
		unreleasedChangelogPath := filepath.Join("changelog", "unreleased")
		versionChangelogPath := filepath.Join("changelog", version)

		if !utils.FileExists(unreleasedChangelogPath) {
			return fmt.Errorf("couldn't find unreleased changelog path: %s", filepath.Join(path, unreleasedChangelogPath))
		}

		if utils.FileExists(versionChangelogPath) {
			return fmt.Errorf("target path already exists: %s", filepath.Join(path, versionChangelogPath))
		}

		out, err := git.GitE("mv", "-v", unreleasedChangelogPath, versionChangelogPath)
		logger.Debug(out)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(unreleasedChangelogPath, 0755); err != nil {
			return fmt.Errorf("couldn't create new unreleased changelog path: %w", err)
		}

		gitKeepPath := filepath.Join(unreleasedChangelogPath, ".gitkeep")
		if err := os.WriteFile(gitKeepPath, []byte(gitkeepContent), 0644); err != nil {
			return fmt.Errorf("couldn't create new .gitkeep file: %w", err)
		}

		if _, err := git.GitE("add", unreleasedChangelogPath); err != nil {
			return fmt.Errorf("couldn't add unreleased changelog path to Git: %w", err)
		}

		commitMsg := fmt.Sprintf("Release changelog for version %s", version)
		if _, err := git.GitE("commit", "-m", commitMsg); err != nil {
			return fmt.Errorf("couldn't commit changelog release: %w", err)
		}

		return nil
	})
}
