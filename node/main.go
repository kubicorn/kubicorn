package main

import "k8s.io/kube-deploy/cluster-api/deploy"

func main() {
	// Hard code these for now
	d := deploy.NewDeployer("aws", "/Users/knova/.kube/config")
	//d.CreateCluster()



	// Create Machine CRD

}

//func RunCreate(co *CreateOptions) error {
//
//	d := deploy.NewDeployer(provider, kubeConfig)
//
//	return d.CreateCluster(cluster, machines)
//}