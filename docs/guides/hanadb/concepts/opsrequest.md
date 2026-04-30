---
title: HanaDBOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: hanadb-opsrequest-concepts
    name: HanaDBOpsRequest
    parent: hanadb-concepts-hanadb
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# HanaDBOpsRequest

## What is HanaDBOpsRequest

`HanaDBOpsRequest` is the operations CRD KubeDB uses for day-2 lifecycle changes of HanaDB databases when supported in a release.

## Current support status

Based on [new_db.md](/new_db.md), no HanaDB operation types are currently listed.

## Expected CRD shape

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hanadb-ops-sample
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: hana-cluster
```

## Next Steps

- Track [HanaDB ops overview](/docs/guides/hanadb/ops-request/overview.md) for support updates.
- Use [HanaDB monitoring guide](/docs/guides/hanadb/monitoring/overview.md) for day-2 operational readiness.