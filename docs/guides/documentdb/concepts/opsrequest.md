---
title: DocumentDBOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: documentdb-opsrequest-concepts
    name: DocumentDBOpsRequest
    parent: documentdb-concepts-documentdb
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# DocumentDBOpsRequest

## What is DocumentDBOpsRequest

`DocumentDBOpsRequest` is the operations CRD KubeDB uses for day-2 lifecycle changes of DocumentDB databases when supported in a release.

## Current support status

Based on [new_db.md](/new_db.md), no DocumentDB operation types are currently listed.

## Expected CRD shape

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DocumentDBOpsRequest
metadata:
  name: documentdb-ops-sample
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: documentdb
```

## Next Steps

- Track [DocumentDB ops overview](/docs/guides/documentdb/ops-request/overview.md) for support updates.
- Use [DocumentDB configuration guide](/docs/guides/documentdb/configuration/overview.md) for declarative spec updates.