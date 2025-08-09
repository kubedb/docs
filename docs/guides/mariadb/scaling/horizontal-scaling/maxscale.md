---
title: Horizontal Scaling MaxScale
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-scaling-horizontal-maxscale
    name: Maxscale
    parent: guides-mariadb-scaling-horizontal
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale MaxScale

This guide will show you how to use `KubeDB` Enterprise operator to scale MaxScale server.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [MariaDB](/docs/guides/mariadb/concepts/mariadb/)
    - [MariaDB Replication](/docs/guides/mariadb/clustering/mariadb-replication)
    - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/opsrequest/)
    - [Horizontal Scaling Overview](/docs/guides/mariadb/scaling/horizontal-scaling/overview/)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Apply Horizontal Scaling on MaxScale Server

Here, we are going to deploy a  `MariaDB` cluster in replication mode using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on `MaxScale` server.

### Deploy MariaDB Cluster

In this section, we are going to deploy a MariaDB cluster in replication mode. Below is the YAML of the `MariaDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: md-replication
  namespace: demo
spec:
  version: "10.5.23"
  replicas: 3
  topology:
    mode: MariaDBReplication
    maxscale:
      replicas: 3
      enableUI: true
      storageType: Durable
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 50Mi
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `MariaDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/scaling/md-replication.yaml
mariadb.kubedb.com/md-replication created
```

Now, wait until `md-replication` has status `Ready`. i.e,

```bash
$ kubectl get mariadb -n demo
NAME             VERSION   STATUS   AGE
md-replication   10.5.23   Ready    2m8s
```

Let's check the number of replicas `Maxscale` has from the MariaDB object, also the number of replicas the petset have,

```bash
$ kubectl get mariadb -n demo md-replication -o json | jq '.spec.topology.maxscale.replicas'
3
$ kubectl get petset -n demo md-replication-mx -o json | jq '.spec.replicas'
3
```

We can see from both command that the `MaxScale` has 3 replicas in the cluster.

## Scale Up Replicas

Here, we are going to scale up the replicas of the replicaset to meet the desired number of replicas after scaling.

#### Create MariaDBOpsRequest

In order to scale up the replicas of the replicaset of the `MaxScale` server, we have to create a `MariaDBOpsRequest` CR with our desired replicas. Below is the YAML of the `MariaDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: maxscale-horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: md-replication
  horizontalScaling:
    maxscale: true
    member: 4
```

Here,

- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `md-replication` database.
- `spec.horizontalScaling.maxscale` specifies that we are performing horizontal scaling operation on maxscale server. If false then horizontal scaling performs on mariadb database.
- `spec.horizontalScaling.member` specifies the desired replicas after scaling.

Let's create the `MariaDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/scaling/horizontal-scaling/mx-hscale-up.yaml
mariadbopsrequest.ops.kubedb.com/maxscale-horizontal-scale-up created
```

#### Verify Cluster replicas scaled up successfully

If everything goes well, `KubeDB` Enterprise operator will update the replicas of `MaxScale` object and related `PetSets` and `Pods`.

Let's wait for `MariaDBOpsRequest` to be `Successful`.  Run the following command to watch `MariaDBOpsRequest` CR,

```bash
$ watch kubectl get mariadbopsrequest -n demo
Every 2.0s: kubectl get mariadbopsrequest -n demo
NAME                           TYPE                STATUS       AGE
maxscale-horizontal-scale-up   HorizontalScaling   Successful   2m31s
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. Now, we are going to verify the number of replicas this database has from the MariaDB object, number of pods the petset have,

```bash
$ kubectl get mariadb -n demo md-replication -o json | jq '.spec.topology.maxscale.replicas'
4
$ kubectl get petset -n demo md-replication-mx -o json | jq '.spec.replicas'
4 
```

From all the above outputs we can see that the replicas of the `MaxScale` server is `4`. That means we have successfully scaled up the replicas of the MariaDB replicaset.

### Scale Down Replicas

Here, we are going to scale down the replicas of the cluster to meet the desired number of replicas after scaling.

#### Create MariaDBOpsRequest

In order to scale down the replicas of the `MaxScale` server, we have to create a `MariaDBOpsRequest` CR with our desired replicas. Below is the YAML of the `MariaDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: maxscale-horizontal-scale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: md-replication
  horizontalScaling:
    maxscale: true
    member: 3
```

Here,

- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `md-replication` database.
- `spec.horizontalScaling.maxscale` specifies that we are performing horizontal scaling operation on maxscale server. If false then horizontal scaling performs on mariadb database.
- `spec.horizontalScaling.member` specifies the desired replicas after scaling.

Let's create the `MariaDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/scaling/horizontal-scaling/mx-hscale-down.yaml
mariadbopsrequest.ops.kubedb.com/maxscale-horizontal-scale-down created
```

#### Verify Cluster replicas scaled down successfully

If everything goes well, `KubeDB` Enterprise operator will update the replicas of `MaxScale` object and related `PetSets` and `Pods`.

Let's wait for `MariaDBOpsRequest` to be `Successful`.  Run the following command to watch `MariaDBOpsRequest` CR,

```bash
$ watch kubectl get mariadbopsrequest -n demo
Every 2.0s: kubectl get mariadbopsrequest -n demo
NAME                             TYPE                STATUS       AGE
maxscale-horizontal-scale-down   HorizontalScaling   Successful   55s
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. Now, we are going to verify the number of replicas `MaxScale` server has from the MariaDB object, number of pods the petset have,

```bash
$ kubectl get mariadb -n demo md-replication -o json | jq '.spec.topology.maxscale.replicas'
3
$ kubectl get petset -n demo md-replication-mx -o json | jq '.spec.replicas'
3
```

From all the above outputs we can see that the replicas of the cluster is `3`. That means we have successfully scaled down the replicas of the MariaDB replicaset.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete mariadb -n demo md-replication
$ kubectl delete mariadbopsrequest -n demo  maxscale-horizontal-scale-up maxscale-horizontal-scale-down
$ kubectl delete ns demo
```