---
title: QdrantAutoscaler CRD
menu:
  docs_{{ .version }}:
    identifier: qdrant-autoscaler-concepts
    name: QdrantAutoscaler
    parent: qdrant-concepts-qdrant
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# QdrantAutoscaler

## What is QdrantAutoscaler

`QdrantAutoscaler` is documented here as a planned resource, but this repository does not currently contain the matching Go type or CRD.

As a result, the sample shown below is illustrative only and should not be treated as a repository-backed manifest for the current release.

## Sample QdrantAutoscaler

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: QdrantAutoscaler
metadata:
  name: qdrant-as-compute
  namespace: demo
spec:
  databaseRef:
    name: qdrant-sample
  compute:
    node:
      trigger: "On"
      minAllowed:
        cpu: 250m
        memory: 512Mi
      maxAllowed:
        cpu: "2"
        memory: 4Gi
```

## Key fields

- `spec.databaseRef.name` points to the target `Qdrant` database.
- `spec.compute` controls CPU and memory autoscaling behavior.
- `spec.storage` controls volume expansion thresholds and bounds.
- `spec.opsRequestOptions` configures generated ops request behavior.

## Next Steps

- Read [Qdrant autoscaler overview](/docs/guides/qdrant/autoscaler/overview.md).
- See [compute autoscaler guide](/docs/guides/qdrant/autoscaler/compute/overview.md) and [storage autoscaler guide](/docs/guides/qdrant/autoscaler/storage/overview.md).