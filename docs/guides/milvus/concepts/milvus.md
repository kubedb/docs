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
      name: my-release-minio
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
      name: my-release-minio
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
      name: my-release-minio
```

### spec.metaStorage

Milvus uses etcd as its metadata store.

- If you omit `spec.metaStorage`, KubeDB provisions and manages an internal etcd cluster for you.
- If you set `spec.metaStorage`, you can point Milvus at an externally managed etcd deployment.

The internal-etcd path is what the current quickstart guides use.

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

See [Configure TLS for Milvus](/docs/guides/milvus/tls/configure/index.md) for the end-to-end flow.

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
