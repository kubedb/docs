---
title: Prepare Dependencies
menu:
  docs_{{ .version }}:
    identifier: dw-prerequisites-quickstart
    name: Prepare Dependencies
    parent: dw-quickstart-durableworkflow
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Prepare DurableWorkflow Dependencies

A `DurableWorkflow` keeps no state of its own, so it cannot start without a database. This page sets up the one every guide in this section uses: a KubeDB `Postgres` named `pg` in the `demo` namespace. KubeDB MySQL works the same way, and so do Cassandra and an external database, see [Persistence](/docs/guides/durableworkflow/configuration/persistence.md).

## Before You Begin

- You need a Kubernetes cluster and a properly configured `kubectl` command-line tool. If you do not have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install the KubeDB operator in your cluster by following the [installation steps](/docs/setup/README.md) and enable the `DurableWorkflow` feature gate:

  ```bash
  helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
    --namespace kubedb --create-namespace \
    --set global.featureGates.DurableWorkflow=true
  ```

- Install [cert-manager](https://cert-manager.io/docs/installation/) only if you want to try TLS.

- This tutorial uses a separate namespace called `demo`:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Check the Versions

```bash
$ kubectl get durableworkflowversions
NAME              DISTRIBUTION   VERSION   SERVER_IMAGE                         DEPRECATED   AGE
cadence-1.4.2     Cadence        1.4.2     docker.io/ubercadence/server:v1.4.2               5m
temporal-1.32.1   Temporal       1.32.1    docker.io/temporalio/server:1.32.1                5m
```

## Deploy Postgres

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg
  namespace: demo
spec:
  version: "18.3"
  replicas: 3
  configuration:
    tuning:
      profile: web
      maxConnections: 300
      storageType: ssd
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/quickstart/pg.yaml
postgres.kubedb.com/pg created
```

Wait until it is `Ready`. KubeDB creates an `AppBinding` named `pg` in the same namespace. That `AppBinding` is what a `DurableWorkflow` refers to:

```bash
$ kubectl get postgres,appbinding -n demo
```

## The Connection Budget

Every server pod keeps up to **5 connections per store** (`maxConns` in `server.yaml`). Each store is the default store and the visibility store, so a pod can use 10 connections when they are both SQL. A cluster with 2 replicas for each of the 4 roles has 8 pods and can open 80 connections, and a rolling operation can add one surge pod per role on top of that.

The default `max_connections` of a Postgres is **100**, which several highly available clusters exceed together. The Postgres above therefore sets `maxConnections: 300`. Plan the limit of the database like this:

```text
max_connections  >=  sum over all DurableWorkflows of (pods + roles)  x  stores  x  5
```

The provisioner records a `ConnectionBudget` warning event on the `DurableWorkflow` when the possible demand of the cluster exceeds 80% of the `max_connections` it reads from Postgres or MySQL. If you see pods fail with `no usable database connection found` or `too many clients`, either raise `max_connections` of the database or lower `maxConns` in `server.yaml`, see [Custom Configuration](/docs/guides/durableworkflow/configuration/custom-config.md).

## Next Steps

- Run [Temporal](/docs/guides/durableworkflow/quickstart/temporal.md).
- Run [Cadence](/docs/guides/durableworkflow/quickstart/cadence.md).
