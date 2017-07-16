package cluster

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/logger"
)

type preProcessorFunc func(initCluster *cluster.Cluster) (*cluster.Cluster, error)

var preProcessors = []preProcessorFunc{
	ssh,
}

func InitCluster(initCluster *cluster.Cluster) (*cluster.Cluster, error) {
	logger.Info("Init Cluster")
	logger.Debug("Running preprocessors")
	for _, f := range preProcessors {
		var err error
		initCluster, err = f(initCluster)
		if err != nil {
			return nil, err
		}
	}
	return initCluster, nil
}
