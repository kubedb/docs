---
bastitle: Pgpool Quickstart
menu:
  docs_{{ .version }}:
    identifier: pp-quickstart-quickstart
    name: Overview
    parent: pp-quickstart-pgpool
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running Pgpool

This tutorial will show you how to use KubeDB to run a Pgpool.

<p align="center">
  <img alt="lifecycle"  src="/docs/images/pgpool/quickstart/lifecycle.png">
</p>

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses two separate namespaces called `demo` for deploying PostgreSQL and `pool` for Pgpool,  throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

```bash
$ kubectl create ns pool
namespace/pool created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/pgpool](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgpool) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> We have designed this tutorial to demonstrate a production setup of KubeDB managed Pgpool. If you just want to try out KubeDB, you can bypass some of the safety features following the tips [here](/docs/guides/pgpool/quickstart/quickstart.md#tips-for-testing).

## Find Available PgpoolVersion

When you have installed KubeDB, it has created `PgpoolVersion` crd for all supported Pgpool versions. Let's check available PgpoolVersion by,

```bash
$ kubectl get pgpoolversions

  NAME    VERSION   PGPOOL_IMAGE                            DEPRECATED   AGE
  4.4.5   4.4.5     ghcr.io/appscode-images/pgpool2:4.4.5                2d17h
  4.5.0   4.5.0     ghcr.io/appscode-images/pgpool2:4.5.0                2d17h
```

Notice the `DEPRECATED` column. Here, `true` means that this PgpoolVersion is deprecated for current KubeDB version. KubeDB will not work for deprecated PgpoolVersion.

In this tutorial, we will use `4.5.0` PgpoolVersion crd to create Pgpool. To know more about what `PgpoolVersion` crd is, please visit [here](/docs/guides/pgpool/concepts/catalog.md). You can also see supported PgpoolVersion [here](/docs/guides/pgpool/README.md#supported-pgpoolversion-crd).

## Get PostgreSQL Server ready

Pgpool is a middleware for PostgreSQL. Therefore you will need to have a PostgreSQL server up and running for Pgpool to connect to.

Luckily PostgreSQL is readily available in KubeDB as crd and can easily be deployed using this guide [here](/docs/guides/postgres/quickstart/quickstart.md). But by default this will create a PostgreSQL server with `max_connections=100`, but we need more than 100 connections for our Pgpool to work as expected. 

Pgpool requires at least `2*num_init_children*max_pool*spec.replicas` connections in PostgreSQL server. So use [this](https://kubedb.com/docs/v2024.4.27/guides/postgres/configuration/using-config-file/) to create a PostgreSQL server with custom `max_connections`.

In this tutorial, we will use a PostgreSQL named `quick-postgres` in the `demo` namespace.
Let's first create a secret for PostgreSQL custom configuration. At first, let’s create `user.conf` file setting `max_connections` parameter.
```bash
$ cat user.conf
max_connections=400
```
> Note that config file name must be `user.conf`

Now, create a Secret with this configuration file.

```bash
$ kubectl create secret generic -n demo pg-configuration --from-literal=user.conf="$(curl -fsSL https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/quickstart/user.conf)"
secret/pg-configuration created
```

Now, create PostgreSQL crd specifying spec.configSecret field.
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/quickstart/quick-postgres.yaml
postgres.kubedb.com/quick-postgres created
```

KubeDB creates all the necessary resources including services, secrets, and appbindings to get this server up and running. A default database `postgres` is created in `quick-postgres`. Database secret `quick-postgres-auth` holds this user's username and password. Following is the yaml file for it.

```bash
$ kubectl get secrets -n demo quick-postgres-auth -o yaml
```
```yaml
apiVersion: v1
data:
  password: M21ufmFwM0ltTmpNUTI1ag==
  username: cG9zdGdyZXM=
kind: Secret
metadata:
  creationTimestamp: "2024-05-02T09:37:01Z"
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-postgres
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: postgreses.kubedb.com
  name: quick-postgres-auth
  namespace: demo
  resourceVersion: "103369"
  uid: ce5462e8-f480-4f8c-827a-66505b3d197b
type: kubernetes.io/basic-auth
```

For the purpose of this tutorial, we will need to extract the username and password from database secret `quick-postgres-auth`.

```bash
$ kubectl get secrets -n demo quick-postgres-auth -o jsonpath='{.data.\password}' | base64 -d
3mn~ap3ImNjMQ25j⏎

$ kubectl get secrets -n demo quick-postgres-auth -o jsonpath='{.data.\username}' | base64 -d
postgres⏎ 
```

Now, to test connection with this database using the credentials obtained above, we will expose the service port associated with `quick-postgres`  to localhost.

```bash
$ kubectl port-forward -n demo svc/quick-postgres 5432
Forwarding from 127.0.0.1:5432 -> 5432
Forwarding from [::1]:5432 -> 5432
```

