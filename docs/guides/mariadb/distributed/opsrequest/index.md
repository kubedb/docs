---
title: Distributed MariaDBOpsRequest
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-distributed-opsrequest
    name: Distributed MariaDBOpsRequest
    parent: guides-mariadb-distributed
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Opsrequest of a Distributed MariaDB Cluster
This guide demonstrates how to use `OpsRequest` with the `MariaDB` CRD in a distributed cluster setup.

> **Note:** All `OpsRequest` operations behave consistently in a distributed environment, except for `VolumeExpansion`, `HorizontalScaling`, and `Reconfigure`, which have cluster-specific considerations.


## Before You Begin

- At first, you need to have a multi-cluster Kubernetes setup with OCM and KubeSlice configured. Follow the [Distributed MariaDB Overview](/docs/guides/mariadb/distributed/overview) guide to set up the required infrastructure.

- Install `KubeDB` Community, Ops-Manager and Autoscaler operator in your hub cluster following the steps [here](/docs/setup/README.md). Make sure to enable OCM support:

```bash
  --set petset.features.ocm.enabled=true
```

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- Install Prometheus in each spoke cluster. The autoscaler queries the per-cluster Prometheus endpoint (configured in the `PlacementPolicy`) to collect resource usage metrics. You can install it from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

- You should be familiar with the following `KubeDB` concepts:
  - [MariaDB](/docs/guides/mariadb/concepts/mariadb)
  - [MariaDBDistributedOverview](/docs/guides/mariadb/distributed/overview/index.md)
  - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/opsrequest)
To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## OpsRequest of Distributed Cluster Database

Here, we are going to deploy a distributed `MariaDB` Galera cluster using a supported version by `KubeDB` operator. Then we are going to apply `MariaDopsRequest` to set up autoscaling.

### Deploy PlacementPolicy
For horizontal scaling, you can increase or decrease the number of replicas as needed. To support this, the `PlacementPolicy` should be defined with the maximum number of replicas you anticipate requiring. This allows you to scale up to that limit while also permitting operation with fewer replicas if needed.

In this example, the `PlacementPolicy` is configured with a maximum of five replicas. It also explicitly defines how replicas are distributed across clusters, specifying which replicas are assigned to each cluster.

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  labels:
    app.kubernetes.io/managed-by: Helm
  name: distributed-mariadb
spec:
  clusterSpreadConstraint:
    distributionRules:
      - clusterName: demo-controller
        replicaIndices:
          - 0
          - 2
      - clusterName: demo-worker
        replicaIndices:
          - 1
    slice:
      projectNamespace: kubeslice-demo-distributed-mariadb
      sliceName: demo-slice
  nodeSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: ScheduleAnyway
  zoneSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: ScheduleAnyway
```

Here,

- `spec.clusterSpreadConstraint.distributionRules[].replicaIndices` specifies which MariaDB replica indices are scheduled on that cluster. Here `demo-controller` hosts replicas `0` and `2`, and `demo-worker` hosts replica `1`.

Apply the `PlacementPolicy` on the hub (`demo-controller`) cluster:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/distributed/autoscaler/compute/cluster/examples/placement-policy.yaml --context demo-controller
placementpolicy.apps.k8s.appscode.com/distributed-mariadb created
```

### Deploy Distributed MariaDB Cluster

In this section, we are going to deploy a distributed MariaDB Galera cluster with version `11.5.2`. Then, in the next section we will set up autoscaling for this database using `MariaDBAutoscaler` CRD.

Below is the YAML of the `MariaDB` CR that we are going to create. Note that `spec.distributed` is set to `true` and the `PlacementPolicy` is referenced via `spec.podTemplate.spec.podPlacementPolicy`:

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "11.5.2"
  distributed: true
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  podTemplate:
    spec:
      podPlacementPolicy:
        name: distributed-mariadb
      containers:
      - name: mariadb
        resources:
          requests:
            cpu: "200m"
            memory: "300Mi"
          limits:
            cpu: "200m"
            memory: "300Mi"
  deletionPolicy: WipeOut
```

Let's create the `MariaDB` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/distributed/autoscaler/compute/cluster/examples/sample-mariadb.yaml --context demo-controller
mariadb.kubedb.com/sample-mariadb created
```

Now, wait until `sample-mariadb` has status `Ready`. i.e,

```bash
$ kubectl get mariadb -n demo --context demo-controller
NAME             VERSION   STATUS   AGE
sample-mariadb   11.5.2   Ready    14m
```

The pods are distributed across clusters as defined by the `PlacementPolicy`:

```bash
$  kubectl get pod -n demo --context demo-controller
NAME        READY   STATUS    RESTARTS   AGE
mariadb-0   3/3     Running   0          14m
mariadb-1   3/3     Running   0          14m

$ kubectl get pod -n demo --context demo-worker
NAME        READY   STATUS    RESTARTS   AGE
mariadb-2   3/3     Running   0          15m


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
    name: mariadb
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

If everything goes well, `KubeDB` Enterprise operator will update the replicas of `MariaDB` object and related `PetSets` and `Pods`.

Let's wait for `MariaDBOpsRequest` to be `Successful`.  Run the following command to watch `MariaDBOpsRequest` CR,

```bash
$ watch kubectl get mariadbopsrequest -n demo
Every 2.0s: kubectl get mariadbopsrequest -n demo
NAME                        TYPE                STATUS       AGE
mdps-scale-horizontal    HorizontalScaling    Successful     106s
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. Now, we are going to verify the number of replicas this database has from the MariaDB object, number of pods the petset have,

```bash
$ kubectl get mariadb -n demo sample-mariadb -o json | jq '.spec.replicas'
5
$ kubectl get sts -n demo sample-mariadb -o json | jq '.spec.replicas'
5
```

