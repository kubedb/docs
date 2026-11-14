---
title: Kafka Branch Configuration
menu:
  docs_{{ .version }}:
    identifier: kafka-branch-configuration
    name: Configuration
    parent: guides-kafka-branch
    weight: 70
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Kafka Branch Configuration

These optional examples reuse `kafka-source` from the [same-cluster tutorial](/docs/guides/kafka/branch/same-cluster/index.md). Keep the source and namespaces before that tutorial’s cleanup. Each Branch below has a distinct target name.

## Storage and resource overrides

`target.storageClassName` selects the clone’s StorageClass; it must resolve the source snapshot on the same backend. `target.resources` configures the target database containers independently of source resource requests. The manifests below set both explicitly. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md) for driver and backend requirements.

## Select a snapshot class

Create a compatible `ceph-rbd-snapclass` before this example. `volumeSnapshotClassName` selects the per-volume/import snapshot class; the compatible VolumeGroupSnapshotClass is discovered separately. Individual fallback snapshots do not provide an atomic snapshot across volumes.

Save the following as `kafka-snapshot.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: kafka-snapshot
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Kafka
      name: kafka-source
    namespace: demo
  target:
    namespace: dev
    name: kafka-snapshot
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1200Mi
      limits:
        memory: 1200Mi
  volumeSnapshotClassName: ceph-rbd-snapclass
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f kafka-snapshot.yaml
kubectl -n demo wait branch/kafka-snapshot --for=condition=Ready --timeout=60m
kubectl -n demo get branch kafka-snapshot
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

Save the following as `kafka-kept.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: kafka-kept
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Kafka
      name: kafka-source
    namespace: demo
  target:
    namespace: dev
    name: kafka-kept
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1200Mi
      limits:
        memory: 1200Mi
  deletionPolicy: Orphan
```

```shell
kubectl -n demo apply -f kafka-kept.yaml
kubectl -n demo wait branch/kafka-kept --for=condition=Ready --timeout=60m
kubectl -n demo get branch kafka-kept
```

```shell
kubectl -n demo delete branch kafka-kept --wait=true --timeout=30m
kubectl -n dev get kafka kafka-kept
kubectl -n dev get pvc,secrets -l app.kubernetes.io/instance=kafka-kept
```

Verify that the retained database still serves its copied data:

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it kafka-kept-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

```shell
kafka-get-offsets.sh --bootstrap-server kafka-kept-pods.dev.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:6
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-kept-pods.dev.svc.cluster.local:9092 \
  --consumer.config /opt/kafka/config/clientauth.properties --topic customers --from-beginning --max-messages 6 --timeout-ms 20000
```

```text
{"id": 1, "full_name": "Customer 1", "email": "customer1@example.com"}
{"id": 2, "full_name": "Customer 2", "email": "customer2@example.com"}
{"id": 3, "full_name": "Customer 3", "email": "customer3@example.com"}
{"id": 4, "full_name": "Customer 4", "email": "customer4@example.com"}
{"id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}
SOURCE-ONLY
```

Run `exit` to return to your terminal.

The results above assume you completed the source edit in the same-cluster tutorial. Preserve the retained auth Secret; the database still needs its credential. To remove the retained deployment and its data explicitly:

```shell
kubectl -n dev patch kafka kafka-kept --type=merge -p '{"spec":{"deletionPolicy":"WipeOut"}}'
kubectl -n dev delete kafka kafka-kept --wait=true --timeout=30m
```

## TLS-enabled sources

Configure a TLS-enabled source following the [Kafka TLS guide](/docs/guides/kafka/tls/combined.md) before creating this optional Branch. Use that TLS-enabled source’s name in `source.databaseRef.name` below if it differs from `kafka-source`. Prepare cert-manager and a target Issuer `branch-ca` in `dev`, including its signing Secret. Courier delivers database resources; it does not create this issuer.

Save the following as `kafka-tls.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: kafka-tls
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Kafka
      name: kafka-source
    namespace: demo
  target:
    namespace: dev
    name: kafka-tls
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1200Mi
      limits:
        memory: 1200Mi
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: branch-ca
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f kafka-tls.yaml
kubectl -n demo wait branch/kafka-tls --for=condition=Ready --timeout=60m
kubectl -n demo get branch kafka-tls
```

```shell
kubectl -n dev get certificates
kubectl -n dev get kafka kafka-tls -o yaml
```

Confirm that target certificates are ready and use branch-specific names. A TLS source’s broker client file uses `SASL_SSL` and a truststore. Use the target broker’s generated `/opt/kafka/config/clientauth.properties` for verification. A separate action pod needs its own matching TLS client configuration and mounted trust material; the plaintext post-action file is not sufficient.

Use `kind: ClusterIssuer` for a cluster-scoped issuer. For OCM, the issuer must exist on the target cluster, and a namespaced Issuer and its supporting Secret must be in the target namespace.

## Cleanup optional examples

```shell
kubectl -n demo delete branch kafka-snapshot kafka-tls --ignore-not-found --wait=true --timeout=30m
```

Return to the [same-cluster cleanup](/docs/guides/kafka/branch/same-cluster/index.md#cleanup) when all dependent Branches are gone.
