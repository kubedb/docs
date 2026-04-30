---
title: Qdrant RBAC
menu:
  docs_{{ .version }}:
    identifier: qdrant-quickstart-rbac
    name: RBAC
    parent: qdrant-quickstart
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Qdrant with RBAC Enabled

This tutorial shows how to run Qdrant with the RBAC permissions required by KubeDB.

## Before You Begin

- Prepare a Kubernetes cluster and `kubectl`.
- Install KubeDB from [/docs/setup/README.md](/docs/setup/README.md).

```bash
kubectl create ns demo
```

## Deploy Qdrant

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-rbac
  namespace: demo
spec:
  version: 1.17.0
  mode: Distributed
  replicas: 3
  storageType: Durable
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
kubectl get qdrant -n demo qdrant-rbac -w
```

## Cleaning up

```bash
kubectl delete qdrant -n demo qdrant-rbac
kubectl delete ns demo
```