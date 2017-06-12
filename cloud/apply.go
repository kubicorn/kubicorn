package cloud

import (
	"crypto/md5"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud/cloudprovider"
	"github.com/kris-nova/kubicorn/cloud/cloudprovider/amazon"
	"github.com/kris-nova/kubicorn/cloud/cloudprovider/azure"
	"github.com/kris-nova/kubicorn/cloud/cloudprovider/baremetal"
	"github.com/kris-nova/kubicorn/cloud/cloudprovider/google"
	"github.com/kris-nova/kubicorn/logger"
	"github.com/kris-nova/kubicorn/state/stores"
	"time"
)

const ApplyTimeoutSeconds = 100

func ApplyCluster(s stores.ClusterStorer, expectedCluster *cluster.Cluster) error {
	pairs, err := GetReconcilePairs(s, expectedCluster)
	if err != nil {
		return err
	}
	resourceCount := 0
	errchan := make(chan error)
	timeoutchan := make(chan int)
	go func() {
		time.Sleep(time.Second * time.Duration(ApplyTimeoutSeconds))
		timeoutchan <- 1
	}()
	for _, pair := range pairs {
		go pair.Provider.ConcurrentApplyResource(pair.Resource, errchan)
		resourceCount++
	}
	for resourceCount > 0 {
		select {
		case err := <-errchan:
			if err != nil {
				return fmt.Errorf("Concurrent Apply error: %v", err)
			}
			resourceCount--
		case <-timeoutchan:
			return fmt.Errorf("Timeout while waiting for concurrent apply of [%d] seconds", ApplyTimeoutSeconds)
		}
	}
	logger.Info("Apply for cluster [%s] complete", expectedCluster.Name)
	return nil
}

func CompareClusters(e, a interface{}) (bool, error) {

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

type NewCloudFunc func(expected *cluster.Cluster) cloudprovider.CloudProvider

var CloudMap = map[string]NewCloudFunc{
	cluster.ServerPoolCloud_Amazon:    amazon.NewAmazonCloudProvider,
	cluster.ServerPoolCloud_Google:    google.NewGoogleCloudProvider,
	cluster.ServerPoolCloud_Azure:     azure.NewAzureCloudProvider,
	cluster.ServerPoolCloud_Baremetal: baremetal.NewBaremetalCloudProvider,
}

var CloudProviderCache = make(map[string]cloudprovider.CloudProvider)

type CloudPair struct {
	Provider cloudprovider.CloudProvider
	Resource cloudprovider.CloudResource
}

func GetReconcilePairs(s stores.ClusterStorer, expectedCluster *cluster.Cluster) ([]*CloudPair, error) {
	var pairs []*CloudPair
	for _, expectedServerPool := range expectedCluster.ServerPools {
		var cloud cloudprovider.CloudProvider
		if f, ok := CloudMap[expectedServerPool.Cloud]; ok {
			cloud = f(expectedCluster)
		} else {
			return nil, fmt.Errorf("Unable to find cloud provider for cloud [%s]", expectedServerPool.Cloud)
		}
		resources := cloud.GetExpectedServerPoolResources(expectedServerPool)
		for _, resource := range resources {
			pairs = append(pairs, &CloudPair{
				Provider: cloud,
				Resource: resource,
			})
		}
	}
	return pairs, nil
}
