---
title: Horizontal Scaling Topology Cassandra
menu:
  docs_{{ .version }}:
    identifier: cas-horizontal-scaling-topology
    name: Topology Cluster
    parent: cas-horizontal-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Cassandra Topology Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to scale the Cassandra topology cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Cassandra](/docs/guides/cassandra/concepts/cassandra.md)
    - [Topology](/docs/guides/cassandra/clustering/topology-cluster/index.md)
    - [CassandraOpsRequest](/docs/guides/cassandra/concepts/cassandraopsrequest.md)
    - [Horizontal Scaling Overview](/docs/guides/cassandra/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/cassandra](/docs/examples/cassandra) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on Topology Cluster

Here, we are going to deploy a `Cassandra` topology cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare Cassandra Topology cluster

Now, we are going to deploy a `Cassandra` topology cluster with version `5.0.3`.

### Deploy Cassandra topology cluster

In this section, we are going to deploy a Cassandra topology cluster. Then, in the next section we will scale the cluster using `CassandraOpsRequest` CRD. Below is the YAML of the `Cassandra` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:
  name: cassandra-prod
  namespace: demo
spec:
  version: 5.0.3
  topology:
    rack:
      - name: r0
        replicas: 2
        podTemplate:
          spec:
            containers:
              - name: cassandra
                resources:
                  limits:
                    memory: 2Gi
                    cpu: 2
                  requests:
                    memory: 1Gi
                    cpu: 1
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
        storageType: Durable
  deletionPolicy: WipeOut
```

Let's create the `Cassandra` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/scaling/cassandra-topology.yaml
cassandra.kubedb.com/cassandra-prod created
```

Now, wait until `cassandra-prod` has status `Ready`. i.e,

```bash
$kubectl get cas -n demo -w
NAME             TYPE                  VERSION   STATUS         AGE
cassandra-prod   kubedb.com/v1alpha2   5.0.3     Provisioning   27s
cassandra-prod   kubedb.com/v1alpha2   5.0.3     Provisioning   1m27s
.
.
cassandra-prod   kubedb.com/v1alpha2   5.0.3     Ready          2m27s
```

Let's check the number of replicas has from cassandra object, number of pods the petset have,

```bash
$ kubectl get petset -n demo cassandra-prod-rack-r0 -o json | jq '.spec.replicas'
2
```

We can see from commands that the cluster has 2 replicas for rack r0 as we have defined in the yaml.

We are now ready to apply the `CassandraOpsRequest` CR to scale this cluster.

## Scale Up Replicas

Here, we are going to scale up the replicas of the topology cluster to meet the desired number of replicas after scaling.

#### Create CassandraOpsRequest

In order to scale up the replicas of the topology cluster, we have to create a `CassandraOpsRequest` CR with our desired replicas. Below is the YAML of the `CassandraOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: cassandra-horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: cassandra-prod
  horizontalScaling:
    node: 4
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `cassandra-prod` cluster.
- `spec.type` specifies that we are performing `HorizontalScaling` on cassandra.
- `spec.horizontalScaling.node` specifies the desired replicas after scaling for cassandra cluster.

Let's create the `CassandraOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/scaling/horizontal-scaling/cassandra-hscale-up-topology.yaml
cassandraopsrequest.ops.kubedb.com/casops-hscale-up-topology created
```

#### Verify Topology cluster replicas scaled up successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Cassandra` object and related `PetSets` and `Pods`.

Let's wait for `CassandraOpsRequest` to be `Successful`. Run the following command to watch `CassandraOpsRequest` CR,

```bash
$ watch kubectl get cassandraopsrequest -n demo
NAME                             TYPE                STATUS       AGE
cassandra-horizontal-scale-up    HorizontalScaling   Successful   106s
```

We can see from the above output that the `CassandraOpsRequest` has succeeded. If we describe the `CassandraOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe cassandraopsrequests -n demo cassandra-horizontal-scale-up 
kubectl describe cassandraopsrequests -n demo cassandra-horizontal-scale-up
Name:         cassandra-horizontal-scale-up
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         CassandraOpsRequest
Metadata:
  Creation Timestamp:  2025-07-18T05:48:27Z
  Generation:          1
  Resource Version:    2808
  UID:                 705dd54c-dc75-4a1f-bbbb-bc1b6028c611
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  cassandra-prod
  Horizontal Scaling:
    Node:  4
  Type:    HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2025-07-18T05:49:37Z
    Message:               Cassandra ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2025-07-18T05:49:55Z
    Message:               Successfully Scaled Up Node
    Observed Generation:   1
    Reason:                HorizontalScaleUp
    Status:                True
    Type:                  HorizontalScaleUp
    Last Transition Time:  2025-07-18T05:49:45Z
    Message:               patch petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetset
    Last Transition Time:  2025-07-18T05:50:04Z
    Message:               successfully reconciled the Cassandra with modified node
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-18T05:50:04Z
    Message:               Successfully completed horizontally scale Cassandra cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                              Age   From                         Message
  ----     ------                              ----  ----                         -------
  Normal   Starting                            39s   KubeDB Ops-manager Operator  Start processing for CassandraOpsRequest: demo/cassandra-horizontal-scale-up
  Normal   Starting                            39s   KubeDB Ops-manager Operator  Pausing Cassandra databse: demo/cassandra-prod
  Normal   Successful                          39s   KubeDB Ops-manager Operator  Successfully paused Cassandra database: demo/cassandra-prod for CassandraOpsRequest: cassandra-horizontal-scale-up
  Warning  patch petset; ConditionStatus:True  31s   KubeDB Ops-manager Operator  patch petset; ConditionStatus:True
  Warning  patch petset; ConditionStatus:True  26s   KubeDB Ops-manager Operator  patch petset; ConditionStatus:True
  Normal   HorizontalScaleUp                   21s   KubeDB Ops-manager Operator  Successfully Scaled Up Node
  Normal   UpdatePetSets                       12s   KubeDB Ops-manager Operator  successfully reconciled the Cassandra with modified node
  Normal   Starting                            12s   KubeDB Ops-manager Operator  Resuming Cassandra database: demo/cassandra-prod
  Normal   Successful                          12s   KubeDB Ops-manager Operator  Successfully resumed Cassandra database: demo/cassandra-prod for CassandraOpsRequest: cassandra-horizontal-scale-up
```

