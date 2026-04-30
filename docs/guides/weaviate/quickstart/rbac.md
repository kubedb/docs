---
title: Weaviate RBAC
menu:
  docs_{{ .version }}:
    identifier: weaviate-quickstart-rbac
    name: RBAC
    parent: weaviate-quickstart
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Weaviate with RBAC Enabled

This tutorial shows how to run Weaviate with the RBAC permissions required by KubeDB.

## Before You Begin

- You need a Kubernetes cluster and the `kubectl` CLI configured for that cluster.
- Install KubeDB operator from [setup guide](/docs/setup/README.md).
- This tutorial uses a dedicated namespace named `demo`.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Deploy Weaviate

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-rbac
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
$ kubectl apply -f weaviate-rbac.yaml
weaviate.kubedb.com/weaviate-rbac created
```

## Verify

```bash
$ kubectl get weaviate -n demo weaviate-rbac
NAME            VERSION   STATUS   AGE
weaviate-rbac   1.33.1    Ready    2m
```

## Cleaning up

```bash
kubectl delete weaviate -n demo weaviate-rbac
kubectl delete ns demo
```