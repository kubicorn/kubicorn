package termios

import "testing"

func TestTcflush(t *testing.T) {
	f := opendev(t)
	defer f.Close()

	if err := Tcflush(f.Fd(), TCIOFLUSH); err != nil {
		t.Fatal(err)
	}
}
