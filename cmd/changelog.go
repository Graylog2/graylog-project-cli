package cmd

import (
	"github.com/Graylog2/graylog-project-cli/changelog"
	c "github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/samber/lo"
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
	Hidden: true, // TODO: Show command once it's fully implemented
	Use:    "release",
	Short:  "Prepare changelogs for release.",
	Long: `Move unreleased changelog entries to a release.

Example:
    graylog-project changelog release path/to/unreleased/changelog
`,
	Run:       changelogReleaseCommand,
	ValidArgs: changelog.AvailableFormatters,
}

var changelogNewCmd = &cobra.Command{
	Use:     "new [flags] changelog/unreleased/(issue|pr)-<num>.toml",
	Aliases: []string{"n"},
	Example: `
  graylog-project changelog new changelog/unreleased/issue-123.toml
  graylog-project changelog new changelog/unreleased/pr-456.toml`,
	Short: "Create new changelog entry.",
	Args:  cobra.ExactArgs(1),
	Long:  "Create a new changelog entry based on a template.",
	Run:   changelogNewCommand,
}

var changelogLintCmd = &cobra.Command{
	Use:     "lint [flags] changelog/unreleased[/(issue|pr)-<num>.toml]",
	Aliases: []string{"l"},
	Example: `
  graylog-project changelog lint changelog/unreleased
  graylog-project changelog lint changelog/unreleased/pr-456.toml`,
	Short: "Check changelog entry for syntax and content errors.",
	Args:  cobra.MinimumNArgs(1),
	Long:  "Checks a changelog entry for syntax and content errors.",
	Run:   changelogLintCommand,
}

var changelogRenderFormat string
var changelogDisableGitHubLinks bool
var changelogReleaseDate string
var changelogReleaseVersion string
var changelogReleaseVersionPattern string
var changelogProduct string
var changelogEntryEdit bool
var changelogEntryMinimalTemplate bool
var changelogEntryInteractive bool

func init() {
	changelogCmd.AddCommand(changelogRenderCmd)
	changelogCmd.AddCommand(changelogReleaseCmd)
	changelogCmd.AddCommand(changelogNewCmd)
	changelogCmd.AddCommand(changelogLintCmd)
	RootCmd.AddCommand(changelogCmd)

	changelogRenderCmd.Flags().StringVarP(&changelogRenderFormat, "format", "f", changelog.FormatMD, "The render format. (e.g., \"md\", \"html\", or \"d360html\")")
	changelogRenderCmd.Flags().BoolVarP(&changelogDisableGitHubLinks, "no-links", "N", false, "Do not render issue or pull-request links for entries.")
	changelogRenderCmd.Flags().StringVarP(&changelogReleaseDate, "date", "d", time.Now().Format("2006-01-02"), "The release date.")
	changelogRenderCmd.Flags().StringVarP(&changelogReleaseVersion, "version", "V", "0.0.0", "The release version.")
	changelogRenderCmd.Flags().StringVarP(&changelogReleaseVersionPattern, "version-pattern", "P", changelog.SemverVersionPattern.String(), "version number pattern")
	changelogRenderCmd.Flags().StringVarP(&changelogProduct, "product", "p", "Graylog", "The product name. (e.g., \"Graylog\", \"Graylog Enterprise\")")

	changelogNewCmd.Flags().BoolVarP(&changelogEntryEdit, "edit", "e", false, "Start $EDITOR after creating new entry")
	changelogNewCmd.Flags().BoolVarP(&changelogEntryMinimalTemplate, "minimal-template", "m", false, "Use a minimal entry template")
	changelogNewCmd.Flags().BoolVarP(&changelogEntryInteractive, "interactive", "i", false, "Fill template values interactively")

	changelogReleaseCmd.Flags().StringVarP(&changelogReleaseVersionPattern, "version-pattern", "P", changelog.SemverVersionPattern.String(), "version number pattern")
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

	snippetsPaths := lo.Map[string, string](args, func(arg string, _ int) string {
		path, err := filepath.Abs(arg)
		if err != nil {
			logger.Fatal("couldn't get absolute path for %s", arg)
		}
		return path
	})

	versionPattern, err := regexp.Compile(changelogReleaseVersionPattern)
	if err != nil {
		logger.Fatal("Invalid version pattern: %s", changelogReleaseVersionPattern)
	}

	// By convention, we use the version in the first snippet path if it's a valid one and no version flag is given.
	releaseVersion := changelogReleaseVersion
	if releaseVersion == "0.0.0" {
		versionPath := filepath.Base(snippetsPaths[0])
		if versionPattern.MatchString(versionPath) {
			releaseVersion = versionPath
		} else {
			logger.Fatal("Missing --version flag and snippets directory doesn't contain a valid version")
		}
	}

	if !versionPattern.MatchString(releaseVersion) {
		logger.Fatal("Invalid version: %s", releaseVersion)
	}

	config := changelog.Config{
		RenderFormat:      changelogRenderFormat,
		RenderGitHubLinks: !changelogDisableGitHubLinks,
		SnippetsPaths:     snippetsPaths,
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

	versionPattern, err := regexp.Compile(changelogReleaseVersionPattern)
	if err != nil {
		logger.Fatal("Invalid version pattern: %s", changelogReleaseVersionPattern)
	}

	if err := changelog.Release(project, versionPattern); err != nil {
		logger.Fatal(err.Error())
	}
}

func changelogNewCommand(cmd *cobra.Command, args []string) {
	if err := changelog.NewEntry(args[0], changelogEntryEdit, changelogEntryMinimalTemplate, changelogEntryInteractive); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
func changelogLintCommand(cmd *cobra.Command, args []string) {
	if err := changelog.LintPaths(args); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
