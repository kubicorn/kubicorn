package fs

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

type FileSystemStoreOptions struct {
	Prefix string
	Path   string
}

type FileSystemStore struct {
	options *FileSystemStoreOptions
}

func NewFileSystemStore(o *FileSystemStoreOptions) *FileSystemStore {
	return &FileSystemStore{
		options: o,
	}
}

func (fs *FileSystemStore) Exists() bool {
	if _, err := os.Stat(fmt.Sprintf("%s/%s", fs.options.Path, fs.options.Prefix)); os.IsNotExist(err) {
		return false
	}
	return true
}

func (fs *FileSystemStore) Write(relativePath string, data []byte) error {
	fqn := fmt.Sprintf("%s/%s/%s", fs.options.Path, fs.options.Prefix, relativePath)
	os.MkdirAll(path.Dir(fqn), 0700)
	fo, err := os.Create(fqn)
	if err != nil {
		return err
	}
	defer fo.Close()
	_, err = io.Copy(fo, strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	return nil
}
