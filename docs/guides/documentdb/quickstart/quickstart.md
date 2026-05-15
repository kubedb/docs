---
title: DocumentDB Quickstart
menu:
  docs_{{ .version }}:
    identifier: documentdb-quickstart-overview
    name: Overview
    parent: documentdb-quickstart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running DocumentDB

This tutorial shows how to run a DocumentDB database with KubeDB.

> Note: YAML files used in this tutorial are stored in [docs/examples/documentdb/quickstart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/documentdb/quickstart).

## Before You Begin

- Ensure you have a Kubernetes cluster and `kubectl` access.
- Install KubeDB from [/docs/setup/README.md](/docs/setup/README.md).
- This tutorial uses `docs/examples/documentdb/quickstart/standalone.yaml` as the working example manifest.
- Create namespace:

```bash
kubectl create ns demo
```

## Check Available StorageClass

```bash
kubectl get storageclass
```

## Check Available DocumentDBVersion

```bash
kubectl get documentdbversions
```

## Create a DocumentDB Database

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: documentdb
  namespace: demo
spec:
  version: "pg17-0.109.0"
  storageType: Durable
  deletionPolicy: Delete
  replicas: 1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/documentdb/quickstart/standalone.yaml
kubectl get documentdb -n demo documentdb -w
```

## Verify DocumentDB Database

```bash
kubectl get documentdb -n demo
kubectl describe documentdb -n demo documentdb
```

When `status.phase` is `Ready`, the database is ready to accept connections.

## Cleaning up

```bash
kubectl delete documentdb -n demo documentdb
kubectl delete ns demo
```