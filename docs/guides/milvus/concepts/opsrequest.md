---
title: MilvusOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: milvus-opsrequest-concepts
    name: MilvusOpsRequest
    parent: milvus-concepts-milvus
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MilvusOpsRequest

## What is MilvusOpsRequest

`MilvusOpsRequest` is the operations CRD KubeDB uses for day-2 lifecycle changes of Milvus databases when supported in a release.

## Current support status

Based on [new_db.md](/new_db.md), no Milvus operation types are currently listed.

## Expected CRD shape

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: milvus-ops-sample
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: milvus-cluster
```

## Next Steps

- Track [Milvus ops overview](/docs/guides/milvus/ops-request/overview.md) for support updates.
- Use [Milvus TLS guide](/docs/guides/milvus/tls/overview.md) and [Milvus monitoring guide](/docs/guides/milvus/monitoring/overview.md) for day-2 operations.