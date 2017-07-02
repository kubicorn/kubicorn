package resources

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud/amazon/awsSdkGo"
	"github.com/kris-nova/kubicorn/cutil/compare"
	"github.com/kris-nova/kubicorn/logger"
)

type Vpc struct {
	Resource
	Actual         *Vpc
	Expected       *Vpc
	CIDR           string
	Tags           map[string]string
	AssociatedAsgs []*Asg
}

func (r *Vpc) Parse() error {
	actual := &Vpc{
		Tags: make(map[string]string),
	}
	expected := &Vpc{
		Tags: make(map[string]string),
	}
	var vpc *ec2.Vpc
	if r.KnownCluster.Network.Identifier != "" {
		input := &ec2.DescribeVpcsInput{
			VpcIds: []*string{&r.KnownCluster.Network.Identifier},
		}
		output, err := r.AwsSdk.Ec2.DescribeVpcs(input)
		if err != nil {
			return err
		}
		lvpc := len(output.Vpcs)
		if lvpc == 1 {
			logger.Debug("Found %s [%s]", r.Type, r.KnownCluster.Network.Identifier)
			vpc = output.Vpcs[0]
			actual.ID = *vpc.VpcId
			actual.CIDR = *vpc.CidrBlock
			for _, tag := range vpc.Tags {
				actual.Tags[*tag.Key] = *tag.Value
				expected.Tags[*tag.Key] = *tag.Value // Always preserve existing tags
			}
		} else if lvpc > 1 {
			return fmt.Errorf("Found more than 1 %s for label [%s] %s", r.Type, r.Label, r.Name)
		} else if lvpc < 1 {
			return fmt.Errorf("Unable to lookup VPC [%s]", r.KnownCluster.Network.Identifier)
		}
	}
	if r.KnownCluster.Network.Identifier != "" {
		expected.ID = r.KnownCluster.Network.Identifier
	}
	expected.CIDR = r.KnownCluster.Network.CIDR
	expected.Tags[r.Label] = r.Name
	expected.Tags["Name"] = r.Name
	r.Expected = expected
	r.Actual = actual

	for _, serverPool := range r.KnownCluster.ServerPools {
		asg := &Asg{
			AssociatedServerPool: serverPool,
		}
		asg.Init(r.KnownCluster, r.AwsSdk)
		err := asg.Parse()
		if err != nil {
			return fmt.Errorf("Unable to add associated ASG [%s]: %v", serverPool.Name, err)
		}
		r.Expected.AssociatedAsgs = append(r.Expected.AssociatedAsgs, asg)
	}

	return nil
}

func (r *Vpc) Apply() error {
	logger.Debug("Compare: [%s]", r.Type)
	if !compare.Compare(r.Actual, r.Expected) {
		{
			input := &ec2.CreateVpcInput{
				CidrBlock: &r.Expected.CIDR,
			}
			output, err := r.AwsSdk.Ec2.CreateVpc(input)
			if err != nil {
				return fmt.Errorf("Unable to create new VPC: %v", err)
			}
			r.Actual.ID = *output.Vpc.VpcId
			r.Actual.CIDR = *output.Vpc.CidrBlock
			logger.Info("Created new VPC [%s]", r.Actual.ID)
		}
		{
			input := &ec2.CreateTagsInput{
				Resources: []*string{&r.Actual.ID},
			}
			for key, val := range r.Expected.Tags {
				logger.Debug("Registering tag [%s] %s", key, val)
				input.Tags = append(input.Tags, &ec2.Tag{
					Key:   S("%s", key),
					Value: S("%s", val),
				})
			}
			_, err := r.AwsSdk.Ec2.CreateTags(input)
			if err != nil {
				return fmt.Errorf("Unable to tag new VPC: %v", err)
			}
		}
	} else {
		logger.Info("Unchanged resource: [%s]", r.Type)
	}

	for _, expectedAsg := range r.Expected.AssociatedAsgs {
		err := expectedAsg.Apply()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Vpc) Init(known, actual, expected *cluster.Cluster, sdk *awsSdkGo.Sdk) error {
	r.Type = "vpc"
	r.Label = "kubicorn_vpc_name"
	r.Name = known.Name
	r.KnownCluster = known
	r.ActualCluster = actual
	r.ExpectedCluster = expected
	r.Tags = make(map[string]string)
	r.AwsSdk = sdk
	logger.Debug("Loading AWS Resource [%s]", r.Type)
	return nil
}

func (r *Vpc) Render() error {
	r.ExpectedCluster.Network = &cluster.Network{
		Identifier: r.Expected.ID,
		CIDR:       r.Expected.CIDR,
	}
	r.ExpectedCluster.Name = r.KnownCluster.Name
	r.ExpectedCluster.Location = r.KnownCluster.Location
	r.ExpectedCluster.Cloud = r.KnownCluster.Cloud

	for _, expectedAsg := range r.Expected.AssociatedAsgs {
		err := expectedAsg.Render()
		if err != nil {
			return err
		}

	}

	r.ActualCluster.Network = &cluster.Network{
		Identifier: r.Actual.ID,
		CIDR:       r.Actual.CIDR,
	}
	r.ActualCluster.Name = r.KnownCluster.Name
	r.ActualCluster.Location = r.KnownCluster.Location
	r.ActualCluster.Cloud = r.KnownCluster.Cloud

	return nil
}

func (r *Vpc) Delete() error {
	input := &ec2.DeleteVpcInput{
		VpcId: &r.ActualCluster.Network.Identifier,
	}
	_, err := r.AwsSdk.Ec2.DeleteVpc(input)
	if err != nil {
		return err
	}
	logger.Info("Destroy resource: [%s] %s", r.ActualCluster.Network.Identifier, r.Type)

	for _, expectedAsg := range r.Expected.AssociatedAsgs {
		err := expectedAsg.Delete()
		if err != nil {
			return err
		}

	}

	return nil
}
