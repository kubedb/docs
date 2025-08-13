---
title: Ignite Storage Autoscaling
menu:
  docs_{{ .version }}:
    identifier: ig-autoscaling-storage-description
    name: Autoscale Storage
    parent: ig-autoscaling-storage
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a Ignite Cluster

This guide will show you how to use `KubeDB` to autoscale the storage of a Ignite cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community, Enterprise and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- Install Prometheus from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

- You must have a `StorageClass` that supports volume expansion.

- You should be familiar with the following `KubeDB` concepts:
  - [Ignite](/docs/guides/ignite/concepts/ignite.md)
  - [IgniteAutoscaler](/docs/guides/ignite/concepts/autoscaler.md)
  - [IgniteOpsRequest](/docs/guides/ignite/concepts/opsrequest.md)
  - [Storage Autoscaling Overview](/docs/guides/ignite/autoscaler/storage/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Storage Autoscaling of Cluster Database

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                  PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)    rancher.io/local-path   Delete          WaitForFirstConsumer   false                  79m
topolvm-provisioner   topolvm.cybozu.com      Delete          WaitForFirstConsumer   true                   78m
```

We can see from the output the `topolvm-provisioner` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it. You can install topolvm from [here](https://github.com/topolvm/topolvm)

Now, we are going to deploy a `Ignite` cluster using a supported version by `KubeDB` operator. Then we are going to apply `IgniteAutoscaler` to set up autoscaling.

#### Deploy Ignite Cluster

In this section, we are going to deploy a Ignite cluster with version `3.13.2`.  Then, in the next section we will set up autoscaling for this database using `IgniteAutoscaler` CRD. Below is the YAML of the `Ignite` CR that we are going to create,

> If you want to autoscale Ignite `Standalone`, Just remove the `spec.Replicas` from the below yaml and rest of the steps are same.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Ignite
metadata:
  name: ignite-autoscale
  namespace: demo
spec:
  version: "3.13.2"
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: WipeOut
  podTemplate:
    spec:
      containers:
        - name: ignite
          resources:
            requests:
              cpu: "0.5m"
              memory: "1Gi"
            limits:
              cpu: "1"
              memory: "2Gi"
  serviceTemplates:
    - alias: primary
      spec:
        type: LoadBalancer
```

Let's create the `Ignite` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/ignite/autoscaler/storage/cluster/examples/sample-ignite.yaml
ignite.kubedb.com/ignite-autoscale created
```

Now, wait until `ignite-autoscale` has status `Ready`. i.e,

```bash
$ kubectl get ignite -n demo
NAME                 VERSION   STATUS   AGE
ignite-autoscale   3.13.2    Ready    3m46s
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo ignite-autoscale -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS          REASON   AGE
pvc-43266d76-f280-4cca-bd78-d13660a84db9   1Gi        RWO            Delete           Bound    demo/data-sample-ignite-2   topolvm-provisioner            57s
pvc-4a509b05-774b-42d9-b36d-599c9056af37   1Gi        RWO            Delete           Bound    demo/data-sample-ignite-0   topolvm-provisioner            58s
pvc-c27eee12-cd86-4410-b39e-b1dd735fc14d   1Gi        RWO            Delete           Bound    demo/data-sample-ignite-1   topolvm-provisioner            57s
```

You can see the petset has 1GB storage, and the capacity of all the persistent volume is also 1GB.

We are now ready to apply the `IgniteAutoscaler` CRO to set up storage autoscaling for this database.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a IgniteAutoscaler Object.

#### Create IgniteAutoscaler Object

In order to set up vertical autoscaling for this replicaset database, we have to create a `IgniteAutoscaler` CRO with our desired configuration. Below is the YAML of the `IgniteAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: IgniteAutoscaler
metadata:
  name: ignite-storage-autosclaer
  namespace: demo
