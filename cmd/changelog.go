package cmd

import (
	"errors"
	"fmt"
	"github.com/Graylog2/graylog-project-cli/changelog"
	c "github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/git"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/hashicorp/go-version"
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
	Run: changelogRenderCommand,
}

var changelogReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Prepare changelogs for release.",
	Long: `Move unreleased changelog entries to a release.

Example:
    graylog-project changelog release path/to/unreleased/changelog
`,
	Run: changelogReleaseCommand,
}

var changelogReleasePathCmd = &cobra.Command{
	Use:   "release:path",
	Short: "Prepare changelogs for release outside a project setup.",
	Long: `Move unreleased changelog entries to a release folder.

To be used when you need to move changelogs of a single repository outside a project setup.

Example:
    graylog-project changelog release:path path/to/unreleased/changelog
`,
	Args: cobra.ExactArgs(1),
	Run:  changelogReleasePathCommand,
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
var changelogSkipHeader bool
var changelogReadStdin bool

func init() {
	changelogCmd.AddCommand(changelogRenderCmd)
	changelogCmd.AddCommand(changelogReleaseCmd)
	changelogCmd.AddCommand(changelogReleasePathCmd)
	changelogCmd.AddCommand(changelogNewCmd)
	changelogCmd.AddCommand(changelogLintCmd)
	RootCmd.AddCommand(changelogCmd)

	applyChangelogRenderFlags(changelogRenderCmd)

	changelogNewCmd.Flags().BoolVarP(&changelogEntryEdit, "edit", "e", false, "Start $EDITOR after creating new entry")
	changelogNewCmd.Flags().BoolVarP(&changelogEntryMinimalTemplate, "minimal-template", "m", false, "Use a minimal entry template")
	changelogNewCmd.Flags().BoolVarP(&changelogEntryInteractive, "interactive", "i", false, "Fill template values interactively")

	changelogReleaseCmd.Flags().StringVarP(&changelogReleaseVersionPattern, "version-pattern", "P", changelog.SemverVersionPattern.String(), "version number pattern")
	changelogReleasePathCmd.Flags().StringVarP(&changelogReleaseVersionPattern, "version-pattern", "P", changelog.SemverVersionPattern.String(), "version number pattern")
}

func applyChangelogRenderFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&changelogRenderFormat, "format", "f", changelog.FormatMD, "The render format. (e.g., \"md\", \"html\", or \"d360html\")")
	cmd.Flags().BoolVarP(&changelogDisableGitHubLinks, "no-links", "N", false, "Do not render issue or pull-request links for entries.")
	cmd.Flags().StringVarP(&changelogReleaseDate, "date", "d", time.Now().Format("2006-01-02"), "The release date.")
	cmd.Flags().StringVarP(&changelogReleaseVersion, "version", "V", "0.0.0", "The release version.")
	cmd.Flags().StringVarP(&changelogReleaseVersionPattern, "version-pattern", "P", changelog.SemverVersionPattern.String(), "version number pattern")
	cmd.Flags().StringVarP(&changelogProduct, "product", "p", "Graylog", "The product name. (e.g., \"Graylog\", \"Graylog Enterprise\")")
	cmd.Flags().BoolVar(&changelogSkipHeader, "skip-header", false, "Don't render the header")
	cmd.Flags().BoolVar(&changelogReadStdin, "stdin", false, "Read paths from STDIN")
}

func changelogRenderCommand(cmd *cobra.Command, args []string) {
	if len(args) == 0 && !changelogReadStdin {
		logger.Error("Missing snippet directories")
		if err := cmd.UsageFunc()(cmd); err != nil {
			logger.Fatal(err.Error())
		}
		os.Exit(1)
	}

	snippetsPaths := lo.Map(args, func(arg string, _ int) string {
		path, err := filepath.Abs(arg)
		if err != nil {
			logger.Fatal("couldn't get absolute path for %s", arg)
		}
		return path
	})

	if err := execChangelogRenderCommand(snippetsPaths); err != nil {
		logger.Error(err.Error())
		if err := cmd.UsageFunc()(cmd); err != nil {
			logger.Fatal(err.Error())
		}
		os.Exit(1)
	}
}

func execChangelogRenderCommand(snippetsPaths []string) error {
	if !lo.Contains(changelog.AvailableFormatters, changelogRenderFormat) {
		return fmt.Errorf("invalid render format: %s (available: %s)", changelogRenderFormat, strings.Join(changelog.AvailableFormatters, ", "))
	}

	if len(snippetsPaths) == 0 && !changelogReadStdin {
		return errors.New("missing snippet directories")
	}

	versionPattern, err := regexp.Compile(changelogReleaseVersionPattern)
	if err != nil {
		return fmt.Errorf("invalid version pattern: %s", changelogReleaseVersionPattern)
	}

	// By convention, we use the version in the first snippet path if it's a valid one and no version flag is given.
	releaseVersion := changelogReleaseVersion
	if releaseVersion == "0.0.0" && !changelogReadStdin {
		versionPath := filepath.Base(snippetsPaths[0])
		if versionPattern.MatchString(versionPath) {
			releaseVersion = versionPath
		} else {
			return errors.New("missing --version flag and snippets directory doesn't contain a valid version")
		}
	}

	if !versionPattern.MatchString(releaseVersion) {
		return fmt.Errorf("invalid version: %s", releaseVersion)
	}

	config := changelog.Config{
		RenderFormat:      changelogRenderFormat,
		RenderGitHubLinks: !changelogDisableGitHubLinks,
		SnippetsPaths:     snippetsPaths,
		ReleaseDate:       changelogReleaseDate,
		ReleaseVersion:    releaseVersion,
		Product:           changelogProduct,
		ReadStdin:         changelogReadStdin,
	}

	if err := changelog.Render(config); err != nil {
		return err
	}

	return nil
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

func changelogReleasePathCommand(cmd *cobra.Command, args []string) {
	v := args[0]
	semver, err := version.NewSemver(v)
	if err != nil {
		logger.Fatal("Invalid version: %s: %s", v, err)
	}

	if semver.Prerelease() != "" {
		logger.Fatal("Not allowing changelog rotation for pre-releases!")
	}

	versionPattern, err := regexp.Compile(changelogReleaseVersionPattern)
	if err != nil {
		logger.Fatal("Invalid version pattern: %s", changelogReleaseVersionPattern)
	}

	path, err := git.ToplevelPath()
	if err != nil {
		logger.Fatal("%s", err)
	}

	if err := changelog.ReleaseInPath(path, v, versionPattern); err != nil {
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
