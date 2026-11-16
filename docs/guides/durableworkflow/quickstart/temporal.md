---
title: Temporal Quickstart
menu:
  docs_{{ .version }}:
    identifier: dw-temporal-quickstart
    name: Temporal
    parent: dw-quickstart-durableworkflow
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Temporal Quickstart

This tutorial shows how to run a Temporal cluster with KubeDB and start a workflow in it.

## Before You Begin

- Complete [Prepare Dependencies](/docs/guides/durableworkflow/quickstart/prerequisites.md). You need the `demo` namespace and a `Ready` Postgres named `pg` with its `AppBinding`.

- You should be familiar with the following `KubeDB` concepts:
  - [DurableWorkflow](/docs/guides/durableworkflow/concepts/durableworkflow/index.md)
  - [DurableWorkflowVersion](/docs/guides/durableworkflow/concepts/catalog/index.md)
  - [AppBinding](/docs/guides/durableworkflow/concepts/appbinding/index.md)

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create a Temporal Cluster

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: temporal
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
  deletionPolicy: WipeOut
```

Here,

- `spec.version` selects the engine. `temporal-1.32.1` is a Temporal `DurableWorkflowVersion`.
- `spec.topology` sets the replicas of each role. Two replicas per role make the cluster highly available, which is what lets day-2 operations run without downtime. A role you leave out gets one replica.
- `spec.persistence.default` refers to the `AppBinding` of the database. The visibility store is not set, so it uses the same Postgres server with its own database.
- `spec.deletionPolicy: WipeOut` also drops the databases and the database user that the operator created, when you delete the object.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/quickstart/temporal.yaml
durableworkflow.kubedb.com/temporal created
```

### What happens

1. The operator waits for the `AppBinding` `pg`.
2. With the admin credentials of the `AppBinding`, it creates the databases `temporal` and `temporal_visibility` and the database user `dw_temporal`, and stores the password in the Secret `temporal-default-auth`.
3. A schema Job applies the Temporal schema to both databases.
4. Only then are the PetSets created, one per role.
5. The `AppBinding` `temporal` is published with the address of the frontend Service.

```bash
$ kubectl get durableworkflow -n demo
NAME       VERSION           STATUS   AGE
temporal   temporal-1.32.1   Ready    3m

$ kubectl get pods -n demo -l app.kubernetes.io/instance=temporal
temporal-frontend-0   1/1   Running
temporal-frontend-1   1/1   Running
temporal-history-0    1/1   Running
temporal-history-1    1/1   Running
temporal-matching-0   1/1   Running
temporal-matching-1   1/1   Running
temporal-worker-0     1/1   Running
temporal-worker-1     1/1   Running
```

`dw` is the short name of the resource.

The frontend is exposed by the Service `temporal` on port `7233` (gRPC) and `7243` (HTTP). Each role also has a headless Service named `temporal-<role>-pods` and the pods expose Prometheus metrics on port `9090`.

## Use the Cluster

Enable the admin tools pod, which has the Temporal CLI with `TEMPORAL_ADDRESS` already pointing to the frontend:

```bash
$ kubectl patch dw -n demo temporal --type merge -p '{"spec":{"adminTools":{"enabled":true}}}'
durableworkflow.kubedb.com/temporal patched
```

Create a Temporal namespace with a [WorkflowNamespace](/docs/guides/durableworkflow/concepts/workflownamespace/index.md) object:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: WorkflowNamespace
metadata:
  name: default
  namespace: demo
spec:
  databaseRef:
    name: temporal
  description: Default namespace
  retention: 72h
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/quickstart/temporal-namespace.yaml
workflownamespace.kubedb.com/default created

$ kubectl get wfns -n demo
NAME      DATABASE   NAME      RETENTION   STATUS   AGE
default   temporal   default   72h         Ready    15s
```

The operator waits for the cluster to be `Ready` and registers the namespace. See [Manage Namespaces and Domains](/docs/guides/durableworkflow/namespaces/workflownamespace.md) to update the retention, enable archival or delete it.

Alternatively, create the namespace by hand with the CLI instead of the object (do not do both for the same name):

```bash
$ kubectl exec -n demo temporal-admin-tools-0 -- temporal operator namespace create --namespace default --retention 72h
```

Then start a workflow (a worker is needed to run it):

```bash
$ kubectl exec -n demo temporal-admin-tools-0 -- temporal workflow start --namespace default --task-queue hello --type HelloWorkflow --workflow-id hello-1
$ kubectl exec -n demo temporal-admin-tools-0 -- temporal workflow list --namespace default
```

Applications connect to `temporal.demo.svc:7233` with any Temporal SDK. To use the web UI, see [UI and Admin Tools](/docs/guides/durableworkflow/configuration/ui-and-admin-tools.md).

## Cleaning up

```bash
kubectl delete dw -n demo temporal
kubectl delete postgres -n demo pg
kubectl delete ns demo
```

`WipeOut` has removed the `temporal` and `temporal_visibility` databases and the `dw_temporal` user before Postgres goes away.

## Next Steps

- Run [Cadence](/docs/guides/durableworkflow/quickstart/cadence.md).
- Understand the [roles and high availability](/docs/guides/durableworkflow/clustering/roles-and-ha.md).
- Use [MySQL, Cassandra or an existing database](/docs/guides/durableworkflow/configuration/persistence.md), or let the operator create the database with [managed persistence](/docs/guides/durableworkflow/configuration/managed-persistence.md).
- Manage [namespaces and domains](/docs/guides/durableworkflow/namespaces/workflownamespace.md) and [archival](/docs/guides/durableworkflow/configuration/archival.md).
- Secure it with [TLS](/docs/guides/durableworkflow/tls/tls.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
