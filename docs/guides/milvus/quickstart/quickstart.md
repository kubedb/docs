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

# Running Milvus

This tutorial shows how to run a Milvus database with KubeDB.

> Note: YAML files used in this tutorial are stored in [docs/examples/milvus/quickstart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/milvus/quickstart).

## Before You Begin

- Prepare a Kubernetes cluster and `kubectl`.
- Install KubeDB from [/docs/setup/README.md](/docs/setup/README.md).
- This tutorial uses `docs/examples/milvus/quickstart/distributed.yaml` as the working example manifest.
- Create namespace:

```bash
kubectl create ns demo
```

## Check Available MilvusVersion

```bash
kubectl get milvusversions
```

## Check Object Storage Secret

Milvus requires an object storage backend for metadata and data files. Make sure the secret referenced by the example manifest exists before creating the database.

```bash
kubectl get secret -n demo my-release-minio
```

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