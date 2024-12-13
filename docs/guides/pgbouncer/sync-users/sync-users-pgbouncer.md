---
title: Runtime users sync to PgBouncer
menu:
  docs_{{ .version }}:
    identifier: pb-sync-users-pgbouncer
    name: Sync users pgbouncer
    parent: pb-sync-users
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Sync Users

KubeDB supports providing a way to add/update users to PgBouncer in runtime simply by creating secret with defined keys and labels. This tutorial will show you how to use KubeDB to sync a user to PgBouncer on runtime.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/pgbouncer](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgbouncer) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB operator allows us to sync additional Postgres users to PgBouncer on runtime by setting `spec.syncUsers` to `true`, if this option is true KubeDB operator searches for secrets in the namespace of the Postgres mentioned with some certain labels. Then if the secret have username and password as key KubeDB operator will sync the username and password to PgBouncer. Again not only to add a user but also this feature can also be used for updating a user's password.

At first, we need to create a secret that contains a `user` key and a `password` key which contains the `username` and `password` respectively. Also, we need to add two labels `<Appbinding name mentioned in .spec.postgresRef.name>` and `postgreses.kubedb.com`. The namespace must be `<Namespace mentioned in .spec.postgresRef.namespace>`. Below given a sample structure of the secret.

Example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  labels:
    app.kubernetes.io/instance: ha-postgres
    app.kubernetes.io/name: postgreses.kubedb.com
  name: pg-user
  namespace: demo
stringData:
  password: "12345"
  username: "alice"
```
- `app.kubernetes.io/instance` should be same as`appbinding name mentioned in .spec.postgresRef.name`.
- `app.kubernetes.io/name` should be `postgreses.kubedb.com`.
- `namespace` should be same as `namespace mentioned in .spec.postgresRef.namespace`.

In every `20 seconds` KubeDB operator will sync all the users to PgBouncer.

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).

### Prepare Postgres
For a PgBouncer surely we will need a Postgres server so, prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgbouncer/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.

### Prepare PgBouncer

Now, we are going to deploy a `PgBouncer` with version `1.23.1`.

### Deploy PgBouncer

Below is the YAML of the `PgBouncer` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  name: pgbouncer-sync
  namespace: demo
spec:
  version: "1.23.1"
  replicas: 1
  database:
    syncUsers: true
    databaseName: "postgres"
    databaseRef:
      name: "ha-postgres"
      namespace: demo
  deletionPolicy: WipeOut
```

Let's create the `PgBouncer` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/sync-users/pgbouncer-sync.yaml
pgbouncer.kubedb.com/pgbouncer-sync created
```

Now, wait until `pgbouncer-sync` has status `Ready`. i.e,

```bash
$ kubectl get pb -n demo
NAME             TYPE                  VERSION   STATUS   AGE
pgbouncer-sync   kubedb.com/v1         1.18.0    Ready    41s
```

### Sync Users

Now, create a secret with structure defined [here](/docs/guides/pgbouncer/concepts/pgbouncer.md#specsyncusers). Below is the YAML of the `secret` that we are going to create,

```yaml
apiVersion: v1
kind: Secret
metadata:
  labels:
    app.kubernetes.io/instance: ha-postgres
    app.kubernetes.io/name: postgreses.kubedb.com
  name: sync-secret
  namespace: demo
stringData:
  password: "12345"
  username: "john"
```

Now, create the secret by applying the yaml above.

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/sync-users/secret.yaml
secret/sync-secret created
```

Now, after `20 seconds` you can exec into the pgbouncer pod and find if the new user is there,

```bash
$ kubectl exec -it -n demo pgbouncer-sync-0 -- /bin/sh
/$ cat /var/run/pgbouncer/secret/userlist
"postgres" "md5AESOmAkfj+zX8zXLm92d6Vup6a5yASiiGScoHNDTIgBwH8="
"john" "md5AEScbLKDSMb+KVrILhh7XEmyQ=="
"pgbouncer" "md5AESOmAkfj+zX8zXLm92d6Vup6a5yASiiGScoHNDTIgBwH8="
/$ exit
exit
```
We can see that the user is there in PgBouncer. So, now let's create this user and try to use this user through PgBouncer.
Now, you can connect to this pgbouncer through [psql](https://www.postgresql.org/docs/current/app-psql.html). Before that we need to port-forward to the primary service of pgbouncer.

```bash
$ kubectl port-forward svc/pgbouncer-sync -n demo 9999:5432
Forwarding from 127.0.0.1:9999 -> 5432
Forwarding from [::1]:9999 -> 5432
```
We will use the root Postgres user to create the user, so let's get the password for the root user, so that we can use it.
```bash
$ kubectl get secrets -n demo ha-postgres-auth -o jsonpath='{.data.\password}' | base64 -d
qEeuU6cu5aH!O9CI⏎ 
```
We can use this password now,
```bash
$ export PGPASSWORD='qEeuU6cu5aH!O9CI'
$ psql --host=localhost --port=9999 --username=postgres postgres
psql (16.3 (Ubuntu 16.3-1.pgdg22.04+1), server 16.1)
Type "help" for help.

postgres=# CREATE USER john WITH PASSWORD '12345';
CREATE ROLE
postgres=# exit
```
Now, let's use this john user.
```bash
$ export PGPASSWORD='12345'
$ psql --host=localhost --port=9999 --username=john postgres
psql (16.3 (Ubuntu 16.3-1.pgdg22.04+1), server 16.1)
Type "help" for help.

postgres=> exit
```
So, we can successfully verify that the user is registered in PgBouncer and also we can use it.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo pb/pgbouncer-sync
kubectl delete -n demo secret/sync-secret
kubectl delete pg -n demo ha-postgres
kubectl delete ns demo
```

## Next Steps

- Monitor your PgBouncer database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/pgbouncer/monitoring/using-prometheus-operator.md).
- Monitor your PgBouncer database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/pgbouncer/monitoring/using-builtin-prometheus.md).
- Detail concepts of [PgBouncer object](/docs/guides/pgbouncer/concepts/pgbouncer.md).
- Detail concepts of [PgBouncerVersion object](/docs/guides/pgbouncer/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
