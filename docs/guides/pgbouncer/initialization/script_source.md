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
> **Note:** The initialization script above is provided only as an example. You can use your own initialization script as long as it performs the required setup for your environment. If your script connects to PostgreSQL, make sure to include the appropriate PostgreSQL credentials (such as the password) so the script can authenticate successfully.
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
$ kubectl exec -it -n demo script-pgbouncer-0 -- \
                                  psql -h localhost -p 5432 -U postgres -d postgres
Password for user postgres: 
psql (16.14, server 17.5)
WARNING: psql major version 16, server major version 17.
         Some psql features might not work.
Type "help" for help.

postgres=# \dt
                    List of relations
 Schema |             Name             | Type  |  Owner   
--------+------------------------------+-------+----------
 public | kubedb_write_check_pgbouncer | table | postgres
 public | my_table                     | table | postgres
(2 rows)
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

- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
