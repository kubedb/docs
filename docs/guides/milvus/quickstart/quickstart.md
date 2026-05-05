---
title: Milvus Quickstart
menu:
  docs_{{ .version }}:
    identifier: milvus-quickstart-overview
    name: Overview
    parent: milvus-quickstart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Milvus QuickStart

This tutorial shows how to run a Milvus database with KubeDB.

<p align="center">
  <img alt="lifecycle"  src="/docs/images/milvus/milvus-lifecycle.png">
</p>

> Note: YAML files used in this tutorial are stored in [docs/examples/milvus/quickstart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/milvus/quickstart).

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md) and make sure to include the flags `--set global.featureGates.Milvus=true` to ensure **Milvus CRD**.

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create namespace demo
namespace/demo created

$ kubectl get namespace
NAME                 STATUS   AGE
demo                 Active   9s
```

## Check Available MilvusVersion

When you install the KubeDB operator, it registers a CRD named [MilvusVersion](/docs/guides/milvus/concepts/catalog.md). The installation process comes with a set of tested MilvusVersion objects. Let's check available MilvusVersions by,

```bash
$ kubectl get milvusversions
NAME     VERSION   DB_IMAGE                                DEPRECATED   AGE
2.6.11   2.6.11    ghcr.io/appscode-images/milvus:2.6.11                6d2h
2.6.7    2.6.7     ghcr.io/appscode-images/milvus:2.6.7                 6d2h
2.6.9    2.6.9     ghcr.io/appscode-images/milvus:2.6.9                 6d2h
```

Notice the `DEPRECATED` column. Here, `true` means that this MilvusVersion is deprecated for the current KubeDB version. KubeDB will not work for deprecated MilvusVersion. You can also use the short from `mvversion` to check available MilvusVersions.

In this tutorial, we will use `2.6.11` MilvusVersion CR to create a Milvus cluster.

## Get External Dependencies Ready

### Object Storage

One of the external dependency of Milvus is object storage where the segments are stored. It is a storage mechanism that Milvus does not provide. **S3-compatible storage** (like **Minio**) are generally convenient options for object storage.

In this tutorial, we will run a `minio-server` as object storage in our local `kind` cluster using `minio-operator` and create a bucket named `milvus` in it, which the deployed milvus database will use.

```bash

$ helm repo add minio https://operator.min.io/
$ helm repo update minio
$ helm upgrade --install --namespace "minio-operator" --create-namespace "minio-operator" minio/operator --set operator.replicaCount=1

$ helm upgrade --install --namespace "demo" --create-namespace milvus-minio minio/tenant \
--set tenant.pools[0].servers=1 \
--set tenant.pools[0].volumesPerServer=1 \
--set tenant.pools[0].size=1Gi \
--set tenant.certificate.requestAutoCert=false \
--set tenant.buckets[0].name="milvus" \
--set tenant.pools[0].name="default"

```

Now we need to create a `Secret` named `my-release-minio`. It contains the necessary connection information using which the milvus database will connect to the object storage.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-release-minio
  namespace: demo
stringData:
  milvus.storage.type: "s3"
  milvus.storage.bucket: "milvus"
  milvus.storage.baseKey: "milvus/segments"
  milvus.s3.accessKey: "minio"
  milvus.s3.secretKey: "minio123"
  milvus.s3.protocol: "http"
  milvus.s3.enablePathStyleAccess: "true"
  milvus.s3.endpoint.signingRegion: "us-east-1"
  milvus.s3.endpoint.url: "http://myminio-hl.demo.svc.cluster.local:9000/"
```

Let’s create the `deep-storage-config` Secret shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/milvus/quickstart/deep-storage-config.yaml
secret/deep-storage-config created
```

You can also use options like **Amazon S3**, **Google Cloud Storage**, **Azure Blob Storage** or **HDFS** and create a connection information `Secret` like this, and you are good to go.

## Create a Milvus Database

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-cluster
  namespace: demo
spec:
  version: "2.6.11"
  objectStorage:
    configSecret:
      name: "my-release-minio"
  topology:
    mode: Distributed
    distributed:
      mixcoord:
        replicas: 2
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    storageClassName: local-path
    resources:
      requests:
        storage: 10Gi
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/milvus/quickstart/distributed.yaml
kubectl get milvus -n demo milvus-cluster -w
```

## Verify Milvus Database

```bash
kubectl get milvus -n demo
kubectl describe milvus -n demo milvus-cluster
```

When `status.phase` becomes `Ready`, the Milvus deployment is ready to serve vector search traffic.

## Cleaning up

```bash
kubectl delete milvus -n demo milvus-cluster
kubectl delete ns demo
```