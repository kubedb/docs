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

This tutorial shows how to run Qdrant with KubeDB.

> Note: YAML files used in this tutorial are stored in [docs/examples/qdrant/quickstart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/qdrant/quickstart).

## Before You Begin

- Prepare a Kubernetes cluster and `kubectl`.
- Install KubeDB from [/docs/setup/README.md](/docs/setup/README.md).
- This tutorial uses `docs/examples/qdrant/quickstart/distributed.yaml` as the working example manifest.
- Create namespace:

```bash
kubectl create ns demo
```

## Check Available StorageClass

```bash
kubectl get storageclass
```

## Check Available QdrantVersion

```bash
kubectl get qdrantversions
```

## Create a Qdrant Database

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
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/quickstart/distributed.yaml
kubectl get qdrant -n demo qdrant-sample -w
```

## Verify Qdrant Database

```bash
kubectl get qdrant -n demo
kubectl describe qdrant -n demo qdrant-sample
```

When `status.phase` becomes `Ready`, the Qdrant cluster is ready to accept vector search and management requests.

## Cleaning up

```bash
kubectl delete qdrant -n demo qdrant-sample
kubectl delete ns demo
```