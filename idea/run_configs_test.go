package idea

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestCreateRunConfigurations(t *testing.T) {
	workdir := t.TempDir()

	config := RunConfig{
		Workdir:      workdir,
		Instances:    DefaultInstanceCounts,
		Force:        false,
		EnvFile:      false,
		RootPassword: "test",
	}

	// No template files yet, should fail
	require.ErrorContains(t, CreateRunConfigurations(config), "update your repository")

	// Copy the templates
	require.NoError(t, copyDir("testdata", workdir))

	// Should not fail now
	require.NoError(t, CreateRunConfigurations(config))

	assert.DirExists(t, filepath.Join(workdir, ".run"))

	for _, instance := range []string{
		"compound-all", "compound-data-nodes", "compound-servers",
		"data-node-1", "data-node-2", "server-1", "server-2", "web-1",
	} {
		assert.FileExists(t, filepath.Join(workdir, ".run", "project-generated-"+instance+".run.xml"))
	}

	for _, file := range []string{"data-node-1", "data-node-2", "server-1", "server-2", "web-1"} {
		assert.NoFileExists(t, filepath.Join(workdir, ".env."+file))
	}

	assertCompoundRunConfigFile(t, filepath.Join(workdir, ".run", "project-generated-compound-all.run.xml"), "All Nodes", map[string]string{
		"Server 1":    "Application",
		"Server 2":    "Application",
		"Data Node 1": "Application",
		"Data Node 2": "Application",
		"Web":         "js.build_tools.npm",
	})
	assertCompoundRunConfigFile(t, filepath.Join(workdir, ".run", "project-generated-compound-servers.run.xml"), "Servers", map[string]string{
		"Server 1": "Application",
		"Server 2": "Application",
	})
	assertCompoundRunConfigFile(t, filepath.Join(workdir, ".run", "project-generated-compound-data-nodes.run.xml"), "Data Nodes", map[string]string{
		"Data Node 1": "Application",
		"Data Node 2": "Application",
	})

	assertFile(t, assert.Contains, filepath.Join(workdir, ".run", "project-generated-web-1.run.xml"),
		`<configuration default="false" name="Web" type="js.build_tools.npm">`)

	for _, instance := range []string{"server-1", "server-2"} {
		path := filepath.Join(workdir, ".run", fmt.Sprintf("project-generated-%s.run.xml", instance))
		num := getInstanceNumber(t, instance)
		portOffset := num - 1

		assertFile(t, assert.Contains, path, fmt.Sprintf(`<configuration default="false" name="Server %d" type="Application" factoryName="Application" singleton="true">`, num))
		assertFile(t, assert.NotContains, path, fmt.Sprintf(`PATH="$PROJECT_DIR$/.env.%s`, instance))

		assertRunConfigFileEnv(t, assert.Contains, path, map[string]any{
			"GRAYLOG_NODE_ID_FILE":        filepath.Join("data", instance, "node-id"),
			"GRAYLOG_DATA_DIR":            filepath.Join("data", instance),
			"GRAYLOG_MESSAGE_JOURNAL_DIR": filepath.Join("data", instance, "journal"),
			"GRAYLOG_IS_LEADER":           num == 1, // First node should be the leader
			"GRAYLOG_HTTP_BIND_ADDRESS":   fmt.Sprintf("127.0.0.1:%d", 9000+portOffset),
			"GRAYLOG_PASSWORD_SECRET":     "hCXFTrzZFF88gnVon2fSV6WmAoQANRUqsYFTRbac8WStamVeJkjTXSykWv6FiXDbTYQQnvdTn59iALnkiT6m93BfhDju9Uqh",
			"GRAYLOG_ROOT_PASSWORD_SHA2":  "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		})
	}
	for _, instance := range []string{"data-node-1", "data-node-2"} {
		path := filepath.Join(workdir, ".run", fmt.Sprintf("project-generated-%s.run.xml", instance))
		num := getInstanceNumber(t, instance)
		portOffset := num - 1

		assertFile(t, assert.Contains, path, fmt.Sprintf(`<configuration default="false" name="Data Node %d" type="Application" factoryName="Application" singleton="true">`, num))
		assertFile(t, assert.NotContains, path, fmt.Sprintf(`PATH="$PROJECT_DIR$/.env.%s`, instance))

		assertRunConfigFileEnv(t, assert.Contains, path, map[string]any{
			"GRAYLOG_DATANODE_PASSWORD_SECRET":           "hCXFTrzZFF88gnVon2fSV6WmAoQANRUqsYFTRbac8WStamVeJkjTXSykWv6FiXDbTYQQnvdTn59iALnkiT6m93BfhDju9Uqh",
			"GRAYLOG_DATANODE_NODE_ID_FILE":              filepath.Join("data", instance, "node-id"),
			"GRAYLOG_DATANODE_CONFIG_LOCATION":           filepath.Join("data", instance, "config"),
			"GRAYLOG_DATANODE_NATIVE_LIB_DIR":            filepath.Join("data", instance, "native_libs"),
			"GRAYLOG_DATANODE_DATANODE_HTTP_PORT":        8999 - portOffset,
			"GRAYLOG_DATANODE_OPENSEARCH_HTTP_PORT":      9200 + portOffset,
			"GRAYLOG_DATANODE_OPENSEARCH_TRANSPORT_PORT": 9300 + portOffset,
		})
	}
}

func TestCreateRunConfigurationsWithEnv(t *testing.T) {
	workdir := t.TempDir()

	config := RunConfig{
		Workdir:      workdir,
		Instances:    DefaultInstanceCounts,
		Force:        false,
		EnvFile:      true,
		RootPassword: "test",
	}

	// No template files yet, should fail
	require.ErrorContains(t, CreateRunConfigurations(config), "update your repository")

	// Copy the templates
	require.NoError(t, copyDir("testdata", workdir))

	// Should not fail now
	require.NoError(t, CreateRunConfigurations(config))

	assert.DirExists(t, filepath.Join(workdir, ".run"))

	for _, file := range []string{
		"compound-all", "compound-data-nodes", "compound-servers",
		"data-node-1", "data-node-2", "server-1", "server-2", "web-1",
	} {
		assert.FileExists(t, filepath.Join(workdir, ".run", "project-generated-"+file+".run.xml"))
	}

	for _, file := range []string{"data-node-1", "data-node-2", "server-1", "server-2", "web-1"} {
		assert.FileExists(t, filepath.Join(workdir, ".env."+file))
	}

	assertCompoundRunConfigFile(t, filepath.Join(workdir, ".run", "project-generated-compound-all.run.xml"), "All Nodes", map[string]string{
		"Server 1":    "Application",
		"Server 2":    "Application",
		"Data Node 1": "Application",
		"Data Node 2": "Application",
		"Web":         "js.build_tools.npm",
	})
	assertCompoundRunConfigFile(t, filepath.Join(workdir, ".run", "project-generated-compound-servers.run.xml"), "Servers", map[string]string{
		"Server 1": "Application",
		"Server 2": "Application",
	})
	assertCompoundRunConfigFile(t, filepath.Join(workdir, ".run", "project-generated-compound-data-nodes.run.xml"), "Data Nodes", map[string]string{
		"Data Node 1": "Application",
		"Data Node 2": "Application",
	})

	assertFile(t, assert.Contains, filepath.Join(workdir, ".run", "project-generated-web-1.run.xml"),
		`<configuration default="false" name="Web" type="js.build_tools.npm">`)

	for _, instance := range []string{"server-1", "server-2"} {
		path := filepath.Join(workdir, ".run", fmt.Sprintf("project-generated-%s.run.xml", instance))
		envPath := filepath.Join(workdir, fmt.Sprintf(".env.%s", instance))
		num := getInstanceNumber(t, instance)
		portOffset := num - 1

		assertFile(t, assert.Contains, path, fmt.Sprintf(`<configuration default="false" name="Server %d" type="Application" factoryName="Application" singleton="true">`, num))
		assertFile(t, assert.Contains, path, fmt.Sprintf(`PATH="$PROJECT_DIR$/.env.%s`, instance))

		expectedData := map[string]any{
			"GRAYLOG_NODE_ID_FILE":        filepath.Join("data", instance, "node-id"),
			"GRAYLOG_DATA_DIR":            filepath.Join("data", instance),
			"GRAYLOG_MESSAGE_JOURNAL_DIR": filepath.Join("data", instance, "journal"),
			"GRAYLOG_IS_LEADER":           num == 1, // First node should be the leader
			"GRAYLOG_HTTP_BIND_ADDRESS":   fmt.Sprintf("127.0.0.1:%d", 9000+portOffset),
			"GRAYLOG_PASSWORD_SECRET":     "hCXFTrzZFF88gnVon2fSV6WmAoQANRUqsYFTRbac8WStamVeJkjTXSykWv6FiXDbTYQQnvdTn59iALnkiT6m93BfhDju9Uqh",
			"GRAYLOG_ROOT_PASSWORD_SHA2":  "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		}
		assertRunConfigFileEnv(t, assert.NotContains, path, expectedData)
		assertEnvFile(t, assert.Contains, envPath, expectedData)
	}
	for _, instance := range []string{"data-node-1", "data-node-2"} {
		path := filepath.Join(workdir, ".run", fmt.Sprintf("project-generated-%s.run.xml", instance))
		envPath := filepath.Join(workdir, fmt.Sprintf(".env.%s", instance))
		num := getInstanceNumber(t, instance)
		portOffset := num - 1

		assertFile(t, assert.Contains, path, fmt.Sprintf(`<configuration default="false" name="Data Node %d" type="Application" factoryName="Application" singleton="true">`, num))
		assertFile(t, assert.Contains, path, fmt.Sprintf(`PATH="$PROJECT_DIR$/.env.%s`, instance))

		expectedData := map[string]any{
			"GRAYLOG_DATANODE_NODE_ID_FILE":              filepath.Join("data", instance, "node-id"),
			"GRAYLOG_DATANODE_CONFIG_LOCATION":           filepath.Join("data", instance, "config"),
			"GRAYLOG_DATANODE_NATIVE_LIB_DIR":            filepath.Join("data", instance, "native_libs"),
			"GRAYLOG_DATANODE_DATANODE_HTTP_PORT":        8999 - portOffset,
			"GRAYLOG_DATANODE_OPENSEARCH_HTTP_PORT":      9200 + portOffset,
			"GRAYLOG_DATANODE_OPENSEARCH_TRANSPORT_PORT": 9300 + portOffset,
		}
		assertRunConfigFileEnv(t, assert.NotContains, path, expectedData)
		assertEnvFile(t, assert.Contains, envPath, expectedData)
	}
}

func getInstanceNumber(t *testing.T, name string) int {
	parts := strings.SplitAfter(name, "-")
	num, err := strconv.Atoi(parts[len(parts)-1])
	require.NoError(t, err)
	return num
}

func assertCompoundRunConfigFile(t *testing.T, path string, title string, values map[string]string) {
	assertFile(t, assert.Contains, path, fmt.Sprintf(`<configuration default="false" name="%s" type="CompoundRunConfigurationType">`, title), "wrong title")

	for key, value := range values {
		assertFile(t, assert.Contains, path, fmt.Sprintf(`<toRun name="%s" type="%s" />`, key, value), fmt.Sprintf("compound file %q - %s=%s", path, key, value))
	}
}

func assertRunConfigFileEnv(t *testing.T, check func(assert.TestingT, interface{}, interface{}, ...interface{}) bool, path string, values map[string]any) {
	for key, value := range values {
		assertFile(t, check, path, fmt.Sprintf(`<env name="%s" value="%v" />`, key, value), fmt.Sprintf("run config file %q - %s=%s", path, key, value))
	}
}

func assertEnvFile(t *testing.T, check func(assert.TestingT, interface{}, interface{}, ...interface{}) bool, path string, values map[string]any) {
	for key, value := range values {
		assertFile(t, check, path, fmt.Sprintf("%s=%v", key, value), fmt.Sprintf("env file %q - %s=%s", path, key, value))
	}
}

func assertFile(t *testing.T, check func(assert.TestingT, interface{}, interface{}, ...interface{}) bool, path string, needle string, message ...string) {
	require.FileExists(t, path)
	buf, err := os.ReadFile(path)
	require.NoError(t, err)
	check(t, string(buf), needle, message)
}

func copyDir(srcPath, dstPath string) error {
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == srcPath {
			return nil
		}

		dstFilePath := filepath.Join(dstPath, path[len(srcPath):])

		if info.IsDir() {
			if err := os.MkdirAll(dstFilePath, info.Mode()); err != nil {
				return err
			}
			return nil
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		//goland:noinspection ALL
		defer srcFile.Close()

		dstFile, err := os.Create(dstFilePath)
		if err != nil {
			return err
		}
		//goland:noinspection ALL
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return err
		}

		return os.Chmod(dstFilePath, info.Mode())
	}

	return filepath.Walk(srcPath, walkFn)
}
