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

package cmd

import (
	"fmt"
	"os"
	"strings"

	"encoding/json"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/pkg/cli"
	"github.com/kubicorn/kubicorn/pkg/kubeconfig"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/kubicorn/kubicorn/pkg/namer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yuroyoro/swalker"
	"k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
)

// CreateCmd represents create command
func CreateCmd() *cobra.Command {
	var co = &cli.CreateOptions{}
	var createCmd = &cobra.Command{
		Use:   "create [NAME] [-p|--profile PROFILENAME] [-c|--cloudid CLOUDID]",
		Short: "Create a Kubicorn API model from a profile",
		Long: `Use this command to create a Kubicorn API model in a defined state store.

	This command will create a cluster API model as a YAML manifest in a state store.
	Once the API model has been created, a user can optionally change the model to their liking.
	After a model is defined and configured properly, the user can then apply the model.`,
		Run: func(cmd *cobra.Command, args []string) {
			switch len(args) {
			case 0:
				co.Name = viper.GetString(keyKubicornName)
				if co.Name == "" {
					co.Name = namer.RandomName()
				}
			case 1:
				co.Name = args[0]
			default:
				logger.Critical("Too many arguments.")
				os.Exit(1)
			}

			if err := RunCreate(co); err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}

		},
	}

	fs := createCmd.Flags()

	bindCommonStateStoreFlags(&co.StateStoreOptions, fs)
	bindCommonAwsFlags(&co.AwsOptions, fs)

	fs.StringVarP(&co.Profile, keyProfile, "p", viper.GetString(keyProfile), descProfile)
	fs.StringVarP(&co.CloudID, keyCloudID, "c", viper.GetString(keyCloudID), descCloudID)
	fs.StringVar(&co.KubeConfigLocalFile, keyKubeConfigLocalFile, viper.GetString(keyKubeConfigLocalFile), descKubeConfigLocalFile)
	fs.StringArrayVarP(&co.Set, keySet, "C", viper.GetStringSlice(keySet), descSet)
	fs.StringArrayVarP(&co.MasterSet, keyMasterSet, "M", viper.GetStringSlice(keyMasterSet), descMasterSet)
	fs.StringArrayVarP(&co.NodeSet, keyNodeSet, "N", viper.GetStringSlice(keyNodeSet), descNodeSet)
	fs.StringVarP(&co.GitRemote, keyGitConfig, "g", viper.GetString(keyGitConfig), descGitConfig)
	fs.StringArrayVar(&co.AwsOptions.PolicyAttachments, keyPolicyAttachments, co.AwsOptions.PolicyAttachments, descPolicyAttachments)

	flagApplyAnnotations(createCmd, "profile", "__kubicorn_parse_profiles")
	flagApplyAnnotations(createCmd, "cloudid", "__kubicorn_parse_cloudid")

	createCmd.SetUsageTemplate(cli.UsageTemplate)

	return createCmd
}

