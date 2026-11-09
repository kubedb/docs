---
title: Branch PostgreSQL into Another Namespace
menu:
  docs_{{ .version }}:
    identifier: pg-branch-same-cluster
    name: Same Cluster
    parent: guides-pg-branch
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Branch PostgreSQL into Another Namespace

This guide creates an HA PostgreSQL source in `demo`, then branches its data into `dev` within the same cluster. Both databases use Ceph RBD.

## Before You Begin

- Complete [Courier setup](/docs/operatormanual/branching/setup.md#single-cluster-installation).
- Have a working `ceph-rbd` StorageClass and compatible snapshot class. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).
- Ensure PostgreSQL version `16.1` is available in the installed KubeDB catalog and use branching-capable KubeDB/database images.
- Be familiar with the [Branch specification](/docs/guides/postgres/branch/overview/index.md).

The storage class name is an example; substitute your Ceph RBD class consistently in the source and Branch manifests.

```bash
kubectl create namespace demo
kubectl create namespace dev
```

## Prepare the source database

Create this PostgreSQL database:

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
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/branch/source-postgres.yaml
kubectl -n demo wait postgres/prod-pg \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
```

Find the elected primary rather than assuming a pod ordinal:

```bash
SOURCE_POD=$(kubectl -n demo get pods \
  -l app.kubernetes.io/instance=prod-pg,kubedb.com/role=primary \
  -o jsonpath='{.items[0].metadata.name}')
curl -fsSL -o seed.sql https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/branch/seed.sql
kubectl -n demo exec -i "$SOURCE_POD" -c postgres -- \
  psql -U postgres -d postgres -v ON_ERROR_STOP=1 < seed.sql
```

The seed creates `branch_demo` with two rows. Check it:

```bash
kubectl -n demo exec "$SOURCE_POD" -c postgres -- \
  psql -U postgres -d postgres -c 'TABLE branch_demo;'
```

## Create the Branch

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: pg-dev
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
    name: pg-dev
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  deletionPolicy: Delete
```

The Branch lives in `demo`, references `prod-pg`, and creates `pg-dev` in `dev`. Omitting `spec.target.clusterName` selects the same-cluster workflow. `storageClassName` selects the target Ceph class; `resources` sets target PostgreSQL container requests.

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/branch/branch-same-cluster.yaml
kubectl -n demo wait branch/pg-dev \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
kubectl -n demo get branch pg-dev
kubectl -n dev get postgres pg-dev
kubectl -n dev get pvc,volumesnapshot
```

The Branch should report `Ready` and mode `Local`. Courier imports source snapshot references into `dev` before creating target PVCs, because a PVC references snapshots within its own namespace.

## Verify the copied data

Read the target auth Secret and connect through its Service:

```bash
BRANCH_POD=$(kubectl -n dev get pods \
  -l app.kubernetes.io/instance=pg-dev,kubedb.com/role=primary \
  -o jsonpath='{.items[0].metadata.name}')
BRANCH_SECRET=$(kubectl -n dev get postgres pg-dev -o jsonpath='{.spec.authSecret.name}')
BRANCH_USER=$(kubectl -n dev get secret "$BRANCH_SECRET" -o jsonpath='{.data.username}' | base64 -d)
BRANCH_PASSWORD=$(kubectl -n dev get secret "$BRANCH_SECRET" -o jsonpath='{.data.password}' | base64 -d)
kubectl -n dev exec "$BRANCH_POD" -c postgres -- \
  env PGPASSWORD="$BRANCH_PASSWORD" psql -h pg-dev -U "$BRANCH_USER" -d postgres \
  -v ON_ERROR_STOP=1 -c 'TABLE branch_demo;'
```

The branch contains the two seeded rows. Verify independent writes:

```bash
kubectl -n dev exec "$BRANCH_POD" -c postgres -- \
  env PGPASSWORD="$BRANCH_PASSWORD" psql -h pg-dev -U "$BRANCH_USER" -d postgres \
  -v ON_ERROR_STOP=1 -c "INSERT INTO branch_demo VALUES (100, 'branch-only@example.com');"
kubectl -n demo exec "$SOURCE_POD" -c postgres -- \
  psql -U postgres -d postgres -v ON_ERROR_STOP=1 \
  -c "INSERT INTO branch_demo VALUES (200, 'source-only@example.com');"
kubectl -n demo exec "$SOURCE_POD" -c postgres -- \
  psql -U postgres -d postgres -c 'SELECT id FROM branch_demo ORDER BY id;'
kubectl -n dev exec "$BRANCH_POD" -c postgres -- \
  env PGPASSWORD="$BRANCH_PASSWORD" psql -h pg-dev -U "$BRANCH_USER" -d postgres \
  -c 'SELECT id FROM branch_demo ORDER BY id;'
```

The source has IDs `1`, `2`, and `200`; the branch has `1`, `2`, and `100`. Source changes reach the branch only through a [refresh](/docs/guides/postgres/branch/refresh.md).

```bash
unset BRANCH_PASSWORD
```

## Cleanup

Delete the Branch first and allow Courier to finish cleanup:

```bash
kubectl -n demo delete branch pg-dev --wait=true --timeout=10m
kubectl -n dev get postgres,pvc,volumesnapshot
```

With `deletionPolicy: Delete`, the target database and its managed storage are removed. The source remains. Keep it for the feature tutorials, or remove the tutorial resources when finished:

```bash
kubectl -n demo delete postgres prod-pg --wait=true --timeout=10m
kubectl delete namespace dev demo
```

The source manifest uses `WipeOut`; deleting it removes its data. See [deletion policies](/docs/guides/postgres/branch/customization/index.md#deletion-policies) to retain a branch independently.
