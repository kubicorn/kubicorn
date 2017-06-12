package resources

type Server struct {
	IP string
}

func (s *Server) Apply() error {
	return nil
}
