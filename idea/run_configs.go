package idea

import (
	"bytes"
	"crypto/sha256"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/pomparse"
	"github.com/google/renameio/v2"
	"github.com/samber/lo"
	"github.com/subosito/gotenv"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
)

var DefaultRootPassword = "admin"
var DefaultInstanceCounts = map[string]int{
	"server":    2,
	"data-node": 2,
}

var runConfigTemplateDir = filepath.Join(".config", "idea", "templates", "run-configurations")
var configFile = filepath.Join(".config", "idea", "config.yml")

const runConfigDir = ".run"
const runConfigSuffix = ".run.xml"
const runConfigTemplateSuffix = ".run.xml.template"
const envFileSuffix = ".env.template"
const generatedFilePrefix = "project-generated-"

// We use a static password secret to ensure that different setups can use the same database.
const staticPasswordSecret = "hCXFTrzZFF88gnVon2fSV6WmAoQANRUqsYFTRbac8WStamVeJkjTXSykWv6FiXDbTYQQnvdTn59iALnkiT6m93BfhDju9Uqh"

type RunConfig struct {
	Workdir      string         `mapstructure:"workdir"`
	Instances    map[string]int `mapstructure:"instances"`
	Force        bool           `mapstructure:"force"`
	EnvFile      bool           `mapstructure:"env-file"`
	RootPassword string         `mapstructure:"root-password"`
}

type ConfigData struct {
	DataDirectories           map[string][]string       `yaml:"data-directories"`
	CompoundConfigs           map[string]CompoundConfig `yaml:"compound-configs"`
	DataNodePomPath           []string                  `yaml:"data-node-pom-path"`
	OpenSearchVersion         string                    `yaml:"opensearch-version"`
	OpenSearchVersionProperty string                    `yaml:"opensearch-version-property"`
}

type CompoundConfig struct {
	Name          string   `yaml:"name"`
	InstanceTypes []string `yaml:"instance-types"`
}

type templateData struct {
	ConfigName        string
	InstanceType      string
	InstanceNumber    int
	UseEnvFile        bool
	Env               map[string]string
	PortOffset        int
	PasswordSecret    string
	RootPasswordHash  string
	DataDir           string
	IsLeaderNode      bool
	OpenSearchVersion string
}

var mathTemplateFuncs = map[string]any{
	"add": func(a, b int) int {
		return a + b
	},
	"sub": func(a, b int) int {
		return a - b
	},
}

// Template string for an IntelliJ compound run configuration entry.
var compoundTemplate = `<component name="ProjectRunConfigurationManager">
  <configuration default="false" name="{{ .Name }}" type="CompoundRunConfigurationType">
{{- range .ToRun }}
    <toRun name="{{ .Name }}" type="{{ .Type }}" />
{{- end }}
    <method v="2" />
  </configuration>
</component>
`

type RunConfigEntry struct {
	Name                string
	InstanceType        string
	InstanceNumber      int
	PortOffset          int
	Template            *template.Template
	RenderedTemplate    bytes.Buffer
	EnvTemplate         *template.Template
	RenderedEnvTemplate bytes.Buffer
	DataDirectories     []string
	Filename            string
	EnvFilename         string
	DataDir             string
}

// XMLRunConfig is used to parse the component type out of run configuration files.
type XMLRunConfig struct {
	XMLName       xml.Name `xml:"component"`
	Configuration struct {
		Type string `xml:"type,attr"`
	} `xml:"configuration"`
}

type CompoundToRun struct {
	Name string
	Type string
}

