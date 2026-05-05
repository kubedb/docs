---
title: HanaDB
menu:
  docs_{{ .version }}:
    identifier: hanadb-readme
    name: HanaDB
    parent: hanadb-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/hanadb/
aliases:
  - /docs/{{ .version }}/guides/hanadb/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

# Overview

KubeDB supports SAP HANA through the `HanaDB` CRD. You can provision standalone SAP HANA instances or system replication clusters from Kubernetes.

## Supported HanaDB Features

| Features                                                      | Availability |
|---------------------------------------------------------------|:------------:|
| Provisioning in Standalone mode                               |   &#10003;   |
| Provisioning in System Replication mode                       |   &#10003;   |
| Custom configuration and auth secret                          |   &#10003;   |
| Custom pod template and service account                       |   &#10003;   |
| Private registry images                                       |   &#10003;   |
| Built-in Prometheus discovery                                 |   &#10003;   |
| Prometheus Operator monitoring                                |   &#10003;   |

## Supported HanaDB Versions

KubeDB supports the following SAP HANA version:

- `2.0.82`

## Example HanaDB Manifest

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hanadb-cluster
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 3
  storageType: "Durable"
  topology:
    mode: SystemReplication
    systemReplication:
      replicationMode: fullsync
      operationMode: logreplay_readaccess
  storage:
    accessModes: ["ReadWriteOnce"]
    resources:
      requests:
        storage: 64Gi
    storageClassName: local-path
```

## User Guide

- [HanaDB Quickstart](/docs/guides/hanadb/quickstart/quickstart.md) with KubeDB operator.
- [HanaDB CRD](/docs/guides/hanadb/concepts/hanadb.md).
- [HanaDBVersion CRD](/docs/guides/hanadb/concepts/catalog.md).
- [AppBinding](/docs/guides/hanadb/concepts/appbinding.md).
- [Standalone and System Replication](/docs/guides/hanadb/clustering/system-replication.md).
- [Custom Configuration](/docs/guides/hanadb/configuration/using-config-file.md).
- [Private Registry](/docs/guides/hanadb/private-registry/using-private-registry.md).
- [Monitoring](/docs/guides/hanadb/monitoring/overview.md) for metrics collection guidance.