Now, we are going to verify the number of replicas this cluster has from the Cassandra object, number of pods the petset have,

```bash
$ kubectl get petset -n demo cassandra-prod-rack-r0 -o json | jq '.spec.replicas'
4
```

### Scale Down Replicas

Here, we are going to scale down the replicas of the cassandra topology cluster to meet the desired number of replicas after scaling.

#### Create CassandraOpsRequest

In order to scale down the replicas of the cassandra topology cluster, we have to create a `CassandraOpsRequest` CR with our desired replicas. Below is the YAML of the `CassandraOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: cassandra-horizontal-scale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: cassandra-prod
  horizontalScaling:
    node: 2
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down operation on `cassandra-prod` cluster.
- `spec.type` specifies that we are performing `HorizontalScaling` on cassandra.
- `spec.horizontalScaling.topology.node` specifies the desired replicas after scaling for the cassandra nodes.

Let's create the `CassandraOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/scaling/horizontal-scaling/cassandra-hscale-down-topology.yaml
cassandraopsrequest.ops.kubedb.com/casops-hscale-down-topology created
```

#### Verify Topology cluster replicas scaled down successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Cassandra` object and related `PetSets` and `Pods`.

Let's wait for `CassandraOpsRequest` to be `Successful`. Run the following command to watch `CassandraOpsRequest` CR,

```bash
$ watch kubectl get cassandraopsrequest -n demo
NAME                              TYPE                STATUS       AGE
cassandra-horizontal-scale-down   HorizontalScaling   Successful   62s
```

We can see from the above output that the `CassandraOpsRequest` has succeeded. If we describe the `CassandraOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe cassandraopsrequests -n demo cassandra-horizontal-scale-down
Name:         cassandra-horizontal-scale-down
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         CassandraOpsRequest
Metadata:
  Creation Timestamp:  2025-07-18T05:53:38Z
  Generation:          1
  Resource Version:    2937
  UID:                 073b5918-d0c5-4a0a-9d17-3927220a727b
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  cassandra-prod
  Horizontal Scaling:
    Node:  2
  Type:    HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2025-07-18T05:53:38Z
    Message:               Cassandra ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2025-07-18T05:53:56Z
    Message:               Successfully Scaled Down Node
    Observed Generation:   1
    Reason:                HorizontalScaleDown
    Status:                True
    Type:                  HorizontalScaleDown
    Last Transition Time:  2025-07-18T05:53:46Z
    Message:               patch petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetset
    Last Transition Time:  2025-07-18T05:54:05Z
    Message:               successfully reconciled the Cassandra with modified node
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-18T05:54:05Z
    Message:               Successfully completed horizontally scale Cassandra cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                              Age    From                         Message
  ----     ------                              ----   ----                         -------
  Normal   Starting                            2m21s  KubeDB Ops-manager Operator  Start processing for CassandraOpsRequest: demo/cassandra-horizontal-scale-down
  Normal   Starting                            2m21s  KubeDB Ops-manager Operator  Pausing Cassandra databse: demo/cassandra-prod
  Normal   Successful                          2m21s  KubeDB Ops-manager Operator  Successfully paused Cassandra database: demo/cassandra-prod for CassandraOpsRequest: cassandra-horizontal-scale-down
  Warning  patch petset; ConditionStatus:True  2m13s  KubeDB Ops-manager Operator  patch petset; ConditionStatus:True
  Warning  patch petset; ConditionStatus:True  2m8s   KubeDB Ops-manager Operator  patch petset; ConditionStatus:True
  Normal   HorizontalScaleDown                 2m3s   KubeDB Ops-manager Operator  Successfully Scaled Down Node
  Normal   UpdatePetSets                       114s   KubeDB Ops-manager Operator  successfully reconciled the Cassandra with modified node
  Normal   Starting                            114s   KubeDB Ops-manager Operator  Resuming Cassandra database: demo/cassandra-prod
  Normal   Successful                          114s   KubeDB Ops-manager Operator  Successfully resumed Cassandra database: demo/cassandra-prod for CassandraOpsRequest: cassandra-horizontal-scale-down
```

Now, we are going to verify the number of replicas this cluster has from the number of pods the petset have,

**Broker Replicas**

```bash
$
$ kubectl get petset -n demo cassandra-prod-rack-r0 -o json | jq '.spec.replicas'
2
```

From all the above outputs we can see that the replicas of the topology cluster is `2`. That means we have successfully scaled down the replicas of the Cassandra topology cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete cas -n demo cassandra-prod
kubectl delete cassandraopsrequest -n demo  cassandra-horizontal-scale-up  cassandra-horizontal-scale-down
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Cassandra object](/docs/guides/cassandra/concepts/cassandra.md).
- Different Cassandra topology clustering modes [here](/docs/guides/cassandra/clustering/_index.md).
- Monitor your Cassandra with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/cassandra/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Cassandra with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/cassandra/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
