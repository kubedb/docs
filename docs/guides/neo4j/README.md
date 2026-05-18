---
title: Neo4j
menu:
  docs_{{ .version }}:
    identifier: neo4j-readme
    name: Neo4j
    parent: neo4j-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/neo4j/
aliases:
  - /docs/{{ .version }}/guides/neo4j/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

# Overview

KubeDB supports graph database deployment with Neo4j using the `Neo4j` CRD.

## Supported Neo4j Features

| Features                         | Availability |
|----------------------------------|:------------:|
| Standalone provisioning          |   &#10003;   |
| Cluster provisioning             |   &#10003;   |
| Monitoring                       |   &#10003;   |
| TLS                              |   &#10003;   |
| Ops Requests                     |   &#10003;   |

## Supported Ops Requests

- Reconfigure
- HorizontalScaling
- VerticalScaling
- VolumeExpansion
- StorageMigration
- UpdateVersion
- ReconfigureTLS
- RotateAuth
- Restart

## Example Neo4j Manifest

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-test
  namespace: demo
spec:
  replicas: 3
  deletionPolicy: WipeOut
  version: "2025.12.1"
  storage:
    storageClassName: local-path
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
```

## User Guide

- [Quickstart Neo4j](/docs/guides/neo4j/quickstart/quickstart.md) with KubeDB operator.
- [Cluster Architecture Overview](/docs/guides/neo4j/clustering/architecture-overview.md)
- [Neo4j CRD Concept](/docs/guides/neo4j/concepts/neo4j.md).
- [Neo4jVersion CRD Concept](/docs/guides/neo4j/concepts/catalog.md).
- [Neo4jOpsRequest CRD Concept](/docs/guides/neo4j/concepts/opsrequest.md).
- [Private Registry](/docs/guides/neo4j/private-registry/using-private-registry.md)
- [Custom RBAC](/docs/guides/neo4j/custom-rbac/using-custom-rbac.md)
- [Custom Configuration](/docs/guides/neo4j/configuration/using-config-file.md)
- [Monitoring](/docs/guides/neo4j/monitoring/overview.md) for metrics collection guidance.
- [Builtin Prometheus Monitoring](/docs/guides/neo4j/monitoring/using-builtin-prometheus.md)
- [Prometheus Operator Monitoring](/docs/guides/neo4j/monitoring/using-prometheus-operator.md)
- [TLS](/docs/guides/neo4j/tls/overview/) for protocol-level TLS guidance.
- [Configure TLS](/docs/guides/neo4j/tls/configure/)
- [Ops Request](/docs/guides/neo4j/ops-request/overview.md) for supported operational changes.
- [Reconfigure](/docs/guides/neo4j/reconfigure/overview.md)
- [Reconfigure Details](/docs/guides/neo4j/reconfigure/reconfigure.md)
- [Reconfigure TLS](/docs/guides/neo4j/reconfigure-tls/overview.md)
- [Reconfigure TLS Details](/docs/guides/neo4j/reconfigure-tls/reconfigure-tls.md)
- [Restart](/docs/guides/neo4j/restart/restart.md)
- [Rotate Auth](/docs/guides/neo4j/rotate-auth/overview.md)
- [Rotate Auth Details](/docs/guides/neo4j/rotate-auth/rotateauth.md)
- [Update Version](/docs/guides/neo4j/update-version/overview.md)
- [Update Version Details](/docs/guides/neo4j/update-version/versionupgrading/)
- [Volume Expansion](/docs/guides/neo4j/volume-expansion/overview.md)
- [Volume Expansion Details](/docs/guides/neo4j/volume-expansion/volume-expansion.md)
- [Migration](/docs/guides/neo4j/migration/)
- [StorageClass Migration](/docs/guides/neo4j/migration/storageMigration.md)
- [Horizontal Scaling](/docs/guides/neo4j/scaling/horizontal-scaling/overview.md)
- [Horizontal Scaling Details](/docs/guides/neo4j/scaling/horizontal-scaling/scale-horizontally/)
- [Vertical Scaling](/docs/guides/neo4j/scaling/vertical-scaling/overview.md)
- [Vertical Scaling Details](/docs/guides/neo4j/scaling/vertical-scaling/scale-vertically/)