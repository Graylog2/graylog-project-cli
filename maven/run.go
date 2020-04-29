package maven

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/manifest"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/google/renameio"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type RunConfig struct {
	HTTPPort      int
	ESPort        int
	MongoDBPort   int
	WorkDir       string
	ConfigFile    string
	ClassPathFile string
}

func (rc RunConfig) NodeIDFile() string {
	return filepath.Join(rc.WorkDir, "server-node-id-01")
}

func (rc RunConfig) DataDir() string {
	return filepath.Join(rc.WorkDir, "data")
}

func (rc RunConfig) ConfigFilePath() string {
	return filepath.Join(rc.WorkDir, rc.ConfigFile)
}

func (rc RunConfig) ServerRepoPath() (string, error) {
	manifestFiles := manifest.ReadState().Files()
	project := p.New(config.Get(), manifestFiles)

	serverPath, err := filepath.Abs(project.Server.Path)
	if err != nil {
		return "", errors.Wrapf(err, "couldn't get absolute path for %s", project.Server.Path)
	}
	return serverPath, nil
}

func RunServer(config RunConfig) error {
	javaBin, err := exec.LookPath("java")
	if err != nil {
		return errors.Wrap(err, "couldn't find java executable. please make sure java is installed.")
	}

	if err := checkRunEnv(config); err != nil {
		return err
	}
	if err := setupRunEnv(config); err != nil {
		return err
	}

	content, err := ioutil.ReadFile(config.ClassPathFile)
	if err != nil {
		return errors.Wrapf(err, "couldn't read classpath file: %s", config.ClassPathFile)
	}
	classPathString := strings.TrimSpace(string(content))

	serverRepoPath, err := config.ServerRepoPath()
	if err != nil {
		return err
	}

	return runServer(
		config.WorkDir,
		javaBin,
		fmt.Sprintf("-Djava.library.path=%s/lib/sigar-1.6.4", serverRepoPath),
		"-Dio.netty.leakDetection.level=paranoid",
		//"-agentlib:jdwp=transport=dt_socket,server=n,address=127.0.0.1:5005,suspend=y", // TODO: Make configurable
		"-Xms1g",
		"-Xmx1g",
		"-server",
		"-XX:-OmitStackTraceInFastThrow",
		"-classpath",
		classPathString,
		"org.graylog2.bootstrap.Main",
		"server",
		"-f",
		config.ConfigFile,
		"-np",
		"--local",
	)
}

func checkRunEnv(config RunConfig) error {
	// TODO: Check for ES and MongoDB ports
	return nil
}

func setupRunEnv(config RunConfig) error {
	// TODO:
	//  - Create a work directory with a data/ dir and generate a node-id file
	//  - Create a default graylog.conf in the work directory with sane defaults (e.g. no versionchecks, lb timeout 0, etc)

	if !utils.FileExists(config.WorkDir) {
		if err := os.MkdirAll(config.WorkDir, 0755); err != nil {
			return errors.Wrapf(err, "couldn't create work dir %s", config.WorkDir)
		}
		if err := os.MkdirAll(config.DataDir(), 0755); err != nil {
			return errors.Wrapf(err, "couldn't create data dir %s", config.DataDir())
		}
	}

	if config.ConfigFile == "" {
		config.ConfigFile = "graylog.conf"
	}

	if !utils.FileExists(config.ConfigFilePath()) {
		if err := createServerConfigFile(config); err != nil {
			return err
		}
	}

	return nil
}

