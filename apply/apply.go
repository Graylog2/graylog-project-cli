package apply

import (
	"github.com/Graylog2/graylog-project-cli/project"
)

type applierCommon interface {
	MavenRun(args ...string)

	MavenRunWithProfiles(profiles []string, args ...string)

	MavenVersionsSet(newVersion string)

	MavenScmCheckinRelease(moduleName string, moduleVersion string)

	MavenScmCheckinDevelopment(moduleName string)

	MavenScmTag(moduleTag string)

	MavenScmBranch(moduleBranch string)

	MavenDependencyVersionSet(module project.Module, groupId string, artifactId string, newVersion string)

	ChangelogRelease(path string, revision string) error
}

type Applier interface {
	applierCommon

	MavenExec(commands []string)

	MavenSetParent(module project.Module, parentVersion string)

	MavenSetProperty(module project.Module, name string, value string)

	NpmVersionSet(module project.Module, newVersion string)

	NpmVersionCommit(module project.Module, newVersion string)
}

// Ensures that the server module gets handled first.
func ForEachModule(p project.Project, includeSubmodules bool, callback func(project.Module)) {
	for _, module := range p.Modules {
		if module.Server {
			callback(module)
		}
	}
	for _, module := range p.Modules {
		if !module.Server {
			callback(module)
			if includeSubmodules && module.HasSubmodules() {
				for _, submodule := range module.Submodules {
					callback(submodule)
				}
			}
		}
	}
}
