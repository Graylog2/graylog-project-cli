package cmd

import (
	"fmt"
	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/project"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"os"
)

var parseManifestCmd = &cobra.Command{
	Hidden: true,
	Use:    "dump-project",
	Short:  "Dump parsed project state",
	Long:   "Parses the given manifest and dumps the generated project object.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			logger.Info("Missing manifest argument")
			cmd.UsageFunc()(cmd)
			os.Exit(1)
		}

		cfg := config.Get()
		cfg.Checkout.ManifestFiles = args[0:]
		projectData := project.New(cfg, cfg.Checkout.ManifestFiles)

		logger.Println("#### Project ####")
		logger.Println(spew.Sdump(projectData))
		logger.Println("")
		logger.Println("#### Maven Dependencies ####")
		logger.Println(spew.Sdump(project.MavenDependencies(projectData)))
		logger.Println("")
		logger.Println("#### Maven Assemblies ####")
		logger.Println(spew.Sdump(dumpMavenAssemblies(projectData)))
		logger.Println("")
	},
}

func dumpMavenAssemblies(p project.Project) []string {
	dependencies := make([]string, 0)

	project.ForEachModuleOrSubmodules(p, func(module project.Module) {
		if module.IsMavenModule() && module.Assembly {
			dependencies = append(dependencies, fmt.Sprintf("%s:%s", module.GroupId(), module.ArtifactId()))
		}
	})

	return dependencies
}

func init() {
	RootCmd.AddCommand(parseManifestCmd)
}
