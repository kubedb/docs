---
title: PostgreSQL Quickstart
menu:
  docs_{{ .version }}:
    identifier: pg-quickstart-quickstart
    name: Overview
    parent: pg-quickstart-postgres
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running PostgreSQL

This tutorial will show you how to use KubeDB to run a PostgreSQL database.

<p align="center">
  <img alt="lifecycle"  src="/docs/images/postgres/lifecycle.png">
</p>

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/postgres](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/postgres) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

>We have designed this tutorial to demonstrate a production setup of KubeDB managed PostgreSQL. If you just want to try out KubeDB, you can bypass some of the safety features following the tips [here](/docs/guides/postgres/quickstart/quickstart.md#tips-for-testing).

## Install pgAdmin

This tutorial will also use a pgAdmin to connect and test PostgreSQL database, once it is running.

Run the following command to install pgAdmin,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/quickstart/pgadmin.yaml
deployment.apps/pgadmin created
service/pgadmin created

$ kubectl get pods -n demo --watch
NAME                      READY     STATUS              RESTARTS   AGE
pgadmin-5b4b96779-lfpfh   0/1       ContainerCreating   0          1m
pgadmin-5b4b96779-lfpfh   1/1       Running   0         2m
^C⏎
```

Now, you can open pgAdmin on your browser using following address `http://<cluster ip>:<NodePort of pgadmin service>`.

If you are using minikube then open pgAdmin in your browser by running `minikube service pgadmin -n demo`. Or you can get the URL of Service `pgadmin` by running following command

```bash
$ minikube service pgadmin -n demo --url
http://192.168.99.100:31983
```

To log into the pgAdmin, use username __`admin`__ and password __`admin`__.

## Find Available StorageClass

We will have to provide `StorageClass` in Postgres crd specification. Check available `StorageClass` in your cluster using following command,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  10d

```

Here, we have `standard` StorageClass in our cluster.

## Find Available PostgresVersion

When you have installed KubeDB, it has created `PostgresVersion` crd for all supported PostgreSQL versions. Let's check available PostgresVersions by,

```bash
$ kubectl get postgresversion
NAME                       VERSION   DISTRIBUTION   DB_IMAGE                               DEPRECATED   AGE
10.16                      10.16     Official       postgres:10.16-alpine                               3d
10.16-debian               10.16     Official       postgres:10.16                                      3d
10.19                      10.19     Official       postgres:10.19-bullseye                             3d
10.19-bullseye             10.19     Official       postgres:10.19-bullseye                             3d
10.20                      10.20     Official       postgres:10.20-bullseye                             3d
10.20-bullseye             10.20     Official       postgres:10.20-bullseye                             3d
11.11                      11.11     Official       postgres:11.11-alpine                               3d
11.11-debian               11.11     Official       postgres:11.11                                      3d
11.14                      11.14     Official       postgres:11.14-alpine                               3d
11.14-bullseye             11.14     Official       postgres:11.14-bullseye                             3d
11.14-bullseye-postgis     11.14     PostGIS        postgis/postgis:11-3.1                              3d
11.15                      11.15     Official       postgres:11.15-alpine                               3d
11.15-bullseye             11.15     Official       postgres:11.15-bullseye                             3d
12.10                      12.10     Official       postgres:12.10-alpine                               3d
12.10-bullseye             12.10     Official       postgres:12.10-bullseye                             3d
12.6                       12.6      Official       postgres:12.6-alpine                                3d
12.6-debian                12.6      Official       postgres:12.6                                       3d
12.9                       12.9      Official       postgres:12.9-alpine                                3d
12.9-bullseye              12.9      Official       postgres:12.9-bullseye                              3d
12.9-bullseye-postgis      12.9      PostGIS        postgis/postgis:12-3.1                              3d
13.2                       13.2      Official       postgres:13.2-alpine                                3d
13.2-debian                13.2      Official       postgres:13.2                                       3d
13.5                       13.5      Official       postgres:13.5-alpine                                3d
13.5-bullseye              13.5      Official       postgres:13.5-bullseye                              3d
13.5-bullseye-postgis      13.5      PostGIS        postgis/postgis:13-3.1                              3d
13.6                       13.6      Official       postgres:13.6-alpine                                3d
13.6-bullseye              13.6      Official       postgres:13.6-bullseye                              3d
14.1                       14.1      Official       postgres:14.1-alpine                                3d
14.1-bullseye              14.1      Official       postgres:14.1-bullseye                              3d
14.1-bullseye-postgis      14.1      PostGIS        postgis/postgis:14-3.1                              3d
14.2                       14.2      Official       postgres:14.2-alpine                                3d
14.2-bullseye              14.2      Official       postgres:14.2-bullseye                              3d
9.6.21                     9.6.21    Official       postgres:9.6.21-alpine                              3d
9.6.21-debian              9.6.21    Official       postgres:9.6.21                                     3d
9.6.24                     9.6.24    Official       postgres:9.6.24-alpine                              3d
9.6.24-bullseye            9.6.24    Official       postgres:9.6.24-bullseye                            3d
timescaledb-2.1.0-pg11     11.11     TimescaleDB    timescale/timescaledb:2.1.0-pg11-oss                3d
timescaledb-2.1.0-pg12     12.6      TimescaleDB    timescale/timescaledb:2.1.0-pg12-oss                3d
timescaledb-2.1.0-pg13     13.2      TimescaleDB    timescale/timescaledb:2.1.0-pg13-oss                3d
timescaledb-2.5.0-pg14.1   14.1      TimescaleDB    timescale/timescaledb:2.5.0-pg14-oss                3d

```

Notice the `DEPRECATED` column. Here, `true` means that this PostgresVersion is deprecated for current KubeDB version. KubeDB will not work for deprecated PostgresVersion.

In this tutorial, we will use `13.2` PostgresVersion crd to create PostgreSQL database. To know more about what is `PostgresVersion` crd and why there is `13.2` and `13.2-debian` variation, please visit [here](/docs/guides/postgres/concepts/catalog.md). You can also see supported PostgresVersion [here](/docs/guides/postgres/README.md#supported-postgresversion-crd).

## Create a PostgreSQL database

KubeDB implements a Postgres CRD to define the specification of a PostgreSQL database.

Below is the Postgres object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: quick-postgres
  namespace: demo
spec:
  version: "13.2"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: DoNotTerminate
```

Here,

- `spec.version` is name of the PostgresVersion crd where the docker images are specified. In this tutorial, a PostgreSQL 13.2 database is created.
- `spec.storageType` specifies the type of storage that will be used for Postgres database. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create Postgres database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies the size and StorageClass of PVC that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests. If you don't specify `spec.storageType: Ephemeral`, then this field is required.
- `spec.terminationPolicy` specifies what KubeDB should do when user try to delete Postgres crd. Termination policy `DoNotTerminate` prevents a user from deleting this object if admission webhook is enabled.

>Note: `spec.storage` section is used to create PVC for database pod. It will create PVC with storage size specified in`storage.resources.requests` field. Don't specify `limits` here. PVC does not get resized automatically.

Let's create Postgres crd,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/quickstart/quick-postgres.yaml
postgres.kubedb.com/quick-postgres created
```

KubeDB operator watches for Postgres objects using Kubernetes api. When a Postgres object is created, KubeDB operator will create a new StatefulSet and two ClusterIP Service with the matching name. KubeDB operator will also create a governing service for StatefulSet with the name `kubedb`, if one is not already present.

If you are using RBAC enabled cluster, PostgreSQL specific RBAC permission is required. For details, please visit [here](/docs/guides/postgres/quickstart/rbac.md).

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created.

```bash
$  kubectl get pg -n demo quick-postgres -o wide
NAME             VERSION   STATUS     AGE
quick-postgres   13.2      Creating   13s
```

Let's describe Postgres object `quick-postgres`

```bash
$ kubectl describe -n demo postgres quick-postgres 
Name:         quick-postgres
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Postgres
Metadata:
  Creation Timestamp:  2022-05-30T09:15:36Z
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
        f:allowedSchemas:
          .:
          f:namespaces:
            .:
            f:from:
        f:storage:
          .:
          f:accessModes:
          f:resources:
            .:
            f:requests:
              .:
              f:storage:
          f:storageClassName:
        f:storageType:
        f:terminationPolicy:
        f:version:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2022-05-30T09:15:36Z
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:finalizers:
          .:
          v:"kubedb.com":
      f:spec:
        f:authSecret:
          .:
          f:name:
    Manager:      pg-operator
    Operation:    Update
    Time:         2022-05-30T09:15:37Z
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         pg-operator
    Operation:       Update
    Subresource:     status
    Time:            2022-05-30T09:16:26Z
  Resource Version:  330717
  UID:               aa9193d0-cd9b-4b63-8403-2b12ec1b04be
Spec:
  Allowed Schemas:
    Namespaces:
      From:  Same
  Auth Secret:
    Name:            quick-postgres-auth
  Client Auth Mode:  md5
  Coordinator:
    Resources:
      Limits:
        Memory:  256Mi
      Requests:
        Cpu:     200m
        Memory:  256Mi
  Leader Election:
    Election Tick:                10
    Heartbeat Tick:               1
    Maximum Lag Before Failover:  67108864
    Period:                       300ms
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Affinity:
        Pod Anti Affinity:
          Preferred During Scheduling Ignored During Execution:
            Pod Affinity Term:
              Label Selector:
                Match Labels:
                  app.kubernetes.io/instance:    quick-postgres
                  app.kubernetes.io/managed-by:  kubedb.com
                  app.kubernetes.io/name:        postgreses.kubedb.com
              Namespaces:
                demo
              Topology Key:  kubernetes.io/hostname
            Weight:          100
            Pod Affinity Term:
              Label Selector:
                Match Labels:
                  app.kubernetes.io/instance:    quick-postgres
                  app.kubernetes.io/managed-by:  kubedb.com
                  app.kubernetes.io/name:        postgreses.kubedb.com
              Namespaces:
                demo
              Topology Key:  failure-domain.beta.kubernetes.io/zone
            Weight:          50
      Container Security Context:
        Capabilities:
          Add:
            IPC_LOCK
            SYS_RESOURCE
        Privileged:    false
        Run As Group:  70
        Run As User:   70
      Resources:
        Limits:
          Memory:  1Gi
        Requests:
          Cpu:     500m
          Memory:  1Gi
      Security Context:
        Fs Group:            70
        Run As Group:        70
        Run As User:         70
      Service Account Name:  quick-postgres
  Replicas:                  1
  Ssl Mode:                  disable
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         1Gi
    Storage Class Name:  standard
  Storage Type:          Durable
  Termination Policy:    DoNotTerminate
  Version:               13.2
Status:
  Conditions:
    Last Transition Time:  2022-05-30T09:15:36Z
    Message:               The KubeDB operator has started the provisioning of Postgres: demo/quick-postgres
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2022-05-30T09:16:26Z
    Message:               All replicas are ready and in Running state
    Observed Generation:   2
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2022-05-30T09:16:26Z
    Message:               The PostgreSQL: demo/quick-postgres is accepting client requests.
    Observed Generation:   2
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2022-05-30T09:16:26Z
    Message:               DB is ready because of server getting Online and Running state
    Observed Generation:   2
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2022-05-30T09:16:26Z
    Message:               The PostgreSQL: demo/quick-postgres is successfully provisioned.
    Observed Generation:   2
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Observed Generation:     2
  Phase:                   Ready
Events:
  Type    Reason      Age   From               Message
  ----    ------      ----  ----               -------
  Normal  Successful  106s  Postgres operator  Successfully created governing service
  Normal  Successful  106s  Postgres operator  Successfully created Service
  Normal  Successful  105s  Postgres operator  Successfully created appbinding
```

KubeDB has created two services for the Postgres object.

```bash
$ kubectl get service -n demo --selector=app.kubernetes.io/name=postgreses.kubedb.com,app.kubernetes.io/instance=quick-postgres
NAME                  TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)                      AGE
quick-postgres        ClusterIP   10.96.52.28   <none>        5432/TCP,2379/TCP            3m19s
quick-postgres-pods   ClusterIP   None          <none>        5432/TCP,2380/TCP,2379/TCP   3m19s


