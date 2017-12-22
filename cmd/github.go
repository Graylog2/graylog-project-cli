package cmd

import (
	"github.com/Graylog2/graylog-project-cli/gh"
	"github.com/Graylog2/graylog-project-cli/logger"
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
var githubRepository string
var githubRepositoryBranch string

func init() {
	githubCmd.AddCommand(githubBranchProtectionCmd)
	RootCmd.AddCommand(githubCmd)

	githubBranchProtectionCmd.Flags().BoolVar(&githubBPEnable, "enable", false, "Enable branch protection for the GitHub repository")
	githubBranchProtectionCmd.Flags().BoolVar(&githubBPDisable, "disable", false, "Disable branch protection for the GitHub repository")
	githubBranchProtectionCmd.Flags().StringVarP(&githubRepository, "repo", "r", "Graylog2/graylog2-server", "The GitHub repository name (e.g. Graylog2/graylog2-server")
	githubBranchProtectionCmd.Flags().StringVarP(&githubRepositoryBranch, "branch", "b", "master", "The GitHub repository branch (e.g. master")
}

func githubBranchProtectionCommand(cmd *cobra.Command, args []string) {
	if !githubBPEnable && !githubBPDisable {
		logger.Fatal("ERROR: You need to use the --enable or --disable flag")
	}

	accessToken := os.Getenv("GPC_GITHUB_TOKEN")

	if accessToken == "" {
		logger.Fatal("ERROR: Missing GPC_GITHUB_TOKEN environment variable")
	}

	client := gh.NewGitHubClient(accessToken)

	owner, name, err := gh.SplitRepoString(githubRepository)
	if err != nil {
		logger.Fatal("ERROR: %s", err)
	}

	if githubBPEnable {
		logger.Info("Adding branch protection from GitHub repository: %s@%s", githubRepository, githubRepositoryBranch)
		if err := client.EnableBranchProtection(owner, name, githubRepositoryBranch); err != nil {
			logger.Fatal("ERROR: Unable to enable branch protection for %s/%s@%s: %s",
				owner, name, githubRepositoryBranch, err)
		}
	} else if githubBPDisable {
		logger.Info("Removing branch protection from GitHub repository: %s@%s", githubRepository, githubRepositoryBranch)
		if err := client.DisableBranchProtection(owner, name, githubRepositoryBranch); err != nil {
			logger.Fatal("ERROR: Unable to disable branch protection for %s/%s@%s: %s",
				owner, name, githubRepositoryBranch, err)
		}
	}
}
