---
title: Milvus
menu:
  docs_{{ .version }}:
    identifier: milvus-readme
    name: Milvus
    parent: milvus-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/milvus/
aliases:
  - /docs/{{ .version }}/guides/milvus/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

# Overview

KubeDB supports vector database deployment with Milvus using the `Milvus` CRD.

## Supported Milvus Features

| Features                         | Availability |
|----------------------------------|:------------:|
| Standalone provisioning          |   &#10003;   |
| Distributed provisioning         |   &#10003;   |
| Monitoring                       |   &#10003;   |
| TLS                              |      No      |
| Ops Requests                     |      No      |

## Example Milvus Manifest

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-cluster
spec:
  version: "2.6.11"
  objectStorage:
    configSecret:
      name: "my-release-minio"
  topology:
    mode: Distributed
    distributed:
      mixcoord:
        replicas: 2
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    storageClassName: local-path
    resources:
      requests:
        storage: 10Gi
```

## User Guide

- [Quickstart Milvus](/docs/guides/milvus/quickstart/quickstart.md) with KubeDB operator.
- [Milvus CRD Concept](/docs/guides/milvus/concepts/milvus.md).
- [MilvusVersion CRD Concept](/docs/guides/milvus/concepts/catalog.md).
- [MilvusOpsRequest CRD Concept](/docs/guides/milvus/concepts/opsrequest.md).
- [RBAC Quickstart](/docs/guides/milvus/quickstart/rbac.md)
- [Private Registry](/docs/guides/milvus/private-registry/using-private-registry.md)
- [Custom RBAC](/docs/guides/milvus/custom-rbac/using-custom-rbac.md)
- [Custom Configuration](/docs/guides/milvus/configuration/using-config-file.md)
- [Monitoring](/docs/guides/milvus/monitoring/overview.md) for metrics collection guidance.
- [Builtin Prometheus Monitoring](/docs/guides/milvus/monitoring/using-builtin-prometheus.md)
- [Prometheus Operator Monitoring](/docs/guides/milvus/monitoring/using-prometheus-operator.md)