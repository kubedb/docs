---
title: KubeDB Common Configuration
description: Enable database engines and verify the KubeDB installation
menu:
  docs_{{ .version }}:
    identifier: install-kubedb-config
    name: Common Configuration
    parent: install-kubedb-enterprise
    weight: 60
product_name: kubedb
menu_name: docs_{{ .version }}
section_menu_id: setup
---

# Common Configuration

The steps below apply regardless of which [installation method](/docs/setup/install/kubedb/_index.md) you used.

## Enable Database Engines

KubeDB ships support for many database engines, gated behind individual feature flags so the operator only installs the components you actually need. Toggle an engine on by setting its `global.featureGates.<Engine>` value to `true`. The defaults below mirror the upstream chart — `Elasticsearch`, `Kafka`, `MariaDB`, `MongoDB`, `MySQL`, `Postgres`, and `Redis` are enabled out of the box; every other engine is disabled.

```yaml
global:
  featureGates:
    Cassandra: false
    ClickHouse: false
    DB2: false
    DocumentDB: false
    Druid: false
    Elasticsearch: true
    HanaDB: false
    Hazelcast: false
    Ignite: false
    Kafka: true
    MariaDB: true
    Memcached: false
    Milvus: false
    MongoDB: true
    MSSQLServer: false
    MySQL: true
    Neo4j: false
    Oracle: false
    PerconaXtraDB: false
    PgBouncer: false
    Pgpool: false
    Postgres: true
    ProxySQL: false
    Qdrant: false
    RabbitMQ: false
    Redis: true
    Singlestore: false
    Solr: false
    Weaviate: false
    ZooKeeper: false
```

Save these values to a file (e.g. `values.yaml`) and pass it to `helm install` / `helm upgrade`:

```bash
$ helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --values values.yaml \
  --wait --burst-limit=10000 --debug
```

Or override individual engines inline with `--set`:

```bash
$ helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --set global.featureGates.Cassandra=true \
  --set global.featureGates.ClickHouse=true \
  --wait --burst-limit=10000 --debug
```

The same `global.featureGates` map works with the ArgoCD `Application` manifests under the `spec.source.helm.values` block, with the `kubedb-certified` chart on OpenShift, and with the `Kubedb` installer CR used by the OperatorHub bundle.

## Verify installation

To check if KubeDB operator pods have started, run the following command:

```bash
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"

NAME                                            READY   STATUS    RESTARTS   AGE
kubedb-kubedb-autoscaler-b5dd47dc5-bxnrq        1/1     Running   0          48s
kubedb-kubedb-ops-manager-6f766b86c6-h9m66      1/1     Running   0          48s
kubedb-kubedb-provisioner-6fd44d5784-d8v9c      1/1     Running   0          48s
kubedb-kubedb-webhook-server-6cf469bdf4-72wvz   1/1     Running   0          48s
```

Once the operator pod is running, you can cancel the above command by typing `Ctrl+C`.

Now, to confirm CRD groups have been registered by the operator, run the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
```

Now, you are ready to [create your first database](/docs/guides/README.md) using KubeDB.
