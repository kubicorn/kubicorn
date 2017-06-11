package apply

import (
	"crypto/md5"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/apply/cloudprovider"
	"github.com/kris-nova/kubicorn/apply/cloudprovider/amazon"
	"github.com/kris-nova/kubicorn/apply/cloudprovider/azure"
	"github.com/kris-nova/kubicorn/apply/cloudprovider/baremetal"
	"github.com/kris-nova/kubicorn/apply/cloudprovider/google"
	"github.com/kris-nova/kubicorn/logger"
	"github.com/kris-nova/kubicorn/state/stores"
)

func ApplyCluster(s stores.ClusterStorer, expectedCluster *cluster.Cluster) error {

	actualCluster, err := GetActualCluster(s, expectedCluster)
	if err != nil {
		return err
	}
	same, err := CompareClusters(expectedCluster, actualCluster)
	if err != nil {
		return err
	}
	if same {
		logger.Info("Cluster config has not changed, not applying.")
		return nil
	}
	logger.Info("Cluster delta detected. Applying changes.")

	for _, serverPool := range expectedCluster.ServerPools {
		if cloud, ok := CloudProviderCache[serverPool.Name]; ok {
			err := cloud.ApplyServerPool(serverPool)
			if err != nil {
				return fmt.Errorf("Failed applying server pool [%s]: %v", serverPool.Name, err)
			}
			logger.Info("Applied server pool [%s]", serverPool.Name)
		} else {
			fmt.Println(CloudProviderCache)
			fmt.Println(serverPool.Name)
			return fmt.Errorf("Fatal error, missing cloudprovider in cache for server pool [%s]", serverPool.Name)
		}
	}
	logger.Info("Apply for cluster [%s] complete", expectedCluster.Name)
	return nil
}

func CompareClusters(e, a *cluster.Cluster) (bool, error) {

	ebytes, err := yaml.Marshal(e)
	if err != nil {
		return false, fmt.Errorf("Unable to marshal (expected) YAML: %v", err)
	}
	abytes, err := yaml.Marshal(a)
	if err != nil {
		return false, fmt.Errorf("Unable to marshal (actual) YAML: %v", err)
	}

	logger.Debug("--------------- Expected ----------------")
	logger.Debug(fmt.Sprintf("%X", md5.Sum(ebytes)))
	logger.Debug("---------------- Actual -----------------")
	logger.Debug(fmt.Sprintf("%X", md5.Sum(abytes)))

	for i, ebyte := range ebytes {
		abyte := abytes[i]
		if ebyte != abyte {
			return false, nil
		}
	}

	return true, nil
}

type NewCloudFunc func() cloudprovider.CloudProvider

var CloudMap = map[string]NewCloudFunc{
	cluster.ServerPoolCloud_Amazon:    amazon.NewAmazonCloudProvider,
	cluster.ServerPoolCloud_Google:    google.NewGoogleCloudProvider,
	cluster.ServerPoolCloud_Azure:     azure.NewAzureCloudProvider,
	cluster.ServerPoolCloud_Baremetal: baremetal.NewBaremetalCloudProvider,
}

var CloudProviderCache = make(map[string]cloudprovider.CloudProvider)

func GetActualCluster(s stores.ClusterStorer, expectedCluster *cluster.Cluster) (*cluster.Cluster, error) {
	actualCluster := &cluster.Cluster{}
	actualCluster.Name = expectedCluster.Name
	for _, serverPool := range expectedCluster.ServerPools {
		var cloud cloudprovider.CloudProvider
		if f, ok := CloudMap[serverPool.Cloud]; ok {
			cloud = f()
			CloudProviderCache[serverPool.Name] = cloud
		} else {
			return actualCluster, fmt.Errorf("Unable to find cloud provider for cloud [%s]", serverPool.Cloud)
		}
		serverPool, err := cloud.GetServerPool()
		if err != nil {
			return actualCluster, fmt.Errorf("Unable to get server pool [%s] for cloud [%s]", serverPool.Name, serverPool.Cloud)
		}
		actualCluster.ServerPools = append(actualCluster.ServerPools, serverPool)

	}
	return actualCluster, nil
}
