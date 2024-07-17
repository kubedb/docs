---
title: Synchronous Replication
menu:
    docs_{{ .version }}:
        identifier: guides-postgres-synchronous
        name: Synchronous Replication Postgres
        parent: pg-postgres-guides
        weight: 42
menu_name: docs_{{ .version }}
section_menu_id: guides
---
> New to KubeDB? Please start [here](/docs/README.md).

# Run as Synchronous Replication Cluster

KubeDB supports Synchronous Replication for PostgreSQL Cluster. This tutorial will show you how to use KubeDB to run PostgreSQL database with Replication Mode as Synchronous.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

## Configure Synchronous Replication Cluster
To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/postgres](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/postgres) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).


Now, create Postgres crd specifying `spec.streamingMode` with `Synchronous` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/synchronous/postgres.yaml
postgres.kubedb.com/demo-pg created
```

Below is the YAML for the Postgres crd we just created.

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: demo-pg
  namespace: demo
spec:
  version: "13.13"
  replicas: 3
  standbyMode: Hot
  streamingMode: Synchronous
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: DoNotTerminate
```

By default, KubeDB create a Synchronous Replication where one Replica Postgres server out of all the replicas will be in `sync` with Current `primary`. 
And others are `potential` candidate to be in sync with primary if the `synchronous` replica failed in any case. 

Let's check in the postgres cluster that we have deployed. Now, exec into the current primary, in our case it is Pod `demo-pg-0`.
```bash
$ kubectl exec -it -n demo demo-pg-0 -c postgres  -- bash
bash-5.1$ psql
psql (14.2)
Type "help" for help.

postgres=# select application_name, client_addr, state, sent_lsn, write_lsn, flush_lsn, replay_lsn, sync_state from pg_stat_replication;
 application_name | client_addr |   state   | sent_lsn  | write_lsn | flush_lsn | replay_lsn | sync_state 
------------------+-------------+-----------+-----------+-----------+-----------+------------+------------
 demo-pg-1        | 10.244.0.22 | streaming | 0/5000060 | 0/5000060 | 0/5000060 | 0/5000060  | sync
 demo-pg-2        | 10.244.0.24 | streaming | 0/5000060 | 0/5000060 | 0/5000060 | 0/5000060  | potential

```
But Users can also configure a Synchronous replication cluster where all the replica are in `sync` with current primary. 
Let's see how a user can do so, Users need to provide `custom configuration` with setting the config for `synchronous_standby_names`. 

For example, If there are 3 nodes in a Postgres cluster where 1 node is a primary and other 2 are acting as replicas. 
In this scenario, We can set all the 2 replicas server as synchronous replica with the current primary. 
We need to provide `synchronous_standby_names = 'FIRST 2 (*)'` inside custom configuration.
That`s all, Then you can see that all the replicas are configured as synchronous replica.
```bash
$ kubectl exec -it -n demo demo-pg-0 -c postgres  -- bash
bash-5.1$ psql
psql (14.2)
Type "help" for help.

postgres=# select application_name, client_addr, state, sent_lsn, write_lsn, flush_lsn, replay_lsn, sync_state from pg_stat_replication;
 application_name | client_addr |   state   | sent_lsn  | write_lsn | flush_lsn | replay_lsn | sync_state 
------------------+-------------+-----------+-----------+-----------+-----------+------------+------------
 demo-pg-1        | 10.244.0.22 | streaming | 0/5000060 | 0/5000060 | 0/5000060 | 0/5000060  | sync
 demo-pg-2        | 10.244.0.24 | streaming | 0/5000060 | 0/5000060 | 0/5000060 | 0/5000060  | sync

```
To know how to set custom configuration for postgres please check [here](/docs/guides/postgres/configuration/using-config-file.md).

### synchronous_commit
`remote_write:` By default `KubeDB Postgres` uses `remote_write` for `synchronous_commit`, which is the least sufficient option for replication
in terms of data preservation as it only guarantees that transaction was replicated over the network and saved into the 
standby's `WAL(write-ahead-log)` without `fsync`. `KubeDB` is using it to ensure minimum latency.

`remote_apply:` which means that the transaction upon completion will be both: persisted to a durable storage and visible 
to a user on standby server(s). Note that this will cause much larger commit delays than other options.

`on:` is a quite safe option when dealing with synchronous replication.
`on` which in context of synchronous replication might be better referred to as `remote_flush`. 
Commits will wait until replies from the current synchronous standby(s) indicate they have received the commit record of
the transaction and flushed it to disk. Although the output of the transaction will not be immediately visible to the users 
on the standby server(s).

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo pg/demo-pg -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo pg/demo-pg

kubectl delete ns demo
```

If you would like to uninstall KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- Learn about [backup and restore](/docs/guides/postgres/backup/overview/index.md) PostgreSQL database using Stash.
- Learn about initializing [PostgreSQL with Script](/docs/guides/postgres/initialization/script_source.md).
- Monitor your PostgreSQL database with KubeDB using [built-in Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md).
- Monitor your PostgreSQL database with KubeDB using [Prometheus operator](/docs/guides/postgres/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
