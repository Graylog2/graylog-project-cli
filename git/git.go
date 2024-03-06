package git

import (
	"bytes"
	"fmt"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"os/exec"
	"strings"
)

func Git(commands ...string) {
	git(true, commands...)
}

func GitE(commands ...string) (string, error) {
	out, err := gitE(true, commands...)
	return out, err
}

func SilentGit(commands ...string) {
	git(false, commands...)
}

func git(verbose bool, commands ...string) {
	var stderr bytes.Buffer

	if verbose {
		logger.ColorInfo(color.FgGreen, "    git %v", strings.Join(commands, " "))
	}

	command := exec.Command("git", commands...)
	command.Stderr = &stderr
	out, err := command.Output()
	if err != nil {
		logger.Error("Git stderr: %v", string(stderr.Bytes()))
		logger.Fatal("Error executing: git %s (%v)", strings.Join(commands, " "), err)
	}

	logOutputBuffer(stderr.Bytes())
	logOutputBuffer(out)
}

func gitE(verbose bool, commands ...string) (string, error) {
	var outBuf bytes.Buffer

	if verbose {
		logger.ColorInfo(color.FgGreen, "    git %v", strings.Join(commands, " "))
	}

	command := exec.Command("git", commands...)
	command.Stderr = &outBuf
	command.Stdout = &outBuf
	if err := command.Run(); err != nil {
		return outBuf.String(), fmt.Errorf("couldn't execute \"git %s\": %w", strings.Join(commands, " "), err)
	}

	return outBuf.String(), nil
}

func GitValue(commands ...string) string {
	var stderr bytes.Buffer

	command := exec.Command("git", commands...)
	command.Stderr = &stderr
	out, err := command.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			logger.Info("Git stderr: %v", string(ee.Stderr))
		}
		logger.Fatal("Error executing: git %v (%v)", commands, err)
	}

	logOutputBuffer(stderr.Bytes())

	return strings.TrimSuffix(string(out), "\n")
}

func GitValueE(commands ...string) (string, error) {
	command := exec.Command("git", commands...)
	out, err := command.Output()
	if err != nil {
		return "", errors.Wrapf(err, "error executing: %s %s", command.Path, strings.Join(command.Args, " "))
	}

	return strings.TrimSpace(string(out)), nil
}

func HasLocalBranch(branch string) bool {
	// Command exits with 1 if the local branch doesn't exist
	_, err := GitValueE("rev-parse", "--verify", "--quiet", branch)
	if err != nil {
		return false
	}
	return true
}

func ToplevelPath() (string, error) {
	path, err := GitValueE("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return path, nil
}

func ExecInPath(path string, commands ...string) error {
	return utils.InDirectoryE(path, func() error {
		return Exec(commands...)
	})
}

func Exec(commands ...string) error {
	var output bytes.Buffer

	logger.ColorInfo(color.FgGreen, "    git %v", strings.Join(commands, " "))
	command := exec.Command("git", commands...)
	command.Stderr = &output
	command.Stdout = &output

	if err := command.Run(); err != nil {
		logOutputBufferWithColor(output.Bytes(), color.FgRed)
		return logger.NewLoggableError(
			err,
			fmt.Sprintf(
				`couldn't execute "%s" in "%s"`,
				fmt.Sprintf("git %s", strings.Join(commands, " ")),
				utils.GetCwd(),
			),
			strings.Split(output.String(), "\n"),
		)
	}

	logOutputBuffer(output.Bytes())

	return nil
}

func logOutputBuffer(buf []byte) {
	logOutputBufferWithColor(buf, color.FgYellow)
}

func logOutputBufferWithColor(buf []byte, c color.Attribute) {
	for _, s := range strings.Split(string(buf), "\n") {
		if len(strings.TrimSpace(s)) > 0 {
			logger.ColorInfo(c, "      %v", s)
		}
	}
}
