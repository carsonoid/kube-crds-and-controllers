# Kubernetes Controllers & CustomResourceDefinitions

## The Goal

Watch all pods in the cluster and make sure they all have a configurable set of labels.

## Dependencies

All the example controllers here use `glide` for dependency management. They are built based on client-go version `v5.0.1` which is most compatible with Kubernetes v1.8.

If you want to use a newer version of client-go, you can find examples that use the latest client-go version, `v6.0.0`, with `glide`, `godep`, and `dep` in the `deptoools` folder. 

All the controllers here should be able to work with client-go `v6.0.0`, but some code change would be required as client-go has made breaking changes between `v5.0.x` and `v6.0.0.`

## Getting Ready

Before running any of the controllers below, make sure you everything is ready by running:

```bash
make
```

It may also be helpful to view the changes between controllers as a series of git commits. To do so you can run:

```bash
make diff-repo
```

And the repo will be generated, You can also find it at `carsonoid/kube-crds-and-controllers-diffs`

## The Controllers

To illustrate basic functionality and common pitfalls the examples are broken up into three different groups: `hard-coded`, `configmap-configued`, and `crd-configured`. All the controllers do roughly the same thing but in different ways with different caveats.

### controllers/hard-coded

A custom controller which has everything hard-coded

##### A simple hard-coded controller

A very simple controller which handles pod labels one at a time based on a hard-coded label map.

```bash
make run-controllers/hard-coded/simple
```

##### A structured hard-coded controller

The simple controller, but with configuration moved to a more defined structure. Suitable for loading from a file.

Hard-coded defaults via the struct:

```bash
make run-controllers/hard-coded/structured
```

Or via a config file:

```bash
make run-controllers/hard-coded/structured OPTS="-config controllers/hard-coded/config.yaml"
```

### controllers/configmap-configured

Use a ConfigMap to configure the controller. This is essentially the same as passing a configmap
based file into the pod. But with a reduced overhead and faster response to config changes.

##### Support for a single configuration in a configmap

Reads a configmap which supports a single configuration

```bash
make run-controllers/configmap-configured/single-config
```

##### Support for multiple configurations in a configmap

Reads a configmap which supports multiple configurations

```bash
make run-controllers/configmap-configured/multi-config
```

### controllers/crd-configured

Use a CustomResourceDefinition to provide configurations to the controller. Using CRDs not only provides a very dynamic and Kubernetes native
way of object handling, but it also provide instant usability by any existin Kubernetes tooling.

##### A blocking controller using CRDs

A controller which uses a CRDs to for all configuratins. Done with simple client-go mechanisms that are easy to understand
but block during processing and can re-process the same resource multiple times

```bash
make run-controllers/crd-configured/simple
```


##### A controller that uses a workqueue to handle pods

A controller which uses CRDs for all configurations. But uses a workqueue to handle multiple resources at the same time. The
workqueue is also a delta queue so multiple changes get combined into one processing run.

```bash
make run-controllers/crd-configured/workqueue
```
