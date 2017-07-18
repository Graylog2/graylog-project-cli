package cmd

import (
	"github.com/Graylog2/graylog-project-cli/apply"
	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	"github.com/Graylog2/graylog-project-cli/pom"
	"github.com/Graylog2/graylog-project-cli/project"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
)

var graylogVersionCmd = &cobra.Command{
	Use:     "graylog-version",
	Aliases: []string{"gv"},
	Short:   "Sets the graylog version",
	Long: `This command sets the given Graylog version in all pom.xml files.

Examples:
    # Set Graylog version for all modules
    graylog-project graylog-version --set 2.3.0

    # Set Graylog version for given modules
    graylog-project -M map,pipeline graylog-version --set 2.3.0
`,
	Run: graylogVersionCommand,
}

var graylogVersion string

func init() {
	RootCmd.AddCommand(graylogVersionCmd)

	graylogVersionCmd.Flags().StringVarP(&graylogVersion, "set", "", "", "Set Graylog version")
}

func graylogVersionCommand(cmd *cobra.Command, args []string) {
	if graylogVersion == "" {
		logger.Info("Missing version option for --set")
		cmd.UsageFunc()(cmd)
		os.Exit(1)
	}

	cfg := config.Get()
	manifestFiles := manifest.ReadState().Files()
	proj := project.New(cfg, manifestFiles)

	msg := func(message string) {
		logger.ColorInfo(color.FgYellow, "===> %s", message)
	}

	applier := apply.NewExecuteApplier([]string{})

	// Set version in all modules
	msg("Setting version in all modules")
	apply.ForEachModule(proj, false, func(module project.Module) {
		applyManifestInDirectory(module.Path, func() {
			applier.MavenVersionsSet(graylogVersion)
		})

		// Update all versions after each change!
		applyManifestUpdateVersions(msg, proj, applier)
	})

	// Regenerate the graylog-project pom and assembly files to get the latest versions
	msg("Regenerate pom and assembly templates")
	pom.WriteTemplates(cfg, proj)
}
