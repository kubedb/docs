---
title: Horizontal Scaling SingleStore
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-scaling-horizontal-cluster
    name: Horizontal Scaling OpsRequest
    parent: guides-sdb-scaling-horizontal
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale SingleStore

This guide will show you how to use `KubeDB` Enterprise operator to scale the cluster of a SingleStore database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [SingleStore](/docs/guides/singlestore/concepts/singlestore.md)
  - [SingleStore Cluster](/docs/guides/singlestore/clustering/)
  - [SingleStoreOpsRequest](/docs/guides/singlestore/concepts/opsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/singlestore/scaling/horizontal-scaling/overview/)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Apply Horizontal Scaling on Cluster

Here, we are going to deploy a  `SingleStore` cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Create SingleStore License Secret

We need SingleStore License to create SingleStore Database. So, Ensure that you have acquired a license and then simply pass the license by secret.

```bash
$ kubectl create secret generic -n demo license-secret \
                --from-literal=username=license \
                --from-literal=password='your-license-set-here'
secret/license-secret created
```

### Deploy SingleStore Cluster

In this section, we are going to deploy a SingleStore cluster. Then, in the next section we will scale the database using `SingleStoreOpsRequest` CRD. Below is the YAML of the `SingleStore` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sample-sdb
  namespace: demo
spec:
  version: "8.7.10"
  topology:
    aggregator:
      replicas: 1
      podTemplate:
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                memory: "2Gi"
                cpu: "600m"
              requests:
                memory: "2Gi"
                cpu: "600m"
      storage:
        storageClassName: "longhorn"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "600m"
                requests:
                  memory: "2Gi"
                  cpu: "600m"                      
      storage:
        storageClassName: "longhorn"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  licenseSecret:
    kind: Secret
    name: license-secret
  storageType: Durable
  deletionPolicy: WipeOut
```

Let's create the `SingleStore` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/scaling/horizontal-scaling/cluster/example/sample-sdb.yaml
singlestore.kubedb.com/sample-sdb created
```

Now, wait until `sample-sdb` has status `Ready`. i.e,

```bash
$ kubectl get singlestore -n demo
NAME         TYPE                  VERSION   STATUS   AGE
sample-sdb   kubedb.com/v1alpha2   8.7.10    Ready    86s
```

Let's check the number of `aggreagtor replicas` and `leaf replicas` this database has from the SingleStore object, number of pods the `aggregator-petset` and `leaf-petset` have,

```bash
$ kubectl get sdb -n demo sample-sdb -o json | jq '.spec.topology.aggregator.replicas'
1
$ kubectl get sdb -n demo sample-sdb -o json | jq '.spec.topology.leaf.replicas'
2

$ kubectl get petset -n demo sample-sdb-aggregator -o=jsonpath='{.spec.replicas}{"\n"}'
1
kubectl get petset -n demo sample-sdb-leaf -o=jsonpath='{.spec.replicas}{"\n"}'
2


```

We can see from both command that the database has 1 `aggregator replicas` and 2 `leaf replicas` in the cluster.

Also, we can verify the replicas of the from an internal memsqlctl command by execing into a replica.

Now let's connect to a singlestore instance and run a memsqlctl internal command to check the number of replicas,

```bash
$ kubectl exec -it -n demo sample-sdb-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@sample-sdb-aggregator-0 /]$ memsqlctl show-cluster
+---------------------+--------------------------------------------------+------+--------------------+-----------+-----------+--------+--------------------+------------------------------+--------+-------------------+
|        Role         |                       Host                       | Port | Availability Group | Pair Host | Pair Port | State  | Opened Connections | Average Roundtrip Latency ms | NodeId | Master Aggregator |
+---------------------+--------------------------------------------------+------+--------------------+-----------+-----------+--------+--------------------+------------------------------+--------+-------------------+
| Leaf                | sample-sdb-leaf-0.sample-sdb-pods.demo.svc       | 3306 | 1                  | null      | null      | online | 2                  |                              | 2      |                   |
| Leaf                | sample-sdb-leaf-1.sample-sdb-pods.demo.svc       | 3306 | 1                  | null      | null      | online | 3                  |                              | 3      |                   |
| Aggregator (Leader) | sample-sdb-aggregator-0.sample-sdb-pods.demo.svc | 3306 |                    | null      | null      | online | 1                  | null                         | 1      | 1                 |
+---------------------+--------------------------------------------------+------+--------------------+-----------+-----------+--------+--------------------+------------------------------+--------+-------------------+


```

We can see from the above output that the cluster has 1 aggregator node and 2 leaf nodes.

We are now ready to apply the `SingleStoreOpsRequest` CR to scale this database.

## Scale Up Replicas

Here, we are going to scale up the replicas of the `leaf nodes` to meet the desired number of replicas after scaling.

#### Create SingleStoreOpsRequest

In order to scale up the replicas of the `leaf nodes` of the database, we have to create a `SingleStoreOpsRequest` CR with our desired replicas. Below is the YAML of the `SingleStoreOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdbops-scale-horizontal-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: sample-sdb
  horizontalScaling:
    leaf: 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `sample-sdb` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.leaf` specifies the desired leaf replicas after scaling.

Let's create the `SingleStoreOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/scaling/horizontal-scaling/cluster/example/sdbops-upscale.yaml
singlestoreopsrequest.ops.kubedb.com/sdbops-scale-horizontal-up created
```

#### Verify Cluster replicas scaled up successfully 

If everything goes well, `KubeDB` Enterprise operator will update the replicas of `SingleStore` object and related `PetSets` and `Pods`.

Let's wait for `SingleStoreOpsRequest` to be `Successful`.  Run the following command to watch `SingleStoreOpsRequest` CR,

```bash
 $ kubectl get singlestoreopsrequest -n demo
