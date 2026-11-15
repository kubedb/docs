---
title: MongoDB Branch Configuration
menu:
  docs_{{ .version }}:
    identifier: mongodb-branch-configuration
    name: Configuration
    parent: guides-mongodb-branch
    weight: 70
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MongoDB Branch Configuration

These optional examples reuse `mongodb-source` from the [same-cluster tutorial](/docs/guides/mongodb/branch/same-cluster/index.md). Keep the source and namespaces before that tutorial’s cleanup. Each Branch below has a distinct target name.

## Storage and resource overrides

`target.storageClassName` selects the clone’s StorageClass; it must resolve the source snapshot on the same backend. `target.resources` configures the target database containers independently of source resource requests. The manifests below set both explicitly. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md) for driver and backend requirements.

## Select a snapshot class

Create a compatible `ceph-rbd-snapclass` before this example. `volumeSnapshotClassName` selects the per-volume/import snapshot class; the compatible VolumeGroupSnapshotClass is discovered separately. Individual fallback snapshots do not provide an atomic snapshot across volumes.

Save the following as `mongodb-snapshot.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mongodb-snapshot
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MongoDB
      name: mongodb-source
    namespace: demo
  target:
    namespace: dev
    name: mongodb-snapshot
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
      limits:
        memory: 1Gi
  volumeSnapshotClassName: ceph-rbd-snapclass
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f mongodb-snapshot.yaml
kubectl -n demo wait branch/mongodb-snapshot --for=condition=Ready --timeout=60m
kubectl -n demo get branch mongodb-snapshot
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

Save the following as `mongodb-kept.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mongodb-kept
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MongoDB
      name: mongodb-source
    namespace: demo
  target:
    namespace: dev
    name: mongodb-kept
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
      limits:
        memory: 1Gi
  deletionPolicy: Orphan
```

```shell
kubectl -n demo apply -f mongodb-kept.yaml
kubectl -n demo wait branch/mongodb-kept --for=condition=Ready --timeout=60m
kubectl -n demo get branch mongodb-kept
```

```shell
kubectl -n demo delete branch mongodb-kept --wait=true --timeout=30m
kubectl -n dev get mongodb mongodb-kept
kubectl -n dev get pvc,secrets -l app.kubernetes.io/instance=mongodb-kept
```

Verify that the retained database still serves its copied data:

Discover the auth Secret referenced by `mongodb-kept`:

```shell
kubectl -n dev get mongodb mongodb-kept -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-kept-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n dev get secret mongodb-kept-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n dev get secret mongodb-kept-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it mongodb-kept-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-kept.dev.svc --port 27017 --username root --authenticationDatabase admin --password
```

Enter the password decoded above. Run JavaScript at the `mongosh` prompt.

```javascript
use branchdb
db.customers.find().sort({_id: 1}).toArray()
```

```text
[
  {"_id": 1, "full_name": "Updated Customer 1", "email": "customer1@example.com"},
  {"_id": 2, "full_name": "Customer 2", "email": "customer2@example.com"},
  {"_id": 3, "full_name": "Customer 3", "email": "customer3@example.com"},
  {"_id": 4, "full_name": "Customer 4", "email": "customer4@example.com"},
  {"_id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}
]
```

Run `exit` in `mongosh`, then `exit` in the pod shell to return to your terminal.

The results above assume you completed the source edit in the same-cluster tutorial. Preserve the retained auth Secret; the database still needs its credential. To remove the retained deployment and its data explicitly:

```shell
kubectl -n dev patch mongodb mongodb-kept --type=merge -p '{"spec":{"deletionPolicy":"WipeOut"}}'
kubectl -n dev delete mongodb mongodb-kept --wait=true --timeout=30m
```

## TLS-enabled sources

Configure a TLS-enabled source following the [MongoDB TLS guide](/docs/guides/mongodb/tls/replicaset.md) before creating this optional Branch. Use that TLS-enabled source’s name in `source.databaseRef.name` below if it differs from `mongodb-source`. Prepare cert-manager and a target Issuer `branch-ca` in `dev`, including its signing Secret. Courier delivers database resources; it does not create this issuer.

Save the following as `mongodb-tls.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mongodb-tls
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MongoDB
      name: mongodb-source
    namespace: demo
  target:
    namespace: dev
    name: mongodb-tls
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
      limits:
        memory: 1Gi
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: branch-ca
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f mongodb-tls.yaml
kubectl -n demo wait branch/mongodb-tls --for=condition=Ready --timeout=60m
kubectl -n demo get branch mongodb-tls
```

```shell
kubectl -n dev get certificates
kubectl -n dev get mongodb mongodb-tls -o yaml
```

Confirm that target certificates are ready and use branch-specific names. For a `requireSSL` source, use `mongosh --tls` with the target CA and client certificate as required by the source authentication mode. The TLS guide covers password and X.509 authentication.

Use `kind: ClusterIssuer` for a cluster-scoped issuer. For OCM, the issuer must exist on the target cluster, and a namespaced Issuer and its supporting Secret must be in the target namespace.

## Cleanup optional examples

```shell
kubectl -n demo delete branch mongodb-snapshot mongodb-tls --ignore-not-found --wait=true --timeout=30m
```

Return to the [same-cluster cleanup](/docs/guides/mongodb/branch/same-cluster/index.md#cleanup) when all dependent Branches are gone.
