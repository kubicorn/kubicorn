package stores

type Storer interface {
	Exists() bool
	Write(relativePath string, data []byte) error
}
