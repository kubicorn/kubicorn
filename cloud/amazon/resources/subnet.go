package resources

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cutil/compare"
	"github.com/kris-nova/kubicorn/logger"
)

type Subnet struct {
	Shared
	ClusterSubnet *cluster.Subnet
	ServerPool    *cluster.ServerPool
	CIDR          string
	VpcId         string
	Zone          string
}

func (r *Subnet) Actual(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("subnet.Actual")
	if r.CachedActual != nil {
		logger.Debug("Using cached subnet [actual]")
		return r.CachedActual, nil
	}
	actual := &Subnet{
		Shared: Shared{
			Name:        r.Name,
			Tags:        make(map[string]string),
			TagResource: r.TagResource,
		},
	}

	if r.ClusterSubnet.Identifier != "" {
		input := &ec2.DescribeSubnetsInput{
			SubnetIds: []*string{S(r.ClusterSubnet.Identifier)},
		}
		output, err := Sdk.Ec2.DescribeSubnets(input)
		if err != nil {
			return nil, err
		}
		lsn := len(output.Subnets)
		if lsn != 1 {
			return nil, fmt.Errorf("Found [%d] Subnets for ID [%s]", lsn, r.ClusterSubnet.Identifier)
		}
		subnet := output.Subnets[0]
		actual.CIDR = *subnet.CidrBlock
		actual.CloudID = *subnet.SubnetId
		actual.VpcId = *subnet.VpcId
		actual.Zone = *subnet.AvailabilityZone
		for _, tag := range subnet.Tags {
			key := *tag.Key
			val := *tag.Value
			actual.Tags[key] = val
		}

	}
	r.CachedActual = actual
	return actual, nil
}

func (r *Subnet) Expected(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("subnet.Expected")
	if r.CachedExpected != nil {
		logger.Debug("Using cached subnet [expected]")
		return r.CachedExpected, nil
	}
	expected := &Subnet{
		Shared: Shared{
			Tags: map[string]string{
				"Name":              r.Name,
				"KubernetesCluster": known.Name,
			},
			CloudID:     r.ClusterSubnet.Identifier,
			Name:        r.Name,
			TagResource: r.TagResource,
		},
		CIDR:  r.ClusterSubnet.CIDR,
		VpcId: known.Network.Identifier,
		Zone:  r.ClusterSubnet.Zone,
	}
	r.CachedExpected = expected
	return expected, nil
}
func (r *Subnet) Apply(actual, expected cloud.Resource, applyCluster *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("subnet.Apply")
	applyResource := expected.(*Subnet)
	isEqual, err := compare.IsEqual(actual.(*Subnet), expected.(*Subnet))
	if err != nil {
		return nil, err
	}
	if isEqual {
		return applyResource, nil
	}
	input := &ec2.CreateSubnetInput{
		CidrBlock:        &expected.(*Subnet).CIDR,
		VpcId:            &applyCluster.Network.Identifier,
		AvailabilityZone: &expected.(*Subnet).Zone,
	}
	output, err := Sdk.Ec2.CreateSubnet(input)
	if err != nil {
		return nil, err
	}
	logger.Info("Created Subnet [%s]", *output.Subnet.SubnetId)
	newResource := &Subnet{}
	newResource.CIDR = *output.Subnet.CidrBlock
	newResource.VpcId = *output.Subnet.VpcId
	newResource.Zone = *output.Subnet.AvailabilityZone
	newResource.Name = applyResource.Name
	newResource.CloudID = *output.Subnet.SubnetId
	return newResource, nil
}
func (r *Subnet) Delete(actual cloud.Resource, known *cluster.Cluster) error {
	logger.Debug("subnet.Delete")
	deleteResource := actual.(*Subnet)
	if deleteResource.CloudID == "" {
		return fmt.Errorf("Unable to delete subnet resource without ID [%s]", deleteResource.Name)
	}

	input := &ec2.DeleteSubnetInput{
		SubnetId: &actual.(*Subnet).CloudID,
	}
	_, err := Sdk.Ec2.DeleteSubnet(input)
	if err != nil {
		return err
	}
	logger.Info("Deleted subnet [%s]", actual.(*Subnet).CloudID)
	return nil
}

func (r *Subnet) Render(renderResource cloud.Resource, renderCluster *cluster.Cluster) (*cluster.Cluster, error) {
	logger.Debug("subnet.Render")
	subnet := &cluster.Subnet{}
	subnet.CIDR = renderResource.(*Subnet).CIDR
	subnet.Zone = renderResource.(*Subnet).Zone
	subnet.Name = renderResource.(*Subnet).Name
	subnet.Identifier = renderResource.(*Subnet).CloudID
	found := false

	// Check if exists. Period.
	for i := 0; i < len(renderCluster.ServerPools); i++ {
		for j := 0; j < len(renderCluster.ServerPools[i].Subnets); j++ {
			if renderCluster.ServerPools[i].Subnets[j].Name == renderResource.(*Subnet).Name {
				renderCluster.ServerPools[i].Subnets[j].CIDR = renderResource.(*Subnet).CIDR
				renderCluster.ServerPools[i].Subnets[j].Zone = renderResource.(*Subnet).Zone
				renderCluster.ServerPools[i].Subnets[j].Identifier = renderResource.(*Subnet).CloudID
				found = true
			}
		}
	}

	// Check if server pool exists.
	if !found {
		for i := 0; i < len(renderCluster.ServerPools); i++ {
			if renderCluster.ServerPools[i].Name == renderResource.(*Subnet).Name {
				renderCluster.ServerPools[i].Subnets = append(renderCluster.ServerPools[i].Subnets, subnet)
				found = true
			}
		}
	}

	// Create a new Server pool
	if !found {
		renderCluster.ServerPools = append(renderCluster.ServerPools, &cluster.ServerPool{
			Name: renderResource.(*Subnet).Name,
			Subnets: []*cluster.Subnet{
				subnet,
			},
		})
	}
	return renderCluster, nil
}

func (r *Subnet) Tag(tags map[string]string) error {
	// Todo tag on another resource
	return nil
}
