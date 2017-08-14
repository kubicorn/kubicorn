package local

import (
	"gopkg.in/src-d/go-git.v4"
	"os"
	"os/user"
	"strings"
)

type KampConfig struct {
	ProjectName string
}

func LocalGit() (*git.Repository, error) {
	r, err := git.PlainOpen("./")
	if err != nil {
		return nil, err
	}

	return r, nil
}

func GetLocal() (*KampConfig, error) {
	_, err := LocalGit()
	if err != nil {
		Warning("Not at root of a git repository")
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return &KampConfig{ProjectName: wd}, nil
}

func Home() string {
	home := os.Getenv("HOME")
	if strings.Contains(home, "root") {
		return "/root"
	}
	usr, err := user.Current()
	if err != nil {
		Warning("Unable to find user: %v", err)
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

func User() string {
	usr, err := user.Current()
	if err != nil {
		Warning("Unable to find user: %v", err)
		return ""
	}
	return usr.Name
}