// RunCreate is the starting point when a user runs the create command.
func RunCreate(options *cli.CreateOptions) error {
	// Create our cluster resource
	name := options.Name
	var newCluster *cluster.Cluster
	if _, ok := cli.ProfileMapIndexed[options.Profile]; ok {
		newCluster = cli.ProfileMapIndexed[options.Profile].ProfileFunc(name)
	} else {
		return fmt.Errorf("Invalid profile [%s]", options.Profile)
	}

	if options.KubeConfigLocalFile != "" {
		if newCluster.Annotations == nil {
			newCluster.Annotations = make(map[string]string)
		}
		newCluster.Annotations[kubeconfig.ClusterAnnotationKubeconfigLocalFile] = options.KubeConfigLocalFile
	}

	if len(options.Set) > 0 {
		// Here we override Set options
		for _, set := range options.Set {
			parts := strings.SplitN(set, "=", 2)
			if len(parts) == 1 {
				continue
			}
			providerConfig := newCluster.ProviderConfig()
			err := swalker.Write(strings.Title(parts[0]), providerConfig, parts[1])
			if err != nil {
				//fmt.Println(1)
				return fmt.Errorf("Invalid --set: %v", err)
			}
			newCluster.SetProviderConfig(providerConfig)
		}
	}

	if len(options.MasterSet) > 0 {
		// Here we override MasterSet options
		for _, set := range options.MasterSet {
			parts := strings.SplitN(set, "=", 2)
			if len(parts) == 1 {
				continue
			}

			for i, ms := range newCluster.MachineSets {
				isMaster := false
				for _, role := range ms.Spec.Template.Spec.Roles {
					if role == v1alpha1.MasterRole {
						isMaster = true
						break
					}
				}
				if !isMaster {
					continue
				}
				pcStr := ms.Spec.Template.Spec.ProviderConfig
				providerConfig := &cluster.MachineProviderConfig{}
				json.Unmarshal([]byte(pcStr), providerConfig)
				err := swalker.Write(strings.Title(parts[0]), providerConfig, parts[1])
				if err != nil {
					//fmt.Println(2)
					return fmt.Errorf("Invalid --set: %v", err)
				}
				// Now set the provider config
				bytes, err := json.Marshal(providerConfig)
				if err != nil {
					logger.Critical("Unable to marshal provider config: %v", err)
					return err
				}
				str := string(bytes)
				newCluster.MachineSets[i].Spec.Template.Spec.ProviderConfig = str
			}

		}
	}

	if len(options.NodeSet) > 0 {
		// Here we override NodeSet options
		for _, set := range options.NodeSet {
			parts := strings.SplitN(set, "=", 2)
			if len(parts) == 1 {
				continue
			}
			for i, ms := range newCluster.MachineSets {
				isNode := false
				for _, role := range ms.Spec.Template.Spec.Roles {
					if role == v1alpha1.NodeRole {
						isNode = true
						break
					}
				}
				if !isNode {
					continue
				}
				pcStr := ms.Spec.Template.Spec.ProviderConfig
				providerConfig := &cluster.MachineProviderConfig{}
				json.Unmarshal([]byte(pcStr), providerConfig)
				err := swalker.Write(strings.Title(parts[0]), providerConfig, parts[1])
				if err != nil {
					//fmt.Println(3)
					return fmt.Errorf("Invalid --set: %v", err)
				}
				// Now set the provider config
				bytes, err := json.Marshal(providerConfig)
				if err != nil {
					logger.Critical("Unable to marshal provider config: %v", err)
					return err
				}
				str := string(bytes)
				newCluster.MachineSets[i].Spec.Template.Spec.ProviderConfig = str
			}

		}
	}

	if len(options.AwsOptions.PolicyAttachments) > 0 {
		for i, ms := range newCluster.MachineSets {
			pcStr := ms.Spec.Template.Spec.ProviderConfig
			providerConfig := &cluster.MachineProviderConfig{}
			if err := json.Unmarshal([]byte(pcStr), providerConfig); err != nil {
				logger.Critical("Unable to unmarshal provider config: %v", err)
				return err
			}
			if providerConfig.ServerPool != nil && providerConfig.ServerPool.InstanceProfile != nil && providerConfig.ServerPool.InstanceProfile.Role != nil {
				providerConfig.ServerPool.InstanceProfile.Role.PolicyAttachments = options.AwsOptions.PolicyAttachments
			}
			// Now set the provider config
			bytes, err := json.Marshal(providerConfig)
			if err != nil {
				logger.Critical("Unable to marshal provider config: %v", err)
				return err
			}
			str := string(bytes)
			newCluster.MachineSets[i].Spec.Template.Spec.ProviderConfig = str
		}
	}

	if newCluster.ProviderConfig().Cloud == cluster.CloudGoogle && options.CloudID == "" {
		return fmt.Errorf("CloudID is required for google cloud. Please set it to your project ID")
	}

	providerConfig := newCluster.ProviderConfig()
	providerConfig.CloudId = options.CloudID
	newCluster.SetProviderConfig(providerConfig)

	// Expand state store path
	// Todo (@kris-nova) please pull this into a filepath package or something
	options.StateStorePath = cli.ExpandPath(options.StateStorePath)

	// Register state store and check if it exists
	stateStore, err := options.NewStateStore()
	if err != nil {
		return err
	} else if stateStore.Exists() {
		return fmt.Errorf("State store [%s] exists, will not overwrite. Delete existing profile [%s] and retry", name, options.StateStorePath+"/"+name)
	}

	// Init new state store with the cluster resource
	err = stateStore.Commit(newCluster)
	if err != nil {
		return fmt.Errorf("Unable to init state store: %v", err)
	}

	logger.Always("The state [%s/%s/cluster.yaml] has been created. You can edit the file, then run `kubicorn apply %s`", options.StateStorePath, name, name)
	return nil
}