spec:
  databaseRef:
    name: ignite-autoscale
  storage:
    ignite:
      expansionMode: "Offline"
      trigger: "On"
      usageThreshold: 20
      scalingThreshold: 30
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `ignite-autoscale` database.
- `spec.storage.ignite.trigger` specifies that storage autoscaling is enabled for this database.
- `spec.storage.ignite.usageThreshold` specifies storage usage threshold, if storage usage exceeds `20%` then storage autoscaling will be triggered.
- `spec.storage.ignite.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `20%` of the current amount.
- `spec.storage.ignite.expansionMode` specifies the expansion mode of volume expansion `igniteOpsRequest` created by `igniteAutoscaler`. topolvm-provisioner supports online volume expansion so here `expansionMode` is set as "Online".

Let's create the `igniteAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/ignite/autoscaler/storage/cluster/examples/ig-storage-autoscale-ops.yaml
igniteautoscaler.autoscaling.kubedb.com/ignite-storage-autosclaer created
```

#### Storage Autoscaling is set up successfully

Let's check that the `igniteautoscaler` resource is created successfully,

```bash
$ kubectl get igniteautoscaler -n demo
NAME                          AGE
ignite-storage-autosclaer   33s

$ kubectl describe igniteautoscaler ignite-storage-autoscaler -n demo
Name:         ignite-storage-autosclaer
Namespace:    demo
Labels:       <none>
Annotations:  API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         igniteAutoscaler
Metadata:
  Creation Timestamp:  2022-01-14T06:08:02Z
  Generation:          1
  Managed Fields:
    ...
  Resource Version:  24009
  UID:               4f45a3b3-fc72-4d04-b52c-a770944311f6
Spec:
  Database Ref:
    Name:  ignite-autoscale
  Storage:
    ignite:
      Scaling Threshold:  20
      Trigger:            On
      Usage Threshold:    20
Events:                   <none>
```

So, the `igniteautoscaler` resource is created successfully.

For this demo we are going to use an opensource tool to manually publish and consume messages in our cluster. This will eventually fill up the storage and trigger a `igniteopsrequest` once the threshold is breached.

We are going to use a docker image called `perf-test`. It runs producers and consumers to continuously publish and consume messages in Ignite cluster. Here's how to run it on kubernetes using the credentials and the address for operator generated primary service.

```bash
kubectl run perf-test --image=pivotalignite/perf-test -- --uri "amqp://admin:password@ignite-autoscale.demo.svc:5672/"
```

You can check the log for this pod which shows publish and consume rates of messages in Ignite.

```bash
$ kubectl logs pod/perf-test -f
id: test-104606-706, starting consumer #0
id: test-104606-706, starting consumer #0, channel #0
id: test-104606-706, starting producer #0
id: test-104606-706, starting producer #0, channel #0
id: test-104606-706, time 1.000 s, sent: 81286 msg/s, received: 23516 msg/s, min/median/75th/95th/99th consumer latency: 6930/174056/361178/503928/519681 µs
id: test-104606-706, time 2.000 s, sent: 30997 msg/s, received: 30686 msg/s, min/median/75th/95th/99th consumer latency: 529789/902251/1057447/1247103/1258790 µs
id: test-104606-706, time 3.000 s, sent: 29032 msg/s, received: 30418 msg/s, min/median/75th/95th/99th consumer latency: 1262421/1661565/1805425/1953992/1989182 µs
id: test-104606-706, time 4.000 s, sent: 30997 msg/s, received: 31228 msg/s, min/median/75th/95th/99th consumer latency: 1572496/1822873/1938918/2035918/2065812 µs
id: test-104606-706, time 5.000 s, sent: 29032 msg/s, received: 33588 msg/s, min/median/75th/95th/99th consumer latency: 1503867/1729779/1831281/1930593/1968284 µs
id: test-104606-706, time 6.000 s, sent: 32704 msg/s, received: 32493 msg/s, min/median/75th/95th/99th consumer latency: 1503915/1749654/1865878/1953439/1971834 µs
id: test-104606-706, time 7.000 s, sent: 38117 msg/s, received: 30759 msg/s, min/median/75th/95th/99th consumer latency: 1511466/1772387/1854642/1918369/1940327 µs
id: test-104606-706, time 8.000 s, sent: 35088 msg/s, received: 31676 msg/s, min/median/75th/95th/99th consumer latency: 1578860/1799719/1915632/1985467/2024141 µs
id: test-104606-706, time 9.000 s, sent: 29706 msg/s, received: 31375 msg/s, min/median/75th/95th/99th consumer latency: 1516415/1743385/1877037/1972570/1988962 µs
id: test-104606-706, time 10.000 s, sent: 15903 msg/s, received: 26711 msg/s, min/median/75th/95th/99th consumer latency: 1569546/1884700/1992762/2096417/2136613 µs
```

Let's watch the `igniteopsrequest` in the demo namespace to see if any `igniteopsrequest` object is created. After some time you'll see that a `igniteopsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
$ kubectl get igniteopsrequest -n demo
NAME                              TYPE              STATUS        AGE
igops-ignite-autoscale-xojkua   VolumeExpansion   Progressing   15s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get igniteopsrequest -n demo
NAME                              TYPE              STATUS       AGE
igops-ignite-autoscale-xojkua   VolumeExpansion   Successful   97s
```

We can see from the above output that the `IgniteOpsRequest` has succeeded. If we describe the `IgniteOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe igniteopsrequest -n demo igops-ignite-autoscale-xojkua
Name:         igops-ignite-autoscaleq-xojkua
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=ignite-autoscale
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=ignites.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         igniteOpsRequest
Metadata:
  Creation Timestamp:  2022-01-14T06:13:10Z
  Generation:          1
  Managed Fields: ...
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  igniteAutoscaler
    Name:                  ignite-storage-autosclaer
    UID:                   4f45a3b3-fc72-4d04-b52c-a770944311f6
  Resource Version:        25557
  UID:                     90763a49-a03f-407c-a233-fb20c4ab57d7
