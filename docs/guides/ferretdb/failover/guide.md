---
title: Failover & Disaster Recovery Overview FerretDB
menu:
  docs_{{ .version }}:
    identifier: fr-failover-disaster-recovery
    name: Overview
    parent: fr-failover-ferretdb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Building Resilient FerretDB Clusters on Kubernetes

## High Availability with KubeDB: Automated Failover & Recovery

In modern Kubernetes environments, downtime in a database like FerretDB isn’t just inconvenient—it can disrupt critical applications. To prevent that, KubeDB provides built-in High Availability (HA) and self-healing failover mechanisms for FerretDB.

KubeDB achieves this by running a sidecar container alongside each FerretDB pod. This sidecar continuously monitors cluster health and uses a consensus algorithm to determine the active leader. If the current leader becomes unavailable—whether due to node failure, crash, or networking issue—the KubeDB operator triggers an automated failover. A healthy replica is then promoted to leader, ensuring requests continue to flow with minimal service disruption.

The entire failover process is designed to be fast and automatic, typically completing within 2–10 seconds depending on cluster conditions. In rare edge cases, such as complex network partitions, failover may extend up to ~45 seconds, but recovery is always guaranteed without manual intervention.

This guide walks through how KubeDB manages FerretDB failover in practice. You’ll learn how to:

- Deploy a FerretDB HA cluster with KubeDB.

- Observe how leader election and failover are coordinated internally.

- Simulate a node failure and watch KubeDB’s automated recovery in action.

By the end, you’ll see how KubeDB keeps your FerretDB workloads always available, even in the face of unexpected failures.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be 
configured to communicate with your cluster. If you do not already have a cluster, you can create
one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) is required to run KubeDB. Check the available StorageClass in cluster.

  ```bash
  $ kubectl get storageclasses
  NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
  standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  2m5s

  ```
- Read [ferretdb replication concept](/docs/guides/ferretdb/clustering/replication-concept.md) to learn about FerretDB Replication clustering.
- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/mongodb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Available FerretDBVersion

When you have installed KubeDB, it has created `FerretDBVersion` CR for all supported FerretDB versions.

```bash
$ kubectl get ferretdbversions
NAME     VERSION   DB_IMAGE                                  DEPRECATED   AGE
1.18.0   1.18.0    ghcr.io/appscode-images/ferretdb:1.18.0                104m
1.23.0   1.23.0    ghcr.io/appscode-images/ferretdb:1.23.0                104m
1.24.0   1.24.0    ghcr.io/appscode-images/ferretdb:1.24.0                104m
2.0.0    2.0.0     ghcr.io/appscode-images/ferretdb:2.0.0                 5d4h
```

## Create a FerretDB database

FerretDB use Postgres as it's main backend. Currently, KubeDB supports Postgres backend as database engine for FerretDB. KubeDB operator will create and manage the backend Postgres for FerretDB

Below is the `FerretDB` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: ferret
  namespace: demo
spec:
  version: "2.0.0"
  backend:
    storage:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
  deletionPolicy: WipeOut
  server:
    primary:
      replicas: 2
    secondary:
      replicas: 2
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/quickstart/ferretdb-internal.yaml
ferretdb.kubedb.com/ferret created
```

You can monitor the status until all pods are ready:
```shell
watch kubectl get ferretdb,petset,pods -n demo
```
See the database is ready.

```shell
$ kubectl get ferretdb,petset,pods -n demo
NAME                         NAMESPACE   VERSION   STATUS   AGE
ferretdb.kubedb.com/ferret   demo        2.0.0     Ready    2m54s

NAME                                             AGE
petset.apps.k8s.appscode.com/ferret              2m1s
petset.apps.k8s.appscode.com/ferret-pg-backend   2m51s
petset.apps.k8s.appscode.com/ferret-secondary    2m1s

NAME                      READY   STATUS    RESTARTS   AGE
pod/ferret-0              1/1     Running   0          2m1s
pod/ferret-1              1/1     Running   0          2m
pod/ferret-pg-backend-0   2/2     Running   0          2m50s
pod/ferret-pg-backend-1   2/2     Running   0          2m43s
pod/ferret-pg-backend-2   2/2     Running   0          2m36s
pod/ferret-secondary-0    1/1     Running   0          2m1s
pod/ferret-secondary-1    1/1     Running   0          2m

```

Inspect who is `primary` and who is `secondary`.

```shell

