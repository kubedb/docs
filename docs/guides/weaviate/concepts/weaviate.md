---
title: Weaviate CRD
menu:
  docs_{{ .version }}:
    identifier: weaviate-concepts-weaviate
    name: Weaviate
    parent: weaviate-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Weaviate

## What is Weaviate

`Weaviate` is a KubeDB CRD for deploying and managing Weaviate vector databases in Kubernetes.

## Weaviate Spec

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-sample
  namespace: demo
spec:
  version: 1.33.1
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

### Key fields

- `spec.version` points to a `WeaviateVersion`.
- `spec.replicas` controls number of Weaviate pods.
- `spec.storageType` and `spec.storage` control data persistence.
- `spec.authSecret` and `spec.disableSecurity` control authentication options.
- `spec.configuration` can provide custom config.
- `spec.podTemplate` and `spec.serviceTemplates` customize runtime resources.
- `spec.deletionPolicy` controls deletion behavior.