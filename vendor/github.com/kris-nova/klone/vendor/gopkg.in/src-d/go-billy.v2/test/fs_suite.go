package test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "gopkg.in/check.v1"
	. "gopkg.in/src-d/go-billy.v2"
)

func Test(t *testing.T) { TestingT(t) }

type FilesystemSuite struct {
	FS Filesystem
}

func (s *FilesystemSuite) TestCreate(c *C) {
	f, err := s.FS.Create("foo")
	c.Assert(err, IsNil)
	c.Assert(f.Filename(), Equals, "foo")
	c.Assert(f.Close(), IsNil)
}

func (s *FilesystemSuite) TestOpen(c *C) {
	f, err := s.FS.Create("foo")
	c.Assert(err, IsNil)
	c.Assert(f.Filename(), Equals, "foo")
	c.Assert(f.Close(), IsNil)

	f, err = s.FS.Open("foo")
	c.Assert(err, IsNil)
	c.Assert(f.Filename(), Equals, "foo")
	c.Assert(f.Close(), IsNil)
}

func (s *FilesystemSuite) TestOpenNotExists(c *C) {
	f, err := s.FS.Open("not-exists")
	c.Assert(err, NotNil)
	c.Assert(f, IsNil)
}

func (s *FilesystemSuite) TestCreateDir(c *C) {
	err := s.FS.MkdirAll("foo", 0644)
	c.Assert(err, IsNil)

	f, err := s.FS.Create("foo")
	c.Assert(err, NotNil)
	c.Assert(f, IsNil)
}

func (s *FilesystemSuite) TestCreateDepth(c *C) {
	f, err := s.FS.Create("bar/foo")
	c.Assert(err, IsNil)
	c.Assert(f.Filename(), Equals, s.FS.Join("bar", "foo"))
	c.Assert(f.Close(), IsNil)
}

func (s *FilesystemSuite) TestCreateDepthAbsolute(c *C) {
	f, err := s.FS.Create("/bar/foo")
	c.Assert(err, IsNil)
	c.Assert(f.Filename(), Equals, s.FS.Join("bar", "foo"))
	c.Assert(f.Close(), IsNil)
}

func (s *FilesystemSuite) TestCreateOverwrite(c *C) {
	for i := 0; i < 3; i++ {
		f, err := s.FS.Create("foo")
		c.Assert(err, IsNil)

		l, err := f.Write([]byte(fmt.Sprintf("foo%d", i)))
		c.Assert(err, IsNil)
		c.Assert(l, Equals, 4)

		err = f.Close()
		c.Assert(err, IsNil)
	}

	f, err := s.FS.Open("foo")
	c.Assert(err, IsNil)

	wrote, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	c.Assert(string(wrote), DeepEquals, "foo2")
	c.Assert(f.Close(), IsNil)
}

func (s *FilesystemSuite) TestCreateClose(c *C) {
	f, err := s.FS.Create("foo")
	c.Assert(err, IsNil)
	c.Assert(f.IsClosed(), Equals, false)

	_, err = f.Write([]byte("foo"))
	c.Assert(err, IsNil)
	c.Assert(f.Close(), IsNil)

	f, err = s.FS.Open(f.Filename())
	c.Assert(err, IsNil)

	wrote, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	c.Assert(string(wrote), DeepEquals, "foo")
	c.Assert(f.Close(), IsNil)
}

func (s *FilesystemSuite) TestOpenFile(c *C) {
	defaultMode := os.FileMode(0666)

	f, err := s.FS.OpenFile("foo1", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, defaultMode)
	c.Assert(err, IsNil)
	s.testWriteClose(c, f, "foo1")

	// Truncate if it exists
	f, err = s.FS.OpenFile("foo1", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, defaultMode)
	c.Assert(err, IsNil)
	c.Assert(f.Filename(), Equals, "foo1")
	s.testWriteClose(c, f, "foo1overwritten")

	// Read-only if it exists
	f, err = s.FS.OpenFile("foo1", os.O_RDONLY, defaultMode)
	c.Assert(err, IsNil)
	c.Assert(f.Filename(), Equals, "foo1")
	s.testReadClose(c, f, "foo1overwritten")

	// Create when it does exist
	f, err = s.FS.OpenFile("foo1", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, defaultMode)
	c.Assert(err, IsNil)
	c.Assert(f.Filename(), Equals, "foo1")
	s.testWriteClose(c, f, "bar")

	f, err = s.FS.OpenFile("foo1", os.O_RDONLY, defaultMode)
	c.Assert(err, IsNil)
	s.testReadClose(c, f, "bar")
}

