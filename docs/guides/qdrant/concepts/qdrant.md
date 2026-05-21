---
title: Qdrant CRD
menu:
  docs_{{ .version }}:
    identifier: qdrant-concepts-qdrant
    name: Qdrant
    parent: qdrant-concepts
    weight: 10
menu_name: docs_{{ .version }}
---

> New to KubeDB? Please start [here](/docs/README.md).

# Qdrant

## What is Qdrant

`Qdrant` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [Qdrant](https://qdrant.tech/) vector databases in a Kubernetes native way. You only need to describe the desired database configuration in a `Qdrant` object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## Qdrant Spec

As with all other Kubernetes objects, a `Qdrant` CR needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

Below is an example `Qdrant` object:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  authSecret:
    name: qdrant-sample-auth
  configSecret:
    name: qdrant-config
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: qdrant-issuer
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          app: kubedb
        interval: 10s
  podTemplate:
    metadata:
      annotations:
        passMe: ToDatabasePod
    spec:
      serviceAccountName: my-custom-sa
      nodeSelector:
        disktype: ssd
      imagePullSecrets:
      - name: myregistrykey
      containers:
      - name: qdrant
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1"
  serviceTemplates:
  - alias: primary
    spec:
      type: LoadBalancer
  deletionPolicy: Halt
```

### spec.version

`spec.version` (required) specifies the name of the [QdrantVersion](/docs/guides/qdrant/concepts/catalog.md) CRD where the docker images are specified.

```bash
$ kubectl get qdrantversions
NAME     VERSION   DB_IMAGE                                       DEPRECATED   AGE
1.15.4   1.15.4    docker.io/qdrant/qdrant:v1.15.4-unprivileged                28d
1.16.2   1.16.2    docker.io/qdrant/qdrant:v1.16.2-unprivileged                28d
1.17.0   1.17.0    docker.io/qdrant/qdrant:v1.17.0-unprivileged                28d
```

### spec.replicas

`spec.replicas` is an optional `<integer>` field that specifies the number of Qdrant pods to run. For a single-node deployment, set it to `1`. For a multi-node distributed cluster, set it to the desired number of replicas (e.g., `3`).

### spec.mode

`spec.mode` specifies the deployment mode for the Qdrant cluster. Supported values are:

- `Standalone` — runs a single Qdrant node.
- `Distributed` — runs a multi-node Qdrant cluster with peer-to-peer communication. Required for `spec.tls.p2p`.

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used for Qdrant API key authentication. If not provided, KubeDB will create one automatically. It contains the following sub-fields:

- `name` — the name of the Secret (required).
- `kind` — the kind of the secret reference (required).
- `apiGroup` — the API group of the secret reference.
- `externallyManaged` — specifies whether the secret is managed externally.
- `activeFrom` — the time from which the secret becomes active.
- `rotateAfter` — the duration after which the secret should be rotated.
- `secretStoreName` — the name of the external secret store.

### spec.configSecret

`spec.configSecret` is an optional field that points to a Secret containing a custom `production.yaml` configuration file for Qdrant. See [Custom Configuration](/docs/guides/qdrant/configuration/using-config-file.md) for details.

### spec.configuration

`spec.configuration` is an optional field for providing custom Qdrant configuration. It has the following sub-fields:

- `inline` — a map of key-value pairs for inline configuration.
- `secretName` — the name of a Secret containing the configuration.

### spec.storageType

`spec.storageType` specifies the type of storage that will be used for Qdrant. Supported values are:

- `Durable` — uses PersistentVolumeClaims (default).
- `Ephemeral` — uses `EmptyDir` volumes (useful for testing).

### spec.storage

`spec.storage` specifies the PersistentVolumeClaim configuration for Qdrant data. It contains standard PVC fields like `storageClassName`, `accessModes`, `resources`, `selector`, etc.

### spec.healthChecker

`spec.healthChecker` specifies the configuration for database health checking. It contains the following sub-fields:

- `disableWriteCheck` — disables write health checks.
- `failureThreshold` — the number of consecutive failures before marking the database as unhealthy.
- `periodSeconds` — the interval between health checks.
- `timeoutSeconds` — the timeout for each health check.

### spec.halted

`spec.halted` is an optional `<boolean>` field. When set to `true`, the database will be halted (all pods stopped) while preserving the PVCs and other resources.

### spec.disableSecurity

`spec.disableSecurity` is an optional `<boolean>` field that disables API key authentication when set to `true`. The default is `false`.

### spec.tls

`spec.tls` specifies the TLS configurations for the Qdrant database. KubeDB uses [cert-manager](https://cert-manager.io/) v1 api to provision and manage TLS certificates.

The following fields are configurable in the `spec.tls` section:

- `issuerRef` is a reference to the `Issuer` or `ClusterIssuer` CR of [cert-manager](https://cert-manager.io/docs/concepts/issuer/) that will be used by `KubeDB` to generate necessary certificates.

  - `apiGroup` is the group name of the resource that is being referenced. Currently, the only supported value is `cert-manager.io`.
  - `kind` — the type of resource. KubeDB supports both `Issuer` and `ClusterIssuer`.
  - `name` — the name of the resource being referenced (required).

- `client` (optional, `<boolean>`, default `false`) enables TLS for client-to-server communication. When set to `true`, the Qdrant server will accept TLS-encrypted connections from clients.

- `p2p` (optional, `<boolean>`, default `false`) enables TLS for peer-to-peer communication between Qdrant nodes. When set to `true`, inter-node communication within the Qdrant cluster will be encrypted using TLS. Requires `spec.mode` to be `Distributed`.

- `certificates` (optional) is a list of additional certificates used to configure the Qdrant server. Each certificate has the following fields:
  - `alias` — the identifier of the certificate (required). The supported value is `server`.
  - `secretName` (optional) specifies the Kubernetes secret name that holds the certificates. If not specified, defaults to `<database-name>-<cert-alias>-cert`.
  - `issuerRef` (optional) specifies a separate issuer for this certificate. If not set, the top-level `issuerRef` is used.
    - `apiGroup` — the API group of the issuer.
    - `kind` — the kind of the issuer (`Issuer` or `ClusterIssuer`).
    - `name` — the name of the issuer.
  - `subject` (optional) specifies an `X.509` distinguished name with the following sub-fields:
    - `organizations` — list of organization names.
    - `organizationalUnits` — list of organization unit names.
    - `countries` — list of country names.
    - `localities` — list of locality names.
    - `provinces` — list of province names.
    - `streetAddresses` — list of street addresses.
    - `postalCodes` — list of postal codes.
    - `serialNumber` — serial number.
  - `duration` (optional) — the validity period of the certificate.
  - `renewBefore` (optional) — the time before expiration to renew the certificate.
  - `dnsNames` (optional) — list of DNS subject alternative names.
  - `ipAddresses` (optional) — list of IP subject alternative names.
  - `uris` (optional) — list of URI Subject Alternative Names.
  - `emailAddresses` (optional) — list of email Subject Alternative Names.
  - `privateKey` (optional) specifies options for the private key:
    - `encoding` — the private key encoding. Supported values are `PKCS1` and `PKCS8`.

### spec.monitor

`spec.monitor` specifies the monitoring configuration for Qdrant. It contains:

- `agent` — the monitoring agent. Supported values are `prometheus.io/builtin`, `prometheus.io`, and `prometheus.io/operator`.
- `prometheus` — Prometheus-specific configuration including `exporter` and `serviceMonitor` settings.

### spec.podTemplate

KubeDB allows providing a template for database pods through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for Qdrant database. Notable sub-fields include:

- `spec.podTemplate.spec.serviceAccountName` — provide a custom ServiceAccount.
- `spec.podTemplate.spec.imagePullSecrets` — pull images from a private registry.
- `spec.podTemplate.spec.nodeSelector` — schedule pods on specific nodes.
- `spec.podTemplate.spec.containers[].resources` — configure CPU and memory resources.
- `spec.podTemplate.spec.containers[].env` — set environment variables for the container.

### spec.serviceTemplates

`spec.serviceTemplates` is an optional list of service templates. Each entry has:

- `alias` — the service alias (required). Supported values include `primary`, `standby`, `stats`, `dashboard`, etc.
- `spec` — the service specification including `type`, `clusterIP`, `ports`, etc.

### spec.deletionPolicy

`spec.deletionPolicy` controls the behavior when a Qdrant object is deleted. Supported values are:

- `Halt` — deletes the Qdrant object but keeps underlying resources (PVCs, Secrets).
- `Delete` — deletes the Qdrant object and its PVCs, but not Secrets.
- `WipeOut` — deletes the Qdrant object and all related resources including PVCs and Secrets.
- `DoNotTerminate` — prevents deletion if admission webhook is enabled.

## Next Steps

- Learn about [QdrantVersion CRD](/docs/guides/qdrant/concepts/catalog.md).
- Deploy your first Qdrant database with KubeDB by following the guide [here](/docs/guides/qdrant/quickstart/quickstart.md).
