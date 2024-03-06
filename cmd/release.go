package cmd

import (
	"github.com/Graylog2/graylog-project-cli/apply"
	"github.com/Graylog2/graylog-project-cli/git"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	"github.com/Graylog2/graylog-project-cli/pomparse"
	"github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/projectstate"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/fatih/color"
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Build a release for the given manifest",
	Long: `This can be used to release a new Graylog version.

It takes an apply-manifest and creates a new Graylog release from that.

Example:

  # Shows all commands that would be executed
  $ graylog-project release manifests/release-2.2.0.json
`,
	Run: releaseCommand,
}

var releasePrepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare a release for the given manifest",
	Long: `This can be used to prepare a new Graylog release.

It takes an apply-manifest and prepares a new Graylog release from that.

Example:

  # Shows all commands that would be executed
  $ graylog-project release prepare manifests/release-2.2.0.json
`,
	RunE: releasePrepareCommand,
}

var releaseForce bool
var releaseSkipMavenDeploy bool
var releaseSkipTests bool

func init() {
	releaseCmd.AddCommand(releasePrepareCmd)
	RootCmd.AddCommand(releaseCmd)

	releaseCmd.Flags().BoolVarP(&releaseForce, "force", "f", false, "Ignore some sanity checks")
	releaseCmd.Flags().BoolVarP(&releaseSkipMavenDeploy, "skip-maven-deploy", "", false, "Skip maven deployment")
	releaseCmd.Flags().BoolVarP(&releaseSkipTests, "skip-tests", "", false, "Skip running tests via maven")

	viper.BindPFlag("release.force", releaseCmd.Flags().Lookup("force"))
	viper.BindPFlag("release.skip-deploy", releaseCmd.Flags().Lookup("skip-maven-deploy"))
	viper.BindPFlag("release.skip-tests", releaseCmd.Flags().Lookup("skip-tests"))
}

func inDirectory(path string, callback utils.DirectoryCallback) {
	logger.ColorInfo(color.FgGreen, "----> in directory: %s", path)
	utils.InDirectory(path, callback)
}

func releasePrepareCommand(cmd *cobra.Command, args []string) error {
	logger.SetPrefix("[graylog-project]")

	mavenProfiles := []string{"release"}
	config, repoManager, proj := prepareCheckoutCommand(cmd, args)
	applier := apply.NewExecuteApplier(mavenProfiles)

	msg := func(message string) {
		logger.ColorInfo(color.FgYellow, "===> %s", message)
	}

	msg("Sanity check for apply manifest")
	validationErrors := 0
	apply.ForEachModule(proj, false, func(module project.Module) {
		if module.ApplyFromRevision() == "" {
			validationErrors++
			logger.Error("Missing `apply.from_revision` field for module: %s", module.Name)
		}
		if module.ApplyNewVersion() == "" {
			validationErrors++
			logger.Error("Missing `apply.new_version` field for module: %s", module.Name)
		}
	})

	if validationErrors > 0 {
		if !releaseForce {
			os.Exit(1)
		}
	}

	repoManager.SetupProjectRepositoriesWithApply(proj, true)

	projectstate.Sync(proj, config)
	manifest.WriteState(config.Checkout.ManifestFiles)

	// Check that there are no modifications in the repositories via "git status --porcelain ."
	msg("Checking every module for uncommitted changes")
	apply.ForEachModule(proj, false, func(module project.Module) {
		inDirectory(module.Path, func() {
			output := git.GitValue("status", "--porcelain")

			if output != "" {
				logger.Error("Module %s has uncommitted changes:", module.Name)
				for _, line := range strings.Split(output, "\n") {
					logger.Error("%s", line)
				}
				if !releaseForce {
					os.Exit(1)
				}
			}
		})
	})

	// Set release version in all web modules
	msg("Setting release version in all web modules")
	apply.ForEachModule(proj, true, func(module project.Module) {
		inDirectory(module.Path, func() {
			applier.NpmVersionSet(module, module.Revision)
		})
	})

	// Set release version in all modules
	msg("Setting release version in all modules")
	apply.ForEachModule(proj, false, func(module project.Module) {
		inDirectory(module.Path, func() {
			applier.MavenVersionsSet(module.Revision)
		})

		// Update all versions after each change!
		applyManifestUpdateVersions(msg, proj, applier)
	})

	// Regenerate the graylog-project pom and assembly files to get the latest versions
	msg("Regenerate pom and assembly templates")
	projectstate.Sync(proj, config)

	// Run tests via package to also test the jar creation
	msg("Running tests and build artifacts")
	logger.ColorInfo(color.FgMagenta, "[%s]", utils.GetCwd())
	if releaseSkipTests {
		msg("Skipping tests!")
		applier.MavenRun("-DskipTests", "clean", "package")
	} else {
		applier.MavenRun("clean", "package")
	}

	msg("Rotate changelogs for release")
	apply.ForEachModule(proj, false, func(module project.Module) {
		inDirectory(module.Path, func() {
			v, err := version.NewSemver(module.Revision)
			if err != nil {
				logger.Fatal("Couldn't create new semver for %s: %s", module.Revision, err)
			}
			// We don't want different changelog folders for each pre-release but one folder for each GA release.
			if v.Prerelease() == "" {
				if err := applier.ChangelogRelease(module.Path, module.Revision); err != nil {
					logger.Fatal("ERROR: %s", err)
				}
			} else {
				logger.Info("Skipping changelog release for pre-release version: %s", v)
			}
		})
	})

	// Committing new version in web modules
	// Run this before the maven scm checkin is pushing to GitHub
	msg("Committing new version in web modules")
	apply.ForEachModule(proj, true, func(module project.Module) {
		inDirectory(module.Path, func() {
			applier.NpmVersionCommit(module, module.Revision)
		})
	})

	// Commit and push new versions and create and push tag
	msg("Committing and pushing new versions and tags")
	apply.ForEachModule(proj, false, func(module project.Module) {
		inDirectory(module.Path, func() {
			applier.MavenScmCheckinRelease(module.Name, module.Revision)
			applier.MavenScmTag(module.Revision)
		})
	})

	return nil
}

