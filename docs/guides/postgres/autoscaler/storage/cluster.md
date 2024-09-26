---
title: Postgres Cluster Autoscaling
menu:
  docs_{{ .version }}:
    identifier: pg-storage-auto-scaling-cluster
    name: Cluster
    parent: pg-storage-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a Postgres Cluster

This guide will show you how to use `KubeDB` to autoscale the storage of a Postgres Replicaset database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community, Enterprise and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- Install Prometheus from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

- You must have a `StorageClass` that supports volume expansion.

- You should be familiar with the following `KubeDB` concepts:
  - [Postgres](/docs/guides/postgres/concepts/postgres.md)
  - [PostgresOpsRequest](/docs/guides/postgres/concepts/opsrequest.md)
  - [Storage Autoscaling Overview](/docs/guides/postgres/autoscaler/storage/overview.md)

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

Now, we are going to deploy a `Postgres` cluster using a supported version by `KubeDB` operator. Then we are going to apply `PostgresAutoscaler` to set up autoscaling.

#### Deploy Postgres Cluster

In this section, we are going to deploy a Postgres cluster database with version `16.1`.  Then, in the next section we will set up autoscaling for this database using `PostgresAutoscaler` CRD. Below is the YAML of the `Postgres` CR that we are going to create,

> If you want to autoscale Postgres `Standalone`, Just remove the `spec.Replicas` from the below yaml and rest of the steps are same.

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: ha-postgres
  namespace: demo
spec:
  version: "16.1"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "topolvm-provisioner"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Postgres` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/autoscaler/storage/ha-postgres.yaml
postgres.kubedb.com/ha-postgres created
```

Now, wait until `ha-postgres` has status `Ready`. i.e,

```bash
$ kubectl get postgres -n demo
NAME             VERSION   STATUS   AGE
ha-postgres        16.1    Ready    3m46s
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get sts -n demo ha-postgres -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS          REASON   AGE
pvc-43266d76-f280-4cca-bd78-d13660a84db9   1Gi        RWO            Delete           Bound    demo/data-ha-postgres-2   topolvm-provisioner            57s
pvc-4a509b05-774b-42d9-b36d-599c9056af37   1Gi        RWO            Delete           Bound    demo/data-ha-postgres-0   topolvm-provisioner            58s
pvc-c27eee12-cd86-4410-b39e-b1dd735fc14d   1Gi        RWO            Delete           Bound    demo/data-ha-postgres-1   topolvm-provisioner            57s
```

You can see the petset has 1GB storage, and the capacity of all the persistent volume is also 1GB.

We are now ready to apply the `PostgresAutoscaler` CRO to set up storage autoscaling for this database.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a `PostgresAutoscaler` Object.

#### Create PostgresAutoscaler Object

In order to set up vertical autoscaling for this cluster database, we have to create a `PostgresAutoscaler` CRO with our desired configuration. Below is the YAML of the `PostgresAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: PostgresAutoscaler
metadata:
  name: pg-as-st
  namespace: demo
spec:
  databaseRef:
    name: ha-postgres
  storage:
    postgres:
      trigger: "On"
      usageThreshold: 20
      scalingThreshold: 20
      expansionMode: "Online"
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `ha-postgres` database.
- `spec.storage.postgres.trigger` specifies that storage autoscaling is enabled for this database.
- `spec.storage.postgres.usageThreshold` specifies storage usage threshold, if storage usage exceeds `20%` then storage autoscaling will be triggered.
- `spec.storage.postgres.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `20%` of the current amount.
- `spec.storage.postgres.expansionMode` specifies the expansion mode of volume expansion `PostgresOpsRequest` created by `PostgresAutoscaler`. topolvm-provisioner supports online volume expansion so here `expansionMode` is set as "Online".

Let's create the `PostgresAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/autoscaler/storage/pgas-storage.yaml
postgresautoscaler.autoscaling.kubedb.com/pg-as-st created
```

#### Storage Autoscaling is set up successfully

Let's check that the `postgresautoscaler` resource is created successfully,

```bash
$ kubectl get postgresautoscaler -n demo
NAME           AGE
pg-as-st   33s

$ kubectl describe postgresautoscaler pg-as-st -n demo
Name:         pg-as-st
Namespace:    demo
Labels:       <none>
Annotations:  API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         PostgresAutoscaler
Metadata:
  Creation Timestamp:  2022-01-14T06:08:02Z
  Generation:          1
  Managed Fields:
    ...
  Resource Version:  24009
  UID:               4f45a3b3-fc72-4d04-b52c-a770944311f6
Spec:
  Database Ref:
    Name:  ha-postgres
  Storage:
    Mariadb:
      Scaling Threshold:  20
      Trigger:            On
      Usage Threshold:    20
