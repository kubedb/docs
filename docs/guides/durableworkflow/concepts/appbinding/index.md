---
title: AppBinding CRD
menu:
  docs_{{ .version }}:
    identifier: dw-appbinding-concepts
    name: AppBinding
    parent: dw-concepts-durableworkflow
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# AppBinding

## What is AppBinding

An `AppBinding` is a Kubernetes `CustomResourceDefinition`(CRD) which points to an application using either its URL (usually for a non-Kubernetes resident service instance) or a Kubernetes service object (if self-hosted in a Kubernetes cluster), some optional parameters and a credential secret. To learn more about AppBinding and the problems it solves, please read this blog post: [The case for AppBinding](https://appscode.com/blog/post/the-case-for-appbinding).

A `DurableWorkflow` uses `AppBinding` objects in two directions.

## The AppBindings that a DurableWorkflow reads

`spec.persistence.default` and `spec.persistence.visibility` do not contain addresses or passwords. They name an `AppBinding` of the database. When you deploy a KubeDB `Postgres`, `MySQL`, `Cassandra`, `Elasticsearch` or `OpenSearch`, KubeDB creates the `AppBinding` with the same name as the database. For a database that KubeDB does not manage, create the `AppBinding` yourself with a `clientConfig` (URL or service) and a `secret` with the admin credentials.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: temporal
  namespace: demo
spec:
  version: temporal-1.32.1
  persistence:
    default:
      name: pg
      namespace: demo
```

The operator uses the address and the admin credentials of the `AppBinding` for two things only: to create the databases and the application user, and to drop them under `WipeOut`. The servers connect with the application user, whose credentials are in a Secret of the `DurableWorkflow` namespace (`<name>-default-auth` and `<name>-visibility-auth`). For Elasticsearch and OpenSearch, the operator creates a role limited to the index and a user with only that role through the security API, and the servers use it. Only when the server has no usable security API do the servers use the account of the `AppBinding`. For OpenSearch, set `type: OpenSearch` on the store. For Cadence's Kafka, the `AppBinding` of the `Kafka` in `spec.persistence.kafkaRef` carries the bootstrap servers in `clientConfig.url` and the SASL PLAIN credentials. See [Advanced Visibility](/docs/guides/durableworkflow/configuration/advanced-visibility.md).

## The AppBinding that a DurableWorkflow publishes

When the cluster is ready, the operator publishes an `AppBinding` with the name of the `DurableWorkflow`. It points to the frontend Service and the gRPC port, so other tools can find the cluster:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: temporal
  namespace: demo
spec:
  appRef:
    apiGroup: kubedb.com
    kind: DurableWorkflow
    name: temporal
    namespace: demo
  clientConfig:
    service:
      name: temporal
      port: 7233
      scheme: grpc
  type: kubedb.com/DurableWorkflow
  version: 1.32.1
```

For Cadence the port is `7833`.

## Next Steps

- Choose the [persistence stores](/docs/guides/durableworkflow/configuration/persistence.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
