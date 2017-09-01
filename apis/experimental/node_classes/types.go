package node_classes

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/api"
)

// This should match kubeadm's API as closely as possible, so its definition is
// uninteresting and not fleshed out.
//
// What is more important (for this PR) is the representation of nodes and how
// they are linked to the concept of the Cluster.
type Cluster struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	// ...
}

// This is the level that cloud-agnostic controllers would interact with. This
// is where one could set the desired size/shape of a NodePool.
type NodePool struct {
	metav1.ObjectMeta

	Spec NodePoolSpec
	Status NodePoolStatus
}

type NodePoolSpec struct {
	Name string // foo
	Class string // name of NodeClass
	Args string // JSON { ... } or ugly substruct with cloud-specific fields.
	            // this is explained in NodeClass.
	Size uint64
	MinSize uint64
	MaxSize uint64
}

type NodePoolStatus struct {
	Phase NodePoolPhase
	Message string
	Reason string

	// In addition to phase, we probably want a representation of in-flight
	// cloud requests so we don't duplicate them in further runs of the
	// controller loop. Once we decide to spin up a new node, we should
	// make note of that because it may still take a few minutes for it to
	// become healthy and register, and we don't want to keep triggering
	// instance creation just because current < desired. This could be
	// avoided if we always rely on an ASG/MIG, but that might not give us
	// the control we want.
	PendingNodeAdditions []string
	PendingNodeDeletions []string
}

type NodePoolPhase string
const (
	NodePoolPending NodePoolPhase = "Pending"
	NodePoolRunning NodePoolPhase = "Running"
	NodePoolTerminated NodePoolPhase = "Terminated"
	// ...
)

// These are objects that a cluster admin (or infrastructure admin, if a
// differnet role) would create in order to set the policy of what kinds of
// nodes are available and allowed for the cluster, analogous to
// PersistentVolumes and PersistentVolumeClaims. They can have generic names
// like "small," "medium," "gpu," or as arbitrarily complex as is useful
// ("large-gpu-us-west-ssd"). In a cloud context, it allows infrastructure
// admins to place restrictions on what resources are allowed to be created,
// and in an on-premise context, allows accurate modeling of what hardware is
// actually available for use.
//
// These serve as the glue between NodePools and NodeBuilders. In order to keep
// the NodeClass registry from becoming a combinatorial explosion, we can
// parameterize attributes that aren't core to the NodeClass' identity. For
// instance, it might be very important that region/zone is a core part of a
// NodeClass, and an admin wants their naming scheme to always start with the
// zone. However, within that zone, the exact disk image you use is far more
// flexible and inconsequential to the cluster admin, so it's modeled as an
// overridable attribute. NodePools can then specify Args to fill in the blanks
// or override defaults, and NodeBuilders will merge the Args into the Params
// and use the result to populate a node's attributes.
type NodeClass struct {
	Name string
	Params NodeClassParams
	NodeBuilder api.ObjectReference
}

type NodeClassParams struct {
	InstanceType string // n1-standard-1, r2.medium, etc.
	Disks []Disk
}

// Open question: is there value in modeling key package versions for a node?
// Can be represented in an OS-agnostic way a la cloud-init, but would let us
// have much more control over version of Docker, kubelet, etc., and perform
// rolling updates (potentially in-place).
type Disk struct {
	Bootable bool
	Size uint64
	Image string
}

// This would be attached to control plane components and NodeSpec so that you
// can trigger upgrades.
type KubernetesVersion struct {
	SemanticVersion string
	Path string // gs://..., s3://...
}

// NodeBuilders are the concrete cloud-provider definitions. They should have
// enough configuration to be able to make the service calls necessary to
// create/destroy/manipulate VMs, etc. The controller that ultimately makes a
// call to AWS to ask for a new node would be watching the AWSNodeBuilder
// objects and all NodePool objects that are linked to them in order to be able
// to reconcile state.
//
// These would likely be created by the administrator (or cluster provisioning
// tool) as a way of linking cloud service accounts and settings with the
// cluster.
//
// There's an open question about how these should link to the Cluster object,
// especially if we want to support representing N clusters as concurrent
// objects (for usecases like Cluster Registry, Federation, etc.). One way is
// to explicitly link to Cluster object via ObjectReference. Another way is to
// always give the controller a label selector of which objects it should care
// about reconciling, so it can be sliced and diced however the admin wants.

type AWSNodeBuilder struct {
	Name string
	ServiceAccount string
}

type GCENodeBuilder struct {
	Name string
	Project string
	Zone string
}

type GKENodeBuilder struct {
	Name string
	Project string
	Zone string
	ClusterID string
}

// AKA NoopNodeBuilder, although this could also know how to trigger a hardware
// provisioning process, like we hint at with PersistentVolumeClaims.
type OnPremiseNodeBuilder struct {
	Name string
}
