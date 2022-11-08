---
title: Horizontal Scaling ProxySQL
menu:
docs_{{ .version }}:
identifier: guides-proxysql-scaling-horizontal-cluster
name: Cluster
parent: guides-proxysql-scaling-horizontal
weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Horizontal Scale ProxySQL

This guide will show you how to use `KubeDB` Enterprise operator to scale the cluster of a ProxySQL server.

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

Also we need a mysql backend for the proxysql server. So we are  creating one with the below yaml. 


```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql-server
  namespace: demo
spec:
  version: "5.7.36"
  replicas: 3
  topology:
    mode: GroupReplication
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

After applying the above yaml wait for the MySQL to be Ready.

## Apply Horizontal Scaling on Cluster

Here, we are going to deploy a  `ProxySQL` cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare ProxySQL Cluster

Now, we are going to deploy a `ProxySQL` cluster with version `2.3.2-debian`.

### Deploy ProxySQL Cluster

In this section, we are going to deploy a ProxySQL cluster. Then, in the next section we will scale the proxy server using `ProxySQLOpsRequest` CRD. Below is the YAML of the `ProxySQL` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ProxySQL
metadata:
  name: proxy-server
  namespace: demo
spec:
  version: "2.3.2-debian"
  replicas: 3
  mode: GroupReplication
  backend:
    name: mysql-server
  syncUsers: true
  terminationPolicy: WipeOut
```

Let's create the `ProxySQL` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/scaling/horizontal-scaling/cluster/example/sample-proxysql.yaml
proxysql.kubedb.com/proxy-server created
```

Now, wait until `proxy-server` has status `Ready`. i.e,

```bash
$ kubectl get proxysql -n demo
NAME             VERSION       STATUS    AGE
proxy-server   2.3.2-debian    Ready    2m36s
```

Let's check the number of replicas this cluster has from the ProxySQL object, number of pods the statefulset have,

```bash
$ kubectl get proxysql -n demo proxy-server -o json | jq '.spec.replicas'
3
$ kubectl get sts -n demo proxy-server -o json | jq '.spec.replicas'
3
```

We can see from both command that the server has 3 replicas in the cluster.

Also, we can verify the replicas of the replicaset from an internal proxysql command by execing into a replica.

Now let's connect to a proxysql instance and run a proxysql internal command to check the cluster status,

```bash
$  kubectl exec -it -n demo proxy-server-0 -- bash
root@proxy-server-1:/# mysql -uadmin -padmin -h127.0.0.1 -P6032 -e "select * from runtime_proxysql_servers;"
+---------------------------------------+------+--------+---------+
| hostname                              | port | weight | comment |
+---------------------------------------+------+--------+---------+
| proxy-server-2.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-1.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-0.proxy-server-pods.demo | 6032 | 1      |         |
+---------------------------------------+------+--------+---------+
root@proxy-server-1:/# 


```

We can see from the above output that the cluster has 3 nodes.

We are now ready to apply the `ProxySQLOpsRequest` CR to scale this server.

## Scale Up Replicas

Here, we are going to scale up the replicas of the replicaset to meet the desired number of replicas after scaling.

#### Create ProxySQLOpsRequest

In order to scale up the replicas of the replicaset of the server, we have to create a `ProxySQLOpsRequest` CR with our desired replicas. Below is the YAML of the `ProxySQLOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  proxyRef:
    name: proxy-server
  horizontalScaling:
    member: 5

```

Here,

- `spec.proxyRef.name` specifies that we are performing horizontal scaling operation on `proxy-server` instance.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.member` specifies the desired replicas after scaling.

