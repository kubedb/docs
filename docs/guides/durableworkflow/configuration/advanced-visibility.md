---
title: Advanced Visibility
menu:
  docs_{{ .version }}:
    identifier: dw-advanced-visibility-configuration
    name: Advanced Visibility
    parent: dw-configuration
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Advanced Visibility with Elasticsearch or OpenSearch

Visibility records (the list and the search of workflows) can live in Postgres or MySQL, or in an Elasticsearch or OpenSearch index. This page covers the search store for both engines: Temporal writes to the index directly, and Cadence needs Kafka in addition.

{{< notice type="warning" message="This is implemented but has not been verified on a real Elasticsearch, OpenSearch or Kafka. The behavior below is taken from the operator code and its unit tests." >}}

## Before You Begin

- A `Ready` KubeDB `Elasticsearch` or `OpenSearch`, with its `AppBinding`. For Cadence also a `Ready` KubeDB `Kafka` with its `AppBinding`.
- A default store, for example the Postgres `pg` of the [quickstart](/docs/guides/durableworkflow/quickstart/prerequisites.md).
- You should be familiar with [Persistence](/docs/guides/durableworkflow/configuration/persistence.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Temporal

Name the `AppBinding` in `spec.persistence.visibility`. See the example in [Persistence](/docs/guides/durableworkflow/configuration/persistence.md#elasticsearch-and-opensearch-visibility-temporal). The operator creates the index and the application user described below. `kafkaRef` is rejected for Temporal.

## Cadence

Cadence needs both a search store and Kafka. The history service writes the visibility records to a Kafka topic, and the indexer in the worker service reads the topic and writes the index.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: cadence-es
  namespace: demo
spec:
  version: cadence-1.4.2
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
    visibility:
      name: es
      namespace: demo
      type: Elasticsearch
    kafkaRef:
      name: kafka
      namespace: demo
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/persistence/cadence-es-visibility.yaml
```

- `kafkaRef` is required for Cadence with a search visibility store, and rejected otherwise (for Temporal, and for Cadence with a Postgres or MySQL visibility store).
- A Cadence index name must start with `cadence-visibility-`. If you do not set `databaseName`, it is `cadence-visibility-<namespace>-<name>`, in lower case.
- **OpenSearch**: the `AppBinding` of a KubeDB `OpenSearch` needs `type: OpenSearch` in the visibility store, as in [this example](https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/persistence/cadence-opensearch-visibility.yaml).

### What the operator does

1. **The index template and the index.** Before the servers start, the operator puts the Cadence index template for the index (`_template/<index>` on Elasticsearch, `_index_template/<index>` on OpenSearch), creates the index if it is missing, and puts the mappings on it. The template is scoped to this one index, so several clusters can share a search server. It repeats this for a new server version.
2. **The Kafka topics.** The operator creates two topics on the Kafka of `kafkaRef`: a topic named after the index, with 3 partitions, and a dead letter queue `<index>-dlq` with 1 partition. The replication factor is the number of brokers, up to 3. They are named after the index, so two clusters on one Kafka never share a topic or a consumer group.
3. **Credentials for the pods.** The Kafka credentials and the CA files go into the Secret `<name>-visibility-client`.

Kafka is reached with the `AppBinding`: the brokers come from `clientConfig.url` (comma separated), the credentials from its Secret and the CA from its `caBundle` or TLS Secret. Only **SASL PLAIN** is supported.

Under `deletionPolicy: WipeOut`, the operator drops the index, the template (Cadence), the application user and its role, and the topics. With the other policies they are kept. The topics are disposable: a Kafka that cannot be reached does not hold up the deletion, and a Warning event `KafkaTopicsNotDeleted` says the topics were left.

## The per-index search user

The servers do not use the admin account of the `AppBinding`. For Elasticsearch (security API) and OpenSearch (security plugin), the operator creates:

- a role limited to the one index, with all privileges on it and read-only monitoring on the cluster;
- a user with only that role, named `dw_<namespace>_<name>_visibility` (dashes and dots become underscores), with a generated password.

The credentials are in the Secret `<name>-visibility-auth`, which is annotated `kubedb.com/search-user: managed`. The schema Job still uses the admin account, kept in a copy `<name>-visibility-auth-admin`, because creating the index is beyond what the application user may do.

[RotateAuth](/docs/guides/durableworkflow/rotateauth/overview.md) rotates this user as well, one user next to the other, before it removes the old one.

**Fallback.** If the search server has no usable security API (security disabled, or no such plugin), the operator cannot create the user. The servers then use a copy of the account of the `AppBinding` (the Secret is annotated `kubedb.com/search-user: admin`), and a Warning event `SearchUserNotCreated` on the `DurableWorkflow` says so. If you set `createDatabase: false` and your own `authSecret`, the operator leaves the user to you.

## Next Steps

- [Persistence](/docs/guides/durableworkflow/configuration/persistence.md) for the other stores.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