func CreateRunConfigurations(config RunConfig) error {
	rcDir := filepath.Join(config.Workdir, runConfigDir)
	tmplDir := filepath.Join(config.Workdir, runConfigTemplateDir)

	if _, err := os.Stat(tmplDir); os.IsNotExist(err) {
		return fmt.Errorf("template directory %q doesn't exist; update your repository", tmplDir)
	}

	if err := os.MkdirAll(rcDir, 0755); err != nil {
		return fmt.Errorf("couldn't create run configurations directory: %w", err)
	}

	templates, err := findRunConfigTemplates(tmplDir)
	if err != nil {
		return err
	}

	invalidParams := make([]string, 0)
	for name := range config.Instances {
		if _, found := templates[name]; !found {
			invalidParams = append(invalidParams, name)
		}
	}
	if len(invalidParams) > 0 {
		return fmt.Errorf("invalid instance count parameter(s): %s (available: %s)",
			strings.Join(invalidParams, ", "), strings.Join(slices.Sorted(maps.Keys(templates)), ", "))
	}

	configData, err := parseConfigFile(filepath.Join(config.Workdir, configFile))
	if err != nil {
		return err
	}

	opensearchVersion, err := detectOpenSearchVersion(config, configData)
	if err != nil {
		return err
	}

	entries := make([]RunConfigEntry, 0)

	// Build all run configuration entries in memory
	for instanceType, tmpl := range templates {
		totalCount := getInstanceCount(config, instanceType)
		for i := range totalCount {
			num := i + 1

			envTmpl, err := findEnvFileTemplate(filepath.Join(config.Workdir, runConfigTemplateDir), instanceType)
			if err != nil {
				return err
			}

			entry := RunConfigEntry{
				Name:            generateConfigName(instanceType, num, totalCount),
				InstanceType:    instanceType,
				InstanceNumber:  num,
				PortOffset:      i,
				Template:        tmpl,
				EnvTemplate:     envTmpl,
				DataDirectories: configData.DataDirectories[instanceType],
				Filename:        fmt.Sprintf("%s%s-%d%s", generatedFilePrefix, instanceType, num, runConfigSuffix),
				EnvFilename:     fmt.Sprintf(".env.%s-%d", instanceType, num),
				DataDir:         filepath.Join("data", fmt.Sprintf("%s-%d", instanceType, num)),
			}

			data := templateData{
				ConfigName:        entry.Name,
				InstanceType:      entry.InstanceType,
				InstanceNumber:    entry.InstanceNumber,
				PortOffset:        entry.PortOffset,
				UseEnvFile:        config.EnvFile,
				PasswordSecret:    staticPasswordSecret,
				RootPasswordHash:  fmt.Sprintf("%x", sha256.Sum256([]byte(config.RootPassword))),
				DataDir:           entry.DataDir,
				IsLeaderNode:      entry.InstanceNumber == 1, // Make the first node the leader node
				OpenSearchVersion: opensearchVersion,
			}

			if entry.EnvTemplate != nil {
				if err := entry.EnvTemplate.Execute(&entry.RenderedEnvTemplate, data); err != nil {
					return fmt.Errorf("couldn't render env-file template: %w", err)
				}
			}

			data.Env = gotenv.Parse(bytes.NewReader(entry.RenderedEnvTemplate.Bytes()))

			if err := entry.Template.Execute(&entry.RenderedTemplate, data); err != nil {
				return fmt.Errorf("couldn't compile template: %w", err)
			}

			entries = append(entries, entry)
		}
	}

	// Write all run configuration entries to the file system
	for _, entry := range entries {
		if err := writeEntryFiles(rcDir, config, entry); err != nil {
			return err
		}
	}

	// Write all compound run configurations
	for name, cfg := range configData.CompoundConfigs {
		filteredEntries := lo.Filter(entries, func(entry RunConfigEntry, index int) bool {
			return slices.Contains(cfg.InstanceTypes, entry.InstanceType)
		})

		toRun := make([]CompoundToRun, 0)

		for _, entry := range filteredEntries {
			var xrc XMLRunConfig
			if err := xml.NewDecoder(bytes.NewReader(entry.RenderedTemplate.Bytes())).Decode(&xrc); err != nil {
				return fmt.Errorf("couldn't parse rendered run configuration %q: %w", entry.Filename, err)
			}

			toRun = append(toRun, CompoundToRun{Name: entry.Name, Type: xrc.Configuration.Type})
		}

		tmpl, err := template.New(cfg.Name).Parse(compoundTemplate)
		if err != nil {
			return fmt.Errorf("couldn't parse compunt template %q: %w", cfg.Name, err)
		}

		compoundFilename := fmt.Sprintf("%scompound-%s%s", generatedFilePrefix, name, runConfigSuffix)

		file, err := renameio.TempFile("", filepath.Join(config.Workdir, runConfigDir, compoundFilename))
		if err != nil {
			return fmt.Errorf("couldn't create temp file for %q: %w", compoundFilename, err)
		}

		if err := tmpl.Execute(file, map[string]any{"Name": cfg.Name, "ToRun": toRun}); err != nil {
			return err
		}

		if err := file.CloseAtomicallyReplace(); err != nil {
			return fmt.Errorf("couldn't close temp file for %q: %w", compoundFilename, err)
		}

		logger.Info("Created compound configuration: %s", compoundFilename)
	}

	return nil
}

