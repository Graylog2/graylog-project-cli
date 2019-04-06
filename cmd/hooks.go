package cmd

import (
	"github.com/Graylog2/graylog-project-cli/hooks"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var hooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Manage hooks",
}

var hooksRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute hooks",
	Long:  "Executes hooks for the given name if any exist.",
	Run:   hooksRunCommand,
}

func init() {
	hooksCmd.AddCommand(hooksRunCmd)
	RootCmd.AddCommand(hooksCmd)

	hooksRunCmd.Flags().BoolP("noop", "n", false, "Don't execute hooks, only show which ones would be executed.")

	viper.BindPFlag("hooks.noop", hooksRunCmd.Flags().Lookup("noop"))
}

func hooksRunCommand(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		logger.Info("Missing hook name")
		cmd.UsageFunc()(cmd)
		os.Exit(1)
	}

	noop := viper.GetBool("hooks.noop")

	if err := hooks.Run(args[0], noop); err != nil {
		logger.Error("Running hooks failed: %s", err)
	}
}