func (s *FilesystemSuite) TestOpenFileNoTruncate(c *C) {
	defaultMode := os.FileMode(0666)

	// Create when it does not exist
	f, err := s.FS.OpenFile("foo1", os.O_CREATE|os.O_WRONLY, defaultMode)
	c.Assert(err, IsNil)
	c.Assert(f.Filename(), Equals, "foo1")
	s.testWriteClose(c, f, "foo1")

	f, err = s.FS.OpenFile("foo1", os.O_RDONLY, defaultMode)
	c.Assert(err, IsNil)
	s.testReadClose(c, f, "foo1")

	// Create when it does exist
	f, err = s.FS.OpenFile("foo1", os.O_CREATE|os.O_WRONLY, defaultMode)
	c.Assert(err, IsNil)
	c.Assert(f.Filename(), Equals, "foo1")
	s.testWriteClose(c, f, "bar")

	f, err = s.FS.OpenFile("foo1", os.O_RDONLY, defaultMode)
	c.Assert(err, IsNil)
	s.testReadClose(c, f, "bar1")
}

func (s *FilesystemSuite) TestOpenFileAppend(c *C) {
	defaultMode := os.FileMode(0666)

	f, err := s.FS.OpenFile("foo1", os.O_CREATE|os.O_WRONLY|os.O_APPEND, defaultMode)
	c.Assert(err, IsNil)
	c.Assert(f.Filename(), Equals, "foo1")
	s.testWriteClose(c, f, "foo1")

	f, err = s.FS.OpenFile("foo1", os.O_WRONLY|os.O_APPEND, defaultMode)
	c.Assert(err, IsNil)
	c.Assert(f.Filename(), Equals, "foo1")
	s.testWriteClose(c, f, "bar1")

	f, err = s.FS.OpenFile("foo1", os.O_RDONLY, defaultMode)
	c.Assert(err, IsNil)
	s.testReadClose(c, f, "foo1bar1")
}

func (s *FilesystemSuite) TestOpenFileReadWrite(c *C) {
	defaultMode := os.FileMode(0666)

	f, err := s.FS.OpenFile("foo1", os.O_CREATE|os.O_TRUNC|os.O_RDWR, defaultMode)
	c.Assert(err, IsNil)
	c.Assert(f.Filename(), Equals, "foo1")

	written, err := f.Write([]byte("foobar"))
	c.Assert(written, Equals, 6)
	c.Assert(err, IsNil)

	_, err = f.Seek(0, os.SEEK_SET)
	c.Assert(err, IsNil)

	written, err = f.Write([]byte("qux"))
	c.Assert(written, Equals, 3)
	c.Assert(err, IsNil)

	_, err = f.Seek(0, os.SEEK_SET)
	c.Assert(err, IsNil)

	s.testReadClose(c, f, "quxbar")
}

func (s *FilesystemSuite) TestOpenFileWithModes(c *C) {
	f, err := s.FS.OpenFile("foo", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, customMode)
	c.Assert(err, IsNil)
	c.Assert(f.Close(), IsNil)

	fi, err := s.FS.Stat("foo")
	c.Assert(err, IsNil)
	c.Assert(fi.Mode(), Equals, os.FileMode(customMode))
}

func (s *FilesystemSuite) testWriteClose(c *C, f File, content string) {
	written, err := f.Write([]byte(content))
	c.Assert(written, Equals, len(content))
	c.Assert(err, IsNil)
	c.Assert(f.Close(), IsNil)
}

