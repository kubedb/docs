---
title: RabbitMQ Storage Autoscaling
menu:
  docs_{{ .version }}:
    identifier: rm-autoscaling-storage-description
    name: storage-autoscaling
    parent: rm-autoscaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a RabbitMQ Cluster

This guide will show you how to use `KubeDB` to autoscale the storage of a RabbitMQ cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community, Enterprise and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- Install Prometheus from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

- You must have a `StorageClass` that supports volume expansion.

- You should be familiar with the following `KubeDB` concepts:
  - [RabbitMQ](/docs/guides/rabbitmq/concepts/rabbitmq.md)
  - [RabbitMQAutoscaler](/docs/guides/rabbitmq/concepts/autoscaler.md)
  - [RabbitMQOpsRequest](/docs/guides/rabbitmq/concepts/opsrequest.md)
  - [Storage Autoscaling Overview](/docs/guides/rabbitmq/autoscaler/storage/overview.md)

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

Now, we are going to deploy a `RabbitMQ` cluster using a supported version by `KubeDB` operator. Then we are going to apply `RabbitMQAutoscaler` to set up autoscaling.

#### Deploy RabbitMQ Cluster

In this section, we are going to deploy a RabbitMQ cluster with version `3.13.2`.  Then, in the next section we will set up autoscaling for this database using `RabbitMQAutoscaler` CRD. Below is the YAML of the `RabbitMQ` CR that we are going to create,

> If you want to autoscale RabbitMQ `Standalone`, Just remove the `spec.Replicas` from the below yaml and rest of the steps are same.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rabbitmq-autoscale
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
        - name: rabbitmq
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

Let's create the `RabbitMQ` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/rabbitmq/autoscaler/storage/cluster/examples/sample-rabbitmq.yaml
rabbitmq.kubedb.com/rabbitmq-autoscale created
```

Now, wait until `rabbitmq-autoscale` has status `Ready`. i.e,

```bash
$ kubectl get rabbitmq -n demo
NAME                 VERSION   STATUS   AGE
rabbitmq-autoscale   3.13.2    Ready    3m46s
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo rabbitmq-autoscale -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS          REASON   AGE
pvc-43266d76-f280-4cca-bd78-d13660a84db9   1Gi        RWO            Delete           Bound    demo/data-sample-rabbitmq-2   topolvm-provisioner            57s
pvc-4a509b05-774b-42d9-b36d-599c9056af37   1Gi        RWO            Delete           Bound    demo/data-sample-rabbitmq-0   topolvm-provisioner            58s
pvc-c27eee12-cd86-4410-b39e-b1dd735fc14d   1Gi        RWO            Delete           Bound    demo/data-sample-rabbitmq-1   topolvm-provisioner            57s
```

You can see the petset has 1GB storage, and the capacity of all the persistent volume is also 1GB.

We are now ready to apply the `RabbitMQAutoscaler` CRO to set up storage autoscaling for this database.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a RabbitMQAutoscaler Object.

#### Create RabbitMQAutoscaler Object

In order to set up vertical autoscaling for this replicaset database, we have to create a `RabbitMQAutoscaler` CRO with our desired configuration. Below is the YAML of the `RabbitMQAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: RabbitMQAutoscaler
metadata:
  name: rabbitmq-storage-autosclaer
  namespace: demo
spec:
  databaseRef:
    name: rabbitmq-autoscale
  storage:
    rabbitmq:
      expansionMode: "Offline"
      trigger: "On"
      usageThreshold: 20
      scalingThreshold: 30
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `rabbitmq-autoscale` database.
- `spec.storage.rabbitmq.trigger` specifies that storage autoscaling is enabled for this database.
- `spec.storage.rabbitmq.usageThreshold` specifies storage usage threshold, if storage usage exceeds `20%` then storage autoscaling will be triggered.
- `spec.storage.rabbitmq.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `20%` of the current amount.
- `spec.storage.rabbitmq.expansionMode` specifies the expansion mode of volume expansion `rabbitmqOpsRequest` created by `rabbitmqAutoscaler`. topolvm-provisioner supports online volume expansion so here `expansionMode` is set as "Online".

Let's create the `rabbitmqAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/rabbitmq/autoscaler/storage/cluster/examples/rm-storage-autoscale-ops.yaml
rabbitmqautoscaler.autoscaling.kubedb.com/rabbitmq-storage-autosclaer created
```

#### Storage Autoscaling is set up successfully

Let's check that the `rabbitmqautoscaler` resource is created successfully,

```bash
$ kubectl get rabbitmqautoscaler -n demo
NAME                          AGE
rabbitmq-storage-autosclaer   33s

$ kubectl describe rabbitmqautoscaler rabbitmq-storage-autoscaler -n demo
Name:         rabbitmq-storage-autosclaer
Namespace:    demo
Labels:       <none>
Annotations:  API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         rabbitmqAutoscaler
Metadata:
  Creation Timestamp:  2022-01-14T06:08:02Z
  Generation:          1
  Managed Fields:
    ...
  Resource Version:  24009
  UID:               4f45a3b3-fc72-4d04-b52c-a770944311f6
Spec:
  Database Ref:
    Name:  rabbitmq-autoscale
  Storage:
    rabbitmq:
      Scaling Threshold:  20
      Trigger:            On
      Usage Threshold:    20
