// +build darwin freebsd openbsd netbsd

package termios

import (
	"syscall"
	"testing"
)

func TestTcflush(t *testing.T) {
	f := opendev(t)
	defer f.Close()

	if err := Tcflush(f.Fd(), syscall.TCIOFLUSH); err != nil {
		t.Fatal(err)
	}
}
