package lol

import (
	"fmt"
	"os"
)

var w = &Writer{Output: os.Stdout, ColorMode: ColorMode256}

func Println(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(w, a...)
}

func Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(w, format, a...)
}
