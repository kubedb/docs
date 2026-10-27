---
title: YugabyteDB CRD
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-concepts-yugabytedb
    name: YugabyteDB
    parent: yugabytedb-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# YugabyteDB

## What is YugabyteDB

`YugabyteDB` is a Kubernetes `CustomResourceDefinition` (CRD) managed by KubeDB (`kubedb.com/v1alpha2`, short name `yb`). You describe a YugabyteDB universe in a `YugabyteDB` object and KubeDB provisions and reconciles the PetSets, Services, Secrets, certificates and AppBinding for it.

## YugabyteDB Spec

A `YugabyteDB` needs `apiVersion`, `kind`, `metadata` and `.spec`. Below is an object that uses most of the fields. The field list comes from the Go types in `apimachinery`. Only the first manifest (without `tls`, `placement`, `configuration`, `monitor`) has been applied end to end, see the notes under each field.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: YugabyteDB
metadata:
  name: yb-demo
  namespace: demo
spec:
  version: 2025.2.8
  replicationFactor: 3
  topology:
    master:
      podTemplate:
        spec:
          podPlacementPolicy:
            name: spread-across-zones
      storage:
        accessModes: [ReadWriteOnce]
        resources: {requests: {storage: 2Gi}}
    tserver:
      replicas: 3
      dataVolumes: 1
      podTemplate:
        spec:
          podPlacementPolicy:
            name: spread-across-zones
      storage:
        accessModes: [ReadWriteOnce]
        resources: {requests: {storage: 5Gi}}
  placement:
    cloud: kubernetes
    region: region1
    preferredZones: [zone-a]
  ysql:
    enabled: true
  ycql:
    enabled: false
  authSecret:
    name: yb-demo-auth
  configuration:
    secretName: yb-custom-config
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: yb-ca-issuer
    nodeToNode: true
    clientToNode: true
  monitor:
    agent: prometheus.io/operator
  deletionPolicy: WipeOut
