Minimalistic Machines API
=========================

This proposal is for a minimalistic start to a new Machines API, as part of the
overall Cluster API project. It is intended to live outside of core Kubernetes
and add optional machine management features to Kubernetes clusters.

## Capabilities

This API strives to be able to add these capabilities:

1. A new Node can be created in a declarative way, including Kubernetes version
   and container runtime version. It should also be able to specify
   provider-specific information such as OS image, instance type, disk
   configuration, etc., though this will not be portable.

2. A specific Node can be deleted, freeing external resources associated with
   it.

3. A specific Node can have its kubelet version upgraded or downgraded in a
   declarative way\*.

4. A specific Node can have its container runtime changed, or its version
   upgraded or downgraded, in a declarative way\*.

5. A specific Node can have its OS image upgraded or downgraded in a declarative
   way\*.

\*  It is an implementation detail of the provider if these operations are
performed in-place or via Node replacement.

## Proposal

This proposal introduces a new API type: Machine. See the full definition in
[types.go](types.go).

A "Machine" is the declarative spec for a Node, as represented in Kuberenetes
core. If a new Machine object is created, a provider-specific controller will
handle provisioning and installing a new host to register as a new Node matching
the Machine spec. If the Machine's spec is updated, a provider-specific
controller is responsible for updating the Node in-place or replacing the host
with a new one matching the updated spec. If a Machine object is deleted, the
corresponding Node should have its external resources released by the
provider-specific controller, and should be deleted as well.

Fields like the kubelet version, the container runtime to use, and its version,
are modeled as fields on the Machine's spec. Any other information that is
provider-specific, though, is part of an opaque ProviderConfig string that is
not portable between different providers.

The ProviderConfig is recommended to be a serialized API object in a format
owned by that provider, akin to the [Component Config](https://goo.gl/opSc2o)
pattern. This will allow the configuration to be strongly typed, versioned, and
have as much nested depth as appropriate. These provider-specific API
definitions are meant to live outside of the Machines API, which will allow them
to evolve independently of it. Attributes like instance type, which network to
use, and the OS image all belong in the ProviderConfig.

## In-place vs. Replace

One simplification that might be controversial in this proposal is the lack of
API control over "in-place" versus "replace" reconciliation strategies. For
instance, if a Machine's spec is updated with a different version of kubelet
than is actually running, it is up to the provider-specific controller whether
the request would best be fulfilled by performing an in-place upgrade on the
Node, or by deleting the Node and creating a new one in its place (or reporting
an error if this particular update is not supported). One can force a Node
replacement by deleting and recreating the Machine object rather than updating
it, but no similar mechanism exists to force an in-place change.

Another approach considered was that modifying an existing Machine should only
ever attempt an in-place modification to the Node, and Node replacement should
only occur by deleting and creating a new Machine. In that case, a provider
would set an error field in the status if it wasn't able to fulfill the
requested in-place change (such as changing the OS image or instance type in a
cloud provider).

The reason this approach wasn't used was because most cluster upgrade tools
built on top of the Machines API would follow the same pattern:

    for machine in machines:
        attempt to upgrade machine in-place
        if error:
            create new machine
            delete old machine

Since updating a Node in-place is likely going to be faster than completely
replacing it, most tools would opt to use this pattern to attempt an in-place
modification first, before falling back to a full replacement.

It seems like a much more powerful concept to allow every tool to instead say:

    for machine in machines:
        update machine

and allow the provider to decide if it is capable of performing an in-place
update, or if a full Node replacement is necessary.

## Omitted Capabilities

* A scalable representation of a group of nodes

Given the existing targeted capabilities, this functionality could easily be
built client-side via label selectors to find groups of Nodes and using (1) and
(2) to add or delete instances to simulate this scaling.

It is natural to extend this API in the future to introduce the concepts of
MachineSets and MachineDeployments that mirror ReplicaSets and Deployments, but
an initial goal is to first solidify the definition and behavior of a single
Machine, similar to how Kubernetes first solidifed Pods.

A nice property of this proposal is that if provider controllers are written
solely against Machines, the concept of MachineSets can be implemented in a
provider-agnostic way with a generic controller that uses the MachineSet
template to create and delete Machine instances. All Machine-based provider
controllers will continue to work, and will get full MachineSet functionality
for free without modification. Similarly, a MachineDeployment controller could
then be introduced to generically operate on MachineSets without having to know
about Machines or providers. Provider-specific controllers that are actually
responsible for creating and deleting hosts would only ever have to worry about
individual Machine objects, unless they explicitly opt into watching
higher-level APIs like MachineSets in order to take advantage of
provider-specific features like AutoScalingGroups or Managed Instance Groups.

However, this leaves the barrier to entry very low for adding new providers:
simply implement creation and deletion of individual Nodes, and get Sets and
Deployments for free.

* A provider-agnostic mechanism to request new nodes

In this proposal, only certain attributes of Machines are provider-agnostic and
can be operated on in a generic way. In other iterations of similar proposals,
much care had been taken to allow the creation of truly provider-agnostic
Machines that could be mapped to provider-specific attributes in order to better
support usecases around automated Machine scaling. This introduced a lot of
upfront complexity in the API proposals.

This proposal starts much more minimalistic, but doesn't preclude the option of
extending the API to support these advanced concepts in the future.

* Dynamic API endpoint

This proposal lacks the ability to declaratively update the kube-apiserver
endpoint for the kubelet to register with. This feature could be added later,
but doesn't seem to have demand now. Rather than modeling the kube-apiserver
endpoint in the Machine object, it is expected that the cluster installation
tool resolves the correct endpoint to use, starts a provider-specific Machines
controller configured with this endpoint, and that the controller injects the
endpoint into any hosts it provisions.

## Conditions

Brian Grant and Eric Tune have indicated that the API pattern of having
"Conditions" lists in object statuses is soon to be deprecated. These have
generally been used as a timeline of state transitions for the object's
reconcilation, and difficult to consume for clients that just want a meaningful
representation of the object's current state. There are no existing examples of
the new pattern to follow instead, just the guidance that we should use
top-level fields in the status to reprensent meaningful information. We can
revisit the specifics when new patterns start to emerge in core.

## Types

Please see the full types [here](types.go).
