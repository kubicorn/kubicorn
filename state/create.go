package state

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/logger"
	"github.com/kris-nova/kubicorn/state/stores"
)

func InitStateStore(s stores.ClusterStorer, c *cluster.Cluster) error {
	logger.Info("Creating new state store for cluster [%s]", c.Name)

	err := s.Commit(c)
	if err != nil {
		return err
	}
	return nil
}
