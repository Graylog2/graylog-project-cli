package cmd

import (
	c "github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/projectstate"
	"github.com/spf13/cobra"
)

var regenerateCmd = &cobra.Command{
	Use:     "regenerate",
	Aliases: []string{"r"},
	Short:   "Regenerate files for the current checkout",
	Long:    "Regenerates all generated files like pom.xml and assembly descriptors",
	Run:     regenerateCommand,
}

func init() {
	RootCmd.AddCommand(regenerateCmd)
}

func regenerateCommand(cmd *cobra.Command, args []string) {
	var defaultConfig c.Config

	defaultConfig.Checkout.ManifestFiles = manifest.ReadState().Files()

	config := c.Merge(defaultConfig)

	logger.Debug("Using manifests: %v", config.Checkout.ManifestFiles)

	project := p.New(config, config.Checkout.ManifestFiles)

	projectstate.Sync(project, config)

	manifest.WriteState(config.Checkout.ManifestFiles)
}
