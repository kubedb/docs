---
title: HanaDB CRD
menu:
  docs_{{ .version }}:
    identifier: hanadb-concepts-hanadb
    name: HanaDB
    parent: guides-hanadb-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# HanaDB

## What is HanaDB

`HanaDB` is a Kubernetes `Custom Resource Definition` (CRD). It provides a declarative configuration for
[SAP HANA](https://www.sap.com/products/technology-platform/hana.html) in a Kubernetes native way. You
only need to describe the desired database configuration in a `HanaDB` object, and the KubeDB operator
will create Kubernetes objects in the desired state for you.

## HanaDB Spec

As with all other Kubernetes objects, a `HanaDB` needs `apiVersion`, `kind`, and `metadata` fields. It
also needs a `.spec` section. Below is an example of a `HanaDB` object describing a System Replication
cluster.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hanadb-cluster
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 2
  storageType: Durable
  topology:
    mode: SystemReplication
    systemReplication:
      replicationMode: fullsync
      operationMode: logreplay_readaccess
  authSecret:
    kind: Secret
    name: hanadb-cluster-auth
  configuration:
    secretName: hanadb-configuration
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: hdb-ca-issuer
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 9668
      serviceMonitor:
        labels:
          release: prometheus
  podTemplate:
    spec:
      containers:
      - name: hanadb
        resources:
          requests:
            cpu: "1500m"
            memory: "8Gi"
          limits:
            cpu: "4"
            memory: "14Gi"
  storage:
    storageClassName: local-path
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 64Gi
  deletionPolicy: WipeOut
```

### spec.version

`spec.version` is a required field specifying the name of the [HanaDBVersion](/docs/guides/hanadb/concepts/catalog.md)
CRD where the docker images are specified. To list the available `HanaDBVersion` objects:

```bash
kubectl get hanadbversions
```

### spec.replicas

`spec.replicas` is an optional field that specifies the number of database instances (pods). It
defaults to `1`.

- For a **Standalone** database (no `spec.topology`), keep `replicas: 1`.
- For a **SystemReplication** cluster, `spec.replicas` is the total number of HANA nodes and must be
  `>= 2`. When `spec.replicas` is even, KubeDB automatically adds an **arbiter** pod (a lightweight raft
  tie-breaker that does not run the HANA database) so the cluster can elect a primary.

### spec.topology

`spec.topology` selects the deployment mode. When `spec.topology` is omitted, the database runs as a
single **Standalone** instance.

- `spec.topology.mode` — one of `Standalone` or `SystemReplication`.
- `spec.topology.systemReplication.replicationMode` — HANA System Replication mode. One of `sync`,
  `syncmem`, `async`, or `fullsync`. Defaults to `sync`.
- `spec.topology.systemReplication.operationMode` — HANA operation mode. One of `logreplay`,
  `delta_datashipping`, or `logreplay_readaccess`. Defaults to `logreplay`. Use `logreplay_readaccess`
  to allow read-only queries on the secondary; in that case KubeDB also creates a dedicated
  `secondary-<name>` Service.

### spec.storageType

`spec.storageType` is an optional field that can be `Durable` (default) or `Ephemeral`.

- `Durable` — the database uses a `PersistentVolumeClaim` created from `spec.storage`.
- `Ephemeral` — the database uses an `emptyDir` volume; data is lost when the pod restarts. HANA is
  disk-heavy, so `Ephemeral` is only suitable for throwaway testing.

### spec.storage

When `spec.storageType` is `Durable`, `spec.storage` defines the PVC for the HANA data volume
(`/hana/mounts`):

- `spec.storage.storageClassName` — the name of the `StorageClass` to use.
- `spec.storage.accessModes` — usually `["ReadWriteOnce"]`.
- `spec.storage.resources.requests.storage` — requested volume size. HANA is disk-heavy; use a
  realistic size (the guides use `64Gi`).

### spec.authSecret

`spec.authSecret` is an optional field referencing the `Secret` that holds the HANA `SYSTEM` user
credentials. If omitted, KubeDB creates a secret named `<name>-auth` with an auto-generated password.
The secret has type `kubernetes.io/basic-auth` with the following keys:

- `username` — always `SYSTEM`.
- `password` — the SYSTEM database password.
- `password.json` — `{"master_password":"<password>"}`, the format consumed by the HANA container.

> The HANA username must remain `SYSTEM`. See [Rotate Authentication](/docs/guides/hanadb/rotate-authentication/rotate-authentication.md)
> for changing the password safely.

### spec.configuration

`spec.configuration` lets you supply a custom HANA `global.ini`:

- `spec.configuration.secretName` — name of a `Secret` whose `global.ini` key holds the configuration.
- `spec.configuration.applyConfig` — an inline `map[string]string` keyed by `global.ini`.

See [Custom Configuration](/docs/guides/hanadb/configuration/using-config-file.md). At runtime, custom
configuration is merged into HANA's `global.ini`.

### spec.tls

`spec.tls` configures TLS using [cert-manager](https://cert-manager.io/). When set, `spec.tls.issuerRef`
is required and must reference a cert-manager `Issuer` or `ClusterIssuer`. KubeDB provisions three
certificates with the aliases `server`, `client`, and `metrics-exporter`. See [TLS/SSL Encryption](/docs/guides/hanadb/tls/overview.md).

### spec.monitor

`spec.monitor` enables Prometheus metrics through the bundled `hanadb_exporter`. Set
`spec.monitor.agent` to `prometheus.io/builtin` or `prometheus.io/operator`. See
[Monitoring](/docs/guides/hanadb/monitoring/using-builtin-prometheus.md).

### spec.podTemplate

`spec.podTemplate` customizes the pods that run the database. The most common use is setting resources
on the named containers. A HanaDB pod can run these containers:

- `hanadb` — the SAP HANA database (always present).
- `hanadb-coordinator` — the raft coordinator sidecar (present only for `SystemReplication` clusters; it
  elects the primary and labels pods).
- `exporter` — the Prometheus exporter (present only when `spec.monitor` is set).

### spec.deletionPolicy

`spec.deletionPolicy` controls what happens to the data and auxiliary objects when the `HanaDB` object
is deleted:

| Value           | PetSet / Pods | PVC      | Auth & config secrets |
|-----------------|:-------------:|:--------:|:---------------------:|
| `DoNotTerminate`| blocks delete | —        | —                     |
| `Halt`          | deleted       | kept     | kept                  |
| `Delete`        | deleted       | deleted  | kept                  |
| `WipeOut`       | deleted       | deleted  | deleted               |

`Delete` is the default. For more details, see [this blog post](https://appscode.com/blog/post/deletion-policy/).

### spec.healthChecker

`spec.healthChecker` tunes how KubeDB probes the database (defaults: `periodSeconds: 20`,
`timeoutSeconds: 10`, `failureThreshold: 3`). Setting `spec.healthChecker.disableWriteCheck: true`
keeps the periodic write probe from creating its bookkeeping tenant database/table.

## HanaDB Status

`spec.status.phase` reflects the overall state of the database and is one of:

`Provisioning`, `DataRestoring`, `Ready`, `Critical`, `NotReady`, `Halted`, or `Unknown`.

KubeDB derives the phase from the `status.conditions` array (for example `ProvisioningStarted`,
`ReplicaReady`, `AcceptingConnection`, `Ready`, and `Provisioned`).

## Next Steps

- Deploy a [Standalone HanaDB](/docs/guides/hanadb/quickstart/quickstart.md).
- Deploy a [System Replication cluster](/docs/guides/hanadb/clustering/system-replication.md).
- Learn about [HanaDBVersion](/docs/guides/hanadb/concepts/catalog.md) and [HanaDBOpsRequest](/docs/guides/hanadb/concepts/opsrequest.md).
