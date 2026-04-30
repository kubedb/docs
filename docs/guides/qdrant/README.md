---
title: Qdrant
menu:
  docs_{{ .version }}:
    identifier: qdrant-readme
    name: Qdrant
    parent: qdrant-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/qdrant/
aliases:
  - /docs/{{ .version }}/guides/qdrant/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

# Overview

KubeDB supports Qdrant vector databases through the `Qdrant` CRD.

## Supported Qdrant Features

| Features                            | Availability |
|-------------------------------------|:------------:|
| Standalone provisioning             |   &#10003;   |
| Distributed provisioning            |   &#10003;   |
| TLS                                 |   &#10003;   |
| Backup (logical and volume snapshot)|   &#10003;   |
| Monitoring                          |   &#10003;   |
| Ops Requests                        |   &#10003;   |
| Autoscaler                          |      No      |

## Supported Ops Requests

- HorizontalScaling
- Reconfigure
- ReconfigureTLS
- Restart
- RotateAuth
- VerticalScaling
- VolumeExpansion
- UpdateVersion

## Example Qdrant Manifest

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
spec:
  version: 1.17.0
  mode: Distributed
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

## User Guide

- [Quickstart Qdrant](/docs/guides/qdrant/quickstart/quickstart.md) with KubeDB operator.
- [Qdrant CRD Concept](/docs/guides/qdrant/concepts/qdrant.md).
- [QdrantVersion CRD Concept](/docs/guides/qdrant/concepts/catalog.md).
- [QdrantOpsRequest CRD Concept](/docs/guides/qdrant/concepts/opsrequest.md).
- [QdrantAutoscaler CRD Concept](/docs/guides/qdrant/concepts/autoscaler.md).
- [Monitoring](/docs/guides/qdrant/monitoring/overview.md) for metrics collection guidance.
- [TLS](/docs/guides/qdrant/tls/overview.md) for client and p2p security guidance.
- [Backup](/docs/guides/qdrant/backup/overview.md) for logical and snapshot backup approaches.
- [Ops Request](/docs/guides/qdrant/ops-request/overview.md) for supported operational changes.
- [Autoscaler](/docs/guides/qdrant/autoscaler/overview.md) for current documentation status.
- [Private Registry](/docs/guides/qdrant/private-registry/using-private-registry.md)
- [Custom RBAC](/docs/guides/qdrant/custom-rbac/using-custom-rbac.md)
- [Custom Configuration](/docs/guides/qdrant/configuration/using-config-file.md)
- [Builtin Prometheus Monitoring](/docs/guides/qdrant/monitoring/using-builtin-prometheus.md)
- [Prometheus Operator Monitoring](/docs/guides/qdrant/monitoring/using-prometheus-operator.md)
- [Configure TLS](/docs/guides/qdrant/tls/configure/)
- [Reconfigure Details](/docs/guides/qdrant/reconfigure/reconfigure.md)
- [Reconfigure TLS Details](/docs/guides/qdrant/reconfigure-tls/reconfigure-tls.md)
- [Rotate Auth Details](/docs/guides/qdrant/rotate-auth/rotateauth.md)
- [Horizontal Scaling Details](/docs/guides/qdrant/scaling/horizontal-scaling/scale-horizontally/)
- [Vertical Scaling Details](/docs/guides/qdrant/scaling/vertical-scaling/scale-vertically/)
- [Update Version Details](/docs/guides/qdrant/update-version/versionupgrading/)
- [Volume Expansion Details](/docs/guides/qdrant/volume-expansion/volume-expansion.md)
- [Reconfigure](/docs/guides/qdrant/reconfigure/overview.md)
- [Reconfigure TLS](/docs/guides/qdrant/reconfigure-tls/overview.md)
- [Restart](/docs/guides/qdrant/restart/restart.md)
- [Rotate Auth](/docs/guides/qdrant/rotate-auth/overview.md)
- [Update Version](/docs/guides/qdrant/update-version/overview.md)
- [Volume Expansion](/docs/guides/qdrant/volume-expansion/overview.md)
- [Horizontal Scaling](/docs/guides/qdrant/scaling/horizontal-scaling/overview.md)
- [Vertical Scaling](/docs/guides/qdrant/scaling/vertical-scaling/overview.md)
- [Compute Autoscaler](/docs/guides/qdrant/autoscaler/compute/overview.md)
- [Storage Autoscaler](/docs/guides/qdrant/autoscaler/storage/overview.md)