Spec:
  Database Ref:
    Name:  ignite-autoscale
  Type:    VolumeExpansion
  Volume Expansion:
    ignite:  1594884096
Status:
  Conditions:
    Last Transition Time:  2022-01-14T06:13:10Z
    Message:               Controller has started to Progress the igniteOpsRequest: demo/igops-ignite-autoscale-xojkua
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-01-14T06:14:25Z
    Message:               Volume Expansion performed successfully in ignite pod for igniteOpsRequest: demo/igops-ignite-autoscale-xojkua
    Observed Generation:   1
    Reason:                SuccessfullyVolumeExpanded
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2022-01-14T06:14:25Z
    Message:               Controller has successfully expand the volume of ignite demo/igops-ignite-autoscale-xojkua
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     3
  Phase:                   Successful
Events:
  Type    Reason      Age    From                        Message
  ----    ------      ----   ----                        -------
  Normal  Starting    2m58s  KubeDB Enterprise Operator  Start processing for igniteOpsRequest: demo/igops-ignite-autoscale-xojkua
  Normal  Starting    2m58s  KubeDB Enterprise Operator  Pausing ignite databse: demo/ignite-autoscale
  Normal  Successful  2m58s  KubeDB Enterprise Operator  Successfully paused ignite database: demo/ignite-autoscale for igniteOpsRequest: igops-ignite-autoscale-xojkua
  Normal  Successful  103s   KubeDB Enterprise Operator  Volume Expansion performed successfully in ignite pod for igniteOpsRequest: demo/igops-ignite-autoscale-xojkua
  Normal  Starting    103s   KubeDB Enterprise Operator  Updating ignite storage
  Normal  Successful  103s   KubeDB Enterprise Operator  Successfully Updated ignite storage
  Normal  Starting    103s   KubeDB Enterprise Operator  Resuming ignite database: demo/ignite-autoscale
  Normal  Successful  103s   KubeDB Enterprise Operator  Successfully resumed ignite database: demo/ignite-autoscale
  Normal  Successful  103s   KubeDB Enterprise Operator  Controller has Successfully expand the volume of ignite: demo/ignite-autoscale
```

Now, we are going to verify from the `Petset`, and the `Persistent Volume` whether the volume of the replicaset database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get petset -n demo ignite-autoscale -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1594884096"
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS          REASON   AGE
pvc-43266d76-f280-4cca-bd78-d13660a84db9   2Gi        RWO            Delete           Bound    demo/data-ignite-autoscale-2   topolvm-provisioner            23m
pvc-4a509b05-774b-42d9-b36d-599c9056af37   2Gi        RWO            Delete           Bound    demo/data-signite-autoscale-0   topolvm-provisioner            24m
pvc-c27eee12-cd86-4410-b39e-b1dd735fc14d   2Gi        RWO            Delete           Bound    demo/data-ignite-autoscale-1   topolvm-provisioner            23m
```

The above output verifies that we have successfully autoscaled the volume of the ignite cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete ignite -n demo ignite-autoscale
kubectl delete igniteautoscaler -n demo igops-ignite-autoscale-xojkua
kubectl delete ns demo
```
