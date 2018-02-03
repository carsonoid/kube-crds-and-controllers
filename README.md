# Kuberenetes CustomResourceDefinitions and Controllers

## The Goal

Watch all pods in the cluster and make sure they all have a configurable set of labels.

### The Projects

To illustrate basic functionality and common pitfalls this is broken up into three different projects:

### hard-coded-controller (Bad)

Make a custom controller which has everything hard-coded

### configmap-configured-controller (Better)

Use a ConfigMap to configure the controller

### crd-configured-controller (Best)

Use a CustomResourceDefinition to provide configurations to the controller
