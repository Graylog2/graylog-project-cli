package cmd

import (
	"fmt"
	"github.com/Graylog2/graylog-project-cli/selfupdate"
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"strings"
)

var selfUpdateCmd = &cobra.Command{
	Use:   "self-update [version]",
	Short: "Update the CLI command",
	RunE:  selfUpdateCommand,
}

var selfUpdateForce bool
var selfUpdateNonInteractive bool

func init() {
	selfUpdateCmd.Flags().BoolVarP(&selfUpdateForce, "force", "F", false, "Force update")
	selfUpdateCmd.Flags().BoolVarP(&selfUpdateNonInteractive, "non-interactive", "I", false, "Non-interactive mode")

	RootCmd.AddCommand(selfUpdateCmd)
}

func selfUpdateCommand(cmd *cobra.Command, args []string) error {
	// The gitTag has a "-dirty" suffix when the binary is built during development
	runningVersionString := strings.ReplaceAll(gitTag, "-dirty", "")

	runningVersion, err := version.NewVersion(runningVersionString)
	if err != nil {
		return fmt.Errorf("couldn't parse running version %q: %w", runningVersionString, err)
	}

	requestedVersion := "latest"
	if len(args) >= 1 && strings.TrimSpace(args[0]) != "" {
		requestedVersion = strings.TrimSpace(args[0])
	}

	return selfupdate.SelfUpdate(runningVersion, requestedVersion, selfUpdateForce, !selfUpdateNonInteractive)
}
