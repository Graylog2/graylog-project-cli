package cmd

import (
	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

const NodeModulesDir = "node_modules"
const NodeDir = "node"

// npmCmd represents the npm command
var npmCleanCmd = &cobra.Command{
	Use:     "npm-clean",
	Aliases: []string{"yarn-clean"},
	Short:   "Cleanup npm/yarn related state",
	Long: `
Removes several npm state folders. (e.g. node_modules)
`,
	Run: npmCleanCommand,
}

func init() {
	RootCmd.AddCommand(npmCleanCmd)
}

func npmCleanCommand(cmd *cobra.Command, args []string) {
	project := p.New(config.Get(), manifest.ReadState().Files())

	p.ForEachSelectedModuleOrSubmodules(project, func(module p.Module) {
		if !module.IsNpmModule() {
			return
		}

		utils.InDirectory(module.Path, func() {
			for _, dir := range []string{NodeModulesDir, NodeDir} {
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					return
				}

				logger.Info("Removing directory: %s", filepath.Join(module.Path, dir))
				if err := os.RemoveAll(dir); err != nil {
					logger.Error("Unable to remove directory `%s': %v", dir, err)
				}
			}
		})
	})
}
