package logger

import (
	"github.com/fatih/color"
	"io"
	"os"
)

var debug bool
var quiet bool
var prefix string

const noColor = color.Reset
const prefixColor = color.FgBlue
const errorColor = color.FgRed

func SetDebug(value bool) {
	debug = value
}

func SetQuiet(value bool) {
	quiet = value
}

func SetPrefix(value string) {
	prefix = value
}

func Info(format string, args ...interface{}) {
	if !quiet {
		println(os.Stdout, prefixColor, noColor, prefix, format, args...)
	}
}

func InfoWithPrefix(customPrefix string, format string, args ...interface{}) {
	if !quiet {
		println(os.Stdout, prefixColor, noColor, customPrefix, format, args...)
	}
}

func ColorInfo(colorValue color.Attribute, format string, args ...interface{}) {
	if !quiet {
		println(os.Stdout, prefixColor, colorValue, prefix, format, args...)
	}
}

func Error(format string, args ...interface{}) {
	println(os.Stderr, prefixColor, errorColor, prefix, format, args...)
}

func Debug(format string, args ...interface{}) {
	if debug && !quiet {
		Info(format, args...)
	}
}

func DebugWithPrefix(customPrefix string, format string, args ...interface{}) {
	if debug && !quiet {
		InfoWithPrefix(customPrefix, format, args...)
	}
}

func Fatal(format string, args ...interface{}) {
	Error(format, args...)
	os.Exit(1)
}

func Println(format string, args ...interface{}) {
	if !quiet {
		println(os.Stdout, noColor, noColor, "", format, args...)
	}
}

func Printf(format string, args ...interface{}) {
	if !quiet {
		print(os.Stdout, noColor, noColor, "", format, args...)
	}
}

func ColorPrintln(colorValue color.Attribute, format string, args ...interface{}) {
	if !quiet {
		println(os.Stdout, noColor, colorValue, "", format, args...)
	}
}

func ColorPrintf(colorValue color.Attribute, format string, args ...interface{}) {
	if !quiet {
		print(os.Stdout, noColor, colorValue, "", format, args...)
	}
}

func println(output io.Writer, prefixColor color.Attribute, textColor color.Attribute, prefix string, format string, args ...interface{}) {
	print(output, prefixColor, textColor, prefix, format+"\n", args...)
}

func print(output io.Writer, prefixColor color.Attribute, textColor color.Attribute, prefix string, format string, args ...interface{}) {
	if prefix != "" {
		color.New(prefixColor).Fprintf(output, "%s ", prefix)
	}
	color.New(textColor).Fprintf(output, format, args...)
}
