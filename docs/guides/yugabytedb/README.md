---
title: YugabyteDB
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-readme
    name: YugabyteDB
    parent: yugabytedb-guides
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/yugabytedb/
aliases:
  - /docs/{{ .version }}/guides/yugabytedb/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

# Overview

KubeDB supports distributed SQL database deployment with YugabyteDB using the `YugabyteDB` CRD. A YugabyteDB universe is made of two roles: `yb-master` (metadata, load balancing, cluster configuration) and `yb-tserver` (tablet storage and the YSQL/YCQL query layers). KubeDB runs each role in its own PetSet and spreads the pods across zones with a `PlacementPolicy`.

## Supported YugabyteDB Features

| Features                                   |  Availability  |
|--------------------------------------------|:--------------:|
| Cluster provisioning (RF 1, 3, 5, 7)       |    &#10003;    |
| Zone aware placement (PlacementPolicy)     |    &#10003;    |
| Generated credentials (YSQL)               |    &#10003;    |
| TLS (cert-manager, hot reload)             |    &#10003;    |
| Custom configuration (flags, ysql_pg.conf) |    &#10003;    |
| Ops Requests                               |    &#10003;    |
| Monitoring (Prometheus)                    | &#10003; (see note) |
| YCQL                                       | off by default, not run in a cluster |
| VolumeExpansion                            | implemented, not verified end to end |
| Backup and restore (KubeStash)             | not verified |
| Read replicas                              | not verified on a cluster |
| YSQL Connection Manager                    | not verified on a cluster |
| Webserver TLS                              | not verified on a cluster |
| StorageMigration                           | not verified on a cluster |
| Point in time recovery (Rewind, Clone)     | not verified on a cluster |
| Autoscaler (compute, storage)              | not verified on a cluster |
| GitOps                                     | not verified on a cluster |

## Supported Ops Requests

- Restart
- VerticalScaling
- HorizontalScaling (tservers only)
- VolumeExpansion (not verified end to end)
- Reconfigure
- ReconfigureTLS
- RotateAuth
- UpdateVersion
- FinalizeUpgrade and RollbackUpgrade (not verified on a cluster)
- StorageMigration (not verified on a cluster)
- Rewind and Clone (not verified on a cluster)
- VerticalScaling in `InPlace` mode, CPU only (not verified on a cluster)

## Example YugabyteDB Manifest

This is the manifest used in the [quickstart](/docs/guides/yugabytedb/quickstart/quickstart.md). It creates a replication factor 3 universe with 3 masters and 3 tservers, one of each in every zone.

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
          containers:
          - name: yb-master
            resources:
              requests: {cpu: 500m, memory: 1Gi}
              limits: {memory: 1Gi}
      storage:
        accessModes: [ReadWriteOnce]
        resources: {requests: {storage: 2Gi}}
    tserver:
      replicas: 3
      podTemplate:
        spec:
          podPlacementPolicy:
            name: spread-across-zones
          containers:
          - name: yb-tserver
            resources:
              requests: {cpu: "1", memory: 2Gi}
              limits: {memory: 2Gi}
      storage:
        accessModes: [ReadWriteOnce]
        resources: {requests: {storage: 5Gi}}
  deletionPolicy: WipeOut
```

## What was verified

The guides in this section state, for each operation, what was measured on a three-node kind cluster (three zones, YugabyteDB `2025.2.8`, replication factor 3) while a load generator inserted and selected rows through the client Service. "Zero downtime" in these guides means: no failed calls except the connections that were open to the pod being restarted. Operations that were not run end to end are marked "not verified".

## User Guide

### Getting Started

- [Quickstart YugabyteDB](/docs/guides/yugabytedb/quickstart/quickstart.md): deploy your first universe and connect with `ysqlsh`.
- [Clustering and placement](/docs/guides/yugabytedb/clustering/architecture-overview.md): masters, tservers, replication factor, zones, and how the PlacementPolicy interacts with scaling.

### Concepts

- [YugabyteDB CRD](/docs/guides/yugabytedb/concepts/yugabytedb.md)
- [YugabyteDBVersion CRD](/docs/guides/yugabytedb/concepts/catalog.md)
- [YugabyteDBOpsRequest CRD](/docs/guides/yugabytedb/concepts/opsrequest.md)
- [AppBinding CRD](/docs/guides/yugabytedb/concepts/appbinding.md)

### Configuration

- [Custom Configuration](/docs/guides/yugabytedb/configuration/using-config-file.md)

### Monitoring

- [Monitoring Overview](/docs/guides/yugabytedb/monitoring/overview.md)

### TLS

- [TLS Overview](/docs/guides/yugabytedb/tls/overview/)
- [Configure TLS](/docs/guides/yugabytedb/tls/configure/)

### Day-2 Operations

- [Restart](/docs/guides/yugabytedb/restart/restart.md)
- [Vertical Scaling](/docs/guides/yugabytedb/scaling/vertical-scaling/scale-vertically/)
- [Horizontal Scaling](/docs/guides/yugabytedb/scaling/horizontal-scaling/scale-horizontally/)
- [Volume Expansion](/docs/guides/yugabytedb/volume-expansion/volume-expansion.md)
- [Reconfigure](/docs/guides/yugabytedb/reconfigure/reconfigure.md)
- [Reconfigure TLS](/docs/guides/yugabytedb/reconfigure-tls/reconfigure-tls.md)
- [Rotate Auth](/docs/guides/yugabytedb/rotate-auth/rotateauth.md)
- [Update Version](/docs/guides/yugabytedb/update-version/versionupgrading/)
- [Backup and Restore](/docs/guides/yugabytedb/backup/kubestash/logical/standalone-and-ha/) (not verified)
- [Finalize an upgrade](/docs/guides/yugabytedb/update-version/manual-finalize/finalize-upgrade.md) and [Roll back an upgrade](/docs/guides/yugabytedb/update-version/manual-finalize/rollback-upgrade.md) (not verified on a cluster)
- [In-place vertical scaling](/docs/guides/yugabytedb/scaling/vertical-scaling/in-place.md) (not verified on a cluster)
- [Webserver TLS](/docs/guides/yugabytedb/reconfigure-tls/webserver-tls.md) (not verified on a cluster)
- [Storage Migration](/docs/guides/yugabytedb/storage-migration/storage-migration.md) (not verified on a cluster)

### Phase 2 Features (none of these has been run on a cluster)

- [Read replicas](/docs/guides/yugabytedb/clustering/read-replicas.md)
- [YSQL Connection Manager](/docs/guides/yugabytedb/connection-manager/connection-manager.md)
- [Point in time recovery: Rewind and Clone](/docs/guides/yugabytedb/pitr/pitr.md)
- [Autoscaler](/docs/guides/yugabytedb/autoscaler/overview.md)
- [GitOps](/docs/guides/yugabytedb/gitops/gitops.md)