func createServerConfigFile(config RunConfig) error {
	serverPath, err := config.ServerRepoPath()
	if err != nil {
		return err
	}

	serverConfig := filepath.Join(serverPath, "misc", "graylog.conf")
	userConfig, err := parseUserConfig(serverPath)
	if err != nil {
		return err
	}

	// Set some defaults for graylog.conf
	userConfig.ConfigValues["node_id_file"] = filepath.Base(config.NodeIDFile())

	data, err := ioutil.ReadFile(serverConfig)
	if err != nil {
		return errors.Wrapf(err, "couldn't read config file from: %s", serverConfig)
	}

	file, err := renameio.TempFile("", config.ConfigFilePath())
	if err != nil {
		return errors.Wrapf(err, "couldn't open config file for writing: %s", config.ConfigFilePath())
	}
	//noinspection ALL
	defer file.Cleanup()

	// Process the example graylog.conf line by line
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		lineWritten := false

		// Check if we have anything in the user config that should be set in the new file
		for key, value := range userConfig.ConfigValues {
			if strings.HasPrefix(line, key) {
				_, err := file.WriteString(fmt.Sprintf("%s = %s\n", key, value))
				if err != nil {
					return errors.Wrapf(err, "couldn't write line to config file %s", file.Name())
				}
				lineWritten = true
			}
		}

		if matched, _ := regexp.MatchString("^#http_bind_address = \\S+?:\\d+", line); matched {
			_, err := file.WriteString(fmt.Sprintf("http_bind_address = 127.0.0.1:%d\n", config.HTTPPort))
			if err != nil {
				return errors.Wrapf(err, "couldn't write line to config file %s", config.ConfigFilePath())
			}
			lineWritten = true
		}
		if strings.HasPrefix(line, "mongodb_uri") {
			_, err := file.WriteString(fmt.Sprintf("mongodb_uri = mongodb://127.0.0.1:%d/graylog\n", config.MongoDBPort))
			if err != nil {
				return errors.Wrapf(err, "couldn't write line to config file %s", config.ConfigFilePath())
			}
			lineWritten = true
		}
		if strings.HasPrefix(line, "elasticsearch_hosts") {
			_, err := file.WriteString(fmt.Sprintf("elasticsearch_hosts = http://127.0.0.1:%d\n", config.ESPort))
			if err != nil {
				return errors.Wrapf(err, "couldn't write line to config file %s", config.ConfigFilePath())
			}
			lineWritten = true
		}

		if !lineWritten {
			_, err := file.WriteString(line + "\n")
			if err != nil {
				return errors.Wrapf(err, "couldn't write line to config file %s", file.Name())
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return errors.Wrap(err, "couldn't scan config file data")
	}

	if err := file.CloseAtomicallyReplace(); err != nil {
		return errors.Wrapf(err, "couldn't write config file to: %s", config.ConfigFilePath())
	}

	if err := writeNodeIdFile(config, userConfig.NodeID); err != nil {
		return err
	}

	return nil
}

func writeNodeIdFile(config RunConfig, nodeIdValue string) error {
	if !utils.FileExists(config.NodeIDFile()) {
		nodeId := strings.TrimSpace(nodeIdValue)

		if nodeId == "" {
			newNodeId, err := uuid.NewRandom()
			if err != nil {
				return errors.Wrap(err, "couldn't generate new node ID")
			}
			nodeId = newNodeId.String()
		}

		if err := ioutil.WriteFile(config.NodeIDFile(), []byte(nodeId), 0644); err != nil {
			return errors.Wrapf(err, "couldn't write node ID to file %s", config.NodeIDFile())
		}
	}

	return nil
}

type userConfig struct {
	NodeID       string
	ConfigValues map[string]string
}

func parseUserConfig(serverPath string) (*userConfig, error) {
	userConfig := &userConfig{
		ConfigValues: map[string]string{},
	}
	userConfigFile := filepath.Join(serverPath, "graylog.conf")

	if !utils.FileExists(userConfigFile) {
		return userConfig, nil
	}

	configData, err := ioutil.ReadFile(userConfigFile)
	if err != nil {
		return userConfig, errors.Wrapf(err, "couldn't read config file from: %s", userConfigFile)
	}

	scanner := bufio.NewScanner(bytes.NewReader(configData))
	for scanner.Scan() {
		line := scanner.Text()

		parseConfigLine(line, "password_secret", userConfig.ConfigValues)
		parseConfigLine(line, "root_password_sha2", userConfig.ConfigValues)
		parseConfigLine(line, "node_id_file", userConfig.ConfigValues)
	}

	// We need the node ID value for the new file so read it from the old file
	if file, ok := userConfig.ConfigValues["node_id_file"]; ok && strings.TrimSpace(file) != "" {
		nodeIdFile := strings.TrimSpace(file)
		// Remove the config option because we want to use our default
		delete(userConfig.ConfigValues, "node_id_file")

		if !filepath.IsAbs(nodeIdFile) {
			nodeIdFile = filepath.Join(serverPath, nodeIdFile)
		}

		nodeIdBytes, err := ioutil.ReadFile(nodeIdFile)
		if err != nil {
			return userConfig, errors.Wrapf(err, "couldn't read file: %s", nodeIdFile)
		}

		userConfig.NodeID = strings.TrimSpace(string(nodeIdBytes))
	}

	return userConfig, nil
}

func parseConfigLine(line string, configKey string, data map[string]string) {
	if !strings.HasPrefix(line, configKey) {
		return
	}
	parts := strings.SplitN(line, "=", 2)
	if len(parts) == 2 {
		data[configKey] = strings.TrimSpace(parts[1])
	}
}
