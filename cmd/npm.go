package cmd

import (
	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// npmCmd represents the npm command
var npmCmd = &cobra.Command{
	Use:   "npm",
	Short: "Run npm commands",
	Long: `
Runs npm commands in npm projects. It checks for the presence of a package.json file.

Example:

# Run "npm install" in every module
$ graylog-project npm install
`,
	Run: npmCommand,
}

func init() {
	RootCmd.AddCommand(npmCmd)
}

func npmCommand(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		logger.Info("Missing npm command")
		cmd.UsageFunc()(cmd)
		os.Exit(1)
	}

	manifestFiles := manifest.ReadState().Files()
	project := p.New(config.Get(), manifestFiles)

	logger.Info("Current manifests: %v", manifestFiles)
	logger.Info("Executing `npm %v` for every selected npm module", strings.Join(args, " "))
	p.ForEachSelectedModuleOrSubmodules(project, func(module p.Module) {
		if module.IsNpmModule() {
			npmExecForPath(module, args)
		}
	})
}

func npmExecForPath(module p.Module, args []string) {
	defer utils.Chdir(utils.GetCwd())

	logger.ColorPrintln(color.FgMagenta, "[command output: %v]", filepath.Base(module.Path))

	utils.Chdir(module.Path)

	command := exec.Command("npm", args...)

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		logger.Fatal("Command failed: %v", err)
	}
	logger.Println("")
}
