package snap

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/state"
)

type SnapshotUtility struct {
	cluster        *cluster.Cluster
	stateStore     state.ClusterStorer
	kubeConfigPath string
}

func NewSnapShotUtility(actualCluster *cluster.Cluster, stateStore state.ClusterStorer, kubeConfigPath string) *SnapshotUtility {
	return &SnapshotUtility{
		cluster:        actualCluster,
		stateStore:     stateStore,
		kubeConfigPath: kubeConfigPath,
	}
}

func (s *SnapshotUtility) Capture(namespaces []string, absolutePath string) (*Snapshot, error) {
	query := NewKubernetesQuery(s.kubeConfigPath, namespaces)
	err := query.Authenticate()
	if err != nil {
		return nil, err
	}
	err = query.Execute()
	if err != nil {
		return nil, err
	}
	snapshot := NewSnapshot(s.cluster, query.Bytes(), absolutePath)
	return snapshot, nil
}
