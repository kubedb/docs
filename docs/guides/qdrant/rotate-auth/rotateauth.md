---
title: Rotate Auth of Qdrant
menu:
  docs_{{ .version }}:
    identifier: qdrant-rotate-auth-cluster
    name: Cluster
    parent: qdrant-rotate-auth
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Auth for Qdrant

This guide shows how to rotate authentication credentials for Qdrant using `QdrantOpsRequest`.

## Before You Begin

- Install KubeDB Community and Enterprise operators.
- Ensure your target `Qdrant` instance is `Ready`.

```bash
$ kubectl create ns demo
```

## Apply RotateAuth OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdrant-rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: qdrant-sample
```

```bash
$ kubectl apply -f qdrant-rotate-auth.yaml
qdrantopsrequest.ops.kubedb.com/qdrant-rotate-auth created
```

## Verify

```bash
$ kubectl get qdrantopsrequest -n demo qdrant-rotate-auth
$ kubectl describe qdrantopsrequest -n demo qdrant-rotate-auth
```

## Cleaning up

```bash
kubectl delete qdrantopsrequest -n demo qdrant-rotate-auth
kubectl delete ns demo
```