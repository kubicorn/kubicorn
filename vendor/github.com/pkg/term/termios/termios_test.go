package termios

import (
	"flag"
	"os"
	"runtime"
	"syscall"
	"testing"
)

var dev = flag.String("device", "/dev/tty", "device to use")

func TestTcgetattr(t *testing.T) {
	f := opendev(t)
	defer f.Close()

	var termios syscall.Termios
	if err := Tcgetattr(f.Fd(), &termios); err != nil {
		t.Fatal(err)
	}
}

func TestTcsetattr(t *testing.T) {
	f := opendev(t)
	defer f.Close()

	var termios syscall.Termios
	if err := Tcgetattr(f.Fd(), &termios); err != nil {
		t.Fatal(err)
	}
	for _, opt := range []uintptr{TCSANOW, TCSADRAIN, TCSAFLUSH} {
		if err := Tcsetattr(f.Fd(), opt, &termios); err != nil {
			t.Fatal(err)
		}
	}
}

func TestTcsendbreak(t *testing.T) {
	f := opendev(t)
	defer f.Close()

	if err := Tcsendbreak(f.Fd(), 0); err != nil {
		t.Fatal(err)
	}
}

func TestTcdrain(t *testing.T) {
	f := opendev(t)
	defer f.Close()

	if err := Tcdrain(f.Fd()); err != nil {
		t.Fatal(err)
	}
}

func TestTiocmget(t *testing.T) {
	f := opendev(t)
	defer f.Close()

	var status int
	if err := Tiocmget(f.Fd(), &status); err != nil {
		checktty(t, err)
		t.Fatal(err)
	}
}

func TestTiocmset(t *testing.T) {
	f := opendev(t)
	defer f.Close()

	var status int
	if err := Tiocmget(f.Fd(), &status); err != nil {
		checktty(t, err)
		t.Fatal(err)
	}
	if err := Tiocmset(f.Fd(), &status); err != nil {
		checktty(t, err)
		t.Fatal(err)
	}
}

func TestTiocmbis(t *testing.T) {
	f := opendev(t)
	defer f.Close()

	status := 0
	if err := Tiocmbis(f.Fd(), &status); err != nil {
		checktty(t, err)
		t.Fatal(err)
	}
}

func TestTiocmbic(t *testing.T) {
	f := opendev(t)
	defer f.Close()

	status := 0
	if err := Tiocmbic(f.Fd(), &status); err != nil {
		checktty(t, err)
		t.Fatal(err)
	}
}

func TestTiocinq(t *testing.T) {
	f := opendev(t)
	defer f.Close()

	var inq int
	if err := Tiocinq(f.Fd(), &inq); err != nil {
		t.Fatal(err)
	}
	if inq != 0 {
		t.Fatal("Expected 0 bytes, got %v", inq)
	}
}

func TestTiocoutq(t *testing.T) {
	f := opendev(t)
	defer f.Close()

	var inq int
	if err := Tiocoutq(f.Fd(), &inq); err != nil {
		t.Fatal(err)
	}
	if inq != 0 {
		t.Fatal("Expected 0 bytes, got %v", inq)
	}
}

func TestCfgetispeed(t *testing.T) {
	f := opendev(t)
	defer f.Close()

	var termios syscall.Termios
	if err := Tcgetattr(f.Fd(), &termios); err != nil {
		t.Fatal(err)
	}
	if baud := Cfgetispeed(&termios); baud == 0 && runtime.GOOS != "linux" {
		t.Fatalf("Cfgetispeed: expected > 0, got %v", baud)
	}
}

func TestCfgetospeed(t *testing.T) {
	f := opendev(t)
	defer f.Close()

	var termios syscall.Termios
	if err := Tcgetattr(f.Fd(), &termios); err != nil {
		t.Fatal(err)
	}
	if baud := Cfgetospeed(&termios); baud == 0 && runtime.GOOS != "linux" {
		t.Fatalf("Cfgetospeed: expected > 0, got %v", baud)
	}
}

func opendev(t *testing.T) *os.File {
	_, pts, err := Pty()
	if err != nil {
		t.Fatal(err)
	}
	return pts
}

func checktty(t *testing.T, err error) {

	// some ioctls fail against char devices if they do not
	// support a particular feature
	if (runtime.GOOS == "darwin" && err == syscall.ENOTTY) || (runtime.GOOS == "linux" && err == syscall.EINVAL) {
		t.Skip(err)
	}
}
