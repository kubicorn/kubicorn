package runner

type Runner interface {
	Run() error
}

type VolumePair struct {
	Host   string
	Remote string
}

type Options struct {
	Command     []string
	ImageQuery  string
	VolumePairs []*VolumePair
	Name        string
}
