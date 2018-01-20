package apply

import (
	"fmt"
	"github.com/Graylog2/graylog-project-cli/git"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/pom"
	"github.com/Graylog2/graylog-project-cli/pomparse"
	"github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/fatih/color"
	"os"
	e "os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Used with pom.SetParentIfMatches() to decide if the parent should be updated
var parentMatchFunc = func(module project.Module, pom pomparse.MavenPom) bool {
	// The parent should only be updated if it is a graylog plugin parent
	return pom.ParentGroupId == "org.graylog.plugins" && (pom.ParentArtifactId == "graylog-plugin-parent" || pom.ParentArtifactId == "graylog-plugin-web-parent")
}

func NewExecuteApplier(profiles []string) Applier {
	applier := executeApplier{}
	applier.CommonMaven = CommonMaven{Profiles: profiles, Applier: applier}

	return applier
}

// An apply.Applier implementation that actually executes the commands.
type executeApplier struct {
	CommonMaven
}

func (execute executeApplier) MavenSetParent(module project.Module, parentVersion string) {
	if module.HasParent() {
		fmt.Println("set parent version: " + parentVersion)
		pom.SetParentIfMatches(module, module.ParentGroupId(), module.ParentArtifactId(), parentVersion, module.ParentRelativePath(), parentMatchFunc)
	}
}

func (execute executeApplier) MavenSetProperty(module project.Module, name string, value string) {
	fmt.Println("set property: <" + name + ">" + value + "</" + name + ">")
	pom.SetProperty(module, name, value)
}

func (execute executeApplier) MavenExec(commands []string) {
	logger.ColorPrintln(color.FgMagenta, "[command output: %v]", strings.Join(commands, " "))

	command := e.Command("sh", "-c", strings.Join(commands, " "))

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		logger.Fatal("Command failed: %v", err)
	}
}

func (execute executeApplier) NpmVersionSet(module project.Module, version string) {
	versionRe := regexp.MustCompile(`^\d+\.\d+\.\d+-?.*?$`)
	filename := "package.json"

	if !versionRe.MatchString(version) {
		logger.Fatal("Invalid version: %s", version)
	}

	if module.Server {
		// The server module needs special treatment
		// TODO: Move list of package.json files to the manifest and default to "package.json"
		files := []string{
			filepath.Join("graylog2-web-interface", filename),
			filepath.Join("graylog2-web-interface/manifests", filename),
			filepath.Join("graylog2-web-interface/packages/graylog-web-plugin", filename),
		}
		for _, file := range files {
			fmt.Println("set version in " + filepath.Join(module.Path, file) + ": " + version)

			err := utils.SetPackageJsonVersion(file, version)
			if err != nil {
				logger.Fatal("Couldn't set version in file %s: %s", filepath.Join(module.Path, file), err)
			}
		}
		return
	}

	if !module.IsNpmModule() {
		return
	}

	absFilename := filepath.Join(module.Path, filename)

	fmt.Println("set version in " + absFilename + ": " + version)

	err := utils.SetPackageJsonVersion(filename, version)
	if err != nil {
		logger.Fatal("Couldn't set version in file %s: %s", absFilename, err)
	}
}

func (execute executeApplier) NpmVersionCommit(module project.Module, version string) {
	commitMsg := fmt.Sprintf("Bump package.json version to %s", version)

	utils.InDirectory(module.Path, func() {
		if module.Server {
			// The server module needs special treatment
			// TODO: Move list of package.json files to the manifest and default to "package.json"
			file1 := "graylog2-web-interface/package.json"
			file2 := "graylog2-web-interface/manifests/package.json"
			file3 := "graylog2-web-interface/packages/graylog-web-plugin/package.json"

			git.Git("commit", "-m", commitMsg, file1, file2, file3)
			return
		}

		if !module.IsNpmModule() {
			return
		}

		git.Git("commit", "-m", commitMsg, "package.json")
	})
}