func (s *FilesystemSuite) testReadClose(c *C, f File, content string) {
	read, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	c.Assert(string(read), Equals, content)
	c.Assert(f.Close(), IsNil)
}

func (s *FilesystemSuite) TestFileWrite(c *C) {
	f, err := s.FS.Create("foo")
	c.Assert(err, IsNil)

	n, err := f.Write([]byte("foo"))
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)

	f.Seek(0, io.SeekStart)
	all, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	c.Assert(string(all), Equals, "foo")
	c.Assert(f.Close(), IsNil)
}

func (s *FilesystemSuite) TestFileWriteClose(c *C) {
	f, err := s.FS.Create("foo")
	c.Assert(err, IsNil)

	c.Assert(f.Close(), IsNil)

	_, err = f.Write([]byte("foo"))
	c.Assert(err, NotNil)
}

func (s *FilesystemSuite) TestFileRead(c *C) {
	err := WriteFile(s.FS, "foo", []byte("foo"), 0644)
	c.Assert(err, IsNil)

	f, err := s.FS.Open("foo")
	c.Assert(err, IsNil)

	all, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	c.Assert(string(all), Equals, "foo")
	c.Assert(f.Close(), IsNil)
}

func (s *FilesystemSuite) TestFileClosed(c *C) {
	err := WriteFile(s.FS, "foo", []byte("foo"), 0644)
	c.Assert(err, IsNil)

	f, err := s.FS.Open("foo")
	c.Assert(err, IsNil)
	c.Assert(f.Close(), IsNil)

	_, err = ioutil.ReadAll(f)
	c.Assert(err, NotNil)
}

func (s *FilesystemSuite) TestFileNonRead(c *C) {
	err := WriteFile(s.FS, "foo", []byte("foo"), 0644)
	c.Assert(err, IsNil)

	f, err := s.FS.OpenFile("foo", os.O_WRONLY, 0)
	c.Assert(err, IsNil)

	_, err = ioutil.ReadAll(f)
	c.Assert(err, NotNil)

	c.Assert(f.Close(), IsNil)
}

func (s *FilesystemSuite) TestFileSeekstart(c *C) {
	s.testFileSeek(c, 10, io.SeekStart)
}

func (s *FilesystemSuite) TestFileSeekCurrent(c *C) {
	s.testFileSeek(c, 5, io.SeekCurrent)
}

func (s *FilesystemSuite) TestFileSeekEnd(c *C) {
	s.testFileSeek(c, -26, io.SeekEnd)
}

func (s *FilesystemSuite) testFileSeek(c *C, offset int64, whence int) {
	err := WriteFile(s.FS, "foo", []byte("0123456789abcdefghijklmnopqrstuvwxyz"), 0644)
	c.Assert(err, IsNil)

	f, err := s.FS.Open("foo")
	c.Assert(err, IsNil)

	some := make([]byte, 5)
	_, err = f.Read(some)
	c.Assert(err, IsNil)
	c.Assert(string(some), Equals, "01234")

	p, err := f.Seek(offset, whence)
	c.Assert(err, IsNil)
	c.Assert(int(p), Equals, 10)

	all, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	c.Assert(all, HasLen, 26)
	c.Assert(string(all), Equals, "abcdefghijklmnopqrstuvwxyz")
	c.Assert(f.Close(), IsNil)
}

func (s *FilesystemSuite) TestSeekToEndAndWrite(c *C) {
	defaultMode := os.FileMode(0666)

	f, err := s.FS.OpenFile("foo1", os.O_CREATE|os.O_TRUNC|os.O_RDWR, defaultMode)
	c.Assert(err, IsNil)
	c.Assert(f.Filename(), Equals, "foo1")

	_, err = f.Seek(10, io.SeekEnd)
	c.Assert(err, IsNil)

	n, err := f.Write([]byte(`TEST`))
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 4)

	_, err = f.Seek(0, io.SeekStart)
	c.Assert(err, IsNil)

	s.testReadClose(c, f, "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00TEST")
}

