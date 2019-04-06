package cmd

import (
	"github.com/Graylog2/graylog-project-cli/git"
	"github.com/Graylog2/graylog-project-cli/hooks"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

const DefaultProjectManifest = "manifests/master.json"

var bootstrapShallowClone bool
var bootstrapCheckoutPath string
var bootstrapManifest string
var bootstrapProjectBranch string
var bootstrapSkipHooks bool

// bootstrapCmd represents the bootstrap command
var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap <repository>",
	Short: "Clone and setup graylog-project repository",
	Long: `This clones the given repository and runs a checkout for the given manifest. (defaults to master.json)

Example:
  # Clones the graylog-project repo and runs a checkout for the "manifests/master.json" manifest
  graylog-project bootstrap github://Graylog2/graylog-project.git

  # Clones the graylog-project repo and runs a checkout for the "manifests/2.1.json" manifest
  graylog-project bootstrap -m manifests/2.1.json github://Graylog2/graylog-project.git
`,
	Run: bootstrapCommand,
}

func init() {
	RootCmd.AddCommand(bootstrapCmd)

	bootstrapCmd.Flags().BoolVarP(&bootstrapShallowClone, "shallow-clone", "s", false, "Create a shallow git clone instead of a regular one")
	bootstrapCmd.Flags().StringVarP(&bootstrapCheckoutPath, "checkout-path", "p", "", "Path for the graylog-project checkout")
	bootstrapCmd.Flags().StringVarP(&bootstrapManifest, "manifest", "m", DefaultProjectManifest, "Manifest to checkout")
	bootstrapCmd.Flags().StringVarP(&bootstrapProjectBranch, "project-branch", "B", "master", "graylog-project branch to check out")
	bootstrapCmd.Flags().StringP("auth-token", "T", "", "Auth token to access protected URLs")
	bootstrapCmd.Flags().BoolVarP(&bootstrapSkipHooks, "skip-hooks", "", false, "Do not execute hooks")

	viper.BindPFlag("checkout.auth-token", bootstrapCmd.Flags().Lookup("auth-token"))
	viper.BindEnv("checkout.auth-token", "GPC_AUTH_TOKEN")
}

func bootstrapCommand(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		logger.Info("Missing repository argument")
		cmd.UsageFunc()(cmd)
		os.Exit(1)
	}

	repoUrl, err := utils.ParseGitHubURL(args[0])
	if err != nil {
		logger.Fatal("%v", err)
	}

	if bootstrapCheckoutPath == "" {
		bootstrapCheckoutPath = repoUrl.Directory()
	}

	if utils.FileExists(bootstrapCheckoutPath) {
		logger.Info("Directory %s already exists", bootstrapCheckoutPath)
		return
	}

	cloneUrl := repoUrl.SSH()

	if repoUrl.IsHTTPS() {
		cloneUrl = repoUrl.HTTPS()
		viper.Set("force-https-repos", true)
	}

	if bootstrapShallowClone {
		viper.Set("checkout.shallow-clone", true)
		git.Git("clone", "--depth=1", "--no-single-branch", cloneUrl, bootstrapCheckoutPath)
	} else {
		git.Git("clone", cloneUrl, bootstrapCheckoutPath)
	}

	utils.InDirectory(bootstrapCheckoutPath, func() {
		git.Git("checkout", bootstrapProjectBranch)
		checkoutCommand(cmd, []string{bootstrapManifest})

		if !bootstrapSkipHooks {
			if hooks.Run(cmd.Name(), false) != nil {
				logger.Error("Couldn't run hooks for %s: %s", cmd.Name(), err)
			}
		}
	})

}
