---
title: Horizontal Scaling MariaDB
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-scaling-horizontal-cluster
    name: Cluster
    parent: guides-mariadb-scaling-horizontal
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale MariaDB

This guide will show you how to use `KubeDB` Enterprise operator to scale the cluster of a MariaDB database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MariaDB](/docs/guides/mariadb/concepts/mariadb/)
  - [MariaDB Cluster](/docs/guides/mariadb/clustering/galera-cluster/)
  - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/opsrequest/)
  - [Horizontal Scaling Overview](/docs/guides/mariadb/scaling/horizontal-scaling/overview/)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Apply Horizontal Scaling on Cluster

Here, we are going to deploy a  `MariaDB` cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare MariaDB Cluster Database

Now, we are going to deploy a `MariaDB` cluster with version `10.5.23`.

### Deploy MariaDB Cluster

In this section, we are going to deploy a MariaDB cluster. Then, in the next section we will scale the database using `MariaDBOpsRequest` CRD. Below is the YAML of the `MariaDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "10.5.23"
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

Let's create the `MariaDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/scaling/horizontal-scaling/cluster/example/sample-mariadb.yaml
mariadb.kubedb.com/sample-mariadb created
```

Now, wait until `sample-mariadb` has status `Ready`. i.e,

```bash
$ kubectl get mariadb -n demo
NAME             VERSION   STATUS   AGE
sample-mariadb   10.5.23    Ready    2m36s
```

Let's check the number of replicas this database has from the MariaDB object, number of pods the statefulset have,

```bash
$ kubectl get mariadb -n demo sample-mariadb -o json | jq '.spec.replicas'
3
$ kubectl get sts -n demo sample-mariadb -o json | jq '.spec.replicas'
3
```

We can see from both command that the database has 3 replicas in the cluster.

Also, we can verify the replicas of the replicaset from an internal mariadb command by execing into a replica.

First we need to get the username and password to connect to a mariadb instance,
```bash
$ kubectl get secrets -n demo sample-mariadb-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo sample-mariadb-auth -o jsonpath='{.data.\password}' | base64 -d
nrKuxni0wDSMrgwy
```

Now let's connect to a mariadb instance and run a mariadb internal command to check the number of replicas,

```bash
$  kubectl exec -it -n demo sample-mariadb-0 -c mariadb -- bash
root@sample-mariadb-0:/ mysql -uroot -p$MYSQL_ROOT_PASSWORD -e "show status like 'wsrep_cluster_size';"
+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| wsrep_cluster_size | 3     |
+--------------------+-------+

```

We can see from the above output that the cluster has 3 nodes.

We are now ready to apply the `MariaDBOpsRequest` CR to scale this database.

## Scale Up Replicas

Here, we are going to scale up the replicas of the replicaset to meet the desired number of replicas after scaling.

#### Create MariaDBOpsRequest

In order to scale up the replicas of the replicaset of the database, we have to create a `MariaDBOpsRequest` CR with our desired replicas. Below is the YAML of the `MariaDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: mdops-scale-horizontal-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: sample-mariadb
  horizontalScaling:
    member : 5
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `sample-mariadb` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.member` specifies the desired replicas after scaling.

Let's create the `MariaDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/scaling/horizontal-scaling/cluster/example/mdops-upscale.yaml
mariadbopsrequest.ops.kubedb.com/mdops-scale-horizontal-up created
```

#### Verify Cluster replicas scaled up successfully 

If everything goes well, `KubeDB` Enterprise operator will update the replicas of `MariaDB` object and related `StatefulSets` and `Pods`.

Let's wait for `MariaDBOpsRequest` to be `Successful`.  Run the following command to watch `MariaDBOpsRequest` CR,

```bash
$ watch kubectl get mariadbopsrequest -n demo
Every 2.0s: kubectl get mariadbopsrequest -n demo
NAME                        TYPE                STATUS       AGE
mdps-scale-horizontal    HorizontalScaling    Successful     106s
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. Now, we are going to verify the number of replicas this database has from the MariaDB object, number of pods the statefulset have,

```bash
$ kubectl get mariadb -n demo sample-mariadb -o json | jq '.spec.replicas'
5
$ kubectl get sts -n demo sample-mariadb -o json | jq '.spec.replicas'
5
```

Now let's connect to a mariadb instance and run a mariadb internal command to check the number of replicas,

```bash
$ $  kubectl exec -it -n demo sample-mariadb-0 -c mariadb -- bash
root@sample-mariadb-0:/ mysql -uroot -p$MYSQL_ROOT_PASSWORD -e "show status like 'wsrep_cluster_size';"
+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| wsrep_cluster_size | 5     |
+--------------------+-------+
```

From all the above outputs we can see that the replicas of the cluster is `5`. That means we have successfully scaled up the replicas of the MariaDB replicaset.

### Scale Down Replicas

Here, we are going to scale down the replicas of the cluster to meet the desired number of replicas after scaling.

#### Create MariaDBOpsRequest

In order to scale down the cluster of the database, we have to create a `MariaDBOpsRequest` CR with our desired replicas. Below is the YAML of the `MariaDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: mdops-scale-horizontal-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: sample-mariadb
  horizontalScaling:
    member : 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down operation on `sample-mariadb` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.replicas` specifies the desired replicas after scaling.

Let's create the `MariaDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/scaling/horizontal-scaling/cluster/example/mdops-downscale.yaml
mariadbopsrequest.ops.kubedb.com/mdops-scale-horizontal-down created
```

#### Verify Cluster replicas scaled down successfully 

If everything goes well, `KubeDB` Enterprise operator will update the replicas of `MariaDB` object and related `StatefulSets` and `Pods`.

Let's wait for `MariaDBOpsRequest` to be `Successful`.  Run the following command to watch `MariaDBOpsRequest` CR,

```bash
$ watch kubectl get mariadbopsrequest -n demo
Every 2.0s: kubectl get mariadbopsrequest -n demo
NAME                          TYPE                STATUS       AGE
mops-hscale-down-replicaset   HorizontalScaling   Successful   2m32s
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. Now, we are going to verify the number of replicas this database has from the MariaDB object, number of pods the statefulset have,

```bash
$ kubectl get mariadb -n demo sample-mariadb -o json | jq '.spec.replicas' 
3
$ kubectl get sts -n demo sample-mariadb -o json | jq '.spec.replicas'
3
```

Now let's connect to a mariadb instance and run a mariadb internal command to check the number of replicas,
```bash
$ $  kubectl exec -it -n demo sample-mariadb-0 -c mariadb -- bash
root@sample-mariadb-0:/ mysql -uroot -p$MYSQL_ROOT_PASSWORD -e "show status like 'wsrep_cluster_size';"
+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| wsrep_cluster_size | 5     |
+--------------------+-------+
```

From all the above outputs we can see that the replicas of the cluster is `5`. That means we have successfully scaled down the replicas of the MariaDB replicaset.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete mariadb -n demo sample-mariadb
$ kubectl delete mariadbopsrequest -n demo  mdops-scale-horizontal-up mdops-scale-horizontal-down
```