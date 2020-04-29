package maven

import (
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
)

const DependencyPluginVersion = "3.1.2"

const relativeClassPathFile = "target/graylog-project-cli-classpath.txt"

type BuildConfig struct {
	Clean     bool
	SkipWeb   bool
	MavenPath string
}

func (bc BuildConfig) MavenBin() (string, error) {
	if bc.MavenPath != "" {
		return bc.MavenPath, nil
	}
	mavenBin, err := exec.LookPath("mvn")
	if err != nil {
		return "", errors.Wrap(err, "couldn't find maven executable. please make sure maven is installed.")
	}
	return mavenBin, nil
}

func BuildForRun(config BuildConfig) (string, error) {
	mavenBin, err := config.MavenBin()
	if err != nil {
		return "", err
	}
	logger.Debug("Maven binary path: %s", mavenBin)

	if info, err := os.Stat("manifests"); os.IsNotExist(err) || !info.IsDir() {
		return "", errors.Wrap(err, "this command only works inside the graylog-project root directory.")
	}

	classPathFile, err := filepath.Abs(filepath.Join("runner", relativeClassPathFile))
	if err != nil {
		return "", errors.Wrapf(err, "couldn't get absolute path for %s", relativeClassPathFile)
	}

	commandArguments := []string{
		"--fail-fast",
		"-Dmaven.javadoc.skip=true",
		"-DskipTests",
		"-Dforbiddenapis.skip=true",
		"-DincludeScope=compile ",
		"-Dmdep.outputFile=" + relativeClassPathFile,
	}
	if config.SkipWeb {
		commandArguments = append(commandArguments, "-Dskip.web.build")
	}
	if config.Clean {
		commandArguments = append(commandArguments, "clean")
	}
	commandArguments = append(
		commandArguments,
		"test", // Use the test goal (even though we skip tests) to avoid a test-jar error
		"org.apache.maven.plugins:maven-dependency-plugin:"+DependencyPluginVersion+":build-classpath",
	)

	command := exec.Command(mavenBin, commandArguments...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		return "", errors.Wrapf(err, "couldn't run maven build")
	}

	return classPathFile, nil
}
