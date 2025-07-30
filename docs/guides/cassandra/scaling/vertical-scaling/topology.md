---
title: Vertical Scaling Cassandra Topology Cluster
menu:
  docs_{{ .version }}:
    identifier: cas-vertical-scaling-topology
    name: Topology Cluster
    parent: cas-vertical-scaling
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale Cassandra Topology Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to update the resources of a Cassandra topology cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Cassandra](/docs/guides/cassandra/concepts/cassandra.md)
    - [Topology](/docs/guides/cassandra/clustering/topology-cluster/index.md)
    - [CassandraOpsRequest](/docs/guides/cassandra/concepts/cassandraopsrequest.md)
    - [Vertical Scaling Overview](/docs/guides/cassandra/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/cassandra](/docs/examples/cassandra) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on Topology Cluster

Here, we are going to deploy a `Cassandra` topology cluster using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare Cassandra Topology Cluster

Now, we are going to deploy a `Cassandra` topology cluster database with version `5.0.3`.

### Deploy Cassandra Topology Cluster

In this section, we are going to deploy a Cassandra topology cluster. Then, in the next section we will update the resources of the database using `CassandraOpsRequest` CRD. Below is the YAML of the `Cassandra` CR that we are going to create,

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
$ kubectl get cas -n demo -w
NAME             TYPE                  VERSION   STATUS         AGE
cassandra-prod   kubedb.com/v1alpha2   5.0.3     Provisioning   22s
cassandra-prod   kubedb.com/v1alpha2   5.0.3     Provisioning   45s
.
.
cassandra-prod   kubedb.com/v1alpha2   5.0.3     Ready   104s
```

Let's check the Pod containers resources of the Cassandra topology cluster. Run the following command to get the resources of the containers of the Cassandra topology cluster

```bash
$ kubectl get pod -n demo cassandra-prod-rack-r0-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "2",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "1",
    "memory": "1Gi"
  }
}
```

We are now ready to apply the `CassandraOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of the topology cluster to meet the desired resources after scaling.

#### Create CassandraOpsRequest

In order to update the resources of the database, we have to create a `CassandraOpsRequest` CR with our desired resources. Below is the YAML of the `CassandraOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: cassandra-vertical-scale
  namespace: default
spec:
  type: VerticalScaling
  databaseRef:
    name: cassandra-prod
  verticalScaling:
    node:
      resources:
        requests:
          memory: "3Gi"
          cpu: "2"
        limits:
          memory: "4Gi"
          cpu: "3"
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `cassandra-prod` cluster.
- `spec.type` specifies that we are performing `VerticalScaling` on cassandra.
- `spec.VerticalScaling.node` specifies the desired resources after scaling.

Let's create the `CassandraOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/scaling/vertical-scaling/cassandra-vertical-scaling-topology.yaml
cassandraopsrequest.ops.kubedb.com/casops-vscale-topology created
```

#### Verify Cassandra Topology cluster resources updated successfully

If everything goes well, `KubeDB` Ops-manager operator will update the resources of `Cassandra` object and related `PetSets` and `Pods`.

Let's wait for `CassandraOpsRequest` to be `Successful`.  Run the following command to watch `CassandraOpsRequest` CR,

```bash
$ kubectl get cassandraopsrequest -n demo
NAME                        TYPE              STATUS       AGE
cassandra-vertical-scale    VerticalScaling   Successful   3m56s
```

We can see from the above output that the `CassandraOpsRequest` has succeeded. If we describe the `CassandraOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$  kubectl describe cassandraopsrequest -n demo cassandra-vertical-scale
Name:         cassandra-vertical-scale
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         CassandraOpsRequest
Metadata:
  Creation Timestamp:  2025-07-18T08:35:29Z
  Generation:          1
  Resource Version:    8364
  UID:                 282858bc-553d-487d-aad5-2026a5dd6a9e
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   cassandra-prod
  Timeout:  5m
  Type:     VerticalScaling
  Vertical Scaling:
    Node:
      Resources:
        Limits:
          Cpu:     3
          Memory:  4Gi
        Requests:
          Cpu:     2
          Memory:  3Gi
Status:
  Conditions:
    Last Transition Time:  2025-07-18T08:35:29Z
    Message:               Cassandra ops-request has started to vertically scaling the Cassandra nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2025-07-18T08:35:32Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-18T08:38:18Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2025-07-18T08:35:38Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-18T08:35:38Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-18T08:35:43Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-07-18T08:36:18Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-18T08:36:18Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-18T08:38:18Z
    Message:               Successfully completed the vertical scaling for Cassandra
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                             Age   From                         Message
  ----     ------                                                             ----  ----                         -------
  Normal   Starting                                                           36m   KubeDB Ops-manager Operator  Start processing for CassandraOpsRequest: demo/cassandra-vertical-scale
  Normal   Starting                                                           36m   KubeDB Ops-manager Operator  Pausing Cassandra databse: demo/cassandra-prod
  Normal   Successful                                                         36m   KubeDB Ops-manager Operator  Successfully paused Cassandra database: demo/cassandra-prod for CassandraOpsRequest: cassandra-vertical-scale
  Normal   UpdatePetSets                                                      36m   KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0  36m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  running pod; ConditionStatus:False                                 36m   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1    36m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1  36m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0    35m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0  35m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1    34m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1  34m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Normal   RestartPods                                                        34m   KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                           34m   KubeDB Ops-manager Operator  Resuming Cassandra database: demo/cassandra-prod
  Normal   Successful                                                         34m   KubeDB Ops-manager Operator  Successfully resumed Cassandra database: demo/cassandra-prod for CassandraOpsRequest: cassandra-vertical-scale                                                              2m18s  KubeDB Ops-manager Operator  Successfully resumed Cassandra database: demo/cassandra-prod for CassandraOpsRequest: casops-vscale-topology
```
Now, we are going to verify from one of the Pod yaml whether the resources of the topology cluster has updated to meet up the desired state, Let's check,

```bash
$  kubectl get pod -n demo cassandra-prod-rack-r0-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "3",
    "memory": "4Gi"
  },
  "requests": {
    "cpu": "2",
    "memory": "3Gi"
  }
}
```

The above output verifies that we have successfully scaled up the resources of the Cassandra topology cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete cas -n demo cassandra-prod
kubectl delete cassandraopsrequest -n demo cassandra-vertical-scale
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Cassandra object](/docs/guides/cassandra/concepts/cassandra.md).
- Different Cassandra topology clustering modes [here](/docs/guides/cassandra/clustering/_index.md).
- Monitor your Cassandra database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/cassandra/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Cassandra database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/cassandra/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