NAME                         TYPE                STATUS       AGE
sdbops-scale-horizontal-up   HorizontalScaling   Successful   74s
```

We can see from the above output that the `SingleStoreOpsRequest` has succeeded. Now, we are going to verify the number of `leaf replicas` this database has from the SingleStore object, number of pods the `leaf petset` have,

```bash
$ kubectl get sdb -n demo sample-sdb -o json | jq '.spec.topology.leaf.replicas'
3
$ kubectl get petset -n demo sample-sdb-leaf -o=jsonpath='{.spec.replicas}{"\n"}'
3

```

Now let's connect to a singlestore instance and run a memsqlctl internal command to check the number of replicas,

```bash
$ kubectl exec -it -n demo sample-sdb-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@sample-sdb-aggregator-0 /]$ memsqlctl show-cluster
+---------------------+--------------------------------------------------+------+--------------------+-----------+-----------+--------+--------------------+------------------------------+--------+-------------------+
|        Role         |                       Host                       | Port | Availability Group | Pair Host | Pair Port | State  | Opened Connections | Average Roundtrip Latency ms | NodeId | Master Aggregator |
+---------------------+--------------------------------------------------+------+--------------------+-----------+-----------+--------+--------------------+------------------------------+--------+-------------------+
| Leaf                | sample-sdb-leaf-0.sample-sdb-pods.demo.svc       | 3306 | 1                  | null      | null      | online | 2                  |                              | 2      |                   |
| Leaf                | sample-sdb-leaf-1.sample-sdb-pods.demo.svc       | 3306 | 1                  | null      | null      | online | 3                  |                              | 3      |                   |
| Leaf                | sample-sdb-leaf-2.sample-sdb-pods.demo.svc       | 3306 | 1                  | null      | null      | online | 2                  |                              | 4      |                   |
| Aggregator (Leader) | sample-sdb-aggregator-0.sample-sdb-pods.demo.svc | 3306 |                    | null      | null      | online | 1                  | null                         | 1      | 1                 |
+---------------------+--------------------------------------------------+------+--------------------+-----------+-----------+--------+--------------------+------------------------------+--------+-------------------+

```

From all the above outputs we can see that the `leaf replicas` of the cluster is `3`. That means we have successfully scaled up the `leaf replicas` of the SingleStore Cluster.

### Scale Down Replicas

Here, we are going to scale down the `leaf replicas` of the cluster to meet the desired number of replicas after scaling.

#### Create SingleStoreOpsRequest

In order to scale down the cluster of the database, we have to create a `SingleStoreOpsRequest` CR with our desired replicas. Below is the YAML of the `SingleStoreOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdbops-scale-horizontal-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: sample-sdb
  horizontalScaling:
    leaf: 2
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down operation on `sample-sdb` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.leaf` specifies the desired `leaf replicas` after scaling.

Let's create the `SingleStoreOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/scaling/horizontal-scaling/cluster/example/sdbops-downscale.yaml
singlestoreopsrequest.ops.kubedb.com/sdbops-scale-horizontal-down created
```

#### Verify Cluster replicas scaled down successfully 

If everything goes well, `KubeDB` Enterprise operator will update the replicas of `SingleStore` object and related `PetSets` and `Pods`.

Let's wait for `SingleStoreOpsRequest` to be `Successful`.  Run the following command to watch `SingleStoreOpsRequest` CR,

```bash
$ kubectl get singlestoreopsrequest -n demo
NAME                           TYPE                STATUS       AGE
sdbops-scale-horizontal-down   HorizontalScaling   Successful   63s
```

We can see from the above output that the `SingleStoreOpsRequest` has succeeded. Now, we are going to verify the number of `leaf replicas` this database has from the SingleStore object, number of pods the `leaf petset` have,

```bash
$ kubectl get sdb -n demo sample-sdb -o json | jq '.spec.topology.leaf.replicas'
2
$ kubectl get petset -n demo sample-sdb-leaf -o=jsonpath='{.spec.replicas}{"\n"}'
2

```

Now let's connect to a singlestore instance and run a memsqlctl internal command to check the number of replicas,
```bash
$ kubectl exec -it -n demo sample-sdb-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
bash: mesqlctl: command not found
[memsql@sample-sdb-aggregator-0 /]$ memsqlctl show-cluster
+---------------------+--------------------------------------------------+------+--------------------+-----------+-----------+--------+--------------------+------------------------------+--------+-------------------+
|        Role         |                       Host                       | Port | Availability Group | Pair Host | Pair Port | State  | Opened Connections | Average Roundtrip Latency ms | NodeId | Master Aggregator |
+---------------------+--------------------------------------------------+------+--------------------+-----------+-----------+--------+--------------------+------------------------------+--------+-------------------+
| Leaf                | sample-sdb-leaf-0.sample-sdb-pods.demo.svc       | 3306 | 1                  | null      | null      | online | 2                  |                              | 2      |                   |
| Leaf                | sample-sdb-leaf-1.sample-sdb-pods.demo.svc       | 3306 | 1                  | null      | null      | online | 3                  |                              | 3      |                   |
| Aggregator (Leader) | sample-sdb-aggregator-0.sample-sdb-pods.demo.svc | 3306 |                    | null      | null      | online | 1                  | null                         | 1      | 1                 |
+---------------------+--------------------------------------------------+------+--------------------+-----------+-----------+--------+--------------------+------------------------------+--------+-------------------+

```

From all the above outputs we can see that the `leaf replicas` of the cluster is `2`. That means we have successfully scaled down the `leaf replicas` of the SingleStore database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete sdb -n demo sample-sdb
$ kubectl delete singlestoreopsrequest -n demo  sdbops-scale-horizontal-up sdbops-scale-horizontal-down
```