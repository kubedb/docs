---
title: Weaviate Quickstart
menu:
  docs_{{ .version }}:
    identifier: weaviate-quickstart-overview
    name: Overview
    parent: weaviate-quickstart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running Weaviate

This tutorial will show you how to use KubeDB to run a Weaviate vector database.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/weaviate/quickstart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/weaviate/quickstart) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Available StorageClass

We will need to provide `StorageClass` in the Weaviate CR specification. Check available `StorageClass` in your cluster using the following command:

```bash
$ kubectl get storageclass
NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  10d
```

Here, we have `standard` StorageClass in our cluster.

## Find Available WeaviateVersion

When you install KubeDB, it creates `WeaviateVersion` CRDs for all supported Weaviate versions. Let's check available `WeaviateVersion`s:

```bash
$ kubectl get weaviateversions
NAME      VERSION   DB_IMAGE                        DEPRECATED   AGE
1.25.0    1.25.0    kubedb/weaviate:1.25.0                       3d
1.28.0    1.28.0    kubedb/weaviate:1.28.0                       3d
1.30.0    1.30.0    kubedb/weaviate:1.30.0                       3d
1.33.1    1.33.1    kubedb/weaviate:1.33.1                       3d
```

Notice the `DEPRECATED` column. `true` means that WeaviateVersion is deprecated for the current KubeDB version and KubeDB will not work for that version.

In this tutorial, we will use `1.33.1` WeaviateVersion CR to create a Weaviate cluster. To know more about what `WeaviateVersion` CR is and why there may be variation in version names, please visit [here](/docs/guides/weaviate/concepts/catalog.md).

## Create a Weaviate Database

KubeDB implements a `Weaviate` CRD to define the specification of a Weaviate database.

Below is the `Weaviate` object created in this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-sample
  namespace: demo
spec:
  version: "1.33.1"
  replicas: 3
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: DoNotTerminate
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/quickstart/weaviate-sample.yaml
weaviate.kubedb.com/weaviate-sample created
```

Here,

- `spec.version` is the name of the WeaviateVersion CR where Docker images are specified. In this tutorial, a Weaviate `1.33.1` cluster is created.
- `spec.replicas` specifies the number of Weaviate nodes in the cluster.
- `spec.storage` specifies the size and StorageClass of the PVC that will be dynamically allocated to store data for each Weaviate pod. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods.
- `spec.deletionPolicy` specifies what KubeDB should do when a user tries to delete the `Weaviate` CR. Deletion policy `DoNotTerminate` prevents deletion of this object if the admission webhook is enabled.

> **Note:** `spec.storage` section is used to create PVC for the database pods. Specify only `requests`, not `limits` — PVC does not resize automatically.

Now, let's watch the progress of creating the `Weaviate` cluster:

```bash
$ kubectl get weaviate -n demo weaviate-sample -w
NAME               VERSION   STATUS         AGE
weaviate-sample    1.33.1    Provisioning   5s
weaviate-sample    1.33.1    Provisioning   30s
weaviate-sample    1.33.1    Ready          2m
```

## Describe Weaviate

Let's describe the `Weaviate` object to see its current state:

```bash
$ kubectl describe weaviate -n demo weaviate-sample
Name:         weaviate-sample
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Weaviate
Metadata:
  ...
Spec:
  Auth Secret:
    Name:  weaviate-sample-auth
  Deletion Policy:  DoNotTerminate
  Replicas:         3
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         1Gi
    Storage Class Name:  standard
  Storage Type:          Durable
  Version:               1.33.1
Status:
  Conditions:
    Last Transition Time:  2024-10-01T10:00:00Z
    Message:               The KubeDB operator has started the provisioning of Weaviate: demo/weaviate-sample
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2024-10-01T10:01:30Z
    Message:               All desired replicas are ready.
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2024-10-01T10:02:00Z
    Message:               The Weaviate: demo/weaviate-sample is accepting client requests.
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2024-10-01T10:02:00Z
    Message:               DB is ready because of reason
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2024-10-01T10:02:00Z
    Message:               The Weaviate: demo/weaviate-sample is successfully provisioned.
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:  Ready
```

## Connect to Weaviate

KubeDB creates a Secret containing the API key for the Weaviate cluster. Let's check it:

```bash
$ kubectl get secret -n demo weaviate-sample-auth -o yaml
apiVersion: v1
data:
  api-key: <base64-encoded-api-key>
kind: Secret
metadata:
  name: weaviate-sample-auth
  namespace: demo
type: Opaque
```

Now, let's connect to the Weaviate cluster using port forwarding. Weaviate exposes a REST API on port `8080` and gRPC on port `50051`:

```bash
$ kubectl port-forward -n demo svc/weaviate-sample 8080:8080 &
$ export WEAVIATE_API_KEY=$(kubectl get secret -n demo weaviate-sample-auth -o jsonpath='{.data.api-key}' | base64 -d)

$ curl -H "Authorization: Bearer $WEAVIATE_API_KEY" http://localhost:8080/v1/meta
{"hostname":"http://[::]:8080","modules":{},"version":"1.33.1"}
```

You can also check the cluster health:

```bash
$ curl -H "Authorization: Bearer $WEAVIATE_API_KEY" http://localhost:8080/v1/.well-known/ready
{}
```

And list collections (empty at first):

```bash
$ curl -H "Authorization: Bearer $WEAVIATE_API_KEY" http://localhost:8080/v1/schema
{"classes":[]}
```

## Database DeletionPolicy

This tutorial has set `deletionPolicy: DoNotTerminate`. This will prevent you from deleting the database. If you try to delete it, you will get an error. Once you are done experimenting, change the `deletionPolicy` to `WipeOut` before deleting the Weaviate CR:

```bash
$ kubectl patch -n demo weaviate/weaviate-sample -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
weaviate.kubedb.com/weaviate-sample patched
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete weaviate -n demo weaviate-sample
kubectl delete ns demo
```