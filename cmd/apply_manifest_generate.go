package cmd

import (
	"github.com/Graylog2/graylog-project-cli/ask"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	"github.com/fatih/color"
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

var applyManifestGenerateCmd = &cobra.Command{
	Use:     "apply-manifest-generate",
	Aliases: []string{"amg"},
	Short:   "Generate apply-manifest from the given manifest",
	Long: `This generates a new apply-manifest which can be used to create a new release.

In non-batch mode, the command will collect all required values interactively and then creates a new release manifest.

  $ graylog-project apply-manifest-generate manifests/3.3.json Please enter the new version: 3.3.1
  From which revision (branch/tag) should this release be created? [3.3]
  Should we create a new branch? [y/N]
  What is the new development version? 3.3.2-SNAPSHOT
  Do you want to use the default settings for all modules? [Y/n]

In batch mode, all required values can be passed as command flags.

  graylog-project apply-manifest-generate \
      --batch \
      --release-version 3.3.0 \
      --dev-version 3.3.1-SNAPSHOT \
      --base-rev master \
      [--new-branch 3.3 \]
      manifests/3.3.json

`,
	Run: applyManifestGenerateCommand,
}

const versionRegex = version.VersionRegexpRaw

var amgBatchMode bool
var amgReleaseVersion string
var amgDevVersion string
var amgNewBranch string
var amgBaseRev string

func init() {
	RootCmd.AddCommand(applyManifestGenerateCmd)

	applyManifestGenerateCmd.Flags().BoolVar(&amgBatchMode, "batch", false, "Enable batch mode for automation")
	applyManifestGenerateCmd.Flags().StringVar(&amgReleaseVersion, "release-version", "", "Release version")
	applyManifestGenerateCmd.Flags().StringVar(&amgDevVersion, "dev-version", "", "Next development version")
	applyManifestGenerateCmd.Flags().StringVar(&amgNewBranch, "new-branch", "", "Create new branch (optional)")
	applyManifestGenerateCmd.Flags().StringVar(&amgBaseRev, "base-rev", "", "Base revision (branch/tag) for the release")
}

func applyManifestGenerateCommand(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		logger.Info("Missing manifest argument")
		_ = cmd.UsageFunc()(cmd)
		os.Exit(1)
	}

	newManifest := manifest.ReadManifest(args[0:])

	var releaseVersion string

	if amgBatchMode {
		releaseVersion = amgCommandBatch(&newManifest)
	} else {
		releaseVersion = amgCommandInteractive(&newManifest)
	}

	buf, err := manifest.Marshal(newManifest)
	if err != nil {
		logger.Fatal("ERROR: %v", err)
	}

	newManifestFile := "manifests/release-" + releaseVersion + ".json"

	if err := ioutil.WriteFile(newManifestFile, buf, 0644); err != nil {
		logger.Fatal("Unable to write new manifest file %s: %v", newManifestFile, err)
	}

	logger.ColorInfo(color.FgGreen, "Wrote new apply-manifest to: %s", newManifestFile)
	logger.ColorInfo(color.FgYellow, "Make sure to review the new manifest to make sure it's correct!")
}

// Create manifest parameters based on command flags.
func amgCommandBatch(newManifest *manifest.Manifest) string {
	if amgReleaseVersion == "" {
		logger.Fatal("Missing --release-version parameter")
	}
	if amgDevVersion == "" {
		logger.Fatal("Missing --dev-version parameter")
	}
	if amgBaseRev == "" {
		logger.Fatal("Missing --base-rev parameter")
	}

	var newModules []manifest.ManifestModule

	for _, module := range newManifest.Modules {
		module.Revision = amgReleaseVersion
		newModules = append(newModules, module)
	}

	newManifest.DefaultApply = manifest.ManifestApply{
		FromRevision: amgBaseRev,
		NewVersion:   amgDevVersion,
	}
	if amgNewBranch != "" {
		newManifest.DefaultApply.NewBranch = amgNewBranch
	}

	newManifest.Modules = newModules

	return amgReleaseVersion
}

func applyManifestGenerateAskModule(asker *ask.Asker, module manifest.ManifestModule) (string, manifest.ManifestApply) {
	newApply := manifest.ManifestApply{}

	newVersion := asker.Ask("Please enter the new version:", "", versionRegex)

	newApply.FromRevision = asker.Ask("From which revision (branch/tag) should this release be created?", module.Revision, `^\S+$`)
	if asker.AskYesNo("Should we create a new branch?", false) {
		newApply.NewBranch = asker.Ask("New branch name:", "", `^\S+$`)
	}
	newApply.NewVersion = asker.Ask("What is the new development version?", "", versionRegex)

	return newVersion, newApply
}

// Asks the user for required input parameters interactively.
func amgCommandInteractive(newManifest *manifest.Manifest) string {
	asker := ask.NewAsker(os.Stdin)

	var serverModule manifest.ManifestModule
	for _, m := range newManifest.Modules {
		if !m.Server {
			continue
		}
		serverModule = m
	}

	newVersion, newApply := applyManifestGenerateAskModule(&asker, serverModule)

	newManifest.DefaultApply = newApply

	var newModules []manifest.ManifestModule

	for _, module := range newManifest.Modules {
		module.Revision = newVersion

		newModules = append(newModules, module)
	}

	newManifest.Modules = newModules

	return newVersion
}
