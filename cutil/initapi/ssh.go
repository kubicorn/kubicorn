package initapi

import (
	//"github.com/gravitational/teleport/lib/sshutils"
	//"github.com/gravitational/teleport/lib/sshutils"
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

		//fp, err := sshutils.PrivateKeyFingerprint(bytes)
		//if err != nil {
		//	return nil, err
		//}
		//initCluster.Ssh.PublicKeyFingerprint = fp
	}

	return initCluster, nil
}