// The GRAYLOG_DATANODE_OPENSEARCH_LOCATION setting needs the OpenSearch version. Let's try to find it
// in the data-node/pom.xml of the server repo. It's ugly. ¯\_(ツ)_/¯
func detectOpenSearchVersion(config RunConfig, configData *ConfigData) (string, error) {
	opensearchVersion := ""
	dataNodePomPath := filepath.Join(append([]string{config.Workdir}, configData.DataNodePomPath...)...)
	pom, err := pomparse.ParsePomE(dataNodePomPath)
	if err == nil {
		// First check the property from the pom.xml file
		opensearchVersion = pom.PropertiesMap()[configData.OpenSearchVersionProperty]
	}
	if opensearchVersion == "" {
		// Use the OpenSearch version setting from the config file as fallback
		opensearchVersion = configData.OpenSearchVersion
	}
	if opensearchVersion == "" {
		// We can't continue without an OpenSearch version
		return "", errors.Join(
			fmt.Errorf("couldn't find the OpenSearch version in the %q property of the %q file and the fallback opensearch-version is not set in %q",
				configData.OpenSearchVersionProperty, dataNodePomPath, configFile),
			err,
		)
	}

	return opensearchVersion, nil
}

func parseConfigFile(path string) (*ConfigData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't open config %q: %w", path, err)
	}
	//goland:noinspection ALL
	defer f.Close()

	var value ConfigData
	if err := yaml.NewDecoder(f).Decode(&value); err != nil {
		return nil, fmt.Errorf("couldn't parse idea config %q: %w", path, err)
	}

	return &value, nil
}

func getInstanceCount(config RunConfig, instanceType string) int {
	count, ok := config.Instances[instanceType]
	if !ok {
		defaultCount, defaultOk := DefaultInstanceCounts[instanceType]
		if defaultOk {
			return defaultCount
		} else {
			return 1
		}
	}
	return count
}

func writeEntryFiles(configDir string, config RunConfig, entry RunConfigEntry) error {
	if _, err := os.Stat(filepath.Join(configDir, entry.Filename)); !os.IsNotExist(err) && !config.Force {
		logger.Info("Skipping existing run configuration: %s", entry.Filename)
		return nil
	}

	if err := renameio.WriteFile(filepath.Join(configDir, entry.Filename), entry.RenderedTemplate.Bytes(), 0600); err != nil {
		return fmt.Errorf("couldn't write file %q: %w", entry.Filename, err)
	}

	logger.Info("Created run configuration: %s", entry.Filename)

	for _, dir := range entry.DataDirectories {
		dirToCreate := filepath.Join(entry.DataDir, dir)
		if _, err := os.Stat(dirToCreate); os.IsNotExist(err) {
			if err := os.MkdirAll(dirToCreate, 0755); err != nil {
				return fmt.Errorf("couldn't create data dir %q: %w", dir, err)
			}
			logger.Info("Created data directory: %s", dirToCreate)
		}
	}

	if config.EnvFile {
		if err := renameio.WriteFile(filepath.Join(config.Workdir, entry.EnvFilename), entry.RenderedEnvTemplate.Bytes(), 0600); err != nil {
			return fmt.Errorf("couldn't write file %q: %w", entry.EnvFilename, err)
		}

		logger.Info("Created run env file: %s", entry.EnvFilename)
	}

	return nil
}

func generateConfigName(instanceType string, num int, total int) string {
	// Data-Node -> Data Node
	configName := strings.ReplaceAll(cases.Title(language.English).String(instanceType), "-", " ")

	if total > 1 {
		return fmt.Sprintf("%s %d", configName, num)
	} else {
		return configName
	}
}

func findEnvFileTemplate(templateDir string, name string) (*template.Template, error) {
	data, err := os.ReadFile(filepath.Join(templateDir, fmt.Sprintf("%s%s", name, envFileSuffix)))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("couldn't read env-file template for %q: %w", name, err)
	}
	tmpl, err := template.New(name).Option("missingkey=error").Funcs(mathTemplateFuncs).Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("couldn't parse template %q: %w", name, err)
	}
	return tmpl, nil
}

func findRunConfigTemplates(templateDir string) (map[string]*template.Template, error) {
	templates := make(map[string]*template.Template)

	err := filepath.WalkDir(templateDir, func(path string, entry fs.DirEntry, err error) error {
		if entry.IsDir() {
			return nil
		}
		if !strings.HasSuffix(entry.Name(), runConfigTemplateSuffix) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("couldn't read template: %w", err)
		}

		name := strings.TrimSuffix(entry.Name(), runConfigTemplateSuffix)

		tmpl, err := template.New(name).Option("missingkey=error").Funcs(mathTemplateFuncs).Parse(string(data))
		if err != nil {
			return fmt.Errorf("couldn't parse template %q: %w", path, err)
		}

		templates[name] = tmpl

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't walk template dir: %w", err)
	}

	return templates, nil
}
