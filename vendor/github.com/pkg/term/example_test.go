package term

import (
	"log"
	"time"
)

// Open a terminal in raw mode at 19200 baud.
func ExampleOpen() {
	Open("/dev/ttyUSB0", Speed(19200), RawMode)
}

// Reset an Arduino by toggling the DTR signal.
func ExampleTerm_SetDTR() {
	t, _ := Open("/dev/USB0")
	t.SetDTR(false) // toggle DTR low
	time.Sleep(250 * time.Millisecond)
	t.SetDTR(true) // raise DTR, resets Ardunio
}

// Send Break to the remote DTE.
func ExampleTerm_SendBreak() {
	t, _ := Open("/dev/ttyUSB0")
	for {
		time.Sleep(3 * time.Second)
		log.Println("Break...")
		t.SendBreak()
	}
}

// Restore the terminal state
func ExampleTerm_Restore() {
	t, _ := Open("/dev/tty")
	// mutate terminal state
	t.Restore()
}
