package changelog

import (
	"bytes"
	"fmt"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/mattn/go-isatty"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

const entryTemplate = `# PLEASE REMOVE COMMENTS AND OPTIONAL FIELDS! THANKS!

# Entry type according to https://keepachangelog.com/en/1.0.0/
# One of: a(dded), c(hanged), d(eprecated), r(emoved), f(ixed), s(ecurity)
type = "fixed"
message = "Fix ."

issues = ["{{ .IssueNumber }}"]
pulls = ["{{ .PRNumber }}"]

contributors = [""]

details.user = """
This text contains a more detailed description of the change for users.

It can be written in Markdown to support [links](https://example.com) and
other elements.
"""

details.ops = """
This text contains operations related details for the change.

It can be written in Markdown to support [links](https://example.com) and
other elements.
"""
`

var filenamePattern = regexp.MustCompile("^(issue|pr)-(\\d+)\\.toml$")

func NewEntry(path string, edit bool) error {
	file := filepath.Base(path)
	directory := filepath.Dir(path)

	if !filenamePattern.MatchString(file) {
		return fmt.Errorf("invalid entry filename - allowed pattern: %s (Examples: issue-123.toml, pr-456.toml)", filenamePattern)
	}

	matches := filenamePattern.FindStringSubmatch(file)
	entryType := matches[1]
	number := matches[2]

	if !strings.HasPrefix(file, "issue-") && !strings.HasPrefix(file, "pr-") {
		return fmt.Errorf("changelog entry file names should have an issue- or pr- prefix")
	}
	if !strings.HasSuffix(file, ".toml") {
		return fmt.Errorf("changelog entry file names should have a .toml suffix")
	}

	if !utils.FileExists(directory) {
		if err := os.MkdirAll(directory, 0755); err != nil {
			return fmt.Errorf("couldn't create entry directory %s: %w", directory, err)
		}
	}

	if !utils.FileExists(path) {
		tmpl, err := template.New(path).Parse(entryTemplate)
		if err != nil {
			return fmt.Errorf("couldn't parse entry template: %w", err)
		}

		issueNumber := ""
		prNumber := ""

		switch entryType {
		case "issue":
			issueNumber = number
		case "pr":
			prNumber = number
		default:
			return fmt.Errorf("unknown changelog entry type: %s", entryType)
		}

		data := struct {
			IssueNumber string
			PRNumber    string
		}{issueNumber, prNumber}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, &data)
		if err != nil {
			return fmt.Errorf("couldn't generate entry: %w", err)
		}

		if err := ioutil.WriteFile(path, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("couldn't write changelog entry %s: %w", path, err)
		}
	}

	if edit && isatty.IsTerminal(os.Stdout.Fd()) {
		editor := os.Getenv("EDITOR")

		if editor == "" {
			return fmt.Errorf("couldn't edit %s because $EDITOR environment variable is not defined", path)
		}

		cmd := exec.Command(editor, path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("editing %s failed: %w", path, err)
		}
	}

	return nil
}
