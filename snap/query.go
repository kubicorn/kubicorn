package snap

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/kris-nova/kubicorn/cutil/logger"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesQuery struct {
	kubeconfigPath string
	namespaces     []string
	config         *restclient.Config
	clientset      *kubernetes.Clientset
	result         []byte
}

func NewKubernetesQuery(kubeConfigPath string, namespaces []string) *KubernetesQuery {
	return &KubernetesQuery{
		kubeconfigPath: kubeConfigPath,
		namespaces:     namespaces,
	}
}

func (q *KubernetesQuery) Authenticate() error {
	config, err := clientcmd.BuildConfigFromFlags("", q.kubeconfigPath)
	if err != nil {
		return err
	}
	q.config = config
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	q.clientset = clientset
	return nil
}

type header struct {
	ApiVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`
}

func (q *KubernetesQuery) Execute() error {

	// Support for wildcard * namespaces
	if len(q.namespaces) == 1 && q.namespaces[0] == "*" {
		nss, err := q.GetAllNamespaces()
		if err != nil {
			return fmt.Errorf("Unable to look up namespaces: %v", err)
		}
		q.namespaces = nss
	}

	for _, namespace := range q.namespaces {
		logger.Info("-----------------------------------------------------------------------------")
		logger.Info("Querying namespace [%s]", namespace)
		logger.Info("-----------------------------------------------------------------------------")

		// ----- certificatesigningrequests -----
		{
			logger.Info(" ⎈ Kubernetes Certificate Signing Requests:")
			resources, err := q.clientset.CertificatesV1beta1().CertificateSigningRequests().List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "certificates/v1beta1",
					Kind:       "CertificateSigningRequest",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- clusterrolebindings -----
		{
			logger.Info(" ⎈ Kubernetes Cluster Role Bindings:")
			resources, err := q.clientset.RbacV1beta1().ClusterRoleBindings().List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "rbac/v1beta1",
					Kind:       "ClusterRoleBindings",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- componentstatuses -----
		{
			logger.Info(" ⎈ Kubernetes Component Statuses:")
			resources, err := q.clientset.CoreV1().ComponentStatuses().List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "v1",
					Kind:       "ComponentStatus",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- configmaps -----
		{
			logger.Info(" ⎈ Kubernetes Config Maps:")
			resources, err := q.clientset.CoreV1().ConfigMaps(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "v1",
					Kind:       "ConfigMap",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- controllerrevisions -----
		{
			logger.Info(" ⎈ Kubernetes Controller Revisions:")
			resources, err := q.clientset.AppsV1beta1().ControllerRevisions(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "apps/v1beta1",
					Kind:       "ControllerRevision",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- cronjobs -----
		{
			logger.Info(" ⎈ Kubernetes Batch Jobs:")
			resources, err := q.clientset.BatchV1().Jobs(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "batch/v1",
					Kind:       "Job",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- daemonsets -----
		{
			logger.Info(" ⎈ Kubernetes Daemon Sets:")
			resources, err := q.clientset.ExtensionsV1beta1().DaemonSets(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "extensions/v1beta1",
					Kind:       "DaemonSet",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- deployments -----
		{
			logger.Info(" ⎈ Kubernetes Deployments:")
			resources, err := q.clientset.AppsV1beta1().Deployments(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "apps/v1beta1",
					Kind:       "Deployment",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- endpoints -----
		{
			logger.Info(" ⎈ Kubernetes Endpoints:")
			resources, err := q.clientset.CoreV1().Endpoints(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "v1",
					Kind:       "Endpoint",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- events -----
		{
			logger.Info(" ⎈ Kubernetes Events:")
			resources, err := q.clientset.CoreV1().Events(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "v1",
					Kind:       "Event",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- horizontalpodautoscalers -----
		{
			logger.Info(" ⎈ Kubernetes Horizontal Pod Autoscaler:")
			resources, err := q.clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "autoscaling/v1",
					Kind:       "HorizontalPodAutoscaler",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- ingresses -----
		{
			logger.Info(" ⎈ Kubernetes Ingresses:")
			resources, err := q.clientset.ExtensionsV1beta1().Ingresses(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "extensions/v2beta1",
					Kind:       "Ingress",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- limitranges -----
		{
			logger.Info(" ⎈ Kubernetes Ingresses:")
			resources, err := q.clientset.CoreV1().LimitRanges(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "v1",
					Kind:       "LimitRange",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- networkpolicies -----
		{
			logger.Info(" ⎈ Kubernetes Network Policies:")
			resources, err := q.clientset.NetworkingV1().NetworkPolicies(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "networking/v1",
					Kind:       "LimitRange",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- persistentvolumeclaims -----
		{
			logger.Info(" ⎈ Kubernetes Persistent Volume Claims:")
			resources, err := q.clientset.CoreV1().PersistentVolumeClaims(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "v1",
					Kind:       "PersistentVolumeClaim",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- poddisruptionbudgets -----
		{
			logger.Info(" ⎈ Kubernetes Pod Disruption Budgets:")
			resources, err := q.clientset.PolicyV1beta1().PodDisruptionBudgets(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "policy/v1beta1",
					Kind:       "PodDisruptionBudget",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- podpreset -----
		{
			logger.Info(" ⎈ Kubernetes Pod Preset:")
			resources, err := q.clientset.SettingsV1alpha1().PodPresets(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "settings/v1alpha1",
					Kind:       "PodPreset",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- pods -----
		{
			logger.Info(" ⎈ Kubernetes Pods:")
			resources, err := q.clientset.CoreV1().Pods(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "v1",
					Kind:       "Pod",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- podsecuritypolicies -----
		{
			logger.Info(" ⎈ Kubernetes Pod Security Policies:")
			resources, err := q.clientset.ExtensionsV1beta1().PodSecurityPolicies().List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "extensions/v1beta1",
					Kind:       "PodSecurityPolicy",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- podtemplates -----
		{
			logger.Info(" ⎈ Kubernetes Pod Templates:")
			resources, err := q.clientset.CoreV1().PodTemplates(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "v1",
					Kind:       "PodSecurityPolicy",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- replicasets -----
		{
			logger.Info(" ⎈ Kubernetes Replica Set:")
			resources, err := q.clientset.ExtensionsV1beta1().ReplicaSets(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "extensions/v1beta1",
					Kind:       "ReplicaSet",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- replicationcontrollers -----
		{
			logger.Info(" ⎈ Kubernetes Replication Controllers:")
			resources, err := q.clientset.CoreV1().ReplicationControllers(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "v1",
					Kind:       "ReplicationController",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- resourcequotas -----
		{
			logger.Info(" ⎈ Kubernetes Resource Quotas:")
			resources, err := q.clientset.CoreV1().ResourceQuotas(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "v1",
					Kind:       "ResourceQuota",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- rolebindings -----
		{
			logger.Info(" ⎈ Kubernetes Role Bindings:")
			resources, err := q.clientset.RbacV1beta1().RoleBindings(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "rbac/v1beta1",
					Kind:       "RoleBinding",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- roles -----
		{
			logger.Info(" ⎈ Kubernetes Roles:")
			resources, err := q.clientset.RbacV1beta1().Roles(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "rbac/v1beta1",
					Kind:       "Role",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- secrets -----
		{
			logger.Info(" ⎈ Kubernetes Secrets:")
			resources, err := q.clientset.CoreV1().Secrets(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "v1",
					Kind:       "Secret",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- serviceaccounts -----
		{
			logger.Info(" ⎈ Kubernetes Service Accounts:")
			resources, err := q.clientset.CoreV1().ServiceAccounts(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "v1",
					Kind:       "ServiceAccount",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- services -----
		{
			logger.Info(" ⎈ Kubernetes Services:")
			resources, err := q.clientset.CoreV1().Services(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "v1",
					Kind:       "Service",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- statefulsets -----
		{
			logger.Info(" ⎈ Kubernetes Stateful Set:")
			resources, err := q.clientset.AppsV1beta1().StatefulSets(namespace).List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "apps/v1beta1",
					Kind:       "StatefulSet",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- storageclasses -----
		{
			logger.Info(" ⎈ Kubernetes Storage Classes:")
			resources, err := q.clientset.StorageV1().StorageClasses().List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "storage/v1",
					Kind:       "StorageClass",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}

		// ----- thirdpartyresources -----
		{
			logger.Info(" ⎈ Kubernetes Third Party Resources:")
			resources, err := q.clientset.ExtensionsV1beta1().ThirdPartyResources().List(meta_v1.ListOptions{})
			if err != nil {
				return fmt.Errorf("Invalid resource list: %v", err)
			}
			for _, item := range resources.Items {
				headerData, err := yaml.Marshal(header{
					ApiVersion: "extensions/v1beta1",
					Kind:       "ThirdPartyResource",
				})
				if err != nil {
					return err
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("Unable to marshal %v", err)
				}

				data = append(headerData, data...)
				logger.Info("    →  %s", item.Name)
				q.result = appendData(q.result, data)
			}
		}
	}
	return nil
}

func appendData(a, b []byte) []byte {
	sepStr := `
---

`
	sep := []byte(sepStr)
	a = append(a, sep...)
	a = append(a, b...)
	return a
}

func (q *KubernetesQuery) GetAllNamespaces() ([]string, error) {
	nsl, err := q.clientset.CoreV1().Namespaces().List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var namespaces []string
	for _, namespace := range nsl.Items {
		namespaces = append(namespaces, namespace.Name)
	}
	return namespaces, nil
}

func (q *KubernetesQuery) Bytes() []byte {
	return q.result
}