func (s *FilesystemSuite) TestFileSeekClosed(c *C) {
	err := WriteFile(s.FS, "foo", []byte("foo"), 0644)
	c.Assert(err, IsNil)

	f, err := s.FS.Open("foo")
	c.Assert(err, IsNil)
	c.Assert(f.Close(), IsNil)

	_, err = f.Seek(0, 0)
	c.Assert(err, NotNil)
}

func (s *FilesystemSuite) TestFileCloseTwice(c *C) {
	f, err := s.FS.Create("foo")
	c.Assert(err, IsNil)

	c.Assert(f.Close(), IsNil)
	c.Assert(f.Close(), NotNil)
}

func (s *FilesystemSuite) TestReadDirAndDir(c *C) {
	files := []string{"foo", "bar", "qux/baz", "qux/qux"}
	for _, name := range files {
		err := WriteFile(s.FS, name, nil, 0644)
		c.Assert(err, IsNil)
	}

	info, err := s.FS.ReadDir("/")
	c.Assert(err, IsNil)
	c.Assert(info, HasLen, 3)

	info, err = s.FS.ReadDir("/qux")
	c.Assert(err, IsNil)
	c.Assert(info, HasLen, 2)

	qux := s.FS.Dir("/qux")
	info, err = qux.ReadDir("/")
	c.Assert(err, IsNil)
	c.Assert(info, HasLen, 2)
}

func (s *FilesystemSuite) TestReadDirAndWithMkDirAll(c *C) {
	err := s.FS.MkdirAll("qux", 0644)
	c.Assert(err, IsNil)

	files := []string{"qux/baz", "qux/qux"}
	for _, name := range files {
		err := WriteFile(s.FS, name, nil, 0644)
		c.Assert(err, IsNil)
	}

	info, err := s.FS.ReadDir("/")
	c.Assert(err, IsNil)
	c.Assert(info, HasLen, 1)
	c.Assert(info[0].IsDir(), Equals, true)

	info, err = s.FS.ReadDir("/qux")
	c.Assert(err, IsNil)
	c.Assert(info, HasLen, 2)

	qux := s.FS.Dir("/qux")
	info, err = qux.ReadDir("/")
	c.Assert(err, IsNil)
	c.Assert(info, HasLen, 2)
}

func (s *FilesystemSuite) TestReadDirFileInfo(c *C) {
	err := WriteFile(s.FS, "foo", []byte{'F', 'O', 'O'}, 0644)
	c.Assert(err, IsNil)

	info, err := s.FS.ReadDir("/")
	c.Assert(err, IsNil)
	c.Assert(info, HasLen, 1)

	c.Assert(info[0].Size(), Equals, int64(3))
	c.Assert(info[0].IsDir(), Equals, false)
	c.Assert(info[0].Name(), Equals, "foo")
}

func (s *FilesystemSuite) TestReadDirFileInfoDirs(c *C) {
	files := []string{"qux/baz/foo"}
	for _, name := range files {
		err := WriteFile(s.FS, name, []byte{'F', 'O', 'O'}, 0644)
		c.Assert(err, IsNil)
	}

	info, err := s.FS.ReadDir("qux")
	c.Assert(err, IsNil)
	c.Assert(info, HasLen, 1)
	c.Assert(info[0].IsDir(), Equals, true)
	c.Assert(info[0].Name(), Equals, "baz")

	info, err = s.FS.ReadDir("qux/baz")
	c.Assert(err, IsNil)
	c.Assert(info, HasLen, 1)
	c.Assert(info[0].Size(), Equals, int64(3))
	c.Assert(info[0].IsDir(), Equals, false)
	c.Assert(info[0].Name(), Equals, "foo")
	c.Assert(info[0].Mode(), Not(Equals), 0)
}

func (s *FilesystemSuite) TestStat(c *C) {
	WriteFile(s.FS, "foo/bar", []byte("foo"), customMode)

	fi, err := s.FS.Stat("foo/bar")
	c.Assert(err, IsNil)
	c.Assert(fi.Name(), Equals, "bar")
	c.Assert(fi.Size(), Equals, int64(3))
	c.Assert(fi.Mode(), Equals, customMode)
	c.Assert(fi.ModTime().IsZero(), Equals, false)
	c.Assert(fi.IsDir(), Equals, false)
}

