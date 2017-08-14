package server

type Server interface {
	Authorize(auth *RsaAuth) error
	Run() error
}

type RsaAuth struct {
	PublicKey []byte
	Username  string
}
