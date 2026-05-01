---
title: Neo4j CRD
menu:
  docs_{{ .version }}:
    identifier: neo4j-concepts-neo4j
    name: Neo4j
    parent: neo4j-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Neo4j

## What is Neo4j

`Neo4j` is a KubeDB CRD that provides declarative management for Neo4j graph databases in Kubernetes.

## Neo4j Spec

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-test
  namespace: demo
spec:
  replicas: 3
  deletionPolicy: WipeOut
  version: "2025.12.1"
  storage:
    storageClassName: local-path
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
```

### Key fields

- `spec.version` refers to a `Neo4jVersion`.
- `spec.replicas` sets number of Neo4j instances.
- `spec.storageType` and `spec.storage` control persistence.
- `spec.authSecret` and `spec.disableSecurity` control authentication behavior.
- `spec.tls` configures TLS per protocol.
- `spec.disabledProtocols` allows protocol-level disablement.
- `spec.monitor` enables monitoring integration.
- `spec.deletionPolicy` controls cleanup.