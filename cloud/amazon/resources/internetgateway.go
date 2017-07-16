package resources

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cutil/compare"
	"github.com/kris-nova/kubicorn/logger"
)

type InternetGateway struct {
	Shared
}

func (r *InternetGateway) Actual(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("internetgateway.Actual")
	if r.CachedActual != nil {
		logger.Debug("Using cached internetgateway [actual]")
		return r.CachedActual, nil
	}
	actual := &InternetGateway{
		Shared: Shared{
			Name:        r.Name,
			Tags:        make(map[string]string),
			TagResource: r.TagResource,
		},
	}
	if known.Network.Identifier != "" {
		input := &ec2.DescribeInternetGatewaysInput{
			Filters: []*ec2.Filter{
				{
					Name:   S("tag:kubicorn-internet-gateway-name"),
					Values: []*string{S(r.Name)},
				},
			},
		}
		output, err := Sdk.Ec2.DescribeInternetGateways(input)
		if err != nil {
			return nil, err
		}
		lsn := len(output.InternetGateways)
		if lsn != 1 {
			return nil, fmt.Errorf("Found [%d] Internet Gateways for ID [%s]", lsn, known.Network.Identifier)
		}
		ig := output.InternetGateways[0]
		for _, tag := range ig.Tags {
			key := *tag.Key
			val := *tag.Value
			actual.Tags[key] = val
		}
		actual.CloudID = r.Name
	}
	r.CachedActual = actual
	return actual, nil
}

func (r *InternetGateway) Expected(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("internetgateway.Expected")
	if r.CachedExpected != nil {
		logger.Debug("Using internetgateway subnet [expected]")
		return r.CachedExpected, nil
	}
	expected := &InternetGateway{
		Shared: Shared{
			Tags: map[string]string{
				"Name":                           r.Name,
				"KubernetesCluster":              known.Name,
				"kubicorn-internet-gateway-name": r.Name,
			},
			CloudID:     r.Name,
			Name:        r.Name,
			TagResource: r.TagResource,
		},
	}
	r.CachedExpected = expected
	return expected, nil
}
func (r *InternetGateway) Apply(actual, expected cloud.Resource, applyCluster *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("internetgateway.Apply")
	applyResource := expected.(*InternetGateway)
	isEqual, err := compare.IsEqual(actual.(*InternetGateway), expected.(*InternetGateway))
	if err != nil {
		return nil, err
	}
	if isEqual {
		return applyResource, nil
	}
	input := &ec2.CreateInternetGatewayInput{}
	output, err := Sdk.Ec2.CreateInternetGateway(input)
	if err != nil {
		return nil, err
	}
	logger.Info("Created Internet Gateway [%s]", *output.InternetGateway.InternetGatewayId)
	ig := output.InternetGateway

	// --- Attach Internet Gateway to VPC
	atchinput := &ec2.AttachInternetGatewayInput{
		InternetGatewayId: ig.InternetGatewayId,
		VpcId:             &applyCluster.Network.Identifier,
	}
	_, err = Sdk.Ec2.AttachInternetGateway(atchinput)
	if err != nil {
		return nil, err
	}
	logger.Info("Attaching Internet Gateway [%s] to VPC [%s]", *ig.InternetGatewayId, applyCluster.Network.Identifier)
	newResource := &InternetGateway{
		Shared: Shared{
			Tags: make(map[string]string),
		},
	}
	newResource.CloudID = expected.(*InternetGateway).Name
	newResource.Name = expected.(*InternetGateway).Name
	for key, value := range expected.(*InternetGateway).Tags {
		newResource.Tags[key] = value
	}
	expected.(*InternetGateway).CloudID = *output.InternetGateway.InternetGatewayId
	err = expected.Tag(expected.(*InternetGateway).Tags)
	if err != nil {
		return nil, err
	}
	return newResource, nil
}

func (r *InternetGateway) Delete(actual cloud.Resource, known *cluster.Cluster) error {
	logger.Debug("internetgateway.Delete")
	deleteResource := actual.(*InternetGateway)
	if deleteResource.CloudID == "" {
		return fmt.Errorf("Unable to delete internetgateway resource without ID [%s]", deleteResource.Name)
	}

	input := &ec2.DescribeInternetGatewaysInput{
		Filters: []*ec2.Filter{
			{
				Name:   S("tag:kubicorn-internet-gateway-name"),
				Values: []*string{S(r.Name)},
			},
		},
	}
	output, err := Sdk.Ec2.DescribeInternetGateways(input)
	if err != nil {
		return err
	}
	lsn := len(output.InternetGateways)
	if lsn == 0 {
		return nil
	}
	if lsn != 1 {
		return fmt.Errorf("Found [%d] Internet Gateways for ID [%s]", lsn, r.Name)
	}
	ig := output.InternetGateways[0]

	detinput := &ec2.DetachInternetGatewayInput{
		InternetGatewayId: ig.InternetGatewayId,
		VpcId:             &known.Network.Identifier,
	}
	_, err = Sdk.Ec2.DetachInternetGateway(detinput)
	if err != nil {
		return err
	}

	delinput := &ec2.DeleteInternetGatewayInput{
		InternetGatewayId: ig.InternetGatewayId,
	}
	_, err = Sdk.Ec2.DeleteInternetGateway(delinput)
	if err != nil {
		return err
	}
	logger.Info("Deleted internetgateway [%s]", actual.(*InternetGateway).CloudID)
	return nil
}

func (r *InternetGateway) Render(renderResource cloud.Resource, renderCluster *cluster.Cluster) (*cluster.Cluster, error) {
	logger.Debug("internetgateway.Render")
	return renderCluster, nil
}

func (r *InternetGateway) Tag(tags map[string]string) error {
	logger.Debug("internetgateway.Tag")
	tagInput := &ec2.CreateTagsInput{
		Resources: []*string{&r.CloudID},
	}
	for key, val := range tags {
		logger.Debug("Registering Internet Gateway tag [%s] %s", key, val)
		tagInput.Tags = append(tagInput.Tags, &ec2.Tag{
			Key:   S("%s", key),
			Value: S("%s", val),
		})
	}
	_, err := Sdk.Ec2.CreateTags(tagInput)
	if err != nil {
		return err
	}
	return nil
}
