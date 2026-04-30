---
title: DocumentDB CRD
menu:
  docs_{{ .version }}:
    identifier: documentdb-concepts-documentdb
    name: DocumentDB
    parent: documentdb-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# DocumentDB

## What is DocumentDB

`DocumentDB` is a Kubernetes `CustomResourceDefinition` (CRD) in KubeDB that manages DocumentDB-compatible databases.

## DocumentDB Spec

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: documentdb
  namespace: demo
spec:
  version: "pg17-0.109.0"
  storageType: Durable
  deletionPolicy: Delete
  replicas: 1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
```

### Key fields

- `spec.version` is required and points to a `DocumentDBVersion` resource.
- `spec.replicas` sets number of database instances.
- `spec.storageType` can be `Durable` or `Ephemeral`.
- `spec.storage` defines PVC settings for persistent data.
- `spec.authSecret` optionally references credentials.
- `spec.podTemplate` optionally customizes pods.
- `spec.serviceTemplates` optionally customizes services.
- `spec.deletionPolicy` controls object deletion behavior.