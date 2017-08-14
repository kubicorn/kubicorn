package filesystem

type Filesystem interface {
	Mount(options *FilesystemOptions) error
	Unmount(options *FilesystemOptions) error
}

type FilesystemOptions struct {
	RemoteServer string
	RemotePort   string
	RemotePath   string
	LocalPath    string
	LocalPort    string
}


