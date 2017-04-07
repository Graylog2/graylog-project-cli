package logger

import (
	"github.com/fatih/color"
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
	println(prefixColor, noColor, prefix, format, args...)
}

func InfoWithPrefix(customPrefix string, format string, args ...interface{}) {
	println(prefixColor, noColor, customPrefix, format, args...)
}

func ColorInfo(colorValue color.Attribute, format string, args ...interface{}) {
	println(prefixColor, colorValue, prefix, format, args...)
}

func Error(format string, args ...interface{}) {
	ColorInfo(errorColor, format, args...)
}

func Debug(format string, args ...interface{}) {
	if debug {
		Info(format, args...)
	}
}

func DebugWithPrefix(customPrefix string, format string, args ...interface{}) {
	if debug {
		InfoWithPrefix(customPrefix, format, args...)
	}
}

func Fatal(format string, args ...interface{}) {
	Error(format, args...)
	os.Exit(1)
}

func Println(format string, args ...interface{}) {
	println(noColor, noColor, "", format, args...)
}

func Printf(format string, args ...interface{}) {
	print(noColor, noColor, "", format, args...)
}

func ColorPrintln(colorValue color.Attribute, format string, args ...interface{}) {
	println(noColor, colorValue, "", format, args...)
}

func ColorPrintf(colorValue color.Attribute, format string, args ...interface{}) {
	print(noColor, colorValue, "", format, args...)
}

func println(prefixColor color.Attribute, textColor color.Attribute, prefix string, format string, args ...interface{}) {
	print(prefixColor, textColor, prefix, format+"\n", args...)
}

func print(prefixColor color.Attribute, textColor color.Attribute, prefix string, format string, args ...interface{}) {
	if quiet {
		return
	}

	if prefix != "" {
		color.New(prefixColor).Printf("%s ", prefix)
	}
	color.New(textColor).Printf(format, args...)
}
