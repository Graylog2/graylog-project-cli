package exec

import (
	"bytes"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/utils"
	"os/exec"
	"path/filepath"
	"strings"
)

type ExecCommandOutput struct {
	Stdout bytes.Buffer
	Stderr bytes.Buffer
}

func ExecCommandInPath(path string, args ...string) (ExecCommandOutput, error) {
	defer utils.Chdir(utils.GetCwd())

	logger.Debug("[%v] %s]", filepath.Base(path), strings.Join(args, " "))

	utils.Chdir(path)

	command := exec.Command("sh", "-c", strings.Join(args, " "))

	var output ExecCommandOutput

	command.Stdout = &output.Stdout
	command.Stderr = &output.Stderr

	err := command.Run()

	return output, err
}
