package snap

import "github.com/kris-nova/kubicorn/apis/cluster"

type Snapshot struct {
	infra        *cluster.Cluster
	app          []byte
	absolutePath string
}

func NewSnapshot(actualCluster *cluster.Cluster, appData []byte, absolutePath string) *Snapshot {
	return &Snapshot{
		infra:        actualCluster,
		app:          appData,
		absolutePath: absolutePath,
	}
}

func (s *Snapshot) AbsolutelyPath() string {
	return s.absolutePath
}
