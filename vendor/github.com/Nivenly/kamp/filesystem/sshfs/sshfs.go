package sshfs

import "github.com/Nivenly/kamp/filesystem"

type SSHFilesystem struct {
}

func (s *SSHFilesystem) Mount(options *filesystem.FilesystemOptions) error {
	return nil

}
func (s *SSHFilesystem) Unmount(options *filesystem.FilesystemOptions) error {
	return nil
}