func releaseCommand(cmd *cobra.Command, args []string) {
	logger.SetPrefix("[graylog-project]")

	t := time.Now()
	mavenProfiles := []string{"release"}
	_, _, proj := prepareCheckoutCommand(cmd, args)
	applier := apply.NewExecuteApplier(mavenProfiles)

	msg := func(message string) {
		logger.ColorInfo(color.FgYellow, "===> %s", message)
	}
	// Run deploy & build artifacts
	msg("Running deploy and build artifacts")
	logger.ColorInfo(color.FgMagenta, "[%s]", utils.GetCwd())
	if releaseSkipMavenDeploy {
		msg("Skipping maven deployment!")
		applier.MavenRunWithProfiles([]string{"release"}, "-DskipTests", "clean", "package")
	} else {
		applier.MavenRunWithProfiles([]string{"release"}, "-DskipTests", "clean", "deploy")
	}

	// Set development version in all web modules
	msg("Setting development version in all web modules")
	apply.ForEachModule(proj, true, func(module project.Module) {
		inDirectory(module.Path, func() {
			applier.NpmVersionSet(module, module.ApplyNewVersion())
		})
	})

	// Set development version
	msg("Set development versions")
	apply.ForEachModule(proj, false, func(module project.Module) {
		inDirectory(module.Path, func() {
			applier.MavenVersionsSet(module.ApplyNewVersion())
		})

		// Update all versions after each change!
		applyManifestUpdateVersions(msg, proj, applier)
	})

	// Committing new development version in web modules
	// Run this before the maven scm checkin is pushing to GitHub
	msg("Committing new development version in web modules")
	apply.ForEachModule(proj, true, func(module project.Module) {
		inDirectory(module.Path, func() {
			applier.NpmVersionCommit(module, module.ApplyNewVersion())
		})
	})

	// Commit development version
	msg("Commit development versions")
	apply.ForEachModule(proj, false, func(module project.Module) {
		inDirectory(module.Path, func() {
			applier.MavenScmCheckinDevelopment(module.Name)
		})
	})

	// Create new branches
	msg("Create new branches")
	apply.ForEachModule(proj, false, func(module project.Module) {
		inDirectory(module.Path, func() {
			if module.ApplyNewBranch() != "" {
				applier.MavenScmBranch(module.ApplyNewBranch())
			}
		})
	})

	msg("Rotate changelogs in source branch when creating new branch")
	apply.ForEachModule(proj, false, func(module project.Module) {
		inDirectory(module.Path, func() {
			// If we create a new branch during the release process, we want to move the "unreleased" changelogs in
			// the source branch to a versioned folder.
			//
			// Example: Release 5.1.0-rc.1 and create a "5.1" branch from the "main" branch in that process.
			// 1. Create "5.1" branch, keeping the "changelog/unreleased" folder for the branch
			// 2. In the "main" branch move "changelog/unreleased" to "changelog/5.1.0-rc.1"
			// Result:
			// branch "main": changelog/5.1.0-rc.1 (and a new empty changelog/unreleased folder)
			// branch "5.1":  changelog/unreleased
			//
			// The drawback here is that the "main" branch only contains the 5.1 changelogs up until the 5.1.0-rc.1
			// release. All newer 5.1 changelogs will only be in the "5.1" branch. There doesn't seem to be a better
			// way without merging changelogs between the branches, so we live with that drawback for now since
			// it doesn't affect changelog generation. For the 5.1.0 GA release the final changelog is generated
			// from the "5.1" branch which has all the changelogs. (changelogs in the "unreleased" folder of the "main"
			// branch moved on to the next feature release already, e.g., 5.2)
			if module.ApplyNewBranch() != "" {
				// We might have done the renaming in the first ChangelogRelease call above when the version is not
				// a pre-release AND we are creating a new branch. In that case this call is a no-op because the
				// changelog rotation code checks if the version changelog folder already exists.
				if err := applier.ChangelogRelease(module.Path, module.Revision); err != nil {
					logger.Fatal("ERROR: %s", err)
				}

				if output, err := git.GitE("push", "origin", module.ApplyFromRevision()); err != nil {
					logger.Fatal("ERROR: %s\n%s", err, output)
				} else {
					logger.Info("%s", output)
				}
			} else {
				logger.Info("Skipping changelog rotation for module: %s (no branch creation requested)", module.Path)
			}
		})
	})

	logger.Info("DONE! - took: %s", time.Since(t))
}

