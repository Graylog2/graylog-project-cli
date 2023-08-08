package logger

import "fmt"

func NewLoggableError(err error, title string, messages []string) *LoggableError {
	return &LoggableError{
		title:    title,
		Messages: messages,
		err:      err,
	}
}

type LoggableError struct {
	title    string
	Messages []string
	err      error
}

func (l LoggableError) Cause() error {
	return l.err
}

func (l LoggableError) Error() string {
	return fmt.Sprintf("%s: %s", l.title, l.err)
}
