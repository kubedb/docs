---
bastitle: PgBouncer Quickstart
menu:
  docs_{{ .version }}:
    identifier: pb-quickstart-quickstart
    name: Overview
    parent: pb-quickstart-pgbouncer
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Running PgBouncer

This tutorial will show you how to use KubeDB to run a PgBouncer.

<p align="center">
  <img alt="lifecycle"  src="/docs/images/postgres/lifecycle.png">
</p>

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/pgbouncer](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgbouncer) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> We have designed this tutorial to demonstrate a production setup of KubeDB managed PgBouncer. If you just want to try out KubeDB, you can bypass some of the safety features following the tips [here](/docs/guides/pgbouncer/quickstart/quickstart.md#tips-for-testing).

## Find Available PgBouncerVersion

When you have installed KubeDB, it has created `PgBouncerVersion` crd for all supported PgBouncer versions. Let's check available PgBouncerVersion by,

```console
$ kubectl get pgbouncerversions

    NAME     VERSION   DB_IMAGE   DEPRECATED   AGE
    1.10.0   1.10.0               false        75m
    1.11.0   1.11.0               false        75m
    1.12.0   1.12.0               false        75m
    1.7      1.7                  false        75m
    1.7.1    1.7.1                false        75m
    1.7.2    1.7.2                false        75m
    1.8.1    1.8.1                false        75m
    1.9.0    1.9.0                false        75m
    latest   latest               false        75m
```

Notice the `DEPRECATED` column. Here, `true` means that this PgBouncerVersion is deprecated for current KubeDB version. KubeDB will not work for deprecated PgBouncerVersion.

In this tutorial, we will use `1.11.0` PgBouncerVersion crd to create PgBouncer. To know more about what `PgBouncerVersion` crd is, please visit [here](/docs/concepts/catalog/pgbouncer.md). You can also see supported PgBouncerVersion [here](/docs/guides/pgbouncer/README.md#supported-pgbouncerversion-crd).

## Get PostgreSQL Server ready

PgBouncer is a connection-pooling middleware for PostgreSQL. Therefore you will need to have a PostgreSQL server up and running for PgBouncer to connect to.

Luckily PostgreSQL is readily available in KubeDB as crd and can easily be deployed using this guide [here](/docs/guides/postgres/quickstart/quickstart.md).

In this tutorial, we will use a Postgres named `quick-postgres` in the `demo` namespace.

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/quickstart/quick-postgres.yaml
postgres.kubedb.com/quick-postgres created
```

KubeDB creates all the necessary resources including services, secrets, and appbindings to get this server up and running. A default database `postgres` is created in `quick-postgres`. Database secret `quick-postgres-auth` holds this user's username and password. Following is the yaml file for it.

```yaml
$kubectl get secrets -n demo quick-postgres-auth -o yaml

apiVersion: v1
data:
  POSTGRES_PASSWORD: cVRPenVkYnp1c2xzNk5UWg==
  POSTGRES_USER: cG9zdGdyZXM=
kind: Secret
metadata:
  creationTimestamp: "2019-08-30T06:08:44Z"
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-postgres
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: postgres
    app.kubernetes.io/version: 11.1-v1
    kubedb.com/kind: Postgres
    kubedb.com/name: quick-postgres
  name: quick-postgres-auth
  namespace: demo
  resourceVersion: "12567"
  selfLink: /api/v1/namespaces/demo/secrets/quick-postgres-auth
  uid: 57c3a271-a834-460f-a802-2544fb608752
type: Opaque
```

For the purpose of this tutorial, we will need to extract the username and password from database secret `quick-postgres-auth`.

```console
$kubectl get secrets -n demo quick-postgres-auth -o jsonpath='{.data.\POSTGRES_PASSWORD}' | base64 -d
qTOzudbzusls6NTZ⏎

$ kubectl get secrets -n demo quick-postgres-auth -o jsonpath='{.data.\POSTGRES_USER}' | base64 -d
postgres⏎
```

Now, to test connection with this database using the credentials obtained above, we will expose the service port associated with `quick-postgres`  to localhost.

```console
$ kubectl port-forward -n demo svc/quick-postgres 5432
Forwarding from 127.0.0.1:5432 -> 5432
Forwarding from [::1]:5432 -> 5432
```

With that done , we should now be able to connect to `postgres` database using username `postgres`, and password `qTOzudbzusls6NTZ`.

```console
$ export PGPASSWORD=qTOzudbzusls6NTZ
$ psql --host=localhost --port=5432 --username=postgres postgres
psql (11.5 (Ubuntu 11.5-1.pgdg18.04+1), server 11.1)
Type "help" for help.

postgres=#
```

After establishing connection successfully, we will create a table in `postgres` database and populate it with data.

```console
postgres=# CREATE TABLE COMPANY( NAME TEXT NOT NULL, EMPLOYEE INT NOT NULL);
CREATE TABLE
postgres=# INSERT INTO COMPANY (name, employee) VALUES ('Apple',10);
INSERT 0 1
postgres=# INSERT INTO COMPANY (name, employee) VALUES ('Google',15);
INSERT 0 1
```

After data insertion, we need to verify that our data have been inserted successfully.

```console
postgres=# SELECT * FROM company ORDER BY name;
  name  | employee
--------+----------
 Apple  |       10
 Google |       15
(2 rows)
postgres=# \q
```

If no error occurs, `quick-postgres` is ready to be used by PgBouncer.

You can also use any other tool to deploy your PostgreSQL server and create a database `postgres` for user `postgres`.

Should you choose not to use KubeDB to deploy Postgres, create AppBinding(s) to point PgBouncer to your PostgreSQL server(s) where your target databases are located. Click [here](/docs/concepts/appbinding.md) for detailed instructions on how to manually create AppBindings for Postgres.

## Create a PgBouncer Server

KubeDB implements a PgBouncer crd to define the specifications of a PgBouncer.

Below is the PgBouncer object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: PgBouncer
metadata:
  name: pgbouncer-server
  namespace: demo
spec:
  version: "1.11.0"
  replicas: 1
  databases:
  - alias: "postgres"
    databaseName: "postgres"
    databaseRef:
      name: "quick-postgres"
  connectionPool:
    maxClientConnections: 20
    reservePoolSize: 5
    adminUsers:
    - admin
    - admin1
  userListSecretRef:
    name: db-user-pass
```

Here,

- `spec.version` is name of the PgBouncerVersion crd where the docker images are specified. In this tutorial, a PgBouncer with base image version 1.11.0 is created.
- `spec.replicas` specifies the number of replica pgbouncer server pods to be created for the PgBouncer object.
- `spec.databases` specifies the databases that are going to be served via PgBouncer.
- `spec.connectionPool` specifies the configurations for connection pool.
- `spec.userListSecretRef` specifies the secret that contains the standard pgbouncer `userlist` file.

### spec.databases

Databases contain three `required` fields and two `optional` fields.

- `spec.databases.alias`:  specifies an alias for the target database located in a postgres server specified by an appbinding.
- `spec.databases.databaseName`:  specifies the name of the target database.
- `spec.databases.databaseRef`:  specifies the name and namespace of the appBinding that contains the path to a PostgreSQL server where the target database can be found.
- `spec.databases.username` (optional):  specifies the user with whom this particular database should have an exclusive connection. By default, if this field is left empty, all users will be able to use the database.
- `spec.databases.password` (optional):  specifies password to authenticate the user with whom this particular database should have an exclusive connection.

### spec.connectionPool

 ConnectionPool is used to configure pgbouncer connection pool. All the fields here are accompanied by default values and can be left unspecified if no customization is required by the user.

- `spec.connectionPool.port`: specifies the port on which pgbouncer should listen to connect with clients. The default is 5432.
- `spec.connectionPool.adminUsers`: specifies the values of admin_users. An array of names of admin users are listed here.
- `spec.connectionPool.authType`: specifies how to authenticate users.
- `spec.connectionPool.poolMode`: specifies the value of pool_mode.
- `spec.connectionPool.maxClientConnections`: specifies the value of max_client_conn.
- `spec.connectionPool.defaultPoolSize`: specifies the value of default_pool_size.
- `spec.connectionPool.minPoolSize`: specifies the value of min_pool_size.
- `spec.connectionPool.reservePoolSize`: specifies the value of reserve_pool_size.
- `spec.connectionPool.reservePoolTimeout`: specifies the value of reserve_pool_timeout.
- `spec.connectionPool.maxDbConnections`: specifies the value of max_db_connections.
- `spec.connectionPool.maxUserConnections`: specifies the value of max_user_connections.

### spec.userListSecretRef

UserList field is used to specify a secret that contains the list of authorized users along with their passwords. Basically this secret is created from the standard pgbouncer userlist file.

- `spec.userListSecretRef.name`: specifies the name of the secret containing userlist in the same namespace as the PgBouncer crd.

In this tutorial we will use a standard userlist text file to create a secret for spec.userListSecretRef. In the userlist text file we have added `pgbouncer` as user and  `qTOzudbzusls6NTZ` as corresponding password. The file looks like this:

```
"postgres" "qTOzudbzusls6NTZ"
"myuser" "mypass"
```

We will need user `myuser` with password  `mypass` later in this tutorial.

```console
$ kubectl create secret -n demo generic db-user-pass --from-file=https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/quickstart/userlist.txt

secret/db-user-pass created
```

Now that we've been introduced to the pgBouncer crd, let's create it,

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/quickstart/pgbouncer-server.yaml

pgbouncer.kubedb.com/pgbouncer-server created
```

## Connect via PgBouncer

To connect via pgBouncer we have to expose its service to localhost.

```console
$ kubectl port-forward -n demo svc/pgbouncer-server 5432

Forwarding from 127.0.0.1:5432 -> 5432
Forwarding from [::1]:5432 -> 5432
```

Now, let's connect to `postgres` database via PgBouncer using psql.

``` bash
$ env PGPASSWORD=qTOzudbzusls6NTZ psql --host=localhost --port=5432 --username=postgres postgres
psql (11.5 (Ubuntu 11.5-1.pgdg18.04+1), server 11.1)
Type "help" for help.

postgres=# \q
```

If everything goes well, we'll be connected to the `postgres` database and be able to execute commands. Let's confirm if the company data we inserted in the  `postgres` database before are available via PgBouncer:

```console
$ env PGPASSWORD=qTOzudbzusls6NTZ psql --host=localhost --port=5432 --username=postgres postgres --command='SELECT * FROM company ORDER BY name;'
  name  | employee
--------+----------
 Apple  |       10
 Google |       15
(2 rows)
```

## Add New Connections to the Pool

We will add a new user and a new database to our PostgreSQL server `quick-postgres` and add this database to the existing pool and connect to this database using newly created user.

First lets create a new user `myuser` with password `mypass`

```console
$ env PGPASSWORD=qTOzudbzusls6NTZ psql --host=localhost --port=5432 --username=postgres postgres --command="create user myuser with encrypted password 'mypass'"
CREATE ROLE
```

And then create a new database `mydb`

```console
$ env PGPASSWORD=qTOzudbzusls6NTZ psql --host=localhost --port=5432 --username=postgres postgres --command="CREATE DATABASE mydb;"
CREATE DATABASE
```

Now we will need to edit our PgBouncer's spec.databases to add this database to the connection pool.

The YAML file should now look like this:

```yaml
apiVersion: kubedb.com/v1alpha1
kind: PgBouncer
metadata:
  name: pgbouncer-server
  namespace: demo
spec:
  version: "1.11.0"
  replicas: 1
  databases:
  - alias: "postgres"
    databaseName: "postgres"
    databaseRef:
      name: "quick-postgres"
  - alias: "tmpdb"
    databaseName: "mydb"
    databaseRef:
      name: "quick-postgres"
  connectionPool:
    maxClientConnections: 20
    reservePoolSize: 5
    adminUsers:
    - admin
    - admin1
  userListSecretRef:
    name: db-user-pass
```

We have given our newly added database an alias `tmpdb`.  We will now apply this modified file.

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/quickstart/pgbouncer-server-mod.yaml
pgbouncer.kubedb.com/pgbouncer-server configured
```

Let's try to connect to `mydb` via PgBouncer.

```console
$ env PGPASSWORD=mypass psql --host=localhost --port=5432 --username=myuser tmpdb
psql (11.5 (Ubuntu 11.5-1.pgdg18.04+1), server 11.1)
Type "help" for help.

tmpdb=>
```

We can now switch our connection between our existing databases `postgres` and `mydb` as well.

```console
tmpdb=>\c postgres
psql (11.5 (Ubuntu 11.5-1.pgdg18.04+1), server 11.1)
You are now connected to database "postgres" as user "myuser".
postgres=>\c mydb
psql (11.5 (Ubuntu 11.5-1.pgdg18.04+1), server 11.1)
You are now connected to database "mydb" as user "myuser".
tmpdb=>\q
```

KubeDB operator watches for PgBouncer objects using Kubernetes api. When a PgBouncer object is created, KubeDB operator will create a new StatefulSet and a Service with the matching name. KubeDB operator will also create a governing service for StatefulSet with the name `kubedb`, if one is not already present.

KubeDB operator sets the `status.phase` to `Running` once the connection-pooling mechanism is ready.

```console
$ kubectl get pb -n demo pgbouncer-server -o wide
NAME               VERSION   STATUS    AGE
pgbouncer-server   1.11.0    Running   2h
```

Let's describe PgBouncer object `pgbouncer-server`

```console
$ kubectl dba describe pb -n demo pgbouncer-server
Name:         pgbouncer-demo
Namespace:    demo
Labels:       <none>
Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                {"apiVersion":"kubedb.com/v1alpha1","kind":"PgBouncer","metadata":{"annotations":{},"name":"pgbouncer-demo","namespace":"demo"},"spec":{"c...
API Version:  kubedb.com/v1alpha1
Kind:         PgBouncer
Metadata:
  Creation Timestamp:  2019-10-31T10:34:04Z
  Finalizers:
    kubedb.com
  Generation:        1
  Resource Version:  4733
  Self Link:         /apis/kubedb.com/v1alpha1/namespaces/demo/pgbouncers/pgbouncer-demo
  UID:               158b7c58-ecb2-4a77-bceb-081489b4921a
Spec:
  Connection Pool:
    Admin Users:
      admin
      admin1
    Pool Mode:          session
    Port:               5432
    Reserve Pool Size:  5
  Databases:
    Alias:          postgres
    Database Name:  postgres
    Database Ref:
      Name:         quick-postgres
      Namespace:
    Alias:          tmpdb
    Database Name:  mydb
    Database Ref:
      Name:       quick-postgres
      Namespace:
  Monitor:
    Agent:  prometheus.io/builtin
    Prometheus:
      Port:  56790
    Resources:
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Resources:
  Replicas:  1
  Service Template:
    Metadata:
    Spec:
  User List Secret Ref:
    Name:   db-user-pass
  Version:  1.12.0
Status:
  Observed Generation:  1$6208915667192219204
  Phase:                Running
Events:                 <none>
```

KubeDB has created a service for the PgBouncer object.

```console
$ kubectl get service -n demo --selector=kubedb.com/kind=PgBouncer,kubedb.com/name=pgbouncer-server
NAME               TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
pgbouncer-server   ClusterIP   10.97.188.32   <none>        5432/TCP   2h
```

Here, Service *`pgbouncer-server`* targets random pods to carry out connection-pooling.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
kubectl delete -n demo pg/quick-postgres

kubectl delete -n demo pb/pgbouncer-server
kubectl delete -n demo pb/pgbouncer-server-mod

kubectl delete secret -n demo db-user-pass

kubectl delete ns demo
```

## Next Steps

- Learn about [custom PgBouncerVersions](/docs/guides/pgbouncer/custom-versions/setup.md).
- Monitor your PgBouncer with KubeDB using [built-in Prometheus](/docs/guides/pgbouncer/monitoring/using-builtin-prometheus.md).
- Monitor your PgBouncer with KubeDB using [CoreOS Prometheus Operator](/docs/guides/pgbouncer/monitoring/using-coreos-prometheus-operator.md).
- Detail concepts of [PgBouncer object](/docs/concepts/database-proxy/pgbouncer.md).
- Use [private Docker registry](/docs/guides/pgbouncer/private-registry/using-private-registry.md) to deploy PgBouncer with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
