---
title: Milvus CRD
menu:
  docs_{{ .version }}:
    identifier: milvus-concepts-milvus
    name: Milvus
    parent: milvus-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Milvus

## What is Milvus

`Milvus` is a Kubernetes `CustomResourceDefinition` (CRD). It provides declarative configuration for [Milvus](https://milvus.io/) in a Kubernetes-native way. You describe the desired Milvus deployment in a `Milvus` object, and the KubeDB operator creates and reconciles the required Kubernetes resources for you.

KubeDB supports Milvus in two topologies:

- `Standalone` - a single all-in-one Milvus workload.
- `Distributed` - Milvus roles run as separate workloads that can be scaled independently.

## Sample Milvus Objects

### Standalone

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-standalone
  namespace: demo
spec:
  version: "2.6.11"
  topology:
    mode: Standalone
  objectStorage:
    configSecret:
      name: milvus-storage-config
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    storageClassName: local-path
    resources:
      requests:
        storage: 1Gi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  tls:
    issuerRef:
      name: milvus-issuer
      kind: Issuer
      apiGroup: cert-manager.io
    external:
      mode: mTLS
    internal:
      mode: TLS
  deletionPolicy: WipeOut
```

### Distributed

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-cluster
  namespace: demo
spec:
  version: "2.6.11"
  objectStorage:
    configSecret:
      name: milvus-storage-config
  topology:
    mode: Distributed
    distributed:
      mixcoord:
        replicas: 2
      datanode:
        replicas: 2
      proxy:
        replicas: 2
      querynode:
        replicas: 2
      streamingnode:
        replicas: 3
        storageType: Durable
        storage:
          accessModes:
            - ReadWriteOnce
          storageClassName: local-path
          resources:
            requests:
              storage: 10Gi
  deletionPolicy: WipeOut
```

## Milvus Spec

Like any Kubernetes resource, a `Milvus` object has `apiVersion`, `kind`, and `metadata` fields. The database-specific configuration lives under `spec`.

### spec.version

`spec.version` is required. It selects the `MilvusVersion` catalog entry that contains the Milvus image and version metadata KubeDB should run.

You can see the available versions with:

```bash
$ kubectl get milvusversions
```

### spec.objectStorage.configSecret

`spec.objectStorage` is required. Milvus stores its segments and logs in object storage, so you must provide a secret with the storage endpoint and credentials before creating the database.

In the current guides, this is usually a MinIO-backed secret referenced as:

```yaml
spec:
  objectStorage:
    configSecret:
      name: milvus-storage-config
```

### spec.objectStorage.tls

Set `spec.objectStorage.tls` when the object storage serves a certificate signed by a private CA, for example Ceph RGW behind an OpenShift service CA. Without it, Milvus fails with `x509: certificate signed by unknown authority`. This field only configures how Milvus verifies the object storage server. It is separate from [`spec.tls`](/docs/guides/milvus/tls/overview.md), which secures Milvus's own listeners.

Name the CA bundle in exactly one of `caSecret` or `caConfigMap`. Both need a `name` and a `key` that holds the PEM CA certificate, in the same namespace as the Milvus object:

```yaml
spec:
  objectStorage:
    configSecret:
      name: milvus-s3
    tls:
      caConfigMap:
        name: openshift-service-ca.crt   # an injected service CA ConfigMap in this namespace
        key: service-ca.crt
```

KubeDB mounts the CA at `/milvus/object-storage-ca/ca.crt` in every Milvus pod and sets `minio.useSSL: true` and `minio.ssl.tlsCACert` in the rendered `milvus.yaml`. The endpoint still comes from the object storage secret, so set its `address` and `port` (usually `443`) for the TLS endpoint.

If the same `minio` key is set in more than one place, the last one in this order wins: KubeDB defaults, the object storage secret, `spec.objectStorage.tls`, then `spec.configuration`.

The matching Milvus archiver, backup-plugin and CSI-plugin images also use this CA. They read the effective storage settings from the generated configuration Secret, so inline overrides apply consistently to Milvus and its backup workers. An arbitrary CA path mounted only through `podTemplate` is not accessible to these workers; use `spec.objectStorage.tls` for backup-compatible private-CA configuration.

This is independent of `spec.tls`: Milvus listener TLS can be disabled while HTTPS to object storage remains enabled. A backup destination has its own TLS settings in KubeStash `BackupStorage`; supply that endpoint's CA in the credential Secret under `CA_CERT_DATA`. The two endpoints can use different CAs.

Before restoring into another namespace, create that target's object-storage credential Secret and selected CA Secret or ConfigMap there. Restore uses the target's configuration and does not replace its CA with the source database's CA.

Each Milvus needs its own object storage path: the `rootPath` key of the secret, which defaults to `files`. Before the first pod of a new Milvus starts, KubeDB checks that path. If the meta etcd is empty and the path already holds objects, the Milvus does not start, because it could not write there. KubeDB then sets the `ObjectStorageChecked` condition to `False` with reason `PathInUse`:

```bash
$ kubectl get milvus.kubedb.com -n demo milvus-standalone -o jsonpath='{.status.conditions[?(@.type=="ObjectStorageChecked")].message}'
```

To continue, clean the path or set a new `rootPath` in the secret. KubeDB checks again every minute and starts the Milvus by itself. A Milvus that already holds its metadata in etcd passes this check, for example an adopted cluster that keeps its etcd. Note that deleting a Milvus with the `Halt` policy also removes the KubeDB-managed etcd it owns, so recreating it stops at this check. If you are sure the path is safe, set the annotation `kubedb.com/milvus-skip-object-storage-check: "true"` on the Milvus. If the check cannot run, for example when S3 is unreachable or `useIAM` is set, KubeDB records a warning and starts the Milvus.

### spec.metaStorage

<!-- The tls/authSecret/storage fields below are drafted from the Milvus/Etcd CRD schemas and controller source, not yet verified against a live cluster. -->

Milvus uses etcd as its metadata store. By default (and whenever `spec.metaStorage.externallyManaged` is not `true`), KubeDB provisions and manages an internal [`Etcd`](/docs/guides/etcd/concepts/etcd.md) database for this - the same `Etcd` CRD and operator used to run etcd as a standalone KubeDB database. This is the path the current quickstart guides use.

Fields on `spec.metaStorage`:

- `externallyManaged` - set to `true` to point Milvus at an etcd cluster you manage yourself. `endpoints` is then required, and `size`/`storage`/`tls`/`authSecret` are ignored.
- `endpoints` - client endpoints of an externally managed etcd. Required (and only used) when `externallyManaged: true`.
- `size` - number of members in the internally-managed etcd cluster.
- `storageType` / `storage` - storage behavior for the internal etcd cluster's PVCs, same shape as any other KubeDB database's `spec.storage`. Point `storage.storageClassName` at an encrypted-volume StorageClass for encryption at rest - no other configuration is needed for that.
- `tls` - issues server/client/peer certificates for the internal etcd cluster, same shape as `Etcd.spec.tls`. Only applies when `externallyManaged` is not `true`; requires `tls.issuerRef`.
- `authSecret` - the internal etcd cluster's root credential. Omit it to let the etcd operator auto-generate one. Only applies when `externallyManaged` is not `true`.

```yaml
spec:
  metaStorage:
    size: 3
    storage:
      storageClassName: encrypted-ssd
      resources:
        requests:
          storage: 10Gi
    tls:
      issuerRef:
        apiGroup: cert-manager.io
        kind: Issuer
        name: milvus-meta-etcd-issuer
    authSecret:
      name: milvus-meta-etcd-auth
```

### spec.topology

`spec.topology.mode` chooses the deployment shape:

- `Standalone` creates one Milvus workload.
- `Distributed` creates one workload per Milvus role.

For distributed Milvus, `spec.topology.distributed` is keyed by role:

- `mixcoord`
- `proxy`
- `datanode`
- `querynode`
- `streamingnode`

Each role can carry its own `replicas`, `podTemplate`, and, where relevant, storage settings.

Only `streamingnode` carries persistent Milvus storage in distributed mode. That is why storage operations such as [volume expansion](/docs/guides/milvus/volume-expansion/guide.md), [storage migration](/docs/guides/milvus/storage-migration/guide.md), and [storage autoscaling](/docs/guides/milvus/autoscaler/storage/guide.md) target `streamingnode`.

### spec.storageType and spec.storage

For standalone Milvus, persistent storage is configured at the top level:

```yaml
spec:
  storageType: Durable
  storage:
    storageClassName: local-path
    resources:
      requests:
        storage: 1Gi
```

- `storageType` can be `Durable` or `Ephemeral`.
- When `storageType` is `Durable`, `spec.storage` defines the PVC template for the standalone workload.

For distributed Milvus, storage is configured under `spec.topology.distributed.streamingnode`.

### spec.authSecret and spec.disableSecurity

Milvus authentication is enabled by default because `spec.disableSecurity` defaults to `false`.

- If `spec.authSecret` is omitted, KubeDB generates a `kubernetes.io/basic-auth` secret named `<db-name>-auth` with user `root` and a random password.
- If `spec.authSecret.name` is set, KubeDB uses that secret instead.
- `spec.authSecret.rotateAfter` and `spec.authSecret.activeFrom` are used by the Recommendation Engine to decide when a [RotateAuth](/docs/guides/milvus/rotate-auth/guide.md) recommendation should be emitted.

### spec.tls

`spec.tls` enables TLS for Milvus.

Milvus has two TLS surfaces:

- `external` controls client-facing traffic.
- `internal` controls inter-component traffic between Milvus roles.

The guides use cert-manager-backed TLS via `spec.tls.issuerRef`. Once TLS is enabled, KubeDB issues the certificate secrets, mounts them into the Milvus pods, and switches the generated `AppBinding` scheme to `https`.

See [Configure TLS for Milvus](/docs/guides/milvus/tls/guide.md) for the end-to-end flow.

### spec.monitor

`spec.monitor` configures metrics exposure for monitoring. The current Milvus guides use Prometheus Operator integration:

```yaml
spec:
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

See [Monitoring Milvus](/docs/guides/milvus/monitoring/using-prometheus-operator.md).

### spec.podTemplate

`spec.podTemplate` lets you customize the Milvus pods and the PetSet metadata KubeDB creates. This is where per-pod scheduling rules, environment variables, annotations, labels, and resource requests can be set.

In distributed mode, role-specific `podTemplate` blocks can also be provided under `spec.topology.distributed.<role>`.

### spec.deletionPolicy

`spec.deletionPolicy` controls what happens to the database resources when the `Milvus` object is deleted. Common values are:

- `Halt`
- `Delete`
- `WipeOut`
- `DoNotTerminate`

The quickstart guides switch this to `WipeOut` before cleanup so all generated resources can be removed cleanly.

No policy removes data from object storage, including `WipeOut`. To reuse the path for a new Milvus, clean it first.

### spec.halted

`spec.halted` is used to stop the database while keeping the `Milvus` object. When halted, the operator tears down the running database workloads but retains the declarative object so it can be brought back later.

## Resources Created by KubeDB

Once a `Milvus` object becomes `Ready`, KubeDB creates and manages:

- one or more Milvus PetSets and pods,
- the client and metrics services,
- persistent volume claims for standalone Milvus or distributed `streamingnode`,
- the authentication secret,
- the rendered configuration secret,
- TLS certificate secrets when TLS is enabled,
- an `AppBinding` that describes how to connect to the database.

## Related Concepts

- [MilvusAutoscaler](/docs/guides/milvus/concepts/milvusautoscaler.md)
- [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md)
