---
title: HanaDB CRD
menu:
  docs_{{ .version }}:
    identifier: hanadb-concepts-hanadb
    name: HanaDB
    parent: hanadb-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# HanaDB

## What is HanaDB

`HanaDB` is a Kubernetes `CustomResourceDefinition` (CRD) in KubeDB to run SAP HANA databases in a Kubernetes-native way.

## HanaDB Spec

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

### Key fields

- `spec.version` refers to a `HanaDBVersion`.
- `spec.replicas` controls number of database instances.
- `spec.topology.mode` supports `Standalone` and `SystemReplication`.
- `spec.topology.systemReplication` configures replication and operation mode.
- `spec.storageType` and `spec.storage` define persistent data configuration.
- `spec.authSecret`, `spec.configuration`, `spec.podTemplate`, and `spec.monitor` are optional tuning controls.
- `spec.deletionPolicy` controls cleanup behavior.