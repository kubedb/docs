---
title: Distributed MariaDBOpsRequest
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-distributed-opsrequest
    name: Distributed MariaDBOpsRequest
    parent: guides-mariadb-distributed-horizontalscaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Distributed MariaDB

This guide will show you how to use `KubeDB` Enterprise operator to horizontally scale a distributed MariaDB Galera cluster across multiple Kubernetes clusters.

> **Note:** All other `OpsRequest` operations behave consistently in a distributed environment. Only `HorizontalScaling` has cluster-specific considerations covered in this guide.

## Before You Begin

- At first, you need to have a multi-cluster Kubernetes setup with OCM and KubeSlice configured. Follow the [Distributed MariaDB Overview](/docs/guides/mariadb/distributed/overview) guide to set up the required infrastructure.

- Install `KubeDB` Community and Enterprise operator in your hub cluster following the steps [here](/docs/setup/README.md). Make sure to enable OCM support:

  ```bash
  --set petset.features.ocm.enabled=true
  ```

- You should be familiar with the following `KubeDB` concepts:
  - [MariaDB](/docs/guides/mariadb/concepts/mariadb/)
  - [Distributed MariaDB Overview](/docs/guides/mariadb/distributed/overview/index.md)
  - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/opsrequest/)
  - [Horizontal Scaling Overview](/docs/guides/mariadb/scaling/horizontal-scaling/overview/)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Apply Horizontal Scaling on Distributed Cluster

Here, we are going to deploy a distributed `MariaDB` Galera cluster and then apply horizontal scaling using `MariaDBOpsRequest`.

### Deploy PlacementPolicy

For a distributed MariaDB cluster, the `PlacementPolicy` must be created **before** deploying the database. It defines which replica index is scheduled on which cluster and acts as the upper bound for scaling — the actual running replicas can be any number less than or equal to the total indices defined in the policy.

In this example, the `PlacementPolicy` is configured with **five** replica indices distributed across two clusters. The MariaDB cluster will start with **three** replicas (indices `0`, `1`, `2`), and can be scaled up to **five** (indices `0`–`4`) without modifying the policy.

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
          - 4
      - clusterName: demo-worker
        replicaIndices:
          - 1
          - 3
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

- `spec.clusterSpreadConstraint.distributionRules[].replicaIndices` specifies which MariaDB replica indices are scheduled on that cluster. `demo-controller` will host replicas `0`, `2`, and `4`; `demo-worker` will host replicas `1` and `3`.

Apply the `PlacementPolicy` on the hub (`demo-controller`) cluster:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/distributed/opsrequest/examples/placement-policy.yaml --context demo-controller
placementpolicy.apps.k8s.appscode.com/distributed-mariadb created
```

### Deploy Distributed MariaDB Cluster

In this section, we are going to deploy a distributed MariaDB Galera cluster with version `11.5.2`. Note that `spec.distributed` is set to `true` and `spec.replicas` is `3` — which is less than the five indices defined in the `PlacementPolicy`. The operator will only create pods for indices `0`, `1`, and `2`.

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: mariadb
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

Let's create the `MariaDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/distributed/opsrequest/examples/mariadb.yaml --context demo-controller
mariadb.kubedb.com/mariadb created
```

Now, wait until `mariadb` has status `Ready`:

```bash
$ kubectl get mariadb -n demo --context demo-controller
NAME             VERSION   STATUS   AGE
mariadb   11.5.2    Ready    2m36s
```

Let's check the number of replicas this database has from the MariaDB object:

```bash
$ kubectl get mariadb -n demo mariadb -o json | jq '.spec.replicas'
3
```

The pods are distributed across clusters as defined by the `PlacementPolicy`. Indices `0` and `2` land on `demo-controller`; index `1` lands on `demo-worker`:

```bash
$ kubectl get pods -n demo --context demo-controller
NAME          READY   STATUS    RESTARTS   AGE
mariadb-0     3/3     Running   0          2m30s
mariadb-2     3/3     Running   0          2m30s

