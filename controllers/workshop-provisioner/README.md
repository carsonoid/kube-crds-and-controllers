# workshop-provisioner

## The Problem

A single Kubernetes cluster has been provisioned for the workshop. But multiple users will need to access and manipulate it at once. This means that many sets of resources will need to be created in the cluster. The number of workshop attendees and their names may not be known until the actual workshop starts.

### Requirements

  * Self-Service
  * Automatic provisioning of resources
  * No custom programs or scripting languages should be required
  * Resources constantly ensured. Deleted sub-resources should be recreated
  * Deletes should be easy and delete all resources recursively

## The Solution

Given the problem, the constraints, and the context of the workshop; it makes the most sense to use a Custom Controller to provision users at the time of the workshop. The controller will use a Custom Resource Definition to define workshop attendees and a multi-threaded workqueue to process add, update, and delete events.

## Running

Running the workshop is a muli-setup process.

### Admin Steps

Only required once before the workshop.

#### Setup the Kubernetes cluster

Creating the demo cluster is not covered here. You should have kubectl configured with admin access to it.s

While in the directory this readme is in, create all the files from the setup directory.

```bash
kubectl create -f setup
```

#### Setup the Workshop Admin kubeconfig

The token can be extracted from the kube-public service account created above.

```bash
CLUSTER=CLUSTER
TOKEN=$(kubectl --namespace=kube-public get secret $(kubectl --namespace=kube-public get sa workshop-admin -o jsonpath="{.secrets[0].name}") -o jsonpath="{.data.token}" |base64 -d -w0)
cat > workshopadmin.kubeconfig <<EOL
apiVersion: v1
kind: Config
preferences: {}
current-context: "workshopadmin"
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: $CLUSTER
  name: workshopadmin
contexts:
- context:
    cluster: workshopadmin
    namespace: wa-user1
    user: workshopadmin
  name: workshopadmin
users:
- name: workshopadmin
  user:
    token: $TOKEN
EOL
```

This file will need to be provided to attendees at the start off the workshop.

#### Run the Workshop Provisioner Controller

In the workshop this is done on the main screen so it can be watched by attendees.

```bash
make workshop-provisioner-all run-provisioner OPTS="-cluster-addr https://cluster-addr"
```

### Attendee Steps

#### Create a WorkshopAttendee resource

Setting up your account is as easy as making one Kubernetes Resource. This can be done by making a file and using `kubectl` or via the commands belows.

Set your username and email, then create the resource.

```bash
WAUSER=
WAEMAIL=

kubectl --kubeconfig=workshopadmin.kubeconfig create -f <(cat <<EOL
apiVersion: provisioner.k8s.carsonoid.net/v1alpha1
kind: WorkshopAttendee
metadata:
  name: $WAUSER
spec:
  email: "$WAEMAIL"
EOL
)
```

#### Wait for it to be ready

```
kubectl get wa $WAUSER -o jsonpath="{.status.state}"
```

#### Get your config file from the  status

This can be simply captured via `kubectl` to a file

```bash
kubectl get wa $WAUSER -o jsonpath="{.status.kubeconfig}" > ./podlabeler.kubeconfig
```

### Running the controllers

```bash
export OPTS="-kubeconfig ./podlabeler.kubeconfig"
make run-controllers/hard-coded/simple
make run-controllers/hard-coded/structured
make run-controllers/hard-coded/structured OPTS="$OPTS -config controllers/hard-coded/config.yaml"
make run-controllers/configmap-configured/single-config
make run-controllers/configmap-configured/multi-config
make run-controllers/crd-configured/simple
make run-controllers/crd-configured/workqueue
```
