---
title: DB2 CRD
menu:
  docs_{{ .version }}:
    identifier: db2-db2-concepts
    name: DB2
    parent: db2-concepts-db2
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# DB2

## What is DB2

`DB2` is a Kubernetes `CustomResourceDefinition` (CRD) provided by KubeDB. It lets you run and manage IBM DB2 with Kubernetes-native declarative APIs.

## DB2 Spec

Like all Kubernetes resources, a `DB2` object needs `apiVersion`, `kind`, and `metadata`, and it uses `.spec` to define the desired state.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DB2
metadata:
  name: db2
  namespace: demo
spec:
  version: 11.5.8.0
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
  deletionPolicy: Delete
```

### Key fields

- `spec.version` is required and points to a `DB2Version`.
- `spec.replicas` controls number of DB2 pods.
- `spec.storageType` supports `Durable` (PVC) and `Ephemeral`.
- `spec.storage` defines the PVC template when using durable storage.
- `spec.authSecret` references user credentials.
- `spec.podTemplate` customizes the DB2 pods.
- `spec.serviceTemplates` customizes services exposed by KubeDB.
- `spec.deletionPolicy` controls cleanup behavior on CR deletion.
- `spec.healthChecker` controls readiness/liveness checks.