```

### spec.version

`spec.version` is required. It is the name of a [YugabyteDBVersion](/docs/guides/yugabytedb/concepts/catalog.md) object.

```bash
$ kubectl get ybversion
NAME       VERSION    DB_IMAGE                                        DEPRECATED   AGE
2025.1.4   2025.1.4   docker.io/yugabytedb/yugabyte:2025.1.4.0-b103   false        5h39m
2025.2.8   2025.2.8   docker.io/yugabytedb/yugabyte:2025.2.8.0-b60    false        5h39m
```

### spec.replicationFactor

The number of masters and the default number of replicas of every tablet. Allowed values are `1`, `3`, `5` and `7`; the default is `3`. It cannot be changed after the database is created. With replication factor 1 there is nothing to fail over to, so no operation on such a database can be zero downtime.

### spec.topology

`spec.topology` configures the two roles, `master` and `tserver`. Each role takes the same fields:

- `replicas`: number of pods. For `master` it always equals the replication factor and cannot be scaled. For `tserver` it must be at least the replication factor and can be changed with [HorizontalScaling](/docs/guides/yugabytedb/scaling/horizontal-scaling/scale-horizontally/).
- `podTemplate`: pod customization (resources of the `yb-master` / `yb-tserver` container, image pull secrets, service account, and so on). Set `podTemplate.spec.podPlacementPolicy.name` to a `PlacementPolicy` (`apps.k8s.appscode.com/v1`) to spread the pods across zones and nodes. See [Clustering](/docs/guides/yugabytedb/clustering/architecture-overview.md) for an important constraint.
- `storageType`: `Durable` (default, PVCs) or `Ephemeral` (emptyDir).
- `storage`: the PVC template of the data volumes when `storageType` is `Durable`.
- `dataVolumes`: number of data volumes per pod, from 1 to 8. They are mounted at `/mnt/disk0` and up. It cannot be changed after creation.

YugabyteDB rejects writes when the data disk has less than 5% free space, so size the volumes with that margin.

### spec.placement

Names the cloud and region reported to YugabyteDB and orders the zones that should hold the tablet leaders.

- `cloud`: defaults to `kubernetes`.
- `region`: defaults to the `topology.kubernetes.io/region` label of the node, else `region1`.
- `preferredZones`: zones, in order of preference, for tablet leaders.

The zone of each pod is read from its node (`topology.kubernetes.io/zone`). The operator writes it to the pod as an annotation.

### spec.ysql and spec.ycql

`spec.ysql.enabled` (default `true`) starts the PostgreSQL compatible API on port 5433. `spec.ycql.enabled` (default `false`) starts the Cassandra compatible API on port 9042. Neither can be changed after creation. The YCQL operator path (auth bootstrap through `gocql`) was unit tested but never run in a cluster.

### spec.authSecret and spec.disableSecurity

On the first reconcile KubeDB creates the Secret `<name>-auth` with the keys `username` (`yugabyte`) and `password` (random) and sets the password in the database. The default `yugabyte/yugabyte` password is rejected afterwards. The client Service only routes to a tserver after this bootstrap finished (the pod label `kubedb.com/auth-bootstrapped=true`), so the well-known default password is never reachable through the Service. `spec.authSecret` references a Secret you provide; the username must stay `yugabyte`. See [Rotate Auth](/docs/guides/yugabytedb/rotate-auth/rotateauth.md). `spec.disableSecurity: true` turns authentication off; it is not recommended.

### spec.configuration

Custom flags. `secretName` references a Secret you create; `inline` holds the same keys as a map. The keys are `master.flags`, `tserver.flags`, `ysql_pg.conf` and `ysql_hba.conf`. Flags that the operator owns (addresses, placement, TLS, ports) are rejected or dropped. See [Custom Configuration](/docs/guides/yugabytedb/configuration/using-config-file.md).

### spec.tls

Cert-manager based TLS. `issuerRef` points to an `Issuer` or `ClusterIssuer`. `nodeToNode` (default `true`) encrypts master/tserver traffic and `clientToNode` (default `true`) encrypts client traffic. See [TLS](/docs/guides/yugabytedb/tls/overview/).

### spec.serviceTemplates

Customizes the Services that expose YugabyteDB (aliases `primary` and `stats`).

### spec.monitor

Enables monitoring. See [Monitoring](/docs/guides/yugabytedb/monitoring/overview.md).

### spec.halted

A shortcut for `deletionPolicy: Halt`. Setting `halted: true` removes the pods, PetSets, Services and AppBinding and keeps the PVCs and Secrets. Setting it back to `false` brings the same universe back with the same credentials (verified: 6 pods back, same credentials, data intact).

### spec.deletionPolicy

What happens to the data when the `YugabyteDB` object is deleted. Observed behavior:

| Policy | Result |
|--------|--------|
| `DoNotTerminate` | Deletion is blocked. |
| `Halt` | PetSets, Services, pods and AppBinding are removed; PVCs and all Secrets are kept. |
| `Delete` | PetSets, Services, pods, AppBinding and PVCs are removed; the auth Secret is kept. |
| `WipeOut` | Everything is removed, including PVCs and Secrets. |

### spec.healthChecker, spec.init

`spec.healthChecker` configures the periodic check (defaults: `periodSeconds: 20`, `timeoutSeconds: 10`, `failureThreshold: 3`). It checks that all masters and tservers are alive, no tablet is under-replicated, and that `SELECT 1` works through the Service. `spec.init` initializes the database from a script.

## Status

`status.phase` is `Provisioning`, `Ready`, `Critical` (a tserver is dead or a tablet is under-replicated) or `NotReady` (quorum lost). The conditions seen on a healthy universe are `ProvisioningStarted`, `ReplicaReady`, `MastersReady`, `TServersReady`, `PlacementSet`, `InternalUsersSynced`, `Ready`, `AcceptingConnection` and `Provisioned`.

## Next Steps

- Learn about the [YugabyteDBVersion CRD](/docs/guides/yugabytedb/concepts/catalog.md).
- Learn about the [AppBinding CRD](/docs/guides/yugabytedb/concepts/appbinding.md).
- Explore [YugabyteDBOpsRequest](/docs/guides/yugabytedb/concepts/opsrequest.md).
- Follow the [quickstart](/docs/guides/yugabytedb/quickstart/quickstart.md).
