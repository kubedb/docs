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

KubeDB supports SAP HANA through the `HanaDB` CRD with standalone and system replication topology.

## Supported HanaDB Features

| Features                                                      | Availability |
|---------------------------------------------------------------|:------------:|
| Provisioning in Standalone mode                              |   &#10003;   |
| Provisioning in System Replication mode                      |   &#10003;   |
| Custom configuration and auth secret                          |   &#10003;   |
| Non-root deployment customization                              |   &#10003;   |
| Monitoring                                                     |   &#10003;   |

## Example HanaDB Manifest

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hana-cluster
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 2
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

- [Quickstart HanaDB](/docs/guides/hanadb/quickstart/quickstart.md) with KubeDB operator.
- [HanaDB CRD Concept](/docs/guides/hanadb/concepts/hanadb.md).
- [HanaDBVersion CRD Concept](/docs/guides/hanadb/concepts/catalog.md).
- [HanaDBOpsRequest CRD Concept](/docs/guides/hanadb/concepts/opsrequest.md).
- [Monitoring](/docs/guides/hanadb/monitoring/overview.md) for metrics collection guidance.
- [Ops Request](/docs/guides/hanadb/ops-request/overview.md).