With that done, we should now be able to connect to `postgres` database using username `postgres`, and password `3mn~ap3ImNjMQ25j`.

```bash
$ export PGPASSWORD='3mn~ap3ImNjMQ25j'
$ psql --host=localhost --port=5432 --username=postgres postgres
psql (16.2 (Ubuntu 16.2-1.pgdg22.04+1), server 13.13)
Type "help" for help.

postgres=# 
```

After establishing connection successfully, we will create a table in `postgres` database and populate it with data.

```bash
postgres=# CREATE TABLE COMPANY( NAME TEXT NOT NULL, EMPLOYEE INT NOT NULL);
CREATE TABLE
postgres=# INSERT INTO COMPANY (name, employee) VALUES ('Apple',10);
INSERT 0 1
postgres=# INSERT INTO COMPANY (name, employee) VALUES ('Google',15);
INSERT 0 1
```

After data insertion, we need to verify that our data have been inserted successfully.

```bash
postgres=# SELECT * FROM company ORDER BY name;
  name  | employee
--------+----------
 Apple  |       10
 Google |       15
(2 rows)
postgres=# \q
```

If no error occurs, `quick-postgres` is ready to be used by Pgpool.

You can also use any other externally managed PostgreSQL server and create a database `postgres` for user `postgres`.

If you choose not to use KubeDB to deploy Postgres, create AppBinding to point Pgpool to your PostgreSQL server. Click [here](/docs/guides/pgpool/concepts/appbinding.md) for detailed instructions on how to manually create AppBindings for Postgres.

## Create a Pgpool Server

KubeDB implements a Pgpool crd to define the specifications of a Pgpool.

Below is the Pgpool object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: quick-pgpool
  namespace: pool
spec:
  version: "4.5.0"
  replicas: 1
  postgresRef:
    name: quick-postgres
    namespace: demo
  sslMode: disable
  clientAuthMode: md5
  syncUsers: true
  terminationPolicy: WipeOut
```

Here,

- `spec.version` is name of the PgpoolVersion crd where the docker images are specified. In this tutorial, a Pgpool with base image version `4.5.0` is created.
- `spec.replicas` specifies the number of replica pgpool server pods to be created for the Pgpool object.
- `spec.postgresRef` specifies the name and the namespace of the appbinding that points to the PostgreSQL server.
- `spec.sslMode` specifies ssl mode for clients.
- `spec.clientAuthMode` specifies the authentication method that will be used for clients.
- `spec.syncUsers` specifies whether user want to sync additional users to Pgpool.
- `spec.terminationPolicy` specifies what policy to apply while deletion.

Now that we've been introduced to the pgpool crd, let's create it,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/quickstart/pgpool-server.yaml
pgpool.kubedb.com/quick-pgpool created
```

## Connect via Pgpool

To connect via pgpool we have to expose its service to localhost.

```bash
$ kubectl port-forward -n pool svc/quick-pgpool 9999
Forwarding from 127.0.0.1:9999 -> 9999
```

Now, let's connect to `postgres` database via Pgpool using psql.

``` bash
$ export PGPASSWORD='3mn~ap3ImNjMQ25j'
$ psql --host=localhost --port=9999 --username=postgres postgres
psql (16.2 (Ubuntu 16.2-1.pgdg22.04+1), server 13.13)
Type "help" for help.

postgres=#
```

If everything goes well, we'll be connected to the `postgres` database and be able to execute commands. Let's confirm if the company data we inserted in the  `postgres` database before are available via Pgpool:

```bash
$ psql --host=localhost --port=9999 --username=postgres postgres --command='SELECT * FROM company ORDER BY name;'
  name  | employee
--------+----------
 Apple  |       10
 Google |       15
(2 rows)
```

KubeDB operator watches for Pgpool objects using Kubernetes api. When a Pgpool object is created, KubeDB operator will create a new PetSet and a Service with the matching name. KubeDB operator will also create a governing service for PetSet, if one is not already present.

KubeDB operator sets the `status.phase` to `Ready` once Pgpool is ready after all checks.

```bash
$ kubectl get pp -n pool quick-pgpool -o wide
NAME           TYPE                  VERSION   STATUS   AGE
quick-pgpool   kubedb.com/v1alpha2   4.5.0     Ready    63m

```

Let's describe Pgpool object `quick-pgpool`