$  kubectl get pods -n demo --show-labels | grep role
ferret-pg-backend-0   2/2     Running   0          3m12s   app.kubernetes.io/component=database,app.kubernetes.io/instance=ferret-pg-backend,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=postgreses.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=ferret-pg-backend-6766dfcdd7,kubedb-role=primary,kubedb.com/role=primary,statefulset.kubernetes.io/pod-name=ferret-pg-backend-0
ferret-pg-backend-1   2/2     Running   0          113s    app.kubernetes.io/component=database,app.kubernetes.io/instance=ferret-pg-backend,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=postgreses.kubedb.com,apps.kubernetes.io/pod-index=1,controller-revision-hash=ferret-pg-backend-6766dfcdd7,kubedb-role=secondary,kubedb.com/role=secondary,statefulset.kubernetes.io/pod-name=ferret-pg-backend-1
ferret-pg-backend-2   2/2     Running   0          110s    app.kubernetes.io/component=database,app.kubernetes.io/instance=ferret-pg-backend,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=postgreses.kubedb.com,apps.kubernetes.io/pod-index=2,controller-revision-hash=ferret-pg-backend-6766dfcdd7,kubedb-role=secondary,kubedb.com/role=secondary,statefulset.kubernetes.io/pod-name=ferret-pg-backend-2
```
The pod having `kubedb.com/role=primary` is the primary and `kubedb.com/role=secondary` are the secondaries.


Let's create a table in the primary.

In terminal connect with the `FerretDB` using `Postgres` client.
```shell
$ kubectl exec -it -n demo ferret-pg-backend-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)

postgres@ferret-pg-backend-0:/$ psql -U postgres
psql (17.4 (Debian 17.4-1.pgdg120+2))
Type "help" for help.

postgres=# CREATE DATABASE Kothakoli;
CREATE DATABASE
postgres=# \l
                                                     List of databases
     Name      |  Owner   | Encoding | Locale Provider |  Collate   |   Ctype    | Locale | ICU Rules |   Access privileges   
---------------+----------+----------+-----------------+------------+------------+--------+-----------+-----------------------
 ferretdb      | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 kothakoli     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 kubedb_system | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 postgres      | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 template0     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | =c/postgres          +
               |          |          |                 |            |            |        |           | postgres=CTc/postgres
 template1     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | =c/postgres          +
               |          |          |                 |            |            |        |   :

```

Verify the table creation in secondary's.

```shell
$ kubectl exec -it -n demo ferret-pg-backend-1 -- bash

Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)

postgres@ferret-pg-backend-0:/$ psql -U postgres
psql (17.4 (Debian 17.4-1.pgdg120+2))
Type "help" for help.

postgres=# \l
                                                     List of databases
     Name      |  Owner   | Encoding | Locale Provider |  Collate   |   Ctype    | Locale | ICU Rules |   Access privileges   
---------------+----------+----------+-----------------+------------+------------+--------+-----------+-----------------------
 ferretdb      | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 kothakoli     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 kubedb_system | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 postgres      | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 template0     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | =c/postgres          +
               |          |          |                 |            |            |        |           | postgres=CTc/postgres
 template1     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | =c/postgres          +
               |          |          |                 |            |            |        |   :

```
### Step 2: Simulating a Failover
Before we simulate a failover, let’s understand how it’s handled in a KubeDB-managed FerretDB cluster.
KubeDB continuously monitors the health of your database pods. If the active pod goes down, KubeDB quickly promotes another healthy pod to take over as the new primary.

This switchover usually happens in just a few seconds, so your applications experience little to no interruption.

Now current running primary is `FerretDB-ag-cluster-0`. Let's open another terminal and run the command below.

```shell
watch -n 2 "kubectl get pods -n demo -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.labels.kubedb\\.com/role}{\"\\n\"}{end}'"

```
It will show current ferretdb cluster roles.
```shell
ferret-0
ferret-1
ferret-pg-backend-0 primary
ferret-pg-backend-1 secondary
ferret-pg-backend-2 secondary
ferret-secondary-0
ferret-secondary-1

```

#### Case 1: Delete the current primary

Let's delete the current primary and see how the role change happens almost immediately.

```shell
$ kubectl delete pods -n demo ferret-pg-backend-0
pod "ferret-pg-backend-0" deleted

```
```shell
ferret-0
ferret-1
ferret-pg-backend-0 
ferret-pg-backend-1 primary
ferret-pg-backend-2 secondary
ferret-secondary-0
ferret-secondary-1

```

Now we know how failover is done, let's check if the new primary `FerretDB-ag-cluster-1` is working.

```shell
$ kubectl exec -it -n demo ferret-pg-backend-1 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
postgres@ferret-pg-backend-1:/$ psql -U postgres
psql (17.4 (Debian 17.4-1.pgdg120+2))
Type "help" for help.

postgres=# CREATE DATABASE Kathak;
CREATE DATABASE
postgres=# \l

(2 rows affected)
                                                      List of databases
     Name      |  Owner   | Encoding | Locale Provider |  Collate   |   Ctype    | Locale | ICU Rules |   Access privileges   
---------------+----------+----------+-----------------+------------+------------+--------+-----------+-----------------------
 ferretdb      | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 kathak        | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 kothakoli     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 kubedb_system | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 postgres      | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 template0     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | =c/postgres          +
               |          |          |                 |            |            |        |           | postgres=CTc/postgres
 template1     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | =c/postgres          +
:


```


You will see the deleted pod (`FerretDB-ag-cluster-0`) is brought back by the kubedb operator and it is
now assigned to `secondary` role.

```shell

