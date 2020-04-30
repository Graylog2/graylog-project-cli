package idea

import (
	"encoding/xml"
	"github.com/Graylog2/graylog-project-cli/logger"
	p "github.com/Graylog2/graylog-project-cli/project"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var WebBuildExcludeURLRE = regexp.MustCompile(`file://.+/target/web/build`)
var runConfigurationDir = filepath.Join(".idea", "runConfigurations")
var runConfigurations = map[string]string{
	"Graylog_Server.xml": `<component name="ProjectRunConfigurationManager">
  <configuration default="false" name="Graylog Server" type="Application" factoryName="Application" singleton="true">
    <envs>
      <env name="DEVELOPMENT" value="true" />
    </envs>
    <option name="MAIN_CLASS_NAME" value="org.graylog2.bootstrap.Main" />
    <module name="runner" />
    <option name="PROGRAM_PARAMETERS" value="server -f graylog.conf -np --local" />
    <option name="VM_PARAMETERS" value="-Djava.library.path=lib/sigar-1.6.4 -Xmx1g -XX:NewRatio=1 -server -XX:+ResizeTLAB -XX:+UseConcMarkSweepGC -XX:+CMSConcurrentMTEnabled -XX:+CMSClassUnloadingEnabled -XX:-OmitStackTraceInFastThrow -XX:+PreserveFramePointer -XX:+UnlockDiagnosticVMOptions -XX:+DebugNonSafepoints -Dio.netty.leakDetection.level=paranoid" />
    <option name="WORKING_DIRECTORY" value="$PROJECT_DIR$/../graylog-project-repos/graylog2-server" />
    <method v="2">
      <option name="Make" enabled="true" />
    </method>
  </configuration>
</component>
`,

	"Web_Devserver.xml": `<component name="ProjectRunConfigurationManager">
  <configuration default="false" name="Web Devserver" type="js.build_tools.npm">
    <package-json value="$PROJECT_DIR$/../graylog-project-repos/graylog2-server/graylog2-web-interface/package.json" />
    <command value="run" />
    <scripts>
      <script value="start" />
    </scripts>
    <node-interpreter value="project" />
    <node-options value="--max-old-space-size=3192" />
    <envs>
      <env name="_NODE_OPTIONS" value="--max-old-space-size=3192" />
      <env name="_GRAYLOG_HTTP_PUBLISH_URI" value="https://127.0.0.1:9000/api/" />
    </envs>
    <method v="2" />
  </configuration>
</component>
`,
}

type IMLModule struct {
	SourceFolders  []IMLSourceFolder  `xml:"component>content>sourceFolder"`
	ExcludeFolders []IMLExcludeFolder `xml:"component>content>excludeFolder"`
}

type IMLSourceFolder struct {
	Url string `xml:"url,attr"`
}

type IMLExcludeFolder struct {
	Url string `xml:"url,attr"`
}

func Setup(project p.Project) {
	ensureRunConfigurations()

	p.ForEachSelectedModuleOrSubmodules(project, func(module p.Module) {
		ensureWebBuildExclude(module)
	})
}

func ensureRunConfigurations() {
	for filename, content := range runConfigurations {
		// We expect us to be in the graylog-project root directory
		runConfigurationFile := filepath.Join(runConfigurationDir, filename)
		if utils.FileExists(runConfigurationFile) {
			logger.Debug("Skipping existing %s file", runConfigurationFile)
			continue
		}

		if os.MkdirAll(runConfigurationDir, 0755) != nil {
			logger.Fatal("Couldn't create run configuration directory: %s")
		}

		tmpFile, err := ioutil.TempFile(runConfigurationDir, filename)
		if err != nil {
			logger.Fatal("Couldn't create temp file for %s: %v", filename, err)
		}

		if _, err := tmpFile.WriteString(content); err != nil {
			logger.Fatal("Couldn't write to temp file for %s: %v", filename, err)
		}

		if err := ioutil.WriteFile(runConfigurationFile, []byte(content), 0644); err != nil {
			logger.Fatal("Couldn't write run configuration file %s: %v", runConfigurationFile, err)
		}

		if err := tmpFile.Close(); err != nil {
			logger.Fatal("Couldn't close temp file for %s: %v", filename, err)
		}

		if err := os.Rename(tmpFile.Name(), runConfigurationFile); err != nil {
			logger.Fatal("Couldn't rename temp file %s to %s: %v", tmpFile.Name(), runConfigurationFile, err)
		}

		if err := os.Chmod(runConfigurationFile, 0644); err != nil {
			logger.Fatal("Couldn't change mode on file %s: %v", runConfigurationFile, err)
		}

		logger.Info("Created run configuration: %s", filename)
	}
}

func ensureWebBuildExclude(module p.Module) {
	files, err := findIMLFiles(module)
	if err != nil {
		return
	}

	for _, file := range *files {
		iml, err := parseIMLFile(file)
		if err != nil {
			logger.Fatal("Couldn't parse IML file", err)
		}

		excludesToAdd := make([]string, 0)

		for _, sf := range iml.SourceFolders {
			if WebBuildExcludeURLRE.MatchString(sf.Url) {
				hasExclude := false
				for _, ef := range iml.ExcludeFolders {
					if sf.Url == ef.Url {
						logger.Debug("Skipping existing exclude in module %s for: %s", module.Name, sf.Url)
						hasExclude = true
					}
				}

				if !hasExclude {
					excludesToAdd = append(excludesToAdd, sf.Url)
				}
			}
		}

		for _, url := range excludesToAdd {
			if addExcludeFolderToIML(file, url) != nil {
				logger.Fatal("Couldn't add %s as exclude folder in %s", url, file)
			}
		}
	}
}

func parseIMLFile(imlFile string) (*IMLModule, error) {
	bytes, err := ioutil.ReadFile(imlFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading IML file: %v", imlFile)
	}

	var imlModule IMLModule

	if err := xml.Unmarshal(bytes, &imlModule); err != nil {
		return nil, errors.Wrapf(err, "Unable to parse IML file: %v", imlFile)
	}

	return &imlModule, nil
}

func findIMLFiles(module p.Module) (*[]string, error) {
	configs := make([]string, 0)

	err := filepath.Walk(module.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip all directories except the root dir
		if info.IsDir() && path != module.Path {
			return filepath.SkipDir
		}

		if strings.HasSuffix(path, ".iml") {
			configs = append(configs, path)
		}

		return nil
	})

	return &configs, err
}

func addExcludeFolderToIML(imlFile string, url string) error {
	excludeFolder := `<excludeFolder url="` + url + `" />`

	logger.Info("Adding [%s] to file %s", excludeFolder, imlFile)

	buf, err := ioutil.ReadFile(imlFile)
	if err != nil {
		return errors.Wrapf(err, "Couldn't read IML file: %v", imlFile)
	}

	// I hope this only matches once per .iml file ;)
	re := regexp.MustCompile(" {4}</content>")

	newContent := re.ReplaceAllLiteralString(string(buf), "      "+excludeFolder+"\n    </content>")

	if err := ioutil.WriteFile(imlFile, []byte(newContent), 0); err != nil {
		return errors.Wrapf(err, "Couldn't write IML file %s", imlFile)
	}

	return nil
}
