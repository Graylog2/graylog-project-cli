package cmd

import (
	"strconv"

	c "github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/git"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"strings"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Overview of the current project state",
	Long: `
Displays the current project state like the used manifest file, module versions and git status.
`,
	Run: statusCommand,
}

func init() {
	RootCmd.AddCommand(statusCmd)
}

func statusCommand(cmd *cobra.Command, args []string) {
	config := c.Get()
	manifestFiles := manifest.ReadState().Files()
	project := p.New(config, manifestFiles)

	logger.Info("Current project status")
	logger.Info("  Manifests: %v", manifestFiles)

	logger.Info("  Module versions")

	maxNameLength := p.MaxModuleNameLength(project)

	for _, module := range project.Modules {
		if !utils.FileExists(module.Path) {
			logger.Info("Skipping module %v because it does not exist yet", module.Name)
			continue
		}
		utils.InDirectory(module.Path, func() {
			var commitId string
			if c.IsCI() {
				// Don't display shortened commit IDs in CI environments to avoid future ambiguities regarding the
				// shortened commit IDs
				commitId = git.GitValue("rev-parse", "HEAD")
			} else  {
				commitId = git.GitValue("rev-parse", "--short", "HEAD")
			}
			revision := git.GitValue("rev-parse", "--abbrev-ref", "HEAD")
			gitStatus := git.GitValue("status", "--porcelain")
			filesModified, filesDeleted, filesAdded := 0, 0, 0

			for _, l := range strings.Split(gitStatus, "\n") {
				line := strings.TrimSpace(l)
				if strings.HasPrefix(line, "M") {
					filesModified++
				}
				if strings.HasPrefix(line, "D") {
					filesDeleted++
				}
				if strings.HasPrefix(line, "A") {
					filesAdded++
				}
			}

			if !module.HasParent() {
				logger.Info("    %-"+strconv.Itoa(int(maxNameLength))+"s  %s (branch: %s, commit: %s)", module.Name, module.Version(), revision, commitId)
				if config.Verbose {
					logger.ColorInfo(color.FgYellow, "        <no parent>")
				}
			} else {
				if config.Verbose {
					logger.Info("    %-"+strconv.Itoa(int(maxNameLength))+"s  %s (branch: %s, commit: %s)", module.Name, module.Version(), revision, commitId)
					logger.ColorInfo(color.FgYellow, "        Parent groupId:      %s", module.ParentGroupId())
					logger.ColorInfo(color.FgYellow, "        Parent artifactId:   %s", module.ParentArtifactId())
					logger.ColorInfo(color.FgYellow, "        Parent version:      %s", module.ParentVersion())
					logger.ColorInfo(color.FgYellow, "        Parent relativePath: %s", module.ParentRelativePath())
				} else {
					logger.Info("    %-"+strconv.Itoa(int(maxNameLength))+"s  %s (branch: %s, commit: %s, parent: %s)", module.Name, module.Version(), revision, commitId, module.ParentVersion())
				}
			}
			if !config.Verbose && (filesModified > 0 || filesDeleted > 0 || filesAdded > 0) {
				logger.Printf("        Git status:")
				if filesAdded > 0 {
					logger.ColorPrintf(color.FgGreen, " %d added", filesAdded)
				}
				if filesDeleted > 0 {

					logger.ColorPrintf(color.FgRed, " %d deleted", filesDeleted)
				}
				if filesModified > 0 {
					logger.ColorPrintf(color.FgYellow, " %d modified", filesModified)
				}
				logger.ColorPrintf(color.FgYellow, "\n")
			}
		})
	}

	if !config.Verbose {
		return
	}

	logger.Info("  Git status")
	for _, module := range project.Modules {
		if !utils.FileExists(module.Path) {
			logger.Info("Skipping module %v because it does not exist yet", module.Name)
			continue
		}
		logger.Info("    %v", module.RelativePath())
		utils.InDirectory(module.Path, func() {
			git.SilentGit("status", "-s")
		})
	}
}
