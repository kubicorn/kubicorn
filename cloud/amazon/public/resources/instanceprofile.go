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
	"net/url"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud"
	"github.com/kubicorn/kubicorn/pkg/compare"
	"github.com/kubicorn/kubicorn/pkg/logger"
)

var _ cloud.Resource = &InstanceProfile{}

type InstanceProfile struct {
	Shared
	Role       *IAMRole
	ServerPool *cluster.ServerPool
}

type IAMRole struct {
	Shared
	Policies []*IAMPolicy
}

type IAMPolicy struct {
	Shared
	Document string
}

func (r *InstanceProfile) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("instanceprofile.Actual")
	newResource := &InstanceProfile{
		Shared: Shared{
			Name: r.Name,
			Tags: map[string]string{
				"Name":              r.Name,
				"KubernetesCluster": immutable.Name,
			},
		},
		ServerPool: r.ServerPool,
	}
	// Get InstanceProfile
	if r.Identifier != "" {
		logger.Debug("Query InstanceProfile: %v", newResource.Name)
		respInstanceProfile, err := Sdk.IAM.GetInstanceProfile(&iam.GetInstanceProfileInput{
			InstanceProfileName: &newResource.Name,
		})
		if err != nil {
			return nil, nil, err
		}
		newResource.Identifier = *respInstanceProfile.InstanceProfile.InstanceProfileName
		// Get Roles
		if len(respInstanceProfile.InstanceProfile.Roles) > 0 {
			//ListRolePolicies
			for _, role := range respInstanceProfile.InstanceProfile.Roles {
				policyList, err := Sdk.IAM.ListRolePolicies(&iam.ListRolePoliciesInput{
					RoleName: role.RoleName,
				})
				if err != nil {
					return nil, nil, err
				}
				//Here we add the role to InstanceProfile
				iamrole := &IAMRole{
					Shared: Shared{
						Tags: map[string]string{
							"Name":              r.Name,
							"KubernetesCluster": immutable.Name,
						},
						Name: *role.RoleName,
					},
				}
				newResource.Role = iamrole

				for _, policyName := range policyList.PolicyNames {
					policyOutput, err := Sdk.IAM.GetRolePolicy(&iam.GetRolePolicyInput{
						PolicyName: policyName,
						RoleName:   role.RoleName,
					})
					if err != nil {
						return nil, nil, err
					}
					//Here we add the policy to the role
					iampolicy := &IAMPolicy{
						Shared: Shared{
							Tags: map[string]string{
								"Name":              r.Name,
								"KubernetesCluster": immutable.Name,
							},
							Name: *policyOutput.PolicyName,
						},
					}
					raw, err := url.QueryUnescape(*policyOutput.PolicyDocument)
					if err != nil {
						return nil, nil, err
					}
					iampolicy.Document = raw
					iamrole.Policies = append(iamrole.Policies, iampolicy)
				}
			}
		}
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *InstanceProfile) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("instanceprofile.Expected %v", r.Name)
	newResource := &InstanceProfile{
		Shared: Shared{
			Tags: map[string]string{
				"Name":              r.Name,
				"KubernetesCluster": immutable.Name,
			},
			Name:       r.Name,
			Identifier: r.Identifier,
		},
		ServerPool: r.ServerPool,
		Role: &IAMRole{
			Shared: Shared{
				Name: r.Role.Name,
				Tags: map[string]string{
					"Name":              r.Name,
					"KubernetesCluster": immutable.Name,
				},
			},
			Policies: []*IAMPolicy{},
		},
	}
	for _, policy := range r.Role.Policies {
		newPolicy := &IAMPolicy{
			Shared: Shared{
				Name: policy.Name,
				Tags: map[string]string{
					"Name":              r.Name,
					"KubernetesCluster": immutable.Name,
				},
			},
			Document: policy.Document,
		}
		newResource.Role.Policies = append(newResource.Role.Policies, newPolicy)
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *InstanceProfile) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("instanceprofile.Apply")
	applyResource := expected.(*InstanceProfile)
	isEqual, err := compare.IsEqual(actual.(*InstanceProfile), expected.(*InstanceProfile))
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, applyResource, nil
	}
	logger.Debug("Actual: %#v", actual)
	logger.Debug("Expectd: %#v", expected)
	newResource := &InstanceProfile{}
	//TODO fill in instanceprofile attributes

	// Create InstanceProfile
	var nameStr, idStr string
	profileinput := &iam.CreateInstanceProfileInput{
		InstanceProfileName: &expected.(*InstanceProfile).Name,
		Path:                S("/"),
	}
	outInstanceProfile, err := Sdk.IAM.CreateInstanceProfile(profileinput)
	if err != nil {
		logger.Debug("CreateInstanceProfile error: %v", err)
		if err.(awserr.Error).Code() != iam.ErrCodeEntityAlreadyExistsException {
			return nil, nil, err
		} else {
			logger.Debug("InstanceProfile found, using existing.")
			profileinput := &iam.GetInstanceProfileInput{
				InstanceProfileName: &expected.(*InstanceProfile).Name,
			}
			outInstanceProfile, err := Sdk.IAM.GetInstanceProfile(profileinput)
			if err != nil {
				return nil, nil, err
			}
			nameStr = *outInstanceProfile.InstanceProfile.InstanceProfileName
			idStr = *outInstanceProfile.InstanceProfile.InstanceProfileId
		}
	} else {
		nameStr = *outInstanceProfile.InstanceProfile.InstanceProfileName
		idStr = *outInstanceProfile.InstanceProfile.InstanceProfileId
	}
	newResource.Name = nameStr
	newResource.Identifier = idStr
	logger.Info("InstanceProfile created: %s", newResource.Name)
	// Create role
	assumeRolePolicy := `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": "ec2.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]}`
	roleinput := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: &assumeRolePolicy,
		RoleName:                 &expected.(*InstanceProfile).Role.Name,
		Description:              S("Kubicorn Role"),
		Path:                     S("/"),
	}
	irName := ""
	outInstanceRole, err := Sdk.IAM.CreateRole(roleinput)
	if err != nil {
		logger.Debug("CreateRole error: %v", err)
		if err.(awserr.Error).Code() != iam.ErrCodeEntityAlreadyExistsException {
			irName = expected.(*InstanceProfile).Role.Name
		}
	} else {
		irName = *outInstanceRole.Role.RoleName
	}
	newIamRole := &IAMRole{
		Shared: Shared{
			Name: irName,
			Tags: map[string]string{
				"Name":              r.Name,
				"KubernetesCluster": immutable.Name,
			},
		},
		Policies: []*IAMPolicy{},
	}
	logger.Info("Role created")
	//Attach Policy to Role
	for _, policy := range expected.(*InstanceProfile).Role.Policies {
		policyinput := &iam.PutRolePolicyInput{
			PolicyDocument: &policy.Document,
			PolicyName:     &policy.Name,
			RoleName:       &expected.(*InstanceProfile).Role.Name,
		}
		_, err := Sdk.IAM.PutRolePolicy(policyinput)
		if err != nil {
			logger.Debug("PutRolePolicy error: %v", err)
			if err.(awserr.Error).Code() != iam.ErrCodeLimitExceededException {
				return nil, nil, err
			}
		}
		newPolicy := &IAMPolicy{
			Shared: Shared{
				Name: policy.Name,
				Tags: map[string]string{
					"Name":              r.Name,
					"KubernetesCluster": immutable.Name,
				},
			},
			Document: policy.Document,
		}
		newIamRole.Policies = append(newIamRole.Policies, newPolicy)
		logger.Info("Policy created")
	}
	//Attach Role to Profile
	roletoprofile := &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: &expected.(*InstanceProfile).Name,
		RoleName:            &expected.(*InstanceProfile).Role.Name,
	}
	_, err = Sdk.IAM.AddRoleToInstanceProfile(roletoprofile)
	if err != nil {
		logger.Debug("AddRoleToInstanceProfile error: %v", err)
		if err.(awserr.Error).Code() != iam.ErrCodeLimitExceededException {
			return nil, nil, err
		}
	}
	newResource.Role = newIamRole
	logger.Info("RoleAttachedToInstanceProfile done")
	//Add ServerPool
	newResource.ServerPool = expected.(*InstanceProfile).ServerPool
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *InstanceProfile) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	for _, policy := range r.Role.Policies {
		_, err := Sdk.IAM.DeleteRolePolicy(&iam.DeleteRolePolicyInput{
			PolicyName: &policy.Name,
			RoleName:   &r.Role.Name,
		})
		if err != nil {
			logger.Debug("Problem deleting Policy %s for Role: %s: %v", policy.Name, r.Role.Name, err)
			if err.(awserr.Error).Code() != iam.ErrCodeNoSuchEntityException {
				return nil, nil, err
			}
		}
	}
	_, err := Sdk.IAM.RemoveRoleFromInstanceProfile(&iam.RemoveRoleFromInstanceProfileInput{
		InstanceProfileName: &r.Name,
		RoleName:            &r.Role.Name,
	})
	if err != nil {
		logger.Debug("Problem remove Role %s from InstanceProfile %s: %v", r.Role.Name, r.Name, err)
		if err.(awserr.Error).Code() != iam.ErrCodeNoSuchEntityException {
			return nil, nil, err
		}
	}
	_, err = Sdk.IAM.DeleteRole(&iam.DeleteRoleInput{
		RoleName: &r.Role.Name,
	})
	if err != nil {
		logger.Debug("Problem delete role %s: %v", r.Role.Name, err)
		if err.(awserr.Error).Code() != iam.ErrCodeNoSuchEntityException {
			return nil, nil, err
		}
	}
	_, err = Sdk.IAM.DeleteInstanceProfile(&iam.DeleteInstanceProfileInput{
		InstanceProfileName: &r.Name,
	})
	if err != nil {
		logger.Debug("Problem delete InstanceProfile %s: %v", r.Name, err)
		if err.(awserr.Error).Code() != iam.ErrCodeNoSuchEntityException {
			return nil, nil, err
		}
	}
	newResource := &InstanceProfile{}
	newCluster := r.immutableRender(newResource, immutable)
	logger.Info("Deleted InstanceProfile: %v", r.Name)
	return newCluster, newResource, nil
}