Let's create the `ProxySQLOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/scaling/horizontal-scaling/cluster/example/proxyops-upscale.yaml
proxysqlopsrequest.ops.kubedb.com/scale-up created
```

#### Verify Cluster replicas scaled up successfully

If everything goes well, `KubeDB` Enterprise operator will update the replicas of `ProxySQL` object and related `StatefulSets` and `Pods`.

Let's wait for `ProxySQLOpsRequest` to be `Successful`.  Run the following command to watch `ProxySQLOpsRequest` CR,

```bash
$ watch kubectl get proxysqlopsrequest -n demo
Every 2.0s: kubectl get proxysqlopsrequest -n demo
NAME                        TYPE                STATUS       AGE
scale-up                HorizontalScaling    Successful     106s
```

We can see from the above output that the `ProxySQLOpsRequest` has succeeded. Now, we are going to verify the number of replicas this database has from the ProxySQL object, number of pods the statefulset have,

```bash
$ kubectl get proxysql -n demo proxy-server -o json | jq '.spec.replicas'
5
$ kubectl get sts -n demo proxy-server -o json | jq '.spec.replicas'
5
```

Now let's connect to a proxysql instance and run a proxysql internal command to check the number of replicas,

```bash
$  kubectl exec -it -n demo proxy-server-0 -- bash
root@proxy-server-1:/# mysql -uadmin -padmin -h127.0.0.1 -P6032 -e "select * from runtime_proxysql_servers;"
+---------------------------------------+------+--------+---------+
| hostname                              | port | weight | comment |
+---------------------------------------+------+--------+---------+
| proxy-server-2.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-1.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-0.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-3.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-4.proxy-server-pods.demo | 6032 | 1      |         |
+---------------------------------------+------+--------+---------+
root@proxy-server-1:/# 


```

From all the above outputs we can see that the replicas of the cluster is `5`. That means we have successfully scaled up the replicas of the ProxySQL replicaset.

### Scale Down Replicas

Here, we are going to scale down the replicas of the cluster to meet the desired number of replicas after scaling.

#### Create ProxySQLOpsRequest

In order to scale down the cluster of the server, we have to create a `ProxySQLOpsRequest` CR with our desired replicas. Below is the YAML of the `ProxySQLOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: scale-down
  namespace: demo
spec:
  type: HorizontalScaling
  proxyRef:
    name: proxy-server
  horizontalScaling:
    member: 4
```

Here,

- `spec.proxyRef.name` specifies that we are performing horizontal scaling operation on `proxy-server` instance.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.member` specifies the desired replicas after scaling.

Let's create the `ProxySQLOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/scaling/horizontal-scaling/cluster/example/proxyops-downscale.yaml
proxysqlopsrequest.ops.kubedb.com/scale-down created
```

#### Verify Cluster replicas scaled down successfully

If everything goes well, `KubeDB` Enterprise operator will update the replicas of `ProxySQL` object and related `StatefulSets` and `Pods`.

Let's wait for `ProxySQLOpsRequest` to be `Successful`.  Run the following command to watch `ProxySQLOpsRequest` CR,

```bash
$ watch kubectl get proxysqlopsrequest -n demo
Every 2.0s: kubectl get proxysqlopsrequest -n demo
NAME                          TYPE                STATUS       AGE
scale-down              HorizontalScaling       Successful   2m32s
```

We can see from the above output that the `ProxySQLOpsRequest` has succeeded. Now, we are going to verify the number of replicas this database has from the ProxySQL object, number of pods the statefulset have,

```bash
$ kubectl get proxysql -n demo proxy-server -o json | jq '.spec.replicas' 
3
$ kubectl get sts -n demo proxy-server -o json | jq '.spec.replicas'
3
```

Now let's connect to a proxysql instance and run a proxysql internal command to check the number of replicas,
```bash
$  kubectl exec -it -n demo proxy-server-0 -- bash
root@proxy-server-1:/# mysql -uadmin -padmin -h127.0.0.1 -P6032 -e "select * from runtime_proxysql_servers;"
+---------------------------------------+------+--------+---------+
| hostname                              | port | weight | comment |
+---------------------------------------+------+--------+---------+
| proxy-server-2.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-1.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-0.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-3.proxy-server-pods.demo | 6032 | 1      |         |
+---------------------------------------+------+--------+---------+
root@proxy-server-1:/# 

```

From all the above outputs we can see that the replicas of the cluster is `4`. That means we have successfully scaled down the replicas of the ProxySQL replicaset.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete proxysql -n demo proxy-server
$ kubectl delete proxysqlopsrequest -n demo  scale-up scale-down
```