Events:                   <none>
```

So, the `postgresautoscaler` resource is created successfully.

Now, for this demo, we are going to manually fill up the persistent volume to exceed the `usageThreshold` using `dd` command to see if storage autoscaling is working or not.

Let's exec into the database pod and fill the database volume(`/var/pv/data`) using the following commands:

```bash
$ kubectl exec -it -n demo ha-postgres-0 -- bash
root@ha-postgres-0:/ df -h /var/pv/data
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/57cd4330-784f-42c1-bf8e-e743241df164 1014M  357M  658M  36% /var/pv/data
root@ha-postgres-0:/ dd if=/dev/zero of=/var/pv/data/file.img bs=500M count=1
1+0 records in
1+0 records out
524288000 bytes (524 MB, 500 MiB) copied, 0.340877 s, 1.5 GB/s
root@ha-postgres-0:/ df -h /var/pv/data
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/57cd4330-784f-42c1-bf8e-e743241df164 1014M  857M  158M  85% /var/pv/data
```

So, from the above output we can see that the storage usage is 83%, which exceeded the `usageThreshold` 20%.

Let's watch the `postgresopsrequest` in the demo namespace to see if any `postgresopsrequest` object is created. After some time you'll see that a `postgresopsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
$ kubectl get postgresopsrequest -n demo
NAME                         TYPE              STATUS        AGE
pgops-ha-postgres-xojkua   VolumeExpansion   Progressing   15s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get postgresopsrequest -n demo
NAME                         TYPE              STATUS       AGE
pgops-ha-postgres-xojkua   VolumeExpansion   Successful   97s
```

We can see from the above output that the `PostgresOpsRequest` has succeeded. If we describe the `PostgresOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe postgresopsrequest -n demo pgops-ha-postgres-xojkua
Name:         pgops-ha-postgres-xojkua
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=ha-postgres
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=postgress.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PostgresOpsRequest
Metadata:
  Creation Timestamp:  2022-01-14T06:13:10Z
  Generation:          1
  Managed Fields: ...
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  PostgresAutoscaler
    Name:                  pg-as-st
    UID:                   4f45a3b3-fc72-4d04-b52c-a770944311f6
  Resource Version:        25557
  UID:                     90763a49-a03f-407c-a233-fb20c4ab57d7
Spec:
  Database Ref:
    Name:  ha-postgres
  Type:    VolumeExpansion
  Volume Expansion:
    Mariadb:  1594884096
Status:
  Conditions:
    Last Transition Time:  2022-01-14T06:13:10Z
    Message:               Controller has started to Progress the PostgresOpsRequest: demo/mops-ha-postgres-xojkua
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-01-14T06:14:25Z
    Message:               Volume Expansion performed successfully in Postgres pod for PostgresOpsRequest: demo/mops-ha-postgres-xojkua
    Observed Generation:   1
    Reason:                SuccessfullyVolumeExpanded
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2022-01-14T06:14:25Z
    Message:               Controller has successfully expand the volume of Postgres demo/mops-ha-postgres-xojkua
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     3
  Phase:                   Successful
Events:
  Type    Reason      Age    From                        Message
  ----    ------      ----   ----                        -------
  Normal  Starting    2m58s  KubeDB Enterprise Operator  Start processing for PostgresOpsRequest: demo/mops-ha-postgres-xojkua
  Normal  Starting    2m58s  KubeDB Enterprise Operator  Pausing Postgres databse: demo/ha-postgres
  Normal  Successful  2m58s  KubeDB Enterprise Operator  Successfully paused Postgres database: demo/ha-postgres for PostgresOpsRequest: mops-ha-postgres-xojkua
  Normal  Successful  103s   KubeDB Enterprise Operator  Volume Expansion performed successfully in Postgres pod for PostgresOpsRequest: demo/mops-ha-postgres-xojkua
  Normal  Starting    103s   KubeDB Enterprise Operator  Updating Postgres storage
  Normal  Successful  103s   KubeDB Enterprise Operator  Successfully Updated Postgres storage
  Normal  Starting    103s   KubeDB Enterprise Operator  Resuming Postgres database: demo/ha-postgres
  Normal  Successful  103s   KubeDB Enterprise Operator  Successfully resumed Postgres database: demo/ha-postgres
  Normal  Successful  103s   KubeDB Enterprise Operator  Controller has Successfully expand the volume of Postgres: demo/ha-postgres
```

Now, we are going to verify from the `Petset`, and the `Persistent Volume` whether the volume of the cluster database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get sts -n demo ha-postgres -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1594884096"
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS          REASON   AGE
pvc-43266d76-f280-4cca-bd78-d13660a84db9   2Gi        RWO            Delete           Bound    demo/data-ha-postgres-2   topolvm-provisioner            23m
pvc-4a509b05-774b-42d9-b36d-599c9056af37   2Gi        RWO            Delete           Bound    demo/data-ha-postgres-0   topolvm-provisioner            24m
pvc-c27eee12-cd86-4410-b39e-b1dd735fc14d   2Gi        RWO            Delete           Bound    demo/data-ha-postgres-1   topolvm-provisioner            23m
```

The above output verifies that we have successfully autoscaled the volume of the Postgres cluster database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete postgres -n demo ha-postgres
kubectl delete postgresautoscaler -n demo pg-as-st
kubectl delete ns demo
```
