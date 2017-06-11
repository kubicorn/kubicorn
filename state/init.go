package state

import (
	"github.com/kris-nova/kubicorn/api"
	"github.com/kris-nova/kubicorn/api/cluster"
	"github.com/kris-nova/kubicorn/logger"
	"github.com/kris-nova/kubicorn/state/stores"
)

func InitStateStore(s stores.Storer, cluster *cluster.Cluster) error {
	logger.Info("Creating new state store for cluster [%s]", cluster.Name)
	yamlBytes, err := api.ToYaml(cluster)
	if err != nil {
		return err
	}
	logger.Info("Converting to YAML")
	err = s.Write("cluster", yamlBytes)
	if err != nil {
		return err
	}
	return nil
}
