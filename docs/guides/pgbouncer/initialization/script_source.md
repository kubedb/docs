---
title: Initialize PgBouncer using Script Source
menu:
  docs_{{ .version }}:
    identifier: pb-script-source-initialization
    name: Using Script
    parent: pb-initialization-pgbouncer
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialize PgBouncer with Script

KubeDB supports PgBouncer initialization using scripts stored in a ConfigMap. This tutorial will show you how to use KubeDB to initialize a PgBouncer connection pooler from a script source.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created

$ kubectl get ns demo
NAME    STATUS  AGE
demo    Active  5s
```

> Note: YAML files used in this tutorial are stored in [docs/examples/pgbouncer](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgbouncer) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prepare PostgreSQL Backend

PgBouncer is a connection pooler for PostgreSQL and requires a running PostgreSQL instance as its backend. Prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md).

## Prepare Initialization Scripts

PgBouncer supports initialization with `.sh` scripts. In this tutorial, we will use an `init.sh` script to configure additional connection pool settings after startup.

We will use a ConfigMap as the script source. You can use any Kubernetes supported [volume](https://kubernetes.io/docs/concepts/storage/volumes) as a script source.

Let's create a ConfigMap with the initialization script:

```bash
$ kubectl create configmap -n demo pb-init-script \
--from-literal=init.sh="$(curl -fsSL https://raw.githubusercontent.com/kubedb/pgbouncer-pgpool-init-scripts/master/pgbouncer/init.sh)"
configmap/pb-init-script created
```

## Create PgBouncer with Script Source

Following YAML describes the PgBouncer object with `init.script`:

```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  name: script-pgbouncer
  namespace: demo
spec:
  version: "1.24.0"
  replicas: 1
  database:
    syncUsers: true
    databaseName: "postgres"
    databaseRef:
      name: "quick-postgres"
      namespace: demo
  connectionPool:
    maxClientConnections: 20
    reservePoolSize: 5
  init:
    script:
      configMap:
        name: pb-init-script
  deletionPolicy: WipeOut
```

Here,

- `init.script` specifies the scripts used to initialize PgBouncer when it is being created.
- `database.databaseRef` points to the backing PostgreSQL instance.

VolumeSource provided in `init.script` will be mounted in the Pod and executed while creating PgBouncer.

Now, let's create the PgBouncer CRD using the YAML shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/initialization/script-pgbouncer.yaml
pgbouncer.kubedb.com/script-pgbouncer created
```

Now, wait until PgBouncer goes in `Ready` state. Verify that it is in `Ready` state using the following command:

```bash
$ kubectl get pgbouncer -n demo script-pgbouncer
NAME               VERSION   STATUS   AGE
script-pgbouncer   1.24.0    Ready    2m
```

## Verify Initialization

Now let's connect to our PgBouncer instance to verify that it has been initialized successfully.

**Connection Information:**

- Host name/address: you can use any of these
  - Service: `script-pgbouncer.demo`
  - Pod IP: (`$ kubectl get pods script-pgbouncer-0 -n demo -o yaml | grep podIP`)
- Port: `5432`

- Username: Run the following command to get the *username*:

  ```bash
  $ kubectl get secret -n demo quick-postgres-auth -o jsonpath='{.data.username}' | base64 -d
  postgres
  ```

- Password: Run the following command to get the *password*:

  ```bash
  $ kubectl get secret -n demo quick-postgres-auth -o jsonpath='{.data.password}' | base64 -d
  S3cur3P@ssw0rd
  ```

Connect to PgBouncer and verify that it is successfully proxying connections to PostgreSQL:

```bash
$ kubectl exec -it -n demo script-pgbouncer-0 -- psql -h localhost -p 5432 -U postgres -d postgres -c "SHOW POOLS;"
 database | user | cl_active | cl_waiting | sv_active | sv_idle | sv_used | sv_tested | sv_login | maxwait | pool_mode
----------+------+-----------+------------+-----------+---------+---------+-----------+----------+---------+-----------
 postgres | postgres |         0 |          0 |         0 |       0 |       0 |         0 |        0 |       0 | session
(1 row)
```

We can see that PgBouncer is running and proxying connections to the PostgreSQL backend through the initialized connection pool.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo pgbouncer/script-pgbouncer -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete -n demo pgbouncer/script-pgbouncer

$ kubectl delete -n demo configmap/pb-init-script
$ kubectl delete ns demo
```

## Next Steps

- Learn about [backup and restore](/docs/guides/pgbouncer/backup/overview/index.md) PgBouncer using Stash.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
