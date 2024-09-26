---
title: Reconfigure Postgres Cluster
menu:
  docs_{{ .version }}:
    identifier: pg-reconfigure-cluster
    name: Cluster
    parent: pg-reconfigure
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---


> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Postgres Cluster Database

This guide will show you how to use `KubeDB` Enterprise operator to reconfigure a Postgres Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Postgres](/docs/guides/postgres/concepts/postgres.md)
  - [PostgresOpsRequest](/docs/guides/postgres/concepts/opsrequest.md)
  - [Reconfigure Overview](/docs/guides/postgres/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

Now, we are going to deploy a  `Postgres` Cluster using a supported version by `KubeDB` operator. Then we are going to apply `PostgresOpsRequest` to reconfigure its configuration.

### Prepare Postgres Cluster

Now, we are going to deploy a `Postgres` Cluster database with version `16.1`.

### Deploy Postgres

At first, we will create `user.conf` file containing required configuration settings.
To know more about this configuration file, check [here](/docs/guides/postgres/configuration/using-config-file.md)
```ini
$ cat user.conf
max_connections=200
shared_buffers=256MB
```

Now, we will create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo pg-configuration --from-file=./user.conf
secret/pg-configuration created
```

In this section, we are going to create a Postgres object specifying `spec.configSecret` field to apply this custom configuration. Below is the YAML of the `Postgres` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: ha-postgres
  namespace: demo
spec:
  version: "16.1"
  replicas: 3
  configSecret:
    name: pg-configuration
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Postgres` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/reconfigure/ha-postgres.yaml
postgres.kubedb.com/ha-postgres created
```

Now, wait until `ha-postgres` has status `Ready`. i.e,

```bash
$ kubectl get pods,pg -n demo
NAME                READY   STATUS    RESTARTS   AGE
pod/ha-postgres-0   2/2     Running   0          2m28s
pod/ha-postgres-1   2/2     Running   0          59s
pod/ha-postgres-2   2/2     Running   0          51s

NAME                              VERSION   STATUS   AGE
postgres.kubedb.com/ha-postgres   16.1      Ready    2m38s

```

Now lets check these parameters,
```bash
$ kubectl exec -it -n demo ha-postgres-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
ha-postgres-0:/$ psql
psql (16.1)
Type "help" for help.

postgres=# show max_connections;
 max_connections 
-----------------
 200
(1 row)

postgres=# show shared_buffers;
 shared_buffers 
----------------
 256MB
(1 row)
```
You can check the other pods same way.
So we have configured custom parameters.
### Reconfigure using new config secret

Now we will reconfigure this database to set `max_connections` to `250`.

Now, we will create new file `user.conf` containing required configuration settings.

```ini
$ cat user.conf 
max_connections = 250
```

Then, we will create a new secret with this configuration file.

```bash
$ kubectl create secret generic -n demo new-pg-configuration --from-file=./user.conf
secret/new-pg-configuration created
```

#### Create PostgresOpsRequest

Now, we will use this secret to replace the previous secret using a `PostgresOpsRequest` CR. The `PostgresOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pgops-reconfigure-config
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: ha-postgres
  configuration:   
    configSecret:
      name: new-pg-configuration
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `ha-postgres` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.configSecret.name` specifies the name of the new secret.

Let's create the `PostgresOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/reconfigure/reconfigure-using-secret.yaml
postgresopsrequest.ops.kubedb.com/pgops-reconfigure-config created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Enterprise operator will update the `configSecret` of `Postgres` object.

Let's wait for `PostgresOpsRequest` to be `Successful`.  Run the following command to watch `PostgresOpsRequest` CR,

```bash
$ kubectl get pgops -n demo
NAME                       TYPE          STATUS       AGE
pgops-reconfigure-config   Reconfigure   Successful   3m21s
```

We can see from the above output that the `PostgresOpsRequest` has succeeded.
Now let's connect to a postgres instance and run a postgres internal command to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo ha-postgres-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
ha-postgres-0:/$ psql
psql (16.1)
Type "help" for help.

postgres=# show max_connections;
 max_connections 
-----------------
 250
(1 row)

```

As we can see from the configuration has changed, the value of `max_connections` has been changed from `200` to `250`.
You can check for other pods in the same way.

### Reconfigure Existing Config Secret

Now, we will create a new `PostgresOpsRequest` to reconfigure our existing secret `new-pg-configuration` by modifying our `user.conf` file using `applyConfig`. The `PostgresOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pgops-reconfigure-apply-config
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: ha-postgres
  configuration:
    applyConfig:
      user.conf: |
        max_connections = 230
        shared_buffers = 512MB
```
> Note: You can modify multiple fields of your current configuration using `applyConfig`. If you don't have any secrets then `applyConfig` will create a secret for you. Here, we modified value of our two existing fields which are `max_connections` and `shared_buffers`.

Here,
- `spec.databaseRef.name` specifies that we are reconfiguring `ha-postgres` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.applyConfig` contains the configuration of existing or newly created secret.


```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/reconfigure/apply-config.yaml
postgresopsrequest.ops.kubedb.com/pgops-reconfigure-apply-config created
```


#### Verify the new configuration is working

If everything goes well, `KubeDB` Enterprise operator will update the `configSecret` of `Postgres` object.

Let's wait for `PostgresOpsRequest` to be `Successful`.  Run the following command to watch `PostgresOpsRequest` CR,

```bash
$ kubectl get postgresopsrequest pgops-reconfigure-apply-config -n demo
NAME            TYPE          STATUS       AGE
apply-config   Reconfigure   Successful   4m59s
```
We can see this ops request was successful.

Now let's connect to a postgres instance and run a postgres internal command to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo ha-postgres-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
ha-postgres-0:/$ cat /etc/config/user.conf 
#________******kubedb.com/inline-config******________#
max_connections=230
shared_buffers=512MB
ha-postgres-0:/$ 
ha-postgres-0:/$ 
ha-postgres-0:/$ psql
psql (16.1)
Type "help" for help.

postgres=# show max_connections;
 max_connections 
-----------------
 230
(1 row)

postgres=# show shared_buffers;
 shared_buffers 
----------------
 512MB
(1 row)

```

As we can see from above the configuration has been changed, the value of `max_connections` has been changed from `250` to `230` and the `shared_buffers` has been changed `256MB` to `512MB`.


### Remove Custom Configuration

We can also remove exisiting custom config using `PostgresOpsRequest`. Provide `true` to field `spec.configuration.removeCustomConfig` and make an Ops Request to remove existing custom configuration.

#### Create PostgresOpsRequest

Lets create an `PostgresOpsRequest` having `spec.configuration.removeCustomConfig` is equal `true`,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: remove-config
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: ha-postgres
  configuration:   
    removeCustomConfig: true
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `remove-config` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.removeCustomConfig` is a bool field that should be `true` when you want to remove existing custom configuration.

Let's create the `PostgresOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/reconfigure/remove-config.yaml
postgresopsrequest.ops.kubedb.com/remove-config created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Enterprise operator will update the `configSecret` of `Postgres` object.

Let's wait for `PostgresOpsRequest` to be `Successful`.  Run the following command to watch `PostgresOpsRequest` CR,

```bash
$ kubectl get pgops -n demo remove-config 
NAME            TYPE          STATUS       AGE
remove-config   Reconfigure   Successful   5m5s

```

Now let's connect to a postgres instance and run a postgres internal command to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo ha-postgres-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
ha-postgres-0:/$ psql
psql (16.1)
Type "help" for help.

postgres=# show max_connections;
 max_connections 
-----------------
 100
(1 row)

postgres=# show shared_buffers;
 shared_buffers 
----------------
 256MB
(1 row)

```

As we can see from the configuration has changed to its default value. So removal of existing custom configuration using `PostgresOpsRequest` is successful.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete postgres -n demo ha-postgres
$ kubectl delete postgresopsrequest -n demo pgops-reconfigure-apply-config pgops-reconfigure-config remove-config
$ kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Postgres object](/docs/guides/postgres/concepts/postgres.md).
- Monitor your Postgres database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/postgres/monitoring/using-prometheus-operator.md).
- Monitor your Postgres database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/postgres/private-registry/using-private-registry.md) to deploy Postgres with KubeDB.
- Use [kubedb cli](/docs/guides/postgres/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [Postgres object](/docs/guides/postgres/concepts/postgres.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
