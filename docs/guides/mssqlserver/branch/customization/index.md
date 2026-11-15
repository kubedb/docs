---
title: MSSQL Server Branch Configuration
menu:
  docs_{{ .version }}:
    identifier: mssqlserver-branch-configuration
    name: Configuration
    parent: guides-mssqlserver-branch
    weight: 70
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MSSQL Server Branch Configuration

These optional examples reuse `mssql-source` from the [same-cluster tutorial](/docs/guides/mssqlserver/branch/same-cluster/index.md). Keep the source and namespaces before that tutorial’s cleanup. Each Branch below has a distinct target name.

## Storage and resource overrides

`target.storageClassName` selects the clone’s StorageClass; it must resolve the source snapshot on the same backend. `target.resources` configures the target database containers independently of source resource requests. The manifests below set both explicitly. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md) for driver and backend requirements.

## Select a snapshot class

Create a compatible `ceph-rbd-snapclass` before this example. `volumeSnapshotClassName` selects the per-volume/import snapshot class; the compatible VolumeGroupSnapshotClass is discovered separately. Individual fallback snapshots do not provide an atomic snapshot across volumes.

Save the following as `mssql-snapshot.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mssql-snapshot
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MSSQLServer
      name: mssql-source
    namespace: demo
  target:
    namespace: dev
    name: mssql-snapshot
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: '1'
        memory: 3Gi
      limits:
        memory: 3Gi
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mssql-source-ca
  volumeSnapshotClassName: ceph-rbd-snapclass
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f mssql-snapshot.yaml
kubectl -n demo wait branch/mssql-snapshot --for=condition=Ready --timeout=60m
kubectl -n demo get branch mssql-snapshot
```

```shell
kubectl -n demo get volumesnapshots,volumegroupsnapshots
kubectl -n dev get volumesnapshots
```

For cross-namespace or cross-cluster imports, the named VolumeSnapshotClass must exist wherever it is used. Both clusters must resolve the snapshot handle against the same Ceph backend.

## Deletion policies

| Policy | When the Branch is deleted |
|---|---|
| `Delete` (default) | Remove the managed target database, cloned storage, action Jobs, and branch snapshot resources. |
| `Orphan` | Stop branch management and refresh, retain the target database and required storage/credentials, and clean up branch snapshot resources. |

The source database and namespaces remain under either policy. Keep the controllers and storage available until finalizer cleanup completes.

### Retain a branch with Orphan

Save the following as `mssql-kept.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mssql-kept
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MSSQLServer
      name: mssql-source
    namespace: demo
  target:
    namespace: dev
    name: mssql-kept
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: '1'
        memory: 3Gi
      limits:
        memory: 3Gi
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mssql-source-ca
  deletionPolicy: Orphan
```

```shell
kubectl -n demo apply -f mssql-kept.yaml
kubectl -n demo wait branch/mssql-kept --for=condition=Ready --timeout=60m
kubectl -n demo get branch mssql-kept
```

```shell
kubectl -n demo delete branch mssql-kept --wait=true --timeout=30m
kubectl -n dev get mssqlserver mssql-kept
kubectl -n dev get pvc,secrets -l app.kubernetes.io/instance=mssql-kept
```

Verify that the retained database still serves its copied data:

Discover the auth Secret referenced by `mssql-kept`:

```shell
kubectl -n dev get mssqlserver mssql-kept -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mssql-kept-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n dev get secret mssql-kept-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n dev get secret mssql-kept-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `sa`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it mssql-kept-0 -c mssql -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-kept.dev.svc,1433 -U sa -No
```

Enter the password decoded above. `-No` makes client encryption optional for this `clientTLS: false` example. Run SQL batches at the `sqlcmd` prompt; `GO` submits each batch.

```sql
USE agdb1;
GO
SELECT * FROM dbo.customers ORDER BY id;
GO
```

```text
id full_name          email
-- ------------------ ---------------------
1  Updated Customer 1 customer1@example.com
2  Customer 2         customer2@example.com
3  Customer 3         customer3@example.com
4  Customer 4         customer4@example.com
5  Customer 5         customer5@example.com

(5 rows affected)
```

Run `QUIT` at the `sqlcmd` prompt, then `exit` in the pod shell to return to your terminal.

The results above assume you completed the source edit in the same-cluster tutorial. Preserve the retained auth Secret; the database still needs its credential. To remove the retained deployment and its data explicitly:

```shell
kubectl -n dev patch mssqlserver mssql-kept --type=merge -p '{"spec":{"deletionPolicy":"WipeOut"}}'
kubectl -n dev delete mssqlserver mssql-kept --wait=true --timeout=30m
```

## TLS-enabled sources

Configure a TLS-enabled source following the [MSSQL Server TLS guide](/docs/guides/mssqlserver/tls/ag_cluster.md) before creating this optional Branch. Use that TLS-enabled source’s name in `source.databaseRef.name` below if it differs from `mssql-source`. Prepare cert-manager and a target Issuer `mssql-source-ca` in `dev`, including its signing Secret. Courier delivers database resources; it does not create this issuer.

The earlier issuer setup is already sufficient for certificate issuance; client TLS is controlled by the source’s `tls.clientTLS` setting and inherited by the branch.

Save the following as `mssql-tls.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mssql-tls
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MSSQLServer
      name: mssql-source
    namespace: demo
  target:
    namespace: dev
    name: mssql-tls
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: '1'
        memory: 3Gi
      limits:
        memory: 3Gi
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mssql-source-ca
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f mssql-tls.yaml
kubectl -n demo wait branch/mssql-tls --for=condition=Ready --timeout=60m
kubectl -n demo get branch mssql-tls
```

```shell
kubectl -n dev get certificates
kubectl -n dev get mssqlserver mssql-tls -o yaml
```

Confirm that target certificates are ready and use branch-specific names. For `clientTLS: true`, use the TLS guide’s encrypted `sqlcmd` connection with the issuing CA available for verification; adapt the plaintext example commands accordingly.

Use `kind: ClusterIssuer` for a cluster-scoped issuer. For OCM, the issuer must exist on the target cluster, and a namespaced Issuer and its supporting Secret must be in the target namespace.

## Cleanup optional examples

```shell
kubectl -n demo delete branch mssql-snapshot mssql-tls --ignore-not-found --wait=true --timeout=30m
```

Return to the [same-cluster cleanup](/docs/guides/mssqlserver/branch/same-cluster/index.md#cleanup) when all dependent Branches are gone.
