package cmd

import (
	c "github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var projectChangelogCmd = &cobra.Command{
	Use:     "project-changelog",
	Aliases: []string{"pcl"},
	Short:   "Project wide changelog management",
}

var projectChangelogRenderCmd = &cobra.Command{
	Use:     "render",
	Aliases: []string{"r"},
	Short:   "Render changelog snippets.",
	Long: `Render the changelog snippets for the project.

Example:
    graylog-project project-changelog render changelog/unreleased
`,
	Run: projectChangelogRenderCommand,
}

func init() {
	projectChangelogCmd.AddCommand(projectChangelogRenderCmd)

	applyChangelogRenderFlags(projectChangelogRenderCmd)

	RootCmd.AddCommand(projectChangelogCmd)
}

func projectChangelogRenderCommand(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		logger.Error("Missing snippet directory")
		if err := cmd.UsageFunc()(cmd); err != nil {
			logger.Fatal(err.Error())
		}
		os.Exit(1)
	}

	snippetDirectory := args[0]

	config := c.Get()
	manifestFiles := manifest.ReadState().Files()
	project := p.New(config, manifestFiles)

	modules := p.SelectedModules(project)

	snippetsPaths := lo.Map(modules, func(module p.Module, _ int) string {
		logger.Debug("Generating changelog for module: %s", module.Path)
		return filepath.Join(module.Path, snippetDirectory)
	})

	if err := execChangelogRenderCommand(snippetsPaths); err != nil {
		logger.Error(err.Error())
		if err := cmd.UsageFunc()(cmd); err != nil {
			logger.Fatal(err.Error())
		}
		os.Exit(1)
	}
}