$ kubectl get pods -n demo --context demo-worker
NAME          READY   STATUS    RESTARTS   AGE
mariadb-1     3/3     Running   0          2m30s
```

We can also verify the cluster size from inside a MariaDB pod:

```bash
$ kubectl exec -it -n demo mariadb-0 -c mariadb --context demo-controller -- bash
root@mariadb-0:/ mysql -uroot -p$MYSQL_ROOT_PASSWORD -e "show status like 'wsrep_cluster_size';"
+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| wsrep_cluster_size | 3     |
+--------------------+-------+
```

We can see the cluster has 3 nodes. We are now ready to apply `MariaDBOpsRequest` to scale this database.

## Scale Up Replicas

Here, we are going to scale up the replicas from `3` to `5`. Because the `PlacementPolicy` already defines indices `0`–`4`, no changes to the policy are required — the operator knows where to place the new pods.

### Create MariaDBOpsRequest

In order to scale up the replicas, we have to create a `MariaDBOpsRequest` CR with our desired replicas. Below is the YAML of the `MariaDBOpsRequest` CR that we are going to create:

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

- `spec.databaseRef.name` specifies that we are performing horizontal scaling on `mariadb`.
- `spec.type` specifies that we are performing `HorizontalScaling`.
- `spec.horizontalScaling.member` specifies the desired replica count after scaling. This value must not exceed the total number of indices defined in the `PlacementPolicy` (5 in this example).

Let's create the `MariaDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/distributed/opsrequest/examples/mdops-upscale.yaml --context demo-controller
mariadbopsrequest.ops.kubedb.com/mdops-scale-horizontal-up created
```

### Verify Cluster Replicas Scaled Up Successfully

If everything goes well, `KubeDB` Enterprise operator will update the replicas of the `MariaDB` object and create the new pods on the clusters specified by the `PlacementPolicy`.

Let's wait for `MariaDBOpsRequest` to be `Successful`:

```bash
$ watch kubectl get mariadbopsrequest -n demo --context demo-controller
Every 2.0s: kubectl get mariadbopsrequest -n demo
NAME                        TYPE                STATUS       AGE
mdops-scale-horizontal-up   HorizontalScaling   Successful   18m
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. Now, let's verify the number of replicas:

```bash
$ kubectl get mariadb -n demo mariadb -o json | jq '.spec.replicas'
5
```

The two new pods are placed on the clusters according to the `PlacementPolicy` — index `4` on `demo-controller` and index `3` on `demo-worker`:

```bash
$ kubectl get pods -n demo --context demo-controller
NAME        READY   STATUS    RESTARTS   AGE
mariadb-0   3/3     Running   0          58m
mariadb-2   3/3     Running   0          55m
mariadb-4   3/3     Running   0          17m


$ kubectl get pods -n demo --context demo-worker
NAME        READY   STATUS    RESTARTS   AGE
mariadb-1   3/3     Running   0          57m
mariadb-3   3/3     Running   0          19m
```

From all the above outputs we can see that the cluster now has `5` replicas. We have successfully scaled up the distributed MariaDB cluster.

## Scale Down Replicas

Here, we are going to scale down the replicas from `5` to `3`. The `PlacementPolicy` does not need to be modified — the operator will remove the highest-indexed pods (`mariadb-3` and `mariadb-4`) and the remaining pods will continue running as defined by the policy.

### Create MariaDBOpsRequest

In order to scale down the cluster, we have to create a `MariaDBOpsRequest` CR with our desired replicas. Below is the YAML of the `MariaDBOpsRequest` CR that we are going to create:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: mdops-scale-horizontal-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: mariadb
  horizontalScaling:
    member: 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down on `mariadb`.
- `spec.type` specifies that we are performing `HorizontalScaling`.
- `spec.horizontalScaling.member` specifies the desired replica count after scaling.

Let's create the `MariaDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/distributed/opsrequest/examples/mdops-downscale.yaml --context demo-controller
mariadbopsrequest.ops.kubedb.com/mdops-scale-horizontal-down created
```

### Verify Cluster Replicas Scaled Down Successfully

If everything goes well, `KubeDB` Enterprise operator will update the replicas of the `MariaDB` object and remove the highest-indexed pods from their respective clusters.

Let's wait for `MariaDBOpsRequest` to be `Successful`:

```bash
$ watch kubectl get mariadbopsrequest -n demo --context demo-controller
Every 2.0s: kubectl get mariadbopsrequest -n demo
NAME                          TYPE               STATUS       AGE
mdops-scale-horizontal-down   HorizontalScaling  Successful   2m32s
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. Now, let's verify the number of replicas:

```bash
$ kubectl get mariadb -n demo mariadb -o json | jq '.spec.replicas'
3
```

Pods `mariadb-3` and `mariadb-4` have been removed from their respective clusters:

```bash
$ kubectl get pods -n demo --context demo-controller
NAME          READY   STATUS    RESTARTS   AGE
mariadb-0     3/3     Running   0          20m
mariadb-2     3/3     Running   0          20m

$ kubectl get pods -n demo --context demo-worker
NAME          READY   STATUS    RESTARTS   AGE
mariadb-1     3/3     Running   0          20m
```

Let's verify the cluster size from inside a MariaDB pod:

```bash
$ kubectl exec -it -n demo mariadb-0 -c mariadb --context demo-controller -- bash
root@mariadb-0:/ mysql -uroot -p$MYSQL_ROOT_PASSWORD -e "show status like 'wsrep_cluster_size';"
+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| wsrep_cluster_size | 3     |
+--------------------+-------+
```

From all the above outputs we can see that the cluster now has `3` replicas. We have successfully scaled down the distributed MariaDB cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete mariadb -n demo mariadb --context demo-controller
$ kubectl delete mariadbopsrequest -n demo mdops-scale-horizontal-up mdops-scale-horizontal-down --context demo-controller
$ kubectl delete placementpolicy distributed-mariadb --context demo-controller
```
