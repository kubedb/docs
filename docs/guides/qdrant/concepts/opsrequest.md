---
title: QdrantOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: qdrant-opsrequest-concepts
    name: QdrantOpsRequest
    parent: qdrant-concepts-qdrant
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# QdrantOpsRequest

## What is QdrantOpsRequest

`QdrantOpsRequest` is the CRD for day-2 operational workflows for KubeDB-managed Qdrant databases.

## Supported operation types

- `Reconfigure`
- `ReconfigureTLS`
- `Restart`
- `RotateAuth`
- `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `VolumeExpansion`

## Sample QdrantOpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdrant-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: qdrant-sample
```

## Key fields

- `spec.type` selects the operation category.
- `spec.databaseRef.name` identifies the target `Qdrant` object.
- Operation-specific fields are provided under keys like `updateVersion`, `horizontalScaling`, `verticalScaling`, or `volumeExpansion`.
- `spec.timeout` and `spec.apply` can control execution behavior where supported.

## Next Steps

- See [Qdrant ops overview](/docs/guides/qdrant/ops-request/overview.md) for operation links.
- Follow operation tutorials like [Restart](/docs/guides/qdrant/restart/restart.md) and [VolumeExpansion](/docs/guides/qdrant/volume-expansion/overview.md).