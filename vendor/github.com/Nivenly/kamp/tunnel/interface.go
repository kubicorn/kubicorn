package tunnel

type Tunneler interface {
	Tunnel(options *TunnelerOptions) error
}

type TunnelerOptions struct {
	RemoteServer string
	RemotePort   string
	LocalPort    string
}
