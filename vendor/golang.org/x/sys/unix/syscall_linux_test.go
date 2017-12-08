// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux

package unix_test

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestFchmodat(t *testing.T) {
	defer chtmpdir(t)()

	touch(t, "file1")
	os.Symlink("file1", "symlink1")

	err := unix.Fchmodat(unix.AT_FDCWD, "symlink1", 0444, 0)
	if err != nil {
		t.Fatalf("Fchmodat: unexpected error: %v", err)
	}

	fi, err := os.Stat("file1")
	if err != nil {
		t.Fatal(err)
	}

	if fi.Mode() != 0444 {
		t.Errorf("Fchmodat: failed to change mode: expected %v, got %v", 0444, fi.Mode())
	}

	err = unix.Fchmodat(unix.AT_FDCWD, "symlink1", 0444, unix.AT_SYMLINK_NOFOLLOW)
	if err != unix.EOPNOTSUPP {
		t.Fatalf("Fchmodat: unexpected error: %v, expected EOPNOTSUPP", err)
	}
}

func TestIoctlGetInt(t *testing.T) {
	f, err := os.Open("/dev/random")
	if err != nil {
		t.Fatalf("failed to open device: %v", err)
	}
	defer f.Close()

	v, err := unix.IoctlGetInt(int(f.Fd()), unix.RNDGETENTCNT)
	if err != nil {
		t.Fatalf("failed to perform ioctl: %v", err)
	}

	t.Logf("%d bits of entropy available", v)
}

func TestPoll(t *testing.T) {
	f, cleanup := mktmpfifo(t)
	defer cleanup()

	const timeout = 100

	ok := make(chan bool, 1)
	go func() {
		select {
		case <-time.After(10 * timeout * time.Millisecond):
			t.Errorf("Poll: failed to timeout after %d milliseconds", 10*timeout)
		case <-ok:
		}
	}()

	fds := []unix.PollFd{{Fd: int32(f.Fd()), Events: unix.POLLIN}}
	n, err := unix.Poll(fds, timeout)
	ok <- true
	if err != nil {
		t.Errorf("Poll: unexpected error: %v", err)
		return
	}
	if n != 0 {
		t.Errorf("Poll: wrong number of events: got %v, expected %v", n, 0)
		return
	}
}

func TestPpoll(t *testing.T) {
	f, cleanup := mktmpfifo(t)
	defer cleanup()

	const timeout = 100 * time.Millisecond

	ok := make(chan bool, 1)
	go func() {
		select {
		case <-time.After(10 * timeout):
			t.Errorf("Ppoll: failed to timeout after %d", 10*timeout)
		case <-ok:
		}
	}()

	fds := []unix.PollFd{{Fd: int32(f.Fd()), Events: unix.POLLIN}}
	timeoutTs := unix.NsecToTimespec(int64(timeout))
	n, err := unix.Ppoll(fds, &timeoutTs, nil)
	ok <- true
	if err != nil {
		t.Errorf("Ppoll: unexpected error: %v", err)
		return
	}
	if n != 0 {
		t.Errorf("Ppoll: wrong number of events: got %v, expected %v", n, 0)
		return
	}
}

// mktmpfifo creates a temporary FIFO and provides a cleanup function.
func mktmpfifo(t *testing.T) (*os.File, func()) {
	err := unix.Mkfifo("fifo", 0666)
	if err != nil {
		t.Fatalf("mktmpfifo: failed to create FIFO: %v", err)
	}

	f, err := os.OpenFile("fifo", os.O_RDWR, 0666)
	if err != nil {
		os.Remove("fifo")
		t.Fatalf("mktmpfifo: failed to open FIFO: %v", err)
	}

	return f, func() {
		f.Close()
		os.Remove("fifo")
	}
}

func TestTime(t *testing.T) {
	var ut unix.Time_t
	ut2, err := unix.Time(&ut)
	if err != nil {
		t.Fatalf("Time: %v", err)
	}
	if ut != ut2 {
		t.Errorf("Time: return value %v should be equal to argument %v", ut2, ut)
	}

	var now time.Time

	for i := 0; i < 10; i++ {
		ut, err = unix.Time(nil)
		if err != nil {
			t.Fatalf("Time: %v", err)
		}

		now = time.Now()

		if int64(ut) == now.Unix() {
			return
		}
	}

	t.Errorf("Time: return value %v should be nearly equal to time.Now().Unix() %v", ut, now.Unix())
}

func TestUtime(t *testing.T) {
	defer chtmpdir(t)()

	touch(t, "file1")

	buf := &unix.Utimbuf{
		Modtime: 12345,
	}

	err := unix.Utime("file1", buf)
	if err != nil {
		t.Fatalf("Utime: %v", err)
	}

	fi, err := os.Stat("file1")
	if err != nil {
		t.Fatal(err)
	}

	if fi.ModTime().Unix() != 12345 {
		t.Errorf("Utime: failed to change modtime: expected %v, got %v", 12345, fi.ModTime().Unix())
	}
}

func TestGetrlimit(t *testing.T) {
	var rlim unix.Rlimit
	err := unix.Getrlimit(unix.RLIMIT_AS, &rlim)
	if err != nil {
		t.Fatalf("Getrlimit: %v", err)
	}
}

<<<<<<< HEAD
func TestSelect(t *testing.T) {
	_, err := unix.Select(0, nil, nil, nil, &unix.Timeval{Sec: 0, Usec: 0})
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
}

<<<<<<< HEAD
func TestFstatat(t *testing.T) {
	defer chtmpdir(t)()

	touch(t, "file1")

	var st1 unix.Stat_t
	err := unix.Stat("file1", &st1)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	var st2 unix.Stat_t
	err = unix.Fstatat(unix.AT_FDCWD, "file1", &st2, 0)
	if err != nil {
		t.Fatalf("Fstatat: %v", err)
	}

	if st1 != st2 {
		t.Errorf("Fstatat: returned stat does not match Stat")
	}

	os.Symlink("file1", "symlink1")

	err = unix.Lstat("symlink1", &st1)
	if err != nil {
		t.Fatalf("Lstat: %v", err)
	}

	err = unix.Fstatat(unix.AT_FDCWD, "symlink1", &st2, unix.AT_SYMLINK_NOFOLLOW)
	if err != nil {
		t.Fatalf("Fstatat: %v", err)
	}

	if st1 != st2 {
		t.Errorf("Fstatat: returned stat does not match Lstat")
	}
=======
func TestUname(t *testing.T) {
	var utsname unix.Utsname
	err := unix.Uname(&utsname)
	if err != nil {
		t.Fatalf("Uname: %v", err)
	}

	// conversion from []byte to string, golang.org/issue/20753
	t.Logf("OS: %s/%s %s", string(utsname.Sysname[:]), string(utsname.Machine[:]), string(utsname.Release[:]))
>>>>>>> Initial dep workover
}

=======
>>>>>>> Working on getting compiling
// utilities taken from os/os_test.go

func touch(t *testing.T, name string) {
	f, err := os.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

// chtmpdir changes the working directory to a new temporary directory and
// provides a cleanup function. Used when PWD is read-only.
func chtmpdir(t *testing.T) func() {
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("chtmpdir: %v", err)
	}
	d, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatalf("chtmpdir: %v", err)
	}
	if err := os.Chdir(d); err != nil {
		t.Fatalf("chtmpdir: %v", err)
	}
	return func() {
		if err := os.Chdir(oldwd); err != nil {
			t.Fatalf("chtmpdir: %v", err)
		}
		os.RemoveAll(d)
	}
}
