package local

import (
	"testing"
	"os"
	"os/user"
)

func TestHomeWithRootHappy(t *testing.T) {
	os.Setenv("HOME", "/root")
	location := Home();
	if location != "/root" {
		t.Errorf("Home location incorrect: %s", location)
	}
}

func TestHomeAsUserHappy(t *testing.T) {
	os.Setenv("HOME", "/user/test")
	usr, _ := user.Current()
	location := Home();
	if usr.HomeDir != location {
		t.Errorf("Home location incorrect: %s should be %s", location, usr.HomeDir)
	}
}

func TestExpandAsUserHappy(t *testing.T) {
	path := Expand("~/test")
	usr, _ := user.Current()
	if path != usr.HomeDir + "/test" {
		t.Errorf("Home location incorrect: %s should be %s", path, usr.HomeDir + "/test")
	}
}

func TestExpandAsRootHappy(t *testing.T) {
	os.Setenv("HOME", "/root")
	path := Expand("~/test")
	if path != "/root/test" {
		t.Errorf("Home location incorrect: %s should be %s", path, "/root/test")
	}
}

func TestExpandAsRootNoTildeHappy(t *testing.T) {
	os.Setenv("HOME", "/root")
	path := Expand("/var/test")
	if path != "/var/test" {
		t.Errorf("Home location incorrect: %s should be %s", path, "/var/test")
	}
}

func TestExpandAsUserNoTildeHappy(t *testing.T) {
	os.Setenv("HOME", "/root")
	path := Expand("/var/test")
	if path != "/var/test" {
		t.Errorf("Home location incorrect: %s should be %s", path, "/var/test")
	}
}