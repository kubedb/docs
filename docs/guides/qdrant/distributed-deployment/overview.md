---
title: Distributed Deployment
menu:
  docs_{{ .version }}:
    identifier: qdrant-distributed-deployment-overview
    name: Overview
    parent: qdrant-distributed-deployment
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Qdrant Distributed Deployment

This tutorial will show you how to deploy a Qdrant database in distributed mode using KubeDB.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/qdrant/quickstart](/docs/examples/qdrant/quickstart) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

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
NAME      VERSION   DB_IMAGE                                       DEPRECATED   AGE
1.17.0   1.17.0    docker.io/qdrant/qdrant:v1.17.0-unprivileged                13d
```

In this tutorial, we will use `1.17.0` QdrantVersion CR to create a distributed Qdrant cluster.

## Deploy Distributed Qdrant

KubeDB implements a `Qdrant` CRD to define the specification of a Qdrant database. For distributed deployment, you need to set `spec.mode` to `Distributed` and specify the number of replicas.

Below is the `Qdrant` object created in this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
  namespace: demo
spec:
  version: 1.17.0
  mode: Distributed
  replicas: 3
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 200Mi
  podTemplate:
    spec:
      containers:
        - name: qdrant
          resources:
            requests:
              cpu: 100m
              memory: 100Mi
            limits:
              cpu: 100m
              memory: 100Mi
  deletionPolicy: WipeOut
```

Here,
- `spec.version` specifies the version of Qdrant to use
- `spec.mode` set to `Distributed` enables distributed mode
- `spec.replicas` specifies the number of Qdrant nodes (default is 1)
- `spec.storage` specifies the storage configuration for each node
- `spec.podTemplate` allows setting resource limits and requests

Let's create the Qdrant object:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/distributed-deployment/yamls/qdrant-distributed.yaml
qdrant.kubedb.com/qdrant-sample created
```

## Verify the Deployment

Let's check the status of the Qdrant object:

```bash
$ kubectl get qdrant -n demo
NAME             VERSION   STATUS   AGE
qdrant-sample    1.17.0    Ready    2m
```

To see the distributed nodes, check the pods:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=qdrant-sample
NAME               READY   STATUS    RESTARTS   AGE
qdrant-sample-0    1/1     Running   0          2m
qdrant-sample-1    1/1     Running   0          2m
qdrant-sample-2    1/1     Running   0          2m
```

In distributed mode, Qdrant creates a StatefulSet with the specified number of replicas.

## Horizontal Scaling

You can scale the Qdrant cluster horizontally by increasing or decreasing the number of replicas using `QdrantOpsRequest`.

To scale up to 5 nodes:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdrant-hor-scaling
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: qdrant-sample
  horizontalScaling:
    node: 5
```

Apply the scaling request:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/distributed-deployment/yamls/qdrant-hor-scaling.yaml
qdrantopsrequest.ops.kubedb.com/qdrant-hor-scaling created
```

Monitor the scaling operation:

```bash
$ kubectl get qdrantopsrequest -n demo
NAME                  TYPE                PHASE   AGE
qdrant-hor-scaling    HorizontalScaling   Done    2m
```

Verify the new replica count:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=qdrant-sample
NAME               READY   STATUS    RESTARTS   AGE
qdrant-sample-0    1/1     Running   0          5m
qdrant-sample-1    1/1     Running   0          5m
qdrant-sample-2    1/1     Running   0          5m
qdrant-sample-3    1/1     Running   0          2m
qdrant-sample-4    1/1     Running   0          2m
```

## Cleaning Up

To delete the Qdrant database and all associated resources:

```bash
$ kubectl delete qdrant -n demo qdrant-sample
qdrant.kubedb.com "qdrant-sample" deleted
```

> **Warning:** If you delete the Qdrant object with `deletionPolicy: WipeOut`, all data will be permanently deleted.
