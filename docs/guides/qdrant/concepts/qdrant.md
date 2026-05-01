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

`spec.version` is a required field that specifies the name of the [QdrantVersion](/docs/guides/qdrant/concepts/catalog.md) CRD where the docker images are specified. Currently, when you install KubeDB, it creates the following `QdrantVersion` CRDs:

```bash
$ kubectl get qdrantversions
NAME      VERSION   DB_IMAGE                    DEPRECATED   AGE
1.7.4     1.7.4     qdrant/qdrant:v1.7.4                     3d
1.10.0    1.10.0    qdrant/qdrant:v1.10.0                    3d
1.14.0    1.14.0    qdrant/qdrant:v1.14.0                    3d
1.17.0    1.17.0    qdrant/qdrant:v1.17.0                    3d
```

### spec.replicas

`spec.replicas` is an optional field that specifies the number of Qdrant pods to run. For a single-node deployment, set it to `1`. For a multi-node cluster, set it to the desired number of replicas (e.g., `3`).

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used for Qdrant API key authentication. If not provided, KubeDB will create one automatically. The Secret must contain an `api-key` data field.

### spec.configSecret

`spec.configSecret` is an optional field that points to a Secret containing a custom `production.yaml` configuration file for Qdrant. See [Custom Configuration](/docs/guides/qdrant/configuration/using-config-file.md) for details.

### spec.storageType

`spec.storageType` specifies the type of storage that will be used for Qdrant. It can be `Durable` or `Ephemeral`. The default value is `Durable`. If `Ephemeral` is used, KubeDB will create Qdrant using `EmptyDir` volume. In this case, you don't have to specify `spec.storage`. This is useful for testing purposes.

### spec.storage

`spec.storage` specifies the StorageClass of PVCs that will be dynamically allocated to store data for Qdrant pods. If `spec.storageType: Ephemeral` is not set, this field is required.

### spec.tls

`spec.tls` specifies TLS/SSL configurations for Qdrant. It contains the following sub-fields:

- `spec.tls.issuerRef` points to a cert-manager `Issuer` or `ClusterIssuer` used to issue certificates.
- `spec.tls.certificates` is an optional field that lists additional certificates for the Qdrant server.

### spec.monitor

`spec.monitor` specifies the monitoring configuration for Qdrant. It contains the following sub-field:

- `spec.monitor.agent` specifies the monitoring agent. Valid values are `prometheus.io/builtin` and `prometheus.io/operator`.

### spec.podTemplate

KubeDB allows providing a template for database pods through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the StatefulSet created for Qdrant database. Notable sub-fields include:

- `spec.podTemplate.spec.serviceAccountName` to provide a custom ServiceAccount.
- `spec.podTemplate.spec.imagePullSecrets` to pull images from a private registry.
- `spec.podTemplate.spec.nodeSelector` to schedule pods on specific nodes.
- `spec.podTemplate.spec.containers[].resources` to configure CPU and memory resources.

### spec.serviceTemplates

`spec.serviceTemplates` is an optional field that contains a list of the service templates for the Qdrant services. KubeDB allows following service template variables:

- `spec.serviceTemplates[].alias`: specifies which service template (e.g., `primary`).
- `spec.serviceTemplates[].spec.type`: specifies the service type (e.g., `ClusterIP`, `LoadBalancer`, `NodePort`).

### spec.deletionPolicy

`spec.deletionPolicy` gives freedom to the user to control the behavior of KubeDB when a Qdrant object is deleted. Possible values are:

- `DoNotTerminate`: prevents deletion of the object if admission webhook is enabled.
- `Halt`: deletes the Qdrant object but keeps the underlying resources (PVCs, Secrets) intact.
- `Delete`: deletes the Qdrant object and its PVCs, but not Secrets.
- `WipeOut`: deletes the Qdrant object and all related resources including PVCs and Secrets.

### spec.disableSecurity

`spec.disableSecurity` is an optional boolean field that disables API key authentication when set to `true`. The default is `false`.

## Next Steps

- Learn about [QdrantVersion CRD](/docs/guides/qdrant/concepts/catalog.md).
- Deploy your first Qdrant database with KubeDB by following the guide [here](/docs/guides/qdrant/quickstart/quickstart.md).