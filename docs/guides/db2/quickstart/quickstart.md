---
title: DB2 Quickstart
menu:
  docs_{{ .version }}:
    identifier: db2-quickstart-overview
    name: Overview
    parent: db2-quickstart-db2
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running DB2

This tutorial shows how to run a DB2 database with KubeDB.

> Note: YAML files used in this tutorial are stored in [docs/examples/db2/quickstart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/db2/quickstart).

## Before You Begin

- You need a Kubernetes cluster and `kubectl` configured.
- Install KubeDB following the setup guide: [/docs/setup/README.md](/docs/setup/README.md).
- This tutorial uses `docs/examples/db2/quickstart/standalone.yaml` as the working example manifest.
- Use an isolated namespace:

```bash
kubectl create namespace demo
```

## Check Available StorageClass

```bash
kubectl get storageclass
```

## Check Available DB2Version

```bash
kubectl get db2versions
```

## Create a DB2 Database

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DB2
metadata:
  name: db2
  namespace: demo
spec:
  version: 11.5.8.0
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
  deletionPolicy: Delete
```

Apply the example manifest:

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/db2/quickstart/standalone.yaml
kubectl get db2 -n demo db2 -w
```

## Verify DB2 Database

```bash
kubectl get db2 -n demo
kubectl describe db2 -n demo db2
```

When `status.phase` becomes `Ready`, your DB2 database is ready to use.

## Cleaning up

```bash
kubectl delete db2 -n demo db2
kubectl delete ns demo
```