package resources

import (
	"context"
	"fmt"
	"github.com/digitalocean/godo"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/bootstrap"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cutil/compare"
	"github.com/kris-nova/kubicorn/logger"
	"strconv"
)

type SSH struct {
	Shared
	Fingerprint   string
	PublicKeyData string
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
		actual.Fingerprint = ssh.Fingerprint
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
		Fingerprint:   known.Ssh.PublicKeyFingerprint,
		PublicKeyData: string(known.Ssh.PublicKeyData),
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
	var userData []byte
	userData, err = bootstrap.Asset(fmt.Sprintf("bootstrap/%s", r.ServerPool.BootstrapScript))
	if err != nil {
		return nil, err
	}

	//fmt.Println(string(userData))
	applyCluster.Values.ItemMap["INJECTEDPORT"] = applyCluster.KubernetesApi.Port
	userData, err = bootstrap.Inject(userData, applyCluster.Values.ItemMap)
	if err != nil {
		return nil, err
	}

	createRequest := &godo.DropletCreateRequest{
		Name:   expected.(*SSH).Name,
		Region: expected.(*SSH).Region,
		Size:   expected.(*SSH).Size,
		Image: godo.DropletCreateImage{
			Slug: expected.(*SSH).Image,
		},
		Tags:              []string{expected.(*SSH).Name},
		PrivateNetworking: true,
		SSHKeys: []godo.DropletCreateSSHKey{
			{
				Fingerprint: expected.(*SSH).SShFingerprint,
			},
		},
		UserData: string(userData),
	}
	droplet, _, err := Sdk.Client.Droplets.Create(context.TODO(), createRequest)
	if err != nil {
		return nil, err
	}

	logger.Info("Created Droplet [%d]", droplet.ID)
	id := strconv.Itoa(droplet.ID)
	newResource := &Droplet{
		Shared: Shared{
			Name:    droplet.Name,
			CloudID: id,
		},
		Image:  droplet.Image.Slug,
		Size:   droplet.Size.Slug,
		Region: droplet.Region.Name,
		Count:  expected.(*SSH).Count,
	}
	return newResource, nil
}
func (r *SSH) Delete(actual cloud.Resource, known *cluster.Cluster) error {
	logger.Debug("droplet.Delete")
	deleteResource := actual.(*SSH)
	if deleteResource.Name == "" {
		return fmt.Errorf("Unable to delete droplet resource without Name [%s]", deleteResource.Name)
	}

	droplets, _, err := Sdk.Client.Droplets.ListByTag(context.TODO(), r.Name, &godo.ListOptions{})
	if err != nil {
		return err
	}
	ld := len(droplets)
	if ld != 1 {
		return fmt.Errorf("Found [%d] Droplets for Name [%s]", ld, r.Name)
	}
	droplet := droplets[0]
	_, err = Sdk.Client.Droplets.Delete(context.TODO(), droplet.ID)
	if err != nil {
		return err
	}
	logger.Info("Deleted Droplet [%d]", droplet.ID)
	return nil
}

func (r *SSH) Render(renderResource cloud.Resource, renderCluster *cluster.Cluster) (*cluster.Cluster, error) {
	logger.Debug("droplet.Render")

	serverPool := &cluster.ServerPool{}
	serverPool.Image = renderResource.(*SSH).Image
	serverPool.Size = renderResource.(*SSH).Size
	serverPool.Name = renderResource.(*SSH).Name
	serverPool.MaxCount = renderResource.(*SSH).Count
	found := false
	for i := 0; i < len(renderCluster.ServerPools); i++ {
		if renderCluster.ServerPools[i].Name == renderResource.(*SSH).Name {
			renderCluster.ServerPools[i].Image = renderResource.(*SSH).Image
			renderCluster.ServerPools[i].Size = renderResource.(*SSH).Size
			renderCluster.ServerPools[i].MaxCount = renderResource.(*SSH).Count
			found = true
		}
	}
	if !found {
		renderCluster.ServerPools = append(renderCluster.ServerPools, serverPool)
	}
	renderCluster.Location = renderResource.(*SSH).Region
	return renderCluster, nil
}

func (r *SSH) Tag(tags map[string]string) error {
	return nil
}
