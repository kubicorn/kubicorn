package amazon

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cloud/amazon/awsSdkGo"
	"github.com/kris-nova/kubicorn/cloud/amazon/resources"
)

type AwsResourceGraph struct {
	Root cloud.ResourceVertex
}

func (g *AwsResourceGraph) WalkInit(known *cluster.Cluster) error {
	sdk, err := awsSdkGo.NewSdk("us-west-2")
	if err != nil {
		return nil
	}
	return g.Root.(AwsVertex).RecursiveInit(known, sdk)
}

func (g *AwsResourceGraph) WalkFind() error {
	return g.Root.(AwsVertex).RecursiveFind()
}

func (g *AwsResourceGraph) RenderActual() (*cluster.Cluster, error) {
	actual := &cluster.Cluster{}
	if err := g.Root.RecursiveRenderActual(actual); err != nil {
		return nil, err
	}
	return actual, nil
}

var Graph = &AwsResourceGraph{
	Root: AwsVertex{
		Implementation: &resources.Vpc{},
	}.Add(AwsVertex{
		Implementation: &resources.Asg{},
	}),
}
