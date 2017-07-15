package subdirfs

import (
	"io/ioutil"
	stdos "os"
	"testing"

	"gopkg.in/src-d/go-billy.v2"
	"gopkg.in/src-d/go-billy.v2/osfs"
	"gopkg.in/src-d/go-billy.v2/test"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type FilesystemSuite struct {
	test.FilesystemSuite
	cfs  billy.Filesystem
	path string
}

var _ = Suite(&FilesystemSuite{})

func (s *FilesystemSuite) SetUpTest(c *C) {
	s.path, _ = ioutil.TempDir(stdos.TempDir(), "go-billy-subdirfs-test")
	fs := osfs.New(s.path)

	s.cfs = New(fs, "test-subdir")
	s.FilesystemSuite.FS = s.cfs
}

func (s *FilesystemSuite) TearDownTest(c *C) {
	fi, err := ioutil.ReadDir(s.path)
	c.Assert(err, IsNil)
	c.Assert(len(fi) <= 1, Equals, true)

	err = stdos.RemoveAll(s.path)
	c.Assert(err, IsNil)
}
