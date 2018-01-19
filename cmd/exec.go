package cmd

import (
	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute arbitrary commands",
	Long:  "Execute arbitrary commands in modules",
	Run:   execCommand,
}

func init() {
	RootCmd.AddCommand(execCmd)

	execCmd.Flags().BoolP("force", "f", false, "Continue to execute the command even when it returns a non-zero code")
	execCmd.Flags().BoolP("web", "w", false, "Exec command only in web modules")
	viper.BindPFlag("exec.force", execCmd.Flags().Lookup("force"))
	viper.BindPFlag("exec.web", execCmd.Flags().Lookup("web"))
}

func execCommand(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		logger.Info("Missing command")
		cmd.UsageFunc()(cmd)
		os.Exit(1)
	}

	manifestFiles := manifest.ReadState().Files()
	project := p.New(config.Get(), manifestFiles)

	logger.Info("Current manifests: %v", manifestFiles)

	if viper.GetBool("exec.web") {
		logger.Info("Executing `%v` for every selected web module", strings.Join(args, " "))
		p.ForEachSelectedModuleOrSubmodules(project, func(module p.Module) {
			if module.IsNpmModule() {
				execForPath(module, args)
			}
		})
	} else {
		logger.Info("Executing `%v` for every selected module", strings.Join(args, " "))
		p.ForEachSelectedModule(project, func(module p.Module) {
			execForPath(module, args)
		})
	}
}

func execForPath(module p.Module, args []string) {
	defer utils.Chdir(utils.GetCwd())

	logger.ColorPrintln(color.FgMagenta, "[command output: %v]", filepath.Base(module.Path))

	utils.Chdir(module.Path)

	command := exec.Command("sh", "-c", strings.Join(args, " "))

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		if viper.GetBool("exec.force") {
			logger.Error("Command failed, continuing in other modules: %v", err)
		} else {
			logger.Fatal("Command failed: %v", err)
		}
	}
	logger.Println("")
}
