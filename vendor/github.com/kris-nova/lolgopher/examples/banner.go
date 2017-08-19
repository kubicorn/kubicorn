package main

import (
	"fmt"
	"os"

	"github.com/CrowdSurge/banner"
	lol "github.com/kris-nova/lolgopher"
)

func main() {
	w := &lol.Writer{Output: os.Stdout, ColorMode: lol.ColorMode256}
	fmt.Fprintln(w, "This is a test of the emergency LOL system...")
	w.Write([]byte(banner.PrintS("lolololololololololol")))
}
