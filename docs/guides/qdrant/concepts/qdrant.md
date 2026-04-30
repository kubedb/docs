---
title: Qdrant CRD
menu:
  docs_{{ .version }}:
    identifier: qdrant-concepts-qdrant
    name: Qdrant
    parent: qdrant-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Qdrant

## What is Qdrant

`Qdrant` is a KubeDB CRD for managing Qdrant vector databases with Kubernetes-native APIs.

## Qdrant Spec

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
  namespace: demo
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

### Key fields

- `spec.version` points to a `QdrantVersion`.
- `spec.mode` supports `Standalone` and `Distributed`.
- `spec.replicas` controls the number of pods.
- `spec.storageType` and `spec.storage` configure persistence.
- `spec.tls` configures client and p2p TLS.
- `spec.authSecret` and `spec.disableSecurity` control authentication.
- `spec.monitor` integrates monitoring.
- `spec.deletionPolicy` controls delete behavior.