---
title: Milvus CRD
menu:
  docs_{{ .version }}:
    identifier: milvus-concepts-milvus
    name: Milvus
    parent: milvus-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Milvus

## What is Milvus

`Milvus` is a KubeDB `CustomResourceDefinition` used to deploy and manage Milvus vector databases.

## Milvus Spec

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-cluster
  namespace: demo
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

### Key fields

- `spec.version` points to a `MilvusVersion`.
- `spec.objectStorage` is required for object data.
- `spec.topology.mode` supports `Standalone` or `Distributed`.
- `spec.topology.distributed` configures distributed roles.
- `spec.metaStorage` can configure external or managed etcd.
- `spec.storageType` and `spec.storage` define persistent data storage.
- `spec.authSecret`, `spec.configuration`, `spec.monitor`, and `spec.serviceTemplates` are optional controls.