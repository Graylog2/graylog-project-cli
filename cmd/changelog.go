package cmd

import (
	"github.com/Graylog2/graylog-project-cli/changelog"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

var changelogCmd = &cobra.Command{
	Use:     "changelog",
	Aliases: []string{"cl"},
	Short:   "Changelog management",
	Long: `Management of repository changelogs.

Examples:
    # Create new changelog snippet based on a built-in template
    graylog-project changelog create

    # Render changelog for the given directory
    graylog-project changelog render path/to/snippets
`,
}

var changelogRenderCmd = &cobra.Command{
	Use:     "render",
	Aliases: []string{"r"},
	Short:   "Render changelog snippets.",
	Long: `Render the changelog snippets in the given directory.

Example:
    graylog-project changelog render path/to/snippets
`,
	Run:       changelogRenderCommand,
	ValidArgs: changelog.AvailableFormatters,
}

var changelogRenderFormat string

func init() {
	changelogCmd.AddCommand(changelogRenderCmd)
	RootCmd.AddCommand(changelogCmd)

	changelogRenderCmd.Flags().StringVarP(&changelogRenderFormat, "format", "f", "md", "The render format. (e.g., \"md\" or \"html\")")
}

func changelogRenderCommand(cmd *cobra.Command, args []string) {
	validFormat := false
	for _, v := range cmd.ValidArgs {
		if changelogRenderFormat == v {
			validFormat = true
			break
		}
	}
	if !validFormat {
		logger.Fatal("Invalid render format: %s (available: %s)", changelogRenderFormat, strings.Join(cmd.ValidArgs, ", "))
	}

	if len(args) == 0 {
		logger.Error("Missing snippets directory")
		if err := cmd.UsageFunc()(cmd); err != nil {
			logger.Fatal(err.Error())
		}
		os.Exit(1)
	}

	snippetsPath, err := filepath.Abs(args[0])
	if err != nil {
		logger.Fatal(err.Error())
	}

	if err := changelog.Render(changelogRenderFormat, snippetsPath); err != nil {
		logger.Fatal(err.Error())
	}
}
