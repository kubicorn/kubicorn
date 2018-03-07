// +build ignore

package main

import (
	"fmt"
	"log"

	"github.com/mattn/go-tty"
)

func main() {
	t, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer t.Close()

	fmt.Println("Hit any key")
	for {
		r, err := t.ReadRune()
		if err != nil {
			log.Fatal(err)
		}
		if r == 0 {
			continue
		}
		fmt.Printf("0x%X: %c\n", r, r)
		if !t.Buffered() {
			break
		}
	}
}