func (s *FilesystemSuite) TestStatDir(c *C) {
	s.FS.MkdirAll("foo/bar", 0644)

	fi, err := s.FS.Stat("foo/bar")
	c.Assert(err, IsNil)
	c.Assert(fi.Name(), Equals, "bar")
	c.Assert(fi.Mode().IsDir(), Equals, true)
	c.Assert(fi.ModTime().IsZero(), Equals, false)
	c.Assert(fi.IsDir(), Equals, true)
}

func (s *FilesystemSuite) TestStatNonExistent(c *C) {
	fi, err := s.FS.Stat("non-existent")
	comment := Commentf("error: %s", err)
	c.Assert(os.IsNotExist(err), Equals, true, comment)
	c.Assert(fi, IsNil)
}

func (s *FilesystemSuite) TestDirStat(c *C) {
	files := []string{"foo", "bar", "qux/baz", "qux/qux"}
	for _, name := range files {
		err := WriteFile(s.FS, name, nil, 0644)
		c.Assert(err, IsNil)
	}

	// Some implementations detect directories based on a prefix
	// for all files; it's easy to miss path separator handling there.
	fi, err := s.FS.Stat("qu")
	c.Assert(os.IsNotExist(err), Equals, true, Commentf("error: %s", err))
	c.Assert(fi, IsNil)

	fi, err = s.FS.Stat("qux")
	c.Assert(err, IsNil)
	c.Assert(fi.Name(), Equals, "qux")
	c.Assert(fi.IsDir(), Equals, true)

	qux := s.FS.Dir("qux")

	fi, err = qux.Stat("baz")
	c.Assert(err, IsNil)
	c.Assert(fi.Name(), Equals, "baz")
	c.Assert(fi.IsDir(), Equals, false)

	fi, err = qux.Stat("/baz")
	c.Assert(err, IsNil)
	c.Assert(fi.Name(), Equals, "baz")
	c.Assert(fi.IsDir(), Equals, false)
}

func (s *FilesystemSuite) TestCreateInDir(c *C) {
	f, err := s.FS.Dir("foo").Create("bar")
	c.Assert(err, IsNil)
	c.Assert(f.Close(), IsNil)
	c.Assert(f.Filename(), Equals, "bar")

	f, err = s.FS.Open("foo/bar")
	c.Assert(f.Filename(), Equals, s.FS.Join("foo", "bar"))
	c.Assert(f.Close(), IsNil)
}

func (s *FilesystemSuite) TestRename(c *C) {
	err := WriteFile(s.FS, "foo", nil, 0644)
	c.Assert(err, IsNil)

	err = s.FS.Rename("foo", "bar")
	c.Assert(err, IsNil)

	foo, err := s.FS.Stat("foo")
	c.Assert(foo, IsNil)
	c.Assert(os.IsNotExist(err), Equals, true)

	bar, err := s.FS.Stat("bar")
	c.Assert(bar, NotNil)
	c.Assert(err, IsNil)
}

func (s *FilesystemSuite) TestRenameToDir(c *C) {
	err := WriteFile(s.FS, "foo", nil, 0644)
	c.Assert(err, IsNil)

	err = s.FS.Rename("foo", "bar/qux")
	c.Assert(err, IsNil)

	old, err := s.FS.Stat("foo")
	c.Assert(old, IsNil)
	c.Assert(os.IsNotExist(err), Equals, true)

	dir, err := s.FS.Stat("bar")
	c.Assert(dir, NotNil)
	c.Assert(err, IsNil)

	file, err := s.FS.Stat("bar/qux")
	c.Assert(file.Name(), Equals, "qux")
	c.Assert(err, IsNil)
}

