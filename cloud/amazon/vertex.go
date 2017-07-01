package amazon

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
)

type AwsVertex struct {
	Children       []cloud.ResourceVertex
	Implementation cloud.ResourceImplementation
}

func (a AwsVertex) Add(vertex cloud.ResourceVertex) cloud.ResourceVertex {
	a.Children = append(a.Children, vertex)
	return a
}

func (a AwsVertex) RecursiveInit(known *cluster.Cluster, sdk cloud.Sdk) error {
	if err := a.Implementation.Init(known, sdk); err != nil {
		return err
	}
	for _, child := range a.Children {
		if err := child.(AwsVertex).RecursiveInit(known, sdk); err != nil {
			return err
		}
	}
	return nil
}

func (a AwsVertex) RecursiveFind() error {
	if err := a.Implementation.Find(); err != nil {
		return err
	}
	for _, child := range a.Children {
		if err := child.(AwsVertex).RecursiveFind(); err != nil {
			return err
		}
	}
	return nil
}

func (a AwsVertex) RecursiveRenderActual(actual *cluster.Cluster) error {
	return nil
}
