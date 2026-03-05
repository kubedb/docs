---
title: Distributed MariaDB Cluster Storage Autoscaling
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-distributed-autoscaling-storage-cluster
    name: Cluster
    parent: guides-mariadb-distributed-autoscaling-storage
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a Distributed MariaDB Cluster

This guide will show you how to use `KubeDB` to autoscale the storage of a distributed MariaDB Galera cluster deployed across multiple Kubernetes clusters.

## Before You Begin

- At first, you need to have a multi-cluster Kubernetes setup with OCM and KubeSlice configured. Follow the [Distributed MariaDB Overview](/docs/guides/mariadb/distributed/overview) guide to set up the required infrastructure.

- Install `KubeDB` Community, Enterprise and Autoscaler operator in your hub cluster following the steps [here](/docs/setup/README.md). Make sure to enable OCM support:
  ```bash
  --set petset.features.ocm.enabled=true
  ```

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- Install Prometheus in each spoke cluster. The autoscaler queries the per-cluster Prometheus endpoint (configured in the `PlacementPolicy`) to collect storage usage metrics. You can install it from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

- You must have a `StorageClass` that supports volume expansion in each spoke cluster.

- You should be familiar with the following `KubeDB` concepts:
  - [MariaDB](/docs/guides/mariadb/concepts/mariadb)
  - [MariaDBAutoscaler](/docs/guides/mariadb/concepts/autoscaler)
  - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/opsrequest)
  - [Distributed MariaDB Storage Autoscaling Overview](/docs/guides/mariadb/distributed/autoscaler/storage/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Storage Autoscaling of Distributed Cluster Database

At first verify that your clusters have a storage class that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass --context demo-worker
NAME                  PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)    rancher.io/local-path   Delete          WaitForFirstConsumer   false                  79m
topolvm-provisioner   topolvm.cybozu.com      Delete          WaitForFirstConsumer   true                   78m
```

We can see from the output the `topolvm-provisioner` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it. You can install topolvm from [here](https://github.com/topolvm/topolvm)

### Deploy PlacementPolicy

For distributed MariaDB autoscaling, the `PlacementPolicy` must include a `monitoring.prometheus.url` for each spoke cluster. The autoscaler uses these endpoints to monitor storage usage across all clusters where MariaDB pods are running.

Below is the YAML of the `PlacementPolicy` that we are going to create. It distributes 4 replicas across two clusters and provides the Prometheus endpoint for each:

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
        monitoring:
          prometheus:
            url: http://prometheus-operated.monitoring.svc.cluster.local:9090
        replicaIndices:
          - 0
          - 2
      - clusterName: demo-worker
        monitoring:
          prometheus:
            url: http://prometheus-operated.monitoring.svc.cluster.local:9090
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

- `spec.clusterSpreadConstraint.distributionRules[].monitoring.prometheus.url` specifies the Prometheus endpoint for the corresponding spoke cluster. The autoscaler uses this URL to scrape storage usage metrics for pods running in that cluster.
- `spec.clusterSpreadConstraint.distributionRules[].replicaIndices` specifies which MariaDB replica indices are scheduled on that cluster. Here `demo-controller` hosts replicas `0` and `2`, and `demo-worker` hosts replica `1`.

Apply the `PlacementPolicy` on the hub (`demo-controller`) cluster:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/distributed/autoscaler/storage/cluster/examples/placement-policy.yaml --context demo-controller
placementpolicy.apps.k8s.appscode.com/distributed-mariadb created
```

> **Note:** Update the `monitoring.prometheus.url` values to match the actual Prometheus service endpoints in each of your spoke clusters.

### Deploy Distributed MariaDB Cluster

In this section, we are going to deploy a distributed MariaDB replicaset database with version `11.5.2`. Then, in the next section we will set up autoscaling for this database using `MariaDBAutoscaler` CRD.

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
    storageClassName: "topolvm-provisioner"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  podTemplate:
    spec:
      podPlacementPolicy:
        name: distributed-mariadb
  deletionPolicy: WipeOut
```

Let's create the `MariaDB` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/distributed/autoscaler/storage/cluster/examples/sample-mariadb.yaml --context demo-controller
mariadb.kubedb.com/sample-mariadb created
```

Now, wait until `sample-mariadb` has status `Ready`. i.e,

```bash
$ kubectl get mariadb -n demo --context demo-controller
NAME             VERSION   STATUS   AGE
sample-mariadb   11.5.2   Ready    3m46s
```

The pods are distributed across clusters as defined by the `PlacementPolicy`:

```bash
$ kubectl get pod -n demo --context demo-controller
NAME               READY   STATUS    RESTARTS   AGE
sample-mariadb-0   3/3     Running   0          3m46s
sample-mariadb-2   3/3     Running   0          3m46s

$ kubectl get pod -n demo --context demo-worker
NAME               READY   STATUS    RESTARTS   AGE
sample-mariadb-1   3/3     Running   0          3m46s
```

Let's check volume size from petset, and from the persistent volume on `demo-worker`,

```bash
$ kubectl get sts -n demo sample-mariadb -o json --context demo-worker | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo --context demo-worker
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS          REASON   AGE
pvc-c27eee12-cd86-4410-b39e-b1dd735fc14d   1Gi        RWO            Delete           Bound    demo/data-sample-mariadb-1   topolvm-provisioner            57s
```

You can see the petset has 1GB storage, and the capacity of all the persistent volumes is also 1GB.

We are now ready to apply the `MariaDBAutoscaler` CRO to set up storage autoscaling for this database.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a MariaDBAutoscaler Object.

#### Create MariaDBAutoscaler Object

