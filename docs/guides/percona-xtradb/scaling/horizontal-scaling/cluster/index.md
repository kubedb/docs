---
title: Horizontal Scaling PerconaXtraDB
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-scaling-horizontal-cluster
    name: Cluster
    parent: guides-perconaxtradb-scaling-horizontal
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale PerconaXtraDB

This guide will show you how to use `KubeDB` Enterprise operator to scale the cluster of a PerconaXtraDB database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [PerconaXtraDB](/docs/guides/percona-xtradb/concepts/perconaxtradb/)
  - [PerconaXtraDB Cluster](/docs/guides/percona-xtradb/clustering/galera-cluster/)
  - [PerconaXtraDBOpsRequest](/docs/guides/percona-xtradb/concepts/opsrequest/)
  - [Horizontal Scaling Overview](/docs/guides/percona-xtradb/scaling/horizontal-scaling/overview/)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Apply Horizontal Scaling on Cluster

Here, we are going to deploy a  `PerconaXtraDB` cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare PerconaXtraDB Cluster Database

Now, we are going to deploy a `PerconaXtraDB` cluster with version `8.0.26`.

### Deploy PerconaXtraDB Cluster

In this section, we are going to deploy a PerconaXtraDB cluster. Then, in the next section we will scale the database using `PerconaXtraDBOpsRequest` CRD. Below is the YAML of the `PerconaXtraDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: sample-pxc
  namespace: demo
spec:
  version: "8.0.26"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Let's create the `PerconaXtraDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/scaling/horizontal-scaling/cluster/example/sample-pxc.yaml
perconaxtradb.kubedb.com/sample-pxc created
```

Now, wait until `sample-pxc` has status `Ready`. i.e,

```bash
$ kubectl get perconaxtradb -n demo
NAME             VERSION   STATUS   AGE
sample-pxc       8.0.26    Ready    2m36s
```

Let's check the number of replicas this database has from the PerconaXtraDB object, number of pods the statefulset have,

```bash
$ kubectl get perconaxtradb -n demo sample-pxc -o json | jq '.spec.replicas'
3
$ kubectl get sts -n demo sample-pxc -o json | jq '.spec.replicas'
3
```

We can see from both command that the database has 3 replicas in the cluster.

Also, we can verify the replicas of the replicaset from an internal perconaxtradb command by execing into a replica.

First we need to get the username and password to connect to a perconaxtradb instance,
```bash
$ kubectl get secrets -n demo sample-pxc-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo sample-pxc-auth -o jsonpath='{.data.\password}' | base64 -d
nrKuxni0wDSMrgwy
```

Now let's connect to a perconaxtradb instance and run a perconaxtradb internal command to check the number of replicas,

```bash
$  kubectl exec -it -n demo sample-pxc-0 -c perconaxtradb -- bash
root@sample-pxc-0:/ mysql -uroot -p$MYSQL_ROOT_PASSWORD -e "show status like 'wsrep_cluster_size';"
+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| wsrep_cluster_size | 3     |
+--------------------+-------+