func (s *FilesystemSuite) TestRenameDir(c *C) {
	err := s.FS.MkdirAll("foo", 0644)
	c.Assert(err, IsNil)

	err = WriteFile(s.FS, "foo/bar", nil, 0644)
	c.Assert(err, IsNil)

	err = s.FS.Rename("foo", "bar")
	c.Assert(err, IsNil)

	dirfoo, err := s.FS.Stat("foo")
	c.Assert(dirfoo, IsNil)
	c.Assert(os.IsNotExist(err), Equals, true)

	dirbar, err := s.FS.Stat("bar")
	c.Assert(err, IsNil)
	c.Assert(dirbar, NotNil)

	foo, err := s.FS.Stat("foo/bar")
	c.Assert(os.IsNotExist(err), Equals, true)
	c.Assert(foo, IsNil)

	bar, err := s.FS.Stat("bar/bar")
	c.Assert(err, IsNil)
	c.Assert(bar, NotNil)
}

func (s *FilesystemSuite) TestTempFile(c *C) {
	f, err := s.FS.TempFile("", "bar")
	c.Assert(err, IsNil)
	c.Assert(f.Close(), IsNil)

	c.Assert(strings.HasPrefix(f.Filename(), "bar"), Equals, true)
}

func (s *FilesystemSuite) TestTempFileWithPath(c *C) {
	f, err := s.FS.TempFile("foo", "bar")
	c.Assert(err, IsNil)
	c.Assert(f.Close(), IsNil)

	c.Assert(strings.HasPrefix(f.Filename(), s.FS.Join("foo", "bar")), Equals, true)
}

func (s *FilesystemSuite) TestTempFileFullWithPath(c *C) {
	f, err := s.FS.TempFile("/foo", "bar")
	c.Assert(err, IsNil)
	c.Assert(f.Close(), IsNil)

	c.Assert(strings.HasPrefix(f.Filename(), s.FS.Join("foo", "bar")), Equals, true)
}

func (s *FilesystemSuite) TestOpenAndWrite(c *C) {
	err := WriteFile(s.FS, "foo", nil, 0644)
	c.Assert(err, IsNil)

	foo, err := s.FS.Open("foo")
	c.Assert(foo, NotNil)
	c.Assert(err, IsNil)

	n, err := foo.Write([]byte("foo"))
	c.Assert(err, NotNil)
	c.Assert(n, Equals, 0)

	c.Assert(foo.Close(), IsNil)
}

func (s *FilesystemSuite) TestOpenAndStat(c *C) {
	err := WriteFile(s.FS, "foo", []byte("foo"), 0644)
	c.Assert(err, IsNil)

	foo, err := s.FS.Open("foo")
	c.Assert(foo, NotNil)
	c.Assert(foo.Filename(), Equals, "foo")
	c.Assert(err, IsNil)
	c.Assert(foo.Close(), IsNil)

	stat, err := s.FS.Stat("foo")
	c.Assert(stat, NotNil)
	c.Assert(err, IsNil)
	c.Assert(stat.Name(), Equals, "foo")
	c.Assert(stat.Size(), Equals, int64(3))
}

func (s *FilesystemSuite) TestRemove(c *C) {
	f, err := s.FS.Create("foo")
	c.Assert(err, IsNil)
	c.Assert(f.Close(), IsNil)

	err = s.FS.Remove("foo")
	c.Assert(err, IsNil)
}

func (s *FilesystemSuite) TestRemoveNonExisting(c *C) {
	err := s.FS.Remove("NON-EXISTING")
	c.Assert(err, NotNil)
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (s *FilesystemSuite) TestRemoveNotEmptyDir(c *C) {
	err := WriteFile(s.FS, "foo", nil, 0644)
	c.Assert(err, IsNil)

	err = s.FS.Remove("no-exists")
	c.Assert(err, NotNil)
}

func (s *FilesystemSuite) TestRemoveTempFile(c *C) {
	f, err := s.FS.TempFile("test-dir", "test-prefix")
	c.Assert(err, IsNil)

	fn := f.Filename()
	c.Assert(err, IsNil)
	c.Assert(f.Close(), IsNil)

	c.Assert(s.FS.Remove(fn), IsNil)
}

func (s *FilesystemSuite) TestJoin(c *C) {
	c.Assert(s.FS.Join("foo", "bar"), Equals, fmt.Sprintf("foo%cbar", filepath.Separator))
}

func (s *FilesystemSuite) TestBase(c *C) {
	c.Assert(s.FS.Base(), Not(Equals), "")
}

func (s *FilesystemSuite) TestReadAtOnReadWrite(c *C) {
	f, err := s.FS.Create("foo")
	c.Assert(err, IsNil)
	_, err = f.Write([]byte("abcdefg"))
	c.Assert(err, IsNil)

	rf, ok := f.(io.ReaderAt)
	c.Assert(ok, Equals, true)

	b := make([]byte, 3)
	n, err := rf.ReadAt(b, 2)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)
	c.Assert(string(b), Equals, "cde")
	c.Assert(f.Close(), IsNil)
}

