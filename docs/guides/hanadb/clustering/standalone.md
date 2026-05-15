---
title: HanaDB Standalone
menu:
  docs_{{ .version }}:
    identifier: hanadb-standalone-clustering
    name: Standalone
    parent: hanadb-clustering
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# HanaDB Standalone

This guide shows how to run a single SAP HANA instance using KubeDB.

## Before You Begin

- Prepare a Kubernetes cluster and configure `kubectl`.
- Install KubeDB following the steps [here](/docs/setup/README.md).
- Create a namespace for this tutorial:

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/hanadb/clustering](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/clustering).

## Deploy a Standalone Instance

The following manifest creates a standalone HanaDB instance. If `spec.topology` is omitted, KubeDB treats the database as standalone and requires `spec.replicas: 1`.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hanadb-standalone
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 1
  storageType: Durable
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 64Gi
    storageClassName: local-path
  deletionPolicy: WipeOut
```

Create the database:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/clustering/standalone.yaml
hanadb.kubedb.com/hanadb-standalone created
```

Wait for the database to become ready:

```bash
$ kubectl get hanadb -n demo hanadb-standalone
NAME               VERSION   STATUS   AGE
hanadb-standalone   2.0.82    Ready    5m
```

## Cleaning up

```bash
kubectl delete hanadb -n demo hanadb-standalone
kubectl delete ns demo
```
