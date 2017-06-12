package resources

type ServerPool struct {
	Servers []*Server
}

func (s *ServerPool) Apply() error {
	return nil
}
