package apply

import (
	"github.com/Graylog2/graylog-project-cli/project"
	"strings"
)

type CommonMaven struct {
	Applier  Applier
	Profiles []string
}

func (common CommonMaven) MavenRunWithProfiles(profiles []string, args ...string) {
	commands := []string{"mvn --show-version --batch-mode --fail-fast"}

	if len(profiles) > 0 {
		commands = append(commands, "--activate-profiles")
		commands = append(commands, strings.Join(profiles, ","))
	}

	commands = append(commands, strings.Join(args, " "))

	common.Applier.MavenExec(commands)
}

func (common CommonMaven) MavenRun(args ...string) {
	common.MavenRunWithProfiles([]string{}, args...)
}

func (common CommonMaven) MavenVersionsSet(newVersion string) {
	common.MavenRunWithProfiles(common.Profiles, "-DnewVersion="+newVersion, "versions:set", "versions:commit")
}

func (common CommonMaven) MavenScmCheckinRelease(moduleName string, moduleVersion string) {
	common.MavenRunWithProfiles(common.Profiles, "-Dmessage=\"["+moduleName+"] prepare release "+moduleVersion+"\"", "-Dincludes=\"**/pom.xml\"", "-Dexcludes=\"**/target/**/pom.xml\"", "scm:checkin")
}

func (common CommonMaven) MavenScmCheckinDevelopment(moduleName string) {
	common.MavenRunWithProfiles(common.Profiles, "-Dmessage=\"["+moduleName+"] prepare for next development iteration\"", "-Dincludes=\"**/pom.xml\"", "-Dexcludes=\"**/target/**/pom.xml\"", "scm:checkin")
}

func (common CommonMaven) MavenScmTag(moduleTag string) {
	common.MavenRunWithProfiles(common.Profiles, "-Dtag="+moduleTag, "scm:tag")
}

func (common CommonMaven) MavenScmBranch(moduleBranch string) {
	common.MavenRunWithProfiles(common.Profiles, "-Dbranch="+moduleBranch, "scm:branch")
}

func (common CommonMaven) MavenDependencyVersionSet(module project.Module, groupId string, artifactId string, newVersion string) {
	common.MavenRunWithProfiles(common.Profiles, "versions:use-dep-version", "-DdepVersion="+newVersion, "-DforceVersion", "-Dincludes="+groupId+":"+artifactId)
}