In order to set up vertical autoscaling for this distributed replicaset database, we have to create a `MariaDBAutoscaler` CRO with our desired configuration. Below is the YAML of the `MariaDBAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MariaDBAutoscaler
metadata:
  name: md-as-st
  namespace: demo
spec:
  databaseRef:
    name: sample-mariadb
  storage:
    mariadb:
      trigger: "On"
      usageThreshold: 20
      scalingThreshold: 20
      expansionMode: "Online"
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `sample-mariadb` database.
- `spec.storage.mariadb.trigger` specifies that storage autoscaling is enabled for this database.
- `spec.storage.mariadb.usageThreshold` specifies storage usage threshold, if storage usage exceeds `20%` then storage autoscaling will be triggered. The autoscaler monitors storage usage by querying the Prometheus endpoint of the spoke cluster where each pod is running, as configured in the `PlacementPolicy`.
- `spec.storage.mariadb.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `20%` of the current amount.
- `spec.storage.mariadb.expansionMode` specifies the expansion mode of volume expansion `MariaDBOpsRequest` created by `MariaDBAutoscaler`. topolvm-provisioner supports online volume expansion so here `expansionMode` is set as "Online".

Let's create the `MariaDBAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/distributed/autoscaler/storage/cluster/examples/mdas-storage.yaml --context demo-controller
mariadbautoscaler.autoscaling.kubedb.com/md-as-st created
```

#### Storage Autoscaling is set up successfully

Let's check that the `mariadbautoscaler` resource is created successfully,

```bash
$ kubectl get mariadbautoscaler -n demo --context demo-controller
NAME           AGE
md-as-st   33s

$ kubectl describe mariadbautoscaler md-as-st -n demo --context demo-controller
Name:         md-as-st
Namespace:    demo
Labels:       <none>
Annotations:  API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         MariaDBAutoscaler
Metadata:
  Creation Timestamp:  2022-01-14T06:08:02Z
  Generation:          1
  ...
  Resource Version:  24009
  UID:               4f45a3b3-fc72-4d04-b52c-a770944311f6
Spec:
  Database Ref:
    Name:  sample-mariadb
  Storage:
    Mariadb:
      Scaling Threshold:  20
      Trigger:            On
      Usage Threshold:    20
Events:                   <none>
```

So, the `mariadbautoscaler` resource is created successfully.

Now, for this demo, we are going to manually fill up the persistent volume to exceed the `usageThreshold` using `dd` command to see if storage autoscaling is working or not.

Let's exec into a database pod on `demo-worker` and fill the database volume(`var/lib/mysql`) using the following commands:

```bash
$ kubectl exec -it -n demo sample-mariadb-0 --context demo-worker -- bash
root@sample-mariadb-0:/ df -h /var/lib/mysql
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/57cd4330-784f-42c1-bf8e-e743241df164 1014M  357M  658M  36% /var/lib/mysql
root@sample-mariadb-0:/ dd if=/dev/zero of=/var/lib/mysql/file.img bs=500M count=1
1+0 records in
1+0 records out
524288000 bytes (524 MB, 500 MiB) copied, 0.340877 s, 1.5 GB/s
root@sample-mariadb-0:/ df -h /var/lib/mysql
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/57cd4330-784f-42c1-bf8e-e743241df164 1014M  857M  158M  85% /var/lib/mysql
```

So, from the above output we can see that the storage usage is 85%, which exceeded the `usageThreshold` 20%.

Let's watch the `mariadbopsrequest` in the demo namespace to see if any `mariadbopsrequest` object is created. After some time you'll see that a `mariadbopsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
$ kubectl get mariadbopsrequest -n demo --context demo-controller
NAME                         TYPE              STATUS        AGE
mops-sample-mariadb-xojkua   VolumeExpansion   Progressing   15s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get mariadbopsrequest -n demo --context demo-controller
NAME                         TYPE              STATUS       AGE
mops-sample-mariadb-xojkua   VolumeExpansion   Successful   97s
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. If we describe the `MariaDBOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe mariadbopsrequest -n demo mops-sample-mariadb-xojkua --context demo-controller
Name:         mops-sample-mariadb-xojkua
Namespace:    demo
...
Spec:
  Database Ref:
    Name:  sample-mariadb
  Type:    VolumeExpansion
  Volume Expansion:
    Mariadb:  1594884096
Status:
  Conditions:
    ...
    Last Transition Time:  2022-01-14T06:14:25Z
    Message:               Volume Expansion performed successfully in MariaDB pod for MariaDBOpsRequest: demo/mops-sample-mariadb-xojkua
    Observed Generation:   1
    Reason:                SuccessfullyVolumeExpanded
    Status:                True
    Type:                  VolumeExpansion
    ...
  Phase:  Successful
```

Now, we are going to verify from the `Petset` and the `Persistent Volume` whether the volume of the distributed replicaset database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get sts -n demo sample-mariadb -o json --context demo-worker | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1594884096"
$ kubectl get pv -n demo --context demo-worker
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS          REASON   AGE
pvc-c27eee12-cd86-4410-b39e-b1dd735fc14d   2Gi        RWO            Delete           Bound    demo/data-sample-mariadb-1   topolvm-provisioner            23m
```

The above output verifies that we have successfully autoscaled the volume of the distributed MariaDB cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mariadb -n demo sample-mariadb --context demo-controller
kubectl delete mariadbautoscaler -n demo md-as-st --context demo-controller
kubectl delete placementpolicy distributed-mariadb --context demo-controller
kubectl delete ns demo --context demo-controller
kubectl delete ns demo --context demo-worker
```
