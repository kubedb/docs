---
title: PostgresAutoscaler CRD
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-concepts-autoscaler
    name: PostgresAutoscaler
    parent: pg-concepts-postgres
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PostgresAutoscaler

## What is PostgresAutoscaler

`PostgresAutoscaler` is the CRD that enables compute and storage autoscaling workflows for KubeDB-managed PostgreSQL databases.

KubeDB Autoscaler observes this resource and creates appropriate `PostgresOpsRequest` objects when scaling actions are needed.

## Sample PostgresAutoscaler

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: PostgresAutoscaler
metadata:
  name: pg-as
  namespace: demo
spec:
  databaseRef:
    name: ha-postgres
  opsRequestOptions:
    timeout: 3m
    apply: IfReady
  compute:
    postgres:
      trigger: "On"
      minAllowed:
        cpu: 250m
        memory: 1Gi
      maxAllowed:
        cpu: "1"
        memory: 2Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: RequestsAndLimits
  storage:
    postgres:
      trigger: "On"
      expansionMode: Online
      usageThreshold: 70
      scalingThreshold: 50
      minAllowed: 2Gi
      maxAllowed: 50Gi
```

## Key fields

- `spec.databaseRef.name` points to the target `Postgres` database.
- `spec.compute` configures CPU and memory autoscaling behavior.
- `spec.storage` configures storage expansion behavior.
- `spec.opsRequestOptions` configures generated ops request behavior.

## Next Steps

- See [Postgres compute autoscaling overview](/docs/guides/postgres/autoscaler/compute/overview.md).
- See [Postgres storage autoscaling overview](/docs/guides/postgres/autoscaler/storage/overview.md).