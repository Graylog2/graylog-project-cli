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

An interactive process will collect all needed values and then creates a new release manifest.`,
	Run: applyManifestGenerateCommand,
}

const versionRegex = version.VersionRegexpRaw

func init() {
	RootCmd.AddCommand(applyManifestGenerateCmd)
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

func applyManifestGenerateCommand(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		logger.Info("Missing manifest argument")
		cmd.UsageFunc()(cmd)
		os.Exit(1)
	}

	newManifest := manifest.ReadManifest(args[0:])
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

	useDefaults := asker.AskYesNo("Do you want to use the default settings for all modules?", true)

	var newModules []manifest.ManifestModule

	for _, module := range newManifest.Modules {
		var newModuleVersion string
		var newModuleApply manifest.ManifestApply

		if useDefaults || asker.AskYesNo("Do you want to use the defaults for module "+module.Repository+"?", true) {
			newModuleVersion = newVersion
		} else {
			newModuleVersion, newModuleApply = applyManifestGenerateAskModule(&asker, module)
		}

		module.Revision = newModuleVersion
		module.Apply = newModuleApply

		newModules = append(newModules, module)
	}

	newManifest.Modules = newModules

	buf, err := manifest.Marshal(newManifest)
	if err != nil {
		logger.Fatal("ERROR: %v", err)
	}

	newManifestFile := "manifests/release-" + newVersion + ".json"

	if err := ioutil.WriteFile(newManifestFile, buf, 0644); err != nil {
		logger.Fatal("Unable to write new manifest file %s: %v", newManifestFile, err)
	}

	logger.ColorInfo(color.FgGreen, "Wrote new apply-manifest to: %s", newManifestFile)
	logger.ColorInfo(color.FgYellow, "Make sure to review the new manifest to make sure it's correct!")
}
