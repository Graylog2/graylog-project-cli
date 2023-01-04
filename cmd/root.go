package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	gitRevision string
	buildDate   string
	gitTag      string
)

var cfgFile string
var repositoryRoot string
var debug bool
var verbose int
var selectedModules string
var selectedAssemblies string
var loggerPrefix string
var noUpdateCheck bool
var forceHttpsRepos bool

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   filepath.Base(os.Args[0]),
	Short: "graylog-project management CLI",
	Long: `
CLI tool to manage a graylog-project setup

Some command line options can be configured with a yaml config file.

Configuration options:

- repository-root
- checkout.update-repos (see checkout command)
- checkout.shallow-clone (see checkout command)

Example config file:

  ---
  repository-root: "../gpc"

Configuration file lookup order

1. $PWD/.graylog-project-cli.yml
2. $HOME/.graylog-project-cli.yml

Environment variables:

- GPC_REPOSITORY_ROOT: can be used instead of the "repository-root" command line flag

Example usage:

  $ graylog-project checkout manifests/master.json

  $ graylog-project npm install

  $ graylog-project status
` + fmt.Sprintf("\n\nVersion:      %v\n", gitTag) + fmt.Sprintf("Build date:   %v\n", buildDate) + fmt.Sprintf("Git revision: %v\n\n", gitRevision),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.SetDebug(viper.GetBool("debug"))
		logger.SetPrefix(viper.GetString("logger-prefix"))

		// We use a special repository root if running in a CI environment and the default hasn't been changed.
		if config.IsCI() && viper.GetString("repository-root") == config.DefaultRepositoryRoot {
			logger.Info("Running in a CI environment, using repository root: %s", config.CIRepositoryRoot)
			viper.Set("repository-root", config.CIRepositoryRoot)
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	RootCmd.PersistentFlags().StringVar(&repositoryRoot, "repository-root", config.DefaultRepositoryRoot, "Git repository root")
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "D", false, "enable debug output (default: false)")
	RootCmd.PersistentFlags().CountVarP(&verbose, "verbose", "v", "enable verbose output - use multiple times to increase verbosity")
	RootCmd.PersistentFlags().StringVarP(&selectedModules, "selected-modules", "M", "", "apply command to given modules (comma separated)")
	RootCmd.PersistentFlags().StringVarP(&selectedAssemblies, "selected-assemblies", "Y", "", "apply command to modules that match the given assembly filter (comma separated - use \"-\" prefix to negate selection)")
	RootCmd.PersistentFlags().StringVarP(&loggerPrefix, "logger-prefix", "", "", "output logger prefix")
	RootCmd.PersistentFlags().BoolVarP(&noUpdateCheck, "disable-update-check", "U", false, "disable checking for graylog-project-cli updates")
	RootCmd.PersistentFlags().BoolVarP(&forceHttpsRepos, "force-https-repos", "", false, "convert all git@github.com:... repository URLs to https://github.com/...")

	viper.BindPFlag("repository-root", RootCmd.PersistentFlags().Lookup("repository-root"))
	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("selected-modules", RootCmd.PersistentFlags().Lookup("selected-modules"))
	viper.BindPFlag("selected-assemblies", RootCmd.PersistentFlags().Lookup("selected-assemblies"))
	viper.BindPFlag("logger-prefix", RootCmd.PersistentFlags().Lookup("logger-prefix"))
	viper.BindPFlag("disable-update-check", RootCmd.PersistentFlags().Lookup("disable-update-check"))
	viper.BindPFlag("force-https-repos", RootCmd.PersistentFlags().Lookup("force-https-repos"))

	viper.BindEnv("repository-root", "GPC_REPOSITORY_ROOT")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".graylog-project-cli") // name of config file (without extension)
	viper.AddConfigPath("$HOME")                // adding home directory as first search path
	viper.AddConfigPath(".")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logger.Debug("Using config file: %v", viper.ConfigFileUsed())
	} else {
		logger.Debug("Error reading config file: %v", err)
	}
}
