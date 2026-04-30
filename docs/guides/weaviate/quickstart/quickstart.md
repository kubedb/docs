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

This tutorial shows how to run Weaviate with KubeDB.

> Note: YAML files used in this tutorial are stored in [docs/examples/weaviate/quickstart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/weaviate/quickstart).

## Before You Begin

- Prepare a Kubernetes cluster and `kubectl`.
- Install KubeDB from [/docs/setup/README.md](/docs/setup/README.md).
- This tutorial uses `docs/examples/weaviate/quickstart/weaviate.yaml` as the working example manifest.
- Create namespace:

```bash
kubectl create ns demo
```

## Check Available StorageClass

```bash
kubectl get storageclass
```

## Check Available WeaviateVersion

```bash
kubectl get weaviateversions
```

## Create a Weaviate Database

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-sample
  namespace: demo
spec:
  version: 1.33.1
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/quickstart/weaviate.yaml
kubectl get weaviate -n demo weaviate-sample -w
```

## Verify Weaviate Database

```bash
kubectl get weaviate -n demo
kubectl describe weaviate -n demo weaviate-sample
```

When `status.phase` becomes `Ready`, the Weaviate cluster is ready for schema and vector indexing requests.

## Cleaning up

```bash
kubectl delete weaviate -n demo weaviate-sample
kubectl delete ns demo
```