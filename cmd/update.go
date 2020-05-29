package cmd

import (
	c "github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	"github.com/Graylog2/graylog-project-cli/pom"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/repo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var updateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"up"},
	Short:   "Update all repositories for the current manifest",
	Long: `
Update all repositories for the current manifest.

This is the equivalent of executing the following git commands in every repository:

  git fetch --all --tags (--prune)
  git merge --ff-only origin/<branch-name>
`,
	Run: updateCommand,
}

func init() {
	RootCmd.AddCommand(updateCmd)

	updateCmd.Flags().BoolP("prune", "p", false, "Prune local branches that no longer exists in the remote repository. (i.e. \"git fetch --prune\")")

	viper.BindPFlag("update.prune", updateCmd.Flags().Lookup("prune"))
}

func updateCommand(cmd *cobra.Command, args []string) {
	config := c.Get()
	manifestFiles := manifest.ReadState().Files()
	repoMgr := repo.NewRepoManager(config)
	project := p.New(config, manifestFiles)

	p.ForEachSelectedModule(project, func(module p.Module) {
		logger.Info("Updating %v", module.Path)
		repoMgr.UpdateRepository(module)
	})

	pom.WriteTemplates(config, project)
}
