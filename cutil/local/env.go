package local

import (
	"github.com/kris-nova/kubicorn/logger"
	"os"
	"os/user"
	"strings"
)

func Home() string {
	home := os.Getenv("HOME")
	if strings.Contains(home, "root") {
		return "/root"
	}
	usr, err := user.Current()
	if err != nil {
		logger.Warning("unable to find user: %v", err)
		return ""
	}
	return usr.HomeDir
}

func Expand(path string) string {
	if strings.Contains(path, "~") {
		return strings.Replace(path, "~", Home(), 1)
	}
	return path
}
