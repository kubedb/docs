---
title: Resetting a PostgreSQL Branch Root Password
menu:
  docs_{{ .version }}:
    identifier: pg-branch-reset-root-password
    name: ResetRootPassword
    parent: guides-pg-branch
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Resetting a PostgreSQL Branch Root Password

Set `spec.resetRootPassword: true` to give a PostgreSQL branch its own root credential. The source credential remains unchanged. See the common [ResetRootPassword guide](/docs/operatormanual/branching/reset-root-password.md) for the shared execution order.

## Before You Begin

Prepare the HA source and seed data using a [creation tutorial](/docs/guides/postgres/branch/overview/index.md).

These commands use explicit contexts. For the same-cluster environment, initialize:

```bash
SOURCE_CONTEXT=$(kubectl config current-context)
TARGET_CONTEXT="$SOURCE_CONTEXT"
```

For the cross-cluster environment, set `SOURCE_CONTEXT=source` and `TARGET_CONTEXT=target`, and add `spec.target.clusterName: target` to the downloaded Branch manifest **before creating it**. Apply it on the source cluster. The `demo/prod-pg` source and `dev` target namespace must already exist as described in the creation tutorials.

## Create the Branch

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: pg-reset
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
    name: pg-reset
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  resetRootPassword: true
  deletionPolicy: Delete
```

For PostgreSQL, Courier leaves the target's auth Secret for KubeDB to generate, and the branching-capable PostgreSQL startup applies that credential to the cloned data. This path does not require a password-rotation OpsRequest.

```bash
curl -fsSL -o branch-reset-root-password.yaml https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/branch/branch-reset-root-password.yaml
# For cross-cluster use, add spec.target.clusterName: target to this file first.
kubectl --context "$SOURCE_CONTEXT" apply -f branch-reset-root-password.yaml
kubectl --context "$SOURCE_CONTEXT" -n demo wait branch/pg-reset \
  --for=jsonpath='{.status.phase}'=Ready --timeout=20m
```

## Verify credential isolation through the Service

Read the source and branch credentials without printing decoded passwords:

```bash
SOURCE_SECRET=$(kubectl --context "$SOURCE_CONTEXT" -n demo get postgres prod-pg -o jsonpath='{.spec.authSecret.name}')
SOURCE_USER=$(kubectl --context "$SOURCE_CONTEXT" -n demo get secret "$SOURCE_SECRET" -o jsonpath='{.data.username}' | base64 -d)
SOURCE_PASSWORD=$(kubectl --context "$SOURCE_CONTEXT" -n demo get secret "$SOURCE_SECRET" -o jsonpath='{.data.password}' | base64 -d)
BRANCH_SECRET=$(kubectl --context "$TARGET_CONTEXT" -n dev get postgres pg-reset -o jsonpath='{.spec.authSecret.name}')
BRANCH_USER=$(kubectl --context "$TARGET_CONTEXT" -n dev get secret "$BRANCH_SECRET" -o jsonpath='{.data.username}' | base64 -d)
BRANCH_PASSWORD=$(kubectl --context "$TARGET_CONTEXT" -n dev get secret "$BRANCH_SECRET" -o jsonpath='{.data.password}' | base64 -d)
BRANCH_POD=$(kubectl --context "$TARGET_CONTEXT" -n dev get pods \
  -l app.kubernetes.io/instance=pg-reset,kubedb.com/role=primary \
  -o jsonpath='{.items[0].metadata.name}')
```

First, verify that the branch's own password authenticates and the seeded data is present:

```bash
kubectl --context "$TARGET_CONTEXT" -n dev exec "$BRANCH_POD" -c postgres -- \
  env PGPASSWORD="$BRANCH_PASSWORD" PGCONNECT_TIMEOUT=10 \
  psql -w -h pg-reset -U "$BRANCH_USER" -d postgres \
  -v ON_ERROR_STOP=1 -c 'SELECT id, email FROM branch_demo WHERE id IN (1, 2) ORDER BY id;'
```

Both seeded rows should be returned. Repeat against the same Service with the **source** password:

```bash
kubectl --context "$TARGET_CONTEXT" -n dev exec "$BRANCH_POD" -c postgres -- \
  env PGPASSWORD="$SOURCE_PASSWORD" PGCONNECT_TIMEOUT=10 \
  psql -w -h pg-reset -U "$BRANCH_USER" -d postgres -c 'SELECT 1;'
```

This command should fail with `password authentication failed`. A connection timeout or another error does not prove credential isolation. Both checks use `-h pg-reset` so that PostgreSQL performs Service-path authentication.

Confirm that the original source credential still works on the source Service:

```bash
SOURCE_POD=$(kubectl --context "$SOURCE_CONTEXT" -n demo get pods \
  -l app.kubernetes.io/instance=prod-pg,kubedb.com/role=primary \
  -o jsonpath='{.items[0].metadata.name}')
kubectl --context "$SOURCE_CONTEXT" -n demo exec "$SOURCE_POD" -c postgres -- \
  env PGPASSWORD="$SOURCE_PASSWORD" PGCONNECT_TIMEOUT=10 \
  psql -w -h prod-pg -U "$SOURCE_USER" -d postgres -c 'SELECT 1;'
unset SOURCE_PASSWORD BRANCH_PASSWORD
```

After a [scheduled refresh](/docs/guides/postgres/branch/refresh.md), wait for readiness and read the branch Secret again before authenticating. Post-actions use the branch's current credential automatically.

## Cleanup

```bash
kubectl --context "$SOURCE_CONTEXT" -n demo delete branch pg-reset --wait=true --timeout=15m
```
