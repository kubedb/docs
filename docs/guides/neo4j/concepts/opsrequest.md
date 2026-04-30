---
title: Neo4jOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: neo4j-opsrequest-concepts
    name: Neo4jOpsRequest
    parent: neo4j-concepts-neo4j
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Neo4jOpsRequest

## What is Neo4jOpsRequest

`Neo4jOpsRequest` is the CRD for day-2 operational workflows for KubeDB-managed Neo4j databases.

## Supported operation types

- `Reconfigure`
- `ReconfigureTLS`
- `Restart`
- `RotateAuth`
- `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `VolumeExpansion`

## Sample Neo4jOpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: neo4j-test
```

## Key fields

- `spec.type` selects the operation category.
- `spec.databaseRef.name` identifies the target `Neo4j` object.
- Operation-specific fields are provided under keys like `updateVersion`, `horizontalScaling`, `verticalScaling`, or `volumeExpansion`.
- `spec.timeout` and `spec.apply` can control execution behavior where supported.

## Next Steps

- See [Neo4j ops overview](/docs/guides/neo4j/ops-request/overview.md) for operation links.
- Follow operation tutorials like [Restart](/docs/guides/neo4j/restart/restart.md) and [UpdateVersion](/docs/guides/neo4j/update-version/overview.md).