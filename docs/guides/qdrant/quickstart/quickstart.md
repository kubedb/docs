---
title: Qdrant Quickstart
menu:
  docs_{{ .version }}:
    identifier: qdrant-quickstart-overview
    name: Overview
    parent: qdrant-quickstart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running Qdrant

This tutorial will show you how to use KubeDB to run a Qdrant database.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/qdrant/quickstart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/qdrant/quickstart) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Available StorageClass

We will need to provide `StorageClass` in the Qdrant CR specification. Check available `StorageClass` in your cluster using the following command:

```bash
$ kubectl get storageclass
NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  10d
```

Here, we have `standard` StorageClass in our cluster.

## Find Available QdrantVersion

When you install KubeDB, it creates `QdrantVersion` CRDs for all supported Qdrant versions. Let's check available `QdrantVersion`s:

```bash
$ kubectl get qdrantversions
NAME      VERSION   DB_IMAGE                              DEPRECATED   AGE
1.7.4     1.7.4     qdrant/qdrant:v1.7.4                              3d
1.8.4     1.8.4     qdrant/qdrant:v1.8.4                              3d
1.10.0    1.10.0    qdrant/qdrant:v1.10.0                             3d
1.11.0    1.11.0    qdrant/qdrant:v1.11.0                             3d
1.12.0    1.12.0    qdrant/qdrant:v1.12.0                             3d
1.13.0    1.13.0    qdrant/qdrant:v1.13.0                             3d
1.14.0    1.14.0    qdrant/qdrant:v1.14.0                             3d
1.15.0    1.15.0    qdrant/qdrant:v1.15.0                             3d
1.16.0    1.16.0    qdrant/qdrant:v1.16.0                             3d
1.17.0    1.17.0    qdrant/qdrant:v1.17.0                             3d
```

Notice the `DEPRECATED` column. `true` means that QdrantVersion is deprecated for the current KubeDB version and KubeDB will not work for that version.

In this tutorial, we will use `1.17.0` QdrantVersion CR to create a Qdrant cluster. To know more about what `QdrantVersion` CR is and why there may be variation in version names, please visit [here](/docs/guides/qdrant/concepts/catalog.md).

## Create a Qdrant Database

KubeDB implements a `Qdrant` CRD to define the specification of a Qdrant database.

Below is the `Qdrant` object created in this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
  namespace: demo
spec:
  version: "1.17.0"
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/quickstart/qdrant-sample.yaml
qdrant.kubedb.com/qdrant-sample created
```

Here,

- `spec.version` is the name of the QdrantVersion CR where Docker images are specified. In this tutorial, a Qdrant `1.17.0` cluster is created.
- `spec.replicas` specifies the number of Qdrant nodes in the cluster.
- `spec.storage` specifies the size and StorageClass of the PVC that will be dynamically allocated to store data for each Qdrant pod. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods.
- `spec.deletionPolicy` specifies what KubeDB should do when a user tries to delete the `Qdrant` CR. Deletion policy `DoNotTerminate` prevents deletion of this object if the admission webhook is enabled.

> **Note:** `spec.storage` section is used to create PVC for the database pods. Specify only `requests`, not `limits` — PVC does not resize automatically.

Now, let's watch the progress of creating the `Qdrant` cluster:

```bash
$ kubectl get qdrant -n demo qdrant-sample -w
NAME             VERSION   STATUS         AGE
qdrant-sample    1.17.0    Provisioning   5s
qdrant-sample    1.17.0    Provisioning   30s
qdrant-sample    1.17.0    Ready          2m
```

## Describe Qdrant

Let's describe the `Qdrant` object to see its current state:

```bash
$ kubectl describe qdrant -n demo qdrant-sample
Name:         qdrant-sample
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Qdrant
Metadata:
  ...
Spec:
  Auth Secret:
    Name:  qdrant-sample-auth
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
  Version:               1.17.0
Status:
  Conditions:
    Last Transition Time:  2024-10-01T10:00:00Z
    Message:               The KubeDB operator has started the provisioning of Qdrant: demo/qdrant-sample
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2024-10-01T10:01:30Z
    Message:               All desired replicas are ready.
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2024-10-01T10:02:00Z
    Message:               The Qdrant: demo/qdrant-sample is accepting client requests.
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2024-10-01T10:02:00Z
    Message:               DB is ready because of reason
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2024-10-01T10:02:00Z
    Message:               The Qdrant: demo/qdrant-sample is successfully provisioned.
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:  Ready
```

## Connect to Qdrant

KubeDB creates a Secret containing authentication credentials for the Qdrant cluster. Let's check it:

```bash
$ kubectl get secret -n demo qdrant-sample-auth -o yaml
apiVersion: v1
data:
  api-key: <base64-encoded-api-key>
kind: Secret
metadata:
  name: qdrant-sample-auth
  namespace: demo
type: Opaque
```

Now, let's connect to the Qdrant cluster using port forwarding:

```bash
$ kubectl port-forward -n demo svc/qdrant-sample 6333:6333 &
$ export QDRANT_API_KEY=$(kubectl get secret -n demo qdrant-sample-auth -o jsonpath='{.data.api-key}' | base64 -d)

$ curl -H "api-key: $QDRANT_API_KEY" http://localhost:6333/collections
{"result":{"collections":[]},"status":"ok","time":0.001}
```

## Database DeletionPolicy

This tutorial has set `deletionPolicy: DoNotTerminate`. This will prevent you from deleting the database. If you try to delete it, you will get an error. Once you are done experimenting, change the `deletionPolicy` to `WipeOut` before deleting the Qdrant CR:

```bash
$ kubectl patch -n demo qdrant/qdrant-sample -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
qdrant.kubedb.com/qdrant-sample patched
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete qdrant -n demo qdrant-sample
kubectl delete ns demo
```