```bash
$ kubectl dba describe pp -n pool quick-pgpool
Name:         quick-pgpool
Namespace:    pool
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Pgpool
Metadata:
  Creation Timestamp:  2024-05-02T10:39:44Z
  Finalizers:
    kubedb.com
  Generation:  2
  Managed Fields:
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:clientAuthMode:
        f:healthChecker:
          .:
          f:failureThreshold:
          f:periodSeconds:
          f:timeoutSeconds:
        f:podPlacementPolicy:
        f:postgresRef:
          .:
          f:name:
          f:namespace:
        f:replicas:
        f:sslMode:
        f:syncUsers:
        f:terminationPolicy:
        f:version:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2024-05-02T10:39:44Z
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:finalizers:
          .:
          v:"kubedb.com":
      f:spec:
        f:authSecret:
    Manager:      pgpool-operator
    Operation:    Update
    Time:         2024-05-02T10:39:44Z
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:phase:
    Manager:         pgpool-operator
    Operation:       Update
    Subresource:     status
    Time:            2024-05-02T10:40:41Z
  Resource Version:  109413
  UID:               f742442c-50e6-4aa7-92a2-bf423efdabb0
Spec:
  Auth Secret:
    Name:            quick-pgpool-auth
  Client Auth Mode:  md5
  Health Checker:
    Failure Threshold:  1
    Period Seconds:     10
    Timeout Seconds:    10
  Pod Placement Policy:
    Name:  default
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Containers:
        Name:  pgpool
        Resources:
          Limits:
            Memory:  1Gi
          Requests:
            Cpu:     500m
            Memory:  1Gi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Group:     70
          Run As Non Root:  true
          Run As User:      70
          Seccomp Profile:
            Type:  RuntimeDefault
      Security Context:
        Fs Group:  70
  Postgres Ref:
    Name:              quick-postgres
    Namespace:         demo
  Replicas:            1
  Ssl Mode:            disable
  Sync Users:          true
  Termination Policy:  WipeOut
  Version:             4.5.0
Status:
  Conditions:
    Last Transition Time:  2024-05-02T10:39:44Z
    Message:               The KubeDB operator has started the provisioning of Pgpool: pool/quick-pgpool
    Observed Generation:   1
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2024-05-02T10:40:28Z
    Message:               All replicas are ready for Pgpool pool/quick-pgpool
    Observed Generation:   2
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2024-05-02T10:40:39Z
    Message:               pgpool pool/quick-pgpool is accepting connection
    Observed Generation:   2
    Reason:                AcceptingConnection
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2024-05-02T10:40:39Z
    Message:               pgpool pool/quick-pgpool is ready
    Observed Generation:   2
    Reason:                AllReplicasReady
    Status:                True
    Type:                  Ready
    Last Transition Time:  2024-05-02T10:40:39Z
    Message:               The Pgpool: pool/quick-pgpool is successfully provisioned.
    Observed Generation:   2
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:                    <none>

```

KubeDB has created a services for the Pgpool object.

```bash
$ `kubectl get service -n pool --selector=app.kubernetes.io/name=pgpools.kubedb.com,app.kubernetes.io/instance=quick-pgpool`
NAME                TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
quick-pgpool        ClusterIP   10.96.33.221   <none>        9999/TCP   67m
quick-pgpool-pods   ClusterIP   None           <none>        9999/TCP   67m
```

Here, Service *`quick-pgpool`* targets random pods to carry out any operation that are made through this service.

## Cleaning up

If you don't set the terminationPolicy, then the kubeDB set the TerminationPolicy to `Delete` by-default.

### Delete
If you want to delete the existing pgpool, but want to keep the secrets intact then you might want to set the pgpool object terminationPolicy to Delete. In this setting, PetSet and the services will be deleted.

When the TerminationPolicy is set to Delete and the pgpool object is deleted, the KubeDB operator will delete the PetSet and its pods along with the services  but leaves the secrets intact.

```bash
$ kubectl patch -n pool pp/quick-pgpool -p '{"spec":{"terminationPolicy":"Delete"}}' --type="merge"
pgpool.kubedb.com/quick-pgpool patched

$ kubectl delete -n pool pp/quick-pgpool
pgpool.kubedb.com "quick-pgpool" deleted

$ kubectl get pp,petset,svc,secret -n pool
NAME                         TYPE                       DATA   AGE
secret/quick-pgpool-auth     kubernetes.io/basic-auth   2      3h22m
secret/quick-pgpool-config   Opaque                     2      3h22m

$ kubectl delete ns pool
namespace "pool" deleted

$ kubectl delete -n demo pg/quick-postgres
pgpool.kubedb.com "quick-postgres" deleted
```

### WipeOut
But if you want to cleanup each of the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n pool pp/quick-pgpool -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"

$ kubectl delete -n pool pp/quick-pgpool
pgpool.kubedb.com "quick-pgpool" deleted

$ kubectl get pp,petset,svc,secret -n pool
No resources found in pool namespace.

$ kubectl delete ns pool
namespace "pool" deleted

$ kubectl delete -n demo pg/quick-postgres
pgpool.kubedb.com "quick-postgres" deleted
```

## Next Steps

- Detail concepts of [Pgpool object](/docs/guides/pgpool/concepts/pgpool.md).
- Detail concepts of [PgpoolVersion object](/docs/guides/pgpool/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
```

