# Reconciler

A `Reconciler` defines a core worker in `kubicorn` and there is one per cloud.
The `Reconciler` for each cloud implements the same interface and is used to interact and audit a cloud.


```go
type Reconciler interface {
	Init() error
	GetActual() (*cluster.Cluster, error)
	GetExpected() (*cluster.Cluster, error)
	Reconcile(actualCluster, expectedCluster *cluster.Cluster) (*cluster.Cluster, error)
	Destroy() error
}
```

The `Reconciler` can be used for the following functionality:
### Init() error

This is used to initialize a `Reconciler` and performs tasks like registering an Sdk in memory to connect to the cloud.

### GetActual() (*cluster.Cluster, error)

This is used to return a cluster API for the existing state of a cloud.

### GetExpected() (*cluster.Cluster, error)

This is used to to sanity check a cluster API. This will return the expected cluster API, after it has been mapped via the internal mapper.
This should never conflict with the original cluster API.

### Reconcile(actualCluster, expectedCluster *cluster.Cluster) (*cluster.Cluster, error)

This can be called to reconcile an actual and expected state of a cluster.
The internal mapper is called, and each of the underlying resources is asserted.
If a delta is detected, the reconciler will take action.
By design as the `Reconciler` is reconciling, if an error is detected it will unwind itself, and delete resources it already created.

### Destroy() error

This will map the cluster API to resources, and attempt to delete each of them.

# Resource

The `Reconciler` will map each cluster API to a set of `Resource`s.
The `Resource` interface defines how the `Reconciler` will interact with each `Resource`.

```go
type Resource interface {
	Actual(known *cluster.Cluster) (Resource, error)
	Expected(known *cluster.Cluster) (Resource, error)
	Apply(actual, expected Resource, expectedCluster *cluster.Cluster) (Resource, error)
	Delete(actual Resource, known *cluster.Cluster) error
	Render(renderResource Resource, renderCluster *cluster.Cluster) (*cluster.Cluster, error)
	Tag(tags map[string]string) error
}
```