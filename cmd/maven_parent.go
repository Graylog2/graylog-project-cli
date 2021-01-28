package cmd

import (
	"fmt"
	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	"github.com/Graylog2/graylog-project-cli/pom"
	"github.com/Graylog2/graylog-project-cli/project"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"strconv"
)

var mavenParentCmd = &cobra.Command{
	Use:     "maven-parent",
	Aliases: []string{"parent"},
	Short:   "Show or modify maven parent",
	Long: `This command can show and/or modify the maven parent entry in a pom.xml.

Examples:
    # Get parent for all selected modules
    graylog-project maven-parent

    # Set parent version for all selected modules
    graylog-project maven-parent -V 2.2.1

    # Set parent groupId, artifactId and version for all selected modules
    graylog-project maven-parent -G org.graylog.plugins -A graylog-plugin-parent -V 2.2.1
`,
	Run: mavenParentCommand,
}

var mavenSetParentGroupId string
var mavenSetParentArtifactId string
var mavenSetParentVersion string
var mavenSetParentRelativePath string

func init() {
	RootCmd.AddCommand(mavenParentCmd)

	mavenParentCmd.Flags().StringVarP(&mavenSetParentGroupId, "set-group-id", "G", "", "Set parent groupId.")
	mavenParentCmd.Flags().StringVarP(&mavenSetParentArtifactId, "set-artifact-id", "A", "", "Set parent artifactId.")
	mavenParentCmd.Flags().StringVarP(&mavenSetParentVersion, "set-version", "V", "", "Set parent version.")
	mavenParentCmd.Flags().StringVarP(&mavenSetParentRelativePath, "set-relative-path", "R", "", "Set parent relativePath.")
}

func mavenParentCommand(cmd *cobra.Command, args []string) {
	config := config.Get()
	manifestFiles := manifest.ReadState().Files()
	p := project.New(config, manifestFiles)

	maxLength := project.MaxModuleNameLength(p)

	logger.Info("Maven parents:")
	project.ForEachSelectedModule(p, func(module project.Module) {
		// The server module is supposed to be the parent of all other modules so we don't touch its parent
		if module.Server {
			logger.ColorPrintln(color.FgYellow, "    %-"+strconv.Itoa(int(maxLength))+"s  %s", module.Name, "Not setting parent for server module")
			return
		}

		if mavenSetParentGroupId != "" {
			pom.SetParent(module, mavenSetParentGroupId, module.ParentArtifactId(), module.ParentVersion(), module.ParentRelativePath())
		}
		if mavenSetParentArtifactId != "" {
			pom.SetParent(module, module.ParentGroupId(), mavenSetParentArtifactId, module.ParentVersion(), module.ParentRelativePath())
		}
		if mavenSetParentVersion != "" {
			pom.SetParent(module, module.ParentGroupId(), module.ParentArtifactId(), mavenSetParentVersion, module.ParentRelativePath())
		}
		if mavenSetParentRelativePath != "" {
			pom.SetParent(module, module.ParentGroupId(), module.ParentArtifactId(), module.ParentVersion(), mavenSetParentRelativePath)
		}

		coordinates := fmt.Sprintf("%s:%s:%s:%s", module.ParentGroupId(), module.ParentArtifactId(), module.ParentVersion(), module.ParentRelativePath())

		if coordinates == ":::" {
			logger.Info("    %-"+strconv.Itoa(int(maxLength))+"s  <none>", module.Name)
		} else {
			logger.Info("    %-"+strconv.Itoa(int(maxLength))+"s  %s", module.Name, coordinates)
		}
	})
}