func (s *FilesystemSuite) TestReadAtOnReadOnly(c *C) {
	err := WriteFile(s.FS, "foo", []byte("abcdefg"), 0644)
	c.Assert(err, IsNil)

	f, err := s.FS.Open("foo")
	c.Assert(err, IsNil)

	rf, ok := f.(io.ReaderAt)
	c.Assert(ok, Equals, true)

	b := make([]byte, 3)
	n, err := rf.ReadAt(b, 2)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)
	c.Assert(string(b), Equals, "cde")
	c.Assert(f.Close(), IsNil)
}

func (s *FilesystemSuite) TestReadAtEOF(c *C) {
	err := WriteFile(s.FS, "foo", []byte("TEST"), 0644)
	c.Assert(err, IsNil)

	f, err := s.FS.Open("foo")
	c.Assert(err, IsNil)

	rf, ok := f.(io.ReaderAt)
	c.Assert(ok, Equals, true)

	b := make([]byte, 5)
	n, err := rf.ReadAt(b, 0)
	c.Assert(err, Equals, io.EOF)
	c.Assert(n, Equals, 4)
	c.Assert(string(b), Equals, "TEST\x00")

	err = f.Close()
	c.Assert(err, IsNil)
}

func (s *FilesystemSuite) TestReadAtOffset(c *C) {
	err := WriteFile(s.FS, "foo", []byte("TEST"), 0644)
	c.Assert(err, IsNil)

	f, err := s.FS.Open("foo")
	c.Assert(err, IsNil)

	rf, ok := f.(io.ReaderAt)
	c.Assert(ok, Equals, true)

	o, err := f.Seek(0, io.SeekCurrent)
	c.Assert(err, IsNil)
	c.Assert(o, Equals, int64(0))

	b := make([]byte, 4)
	n, err := rf.ReadAt(b, 0)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 4)
	c.Assert(string(b), Equals, "TEST")

	o, err = f.Seek(0, io.SeekCurrent)
	c.Assert(err, IsNil)
	c.Assert(o, Equals, int64(0))

	err = f.Close()
	c.Assert(err, IsNil)
}

func (s *FilesystemSuite) TestReadWriteLargeFile(c *C) {
	f, err := s.FS.Create("foo")
	c.Assert(err, IsNil)

	size := 1 << 20

	n, err := f.Write(bytes.Repeat([]byte("F"), size))
	c.Assert(err, IsNil)
	c.Assert(n, Equals, size)

	c.Assert(f.Close(), IsNil)

	f, err = s.FS.Open("foo")
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	c.Assert(len(b), Equals, size)
	c.Assert(f.Close(), IsNil)
}

func (s *FilesystemSuite) TestRemoveAllNonExistent(c *C) {
	c.Assert(RemoveAll(s.FS, "non-existent"), IsNil)
}

