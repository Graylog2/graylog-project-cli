package cmd

import (
	"github.com/Graylog2/graylog-project-cli/apply"
	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var npmVersionCmd = &cobra.Command{
	Use:     "npm-version",
	Aliases: []string{"nv"},
	Short:   "Set package.json version",
	Long: `
Sets version in all package.json files.
`,
	Run: npmVersionCommand,
}

func init() {
	RootCmd.AddCommand(npmVersionCmd)

	npmVersionCmd.Flags().StringP("set", "s", "", "New version")
	npmVersionCmd.Flags().BoolP("commit", "c", false, "If new version should be committed with Git")

	viper.BindPFlag("npm-version.set", npmVersionCmd.Flags().Lookup("set"))
	viper.BindPFlag("npm-version.commit", npmVersionCmd.Flags().Lookup("commit"))
}

func npmVersionCommand(cmd *cobra.Command, args []string) {
	version := viper.GetString("npm-version.set")
	shouldCommit := viper.GetBool("npm-version.commit")

	if version == "" {
		logger.Fatal("Missing --set argument")
		cmd.UsageFunc()(cmd)
		os.Exit(1)
	}

	manifestFiles := manifest.ReadState().Files()
	project := p.New(config.Get(), manifestFiles)
	applier := apply.NewExecuteApplier([]string{})

	apply.ForEachModule(project, true, func(module p.Module) {
		utils.InDirectory(module.Path, func() {
			applier.NpmVersionSet(module, version)
			if shouldCommit {
				applier.NpmVersionCommit(module, version)
			}
		})
	})
}
