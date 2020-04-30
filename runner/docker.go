package runner

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

// Exec a script command. This will install a SIGINT signal handler hat swallows the signal so we wait
// for the script process to finish. (which also gets a SIGINT because it's running in the same process group)
func execRunnerScript(config Config, env []string) error {
	// Command "dev:server" becomes "bin/dev-server.sh"
	scriptName := fmt.Sprintf("./bin/%s.sh", strings.ReplaceAll(config.Command, ":", "-"))
	command := exec.Command(scriptName)
	command.Dir = config.RunnerRoot
	command.Env = append(os.Environ(), env...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	go monitorCommandExecSignals()

	if err := command.Run(); err != nil {
		return errors.Wrapf(err, "couldn't run script %s", strings.Join(command.Args, " "))
	}

	return nil
}

func monitorCommandExecSignals() {
	handler := make(chan os.Signal, 1)
	signal.Notify(handler, syscall.SIGINT)

	for sig := range handler {
		switch sig {
		case syscall.SIGINT:
			fmt.Println(" | ctrl-c detected")
			fallthrough
		default:
			// We assume that this is function is only used to monitor signals when running a blocking exec call.
			// So we don't have to do anything here and just continue the program flow.
		}
	}
}
