package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var yarnCmd = &cobra.Command{
	Use:   "yarn",
	Short: "Run yarn commands",
	Long: `
Runs yarn commands in javascript projects. It checks for the presence of a package.json file.

Example:

# Run "yarn install" in every module
$ graylog-project yarn install
`,
	Run: yarnCommand,
}

func init() {
	RootCmd.AddCommand(yarnCmd)
}

func yarnCommand(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		logger.Info("Missing yarn command")
		cmd.UsageFunc()(cmd)
		os.Exit(1)
	}

	manifestFiles := manifest.ReadState().Files()
	project := p.New(config.Get(), manifestFiles)

	logger.Info("Current manifests: %v", manifestFiles)
	logger.Info("Executing `yarn %v` for every selected javascript module", strings.Join(args, " "))
	p.ForEachSelectedModuleOrSubmodules(project, func(module p.Module) {
		if module.IsNpmModule() {
			yarnExecForPath(module, args)
		}
	})
}

func yarnExecForPath(module p.Module, args []string) {
	defer utils.Chdir(utils.GetCwd())

	logger.ColorPrintln(color.FgMagenta, "[command output: %v]", filepath.Base(module.Path))

	utils.Chdir(module.Path)

	command := exec.Command("yarn", args...)

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		logger.Fatal("Command failed: %v", err)
	}
	logger.Println("")
}
