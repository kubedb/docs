---
title: Refreshing a PostgreSQL Branch
menu:
  docs_{{ .version }}:
    identifier: pg-branch-refresh
    name: Refresh
    parent: guides-pg-branch
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Refreshing a PostgreSQL Branch

This tutorial periodically replaces a PostgreSQL branch with a new snapshot of its source. Read the common [refresh guide](/docs/operatormanual/branching/refresh.md) for scheduling and lifecycle behavior.

## Before You Begin

Prepare and seed `demo/prod-pg` using the [same-cluster](/docs/guides/postgres/branch/same-cluster/index.md) or [cross-cluster](/docs/guides/postgres/branch/cross-cluster/index.md) tutorial. Retain the source and target namespaces for this tutorial.

These commands use explicit contexts. For the same-cluster environment, initialize:

```bash
SOURCE_CONTEXT=$(kubectl config current-context)
TARGET_CONTEXT="$SOURCE_CONTEXT"
```

For the cross-cluster environment, set `SOURCE_CONTEXT=source` and `TARGET_CONTEXT=target`, and add `spec.target.clusterName: target` to the downloaded Branch manifest **before creating it**. Apply it on the source cluster. The `demo/prod-pg` source and `dev` target namespace must already exist as described in the creation tutorials.

## Create a scheduled Branch

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: pg-refresh
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Postgres
      name: prod-pg
    namespace: demo
  target:
    namespace: dev
    name: pg-refresh
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  schedule:
    cron: "*/10 * * * *"
  historyLimit:
    success: 3
    failed: 2
  deletionPolicy: Delete
```

The ten-minute schedule leaves time to inspect the branch. Refresh overwrites branch-only data and interrupts target availability during storage replacement and restart. Select a production schedule according to the target application's needs.

```bash
curl -fsSL -o branch-refresh.yaml https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/branch/branch-refresh.yaml
# For cross-cluster use, add spec.target.clusterName: target to this file first.
kubectl --context "$SOURCE_CONTEXT" apply -f branch-refresh.yaml
kubectl --context "$SOURCE_CONTEXT" -n demo wait branch/pg-refresh \
  --for=jsonpath='{.status.phase}'=Ready --timeout=20m
```

## Change source and branch data

Run these checks after initial readiness and before the next scheduled refresh. Record the current generation and locate the elected primaries:

```bash
BEFORE_GENERATION=$(kubectl --context "$SOURCE_CONTEXT" -n demo get branch pg-refresh \
  -o jsonpath='{.status.refreshGeneration}')
NEXT_GENERATION=$((${BEFORE_GENERATION:-0} + 1))
SOURCE_POD=$(kubectl --context "$SOURCE_CONTEXT" -n demo get pods \
  -l app.kubernetes.io/instance=prod-pg,kubedb.com/role=primary \
  -o jsonpath='{.items[0].metadata.name}')
BRANCH_POD=$(kubectl --context "$TARGET_CONTEXT" -n dev get pods \
  -l app.kubernetes.io/instance=pg-refresh,kubedb.com/role=primary \
  -o jsonpath='{.items[0].metadata.name}')
BRANCH_SECRET=$(kubectl --context "$TARGET_CONTEXT" -n dev get postgres pg-refresh -o jsonpath='{.spec.authSecret.name}')
BRANCH_USER=$(kubectl --context "$TARGET_CONTEXT" -n dev get secret "$BRANCH_SECRET" -o jsonpath='{.data.username}' | base64 -d)
BRANCH_PASSWORD=$(kubectl --context "$TARGET_CONTEXT" -n dev get secret "$BRANCH_SECRET" -o jsonpath='{.data.password}' | base64 -d)
kubectl --context "$SOURCE_CONTEXT" -n demo exec "$SOURCE_POD" -c postgres -- \
  psql -U postgres -d postgres -v ON_ERROR_STOP=1 \
  -c "INSERT INTO branch_demo VALUES (300, 'refresh-source@example.com');"
kubectl --context "$TARGET_CONTEXT" -n dev exec "$BRANCH_POD" -c postgres -- \
  env PGPASSWORD="$BRANCH_PASSWORD" psql -h pg-refresh -U "$BRANCH_USER" -d postgres \
  -v ON_ERROR_STOP=1 -c "INSERT INTO branch_demo VALUES (900, 'refresh-branch@example.com');"
kubectl --context "$TARGET_CONTEXT" -n dev exec "$BRANCH_POD" -c postgres -- \
  env PGPASSWORD="$BRANCH_PASSWORD" psql -h pg-refresh -U "$BRANCH_USER" -d postgres \
  -c 'SELECT id FROM branch_demo WHERE id IN (300, 900) ORDER BY id;'
```

Before refresh, the branch query returns `900` and does not return `300`.

## Wait for replacement and verify

Wait for the next completed generation and readiness:

```bash
kubectl --context "$SOURCE_CONTEXT" -n demo wait branch/pg-refresh \
  --for=jsonpath='{.status.refreshGeneration}'="$NEXT_GENERATION" --timeout=20m
kubectl --context "$SOURCE_CONTEXT" -n demo wait branch/pg-refresh \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
kubectl --context "$SOURCE_CONTEXT" -n demo get branch pg-refresh -o yaml
```

Rediscover the target pod and read its current Secret after replacement:

```bash
BRANCH_POD=$(kubectl --context "$TARGET_CONTEXT" -n dev get pods \
  -l app.kubernetes.io/instance=pg-refresh,kubedb.com/role=primary \
  -o jsonpath='{.items[0].metadata.name}')
BRANCH_SECRET=$(kubectl --context "$TARGET_CONTEXT" -n dev get postgres pg-refresh -o jsonpath='{.spec.authSecret.name}')
BRANCH_USER=$(kubectl --context "$TARGET_CONTEXT" -n dev get secret "$BRANCH_SECRET" -o jsonpath='{.data.username}' | base64 -d)
BRANCH_PASSWORD=$(kubectl --context "$TARGET_CONTEXT" -n dev get secret "$BRANCH_SECRET" -o jsonpath='{.data.password}' | base64 -d)
kubectl --context "$TARGET_CONTEXT" -n dev exec "$BRANCH_POD" -c postgres -- \
  env PGPASSWORD="$BRANCH_PASSWORD" psql -h pg-refresh -U "$BRANCH_USER" -d postgres \
  -c 'SELECT id FROM branch_demo WHERE id IN (300, 900) ORDER BY id;'
unset BRANCH_PASSWORD
```

The query now returns `300`; branch-only row `900` is gone. Check `lastSuccessfulRefreshTime`, `nextRefreshTime`, `refreshGeneration`, and `history` on the source Branch. With `resetRootPassword` or post-actions configured, Courier completes those steps again before readiness.

## Cleanup

```bash
kubectl --context "$SOURCE_CONTEXT" -n demo delete branch pg-refresh --wait=true --timeout=15m
kubectl --context "$SOURCE_CONTEXT" -n demo exec "$SOURCE_POD" -c postgres -- \
  psql -U postgres -d postgres -c 'DELETE FROM branch_demo WHERE id = 300;'
```
