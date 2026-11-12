---
title: Redis Branch Configuration
menu:
  docs_{{ .version }}:
    identifier: redis-branch-configuration
    name: Configuration
    parent: guides-redis-branch
    weight: 70
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Redis Branch Configuration

These optional examples reuse `redis-source` from the [same-cluster tutorial](/docs/guides/redis/branch/same-cluster/index.md). Keep that source and its namespaces before cleanup. Run one optional example at a time, deleting its Branch before creating the next to limit resource usage.

## Storage and resource overrides

`target.storageClassName` selects the clone’s StorageClass; it must resolve the source snapshot on the same backend. `target.resources` sets target container CPU and memory requests and limits. It does not resize storage. The examples set both explicitly.

## Select a snapshot class

Create a compatible `ceph-rbd-snapclass` before applying this example:

Save the following as `redis-snapshot.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: redis-snapshot
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Redis
      name: redis-source
    namespace: demo
  target:
    namespace: dev
    name: redis-snapshot
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
kubectl -n demo apply -f redis-snapshot.yaml
kubectl -n demo wait branch/redis-snapshot --for=condition=Ready --timeout=60m
kubectl -n demo get branch redis-snapshot
```

```shell
kubectl -n demo get volumesnapshots
kubectl -n dev get volumesnapshots
```

For cross-namespace or cross-cluster imports, the named VolumeSnapshotClass must exist wherever it is used. Cross-cluster storage must resolve handles against the same Ceph backend. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).

```shell
kubectl -n demo delete branch redis-snapshot --wait=true --timeout=30m
```

## Deletion policies

| Policy | When the Branch is deleted |
|---|---|
| `Delete` (default) | Remove the managed target database, cloned PVCs, action Jobs, and branch snapshot resources. |
| `Orphan` | Stop refresh and branch management; retain the target database and required storage and credentials, while cleaning up branch snapshot resources. |

Both policies preserve the source and namespaces. Keep the controllers and storage available until finalizer cleanup finishes.

### Retain a branch with Orphan

Save the following as `redis-kept.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: redis-kept
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Redis
      name: redis-source
    namespace: demo
  target:
    namespace: dev
    name: redis-kept
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
kubectl -n demo apply -f redis-kept.yaml
kubectl -n demo wait branch/redis-kept --for=condition=Ready --timeout=60m
kubectl -n demo get branch redis-kept
```

```shell
kubectl -n demo delete branch redis-kept --wait=true --timeout=30m
kubectl -n dev get redis redis-kept
kubectl -n dev get pvc,secrets -l app.kubernetes.io/instance=redis-kept
```

Verify that the retained database still serves data:

Discover the auth Secret referenced by `redis-kept`:

```shell
kubectl -n dev get redis redis-kept -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
redis-kept-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n dev get secret redis-kept-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it redis-kept-0 -c redis -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
redis-cli -h redis-kept.dev.svc -p 6379 --askpass
```

Enter Redis commands at its prompt. Connections through the source Service reach its writable primary even when the pod you opened is a replica.

```redis
HMGET customer:1 full_name email
HMGET customer:2 full_name email
HMGET customer:3 full_name email
HMGET customer:4 full_name email
HMGET customer:5 full_name email
```

```text
1) "Updated Customer 1"
2) "customer1@example.com"
1) "Customer 2"
2) "customer2@example.com"
1) "Customer 3"
2) "customer3@example.com"
1) "Customer 4"
2) "customer4@example.com"
1) "Customer 5"
2) "customer5@example.com"
```

Run `QUIT` to leave `redis-cli`, then `exit` to leave the pod shell.

The expected values assume you completed the source update in the same-cluster tutorial. Preserve the auth Secret; the retained database still needs its credential. To explicitly remove the retained database and its durable data:

```shell
kubectl -n dev patch redis redis-kept --type=merge -p '{"spec":{"deletionPolicy":"WipeOut"}}'
kubectl -n dev delete redis redis-kept --wait=true --timeout=30m
```

## TLS-enabled sources

Configure a TLS-enabled source following the [Redis TLS guide](/docs/guides/redis/tls/sentinel.md) before this example. Substitute its name in `source.databaseRef.name` if it differs from `redis-source`. Prepare cert-manager, an Issuer named `branch-ca`, and its signing Secret in `dev` on the target cluster. Courier does not install that issuer.

Save the following as `redis-tls.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: redis-tls
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Redis
      name: redis-source
    namespace: demo
  target:
    namespace: dev
    name: redis-tls
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
kubectl -n demo apply -f redis-tls.yaml
kubectl -n demo wait branch/redis-tls --for=condition=Ready --timeout=60m
kubectl -n demo get branch redis-tls
```

```shell
kubectl -n dev get certificates
kubectl -n dev get redis redis-tls -o yaml
```

Confirm that target certificates are ready and use branch-specific names. Use `redis-cli --tls` with the target CA, client certificate, and key required by the source TLS policy. The TLS guide documents their mounted paths. The plaintext commands and action connection settings from the main walkthroughs must be adapted to that TLS policy.

Use `kind: ClusterIssuer` when referencing a cluster-scoped issuer. A namespaced Issuer and its signing Secret must exist in the target namespace, including on another cluster.

## Cleanup optional examples

```shell
kubectl -n demo delete branch redis-tls --ignore-not-found --wait=true --timeout=30m
```

Return to the [same-cluster cleanup](/docs/guides/redis/branch/same-cluster/index.md#cleanup) after all dependent Branches are gone.
