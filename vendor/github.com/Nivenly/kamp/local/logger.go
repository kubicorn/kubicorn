package local

import (
	"fmt"
	"github.com/fatih/color"
	"strings"
	"time"
)

var (
	LogLevel = 2
	LogColor = true
)

func Debug(format string, a ...interface{}) {
	if LogLevel >= 4 {
		fmt.Printf(label(format, "▶"), a...)
	}
}

func Info(format string, a ...interface{}) {
	if LogLevel >= 3 {
		if LogColor {
			color.Blue(label(format, "✔"), a...)
		} else {
			fmt.Printf(label(format, "✔"), a...)
		}
	}
}

func Warning(format string, a ...interface{}) {
	if LogLevel >= 2 {
		if LogColor {
			color.Green(label(format, "!"), a...)
		} else {
			fmt.Printf(label(format, "!"), a...)
		}
	}
}

func Critical(format string, a ...interface{}) {
	if LogLevel >= 1 {
		if LogColor {
			color.Red(label(format, "✖"), a...)
		} else {
			fmt.Printf(label(format, "✖"), a...)
		}
	}
}

func label(format, label string) string {
	t := time.Now()
	rfct := t.Format(time.RFC3339)
	if !strings.Contains(format, "\n") {
		format = fmt.Sprintf("%s%s", format, "\n")
	}
	return fmt.Sprintf("%s [%s]  %s", rfct, label, format)
}
