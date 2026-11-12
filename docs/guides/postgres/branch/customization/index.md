---
title: PostgreSQL Branch Configuration
menu:
  docs_{{ .version }}:
    identifier: pg-branch-configuration
    name: Configuration
    parent: guides-pg-branch
    weight: 70
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PostgreSQL Branch Configuration

This page covers PostgreSQL branch overrides, snapshot selection, TLS, and deletion policies. Start with the [Branch overview](/docs/guides/postgres/branch/overview/index.md) and one of the creation tutorials.

The examples below use the same HA source `demo/prod-pg`, Ceph RBD storage, and target namespace `dev`. They show same-cluster configuration. For cross-cluster use, set `spec.target.clusterName` to the target ManagedCluster name, apply the Branch on the source, and inspect the database resources on the target.

## Storage and resource overrides

`spec.target.storageClassName` selects the target StorageClass. It must be compatible with the source snapshot's CSI driver and backend. The tutorials use `ceph-rbd` on both sides, including for same-cluster branching.

`spec.target.resources` sets PostgreSQL container CPU/memory requests and limits. It does not resize storage. Ensure that any resource limits you add are at least the corresponding requests.

## Select a snapshot class

Use an explicit `volumeSnapshotClassName` when your cluster has multiple classes:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: pg-snapshot
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
    name: pg-snapshot
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  volumeSnapshotClassName: ceph-rbd-snapclass
  deletionPolicy: Delete
```

Create `ceph-rbd-snapclass` with settings appropriate for your CSI installation before applying this example. For another namespace, the importing snapshots use that class too. For another cluster, the named class must exist on both clusters and use compatible driver/backend settings. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/branch/branch-snapshot-class.yaml
kubectl -n demo wait branch/pg-snapshot \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
kubectl -n demo get volumesnapshot
kubectl -n dev get volumesnapshot
```

## TLS-enabled sources

Before applying this example, configure TLS on source `prod-pg` following the [PostgreSQL TLS guide](/docs/guides/postgres/tls/configure/index.md). Prepare cert-manager on the target and an issuer named `branch-ca` in `dev`, with its required signing Secret available there.

When the source has TLS enabled, supply `spec.target.issuerRef` so KubeDB can issue certificates for the branch's own Service names:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: pg-tls
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
    name: pg-tls
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: branch-ca
  deletionPolicy: Delete
```

Use `kind: ClusterIssuer` when referencing a cluster-scoped issuer. The issuer must exist on the target cluster. A namespaced `Issuer` and its supporting Secrets must be in the target namespace.

```bash
kubectl -n dev get issuer branch-ca
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/branch/branch-tls.yaml
kubectl -n demo wait branch/pg-tls \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
kubectl -n dev get certificates,secrets
kubectl -n dev get postgres pg-tls -o yaml
```

Check that the branch's certificates are ready and use the issuer's CA for client verification. Configure client TLS parameters as described in the PostgreSQL TLS guide; the plaintext connection commands from the creation tutorials need to be adapted for the source's TLS and authentication policy.

## Deletion policies

The Branch's `spec.deletionPolicy` controls what happens when you delete the **Branch resource**. Keep the source cluster, target cluster, Courier agents, OCM hub when used, and storage backend available until cleanup completes.

| Policy | Behavior |
|---|---|
| `Delete` (default) | Delete the target database, cloned PVCs, managed auth/config Secrets, and action Jobs; clean up branch snapshot resources. |
| `Orphan` | Retain the target database, cloned PVCs, and required auth/config Secrets as independent resources; stop refresh and remove Courier's branch management and action Jobs. |

Both policies retain the source database and the target namespace. Cleanup uses finalizers and can take time while CSI resources and backend clone dependencies are released. Let the controller complete this ordering before uninstalling Courier.

### Retain a branch with Orphan

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: pg-kept
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
    name: pg-kept
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  deletionPolicy: Orphan
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/branch/branch-orphan.yaml
kubectl -n demo wait branch/pg-kept \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
kubectl -n demo delete branch pg-kept --wait=true --timeout=15m
kubectl -n dev get postgres pg-kept
kubectl -n dev get pvc,secrets
```

The database and its data remain usable after the Branch is gone. Read its auth Secret and verify a query through Service `pg-kept`, following the Service connection steps in the [same-cluster tutorial](/docs/guides/postgres/branch/same-cluster/index.md). Retain the original Secret; it contains the credential for the preserved data.

To remove the retained database and its durable data later, choose the database's own deletion policy explicitly:

```bash
kubectl -n dev patch postgres pg-kept --type=merge \
  -p '{"spec":{"deletionPolicy":"WipeOut"}}'
kubectl -n dev delete postgres pg-kept --wait=true --timeout=10m
```

This final step removes the retained database's data.

## Cleanup optional examples

Delete any Branches created by the snapshot and TLS examples before deleting their namespaces:

```bash
kubectl -n demo delete branch pg-snapshot pg-tls --ignore-not-found --wait=true --timeout=15m
```

For cross-cluster variants, run Branch cleanup on the source cluster and inspect remaining resources on the target.
