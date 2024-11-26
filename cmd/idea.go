package cmd

import (
	"fmt"
	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/idea"
	"github.com/Graylog2/graylog-project-cli/manifest"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ideaCmd = &cobra.Command{
	Use:   "idea",
	Short: "Commands for IntelliJ IDEA",
	Long: `
Commands to help working with the IntelliJ IDE.
`,
}

var ideaSetupCmd = &cobra.Command{
	Deprecated: `use the "run-config create" command instead`,
	Use:        "setup",
	Short:      "Setup IntelliJ IDEA",
	Long: `
This command will do the following:

- Add <excludeFolder url="file://$MODULE_DIR$/target/web/build" /> to the
  *.iml file of each module to avoid indexing built JavaScript files
- Add default run configurations for Graylog server and the web
  development server
`,
	Run: ideaSetupCommand,
}

var ideaRunConfigsCmd = &cobra.Command{
	Use:     "run-configs",
	Aliases: []string{"rc"},
	Short:   "Manage IntelliJ IDEA run configurations",
}

var ideaRunConfigsCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"c"},
	Short:   "Create IntelliJ IDEA run configurations",
	Long: `This command adds default IntelliJ run configurations for Graylog Server,
Data Node, and the web development server.

The run configurations are created in the $PWD/.idea/runConfigurations/ directory.

Examples:
    # Create default run configurations
    graylog-project idea run-config create

    # Create default run configurations and .env files (requires installation of EnvFile plugin in IntelliJ)
    graylog-project idea run-config create -E

    # Create run configurations for two Server and three Data Node instances
    graylog-project idea run-config create --instances server=2,data-node=3
`,
	RunE: ideaRunConfigCreateCommand,
}

func init() {
	ideaRunConfigsCreateCmd.Flags().BoolP("force", "f", false, "Overwrite existing run configurations")
	ideaRunConfigsCreateCmd.Flags().BoolP("env-file", "E", false, "Use .env files (requires the IntelliJ EnvFile plugin)")
	ideaRunConfigsCreateCmd.Flags().StringToIntP("instances", "i", idea.DefaultInstanceCounts, "Number of instances - example: server=1,data-node=3")
	ideaRunConfigsCreateCmd.Flags().String("root-password", idea.DefaultRootPassword, "The root user password")
	ideaRunConfigsCmd.AddCommand(ideaRunConfigsCreateCmd)

	if err := viper.BindPFlags(ideaRunConfigsCreateCmd.Flags()); err != nil {
		panic(err)
	}

	ideaCmd.AddCommand(ideaSetupCmd)
	ideaCmd.AddCommand(ideaRunConfigsCmd)
	RootCmd.AddCommand(ideaCmd)
}

func ideaSetupCommand(cmd *cobra.Command, args []string) {
	manifestFiles := manifest.ReadState().Files()
	project := p.New(config.Get(), manifestFiles)

	idea.Setup(project)
}

func ideaRunConfigCreateCommand(cmd *cobra.Command, args []string) error {
	var cfg idea.RunConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("couldn't parse configuration: %w", err)
	}

	wd, err := utils.GetCwdE()
	if err != nil {
		return err
	}

	cfg.Workdir = wd

	return idea.CreateRunConfigurations(cfg)
}
