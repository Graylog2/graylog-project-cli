package cmd

import (
	cfg "github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/gh"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/spf13/cobra"
	"os"
)

var githubCmd = &cobra.Command{
	Use:     "github",
	Aliases: []string{"gh"},
	Short:   "GitHub management",
	Long: `Management of GitHub projects.

Examples:
    # Enable branch protection for a GitHub repository branch
    graylog-project github branch-protection --enable --repo Graylog2/graylog2-server --branch 2.4

    # Disable branch protection for a GitHub repository branch
    graylog-project github branch-protection --disable --repo Graylog2/graylog2-server --branch 2.4

    # Enable branch protection for all GitHub repositories in the currently checked out manifest
    graylog-project github branch-protection --enable --manifest --branch 2.4
`,
}

var githubBranchProtectionCmd = &cobra.Command{
	Use:     "branch-protection",
	Aliases: []string{"bp"},
	Short:   "Manages the branch protection for a repository",
	Run:     githubBranchProtectionCommand,
}

var githubBPEnable bool
var githubBPDisable bool
var githubBPManifest bool
var githubBPDryRun bool
var githubRepository string
var githubRepositoryBranch string

func init() {
	githubCmd.AddCommand(githubBranchProtectionCmd)
	RootCmd.AddCommand(githubCmd)

	githubBranchProtectionCmd.Flags().BoolVar(&githubBPEnable, "enable", false, "Enable branch protection for the GitHub repository")
	githubBranchProtectionCmd.Flags().BoolVar(&githubBPDisable, "disable", false, "Disable branch protection for the GitHub repository")
	githubBranchProtectionCmd.Flags().BoolVar(&githubBPManifest, "manifest", false, "Toggle branch protection for all repos in the manifest")
	githubBranchProtectionCmd.Flags().BoolVar(&githubBPDryRun, "dry-run", false, "Don't execute branch protection commands, only show what would be done.")
	githubBranchProtectionCmd.Flags().StringVarP(&githubRepository, "repo", "r", "", "The GitHub repository name (e.g. Graylog2/graylog2-server")
	githubBranchProtectionCmd.Flags().StringVarP(&githubRepositoryBranch, "branch", "b", "", "The GitHub repository branch (e.g. master")
}

func githubBranchProtectionCommand(cmd *cobra.Command, args []string) {
	if !githubBPEnable && !githubBPDisable {
		logger.Fatal("ERROR: You need to use the --enable or --disable flag")
	}
	if githubRepositoryBranch == "" {
		logger.Fatal("--branch flag must be set")
	}

	if githubBPManifest {
		config := cfg.Get()
		manifestFiles := manifest.ReadState().Files()
		project := p.New(config, manifestFiles)

		p.ForEachSelectedModule(project, func(module p.Module) {
			url, err := utils.ParseGitHubURL(module.Repository)
			if err != nil {
				logger.Fatal("ERROR: %s", err)
			}
			toggleGitHubBranchProtection(url.Repository)
		})
	} else {
		if githubRepository == "" {
			logger.Fatal("--repo flag must be set")
		}
		toggleGitHubBranchProtection(githubRepository)
	}
}

func toggleGitHubBranchProtection(repo string) {
	accessToken := os.Getenv("GPC_GITHUB_TOKEN")

	if accessToken == "" {
		logger.Fatal("ERROR: Missing GPC_GITHUB_TOKEN environment variable")
	}

	owner, name, err := gh.SplitRepoString(repo)
	if err != nil {
		logger.Fatal("ERROR: %s", err)
	}

	branch := githubRepositoryBranch
	client := gh.NewGitHubClient(accessToken)

	if githubBPEnable {
		if !githubBPDryRun {
			logger.Info("Adding branch protection from GitHub repository: %s/%s@%s", owner, name, branch)
			if err := client.EnableBranchProtection(owner, name, branch); err != nil {
				logger.Fatal("ERROR: Unable to enable branch protection for %s/%s@%s: %s",
					owner, name, branch, err)
			}
		} else {
			logger.Info("Would add branch protection from GitHub repository: %s/%s@%s", owner, name, branch)
		}
	} else if githubBPDisable {
		if !githubBPDryRun {
			logger.Info("Removing branch protection from GitHub repository: %s/%s@%s", owner, name, branch)
			if err := client.DisableBranchProtection(owner, name, branch); err != nil {
				logger.Fatal("ERROR: Unable to disable branch protection for %s/%s@%s: %s",
					owner, name, branch, err)
			}
		} else {
			logger.Info("Would remove branch protection from GitHub repository: %s/%s@%s", owner, name, branch)
		}
	}
}
