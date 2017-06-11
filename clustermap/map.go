package clustermap

import "github.com/kris-nova/kubicorn/apis/cluster"

type ClusterMapFunc func(name string) *cluster.Cluster

var ClusterMaps = map[string]ClusterMapFunc{
	"baremetal": NewSimpleBareMetal,
	"amazon":    NewSimpleAmazonCluster,
	"azure":     NewSimpleAzureCluster,
	"google":    NewSimpleGoogleCluster,
}
