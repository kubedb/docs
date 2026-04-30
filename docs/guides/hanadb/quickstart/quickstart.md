---
title: HanaDB Quickstart
menu:
  docs_{{ .version }}:
    identifier: hanadb-quickstart-overview
    name: Overview
    parent: hanadb-quickstart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running HanaDB

This tutorial shows how to run a HanaDB database with KubeDB.

> Note: YAML files used in this tutorial are stored in [docs/examples/hanadb/quickstart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/quickstart).

## Before You Begin

- Prepare a Kubernetes cluster and `kubectl`.
- Install KubeDB from [/docs/setup/README.md](/docs/setup/README.md).
- This tutorial uses `docs/examples/hanadb/quickstart/system-replication.yaml` as the working example manifest.
- Create namespace:

```bash
kubectl create ns demo
```

## Check Available StorageClass

```bash
kubectl get storageclass
```

## Check Available HanaDBVersion

```bash
kubectl get hanadbversions
```

## Create a HanaDB Database

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hana-cluster
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 2
  storageType: "Durable"
  topology:
    mode: SystemReplication
    systemReplication:
      replicationMode: fullsync
      operationMode: logreplay_readaccess
  storage:
    accessModes: ["ReadWriteOnce"]
    resources:
      requests:
        storage: 64Gi
    storageClassName: local-path
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/quickstart/system-replication.yaml
kubectl get hanadb -n demo hana-cluster -w
```

## Verify HanaDB Database

```bash
kubectl get hanadb -n demo
kubectl describe hanadb -n demo hana-cluster
```

When `status.phase` becomes `Ready`, the HanaDB deployment is ready for application traffic.

## Cleaning up

```bash
kubectl delete hanadb -n demo hana-cluster
kubectl delete ns demo
```