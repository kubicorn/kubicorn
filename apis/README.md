# Cluster API

This represents a Kubernetes cluster.

### Glossary

✖ This is not a `spec`.

✖ This is not a `model`.

✔ This is an `API`.

### What it is

This is a representation of a Kubernetes cluster.
This is cloud agnostic, and is defined by `struct{}`'s in Go.

# Cluster API

We will be adopting the Kubernetes cluster API

For now we are vendoring it in from `kube-deploy` using the suggested `client-go` and `apimachinery` packages.

Ultimately we will vendor this in from the new repository as we begin to lean on it more.

