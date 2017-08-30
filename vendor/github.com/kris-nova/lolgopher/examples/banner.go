package main

import (
	"os"

	"github.com/CrowdSurge/banner"
	lol "github.com/kris-nova/lolgopher"
)

func main() {
	w := &lol.Writer{Output: os.Stdout, ColorMode: lol.ColorMode256}
	w.Write([]byte(banner.PrintS("lolgopher")))
}
