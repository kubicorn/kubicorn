package logger

import (
	"fmt"
	"github.com/fatih/color"
	"strings"
	"time"
)

var (
	Level = 2
	Color = true
)

func Always(format string, a ...interface{}) {
	color.Green(label(format, "✿"), a...)
}

func Info(format string, a ...interface{}) {
	if Level >= 3 {
		if Color {
			color.Cyan(label(format, "✔"), a...)
		} else {
			fmt.Printf(label(format, "✔"), a...)
		}
	}
}

func Debug(format string, a ...interface{}) {
	if Level >= 4 {
		fmt.Printf(label(format, "▶"), a...)
	}
}

func Warning(format string, a ...interface{}) {
	if Level >= 2 {
		if Color {
			color.Green(label(format, "!"), a...)
		} else {
			fmt.Printf(label(format, "!"), a...)
		}
	}
}

func Critical(format string, a ...interface{}) {
	if Level >= 1 {
		if Color {
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