func _applyManifestUpdateVersions(msg func(string), proj project.Project, applier apply.Applier) {
	// Set parent versions and graylog.version properties in non-server modules
	msg("Setting parent and graylog.version properties in non-server modules")
	serverVersion := proj.Server.Version()
	apply.ForEachModule(proj, true, func(module project.Module) {
		if module.Server {
			// Don't change parent and graylog.version property for server
			return
		}
		inDirectory(module.Path, func() {
			applier.MavenSetParent(module, serverVersion)
			applier.MavenSetProperty(module, "graylog.version", serverVersion)
			applier.MavenSetProperty(module, "graylog2.version", serverVersion)
		})
	})

	// Check if any module uses another module as dependency and update the dependency version to the new one.
	msg("Checking if any module uses another module as dependency and update dependency versions")
	checkDep := func(module project.Module, dep pomparse.MavenDependency) {
		// Skip entries with an empty version, the version is probably defined somewhere else. (<dependencyManagement/>)
		if dep.Version == "" {
			return
		}
		// Skip versions which use a property for the version.
		if strings.HasPrefix(dep.Version, "${") {
			return
		}

		match, matchedModule := project.HasModule(proj, dep.GroupId, dep.ArtifactId)
		if match && dep.Version != matchedModule.Version() {
			inDirectory(module.Path, func() {
				applier.MavenDependencyVersionSet(module, dep.GroupId, dep.ArtifactId, matchedModule.Version())
			})
		}
	}
	apply.ForEachModule(proj, true, func(module project.Module) {
		pomFile := filepath.Join(module.Path, "pom.xml")
		pom := pomparse.ParsePom(pomFile)

		for _, dep := range pom.Dependencies {
			checkDep(module, dep)
		}
		for _, dep := range pom.DependencyManagement {
			checkDep(module, dep)
		}
	})
}
