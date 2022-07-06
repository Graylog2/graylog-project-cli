package cmd

import (
	"github.com/Graylog2/graylog-project-cli/changelog"
	c "github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"regexp"
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

var changelogReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Prepare changelogs for release.",
	Long: `Move unreleased changelog entries to a release.

Example:
    graylog-project changelog release path/to/unreleased/changelog
`,
	Run:       changelogReleaseCommand,
	ValidArgs: changelog.AvailableFormatters,
}

var changelogNewCmd = &cobra.Command{
	Use:     "new",
	Aliases: []string{"n"},
	Short:   "Create new changelog entry.",
	Long: `Create a new changelog entry based on a template.

Example:
    graylog-project changelog new path/to/unreleased/issue-123.toml
`,
	Run: changelogNewCommand,
}

var changelogRenderFormat string
var changelogDisableGitHubLinks bool
var changelogReleaseDate string
var changelogReleaseVersion string
var changelogProduct string
var changelogEntryEdit bool

func init() {
	changelogCmd.AddCommand(changelogRenderCmd)
	changelogCmd.AddCommand(changelogReleaseCmd)
	changelogCmd.AddCommand(changelogNewCmd)
	RootCmd.AddCommand(changelogCmd)

	changelogRenderCmd.Flags().StringVarP(&changelogRenderFormat, "format", "f", changelog.FormatMD, "The render format. (e.g., \"md\", \"html\", or \"d360html\")")
	changelogRenderCmd.Flags().BoolVarP(&changelogDisableGitHubLinks, "no-links", "N", false, "Do not render issue or pull-request links for entries.")
	changelogRenderCmd.Flags().StringVarP(&changelogReleaseDate, "date", "d", time.Now().Format("2006-01-02"), "The release date.")
	changelogRenderCmd.Flags().StringVarP(&changelogReleaseVersion, "version", "V", "0.0.0", "The release version.")
	changelogRenderCmd.Flags().StringVarP(&changelogProduct, "product", "p", "Graylog", "The product name. (e.g., \"Graylog\", \"Graylog Enterprise\")")

	changelogNewCmd.Flags().BoolVarP(&changelogEntryEdit, "edit", "e", false, "start $EDITOR after creating new entry")
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

	// By convention, we use the version in the snippet path if it's a valid one and no version flag is given.
	releaseVersion := changelogReleaseVersion
	if releaseVersion == "0.0.0" {
		versionPath := filepath.Base(snippetsPath)
		if regexp.MustCompile("^\\d+\\.\\d+\\.\\d+$").MatchString(versionPath) {
			releaseVersion = versionPath
		} else {
			logger.Fatal("Missing --version flag and snippets directory doesn't contain a valid version")
		}
	}

	config := changelog.Config{
		RenderFormat:      changelogRenderFormat,
		RenderGitHubLinks: !changelogDisableGitHubLinks,
		SnippetsPath:      snippetsPath,
		ReleaseDate:       changelogReleaseDate,
		ReleaseVersion:    releaseVersion,
		Product:           changelogProduct,
	}

	if err := changelog.Render(config); err != nil {
		logger.Fatal(err.Error())
	}
}
func changelogReleaseCommand(cmd *cobra.Command, args []string) {
	// TODO: We might have to take the manifest as argument
	config := c.Get()
	manifestFiles := manifest.ReadState().Files()
	project := p.New(config, manifestFiles)

	if err := changelog.Release(project); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func changelogNewCommand(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		logger.Error("Expecting a single argument")
		if err := cmd.UsageFunc()(cmd); err != nil {
			logger.Fatal(err.Error())
		}
		os.Exit(1)
	}

	if err := changelog.NewEntry(args[0], changelogEntryEdit); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
