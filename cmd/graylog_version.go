package cmd

import (
	"fmt"
	"github.com/Graylog2/graylog-project-cli/apply"
	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	"github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/projectstate"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"strings"
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
var graylogVersionShort bool
var graylogVersionTruncate bool

func init() {
	RootCmd.AddCommand(graylogVersionCmd)

	graylogVersionCmd.Flags().StringVar(&graylogVersion, "set", "", "Set Graylog version")
	graylogVersionCmd.Flags().BoolVarP(&graylogVersionShort, "short", "s", false, "Only show the version number(s)")
	graylogVersionCmd.Flags().BoolVarP(&graylogVersionTruncate, "truncate", "t", false, "Truncate any -SNAPSHOT suffix from the version")
}

func graylogVersionCommand(cmd *cobra.Command, args []string) {
	cfg := config.Get()
	manifestFiles := manifest.ReadState().Files()
	proj := project.New(cfg, manifestFiles)

	if graylogVersion == "" {
		project.ForEachSelectedModule(proj, func(module project.Module) {
			utils.InDirectory(module.Path, func() {
				version := module.Version()
				if graylogVersionTruncate {
					version = strings.TrimSuffix(module.Version(), "-SNAPSHOT")
				}
				if graylogVersionShort {
					fmt.Println(version)
				} else {
					url, err := utils.ParseGitHubURL(module.Repository)
					if err != nil {
						logger.Error("Couldn't parse repository URL: %s", err)
						return
					}
					logger.Info("%-50s %s", strings.TrimSuffix(url.Repository(), ".git"), version)
				}
			})
		})
		return
	}

	msg := func(message string) {
		logger.ColorInfo(color.FgYellow, "===> %s", message)
	}

	applier := apply.NewExecuteApplier([]string{})

	msg("Setting version in all modules")
	apply.ForEachModule(proj, false, func(module project.Module) {
		applyManifestInDirectory(module.Path, func() {
			applier.MavenVersionsSet(graylogVersion)
		})

		// Update all versions after each change!
		applyManifestUpdateVersions(msg, proj, applier)
	})

	msg("Setting version in all web modules")
	apply.ForEachModule(proj, true, func(module project.Module) {
		applyManifestInDirectory(module.Path, func() {
			applier.NpmVersionSet(module, graylogVersion)
		})
	})

	// Regenerate the graylog-project pom and assembly files to get the latest versions
	msg("Regenerate pom and assembly templates")
	projectstate.Sync(proj, cfg)
}
