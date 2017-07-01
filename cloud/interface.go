package cloud

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
)

type Reconciler interface {
	GetActual(known *cluster.Cluster) (*cluster.Cluster, error)
	GetExpected(known *cluster.Cluster) (*cluster.Cluster, error)
	Reconcile(actual, expected *cluster.Cluster) error
}

type Sdk interface {
}

type ResourceGraph interface {
	WalkInit(known *cluster.Cluster) error
	WalkFind() error
	RenderActual() (*cluster.Cluster, error)
}

type ResourceVertex interface {
	Add(ResourceVertex) ResourceVertex
	RecursiveInit(known *cluster.Cluster, sdk Sdk) error
	RecursiveFind() error
	RecursiveRenderActual(actual *cluster.Cluster) error
}

type ResourceImplementation interface {
	Init(known *cluster.Cluster, sdk Sdk) error
	Apply() error
	Find() error
	RenderActual(actual *cluster.Cluster) error
}
