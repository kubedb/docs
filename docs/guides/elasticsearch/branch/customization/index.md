---
title: Elasticsearch Branch Configuration
menu:
  docs_{{ .version }}:
    identifier: elasticsearch-branch-configuration
    name: Configuration
    parent: guides-elasticsearch-branch
    weight: 70
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Elasticsearch Branch Configuration

These optional examples reuse `elasticsearch-source` from the [same-cluster tutorial](/docs/guides/elasticsearch/branch/same-cluster/index.md). Keep that source and its namespaces before cleanup. Run one optional example at a time, deleting its Branch before creating the next to limit resource usage.

## Storage and resource overrides

`target.storageClassName` selects the clone’s StorageClass; it must resolve the source snapshot on the same backend. `target.resources` sets target container CPU and memory requests and limits. It does not resize storage. The examples set both explicitly.

## Select a snapshot class

Create a compatible `ceph-rbd-snapclass` before applying this example:

Save the following as `elasticsearch-snapshot.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: elasticsearch-snapshot
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Elasticsearch
      name: elasticsearch-source
    namespace: demo
  target:
    namespace: dev
    name: elasticsearch-snapshot
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
kubectl -n demo apply -f elasticsearch-snapshot.yaml
kubectl -n demo wait branch/elasticsearch-snapshot --for=condition=Ready --timeout=60m
kubectl -n demo get branch elasticsearch-snapshot
```

```shell
kubectl -n demo get volumesnapshots
kubectl -n dev get volumesnapshots
```

The compatible VolumeGroupSnapshotClass is discovered separately. Individual per-volume fallback does not provide an atomic snapshot across volumes.

For cross-namespace or cross-cluster imports, the named VolumeSnapshotClass must exist wherever it is used. Cross-cluster storage must resolve handles against the same Ceph backend. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).

```shell
kubectl -n demo delete branch elasticsearch-snapshot --wait=true --timeout=30m
```

## Deletion policies

| Policy | When the Branch is deleted |
|---|---|
| `Delete` (default) | Remove the managed target database, cloned PVCs, action Jobs, and branch snapshot resources. |
| `Orphan` | Stop refresh and branch management; retain the target database and required storage and credentials, while cleaning up branch snapshot resources. |

Both policies preserve the source and namespaces. Keep the controllers and storage available until finalizer cleanup finishes.

### Retain a branch with Orphan

Save the following as `elasticsearch-kept.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: elasticsearch-kept
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Elasticsearch
      name: elasticsearch-source
    namespace: demo
  target:
    namespace: dev
    name: elasticsearch-kept
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
kubectl -n demo apply -f elasticsearch-kept.yaml
kubectl -n demo wait branch/elasticsearch-kept --for=condition=Ready --timeout=60m
kubectl -n demo get branch elasticsearch-kept
```

```shell
kubectl -n demo delete branch elasticsearch-kept --wait=true --timeout=30m
kubectl -n dev get elasticsearch elasticsearch-kept
kubectl -n dev get pvc,secrets -l app.kubernetes.io/instance=elasticsearch-kept
```

Verify that the retained database still serves data:

Discover the auth Secret referenced by `elasticsearch-kept`:

```shell
kubectl -n dev get elasticsearch elasticsearch-kept -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
elasticsearch-kept-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n dev get secret elasticsearch-kept-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it elasticsearch-kept-0 -c elasticsearch -- bash
```

Run the following `curl` commands inside the pod shell. `--user elastic` prompts for the decoded password on each request. Requests use the database Service.

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-kept.dev.svc:9200/customers/_count?filter_path=count&pretty"
```

```json
{
  "count": 5
}
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-kept.dev.svc:9200/customers/_search?size=5&sort=id:asc&filter_path=hits.hits._source&pretty"
```

```json
{
  "hits": {
    "hits": [
      {"_source": {"id": 1, "full_name": "Updated Customer 1", "email": "customer1@example.com"}},
      {"_source": {"id": 2, "full_name": "Customer 2", "email": "customer2@example.com"}},
      {"_source": {"id": 3, "full_name": "Customer 3", "email": "customer3@example.com"}},
      {"_source": {"id": 4, "full_name": "Customer 4", "email": "customer4@example.com"}},
      {"_source": {"id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}}
    ]
  }
}
```

Run `exit` to return to your terminal.

The expected values assume you completed the source update in the same-cluster tutorial. Preserve the auth Secret; the retained database still needs its credential. To explicitly remove the retained database and its durable data:

```shell
kubectl -n dev patch elasticsearch elasticsearch-kept --type=merge -p '{"spec":{"deletionPolicy":"WipeOut"}}'
kubectl -n dev delete elasticsearch elasticsearch-kept --wait=true --timeout=30m
```

## TLS-enabled sources

Configure a TLS-enabled source following the [Elasticsearch TLS guide](/docs/guides/elasticsearch/tls/elasticsearch-combined.md) before this example. Substitute its name in `source.databaseRef.name` if it differs from `elasticsearch-source`. Prepare cert-manager, an Issuer named `branch-ca`, and its signing Secret in `dev` on the target cluster. Courier does not install that issuer.

Set `target.issuerRef` when the source has `enableSSL: true`. Transport encryption alone does not require a user-provided issuer.

Save the following as `elasticsearch-tls.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: elasticsearch-tls
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Elasticsearch
      name: elasticsearch-source
    namespace: demo
  target:
    namespace: dev
    name: elasticsearch-tls
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
kubectl -n demo apply -f elasticsearch-tls.yaml
kubectl -n demo wait branch/elasticsearch-tls --for=condition=Ready --timeout=60m
kubectl -n demo get branch elasticsearch-tls
```

```shell
kubectl -n dev get certificates
kubectl -n dev get elasticsearch elasticsearch-tls -o yaml
```

Confirm that target certificates are ready and use branch-specific names. Use HTTPS and the issuer’s CA with `curl --cacert` for client verification. The plaintext commands and action connection settings from the main walkthroughs must be adapted to that TLS policy.

Use `kind: ClusterIssuer` when referencing a cluster-scoped issuer. A namespaced Issuer and its signing Secret must exist in the target namespace, including on another cluster.

## Cleanup optional examples

```shell
kubectl -n demo delete branch elasticsearch-tls --ignore-not-found --wait=true --timeout=30m
```

Return to the [same-cluster cleanup](/docs/guides/elasticsearch/branch/same-cluster/index.md#cleanup) after all dependent Branches are gone.