```

Here,

- Service *`quick-postgres`* targets only one Pod which is acting as *primary* server
- Service *`quick-postgres-pods`* targets all Pods created by StatefulSet

KubeDB supports PostgreSQL clustering where Pod can be either *primary* or *standby*. To learn how to configure highly available PostgreSQL cluster, click [here](/docs/guides/postgres/clustering/ha_cluster.md).

Here, we have created a PostgreSQL database with single node, *primary* only.

## Connect with PostgreSQL database

KubeDB operator has created a new Secret called `quick-postgres-auth` for storing the *username* and *password* for `postgres` database.

```yaml
 $ kubectl get secret -n demo quick-postgres-auth -o yaml
apiVersion: v1
data:
  POSTGRES_PASSWORD: REQ4aTU2VUJJY3M2M1BWTw==
  POSTGRES_USER: cG9zdGdyZXM=
kind: Secret
metadata:
  creationTimestamp: 2018-09-03T11:25:39Z
  labels:
    app.kubernetes.io/name: postgreses.kubedb.com
    app.kubernetes.io/instance: quick-postgres
  name: quick-postgres-auth
  namespace: demo
  resourceVersion: "1677"
  selfLink: /api/v1/namespaces/demo/secrets/quick-postgres-auth
  uid: 15b3e8a1-af6c-11e8-996d-0800270d7bae
