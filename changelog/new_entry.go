package changelog

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/manifoldco/promptui"
	"github.com/mattn/go-isatty"
	"github.com/samber/lo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
type = "{{ .Type }}"
message = "{{ .Message }}"

issues = [{{ .Issues }}]
pulls = [{{ .PullRequests }}]

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

const minimalEntryTemplate = `type = "{{ .Type }}" # One of: a(dded), c(hanged), d(eprecated), r(emoved), f(ixed), s(ecurity)
message = "{{ .Message }}"

issues = [{{ .Issues }}]
pulls = [{{ .PullRequests }}]
`

var filenamePattern = regexp.MustCompile("^(issue|pr)-(\\d+)\\.toml$")

type TemplateData struct {
	Type         string
	Message      string
	Issues       string
	PullRequests string
}

func NewEntry(path string, edit bool, useMinimalTemplate bool, interactive bool) error {
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
		tmplString := entryTemplate
		if useMinimalTemplate {
			tmplString = minimalEntryTemplate
		}
		tmpl, err := template.New(path).Parse(tmplString)
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

		data := TemplateData{
			Type:         "fixed",
			Message:      "Fix [...].",
			Issues:       fmt.Sprintf("\"%s\"", issueNumber),
			PullRequests: fmt.Sprintf("\"%s\"", prNumber),
		}

		if interactive {
			if !isatty.IsTerminal(os.Stdout.Fd()) {
				return errors.New("unable to use interactive mode, output is not a terminal")
			}
			if err := askForContent(&data); err != nil {
				return err
			}
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, &data)
		if err != nil {
			return fmt.Errorf("couldn't generate entry: %w", err)
		}

		if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("couldn't write changelog entry %s: %w", path, err)
		}

		fmt.Println("Created changelog entry file:", path)
	} else {
		fmt.Println("Skipping existing changelog entry file:", path)
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

func askForContent(data *TemplateData) error {
	types := []string{"added", "changed", "deprecated", "removed", "fixed", "security"}
	defaultType := 4 // Fixed should be the default choice
	defaultMessages := map[string]string{
		"added":      "Add [...].",
		"changed":    "Change [...].",
		"deprecated": "Deprecate [...].",
		"removed":    "Remove [...].",
		"fixed":      "Fix [...].",
		"security":   "Fix [...].",
	}

	promptType := promptui.Select{
		Label:     "Please select an entry type",
		Items:     types,
		Size:      len(types),
		CursorPos: defaultType,
	}

	_, typeResult, err := promptType.Run()
	if err != nil {
		return fmt.Errorf("couldn't get type interactively: %w", err)
	}

	promptMessage := promptui.Prompt{
		Label:   "Message",
		Default: defaultMessages[typeResult],
		Validate: func(input string) error {
			cleanedInput := cleanupInput(input)
			if len(cleanedInput) == 0 {
				return errors.New("message must not be empty")
			}
			firstWord := strings.Split(input, " ")[0]
			if firstWord != cases.Title(language.English).String(firstWord) {
				return errors.New("message should be a full sentence and start with an uppercase character")
			}
			return nil
		},
	}

	fmt.Println("Please enter a message. (make sure it's a full sentence and ends with a \".\"")
	messageResult, err := promptMessage.Run()
	if err != nil {
		return fmt.Errorf("couldn't get message interactively: %w", err)
	}

	promptIssues := promptui.Prompt{
		Label:    "Issues (comma separated string)",
		Default:  cleanupInput(data.Issues),
		Validate: validateIssueNumberInput,
	}

	issuesResult, err := promptIssues.Run()
	if err != nil {
		return err
	}

	promptPrs := promptui.Prompt{
		Label:    "Pull-requests (comma separated string)",
		Default:  cleanupInput(data.PullRequests),
		Validate: validateIssueNumberInput,
	}

	prsResult, err := promptPrs.Run()
	if err != nil {
		return err
	}

	data.Type = typeResult
	data.Message = messageResult
	data.Issues = strings.Join(cleanInputList(issuesResult, true), ", ")
	data.PullRequests = strings.Join(cleanInputList(prsResult, true), ", ")

	return nil
}

func cleanupInput(input string) string {
	return strings.ReplaceAll(strings.TrimSpace(input), "\"", "")
}

func cleanInputList(input string, quote bool) []string {
	return lo.Map[string, string](strings.Split(input, ","), func(issue string, _ int) string {
		if quote {
			return fmt.Sprintf("\"%s\"", cleanupInput(issue))
		} else {
			return fmt.Sprintf("%s", cleanupInput(issue))
		}
	})
}

func validateIssueNumberInput(input string) error {
	if cleanupInput(input) == "" {
		return nil
	}
	list := cleanInputList(input, false)
	if len(list) > 0 {
		for _, issue := range list {
			if !regexp.MustCompile("\\d+$").MatchString(cleanupInput(issue)) {
				return errors.New("issue or pull-request value must end with a number")
			}
		}
	}
	return nil
}
