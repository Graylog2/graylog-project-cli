package cmd

import (
	"fmt"
	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/project"
	"github.com/k0kubun/pp/v3"
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
		logger.Println(pp.Sprint(projectData))
		logger.Println("")
		logger.Println("#### Maven Dependencies ####")
		logger.Println(pp.Sprint(project.MavenDependencies(projectData)))
		logger.Println("")
		logger.Println("#### Maven Assemblies ####")
		logger.Println(pp.Sprint(dumpMavenAssemblies(projectData)))
		logger.Println("")
	},
}

func dumpMavenAssemblies(p project.Project) map[string][]string {
	assemblies := make(map[string][]string)

	project.ForEachModuleOrSubmodules(p, func(module project.Module) {
		if module.IsMavenModule() && len(module.Assemblies) > 0 {
			for _, assemblyId := range module.Assemblies {
				if module.AssemblyAttachment != "" {
					assemblies[assemblyId] = append(assemblies[assemblyId], fmt.Sprintf("%s:%s:%s", module.GroupId(), module.ArtifactId(), module.AssemblyAttachment))
				} else {
					assemblies[assemblyId] = append(assemblies[assemblyId], fmt.Sprintf("%s:%s", module.GroupId(), module.ArtifactId()))
				}
			}
		}
	})

	return assemblies
}

func init() {
	RootCmd.AddCommand(parseManifestCmd)
}
