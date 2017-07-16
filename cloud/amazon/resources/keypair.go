package resources

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cutil/compare"
	"github.com/kris-nova/kubicorn/logger"
)

type KeyPair struct {
	Shared
	PublicKeyData        string
	PublicKeyPath        string
	PublicKeyFingerprint string
	User string
}

func (r *KeyPair) Actual(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("keypair.Actual")
	if r.CachedActual != nil {
		logger.Debug("Using cached keypair [actual]")
		return r.CachedActual, nil
	}
	actual := &KeyPair{
		Shared: Shared{
			Name:        r.Name,
			Tags:        make(map[string]string),
			TagResource: r.TagResource,
		},
	}

	if known.Ssh.Identifier != "" {
		input := &ec2.DescribeKeyPairsInput{
			KeyNames: []*string{&known.Ssh.Identifier},
		}
		output, err := Sdk.Ec2.DescribeKeyPairs(input)
		if err != nil {
			return nil, err
		}
		lsn := len(output.KeyPairs)
		if lsn != 1 {
			return nil, fmt.Errorf("Found [%d] Keypairs for ID [%s]", lsn, known.Ssh.Identifier)
		}
		keypair := output.KeyPairs[0]
		actual.CloudID = *keypair.KeyName
		actual.PublicKeyFingerprint = *keypair.KeyFingerprint
	}
	actual.User = known.Ssh.User
	r.CachedActual = actual
	return actual, nil
}

func (r *KeyPair) Expected(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("keypair.Expected")
	if r.CachedExpected != nil {
		logger.Debug("Using keypair subnet [expected]")
		return r.CachedExpected, nil
	}
	expected := &KeyPair{
		Shared: Shared{
			Tags: map[string]string{
				"Name":              r.Name,
				"KubernetesCluster": known.Name,
			},
			CloudID:     known.Ssh.Identifier,
			Name:        r.Name,
			TagResource: r.TagResource,
		},
		PublicKeyPath: known.Ssh.PublicKeyPath,
		PublicKeyData: string(known.Ssh.PublicKeyData),
		User: known.Ssh.User,
	}
	r.CachedExpected = expected
	return expected, nil
}
func (r *KeyPair) Apply(actual, expected cloud.Resource, applyCluster *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("keypair.Apply")
	applyResource := expected.(*KeyPair)
	isEqual, err := compare.IsEqual(actual.(*KeyPair), expected.(*KeyPair))
	if err != nil {
		return nil, err
	}
	if isEqual {
		return applyResource, nil
	}
	input := &ec2.ImportKeyPairInput{
		KeyName:           &expected.(*KeyPair).Name,
		PublicKeyMaterial: []byte(expected.(*KeyPair).PublicKeyData),
	}
	output, err := Sdk.Ec2.ImportKeyPair(input)
	if err != nil {
		return nil, err
	}
	logger.Info("Created KeyPair [%s]", *output.KeyName)

	newResource := &KeyPair{}
	newResource.PublicKeyData = expected.(*KeyPair).PublicKeyData
	newResource.PublicKeyPath = expected.(*KeyPair).PublicKeyPath
	newResource.User = expected.(*KeyPair).User
	newResource.PublicKeyFingerprint = *output.KeyFingerprint
	newResource.CloudID = expected.(*KeyPair).Name
	newResource.Name = expected.(*KeyPair).Name
	return newResource, nil
}
func (r *KeyPair) Delete(actual cloud.Resource, known *cluster.Cluster) error {
	logger.Debug("keypair.Delete")
	deleteResource := actual.(*KeyPair)
	if deleteResource.CloudID == "" {
		return fmt.Errorf("Unable to delete keypair resource without ID [%s]", deleteResource.Name)
	}

	input := &ec2.DeleteKeyPairInput{
		KeyName: &actual.(*KeyPair).Name,
	}
	_, err := Sdk.Ec2.DeleteKeyPair(input)
	if err != nil {
		return err
	}
	logger.Info("Deleted keypair [%s]", actual.(*KeyPair).CloudID)
	return nil
}

func (r *KeyPair) Render(renderResource cloud.Resource, renderCluster *cluster.Cluster) (*cluster.Cluster, error) {
	logger.Debug("keypair.Render")
	renderCluster.Ssh.Name = renderResource.(*KeyPair).Name
	renderCluster.Ssh.Identifier = renderResource.(*KeyPair).Name
	renderCluster.Ssh.PublicKeyData = []byte(renderResource.(*KeyPair).PublicKeyData)
	renderCluster.Ssh.PublicKeyFingerprint = renderResource.(*KeyPair).PublicKeyFingerprint
	renderCluster.Ssh.PublicKeyPath = renderResource.(*KeyPair).PublicKeyPath
	renderCluster.Ssh.User = renderResource.(*KeyPair).User
	return renderCluster, nil
}

func (r *KeyPair) Tag(tags map[string]string) error {
	// Todo tag on another resource
	return nil
}
