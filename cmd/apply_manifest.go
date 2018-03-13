package cmd

import (
	"github.com/Graylog2/graylog-project-cli/apply"
	"github.com/Graylog2/graylog-project-cli/git"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	"github.com/Graylog2/graylog-project-cli/pom"
	"github.com/Graylog2/graylog-project-cli/pomparse"
	"github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var applyManifestCmd = &cobra.Command{
	Use:   "apply-manifest",
	Short: "Apply the given manifest",
	Long: `This can be used to release a new Graylog version.

It takes an apply-manfiest and creates a new Graylog release from that.

Example:

  # Shows all commands that would be executed
  $ graylog-project apply-manifest manifests/release-2.2.0.json

  # Actually execute all commands!
  $ graylog-project apply-manifest --execute manifests/release-2.2.0.json
`,
	Run: applyManifestCommand,
}

var applyManifestExecute bool
var applyManifestForce bool

func init() {
	RootCmd.AddCommand(applyManifestCmd)

	applyManifestCmd.Flags().BoolVarP(&applyManifestExecute, "execute", "", false, "Actually apply the manifest!")
	applyManifestCmd.Flags().BoolVarP(&applyManifestForce, "force", "f", false, "Ignore some sanity checks")

	viper.BindPFlag("apply-manifest.execute", applyManifestCmd.Flags().Lookup("execute"))
	viper.BindPFlag("apply-manifest.force", applyManifestCmd.Flags().Lookup("force"))
}

func applyManifestInDirectory(path string, callback utils.DirectoryCallback) {
	logger.ColorInfo(color.FgGreen, "----> in directory: %s", path)
	utils.InDirectory(path, callback)
}

func applyManifestCommand(cmd *cobra.Command, args []string) {
	logger.SetPrefix("[graylog-project]")

	t := time.Now()
	mavenProfiles := []string{"release"}
	config, repoManager, proj := prepareCheckoutCommand(cmd, args)
	var applier apply.Applier

	if applyManifestExecute {
		applier = apply.NewExecuteApplier(mavenProfiles)
	} else {
		applier = apply.NewNoopApplier(mavenProfiles)
	}

	msg := func(message string) {
		logger.ColorInfo(color.FgYellow, "===> %s", message)
	}

	msg("Sanity check for apply manifest")
	applyManifestErrors := 0
	apply.ForEachModule(proj, false, func(module project.Module) {
		if module.ApplyFromRevision() == "" {
			applyManifestErrors++
			logger.Error("Missing `apply.from_revision` field for module: %s", module.Name)
		}
		if module.ApplyNewVersion() == "" {
			applyManifestErrors++
			logger.Error("Missing `apply.new_version` field for module: %s", module.Name)
		}
	})

	if applyManifestErrors > 0 {
		if !applyManifestForce {
			os.Exit(1)
		}
	}

	repoManager.SetupProjectRepositoriesWithApply(proj, true)

	pom.WriteTemplates(config, proj)
	manifest.WriteState(config.Checkout.ManifestFiles)

	// Check that there are no modifications in the repositories via "git status --porcelain ."
	msg("Checking every module for uncommitted changes")
	apply.ForEachModule(proj, false, func(module project.Module) {
		applyManifestInDirectory(module.Path, func() {
			output := git.GitValue("status", "--porcelain")

			if output != "" {
				logger.Error("Module %s has uncommitted changes:", module.Name)
				for _, line := range strings.Split(output, "\n") {
					logger.Error("%s", line)
				}
				if !applyManifestForce {
					os.Exit(1)
				}
			}
		})
	})

	// Set release version in all web modules
	msg("Setting release version in all web modules")
	apply.ForEachModule(proj, true, func(module project.Module) {
		applyManifestInDirectory(module.Path, func() {
			applier.NpmVersionSet(module, module.Revision)
		})
	})

	// Set release version in all modules
	msg("Setting release version in all modules")
	apply.ForEachModule(proj, false, func(module project.Module) {
		applyManifestInDirectory(module.Path, func() {
			applier.MavenVersionsSet(module.Revision)
		})

		// Update all versions after each change!
		applyManifestUpdateVersions(msg, proj, applier)
	})

	// Regenerate the graylog-project pom and assembly files to get the latest versions
	msg("Regenerate pom and assembly templates")
	pom.WriteTemplates(config, proj)

	// Run tests via package to also test the jar creation
	msg("Running tests and build artifacts")
	logger.ColorInfo(color.FgMagenta, "[%s]", utils.GetCwd())
	applier.MavenRun("clean", "package")

	// Committing new version in web modules
	// Run this before the maven scm checkin is pushing to GitHub
	msg("Committing new version in web modules")
	apply.ForEachModule(proj, true, func(module project.Module) {
		applyManifestInDirectory(module.Path, func() {
			applier.NpmVersionCommit(module, module.Revision)
		})
	})

	// Commit and push new versions and create and push tag
	msg("Committing and pushing new versions and tags")
	apply.ForEachModule(proj, false, func(module project.Module) {
		applyManifestInDirectory(module.Path, func() {
			applier.MavenScmCheckinRelease(module.Name, module.Revision)
			applier.MavenScmTag(module.Revision)
		})
	})

	// Run deploy & build artifacts
	msg("Running deploy and build artifacts")
	logger.ColorInfo(color.FgMagenta, "[%s]", utils.GetCwd())
	applier.MavenRunWithProfiles([]string{"release"}, "-DskipTests", "clean", "deploy")

	// Set development version in all web modules
	msg("Setting development version in all web modules")
	apply.ForEachModule(proj, true, func(module project.Module) {
		applyManifestInDirectory(module.Path, func() {
			applier.NpmVersionSet(module, module.ApplyNewVersion())
		})
	})

	// Set development version
	msg("Set development versions")
	apply.ForEachModule(proj, false, func(module project.Module) {
		applyManifestInDirectory(module.Path, func() {
			applier.MavenVersionsSet(module.ApplyNewVersion())
		})

		// Update all versions after each change!
		applyManifestUpdateVersions(msg, proj, applier)
	})

	// Committing new development version in web modules
	// Run this before the maven scm checkin is pushing to GitHub
	msg("Committing new development version in web modules")
	apply.ForEachModule(proj, true, func(module project.Module) {
		applyManifestInDirectory(module.Path, func() {
			applier.NpmVersionCommit(module, module.ApplyNewVersion())
		})
	})

	// Commit development version
	msg("Commit development versions")
	apply.ForEachModule(proj, false, func(module project.Module) {
		applyManifestInDirectory(module.Path, func() {
			applier.MavenScmCheckinDevelopment(module.Name)
		})
	})

	// Create new branches
	msg("Create new branches")
	apply.ForEachModule(proj, false, func(module project.Module) {
		applyManifestInDirectory(module.Path, func() {
			if module.ApplyNewBranch() != "" {
				applier.MavenScmBranch(module.ApplyNewBranch())
			}
		})
	})

	logger.Info("DONE! - took: %s", time.Since(t))
}

func applyManifestUpdateVersions(msg func(string), proj project.Project, applier apply.Applier) {
	// Set parent versions and graylog.version properties in non-server modules
	msg("Setting parent and graylog.version properties in non-server modules")
	serverVersion := proj.Server.Version()
	apply.ForEachModule(proj, true, func(module project.Module) {
		if module.Server {
			// Don't change parent and graylog.version property for server
			return
		}
		applyManifestInDirectory(module.Path, func() {
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
			applyManifestInDirectory(module.Path, func() {
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
