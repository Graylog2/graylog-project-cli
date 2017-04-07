package ask

import (
	"bufio"
	"fmt"
	"github.com/Graylog2/graylog-project-cli/logger"
	"io"
	"regexp"
	"strings"
)

type Asker struct {
	reader *bufio.Reader
}

func NewAsker(reader io.Reader) Asker {
	return Asker{reader: bufio.NewReader(reader)}
}

func (asker *Asker) Ask(question string, defaultValue string, validatorString string) string {
	validator := regexp.MustCompile(validatorString)

	for {
		if defaultValue == "" {
			fmt.Printf("%s ", question)
		} else {
			fmt.Printf("%s [%s] ", question, defaultValue)
		}

		text, err := asker.reader.ReadString('\n')

		if err != nil {
			logger.Fatal("Unable to read input: %v", err)
		}

		answer := strings.TrimSpace(text)

		if answer == "" && defaultValue != "" {
			return defaultValue
		} else if validator.Match([]byte(answer)) {
			return answer
		} else {
			fmt.Printf("Invalid input: %s (validator: \"%s\")\n", answer, validator.String())
		}
	}
}

func (asker *Asker) AskYesNo(question string, defaultValue bool) bool {
	if defaultValue {
		fmt.Printf("%s [Y/n] ", question)
	} else {
		fmt.Printf("%s [y/N] ", question)
	}

	text, err := asker.reader.ReadString('\n')

	if err != nil {
		logger.Fatal("Unable to read input: %v", err)
	}

	answer := strings.TrimSpace(text)

	if answer == "" {
		return defaultValue
	} else if strings.ToLower(answer) == "y" {
		return true
	} else {
		return false
	}
}
