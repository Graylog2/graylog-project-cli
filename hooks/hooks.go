package hooks

import (
	"fmt"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
)

const path = "hooks"

func Run(name string, noop bool) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.Wrap(err, "scripts directory doesn't exist")
	}

	// Supported scripts are named after the command that is executed. (e.g. bootstrap.sh)
	// To support custom scripts, we also look for "<name>-custom.sh".
	scripts := []string{
		filepath.Join(path, fmt.Sprintf("%s.sh", name)),
		filepath.Join(path, fmt.Sprintf("%s-custom.sh", name)),
	}

	for _, script := range scripts {
		if _, err := os.Stat(script); os.IsNotExist(err) {
			logger.Debug("script %s doesn't exist, skipping", script)
			continue
		}

		if noop {
			logger.Info("Would execute script %s", script)
			continue
		}

		customEnv := []string{"GPC_COMMAND=" + name, "GPC_HOOK_SCRIPT=" + script, "GPC_BIN=" + os.Args[0]}

		cmd := exec.Command(script)
		cmd.Env = append(os.Environ(), customEnv...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		logger.Debug("Running script %s for hook %s with environment: %v", script, name, customEnv)

		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "couldn't run hook script %s", script)
		}
	}

	return nil
}

func Simulate(name string) error {
	return nil
}
