package git

import (
	"bytes"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"os/exec"
	"strings"
)

func Git(commands ...string) {
	git(true, commands...)
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

func GitErrOk(commands ...string) {
	var stderr bytes.Buffer

	logger.ColorInfo(color.FgGreen, "    git %v", strings.Join(commands, " "))
	command := exec.Command("git", commands...)
	command.Stderr = &stderr
	out, err := command.Output()
	if err != nil {
		logOutputBufferWithColor(stderr.Bytes(), color.FgRed)
		logOutputBufferWithColor(out, color.FgRed)
		return
	}

	logOutputBuffer(stderr.Bytes())
	logOutputBuffer(out)
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
