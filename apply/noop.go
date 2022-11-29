package apply

import (
	"fmt"
	"github.com/Graylog2/graylog-project-cli/project"
	"strings"
)

func NewNoopApplier(profiles []string) Applier {
	applier := noopApplier{}
	applier.CommonMaven = CommonMaven{Profiles: profiles, Applier: applier}

	return applier
}

// Check that noopApplier implements the Applier interface
var _ Applier = (*noopApplier)(nil)

// A no-op implementation of the apply.Applier interface which just prints the commands.
type noopApplier struct {
	CommonMaven
}

func (noop noopApplier) MavenSetParent(module project.Module, parentVersion string) {
	fmt.Println("set parent version: " + parentVersion)
}

func (noop noopApplier) MavenSetProperty(module project.Module, name string, value string) {
	fmt.Println("set property: <" + name + ">" + value + "</" + name + ">")
}

func (noop noopApplier) MavenExec(commands []string) {
	fmt.Println(strings.Join(commands, " "))
}

func (noop noopApplier) NpmVersionSet(module project.Module, newVersion string) {
	fmt.Println("set web module version: " + newVersion)
}

func (noop noopApplier) NpmVersionCommit(module project.Module, newVersion string) {
	fmt.Println("commit web module version: " + newVersion)
}

func (noop noopApplier) ChangelogRelease(path string, revision string) error {
	fmt.Println("rotating changelog: ", path, revision)
	return nil
}