func (r *InstanceProfile) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("instanceprofile.Render")
	newCluster := inaccurateCluster
	instanceProfile := &cluster.IAMInstanceProfile{}
	instanceProfile.Name = newResource.(*InstanceProfile).Name
	instanceProfile.Identifier = newResource.(*InstanceProfile).Identifier
	instanceProfile.Role = &cluster.IAMRole{}
	if newResource.(*InstanceProfile).Role != nil {
		instanceProfile.Role.Name = newResource.(*InstanceProfile).Role.Name
		if len(newResource.(*InstanceProfile).Role.Policies) > 0 {
			for i := range newResource.(*InstanceProfile).Role.Policies {
				newPolicy := &cluster.IAMPolicy{
					Name:       newResource.(*InstanceProfile).Role.Policies[i].Name,
					Identifier: newResource.(*InstanceProfile).Role.Policies[i].Identifier,
					Document:   newResource.(*InstanceProfile).Role.Policies[i].Document,
				}
				instanceProfile.Role.Policies = append(instanceProfile.Role.Policies, newPolicy)
			}
		}
	}
	found := false
	machineProviderConfigs := newCluster.MachineProviderConfigs()
	for i := 0; i < len(machineProviderConfigs); i++ {
		machineProviderConfig := machineProviderConfigs[i]
		if newResource.(*InstanceProfile).ServerPool != nil && machineProviderConfig.Name == newResource.(*InstanceProfile).ServerPool.Name {
			machineProviderConfigs[i].ServerPool.InstanceProfile = instanceProfile
			newCluster.SetMachineProviderConfigs(machineProviderConfigs)
			found = true
		}
	}
	if !found {
		logger.Debug("Not found ServerPool for InstanceProfile!")
	}
	return newCluster
}
