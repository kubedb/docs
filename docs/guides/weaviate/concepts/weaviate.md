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

`Weaviate` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [Weaviate](https://weaviate.io/) vector databases in a Kubernetes native way. You only need to describe the desired database configuration in a `Weaviate` object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## Weaviate Spec

As with all other Kubernetes objects, a `Weaviate` CR needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

Below is an example `Weaviate` object:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-sample
  namespace: demo
spec:
  version: "1.33.1"
  replicas: 3
  authSecret:
    name: weaviate-sample-auth
  configuration:
    secretName: weaviate-config
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
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
      - name: weaviate
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

`spec.version` is a required field that specifies the name of the [WeaviateVersion](/docs/guides/weaviate/concepts/catalog.md) CRD where the docker images are specified. Currently, when you install KubeDB, it creates the following `WeaviateVersion` CRDs:

```bash
$ kubectl get weaviateversions
NAME      VERSION   DB_IMAGE                        DEPRECATED   AGE
1.25.0    1.25.0    kubedb/weaviate:1.25.0                       3d
1.28.0    1.28.0    kubedb/weaviate:1.28.0                       3d
1.30.0    1.30.0    kubedb/weaviate:1.30.0                       3d
1.33.1    1.33.1    kubedb/weaviate:1.33.1                       3d
```

### spec.replicas

`spec.replicas` is an optional field that specifies the number of Weaviate pods to run. For a single-node deployment, set it to `1`. For a multi-node cluster, set it to the desired number of replicas (e.g., `3`).

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used for Weaviate API key authentication. If not provided, KubeDB will create one automatically. The Secret must contain an `api-key` data field that is injected into pods via the `AUTHENTICATION_APIKEY_ALLOWED_KEYS` environment variable.

### spec.configuration.secretName

`spec.configuration.secretName` is an optional field that points to a Secret containing a custom `weaviate.yaml` configuration file for Weaviate. See [Custom Configuration](/docs/guides/weaviate/configuration/using-config-file.md) for details.

### spec.disableSecurity

`spec.disableSecurity` is an optional boolean field that disables API key authentication when set to `true`. When `false` (the default), KubeDB sets `AUTHENTICATION_APIKEY_ENABLED=true` automatically.

### spec.storageType

`spec.storageType` specifies the type of storage that will be used for Weaviate. It can be `Durable` or `Ephemeral`. The default value is `Durable`. If `Ephemeral` is used, KubeDB will create Weaviate using `EmptyDir` volume. This is useful for testing purposes.

### spec.storage

`spec.storage` specifies the StorageClass of PVCs that will be dynamically allocated to store data for Weaviate pods. If `spec.storageType: Ephemeral` is not set, this field is required.

### spec.monitor

`spec.monitor` specifies the monitoring configuration for Weaviate. It contains the following sub-field:

- `spec.monitor.agent` specifies the monitoring agent. Valid values are `prometheus.io/builtin` and `prometheus.io/operator`.

### spec.podTemplate

KubeDB allows providing a template for database pods through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the StatefulSet created for Weaviate database. Notable sub-fields include:

- `spec.podTemplate.spec.serviceAccountName` to provide a custom ServiceAccount.
- `spec.podTemplate.spec.imagePullSecrets` to pull images from a private registry.
- `spec.podTemplate.spec.nodeSelector` to schedule pods on specific nodes.
- `spec.podTemplate.spec.containers[].resources` to configure CPU and memory resources.

### spec.serviceTemplates

`spec.serviceTemplates` is an optional field that contains a list of the service templates for the Weaviate services. KubeDB allows following service template variables:

- `spec.serviceTemplates[].alias`: specifies which service template (e.g., `primary`).
- `spec.serviceTemplates[].spec.type`: specifies the service type (e.g., `ClusterIP`, `LoadBalancer`, `NodePort`).

### spec.deletionPolicy

`spec.deletionPolicy` gives freedom to the user to control the behavior of KubeDB when a Weaviate object is deleted. Possible values are:

- `DoNotTerminate`: prevents deletion of the object if admission webhook is enabled.
- `Halt`: deletes the Weaviate object but keeps the underlying resources (PVCs, Secrets) intact.
- `Delete`: deletes the Weaviate object and its PVCs, but not Secrets.
- `WipeOut`: deletes the Weaviate object and all related resources including PVCs and Secrets.

## Next Steps

- Learn about [WeaviateVersion CRD](/docs/guides/weaviate/concepts/catalog.md).
- Deploy your first Weaviate database with KubeDB by following the guide [here](/docs/guides/weaviate/quickstart/quickstart.md).