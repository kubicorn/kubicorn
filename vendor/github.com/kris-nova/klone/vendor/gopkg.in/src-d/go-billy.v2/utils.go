package billy

import (
	"io"
	"os"
)

// RemoveAll removes path and any children it contains.
// It removes everything it can but returns the first error
// it encounters. If the path does not exist, RemoveAll
// returns nil (no error).
func RemoveAll(fs Filesystem, path string) error {
	r, ok := fs.(removerAll)
	if ok {
		return r.RemoveAll(path)
	}

	return removeAll(fs, path)
}

func removeAll(fs Filesystem, path string) error {
	// This implementation is adapted from os.RemoveAll.

	// Simple case: if Remove works, we're done.
	err := fs.Remove(path)
	if err == nil || os.IsNotExist(err) {
		return nil
	}

	// Otherwise, is this a directory we need to recurse into?
	dir, serr := fs.Stat(path)
	if serr != nil {
		if os.IsNotExist(serr) {
			return nil
		}

		return serr
	}

	if !dir.IsDir() {
		// Not a directory; return the error from Remove.
		return err
	}

	// Directory.
	fis, err := fs.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Race. It was deleted between the Lstat and Open.
			// Return nil per RemoveAll's docs.
			return nil
		}

		return err
	}

	// Remove contents & return first error.
	err = nil
	for _, fi := range fis {
		cpath := fs.Join(path, fi.Name())
		err1 := removeAll(fs, cpath)
		if err == nil {
			err = err1
		}
	}

	// Remove directory.
	err1 := fs.Remove(path)
	if err1 == nil || os.IsNotExist(err1) {
		return nil
	}

	if err == nil {
		err = err1
	}

	return err

}

// WriteFile writes data to a file named by filename in the given filesystem.
// If the file does not exist, WriteFile creates it with permissions perm;
// otherwise WriteFile truncates it before writing.
func WriteFile(fs Filesystem, filename string, data []byte, perm os.FileMode) error {
	f, err := fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}

	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}

	if err1 := f.Close(); err == nil {
		err = err1
	}

	return err
}
