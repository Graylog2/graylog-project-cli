package main

import (
	"github.com/Graylog2/graylog-project-cli/cmd"
	"github.com/k0kubun/pp/v3"
	"github.com/mattn/go-isatty"
	"os"
)

func init() {
	pp.Default.SetColoringEnabled(isatty.IsTerminal(os.Stdout.Fd()))
}

func main() {
	cmd.Execute()
}