Events:                   <none>
```

So, the `rabbitmqautoscaler` resource is created successfully.

For this demo we are going to use an opensource tool to manually publish and consume messages in our cluster. This will eventually fill up the storage and trigger a `rabbitmqopsrequest` once the threshold is breached.

We are going to use a docker image called `perf-test`. It runs producers and consumers to continuously publish and consume messages in RabbitMQ cluster. Here's how to run it on kubernetes using the credentials and the address for operator generated primary service.

```bash
kubectl run perf-test --image=pivotalrabbitmq/perf-test -- --uri "amqp://admin:password@rabbitmq-autoscale.demo.svc:5672/"
```

You can check the log for this pod which shows publish and consume rates of messages in RabbitMQ.

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

Let's watch the `rabbitmqopsrequest` in the demo namespace to see if any `rabbitmqopsrequest` object is created. After some time you'll see that a `rabbitmqopsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
$ kubectl get rabbitmqopsrequest -n demo
NAME                              TYPE              STATUS        AGE
rmops-rabbitmq-autoscale-xojkua   VolumeExpansion   Progressing   15s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get rabbitmqopsrequest -n demo
NAME                              TYPE              STATUS       AGE
rmops-rabbitmq-autoscale-xojkua   VolumeExpansion   Successful   97s
```

We can see from the above output that the `RabbitMQOpsRequest` has succeeded. If we describe the `RabbitMQOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe rabbitmqopsrequest -n demo rmops-rabbitmq-autoscale-xojkua
Name:         rmops-rabbitmq-autoscaleq-xojkua
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=rabbitmq-autoscale
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=rabbitmqs.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         rabbitmqOpsRequest
Metadata:
  Creation Timestamp:  2022-01-14T06:13:10Z
  Generation:          1
  Managed Fields: ...
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  rabbitmqAutoscaler
    Name:                  rabbitmq-storage-autosclaer
    UID:                   4f45a3b3-fc72-4d04-b52c-a770944311f6
  Resource Version:        25557
  UID:                     90763a49-a03f-407c-a233-fb20c4ab57d7
Spec:
  Database Ref:
    Name:  rabbitmq-autoscale
  Type:    VolumeExpansion
  Volume Expansion:
    rabbitmq:  1594884096
Status:
  Conditions:
    Last Transition Time:  2022-01-14T06:13:10Z
    Message:               Controller has started to Progress the rabbitmqOpsRequest: demo/rmops-rabbitmq-autoscale-xojkua
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-01-14T06:14:25Z
    Message:               Volume Expansion performed successfully in rabbitmq pod for rabbitmqOpsRequest: demo/rmops-rabbitmq-autoscale-xojkua
    Observed Generation:   1
    Reason:                SuccessfullyVolumeExpanded
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2022-01-14T06:14:25Z
    Message:               Controller has successfully expand the volume of rabbitmq demo/rmops-rabbitmq-autoscale-xojkua
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     3
  Phase:                   Successful
Events:
  Type    Reason      Age    From                        Message
  ----    ------      ----   ----                        -------
  Normal  Starting    2m58s  KubeDB Enterprise Operator  Start processing for rabbitmqOpsRequest: demo/rmops-rabbitmq-autoscale-xojkua
  Normal  Starting    2m58s  KubeDB Enterprise Operator  Pausing rabbitmq databse: demo/rabbitmq-autoscale
  Normal  Successful  2m58s  KubeDB Enterprise Operator  Successfully paused rabbitmq database: demo/rabbitmq-autoscale for rabbitmqOpsRequest: rmops-rabbitmq-autoscale-xojkua
  Normal  Successful  103s   KubeDB Enterprise Operator  Volume Expansion performed successfully in rabbitmq pod for rabbitmqOpsRequest: demo/rmops-rabbitmq-autoscale-xojkua
  Normal  Starting    103s   KubeDB Enterprise Operator  Updating rabbitmq storage
  Normal  Successful  103s   KubeDB Enterprise Operator  Successfully Updated rabbitmq storage
  Normal  Starting    103s   KubeDB Enterprise Operator  Resuming rabbitmq database: demo/rabbitmq-autoscale
  Normal  Successful  103s   KubeDB Enterprise Operator  Successfully resumed rabbitmq database: demo/rabbitmq-autoscale
  Normal  Successful  103s   KubeDB Enterprise Operator  Controller has Successfully expand the volume of rabbitmq: demo/rabbitmq-autoscale
```

Now, we are going to verify from the `Petset`, and the `Persistent Volume` whether the volume of the replicaset database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get petset -n demo rabbitmq-autoscale -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1594884096"
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS          REASON   AGE
pvc-43266d76-f280-4cca-bd78-d13660a84db9   2Gi        RWO            Delete           Bound    demo/data-rabbitmq-autoscale-2   topolvm-provisioner            23m
pvc-4a509b05-774b-42d9-b36d-599c9056af37   2Gi        RWO            Delete           Bound    demo/data-srabbitmq-autoscale-0   topolvm-provisioner            24m
pvc-c27eee12-cd86-4410-b39e-b1dd735fc14d   2Gi        RWO            Delete           Bound    demo/data-rabbitmq-autoscale-1   topolvm-provisioner            23m
```

The above output verifies that we have successfully autoscaled the volume of the rabbitmq cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete rabbitmq -n demo rabbitmq-autoscale
kubectl delete rabbitmqautoscaler -n demo rmops-rabbitmq-autoscale-xojkua
kubectl delete ns demo
```
