package runner

import (
	"fmt"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/go-cmd/cmd"
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"syscall"
)

func dockerCompose(workDir string, subCommand string, args ...string) error {
	// Depending on the command, this should compose (ha!) different docker-compose.yml files
	// See: https://docs.docker.com/compose/extends/
	//
	// Example: docker-compose -f docker-compose.services.yml -f docker-compose.dev-server.yml
	//          docker-compose -f docker-compose.services.yml
	// Use github.com/go-cmd/cmd here instead of os/exec to ensure proper signal handling

	command := cmd.NewCmdOptions(
		cmd.Options{Buffered: false, Streaming: true},
		"docker-compose",
		append([]string{subCommand}, args...)...,
	)
	command.Dir = workDir

	// Print STDOUT and STDERR lines streaming from Cmd
	doneChan := make(chan struct{})

	go monitorSignals(command)
	go handleProcessOutput(doneChan, command)

	finalStatus := <-command.Start() // Wait for command to be finished
	<-doneChan                       // Wait for output printer to be finished

	if finalStatus.Error != nil {
		return errors.Wrapf(finalStatus.Error, "error running command: %v", finalStatus)
	}

	return nil
}

func monitorSignals(command *cmd.Cmd) {
	handler := make(chan os.Signal, 1)
	signal.Notify(handler, syscall.SIGALRM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	for sig := range handler {
		switch sig {
		case syscall.SIGINT:
			fmt.Println(" | ctrl-c detected")
			fallthrough
		default:
			if err := command.Stop(); err != nil {
				logger.Error("error stopping command: %v", err)
			}
		}
	}
}

func handleProcessOutput(doneChan chan struct{}, command *cmd.Cmd) {
	defer close(doneChan)
	// Done when both channels have been closed
	// https://dave.cheney.net/2013/04/30/curious-channels
	for command.Stdout != nil || command.Stderr != nil {
		select {
		case line, open := <-command.Stdout:
			if !open {
				command.Stdout = nil
				continue
			}
			fmt.Println(line)
		case line, open := <-command.Stderr:
			if !open {
				command.Stderr = nil
				continue
			}
			_, _ = fmt.Fprintln(os.Stderr, line)
		}
	}
}
