# apiserver-framework

The apiserver-framework implements libraries and tools to quickly and easily build Kubernetes apiservers
to support custom resource types.  The intention is provide 100% feature parity with
apiservers built from scratch, but without the boilerplate and copy-pasting.

## Getting started

Instructions for building your first apiserver using the `apiserver-boot` tool can be found
[here](https://github.com/kubernetes-incubator/apiserver-builder/blob/master/docs/getting_started.md).

## Motivation

Standing up apiservers from scratch and adding apis requires 100's of lines of boilerplate
code that must be understood and maintained (rebased against master).  There are few defaults,
requiring the common case configuration to be repeated for each new apiserver and resource.
Apiservers rely heavily on code generation to build libraries used by the apiserver, putting a
steep learning curve on Kubernetes community members that want to implement a native api.
Frameworks like Ruby on Rails and Spring have made standing up REST apis trivial by eliminating
boilerplate and defaulting common values, allowing developers to focus on creating
implementing the business logic of their component.

## Goals

- Working hello-world apiserver in ~5 lines
- Declaring new resource types only requires defining the struct definition
  and taging it as a resource
- Adding sub-resources only requires defining the request-type struct definition,
  implementing the REST implementation, and tagging the parent resource.
- Adding validation / defaulting to a type only requires defining the validation / defaulting method
  as a function of the appropriate struct type.
- All necessary generated code can be generated running a single command, passing in repo root.

## Proposal

Construct a set of libraries and tools to reduce and generate the boilerplate.

### Binary distribution of build tools

- Distribute binaries for all of the code-generators
- Write porcelian wrapper for code-generators that is able to detect
  the appropriate arguments for each from the go PATH and types.go files

### Helper libraries

- Implement common-case defaults for create/update strategies
  - Define implementable interfaces for default actions requiring
    type specific knowledge - e.g. HasStatus - how to set and get Status
- Implement libraries for registering types and setting up strategies
  - Implement structs to defining wiring semantics instead of linking
    directly to package variables for declarations
- Implement libraries for registering subresources

### Generate code for common defaults that require type or variable declarations

- Implementations for "unversioned" types
- Implementations for "List" types
- Package variables used by code generation
- Generate invocations of helper libraries from observered types.go types

### Support hooks for overriding defaults

- Try to support 100% of the flexibility of manually writing the boilerplate by 
  providing hooks.
  - Implement functions that can be invoked to register overrides
  - Use type embeding to inherit defaults but allow new functions to override the defaults
  
### Support for generating reference documentation

- Generate k8s.io style reference documentation for declared types
  - Support for request / response examples and manual edits

### Thorough documentation and examples for how to use the framework

- Hello-world example
- How to override each default
- Build tools
- How to use libraries directly (without relying on code generation)
