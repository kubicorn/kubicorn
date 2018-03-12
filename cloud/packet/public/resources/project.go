// Copyright © 2017 The Kubicorn Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"fmt"
	"os"
	"strings"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/packethost/packngo"
)

var _ cloud.Resource = &Project{}

type Project struct {
	Shared
}

func (r *Project) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("project.Actual")
	newResource := &Project{
		Shared: Shared{
			Name: r.Name,
		},
	}

	if immutable.ProviderConfig().Project == nil || immutable.ProviderConfig().Project.Name == "" {
		return nil, nil, fmt.Errorf("Cannot work with empty project")
	}
	logger.Debug("project.Actual searching for project %s", immutable.ProviderConfig().Project.Name)
	project, err := GetProjectByName(immutable.ProviderConfig().Project.Name)
	if err != nil {
		return nil, nil, err
	}
	var id string
	if project != nil {
		id = project.ID
	}
	logger.Debug("project.Actual found? [%s]", id)
	newResource.Identifier = id

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Project) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("project.Expected")
	newResource := &Project{
		Shared: Shared{
			Name: r.Name,
		},
	}
	logger.Debug("project.Expected newResource %v", newResource)
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Project) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("project.Apply")
	expectedResource := expected.(*Project)
	actualResource := actual.(*Project)
	logger.Debug("project.Apply expectedResource vs actualResource %v %v", *expectedResource, *actualResource)
	if expectedResource.Name == actualResource.Name && actualResource.Identifier != "" {
		logger.Debug("already equal")
		newCluster := r.immutableRender(actualResource, immutable)
		logger.Debug("newCluster.Project %v", newCluster.ProviderConfig().Project)
		return newCluster, actualResource, nil
	}

	// if we made it here, we do not have that key, so create it
	request := &packngo.ProjectCreateRequest{
		Name: expected.(*Project).Name,
	}
	project, _, err := Sdk.Client.Projects.Create(request)
	if err != nil {
		// really should check if key already is in use
		return nil, nil, err
	}
	logger.Success("Created Project [%s]", project.ID)

	newResource := &Project{
		Shared: Shared{
			Name:       r.Name,
			Identifier: project.ID,
		},
	}
	logger.Debug("project.Apply newResource %v", newResource)
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Project) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("project.Delete")
	force := strings.ToLower(os.Getenv("KUBICORN_FORCE_DELETE_PROJECT"))
	logger.Debug("project.Delete force %s", force)
	if force == "true" {
		deleteResource := actual.(*Project)
		if deleteResource.Identifier == "" {
			return nil, nil, fmt.Errorf("Unable to delete project resource without ID [%s]", deleteResource.Name)
		}
		logger.Debug("project.Delete deleting project %s", deleteResource.Identifier)
		_, err := Sdk.Client.Projects.Delete(deleteResource.Identifier)
		if err != nil {
			return nil, nil, err
		}
		logger.Success("Deleted Project [%s]", deleteResource.Identifier)
	}

	newResource := &Project{}
	newResource.Name = actual.(*Project).Name

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Project) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("project.Render %v", newResource)
	newCluster := inaccurateCluster
	providerConfig := newCluster.ProviderConfig()
	providerConfig.Project = &cluster.Project{
		Name:       newResource.(*Project).Name,
		Identifier: newResource.(*Project).Identifier,
	}
	newCluster.SetProviderConfig(providerConfig)
	return newCluster
}

func GetProjectByName(name string) (*packngo.Project, error) {
	projects, _, err := Sdk.Client.Projects.List()
	if err != nil {
		return nil, err
	}
	for _, p := range projects {
		if p.Name == name {
			return &p, nil
		}
	}
	return nil, nil
}
