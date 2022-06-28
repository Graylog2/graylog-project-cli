package runner

import (
	"fmt"
	"github.com/pkg/errors"
	"os/exec"
	"strings"
)

const (
	DevCleanupCommand  = "dev:cleanup"
	DevExecCommand     = "dev:exec"
	DevCommand         = "dev"
	DevServicesCommand = "dev:services"
	DevServerCommand   = "dev:server"
	DevWebCommand      = "dev:web"
	ReleaseCommand     = "release"
	SnapshotCommand    = "snapshot"

	EnvDockerComposeBuildImages    = "DOCKER_COMPOSE_BUILD_IMAGES"
	EnvDockerComposeCleanupVolumes = "DOCKER_COMPOSE_CLEANUP_VOLUMES"
	EnvGraylogWebHTTPPort          = "GRAYLOG_WEB_HTTP_PORT"
	EnvGraylogAPIHTTPPort          = "GRAYLOG_API_HTTP_PORT"
	EnvGraylogBuildSkipWeb         = "GRAYLOG_BUILD_SKIP_WEB"
	EnvGraylogBuildClean           = "GRAYLOG_BUILD_CLEAN"
	EnvDebug                       = "DEBUG"
	EnvMongoDBPort                 = "MONGODB_PORT"
	EnvElasticsearchPort           = "ELASTICSEARCH_PORT"

	GraylogAPIDefaultPort    = "9000"
	GraylogWebDefaultPort    = "8080"
	ElasticsearchDefaultPort = "9200"
	MongoDBDefaultPort       = "27017"
)

type Config struct {
	Command        string
	Arguments      []string
	Graylog        GraylogConfig
	Elasticsearch  ElasticsearchConfig
	MongoDB        MongoDBConfig
	RunnerRoot     string
	BuildImages    bool
	CleanupVolumes bool
}

type GraylogConfig struct {
	APIPort    string
	WebPort    string
	BuildClean bool
	BuildWeb   bool
	BuildDebug bool
}

type ElasticsearchConfig struct {
	Port string
}

type MongoDBConfig struct {
	Port string
}

func DispatchCommand(config Config) error {
	switch config.Command {
	case DevCleanupCommand:
		fallthrough
	case DevExecCommand:
		fallthrough
	case DevCommand:
		fallthrough
	case DevServerCommand:
		fallthrough
	case DevWebCommand:
		fallthrough
	case DevServicesCommand:
		return execRunnerScript(config, getEnv(config))
	default:
		return errors.Errorf("%s command not supported yet", config.Command)
	}
}

func getEnv(config Config) []string {
	var env []string

	if config.BuildImages {
		env = append(env, fmt.Sprintf("%s=%s", EnvDockerComposeBuildImages, "true"))
	}
	if config.Graylog.BuildClean {
		env = append(env, fmt.Sprintf("%s=%s", EnvGraylogBuildClean, "true"))
	}
	if config.Graylog.BuildWeb {
		env = append(env, fmt.Sprintf("%s=%s", EnvGraylogBuildSkipWeb, "false"))
	}
	if config.Graylog.BuildDebug {
		env = append(env, fmt.Sprintf("%s=%s", EnvDebug, "true"))
	}
	if config.CleanupVolumes {
		env = append(env, fmt.Sprintf("%s=%s", EnvDockerComposeCleanupVolumes, "true"))
	}

	// Don't set the environment variables if the default doesn't change to make sure the .env file for
	// docker-compose works. Variables defined in the environment are preferred over the one in the .env file by
	// docker-compose.
	if config.Graylog.APIPort != GraylogAPIDefaultPort {
		env = append(env, fmt.Sprintf("%s=%s", EnvGraylogAPIHTTPPort, config.Graylog.APIPort))
	}
	if config.Graylog.WebPort != GraylogWebDefaultPort {
		env = append(env, fmt.Sprintf("%s=%s", EnvGraylogWebHTTPPort, config.Graylog.WebPort))
	}
	if config.MongoDB.Port != MongoDBDefaultPort {
		env = append(env, fmt.Sprintf("%s=%s", EnvMongoDBPort, config.MongoDB.Port))
	}
	if config.Elasticsearch.Port != ElasticsearchDefaultPort {
		env = append(env, fmt.Sprintf("%s=%s", EnvElasticsearchPort, config.Elasticsearch.Port))
	}

	return env
}

func CheckSetup() error {
	if err := checkCommandSetup("docker", "version"); err != nil {
		return errors.Wrapf(err, "docker check failed - make sure it's installed and works properly (e.g. add your own user to the docker system group)")
	}
	if err := checkCommandSetup("docker-compose", "version", "--short"); err != nil {
		return errors.Wrapf(err, "docker-compose check failed - make sure it's installed and works properly")
	}
	return nil
}

func checkCommandSetup(cmd string, args ...string) error {
	if _, err := exec.Command(cmd, args...).Output(); err != nil {
		return errors.Wrapf(err, "couldn't execute: %s %s", cmd, strings.Join(args, " "))
	}
	return nil
}
