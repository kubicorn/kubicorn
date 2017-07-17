package resources

import (
	"context"
	"fmt"
	"github.com/digitalocean/godo"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cutil/compare"
	"github.com/kris-nova/kubicorn/logger"
	"strconv"
)

type SSH struct {
	Shared
	PublicKeyFingerprint string
	PublicKeyData        string
	PublicKeyPath        string
}

func (r *SSH) Actual(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("ssh.Actual")
	if r.CachedActual != nil {
		logger.Debug("Using cached ssh [actual]")
		return r.CachedActual, nil
	}
	actual := &SSH{
		Shared: Shared{
			Name:    r.Name,
			CloudID: known.Ssh.Identifier,
		},
	}

	if r.CloudID != "" {

		id, err := strconv.Atoi(r.CloudID)
		if err != nil {
			return nil, err
		}
		ssh, _, err := Sdk.Client.Keys.GetByID(context.TODO(), id)
		if err != nil {
			return nil, err
		}
		strid := strconv.Itoa(ssh.ID)
		actual.Name = ssh.Name
		actual.CloudID = strid
		actual.PublicKeyFingerprint = ssh.Fingerprint
		actual.PublicKeyPath = known.Ssh.PublicKeyPath
		actual.PublicKeyData = ssh.PublicKey
	}
	r.CachedActual = actual
	return actual, nil
}

func (r *SSH) Expected(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("ssh.Expected")
	if r.CachedExpected != nil {
		logger.Debug("Using cached ssh [expected]")
		return r.CachedExpected, nil
	}
	expected := &SSH{
		Shared: Shared{
			Name:    r.Name,
			CloudID: known.Ssh.Identifier,
		},
		PublicKeyFingerprint: known.Ssh.PublicKeyFingerprint,
		PublicKeyData:        string(known.Ssh.PublicKeyData),
		PublicKeyPath:        known.Ssh.PublicKeyPath,
	}
	r.CachedExpected = expected
	return expected, nil
}

func (r *SSH) Apply(actual, expected cloud.Resource, applyCluster *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("droplet.Apply")
	applyResource := expected.(*SSH)
	isEqual, err := compare.IsEqual(actual.(*SSH), expected.(*SSH))
	if err != nil {
		return nil, err
	}
	if isEqual {
		return applyResource, nil
	}
	request := &godo.KeyCreateRequest{
		Name:      expected.(*SSH).Name,
		PublicKey: expected.(*SSH).PublicKeyData,
	}
	key, _, err := Sdk.Client.Keys.Create(context.TODO(), request)
	if err != nil {
		return nil, err
	}

	logger.Info("Created SSH Key [%d]", key.ID)
	id := strconv.Itoa(key.ID)
	newResource := &SSH{
		Shared: Shared{
			Name:    key.Name,
			CloudID: id,
		},
		PublicKeyFingerprint: key.Fingerprint,
		PublicKeyData:        key.PublicKey,
		PublicKeyPath:        expected.(*SSH).PublicKeyPath,
	}
	return newResource, nil
}
func (r *SSH) Delete(actual cloud.Resource, known *cluster.Cluster) error {
	logger.Debug("ssh.Delete")
	deleteResource := actual.(*SSH)
	if deleteResource.CloudID == "" {
		return fmt.Errorf("Unable to delete ssh resource without Id [%s]", deleteResource.Name)
	}
	id, err := strconv.Atoi(r.CloudID)
	if err != nil {
		return err
	}

	_, err = Sdk.Client.Keys.DeleteByID(context.TODO(), id)
	if err != nil {
		return err
	}

	logger.Info("Deleted SSH Key [%d]", id)
	return nil
}

func (r *SSH) Render(renderResource cloud.Resource, renderCluster *cluster.Cluster) (*cluster.Cluster, error) {
	logger.Debug("ssh.Render")
	renderCluster.Ssh.PublicKeyData = []byte(renderResource.(*SSH).PublicKeyData)
	renderCluster.Ssh.PublicKeyFingerprint = renderResource.(*SSH).PublicKeyFingerprint
	renderCluster.Ssh.PublicKeyPath = renderResource.(*SSH).PublicKeyPath
	renderCluster.Ssh.Identifier = renderResource.(*SSH).CloudID
	return renderCluster, nil
}

func (r *SSH) Tag(tags map[string]string) error {
	return nil
}
