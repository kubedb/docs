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

`Neo4j` is a Kubernetes `CustomResourceDefinition` (CRD) managed by KubeDB. It provides declarative configuration for Neo4j in a Kubernetes-native way. You define the desired state in a `Neo4j` object, and KubeDB provisions and reconciles the database resources.

## Neo4j Spec

As with all other Kubernetes objects, a Neo4j needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

Below is an example Neo4j object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-test
  namespace: demo
spec:
  version: "2025.12.1"
  replicas: 3
  storageType: Durable
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
      apiGroup: cert-manager.io
      kind: Issuer
      name: neo4j-ca-issuer
    bolt:
      mode: mTLS
    cluster:
      mode: mTLS
  podTemplate:
    spec:
      serviceAccountName: neo4j-test
      imagePullSecrets:
        - name: myregistrykey
  storage:
    storageClassName: local-path
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

### spec.version

`spec.version` is a required field that specifies the name of the [Neo4jVersion](/docs/guides/neo4j/concepts/catalog.md) CRD where database image and version metadata are defined.

To see available versions in your cluster:

```bash
$ kubectl get neo4jversions
NAME        VERSION                DB_IMAGE                                       DEPRECATED   AGE
2025.10.1   2025.10.1-enterprise   docker.io/library/neo4j:2025.10.1-enterprise                12d
2025.11.2   2025.11.2-enterprise   docker.io/library/neo4j:2025.11.2-enterprise                12d
2025.12.1   2025.12.1-enterprise   docker.io/library/neo4j:2025.12.1-enterprise                12d
```

### spec.replicas

`spec.replicas` specifies the number of Neo4j server pods in the cluster.

For production deployments, use multiple replicas so the cluster can continue serving traffic during node or pod disruptions.

### spec.storageType

`spec.storageType` is an optional field that specifies the storage mode. It can be `Durable` or `Ephemeral`.

- `Durable`: Uses persistent volumes and keeps data across pod restarts.
- `Ephemeral`: Uses temporary storage and is mainly suitable for testing.

### spec.storage

If `spec.storageType` is `Durable` (or not explicitly set), `spec.storage` is required. This field defines PVC settings used by Neo4j pods.

- `spec.storage.storageClassName` is the StorageClass used to provision PVCs.
- `spec.storage.accessModes` defines how volumes are mounted (for example, `ReadWriteOnce`).
- `spec.storage.resources.requests.storage` defines the requested volume size.

To check available StorageClass resources:

```bash
$ kubectl get storageclass
```

### spec.configuration

`spec.configuration` is an optional field used to provide custom Neo4j configuration.

- `spec.configuration.secretName` references a Secret containing Neo4j key-value settings.

KubeDB merges the provided settings into Neo4j configuration and reconciles the cluster accordingly. See the [configuration guide](/docs/guides/neo4j/configuration/using-config-file.md).

### spec.monitor

`spec.monitor` is an optional field that enables monitoring integration for Neo4j.

In the example above, KubeDB is configured for Prometheus Operator (`prometheus.io/operator`) with a `ServiceMonitor` definition.

See [monitoring overview](/docs/guides/neo4j/monitoring/overview.md) for setup and verification.

### spec.tls

`spec.tls` is an optional field for enabling and configuring TLS for Neo4j protocols.

- `spec.tls.issuerRef` tells KubeDB which cert-manager issuer to use for certificates.
- `spec.tls.bolt.mode` and `spec.tls.cluster.mode` set the TLS mode for the Bolt and cluster channels respectively.

**TLS mode values:**

| Mode | Behavior |
|------|----------|
| `TLS` | Encrypts traffic in transit. Clients do not need to present a certificate. |
| `mTLS` | Mutual TLS — both sides present certificates. Clients must have a valid certificate signed by the same CA. |

See [TLS overview](/docs/guides/neo4j/tls/overview/) and [Reconfigure TLS](/docs/guides/neo4j/reconfigure-tls/overview.md).

### spec.podTemplate

`spec.podTemplate` is an optional field for customizing Neo4j pods.

Common examples:

- `spec.podTemplate.spec.serviceAccountName` to use custom RBAC resources.
- `spec.podTemplate.spec.imagePullSecrets` to pull images from a private registry.
- `spec.podTemplate.spec.resources` to set CPU and memory requests/limits.

To learn more:

- [Custom RBAC for Neo4j](/docs/guides/neo4j/custom-rbac/using-custom-rbac.md)
- [Private Registry for Neo4j](/docs/guides/neo4j/private-registry/using-private-registry.md)

### spec.deletionPolicy

`spec.deletionPolicy` controls what KubeDB does when a `Neo4j` object is deleted.

| Policy | What gets deleted |
|--------|-------------------|
| `DoNotTerminate` | Nothing — KubeDB blocks deletion of the `Neo4j` object entirely |
| `Halt` | The database pods and services are removed, but PVCs and Secrets are kept |
| `Delete` | Pods, services, and PVCs are removed, but the auth Secret is kept |
| `WipeOut` | Everything is removed: pods, services, PVCs, and generated Secrets |

Use `WipeOut` for development clusters where you want full cleanup. Use `Halt` or `Delete` in production when you need to preserve data or credentials for recovery.

For more details, see the [deletion policy reference](https://appscode.com/blog/post/deletion-policy/).

## Next Steps

- Learn about [Neo4jVersion CRD](/docs/guides/neo4j/concepts/catalog.md).
- Learn about [AppBinding CRD](/docs/guides/neo4j/concepts/appbinding.md).
- Explore [Neo4j OpsRequest](/docs/guides/neo4j/concepts/opsrequest.md).
- Follow the [Neo4j quickstart](/docs/guides/neo4j/quickstart/quickstart.md)