```

We can see from the above output that the cluster has 3 nodes.

We are now ready to apply the `PerconaXtraDBOpsRequest` CR to scale this database.

## Scale Up Replicas

Here, we are going to scale up the replicas of the replicaset to meet the desired number of replicas after scaling.

#### Create PerconaXtraDBOpsRequest

In order to scale up the replicas of the replicaset of the database, we have to create a `PerconaXtraDBOpsRequest` CR with our desired replicas. Below is the YAML of the `PerconaXtraDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: pxops-scale-horizontal-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: sample-pxc
  horizontalScaling:
    member : 5
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `sample-pxc` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.member` specifies the desired replicas after scaling.

Let's create the `PerconaXtraDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/scaling/horizontal-scaling/cluster/example/pxops-upscale.yaml
perconaxtradbopsrequest.ops.kubedb.com/pxops-scale-horizontal-up created
```

#### Verify Cluster replicas scaled up successfully 

If everything goes well, `KubeDB` Enterprise operator will update the replicas of `PerconaXtraDB` object and related `StatefulSets` and `Pods`.

Let's wait for `PerconaXtraDBOpsRequest` to be `Successful`.  Run the following command to watch `PerconaXtraDBOpsRequest` CR,

```bash
$ watch kubectl get perconaxtradbopsrequest -n demo
Every 2.0s: kubectl get perconaxtradbopsrequest -n demo
NAME                        TYPE                STATUS       AGE
pxps-scale-horizontal    HorizontalScaling    Successful     106s
```

We can see from the above output that the `PerconaXtraDBOpsRequest` has succeeded. Now, we are going to verify the number of replicas this database has from the PerconaXtraDB object, number of pods the statefulset have,

```bash
$ kubectl get perconaxtradb -n demo sample-pxc -o json | jq '.spec.replicas'
5
$ kubectl get sts -n demo sample-pxc -o json | jq '.spec.replicas'
5
```

Now let's connect to a perconaxtradb instance and run a perconaxtradb internal command to check the number of replicas,

```bash
$ $  kubectl exec -it -n demo sample-pxc-0 -c perconaxtradb -- bash
root@sample-pxc-0:/ mysql -uroot -p$MYSQL_ROOT_PASSWORD -e "show status like 'wsrep_cluster_size';"
+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| wsrep_cluster_size | 5     |
+--------------------+-------+
```

From all the above outputs we can see that the replicas of the cluster is `5`. That means we have successfully scaled up the replicas of the PerconaXtraDB replicaset.

### Scale Down Replicas

Here, we are going to scale down the replicas of the cluster to meet the desired number of replicas after scaling.

#### Create PerconaXtraDBOpsRequest

In order to scale down the cluster of the database, we have to create a `PerconaXtraDBOpsRequest` CR with our desired replicas. Below is the YAML of the `PerconaXtraDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: pxops-scale-horizontal-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: sample-pxc
  horizontalScaling:
    member : 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down operation on `sample-pxc` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.replicas` specifies the desired replicas after scaling.

Let's create the `PerconaXtraDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/scaling/horizontal-scaling/cluster/example/pxops-downscale.yaml
perconaxtradbopsrequest.ops.kubedb.com/pxops-scale-horizontal-down created
```

#### Verify Cluster replicas scaled down successfully 

If everything goes well, `KubeDB` Enterprise operator will update the replicas of `PerconaXtraDB` object and related `StatefulSets` and `Pods`.

Let's wait for `PerconaXtraDBOpsRequest` to be `Successful`.  Run the following command to watch `PerconaXtraDBOpsRequest` CR,

```bash
$ watch kubectl get perconaxtradbopsrequest -n demo
Every 2.0s: kubectl get perconaxtradbopsrequest -n demo
NAME                          TYPE                STATUS       AGE
mops-hscale-down-replicaset   HorizontalScaling   Successful   2m32s
```

We can see from the above output that the `PerconaXtraDBOpsRequest` has succeeded. Now, we are going to verify the number of replicas this database has from the PerconaXtraDB object, number of pods the statefulset have,

```bash
$ kubectl get perconaxtradb -n demo sample-pxc -o json | jq '.spec.replicas' 
3
$ kubectl get sts -n demo sample-pxc -o json | jq '.spec.replicas'
3
```

Now let's connect to a perconaxtradb instance and run a perconaxtradb internal command to check the number of replicas,
```bash
$ $  kubectl exec -it -n demo sample-pxc-0 -c perconaxtradb -- bash
root@sample-pxc-0:/ mysql -uroot -p$MYSQL_ROOT_PASSWORD -e "show status like 'wsrep_cluster_size';"
+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| wsrep_cluster_size | 5     |
+--------------------+-------+
```

From all the above outputs we can see that the replicas of the cluster is `5`. That means we have successfully scaled down the replicas of the PerconaXtraDB replicaset.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete perconaxtradb -n demo sample-pxc
$ kubectl delete perconaxtradbopsrequest -n demo  pxops-scale-horizontal-up pxops-scale-horizontal-down
```