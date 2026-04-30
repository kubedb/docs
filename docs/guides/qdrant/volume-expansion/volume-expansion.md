---
title: Expand Qdrant Volume
menu:
  docs_{{ .version }}:
    identifier: qdrant-volume-expansion-cluster
    name: Cluster
    parent: qdrant-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Expand Qdrant Volume

This guide shows how to expand persistent volume size for Qdrant nodes using `QdrantOpsRequest`.

## Before You Begin

- Your `StorageClass` must support volume expansion.
- Install KubeDB Community and Enterprise operators.

```bash
$ kubectl get storageclass
```

## Apply VolumeExpansion OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdrant-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: qdrant-sample
  volumeExpansion:
    mode: Online
    node: 5Gi
```

```bash
$ kubectl apply -f qdrant-volume-expansion.yaml
qdrantopsrequest.ops.kubedb.com/qdrant-volume-expansion created
```

## Verify

```bash
$ kubectl get qdrantopsrequest -n demo qdrant-volume-expansion
$ kubectl describe qdrantopsrequest -n demo qdrant-volume-expansion
```