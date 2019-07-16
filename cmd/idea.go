package cmd

import (
	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/idea"
	"github.com/Graylog2/graylog-project-cli/manifest"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/spf13/cobra"
)

var ideaCmd = &cobra.Command{
	Use:   "idea",
	Short: "Commands for IntelliJ IDEA",
	Long: `
Commands to help working with the IntelliJ IDE.
`,
}

var ideaSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup IntelliJ IDEA",
	Long: `
This command will do the following:

- Add <excludeFolder url="file://$MODULE_DIR$/target/web/build" /> to the
  *.iml file of each module to avoid indexing built JavaScript files
- Add default run configurations for Graylog server and the web
  development server
`,
	Run: ideaSetupCommand,
}

func init() {
	ideaCmd.AddCommand(ideaSetupCmd)
	RootCmd.AddCommand(ideaCmd)
}

func ideaSetupCommand(cmd *cobra.Command, args []string) {
	manifestFiles := manifest.ReadState().Files()
	project := p.New(config.Get(), manifestFiles)

	idea.Setup(project)
}
