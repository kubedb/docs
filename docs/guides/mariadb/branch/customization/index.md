---
title: MariaDB Branch Configuration
menu:
  docs_{{ .version }}:
    identifier: mariadb-branch-configuration
    name: Configuration
    parent: guides-mariadb-branch
    weight: 70
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MariaDB Branch Configuration

This page covers MariaDB branch overrides, snapshot selection, TLS, and deletion policies. Start with the [Branch overview](/docs/guides/mariadb/branch/overview/index.md) and one of the creation tutorials.

The examples below reuse the three-replica Galera source `demo/mariadb-source`, Ceph RBD storage, and target namespace `dev` from the [same-cluster tutorial](/docs/guides/mariadb/branch/same-cluster/index.md). Keep that source available and seed its five customer rows before trying these examples; skip the tutorial's source cleanup until you finish here. The snapshot and Orphan examples use the source without a cert-manager TLS configuration; the final example requires `spec.tls` on the source. All examples show same-cluster configuration. For cross-cluster use, set `spec.target.clusterName` to the target ManagedCluster name, apply the Branch on the source, and inspect the database resources on the target.

## Storage and resource overrides

`spec.target.storageClassName` selects the target StorageClass. It must be compatible with the source snapshot's CSI driver and backend. The tutorials use `ceph-rbd` on both sides, including for same-cluster branching.

`spec.target.resources` sets MariaDB container CPU/memory requests and limits. It does not resize storage. Ensure that any resource limits you add are at least the corresponding requests.

## Select a snapshot class

Use an explicit `volumeSnapshotClassName` when your cluster has multiple classes. Save this Branch as `branch-snapshot-class.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mariadb-snapshot
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MariaDB
      name: mariadb-source
    namespace: demo
  target:
    namespace: dev
    name: mariadb-snapshot
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  volumeSnapshotClassName: ceph-rbd-snapclass
  deletionPolicy: Delete
```

Create `ceph-rbd-snapclass` with settings appropriate for your CSI installation before applying this example. For another namespace, the importing snapshots use that class too. For another cluster, the named class must exist on both clusters and use compatible driver/backend settings. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).

```shell
kubectl apply -f branch-snapshot-class.yaml
kubectl -n demo wait branch/mariadb-snapshot \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
kubectl -n demo get volumesnapshot
kubectl -n dev get volumesnapshot
```

## Deletion policies

The Branch's `spec.deletionPolicy` controls what happens when you delete the **Branch resource**. Keep the source cluster, target cluster, Courier agents, OCM hub when used, and storage backend available until cleanup completes.

| Policy | Behavior |
|---|---|
| `Delete` (default) | Delete the target database, cloned PVCs, managed auth/config Secrets, and action Jobs; clean up branch snapshot resources. |
| `Orphan` | Retain the target database, cloned PVCs, and required auth/config Secrets as independent resources; stop refresh and remove Courier's branch management and action Jobs. |

Both policies retain the source database and the target namespace. Cleanup uses finalizers and can take time while CSI resources and backend clone dependencies are released. Let the controller complete this ordering before uninstalling Courier.

### Retain a branch with Orphan

Save this Branch as `branch-orphan.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mariadb-kept
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MariaDB
      name: mariadb-source
    namespace: demo
  target:
    namespace: dev
    name: mariadb-kept
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  deletionPolicy: Orphan
```

```shell
kubectl apply -f branch-orphan.yaml
kubectl -n demo wait branch/mariadb-kept \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
kubectl -n demo delete branch mariadb-kept --wait=true --timeout=15m
kubectl -n dev get mariadb mariadb-kept
kubectl -n dev get pvc,secrets
```

The database and its data remain usable after the Branch is gone. Read the retained database's auth Secret name and password from your terminal:

```shell
kubectl -n dev get mariadb mariadb-kept -o jsonpath='{.spec.authSecret.name}{"\n"}'
kubectl -n dev get secret mariadb-kept-auth -o jsonpath='{.data.password}' | base64 -d
```

Use the returned Secret name if it differs from `mariadb-kept-auth`. Open the retained pod shell:

```shell
kubectl -n dev exec -it mariadb-kept-0 -c mariadb -- bash
```

Inside the pod, connect through the retained Service and enter its password:

```shell
mariadb --protocol=TCP -h mariadb-kept.dev.svc -P 3306 -u root -p branchdb
```

At the MariaDB prompt, verify the retained data:

```sql
SELECT * FROM customers ORDER BY id;
```

The five customer records should match the source snapshot. Their values reflect any source edits you made before creating `mariadb-kept`. Leave the client with `\q`, then run `exit` to return to your terminal. Retain the auth Secret; its credential is needed to access the preserved data.

To remove the retained database and its durable data later, choose the database's own deletion policy explicitly:

```shell
kubectl -n dev patch mariadb mariadb-kept --type=merge \
  -p '{"spec":{"deletionPolicy":"WipeOut"}}'
kubectl -n dev delete mariadb mariadb-kept --wait=true --timeout=10m
```

This final step removes the retained database's data.

## Sources with cert-manager TLS

Before applying this example, configure TLS on source `mariadb-source` following the [MariaDB TLS guide](/docs/guides/mariadb/tls/configure/index.md). Prepare cert-manager on the target and an issuer named `branch-ca` in `dev`, with its required signing Secret available there.

When the source has `spec.tls` configured, supply `spec.target.issuerRef` so KubeDB can issue certificates for the branch's own Service names. Save this Branch as `branch-tls.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mariadb-tls
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MariaDB
      name: mariadb-source
    namespace: demo
  target:
    namespace: dev
    name: mariadb-tls
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

```shell
kubectl -n dev get issuer branch-ca
kubectl apply -f branch-tls.yaml
kubectl -n demo wait branch/mariadb-tls \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
kubectl -n dev get certificates,secrets
kubectl -n dev get mariadb mariadb-tls -o yaml
```

Check that the branch's certificates are ready. For client verification, follow the [MariaDB TLS guide](/docs/guides/mariadb/tls/configure/index.md) using the target issuer's CA certificate with `--ssl-ca` and `--ssl-verify-server-cert`, and connect through the branch Service's DNS name. Supply client certificate and key options when required by the server's authentication policy. See the [MariaDB client reference](https://mariadb.com/docs/server/clients-and-utilities/mariadb-client/mariadb-command-line-client) for these TLS options.

## Cleanup optional examples

Delete any Branches created by the snapshot and TLS examples before deleting their namespaces:

```shell
kubectl -n demo delete branch mariadb-snapshot mariadb-tls --ignore-not-found --wait=true --timeout=15m
```

For cross-cluster variants, run Branch cleanup on the source cluster and inspect remaining resources on the target.
