package local

import (
	"os/user"
	"strings"
)

func Home() string {
	usr, err := user.Current()
	if err != nil {
		Printf("unable to find user: %v", err)
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
