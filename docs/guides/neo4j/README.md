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

| Features                                                      | Availability |
|---------------------------------------------------------------|:------------:|
| Standalone provisioning                                       |   &#10003;   |
| Cluster provisioning                                          |   &#10003;   |
| Custom Configuration                                          |   &#10003;   |
| Custom RBAC                                                   |   &#10003;   |
| Private Registry                                              |   &#10003;   |
| Monitoring using Prometheus and Grafana                       |   &#10003;   |
| Builtin Prometheus Discovery                                  |   &#10003;   |
| Operator managed Prometheus Discovery                         |   &#10003;   |
| Authentication & Authorization (TLS)                          |   &#10003;   |
| Reconfigurable TLS Certificates (Add, Remove, Rotate, Update) |   &#10003;   |
| Rotate Authentication                                         |   &#10003;   |
| Automated Version Update                                      |   &#10003;   |
| Automated Horizontal Scaling                                  |   &#10003;   |
| Automated Vertical Scaling                                    |   &#10003;   |
| Automated Volume Expansion                                    |   &#10003;   |
| StorageClass Migration                                        |   &#10003;   |
| Reconfigure                                                   |   &#10003;   |
| Restart                                                       |   &#10003;   |

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

### Getting Started

- [Quickstart Neo4j](/docs/guides/neo4j/quickstart/quickstart.md) — deploy your first Neo4j cluster with KubeDB.
- [Cluster Architecture Overview](/docs/guides/neo4j/clustering/architecture-overview.md) — understand cluster topology, Raft consensus, and fault tolerance.
- [RBAC Permissions](/docs/guides/neo4j/quickstart/rbac.md) — RBAC resources KubeDB creates for Neo4j pods.

### Concepts

- [Neo4j CRD](/docs/guides/neo4j/concepts/neo4j.md) — full reference for all `Neo4j` spec fields.
- [Neo4jVersion CRD](/docs/guides/neo4j/concepts/catalog.md) — image and version catalog.
- [Neo4jOpsRequest CRD](/docs/guides/neo4j/concepts/opsrequest.md) — day-2 operations reference with sample manifests.
- [AppBinding CRD](/docs/guides/neo4j/concepts/appbinding.md) — how KubeDB exposes connection details for backup tools.

### Configuration & Infrastructure

- [Custom Configuration](/docs/guides/neo4j/configuration/using-config-file.md) — pass Neo4j settings via a Kubernetes Secret.
- [Private Registry](/docs/guides/neo4j/private-registry/using-private-registry.md) — pull Neo4j images from a private Docker registry.
- [Custom RBAC](/docs/guides/neo4j/custom-rbac/using-custom-rbac.md) — provide your own ServiceAccount and Role instead of the auto-generated ones.

### Monitoring

- [Monitoring Overview](/docs/guides/neo4j/monitoring/overview.md) — how KubeDB exposes Neo4j metrics.
- [Builtin Prometheus](/docs/guides/neo4j/monitoring/using-builtin-prometheus.md) — scrape metrics without the Prometheus Operator.
- [Prometheus Operator](/docs/guides/neo4j/monitoring/using-prometheus-operator.md) — use a `ServiceMonitor` with the Prometheus Operator.

### Day-2 Operations

- [TLS — How It Works](/docs/guides/neo4j/tls/overview/) — how KubeDB provisions TLS certificates via cert-manager.
- [Configure TLS](/docs/guides/neo4j/tls/configure/) — enable TLS on a new or existing cluster.
- [Reconfigure — How It Works](/docs/guides/neo4j/reconfigure/overview.md) — how KubeDB applies config changes internally.
- [Reconfigure](/docs/guides/neo4j/reconfigure/reconfigure.md) — change Neo4j settings at runtime.
- [Reconfigure TLS — How It Works](/docs/guides/neo4j/reconfigure-tls/overview.md) — how KubeDB rotates or removes TLS.
- [Reconfigure TLS](/docs/guides/neo4j/reconfigure-tls/reconfigure-tls.md) — add, rotate, change issuer, or remove TLS.
- [Restart](/docs/guides/neo4j/restart/restart.md) — rolling restart of all Neo4j pods.
- [Rotate Auth — How It Works](/docs/guides/neo4j/rotate-auth/overview.md) — how KubeDB rotates credentials.
- [Rotate Auth](/docs/guides/neo4j/rotate-auth/rotateauth.md) — rotate Neo4j passwords with or without a user-provided Secret.
- [Update Version — How It Works](/docs/guides/neo4j/update-version/overview.md) — how KubeDB performs rolling version upgrades.
- [Update Version](/docs/guides/neo4j/update-version/versionupgrading/) — upgrade to a newer Neo4j release.
- [Horizontal Scaling — How It Works](/docs/guides/neo4j/scaling/horizontal-scaling/overview.md) — how KubeDB adds or removes cluster members.
- [Horizontal Scaling](/docs/guides/neo4j/scaling/horizontal-scaling/scale-horizontally/) — add or remove Neo4j server pods.
- [Vertical Scaling — How It Works](/docs/guides/neo4j/scaling/vertical-scaling/overview.md) — how KubeDB adjusts pod resources.
- [Vertical Scaling](/docs/guides/neo4j/scaling/vertical-scaling/scale-vertically/) — resize CPU and memory for Neo4j pods.
- [Volume Expansion — How It Works](/docs/guides/neo4j/volume-expansion/overview.md) — how KubeDB expands PVCs.
- [Volume Expansion](/docs/guides/neo4j/volume-expansion/volume-expansion.md) — increase persistent storage size online or offline.
- [StorageClass Migration](/docs/guides/neo4j/migration/storageMigration.md) — migrate Neo4j data to a different StorageClass.