type: Opaque
```

This secret contains superuser name for `postgres` database as `POSTGRES_USER` key and
password as `POSTGRES_PASSWORD` key. By default, superuser name is `postgres` and password is randomly generated.

If you want to use custom password, please create the secret manually and specify that when creating the Postgres object using `spec.authSecret.name`. For more details see [here](/docs/guides/postgres/concepts/postgres.md#specdatabasesecret).

> Note: Auth Secret name format: `{postgres-name}-auth`

Now, you can connect to this database from the pgAdmin dashboard using `quick-postgres.demo` service and *username* and *password* created in `quick-postgres-auth` secret.

**Connection information:**

- Host name/address: you can use any of these
  - Service: `quick-postgres.demo`
  - Pod IP: (`$ kubectl get pods quick-postgres-0 -n demo -o yaml | grep podIP`)
- Port: `5432`
- Maintenance database: `postgres`

- Username: Run following command to get *username*,

  ```bash
  $ kubectl get secrets -n demo quick-postgres-auth -o jsonpath='{.data.\POSTGRES_USER}' | base64 -d
  postgres
  ```

- Password: Run the following command to get *password*,

  ```bash
  $ kubectl get secrets -n demo quick-postgres-auth -o jsonpath='{.data.\POSTGRES_PASSWORD}' | base64 -d
  DD8i56UBIcs63PVO
  ```

Now, go to pgAdmin dashboard and connect to the database using the connection information as shown below,

<p align="center">
  <kbd>
    <img alt="quick-postgres"  src="/docs/images/postgres/quick-postgres.gif">
  </kbd>
</p>

## Halt Database

KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` termination policy. If admission webhook is enabled, it prevents user from deleting the database as long as the `spec.terminationPolicy` is set `DoNotTerminate`.

