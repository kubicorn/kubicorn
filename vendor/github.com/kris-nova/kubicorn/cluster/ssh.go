package cluster

import (
	"github.com/kris-nova/klone/pkg/local"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"io/ioutil"
)

func ssh(initCluster *cluster.Cluster) (*cluster.Cluster, error) {
	if initCluster.Ssh.PublicKeyPath != "" {
		bytes, err := ioutil.ReadFile(local.Expand(initCluster.Ssh.PublicKeyPath))
		if err != nil {
			return nil, err
		}
		initCluster.Ssh.PublicKeyData = bytes
	}
	return initCluster, nil
}
