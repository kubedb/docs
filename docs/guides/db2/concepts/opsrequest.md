---
title: DB2OpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: db2-opsrequest-concepts
    name: DB2OpsRequest
    parent: db2-concepts-db2
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# DB2OpsRequest

## What is DB2OpsRequest

`DB2OpsRequest` is the operations CRD KubeDB uses for day-2 lifecycle changes of DB2 databases when that feature is available in a release.

## Current support status

Based on [new_db.md](/new_db.md), no DB2 operation types are currently listed.

## Expected CRD shape

When available, an ops request follows this structure:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DB2OpsRequest
metadata:
  name: db2-ops-sample
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: db2
```

## Next Steps

- Track [DB2 ops overview](/docs/guides/db2/ops-request/overview.md) for support updates.
- Use [DB2 configuration guide](/docs/guides/db2/configuration/overview.md) for declarative spec updates.