In this tutorial, Postgres `quick-postgres` is created with `spec.terminationPolicy: DoNotTerminate`. So if you try to delete this Postgres object, admission webhook will nullify the delete operation.

```bash
$  kubectl delete pg -n demo quick-postgres
Error from server (BadRequest): admission webhook "postgreswebhook.validators.kubedb.com" denied the request: postgres "demo/quick-postgres" can't be terminated. To delete, change spec.terminationPolicy
```

To halt the database, we have to set `spec.terminationPolicy:` to `Halt` by updating it,

```bash
$ kubectl edit pg -n demo quick-postgres
spec:
  terminationPolicy: Halt
```

Now, if you delete the Postgres object, the KubeDB operator will delete every resource created for this Elasticsearch CR, but leaves the auth secrets, and PVCs.

Let's delete the Postgres object,

```bash
$ kubectl delete pg -n demo quick-postgres
postgres.kubedb.com "quick-postgres" deleted
```
Check resources:
```bash
$ kubectl get all,secret,pvc -n demo -l 'app.kubernetes.io/instance=quick-postgres'
NAME                         TYPE                       DATA   AGE
secret/quick-postgres-auth   kubernetes.io/basic-auth   2      27m

NAME                                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-quick-postgres-0   Bound    pvc-b30e3255-a7ea-4f61-8637-f60e283236b2   1Gi        RWO            standard       27m
```

## Resume Elasticsearch
Say, the Postgres CR was deleted with `spec.terminationPolicy` to `Halt` and you want to re-create the Postgres using the existing auth secrets and the PVCs.

You can do it by simpily re-deploying the original Postgres object:
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/quickstart/quick-postgres.yaml
postgres.kubedb.com/quick-postgres created
```
## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo pg/quick-postgres -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo pg/quick-postgres

kubectl delete ns demo
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for production environment. You can follow these tips to avoid them.

1. **Use `storageType: Ephemeral`**. Databases are precious. You might not want to lose your data in your production environment if database pod fail. So, we recommend to use `spec.storageType: Durable` and provide storage spec in `spec.storage` section. For testing purpose, you can just use `spec.storageType: Ephemeral`. KubeDB will use [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) for storage. You will not require to provide `spec.storage` section.
2. **Use `terminationPolicy: WipeOut`**. It is nice to be able to resume database from previous one. So, we create `DormantDatabase` and preserve all your `PVCs`, `Secrets`, `Snapshots` etc. If you don't want to resume database, you can just use `spec.terminationPolicy: WipeOut`. It will not create `DormantDatabase` and it will delete everything created by KubeDB for a particular Postgres crd when you delete the crd. For more details about termination policy, please visit [here](/docs/guides/postgres/concepts/postgres.md#specterminationpolicy).

## Next Steps

- Learn about [backup and restore](/docs/guides/postgres/backup/overview/index.md) PostgreSQL database using Stash.
- Learn about initializing [PostgreSQL with Script](/docs/guides/postgres/initialization/script_source.md).
- Learn about [custom PostgresVersions](/docs/guides/postgres/custom-versions/setup.md).
- Want to setup PostgreSQL cluster? Check how to [configure Highly Available PostgreSQL Cluster](/docs/guides/postgres/clustering/ha_cluster.md)
- Monitor your PostgreSQL database with KubeDB using [built-in Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md).
- Monitor your PostgreSQL database with KubeDB using [Prometheus operator](/docs/guides/postgres/monitoring/using-prometheus-operator.md).
- Detail concepts of [Postgres object](/docs/guides/postgres/concepts/postgres.md).
- Use [private Docker registry](/docs/guides/postgres/private-registry/using-private-registry.md) to deploy PostgreSQL with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
