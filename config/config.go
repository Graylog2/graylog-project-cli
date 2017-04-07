package config

import (
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"
)

type Checkout struct {
	UpdateRepos   bool     `mapstructure:"update-repos"`
	ShallowClone  bool     `mapstructure:"shallow-clone"`
	ManifestFiles []string `mapstructure:"manifest-files"`
	Force         bool     `mapstructure:"force"`
}

type ApplyManifest struct {
	Execute bool `mapstructure:"execute"`
	Force   bool `mapstructure:"force"`
}

type Config struct {
	RepositoryRoot  string        `mapstructure:"repository-root"`
	SelectedModules string        `mapstructure:"selected-modules"`
	Checkout        Checkout      `mapstructure:"checkout"`
	ApplyManifest   ApplyManifest `mapstructure:"apply-manifest"`
	Verbose         bool          `mapstructure:"verbose"`
	NoUpdateCheck   bool          `mapstructure:"disable-update-check"`
	ForceHttpsRepos bool          `mapstructure:"force-https-repos"`
}

func Merge(config Config) Config {
	return get(config)
}

func Get() Config {
	var config Config

	return get(config)
}

func get(config Config) Config {
	if err := viper.Unmarshal(&config); err != nil {
		logger.Fatal("Unable to unmarshal config: %v", err)
	}

	logger.Debug("Active configuration:\n%v", spew.Sdump(config))

	return config
}
