package config

import (
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/k0kubun/pp/v3"
	"github.com/spf13/viper"
	"os"
)

const DefaultRepositoryRoot = "../graylog-project-repos"
const CIRepositoryRoot = ".repos"

type Checkout struct {
	UpdateRepos    bool     `mapstructure:"update-repos"`
	ShallowClone   bool     `mapstructure:"shallow-clone"`
	ManifestFiles  []string `mapstructure:"manifest-files"`
	Force          bool     `mapstructure:"force"`
	ModuleOverride []string `mapstructure:"module-override"`
	PullRequests   []string `mapstructure:"pull-requests"`
}

type ApplyManifest struct {
	Execute bool `mapstructure:"execute"`
	Force   bool `mapstructure:"force"`
}

type Update struct {
	Prune   bool `mapstructure:"prune"`
	Relaxed bool `mapstructure:"relaxed"`
}

type Config struct {
	RepositoryRoot     string        `mapstructure:"repository-root"`
	SelectedModules    string        `mapstructure:"selected-modules"`
	SelectedAssemblies string        `mapstructure:"selected-assemblies"`
	Checkout           Checkout      `mapstructure:"checkout"`
	ApplyManifest      ApplyManifest `mapstructure:"apply-manifest"`
	Verbose            int           `mapstructure:"verbose"`
	NoUpdateCheck      bool          `mapstructure:"disable-update-check"`
	ForceHttpsRepos    bool          `mapstructure:"force-https-repos"`
	Update             Update        `mapstructure:"update"`
}

// Returns true if running a CI environment. Detected environments: Jenkins, TravisCI
func IsCI() bool {
	return os.Getenv("CI") != "" || os.Getenv("TRAVIS") != "" || os.Getenv("BUILD_ID") != ""
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

	logger.Debug("Active configuration:\n%v", pp.Sprint(config))

	return config
}
