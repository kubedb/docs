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

`Weaviate` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [Weaviate](https://weaviate.io/) in a Kubernetes native way. You only need to describe the desired database configuration in a `Weaviate` object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## Weaviate Spec

As with all other Kubernetes objects, a Weaviate needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example `Weaviate` object.

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
  authSecret:
    kind: Secret
    name: weaviate-sample-auth
  configuration:
    secretName: weaviate-custom-config
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: weaviate-issuer
    clientAuth: true
  podTemplate:
    spec:
      containers:
        - name: weaviate
          resources:
            requests:
              cpu: 500m
              memory: 1Gi
            limits:
              cpu: 500m
              memory: 1Gi
  deletionPolicy: WipeOut
  healthChecker:
    periodSeconds: 10
    timeoutSeconds: 10
    failureThreshold: 3
```

### spec.version

`spec.version` is a required field specifying the name of the [WeaviateVersion](/docs/guides/weaviate/concepts/catalog.md) CR where the docker images are specified. Run `kubectl get weaviateversions` to list the versions available in your cluster.

### spec.replicas

`spec.replicas` specifies the number of nodes (pods) in the Weaviate cluster. A multi-node cluster lets Weaviate replicate collections across nodes for high availability.

### spec.replication

`spec.replication.factor` configures the default replication factor for collections (between `1` and `5`). `1` means no replication (default); `2`–`3` are typical for production high availability.

### spec.storageType

`spec.storageType` can be `Durable` or `Ephemeral`. `Durable` uses a `PersistentVolumeClaim` to persist data; `Ephemeral` uses an `emptyDir` that is lost when the pod is deleted.

### spec.storage

If `spec.storageType` is `Durable`, then `spec.storage` is required. It accepts a standard `PersistentVolumeClaimSpec`:

- `spec.storage.storageClassName` — the name of the `StorageClass` used to provision the PVCs.
- `spec.storage.accessModes` — the PVC access modes (e.g. `ReadWriteOnce`).
- `spec.storage.resources` — the storage request for each node.

### spec.disableSecurity

`spec.disableSecurity` (default `false`) disables API-key authentication when set to `true`. When security is enabled (the default), KubeDB generates an API key.

### spec.authSecret

`spec.authSecret` references the Secret holding the Weaviate API-key credentials. If not provided, KubeDB creates one named `<database-name>-auth` containing the standard Weaviate API-key environment variables (`AUTHENTICATION_APIKEY_ENABLED`, `AUTHENTICATION_APIKEY_ALLOWED_KEYS`, and `AUTHENTICATION_APIKEY_USERS`). You can rotate it with a [RotateAuth](/docs/guides/weaviate/rotate-auth/rotate-auth.md) ops request.

### spec.configuration

`spec.configuration` provides a custom Weaviate configuration (the `conf.yaml` file). It can be supplied either through a Secret (`spec.configuration.secretName`) or inline (`spec.configuration.inline`). See [Using Custom Configuration File](/docs/guides/weaviate/configuration/using-config-file.md).

### spec.tls

`spec.tls` enables TLS encryption using [cert-manager](https://cert-manager.io/). It references an `Issuer`/`ClusterIssuer` through `spec.tls.issuerRef`, and `spec.tls.clientAuth` controls whether mutual TLS (client certificate authentication) is required. See [Weaviate TLS](/docs/guides/weaviate/tls/overview.md).

### spec.podTemplate

`spec.podTemplate` lets you customize the pod specification for the Weaviate pods — container resources, security context, scheduling, environment variables, etc.

### spec.serviceTemplates

`spec.serviceTemplates` allows customizing the Kubernetes Services that KubeDB creates for the Weaviate cluster.

### spec.deletionPolicy

`spec.deletionPolicy` controls what happens to the database resources when the `Weaviate` object is deleted. The available options are `DoNotTerminate`, `Halt`, `Delete`, and `WipeOut`. See the [Quickstart](/docs/guides/weaviate/quickstart/quickstart.md#database-deletionpolicy) for details.

### spec.healthChecker

`spec.healthChecker` configures how KubeDB health-checks the database. It has `periodSeconds`, `timeoutSeconds`, and `failureThreshold` fields.

## Next Steps

- Learn how to use KubeDB to run a Weaviate database [here](/docs/guides/weaviate/quickstart/quickstart.md).
- Detail concepts of [WeaviateVersion object](/docs/guides/weaviate/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
