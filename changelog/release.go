package changelog

import (
	"fmt"
	"github.com/Graylog2/graylog-project-cli/git"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/utils"
	"path/filepath"
)

func Release(project p.Project) error {
	return p.ForEachSelectedModuleE(project, func(module p.Module) error {
		err := utils.InDirectoryE(module.Path, func() error {
			unreleasedChangelogPath := filepath.Join("changelog", "unreleased")
			versionChangelogPath := filepath.Join("changelog", module.Revision)

			if !utils.FileExists(unreleasedChangelogPath) {
				// Nothing to do
				return nil
			}

			fmt.Printf("%s --> %s\n", unreleasedChangelogPath, versionChangelogPath)

			if module.ApplyExecute {
				out, err := git.GitE("mv", "-v", unreleasedChangelogPath, versionChangelogPath)
				fmt.Println(out)
				if err != nil {
					return err
				}
			} else {
				// Use Git's "-n" parameter to simulate the rename operation
				out, err := git.GitE("mv", "-v", "-n", unreleasedChangelogPath, versionChangelogPath)
				fmt.Println(out)
				if err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
}
