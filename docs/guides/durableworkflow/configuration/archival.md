---
title: Archival
menu:
  docs_{{ .version }}:
    identifier: dw-archival-configuration
    name: Archival
    parent: dw-configuration
    weight: 35
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Archival

Archival copies the history and the visibility records of closed workflows to an object store, so that they outlive the retention of the namespace (or domain). `spec.archival` of a `DurableWorkflow` configures where, for Temporal and Cadence alike.

{{< notice type="warning" message="Archival is implemented but has not been verified against a real bucket. What was verified on a kind cluster is enabling archival on a WorkflowNamespace." >}}

## Before You Begin

- A `Ready` cluster, as in the [quickstart](/docs/guides/durableworkflow/quickstart/prerequisites.md).
- You should be familiar with [WorkflowNamespace](/docs/guides/durableworkflow/concepts/workflownamespace/index.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## spec.archival

`spec.archival` has two targets, `history` and `visibility`. A target is enabled by being set. Both have:

| Field          | Description                                                                                              |
|----------------|----------------------------------------------------------------------------------------------------------|
| `provider`     | `s3`, `gcs` or `filestore`.                                                                               |
| `uri`          | `s3://bucket/prefix`, `gs://bucket/prefix` or `file:///absolute/path`. The scheme must match the provider and a bucket is required. |
| `configSecret` | The credentials, in a Secret in the namespace of the `DurableWorkflow`. Required for `s3` and `gcs`, not allowed for `filestore`. |

The operator renders the static `archival` block and the defaults for the namespaces (Temporal) or domains (Cadence) that are created afterwards. These are static configuration, so **a change of `spec.archival` rolls the pods**. So does a change of the data of a `configSecret`.

## S3

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: archival-s3
  namespace: demo
type: Opaque
stringData:
  AWS_REGION: us-east-1
  AWS_ACCESS_KEY_ID: <access key id>
  AWS_SECRET_ACCESS_KEY: <secret access key>
---
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: temporal-archival
  namespace: demo
spec:
  version: temporal-1.32.1
  topology:
    frontend:
      replicas: 2
    history:
      replicas: 2
    matching:
      replicas: 2
    worker:
      replicas: 2
  persistence:
    default:
      name: pg
      namespace: demo
  archival:
    history:
      provider: s3
      uri: s3://my-archive/history
      configSecret:
        name: archival-s3
    visibility:
      provider: s3
      uri: s3://my-archive/visibility
      configSecret:
        name: archival-s3
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/archival/s3.yaml
```

The keys of the Secret:

| Key                       | Description                                                                              |
|---------------------------|------------------------------------------------------------------------------------------|
| `AWS_REGION`              | Required.                                                                                |
| `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` | Both or none. They are passed to the servers as environment variables. Leave them out to use a node role or a workload identity. |
| `AWS_ENDPOINT_URL`        | Optional. An S3-compatible endpoint, such as MinIO.                                       |
| `AWS_S3_FORCE_PATH_STYLE` | Optional. `"true"` for path style addressing.                                             |

`history` and `visibility` may share one S3 Secret, but cannot use two different ones: the credentials are environment variables of the whole process, and the webhook rejects the second Secret.

## GCS

The Secret holds the service account key in the key `credentials.json`. It is mounted read-only into the pods and set as the credentials path:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/archival/gcs.yaml
```

The example uses `provider: gcs` with `uri: gs://my-archive/history` for Cadence.

{{< notice type="warning" message="A Temporal server must be built with the temporal_with_gstorage tag to archive to GCS. The stock image refuses to start with a gcs target. The webhook warns about it when you use gcs with a Temporal version." >}}

## Filestore

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/archival/filestore.yaml
```

The path is an `emptyDir` that the operator mounts into every pod of the cluster. It is **not durable and not shared** between the pods: the archives are lost with the pod, and cannot be read through another pod. Use it for tests only. The webhook warns about it. The two filestore paths may be equal (one mount), but they cannot be nested. A `filestore` target has no `configSecret`.

## Enable archival on a namespace

The settings above are what the servers allow. Whether a namespace archives is a setting of the namespace. Namespaces that exist already keep their own archival state. With a [WorkflowNamespace](/docs/guides/durableworkflow/namespaces/workflownamespace.md), set `spec.archival.enabled: true`, and the operator enables history and visibility archival on it, with the URIs of `spec.archival`:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/namespaces/archival-namespace.yaml
```

For a namespace you manage with the CLI, enable it there. The engines fix the URI once it is set.

## Next Steps

- [Manage namespaces and domains](/docs/guides/durableworkflow/namespaces/workflownamespace.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
