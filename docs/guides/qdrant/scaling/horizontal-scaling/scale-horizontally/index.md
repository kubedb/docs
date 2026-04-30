---
title: Scale Qdrant Horizontally
menu:
  docs_{{ .version }}:
    identifier: qdrant-scale-horizontally
    name: Scale Horizontally
    parent: qdrant-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scaling for Qdrant

This guide shows how to scale Qdrant nodes horizontally using `QdrantOpsRequest`.

## Before You Begin

- Install KubeDB Community and Enterprise operators.
- Deploy Qdrant in distributed mode.

## Apply HorizontalScaling OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdrant-horizontal-scale
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: qdrant-sample
  horizontalScaling:
    node: 5
```

```bash
$ kubectl apply -f qdrant-horizontal-scale.yaml
qdrantopsrequest.ops.kubedb.com/qdrant-horizontal-scale created
```

## Verify

```bash
$ kubectl get qdrantopsrequest -n demo qdrant-horizontal-scale
$ kubectl get pod -n demo -l app.kubernetes.io/instance=qdrant-sample
```

## Cleaning up

```bash
kubectl delete qdrantopsrequest -n demo qdrant-horizontal-scale
```