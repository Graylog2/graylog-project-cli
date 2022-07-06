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
			} else {
				commitId = git.GitValue("rev-parse", "--short", "HEAD")
			}
			revision := git.GitValue("rev-parse", "--abbrev-ref", "HEAD")
			gitStatus := git.GitValue("status", "--porcelain")
			filesModified, filesDeleted, filesAdded := 0, 0, 0
			filesModifiedStaged, filesDeletedStaged, filesAddedStaged := 0, 0, 0

			for _, line := range strings.Split(gitStatus, "\n") {
				if strings.HasPrefix(line, "M") {
					filesModifiedStaged++
				}
				if strings.HasPrefix(line, " M") {
					filesModified++
				}
				if strings.HasPrefix(line, "D") {
					filesDeletedStaged++
				}
				if strings.HasPrefix(line, " D") {
					filesDeleted++
				}
				if strings.HasPrefix(line, "A") {
					filesAddedStaged++
				}
				if strings.HasPrefix(line, " A") {
					filesAdded++
				}
			}

			if !module.HasParent() {
				logger.Info("    %-"+strconv.Itoa(int(maxNameLength))+"s  %s (branch: %s, commit: %s)", module.Name, module.Version(), revision, commitId)
				if config.Verbose > 1 {
					logger.ColorInfo(color.FgYellow, "        <no parent>")
				}
			} else {
				if config.Verbose > 1 {
					logger.Info("    %-"+strconv.Itoa(int(maxNameLength))+"s  %s (branch: %s, commit: %s)", module.Name, module.Version(), revision, commitId)
					logger.ColorInfo(color.FgYellow, "        Parent groupId:      %s", module.ParentGroupId())
					logger.ColorInfo(color.FgYellow, "        Parent artifactId:   %s", module.ParentArtifactId())
					logger.ColorInfo(color.FgYellow, "        Parent version:      %s", module.ParentVersion())
					logger.ColorInfo(color.FgYellow, "        Parent relativePath: %s", module.ParentRelativePath())
				} else {
					logger.Info("    %-"+strconv.Itoa(int(maxNameLength))+"s  %s (branch: %s, commit: %s, parent: %s)", module.Name, module.Version(), revision, commitId, module.ParentVersion())
				}
			}
			if config.Verbose == 0 &&
				((filesModified > 0 || filesDeleted > 0 || filesAdded > 0) ||
					(filesModifiedStaged > 0 || filesDeletedStaged > 0 || filesAddedStaged > 0)) {
				logger.Printf("        Git status:")
				if filesAdded > 0 || filesAddedStaged > 0 {
					logger.ColorPrintf(color.FgGreen, " %d added", filesAdded+filesAddedStaged)
				}
				if filesDeleted > 0 || filesDeletedStaged > 0 {
					logger.ColorPrintf(color.FgRed, " %d deleted", filesDeleted+filesDeletedStaged)
				}
				if filesModified > 0 || filesModifiedStaged > 0 {
					logger.ColorPrintf(color.FgYellow, " %d modified", filesModified+filesModifiedStaged)
				}
				logger.ColorPrintf(color.FgYellow, "\n")
			}
			if config.Verbose > 0 && config.Verbose < 2 {
				statusValue := git.GitValue("status", "-s")
				statusLines := strings.Split(strings.TrimSuffix(statusValue, "\n"), "\n")

				if len(statusLines) > 0 && statusLines[0] != "" {
					logger.Printf("        Git status:\n")
					for _, line := range statusLines {
						var lineColor color.Attribute
						switch {
						case strings.HasPrefix(line, "M") || strings.HasPrefix(line, "D") || strings.HasPrefix(line, "A"):
							lineColor = color.FgGreen
						default:
							lineColor = color.FgRed
						}
						logger.ColorPrintf(lineColor, "          %s\n", line)
					}
				}
			}
		})
	}

	if config.Verbose <= 1 {
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
