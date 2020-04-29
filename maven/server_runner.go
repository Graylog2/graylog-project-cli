package maven

import (
	"fmt"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/go-cmd/cmd"
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"syscall"
)

func runServer(workDir string, bin string, args ...string) error {
	// Use github.com/go-cmd/cmd here instead of os/exec to ensure proper signal handling
	runCmd := cmd.NewCmdOptions(cmd.Options{Buffered: false, Streaming: true}, bin, args...)
	runCmd.Dir = workDir

	// Print STDOUT and STDERR lines streaming from Cmd
	doneChan := make(chan struct{})

	go monitorSignals(runCmd)
	go handleProcessOutput(doneChan, runCmd)

	finalStatus := <-runCmd.Start() // Wait for command to be finished
	<-doneChan                      // Wait for output printer to be finished

	if finalStatus.Error != nil {
		return errors.Wrapf(finalStatus.Error, "error running server: %v", finalStatus)
	}

	return nil
}

func monitorSignals(runCmd *cmd.Cmd) {
	handler := make(chan os.Signal, 1)
	signal.Notify(handler, syscall.SIGALRM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	for sig := range handler {
		switch sig {
		case syscall.SIGINT:
			fmt.Println(" | ctrl-c detected")
			fallthrough
		default:
			if err := runCmd.Stop(); err != nil {
				logger.Error("error stopping command: %v", err)
			}
		}
	}
}

func handleProcessOutput(doneChan chan struct{}, runCmd *cmd.Cmd) {
	defer close(doneChan)
	// Done when both channels have been closed
	// https://dave.cheney.net/2013/04/30/curious-channels
	for runCmd.Stdout != nil || runCmd.Stderr != nil {
		select {
		case line, open := <-runCmd.Stdout:
			if !open {
				runCmd.Stdout = nil
				continue
			}
			fmt.Println(line)
		case line, open := <-runCmd.Stderr:
			if !open {
				runCmd.Stderr = nil
				continue
			}
			_, _ = fmt.Fprintln(os.Stderr, line)
		}
	}
}
