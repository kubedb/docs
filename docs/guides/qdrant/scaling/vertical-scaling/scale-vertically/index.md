---
title: Scale Qdrant Vertically
menu:
  docs_{{ .version }}:
    identifier: qdrant-scale-vertically
    name: Scale Vertically
    parent: qdrant-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scaling for Qdrant

This guide shows how to change CPU and memory resources of Qdrant nodes using `QdrantOpsRequest`.

## Before You Begin

- Install KubeDB Community and Enterprise operators.
- Ensure database is healthy before applying scaling changes.

## Apply VerticalScaling OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdrant-vertical-scale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: qdrant-sample
  verticalScaling:
    node:
      resources:
        requests:
          cpu: "500m"
          memory: "1Gi"
        limits:
          cpu: "1"
          memory: "2Gi"
```

```bash
$ kubectl apply -f qdrant-vertical-scale.yaml
qdrantopsrequest.ops.kubedb.com/qdrant-vertical-scale created
```

## Verify

```bash
$ kubectl get qdrantopsrequest -n demo qdrant-vertical-scale
$ kubectl describe qdrantopsrequest -n demo qdrant-vertical-scale
```