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
  storageType: "Durable"
  version: "2025.11.2"
  configuration:
    secretName: neo4j-custom-config
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: neo4j-ca-issuer
  storage:
    storageClassName: local-path
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
```

### Key fields

- `spec.version` selects the `Neo4jVersion` to run.
- `spec.replicas` sets the number of Neo4j instances in the cluster.
- `spec.deletionPolicy` defines what KubeDB does with database resources when the `Neo4j` object is deleted.
- `spec.storageType` specifies the persistence mode. With `Durable`, you should provide `spec.storage`.
- `spec.storage` defines the persistent volume claim settings, such as `storageClassName`, `accessModes`, and requested storage size.
- `spec.configuration.secretName` references a Secret containing a custom `neo4j.conf`.
- `spec.monitor` enables monitoring integration. In this example, KubeDB creates a `ServiceMonitor` for Prometheus Operator.
- `spec.tls.issuerRef` tells KubeDB which cert-manager issuer to use for generating Neo4j TLS certificates.
