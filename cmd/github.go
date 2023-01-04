package cmd

import (
	"fmt"
	"github.com/Graylog2/graylog-project-cli/gh"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var githubCmd = &cobra.Command{
	Use:     "github",
	Aliases: []string{"gh"},
	Short:   "GitHub management",
	Long: `Management of GitHub projects.

Examples:
    # Create app installation access token for a GitHub org
    graylog-project github generate-app-token -o GitHub-org-name -a 1234 -k path/to/app/private.key
`,
}

var githubAppAccessTokenGenerateCmd = &cobra.Command{
	Use:     "generate-app-access-token",
	Aliases: []string{"gaat"},
	Short:   "Create an access token for an installed GitHub App",
	Run:     githubAppAccessTokenGenerateCommand,
}

var githubBranchProtectionCmd = &cobra.Command{
	Use:     "branch-protection",
	Aliases: []string{"bp"},
	Short:   "Manages the branch protection for a repository",
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Fatal("This command is deprecated and doesn't work anymore.")
	},
}

func init() {
	githubAppAccessTokenGenerateCmd.Flags().StringP("app-id", "a", "", "the GitHub app ID (env: GPC_GITHUB_APP_ID)")
	githubAppAccessTokenGenerateCmd.Flags().StringP("key", "k", "", "path to the private key to use for token generation (env: GPC_GITHUB_APP_KEY)")
	githubAppAccessTokenGenerateCmd.Flags().StringP("org", "o", "", "GitHub org for the generated token (app needs to be installed in the org) (env: GPC_GITHUB_ORG)")

	viper.BindPFlag("github.app-id", githubAppAccessTokenGenerateCmd.Flags().Lookup("app-id"))
	viper.BindPFlag("github.app-key", githubAppAccessTokenGenerateCmd.Flags().Lookup("key"))
	viper.BindPFlag("github.org", githubAppAccessTokenGenerateCmd.Flags().Lookup("org"))

	viper.MustBindEnv("github.app-id", "GPC_GITHUB_APP_ID")
	viper.MustBindEnv("github.app-key", "GPC_GITHUB_APP_KEY")
	viper.MustBindEnv("github.org", "GPC_GITHUB_ORG")

	githubCmd.AddCommand(githubAppAccessTokenGenerateCmd)
	githubCmd.AddCommand(githubBranchProtectionCmd)
	RootCmd.AddCommand(githubCmd)
}

type gitHubCmdConfig struct {
	GitHub struct {
		AppID  string `mapstructure:"app-id"`
		AppKey string `mapstructure:"app-key"`
		Org    string `mapstructure:"org"`
	} `mapstructure:"github"`
}

func githubAppAccessTokenGenerateCommand(cmd *cobra.Command, args []string) {
	var cfg gitHubCmdConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		logger.Fatal("Couldn't deserialize config: %s", err.Error())
	}

	if cfg.GitHub.AppID == "" {
		exitWithUsage(cmd, "Missing app ID flag")
	}
	if cfg.GitHub.AppKey == "" {
		exitWithUsage(cmd, "Missing app key flag")
	}
	if cfg.GitHub.Org == "" {
		exitWithUsage(cmd, "Missing GitHub org flag")
	}

	token, err := gh.GenerateAppToken(cfg.GitHub.Org, cfg.GitHub.AppID, cfg.GitHub.AppKey)
	if err != nil {
		logger.Fatal("ERROR: %s", err)
	}

	fmt.Println(token)
}
