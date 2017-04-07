package apply

import (
	"fmt"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/pom"
	"github.com/Graylog2/graylog-project-cli/pomparse"
	"github.com/Graylog2/graylog-project-cli/project"
	"github.com/fatih/color"
	"os"
	e "os/exec"
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
