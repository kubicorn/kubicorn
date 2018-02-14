// +build ignore

package main

import (
	"encoding/base64"
	"fmt"

	"github.com/mattn/go-tty"
)

func main() {
	tty, err := tty.Open()
	defer tty.Close()

	fmt.Print("Username: ")
	username, err := tty.ReadString()
	if err != nil {
		println("canceled")
		return
	}
	fmt.Print("Password: ")
	password, err := tty.ReadPassword()
	if err != nil {
		println("canceled")
		return
	}
	fmt.Println(base64.StdEncoding.EncodeToString([]byte(username + ":" + password)))
}
