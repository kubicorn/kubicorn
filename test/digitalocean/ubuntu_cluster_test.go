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
	logger.TestMode = true
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
	if err != nil {
		fmt.Printf("Unable to create digital ocean test cluster: %v\n", err)
		os.Exit(1)
	}
	status := m.Run()
	exitCode := 0
	if status != 0 {
		fmt.Printf("-----------------------------------------------------------------------\n")
		fmt.Printf("[FAILURE]\n")
		fmt.Printf("-----------------------------------------------------------------------\n")
		exitCode = 1
	}
	_, err = test.Delete(testCluster)
	if err != nil {
		exitCode = 99
		fmt.Println("Failure cleaning up cluster! Abandoned resources!")
	}
	os.Exit(exitCode)
}

func TestApiListen(t *testing.T) {
	_, err := network.AssertTcpSocketAcceptsConnection(fmt.Sprintf("%s:%s", testCluster.KubernetesApi.Endpoint, testCluster.KubernetesApi.Port), "opening a new socket connection against the Kubernetes API")
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
}
