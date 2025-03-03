package cmd

import (
	"fmt"
	"github.com/Graylog2/graylog-project-cli/gh"
	"github.com/Graylog2/graylog-project-cli/git"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
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

var githubRulesetsCmd = &cobra.Command{
	Use:     "rulesets",
	Aliases: []string{"rs"},
	Short:   "Manage repository rulesets",
}

var githubRulesetsEnableCmd = &cobra.Command{
	Use:     "enable [flags] OWNER/REPO RULESET",
	Short:   "Enable repository ruleset",
	RunE:    githubEnableRuleset,
	Example: "graylog-project github rulesets enable myorg/myrepo custom-ruleset-name",
}

var githubRulesetsDisableCmd = &cobra.Command{
	Use:     "disable [flags] OWNER/REPO RULESET",
	Short:   "Disable repository ruleset",
	RunE:    githubDisableRuleset,
	Example: "graylog-project github rulesets disable myorg/myrepo custom-ruleset-name",
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

var githubParsePullDependenciesCmd = &cobra.Command{
	Use:     "parse-pull-request-dependencies",
	Aliases: []string{"parse-pr-deps"},
	Short:   "Parses dependencies for a pull request",
	RunE:    githubParsePullDependencies,
}

func init() {
	githubAppAccessTokenGenerateCmd.Flags().StringP("app-id", "a", "", "the GitHub app ID (env: GPC_GITHUB_APP_ID)")
	githubAppAccessTokenGenerateCmd.Flags().StringP("key-file", "k", "", "path to the private key to use for token generation (or key from env: GPC_GITHUB_APP_KEY)")
	githubAppAccessTokenGenerateCmd.Flags().StringP("org", "o", "", "GitHub org for the generated token (app needs to be installed in the org) (env: GPC_GITHUB_ORG)")
	githubRulesetsCmd.PersistentFlags().Bool("detect-repo", false, "Auto detect the repository")

	viper.BindPFlag("github.app-id", githubAppAccessTokenGenerateCmd.Flags().Lookup("app-id"))
	viper.BindPFlag("github.app-key-file", githubAppAccessTokenGenerateCmd.Flags().Lookup("key-file"))
	viper.BindPFlag("github.org", githubAppAccessTokenGenerateCmd.Flags().Lookup("org"))
	viper.BindPFlag("github.detect-repo", githubRulesetsCmd.PersistentFlags().Lookup("detect-repo"))

	viper.MustBindEnv("github.app-id", "GPC_GITHUB_APP_ID")
	viper.MustBindEnv("github.app-key", "GPC_GITHUB_APP_KEY")
	viper.MustBindEnv("github.org", "GPC_GITHUB_ORG")
	viper.MustBindEnv("github.access-token", "GPC_GITHUB_TOKEN", "GITHUB_ACCESS_TOKEN")

	githubRulesetsCmd.AddCommand(githubRulesetsEnableCmd)
	githubRulesetsCmd.AddCommand(githubRulesetsDisableCmd)

	githubParsePullDependenciesCmd.Flags().StringSliceP("add", "a", []string{}, "Add additional dependencies")
	githubParsePullDependenciesCmd.Flags().StringP("file", "f", "-", "Read pull request body from file")
	githubParsePullDependenciesCmd.Flags().StringP("join", "j", "", "Join output to one line by given join string")

	githubCmd.AddCommand(githubAppAccessTokenGenerateCmd)
	githubCmd.AddCommand(githubRulesetsCmd)
	githubCmd.AddCommand(githubBranchProtectionCmd)
	githubCmd.AddCommand(githubParsePullDependenciesCmd)
	RootCmd.AddCommand(githubCmd)
}

type gitHubCmdConfig struct {
	GitHub struct {
		AppID       string `mapstructure:"app-id"`
		AppKeyFile  string `mapstructure:"app-key-file"`
		AppKey      string `mapstructure:"app-key"`
		Org         string `mapstructure:"org"`
		AccessToken string `mapstructure:"access-token"`
		DetectRepo  bool   `mapstructure:"detect-repo"`
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
	if cfg.GitHub.AppKey == "" && cfg.GitHub.AppKeyFile == "" {
		exitWithUsage(cmd, "Missing app key GPC_GITHUB_APP_KEY or key file flag")
	}
	if cfg.GitHub.Org == "" {
		exitWithUsage(cmd, "Missing GitHub org flag")
	}

	appKey := cfg.GitHub.AppKey
	if cfg.GitHub.AppKeyFile != "" {
		buf, err := os.ReadFile(cfg.GitHub.AppKeyFile)
		appKey = string(buf)
		if err != nil {
			logger.Fatal("ERROR: couldn't read key file: %s", err)
		}
	}

	token, err := gh.GenerateAppToken(cfg.GitHub.Org, cfg.GitHub.AppID, appKey)
	if err != nil {
		logger.Fatal("ERROR: %s", err)
	}

	fmt.Println(token)
}

func githubToggleRuleset(_ *cobra.Command, args []string, cb func(client *gh.Client, owner, repo, ruleset string) (*gh.Ruleset, error)) error {
	var cfg gitHubCmdConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("couldn't deserialize config: %w", err)
	}

	var repoString string
	var rulesetName string

	if cfg.GitHub.DetectRepo {
		if len(args) != 1 {
			return fmt.Errorf("expected one argument")
		}
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("couldn't get current working directory: %w", err)
		}
		urlString, err := git.GetRemoteUrl(cwd, "origin")
		if err != nil {
			return err
		}
		url, err := utils.ParseGitHubURL(urlString)
		if err != nil {
			return err
		}
		repoString = url.Repository()
		rulesetName = args[0]
	} else {
		if len(args) != 2 {
			return fmt.Errorf("expected two arguments")
		}
		repoString = args[0]
		rulesetName = args[1]
	}

	owner, repo, err := gh.SplitRepoString(repoString)
	if err != nil {
		return err
	}

	ruleset := strings.TrimSpace(rulesetName)

	if ruleset == "" {
		return fmt.Errorf("ruleset can't be blank")
	}

	if cfg.GitHub.AccessToken == "" {
		return fmt.Errorf("missing GitHub access token (GITHUB_ACCESS_TOKEN)")
	}

	client := gh.NewGitHubClient(cfg.GitHub.AccessToken)

	if ruleset, err := cb(client, owner, repo, ruleset); err != nil {
		return err
	} else {
		logger.Info("Ruleset enforcement for ruleset %q in %s/%s: %s (ID: %d)", rulesetName, ruleset.Owner, ruleset.Repo, ruleset.Enforcement, ruleset.ID)
		return nil
	}
}

func githubEnableRuleset(cmd *cobra.Command, args []string) error {
	return githubToggleRuleset(cmd, args, func(client *gh.Client, owner, repo, ruleset string) (*gh.Ruleset, error) {
		return client.EnableRulesetByName(owner, repo, ruleset)
	})
}

func githubDisableRuleset(cmd *cobra.Command, args []string) error {
	return githubToggleRuleset(cmd, args, func(client *gh.Client, owner, repo, ruleset string) (*gh.Ruleset, error) {
		return client.DisableRulesetByName(owner, repo, ruleset)
	})
}

func githubParsePullDependencies(cmd *cobra.Command, args []string) error {
	additional, err := cmd.Flags().GetStringSlice("add")
	if err != nil {
		return fmt.Errorf("couldn't get additional dependencies option: %w", err)
	}
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return fmt.Errorf("couldn't get file option: %w", err)
	}
	join, err := cmd.Flags().GetString("join")
	if err != nil {
		return fmt.Errorf("couldn't get join option: %w", err)
	}

	fd := os.Stdin
	if file == "-" {
		if isatty.IsTerminal(os.Stdin.Fd()) {
			return fmt.Errorf("couldn't read pull request body from standard input because it's a terminal")
		}
	} else {
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("couldn't open file: %w", err)
		}
		fd = f
	}

	dependencies, err := gh.ParsePullDependencies(fd)
	if err != nil {
		return err
	}

	if (len(dependencies) + len(additional)) > 0 {
		joinValue := "\n"
		if join != "" {
			joinValue = join
		}
		fmt.Println(strings.Join(append(additional, dependencies...), joinValue))
	}

	return nil
}
