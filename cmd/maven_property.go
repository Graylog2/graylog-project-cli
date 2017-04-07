package cmd

import (
	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	"github.com/Graylog2/graylog-project-cli/pom"
	"github.com/Graylog2/graylog-project-cli/pomparse"
	"github.com/Graylog2/graylog-project-cli/project"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strconv"
)

var mavenPropertyCmd = &cobra.Command{
	Use:     "maven-property",
	Aliases: []string{"prop"},
	Short:   "Gets or sets a maven property",
	Long: `This command can get or set a maven property in the project's pom.xml file.

Examples:
    # Get property value for all modules
    graylog-project maven-property graylog.version

    # Get property value for given modules
    graylog-project -M map,pipeline maven-property graylog.version

    # Sets the graylog.version property for all modules to 2.1.1
    graylog-project maven-property --set graylog.version 2.1.1

    # Sets the graylog.version property in the map-widget module to 2.1.1
    graylog-project -M map-widget maven-property --set graylog.version 2.1.1
`,
	Run: mavenPropertyCommand,
}

var mavenPropertySet bool
var mavenPropertyAll bool

func init() {
	RootCmd.AddCommand(mavenPropertyCmd)

	mavenPropertyCmd.Flags().BoolVarP(&mavenPropertySet, "set", "", false, "Set property. Requires a second argument that is the new value.")
	mavenPropertyCmd.Flags().BoolVarP(&mavenPropertyAll, "all", "a", false, "Show all properties")
}

func mavenPropertyCommand(cmd *cobra.Command, args []string) {
	if len(args) < 1 && !mavenPropertyAll {
		logger.Info("Missing property argument")
		cmd.UsageFunc()(cmd)
		os.Exit(1)
	}
	if mavenPropertySet && len(args) != 2 {
		logger.Info("Invalid arguments. Setting a property requires the property name and value")
		cmd.UsageFunc()(cmd)
		os.Exit(1)
	}

	config := config.Get()
	manifestFiles := manifest.ReadState().Files()
	p := project.New(config, manifestFiles)

	maxLength := project.MaxModuleNameLength(p)

	project.ForEachSelectedModuleAndSubmodules(p, func(module project.Module) {
		if !module.IsMavenModule() {
			return
		}

		pomObject := pomparse.ParsePom(filepath.Join(module.Path, "pom.xml"))
		propertyMap := pomObject.PropertiesMap()

		if mavenPropertyAll {
			logger.ColorPrintln(color.FgBlue, "[%v]", module.Name)
			for name, value := range propertyMap {
				logger.Info("    %s=%v", name, value)
			}
			return
		}

		if mavenPropertySet {
			pom.SetProperty(module, args[0], args[1])
		} else {
			logger.Info("%-"+strconv.Itoa(int(maxLength))+"s  %s=%v", module.Name, args[0], propertyMap[args[0]])
		}
	})
}