ferret-0
ferret-1
ferret-pg-backend-0 secondary
ferret-pg-backend-1 primary
ferret-pg-backend-2 secondary
ferret-secondary-0
ferret-secondary-1

```

Let's check if the secondary(`FerretDB-ag-cluster-0`) got the updated data from new primary `FerretDB-ag-cluster-2`.

```shell
$ kubectl exec -it -n demo ferret-pg-backend-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
postgres@ferret-pg-backend-0:/$ psql -U postgres
psql (17.4 (Debian 17.4-1.pgdg120+2))
Type "help" for help.

postgres=# \l

(2 rows affected)
                                                      List of databases
     Name      |  Owner   | Encoding | Locale Provider |  Collate   |   Ctype    | Locale | ICU Rules |   Access privileges   
---------------+----------+----------+-----------------+------------+------------+--------+-----------+-----------------------
 ferretdb      | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 kathak        | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 kothakoli     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 kubedb_system | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 postgres      | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 template0     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | =c/postgres          +
               |          |          |                 |            |            |        |           | postgres=CTc/postgres
 template1     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | =c/postgres          +
:

```

#### Case 2: Delete the current primary and one secondary
```shell
$ kubectl delete pods -n demo ferret-pg-backend-0 ferret-pg-backend-1
pod "ferret-pg-backend-0" deleted
pod "ferret-pg-backend-1" deleted

```
Again we can see the failover happened pretty quickly.
```shell

ferret-0
ferret-1
ferret-pg-backend-0 
ferret-pg-backend-1 
ferret-pg-backend-2 primary
ferret-secondary-0
ferret-secondary-1

```

After 10-20 second, the deleted pods will be back and will have its role.

```shell
ferret-0
ferret-1
ferret-pg-backend-0 secondary
ferret-pg-backend-1 secondary
ferret-pg-backend-2 primary
ferret-secondary-0
ferret-secondary-1

```

Let's validate the cluster state from new primary(`FerretDB-ag-cluster-2`).

```shell
$ kubectl exec -it -n demo ferret-pg-backend-2 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
postgres@ferret-pg-backend-2:/$ psql -U postgres
psql (17.4 (Debian 17.4-1.pgdg120+2))
Type "help" for help.

postgres=# \l

(2 rows affected)
                                                      List of databases
     Name      |  Owner   | Encoding | Locale Provider |  Collate   |   Ctype    | Locale | ICU Rules |   Access privileges   
---------------+----------+----------+-----------------+------------+------------+--------+-----------+-----------------------
 ferretdb      | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 kathak        | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 kothakoli     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 kubedb_system | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 postgres      | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | 
 template0     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | =c/postgres          +
               |          |          |                 |            |            |        |           | postgres=CTc/postgres
 template1     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |        |           | =c/postgres          +
:


```

#### Case3: Delete any of the replica's

Let's delete both of the secondary's.

```shell
$ kubectl delete pods -n demo ferret-pg-backend-0 ferret-pg-backend-1
pod "ferret-pg-backend-0" deleted
pod "ferret-pg-backend-1" deleted

```
```shell
ferret-0
ferret-1
ferret-pg-backend-0 
ferret-pg-backend-1
ferret-pg-backend-2 primary
ferret-secondary-0
ferret-secondary-1

```

Shortly both of the pods will be back with its role.
```shell
ferret-0
ferret-1
ferret-pg-backend-0 secondary
ferret-pg-backend-1 secondary
ferret-pg-backend-2 primary
ferret-secondary-0
ferret-secondary-1

```


#### Case 4: Delete both primary and all replicas

Let's delete all the pods.

```shell
$ kubectl delete pods -n demo FerretDB-ag-cluster-0 FerretDB-ag-cluster-1 FerretDB-ag-cluster-2
pod "FerretDB-ag-cluster-0" deleted
pod "FerretDB-ag-cluster-1" deleted
pod "FerretDB-ag-cluster-2" deleted

```
```shell
ferret-0
ferret-1
ferret-pg-backend-0 
ferret-pg-backend-1
ferret-pg-backend-2 
ferret-secondary-0
ferret-secondary-1

```

Within 20-30 second, all  the pod should be back.
```shell
ferret-0
ferret-1
ferret-pg-backend-0 secondary
ferret-pg-backend-1 secondary
ferret-pg-backend-2 primary
ferret-secondary-0
ferret-secondary-1

```

## CleanUp
If you want to clean up each of the Kubernetes resources created by this tutorial, run:
```shell
$  kubectl delete -n demo fr/ferret
ferretdb.kubedb.com "ferret" deleted

$ kubectl delete ns demo
```


## Next Steps

- Want to setup FerretDB cluster? Check how to [configure Highly Available FerretDB Replication](/docs/guides/ferretdb/clustering/replication.md)
- Monitor your FerretDB database with KubeDB using [Prometheus operator](/docs/guides/ferretdb/monitoring/using-prometheus-operator.md).
- Detail concepts of [FerretDB object](/docs/guides/ferretdb/concepts/ferretdb.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).