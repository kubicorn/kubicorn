package subdirfs

import (
	"os"
	"time"

	"gopkg.in/src-d/go-billy.v2"
)

type fileInfo struct {
	filename string
	fi       billy.FileInfo
}

func newFileInfo(filename string, fi billy.FileInfo) billy.FileInfo {
	return &fileInfo{filename, fi}
}

func (fi *fileInfo) Name() string {
	return fi.filename
}
func (fi *fileInfo) Size() int64 {
	return fi.fi.Size()
}

func (fi *fileInfo) Mode() os.FileMode {
	return fi.fi.Mode()
}

func (fi *fileInfo) ModTime() time.Time {
	return fi.fi.ModTime()
}

func (fi *fileInfo) IsDir() bool {
	return fi.fi.IsDir()
}

func (fi *fileInfo) Sys() interface{} {
	return fi.fi.Sys()
}