func (s *FilesystemSuite) TestRemoveAllEmptyDir(c *C) {
	c.Assert(s.FS.MkdirAll("empty", os.FileMode(0755)), IsNil)
	c.Assert(RemoveAll(s.FS, "empty"), IsNil)
	_, err := s.FS.Stat("empty")
	c.Assert(err, NotNil)
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (s *FilesystemSuite) TestRemoveAll(c *C) {
	fnames := []string{
		"foo/1",
		"foo/2",
		"foo/bar/1",
		"foo/bar/2",
		"foo/bar/baz/1",
		"foo/bar/baz/qux/1",
		"foo/bar/baz/qux/2",
		"foo/bar/baz/qux/3",
	}

	for _, fname := range fnames {
		err := WriteFile(s.FS, fname, nil, 0644)
		c.Assert(err, IsNil)
	}

	c.Assert(RemoveAll(s.FS, "foo"), IsNil)

	for _, fname := range fnames {
		_, err := s.FS.Stat(fname)
		comment := Commentf("not removed: %s %s", fname, err)
		c.Assert(os.IsNotExist(err), Equals, true, comment)
	}
}

func (s *FilesystemSuite) TestRemoveAllRelative(c *C) {
	fnames := []string{
		"foo/1",
		"foo/2",
		"foo/bar/1",
		"foo/bar/2",
		"foo/bar/baz/1",
		"foo/bar/baz/qux/1",
		"foo/bar/baz/qux/2",
		"foo/bar/baz/qux/3",
	}

	for _, fname := range fnames {
		err := WriteFile(s.FS, fname, nil, 0644)
		c.Assert(err, IsNil)
	}

	c.Assert(RemoveAll(s.FS, "foo/bar/.."), IsNil)

	for _, fname := range fnames {
		_, err := s.FS.Stat(fname)
		comment := Commentf("not removed: %s %s", fname, err)
		c.Assert(os.IsNotExist(err), Equals, true, comment)
	}
}

func (s *FilesystemSuite) TestMkdirAll(c *C) {
	err := s.FS.MkdirAll("empty", os.FileMode(0755))
	c.Assert(err, IsNil)

	fi, err := s.FS.Stat("empty")
	c.Assert(err, IsNil)
	c.Assert(fi.IsDir(), Equals, true)
}

func (s *FilesystemSuite) TestMkdirAllNested(c *C) {
	err := s.FS.MkdirAll("foo/bar/baz", os.FileMode(0755))
	c.Assert(err, IsNil)

	fi, err := s.FS.Stat("foo/bar/baz")
	c.Assert(err, IsNil)
	c.Assert(fi.IsDir(), Equals, true)
}

func (s *FilesystemSuite) TestMkdirAllIdempotent(c *C) {
	err := s.FS.MkdirAll("empty", 0755)
	c.Assert(err, IsNil)
	fi, err := s.FS.Stat("empty")
	c.Assert(err, IsNil)
	c.Assert(fi.IsDir(), Equals, true)

	// idempotent
	err = s.FS.MkdirAll("empty", 0755)
	c.Assert(err, IsNil)
	fi, err = s.FS.Stat("empty")
	c.Assert(err, IsNil)
	c.Assert(fi.IsDir(), Equals, true)
}

func (s *FilesystemSuite) TestMkdirAllAndOpenFile(c *C) {
	err := s.FS.MkdirAll("dir", os.FileMode(0755))
	c.Assert(err, IsNil)

	f, err := s.FS.Create("dir/bar/foo")
	c.Assert(err, IsNil)
	c.Assert(f.Close(), IsNil)

	fi, err := s.FS.Stat("dir/bar/foo")
	c.Assert(err, IsNil)
	c.Assert(fi.IsDir(), Equals, false)
}

func (s *FilesystemSuite) TestMkdirAllWithExistingFile(c *C) {
	f, err := s.FS.Create("dir/foo")
	c.Assert(err, IsNil)
	c.Assert(f.Close(), IsNil)

	err = s.FS.MkdirAll("dir/foo", os.FileMode(0755))
	c.Assert(err, NotNil)

	fi, err := s.FS.Stat("dir/foo")
	c.Assert(err, IsNil)
	c.Assert(fi.IsDir(), Equals, false)
}

func (s *FilesystemSuite) TestWriteFile(c *C) {
	err := WriteFile(s.FS, "foo", []byte("bar"), 0777)
	c.Assert(err, IsNil)

	f, err := s.FS.Open("foo")
	c.Assert(err, IsNil)

	wrote, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	c.Assert(string(wrote), DeepEquals, "bar")

	c.Assert(f.Close(), IsNil)
}
