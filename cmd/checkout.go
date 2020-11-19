package cmd

import (
	c "github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/manifest"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/projectstate"
	"github.com/Graylog2/graylog-project-cli/repo"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// checkoutCmd represents the checkout command
var checkoutCmd = &cobra.Command{
	Use:     "checkout",
	Aliases: []string{"co"},
	Short:   "Update project for the given manifest",
	Long: `
This will update the project based on the given manifest.

Examples:

  $ graylog-project checkout manifests/master.json

  $ graylog-project co manifests/master.json /path/to/other/manifest.json

  # Override graylog-plugin-collector module in manifest to checkout revision
  # "abc123"  of the a-contributor/graylog-plugin-collector repository
  # instead of Graylog2/graylog-plugin-collector
  $ graylog-project co --module-override Graylog2/graylog-plugin-collector=a-contributor/graylog-plugin-collector@abc123

  # To checkout GitHub pull-requests, use the --pull-requests flag. (only one PR per repository, last one wins)
  $ graylog-project co --pull-requests Graylog2/graylog-plugin-collector#123
`,
	Run: checkoutCommand,
}

func init() {
	RootCmd.AddCommand(checkoutCmd)

	checkoutCmd.Flags().BoolP("update-repos", "u", false, "Fetch latest commits from remote")
	checkoutCmd.Flags().BoolP("shallow-clone", "s", false, "Create a shallow git clone instead of a regular one")
	checkoutCmd.Flags().BoolP("force", "f", false, "Force checkout event though repository is unexpected")
	checkoutCmd.Flags().StringP("auth-token", "T", "", "Auth token to access protected URLs")
	checkoutCmd.Flags().StringSliceP("module-override", "O", []string{}, "Override manifest modules, see help for details")
	checkoutCmd.Flags().StringSliceP("pull-requests", "p", []string{}, "Checkout GitHub pull requests (e.g. Graylog2/graylog2-server#123)")

	viper.BindPFlag("checkout.update-repos", checkoutCmd.Flags().Lookup("update-repos"))
	viper.BindPFlag("checkout.shallow-clone", checkoutCmd.Flags().Lookup("shallow-clone"))
	viper.BindPFlag("checkout.force", checkoutCmd.Flags().Lookup("force"))
	viper.BindPFlag("checkout.auth-token", checkoutCmd.Flags().Lookup("auth-token"))
	viper.BindPFlag("checkout.module-override", checkoutCmd.Flags().Lookup("module-override"))
	viper.BindPFlag("checkout.pull-requests", checkoutCmd.Flags().Lookup("pull-requests"))

	viper.BindEnv("checkout.auth-token", "GPC_AUTH_TOKEN")
}

func prepareCheckoutCommand(cmd *cobra.Command, args []string) (c.Config, *repo.RepoManager, p.Project) {
	if len(args) < 1 {
		logger.Info("Missing manifest argument")
		cmd.UsageFunc()(cmd)
		os.Exit(1)
	}

	var defaultConfig c.Config

	// The only arguments to the command are the manifest files
	defaultConfig.Checkout.ManifestFiles = handleManifestArguments(args[0:])

	config := c.Merge(defaultConfig)

	repoManager := repo.NewRepoManager(config)

	logger.Debug("Using manifests: %v", config.Checkout.ManifestFiles)

	project := p.New(config, config.Checkout.ManifestFiles, p.WithModuleOverride(), p.WithPullRequests())

	return config, repoManager, project
}

func handleManifestArguments(manifests []string) []string {
	files := make([]string, 0)
	authToken := viper.GetString("checkout.auth-token")

	for _, file := range manifests {
		// If the manifest argument looks like a GitHub URL we try to download it
		if strings.HasPrefix(file, "http") && strings.Contains(file, "github") {
			files = append(files, manifest.DownloadManifestFromGitHub(file, authToken))
		} else {
			files = append(files, file)
		}
	}

	return files
}

func cleanupManifestFiles(manifestFiles []string) []string {
	projectManifests := make([]string, 0)
	files := make([]string, 0)

	// Collect all manifest files in the project
	filepath.Walk(utils.GetCwd(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".json") {
			projectManifests = append(projectManifests, path)
		}

		return nil
	})

manifest_files:
	for _, file := range manifestFiles {
		// We only touch downloaded manifest files
		if !strings.Contains(file, manifest.DownloadedManifestPrefix) {
			files = append(files, file)
			continue
		}

		// Check if the downloaded manifest is the same as an existing one so we don't have to copy the
		// downloaded one into the manifests folder
		for _, projectFile := range projectManifests {
			if utils.FilesIdentical(file, projectFile) {
				logger.Info("Downloaded manifest <%s> is the same as <%s>", file, projectFile)
				files = append(files, utils.GetRelativePath(projectFile))
				if err := os.Remove(file); err != nil {
					logger.Error("Unable to remove file <%s>: %v", file, err)
				}
				continue manifest_files
			}
		}

		// If the manifest file looks like a downloaded one, copy it into the local manifests folder
		// so we can access it later
		buf, err := ioutil.ReadFile(file)
		if err != nil {
			logger.Fatal("Unable to read file <%s>: %v", file, err)
		}

		output := filepath.Join("manifests", filepath.Base(file))
		logger.Info("Writing downloaded manifest to: %s", output)
		if err = ioutil.WriteFile(output, buf, 0644); err != nil {
			logger.Fatal("Unable to write file <%s>: %v", output, err)
		}
		files = append(files, output)

		if err := os.Remove(file); err != nil {
			logger.Error("Unable to remove file <%s>: %v", file, err)
		}
	}

	return files
}

func checkoutCommand(cmd *cobra.Command, args []string) {
	config, repoManager, project := prepareCheckoutCommand(cmd, args)

	repoManager.SetupProjectRepositories(project)

	projectstate.Sync(project, config)

	manifest.WriteState(cleanupManifestFiles(config.Checkout.ManifestFiles))

	CheckForUpdate()
}
