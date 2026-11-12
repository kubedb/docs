---
title: Neo4j Branch Configuration
menu:
  docs_{{ .version }}:
    identifier: neo4j-branch-configuration
    name: Configuration
    parent: guides-neo4j-branch
    weight: 70
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Neo4j Branch Configuration

These optional examples reuse `neo4j-source` from the [same-cluster tutorial](/docs/guides/neo4j/branch/same-cluster/index.md). Keep that source and its namespaces before cleanup. Run one optional example at a time, deleting its Branch before creating the next to limit resource usage.

## Storage and resource overrides

`target.storageClassName` selects the clone’s StorageClass; it must resolve the source snapshot on the same backend. `target.resources` sets target container CPU and memory requests and limits. It does not resize storage. The examples set both explicitly.

## Select a snapshot class

Create a compatible `ceph-rbd-snapclass` before applying this example:

Save the following as `neo4j-snapshot.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: neo4j-snapshot
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Neo4j
      name: neo4j-source
    namespace: demo
  target:
    namespace: dev
    name: neo4j-snapshot
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 2Gi
      limits:
        memory: 2Gi
  volumeSnapshotClassName: ceph-rbd-snapclass
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f neo4j-snapshot.yaml
kubectl -n demo wait branch/neo4j-snapshot --for=condition=Ready --timeout=60m
kubectl -n demo get branch neo4j-snapshot
```

```shell
kubectl -n demo get volumesnapshots
kubectl -n dev get volumesnapshots
```

The compatible VolumeGroupSnapshotClass is discovered separately. Individual per-volume fallback does not provide an atomic snapshot across volumes.

For cross-namespace or cross-cluster imports, the named VolumeSnapshotClass must exist wherever it is used. Cross-cluster storage must resolve handles against the same Ceph backend. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).

```shell
kubectl -n demo delete branch neo4j-snapshot --wait=true --timeout=30m
```

## Deletion policies

| Policy | When the Branch is deleted |
|---|---|
| `Delete` (default) | Remove the managed target database, cloned PVCs, action Jobs, and branch snapshot resources. |
| `Orphan` | Stop refresh and branch management; retain the target database and required storage and credentials, while cleaning up branch snapshot resources. |

Both policies preserve the source and namespaces. Keep the controllers and storage available until finalizer cleanup finishes.

### Retain a branch with Orphan

Save the following as `neo4j-kept.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: neo4j-kept
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Neo4j
      name: neo4j-source
    namespace: demo
  target:
    namespace: dev
    name: neo4j-kept
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 2Gi
      limits:
        memory: 2Gi
  deletionPolicy: Orphan
```

```shell
kubectl -n demo apply -f neo4j-kept.yaml
kubectl -n demo wait branch/neo4j-kept --for=condition=Ready --timeout=60m
kubectl -n demo get branch neo4j-kept
```

```shell
kubectl -n demo delete branch neo4j-kept --wait=true --timeout=30m
kubectl -n dev get neo4j neo4j-kept
kubectl -n dev get pvc,secrets -l app.kubernetes.io/instance=neo4j-kept
```

Verify that the retained database still serves data:

Discover the auth Secret referenced by `neo4j-kept`:

```shell
kubectl -n dev get neo4j neo4j-kept -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
neo4j-kept-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n dev get secret neo4j-kept-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it neo4j-kept-0 -c neo4j -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
cypher-shell -a neo4j://neo4j-kept.dev.svc:7687 -u neo4j -d neo4j --format plain
```

Enter Cypher at the client prompt. `neo4j://` discovers the cluster’s writer; a direct `bolt://` connection to an arbitrary member can reject writes.

```cypher
MATCH (c:Customer) RETURN c.id AS id, c.full_name AS full_name, c.email AS email ORDER BY id;
```

```text
id, full_name, email
1, "Updated Customer 1", "customer1@example.com"
2, "Customer 2", "customer2@example.com"
3, "Customer 3", "customer3@example.com"
4, "Customer 4", "customer4@example.com"
5, "Customer 5", "customer5@example.com"
```

Run `:exit` to leave `cypher-shell`, then `exit` to leave the pod shell.

The expected values assume you completed the source update in the same-cluster tutorial. Preserve the auth Secret; the retained database still needs its credential. To explicitly remove the retained database and its durable data:

```shell
kubectl -n dev patch neo4j neo4j-kept --type=merge -p '{"spec":{"deletionPolicy":"WipeOut"}}'
kubectl -n dev delete neo4j neo4j-kept --wait=true --timeout=30m
```

## TLS-enabled sources

Configure a TLS-enabled source following the [Neo4j TLS guide](/docs/guides/neo4j/tls/configure/index.md) before this example. Substitute its name in `source.databaseRef.name` if it differs from `neo4j-source`. Prepare cert-manager, an Issuer named `branch-ca`, and its signing Secret in `dev` on the target cluster. Courier does not install that issuer.

Save the following as `neo4j-tls.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: neo4j-tls
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Neo4j
      name: neo4j-source
    namespace: demo
  target:
    namespace: dev
    name: neo4j-tls
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 2Gi
      limits:
        memory: 2Gi
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: branch-ca
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f neo4j-tls.yaml
kubectl -n demo wait branch/neo4j-tls --for=condition=Ready --timeout=60m
kubectl -n demo get branch neo4j-tls
```

```shell
kubectl -n dev get certificates
kubectl -n dev get neo4j neo4j-tls -o yaml
```

Confirm that target certificates are ready and use branch-specific names. Use the TLS guide’s routing URI and CA trust configuration for the target. The protocol and trust settings must match the source’s Bolt TLS policy. The plaintext commands and action connection settings from the main walkthroughs must be adapted to that TLS policy.

Use `kind: ClusterIssuer` when referencing a cluster-scoped issuer. A namespaced Issuer and its signing Secret must exist in the target namespace, including on another cluster.

## Cleanup optional examples

```shell
kubectl -n demo delete branch neo4j-tls --ignore-not-found --wait=true --timeout=30m
```

Return to the [same-cluster cleanup](/docs/guides/neo4j/branch/same-cluster/index.md#cleanup) after all dependent Branches are gone.
