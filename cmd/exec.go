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
	"runtime"
	"strconv"
	"strings"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute arbitrary commands",
	Long: `Execute arbitrary commands in modules.

The command has access to the following environment variables:

- GPC_MODULE_NAME: Name of the module
- GPC_MODULE_PATH: Path to the module
- GPC_MODULE_REPO: Repository URL of the module
- GPC_MODULE_REVISION: Branch or commit revision of the module
- GPC_MODULE_GROUP_ID: Maven group ID of the module
- GPC_MODULE_ARTIFACT_ID: Maven artifact ID of the module
- GPC_MODULE_VERSION: Maven version of the module
- GPC_MODULE_SERVER: Whether the module is a server module
- GPC_MODULE_SKIP_RELEASE: Whether the module is skipped for release
`,
	Run: execCommand,
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
		logger.Error("Missing command")
		_ = cmd.Help()
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

	var command *exec.Cmd
	if runtime.GOOS == "windows" {
		command = exec.Command("cmd.exe", "/c", strings.Join(args, " "))
	} else {
		command = exec.Command("sh", "-c", strings.Join(args, " "))
	}

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Env = append(
		os.Environ(),
		"GPC_MODULE_NAME="+module.Name,
		"GPC_MODULE_PATH="+module.Path,
		"GPC_MODULE_REPO="+module.Repository,
		"GPC_MODULE_REVISION="+module.Revision,
		"GPC_MODULE_GROUP_ID="+module.GroupId(),
		"GPC_MODULE_ARTIFACT_ID="+module.ArtifactId(),
		"GPC_MODULE_VERSION="+module.Version(),
		"GPC_MODULE_SERVER="+strconv.FormatBool(module.Server),
		"GPC_MODULE_SKIP_RELEASE="+strconv.FormatBool(module.SkipRelease),
	)

	if err := command.Run(); err != nil {
		if viper.GetBool("exec.force") {
			logger.Error("Command failed, continuing in other modules: %v", err)
		} else {
			logger.Fatal("Command failed: %v", err)
		}
	}
	logger.Println("")
}
