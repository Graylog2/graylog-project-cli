package runner

import (
	"github.com/pkg/errors"
	"os/exec"
	"strings"
)

type Config struct {
	Command       string
	Graylog       GraylogConfig
	Elasticsearch ElasticsearchConfig
	MongoDB       MongoDBConfig
}

type GraylogConfig struct {
	HTTPPort int
}

type ElasticsearchConfig struct {
	Port int
}

type MongoDBConfig struct {
	Port int
}

func DispatchCommand(config Config) error {
	return nil
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
