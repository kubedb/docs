---
title: Branch PostgreSQL Across Clusters
menu:
  docs_{{ .version }}:
    identifier: pg-branch-cross-cluster
    name: Cross Cluster
    parent: guides-pg-branch
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Branch PostgreSQL Across Clusters

This guide creates an HA PostgreSQL source on the `source` managed cluster and branches it onto `target`. Both clusters access the **same network-reachable Ceph RBD backend** with CoW snapshot/clone support.

## Before You Begin

- Complete [Courier cross-cluster setup](/docs/operatormanual/branching/setup-ocm.md) in an existing OCM environment.
- Have `kubectl` contexts named `hub`, `source`, and `target`. The OCM managed-cluster names in this tutorial are `source` and `target`.
- Install compatible branching-capable KubeDB/database images and valid licenses on both database clusters. PostgreSQL version `16.1` must be available on the target as well as the source.
- Configure compatible `ceph-rbd` storage and snapshot classes on both clusters. Source snapshots must be resolvable on the target against the shared backend.

Using two independent Ceph backends does not satisfy the storage requirement. Confirm [CSI/backend compatibility](/docs/operatormanual/branching/snapshots.md#cross-cluster-backend-compatibility) before creating the source.

```bash
kubectl --context source create namespace demo
kubectl --context target create namespace dev
```

## Prepare and seed the source

Use the following HA source manifest on the source cluster:

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: prod-pg
  namespace: demo
spec:
  version: "16.1"
  replicas: 3
  standbyMode: Hot
  storageType: Durable
  storage:
    storageClassName: ceph-rbd
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

```bash
kubectl --context source apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/branch/source-postgres.yaml
kubectl --context source -n demo wait postgres/prod-pg \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
SOURCE_POD=$(kubectl --context source -n demo get pods \
  -l app.kubernetes.io/instance=prod-pg,kubedb.com/role=primary \
  -o jsonpath='{.items[0].metadata.name}')
curl -fsSL -o seed.sql https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/branch/seed.sql
kubectl --context source -n demo exec -i "$SOURCE_POD" -c postgres -- \
  psql -U postgres -d postgres -v ON_ERROR_STOP=1 < seed.sql
```

The `branch_demo` table now contains the two seeded rows.

## Create the cross-cluster Branch

Apply the Branch to the **source** cluster:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: pg-cross
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Postgres
      name: prod-pg
    namespace: demo
  target:
    clusterName: target
    namespace: dev
    name: pg-cross
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  deletionPolicy: Delete
```

`spec.target.clusterName` is the target's **OCM ManagedCluster name**, which can differ from a kubeconfig context name. Courier routes the request through the hub and provisions `dev/pg-cross` on that cluster.

```bash
kubectl --context source apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/branch/branch-cross-cluster.yaml
kubectl --context source -n demo wait branch/pg-cross \
  --for=jsonpath='{.status.phase}'=Ready --timeout=20m
kubectl --context source -n demo get branch pg-cross
kubectl --context target -n dev get postgres pg-cross
kubectl --context target -n dev get pvc,volumesnapshot
```

The original Branch reports `Ready` and mode `Initiator`. To inspect delivery and the target's execution:

```bash
kubectl --context hub -n source get branchwork
kubectl --context hub -n target get manifestwork
kubectl --context target -n dev get branch
```

Manage the branch through its original source resource. Courier owns the delivered target resources.

## Verify data and write isolation

Find the target primary and connect using the target Secret and Service:

```bash
BRANCH_POD=$(kubectl --context target -n dev get pods \
  -l app.kubernetes.io/instance=pg-cross,kubedb.com/role=primary \
  -o jsonpath='{.items[0].metadata.name}')
BRANCH_SECRET=$(kubectl --context target -n dev get postgres pg-cross -o jsonpath='{.spec.authSecret.name}')
BRANCH_USER=$(kubectl --context target -n dev get secret "$BRANCH_SECRET" -o jsonpath='{.data.username}' | base64 -d)
BRANCH_PASSWORD=$(kubectl --context target -n dev get secret "$BRANCH_SECRET" -o jsonpath='{.data.password}' | base64 -d)
kubectl --context target -n dev exec "$BRANCH_POD" -c postgres -- \
  env PGPASSWORD="$BRANCH_PASSWORD" psql -h pg-cross -U "$BRANCH_USER" -d postgres \
  -v ON_ERROR_STOP=1 -c 'TABLE branch_demo;'
```

The branch should contain both seeded rows. Write a row to each database:

```bash
kubectl --context target -n dev exec "$BRANCH_POD" -c postgres -- \
  env PGPASSWORD="$BRANCH_PASSWORD" psql -h pg-cross -U "$BRANCH_USER" -d postgres \
  -v ON_ERROR_STOP=1 -c "INSERT INTO branch_demo VALUES (100, 'branch-only@example.com');"
kubectl --context source -n demo exec "$SOURCE_POD" -c postgres -- \
  psql -U postgres -d postgres -v ON_ERROR_STOP=1 \
  -c "INSERT INTO branch_demo VALUES (200, 'source-only@example.com');"
kubectl --context source -n demo exec "$SOURCE_POD" -c postgres -- \
  psql -U postgres -d postgres -c 'SELECT id FROM branch_demo ORDER BY id;'
kubectl --context target -n dev exec "$BRANCH_POD" -c postgres -- \
  env PGPASSWORD="$BRANCH_PASSWORD" psql -h pg-cross -U "$BRANCH_USER" -d postgres \
  -c 'SELECT id FROM branch_demo ORDER BY id;'
unset BRANCH_PASSWORD
```

The source has IDs `1`, `2`, and `200`; the branch has `1`, `2`, and `100`. The branch is an independent database copy.

## Cleanup

Delete the original Branch on the source cluster while the hub, agents, and storage backend remain available:

```bash
kubectl --context source -n demo delete branch pg-cross --wait=true --timeout=15m
kubectl --context target -n dev get postgres,pvc,volumesnapshot
```

`deletionPolicy: Delete` removes the branch database and its managed storage on the target. Keep the source for other tutorials, or remove the remaining example resources:

```bash
kubectl --context source -n demo delete postgres prod-pg --wait=true --timeout=10m
kubectl --context source delete namespace demo
kubectl --context target delete namespace dev
```

The source uses `WipeOut`, so its deletion removes source data. For a retained branch, use [Orphan](/docs/guides/postgres/branch/customization/index.md#deletion-policies).
