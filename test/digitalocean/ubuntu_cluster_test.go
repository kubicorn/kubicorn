package digitalocean

import (
	"fmt"
	"github.com/kris-nova/charlie/network"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/logger"
	"github.com/kris-nova/kubicorn/profiles"
	"github.com/kris-nova/kubicorn/test"
	"os"
	"testing"
)

var testCluster *cluster.Cluster

func TestMain(m *testing.M) {
	var err error
	//func() {
	//	for {
	//		if r := recover(); r != nil {
	//			logger.Critical("Panic: %v", r)
	//		}
	//	}
	//}()
	testCluster = profiles.NewSimpleDigitalOceanCluster("ubuntu-test")
	testCluster, err = test.Create(testCluster)
	defer func() {
		_, err := test.Delete(testCluster)
		if err != nil {
			logger.Critical(err.Error())
		}
	}()
	if err != nil {
		logger.Critical("Unable to create digital ocean test cluster: %v", err)
		os.Exit(1)
	}
	status := m.Run()
	if status != 0 {
		os.Exit(1)
	}
}

func TestApiListen(t *testing.T) {
	_, err := network.AssertTcpSocketAcceptsConnection(fmt.Sprintf("%s:%s", testCluster.KubernetesApi.Endpoint, testCluster.KubernetesApi.Port), "opening a new socket connection against the Kubernetes API")
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
}
