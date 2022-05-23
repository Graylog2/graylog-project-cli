package cmd

import (
	"github.com/Graylog2/graylog-project-cli/changelog"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var changelogCmd = &cobra.Command{
	Use:     "changelog",
	Aliases: []string{"cl"},
	Short:   "Changelog management",
	Long: `Management of repository changelogs.

Examples:
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
var changelogReleaseDate string
var changelogReleaseVersion string
var changelogProduct string

func init() {
	changelogCmd.AddCommand(changelogRenderCmd)
	RootCmd.AddCommand(changelogCmd)

	changelogRenderCmd.Flags().StringVarP(&changelogRenderFormat, "format", "f", changelog.FormatMD, "The render format. (e.g., \"md\" or \"html\")")
	changelogRenderCmd.Flags().StringVarP(&changelogReleaseDate, "date", "d", time.Now().Format("2006-01-02"), "The release date.")
	changelogRenderCmd.Flags().StringVarP(&changelogReleaseVersion, "version", "V", "0.0.0", "The release version.")
	changelogRenderCmd.Flags().StringVarP(&changelogProduct, "product", "p", "Graylog", "The product name. (e.g., \"Graylog\", \"Graylog Enterprise\")")
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

	config := changelog.Config{
		RenderFormat:   changelogRenderFormat,
		SnippetsPath:   snippetsPath,
		ReleaseDate:    changelogReleaseDate,
		ReleaseVersion: changelogReleaseVersion,
		Product:        changelogProduct,
	}

	if err := changelog.Render(config); err != nil {
		logger.Fatal(err.Error())
	}
}
