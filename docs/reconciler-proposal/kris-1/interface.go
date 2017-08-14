package kris_1


type Reconciler interface {
	Init() error
	GetActual() (*cluster.Cluster, error)
	GetExpected() (*cluster.Cluster, error)
	Reconcile(actualCluster, expectedCluster *cluster.Cluster) (*cluster.Cluster, error)
	Destroy() (*cluster.Cluster, error)
}

type Resource interface {
	Actual(known *cluster.Cluster) (Resource, error)
	Expected(known *cluster.Cluster) (Resource, error)
	Apply(actual, expected Resource, expectedCluster *cluster.Cluster) (Resource, error)
	Delete(actual Resource, known *cluster.Cluster) (Resource, error)
	Render(renderResource Resource, renderCluster *cluster.Cluster) (*cluster.Cluster, error)
	Tag(tags map[string]string) error
}
