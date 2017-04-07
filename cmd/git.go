package cmd

import (
	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// gitCmd represents the git command
var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "Run git commands",
	Long:  "Run git commands in each module.",
	Run:   gitCommand,
}

func init() {
	RootCmd.AddCommand(gitCmd)

	gitCmd.Flags().BoolP("force", "f", false, "Continue to execute the git command in other modules even when it fails")
	viper.BindPFlag("git.force", gitCmd.Flags().Lookup("force"))
}

func gitCommand(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		logger.Info("Missing git command")
		cmd.UsageFunc()(cmd)
		os.Exit(1)
	}

	// Do not print any log message if stdout is no terminal
	logger.SetQuiet(!isatty.IsTerminal(os.Stdout.Fd()))

	manifestFiles := manifest.ReadState().Files()
	project := p.New(config.Get(), manifestFiles)

	logger.Info("Current manifests: %v", manifestFiles)
	logger.Info("Executing `git %v` for every selected module", strings.Join(args, " "))
	p.ForEachSelectedModule(project, func(module p.Module) {
		gitExecForPath(module, args)
	})
}

func gitExecForPath(module p.Module, args []string) {
	defer utils.Chdir(utils.GetCwd())

	logger.ColorPrintln(color.FgMagenta, "[command output: %v]", filepath.Base(module.Path))

	utils.Chdir(module.Path)

	command := exec.Command("git", args...)

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		if viper.GetBool("git.force") {
			logger.Error("Command failed, continuing in other modules: %v", err)
		} else {
			logger.Fatal("Command failed: %v", err)
		}
	}
	logger.Println("")
}
