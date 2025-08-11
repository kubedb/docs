---
title: MaxScale Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-volume-expansion-maxscale
    name: MaxScale Volume Expansion
    parent: guides-mariadb-volume-expansion
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MaxScale Volume Expansion

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of MaxScale server.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [MariaDB](/docs/guides/mariadb/concepts/mariadb)
    - [MariaDB Replication](/docs/guides/mariadb/clustering/mariadb-replication)
    - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/opsrequest)
    - [Volume Expansion Overview](/docs/guides/mariadb/volume-expansion/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Expand Volume of MaxScale

Here, we are going to deploy a  `MariaDB` cluster in replication mode using a supported version by `KubeDB` operator. Then we are going to apply `MariaDBOpsRequest` to expand its volume.

### Prepare MariaDB Database

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path             rancher.io/local-path   Delete          WaitForFirstConsumer   false                  46h
longhorn               driver.longhorn.io      Delete          Immediate              true                   2m27s
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   2m24s
```

We can see from the output that `longhorn` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We will use this storage class. You can install longhorn from [here](https://longhorn.io/docs/1.9.0/deploy/install/install-with-kubectl/).

Now, we are going to deploy a `MariaDB` database with `MaxScale` in replication mode.

### Deploy MariaDB with MaxScale

In this section, we are going to deploy a MariaDB database along with `MaxScale` with 50Mi volume. 
Then, in the next section we will expand its volume to 100Mi using `MariaDBOpsRequest` CRD. Below is the YAML of the `MariaDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: mariadb-replication
  namespace: demo
spec:
  version: "10.5.23"
  replicas: 3
  topology:
    mode: MariaDBReplication
    maxscale:
      replicas: 3
      enableUI: true
      storageType: Durable
      storage:
        storageClassName: "longhorn"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 50Mi
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `MariaDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/volume-expansion/md-replication.yaml
mariadb.kubedb.com/md-replication created
```

Now, wait until `mariadb-replication` has status `Ready`. i.e,

```bash
$ kubectl get mariadb -n demo
NAME             VERSION   STATUS   AGE
md-replication   10.5.23   Ready    2m30s
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo md-replication-mx -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"50Mi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                           STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-27e4f4b2-289b-44bb-97a2-729d9420f668   1Gi        RWO            Delete           Bound    demo/data-md-replication-2      longhorn       <unset>                          3m48s
pvc-2df9a141-5d32-4c92-b0ec-a8043975c2ae   1Gi        RWO            Delete           Bound    demo/data-md-replication-1      longhorn       <unset>                          3m48s
pvc-7609183e-f9a5-4177-b260-0d24796fb04c   1Gi        RWO            Delete           Bound    demo/data-md-replication-0      longhorn       <unset>                          3m48s
pvc-96449ed7-305e-4857-a2b6-6eda33c99207   50Mi       RWO            Delete           Bound    demo/data-md-replication-mx-2   longhorn       <unset>                          3m51s
pvc-c1424029-4a52-4ff4-9888-14d5e7b4fb61   50Mi       RWO            Delete           Bound    demo/data-md-replication-mx-0   longhorn       <unset>                          3m51s
pvc-d12d301c-58bd-4c59-bd5a-d9167df2b53d   50Mi       RWO            Delete           Bound    demo/data-md-replication-mx-1   longhorn       <unset>                          3m51s

```

You can see that `MaxScale` petset has 50Mi storage, and the capacity of the `MaxScale` persistent volumes are also 50Mi.

We are now ready to apply the `MariaDBOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the `MaxScale` cluster.

#### Create MariaDBOpsRequest

In order to expand the volume of the maxscale server, we have to create a `MariaDBOpsRequest` CR with our desired volume size. Below is the YAML of the `MariaDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: maxscale-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: md-replication
  volumeExpansion:
    mode: Online
    maxscale: 100Mi
```

Here,
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `md-replication` database.
- `spec.volumeExpansion.mode` specifies the desired volume expansion mode (`Online` or `Offline`). Storageclass `longhorn` supports `Online` volume expansion.
- `spec.volumeExpansion.maxscale` specifies the desired volume size of maxscale server.

> **Note:** If the Storageclass you are using doesn't support `Online` Volume Expansion, Try offline volume expansion by using `spec.volumeExpansion.mode:"Offline"`. The pods need to be restarted during offline volume expansion.

Let's create the `MariaDBOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/volume-expansion/maxscale-volume-expansion.yaml
mariadbopsrequest.ops.kubedb.com/maxscale-volume-expansion created
```

#### Verify MaxScale volume expanded successfully

If everything goes well, `KubeDB` Ops-manager operator will update the volume size of `MaxScale` object and related `PetSets` and `Persistent Volumes`.

Let's wait for `MariaDBOpsRequest` to be `Successful`.  Run the following command to watch `MariaDBOpsRequest` CR,

```bash
$ kubectl get mariadbopsrequest -n demo
NAME                        TYPE              STATUS       AGE
maxscale-volume-expansion   VolumeExpansion   Successful   3m
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. If we describe the `MariaDBOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe mariadbopsrequest -n demo maxscale-volume-expansion
Name:         maxscale-volume-expansion
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MariaDBOpsRequest
Metadata:
  Creation Timestamp:  2025-07-23T06:46:12Z
  Generation:          1
  Resource Version:    96939
  UID:                 c33c6a08-2514-4818-932e-f4460bd612a7
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  md-replication
  Type:    VolumeExpansion
  Volume Expansion:
    Maxscale:  100Mi
    Mode:      Online
Status:
  Conditions:
    Last Transition Time:  2025-07-23T06:46:12Z
    Message:               Controller has started to Progress the MariaDBOpsRequest: demo/maxscale-volume-expansion
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2025-07-23T06:46:20Z
    Message:               all pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  AllPvc
    Last Transition Time:  2025-07-23T06:46:20Z
    Message:               is pvc patch; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPvcPatch
    Last Transition Time:  2025-07-23T06:47:10Z
    Message:               delete pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePvc
    Last Transition Time:  2025-07-23T06:47:15Z
    Message:               Online Volume Expansion performed successfully in MaxScale pod for MariaDBOpsRequest: demo/maxscale-volume-expansion
    Observed Generation:   1
    Reason:                VolumeExpansionSucceeded
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2025-07-23T06:47:15Z
    Message:               Controller has successfully expand the volume of MaxScale: demo/maxscale-volume-expansion
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                              Age   From                         Message
  ----     ------                              ----  ----                         -------
  Normal   Starting                            15m   KubeDB Ops-manager Operator  Start processing for MariaDBOpsRequest: demo/maxscale-volume-expansion
  Normal   Starting                            15m   KubeDB Ops-manager Operator  Pausing MariaDB databse: demo/md-replication
  Normal   Successful                          15m   KubeDB Ops-manager Operator  Successfully paused MariaDB database: demo/md-replication for MariaDBOpsRequest: maxscale-volume-expansion
  Warning  all pvc; ConditionStatus:True       15m   KubeDB Ops-manager Operator  all pvc; ConditionStatus:True
  Warning  is pvc patch; ConditionStatus:True  15m   KubeDB Ops-manager Operator  is pvc patch; ConditionStatus:True
  Warning  is pvc patch; ConditionStatus:True  15m   KubeDB Ops-manager Operator  is pvc patch; ConditionStatus:True
  Warning  is pvc patch; ConditionStatus:True  15m   KubeDB Ops-manager Operator  is pvc patch; ConditionStatus:True
  Warning  all pvc; ConditionStatus:True       15m   KubeDB Ops-manager Operator  all pvc; ConditionStatus:True
  Warning  all pvc; ConditionStatus:True       15m   KubeDB Ops-manager Operator  all pvc; ConditionStatus:True
  Warning  all pvc; ConditionStatus:True       15m   KubeDB Ops-manager Operator  all pvc; ConditionStatus:True
  Warning  all pvc; ConditionStatus:True       15m   KubeDB Ops-manager Operator  all pvc; ConditionStatus:True
  Warning  all pvc; ConditionStatus:True       15m   KubeDB Ops-manager Operator  all pvc; ConditionStatus:True
  Warning  all pvc; ConditionStatus:True       15m   KubeDB Ops-manager Operator  all pvc; ConditionStatus:True
  Warning  all pvc; ConditionStatus:True       15m   KubeDB Ops-manager Operator  all pvc; ConditionStatus:True
  Warning  all pvc; ConditionStatus:True       15m   KubeDB Ops-manager Operator  all pvc; ConditionStatus:True
  Warning  all pvc; ConditionStatus:True       15m   KubeDB Ops-manager Operator  all pvc; ConditionStatus:True
  Warning  all pvc; ConditionStatus:True       14m   KubeDB Ops-manager Operator  all pvc; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True    14m   KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  all pvc; ConditionStatus:True       14m   KubeDB Ops-manager Operator  all pvc; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True    14m   KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Normal   Successful                          14m   KubeDB Ops-manager Operator  Online Volume Expansion performed successfully in MaxScale pod for MariaDBOpsRequest: demo/maxscale-volume-expansion
  Normal   Starting                            14m   KubeDB Ops-manager Operator  Updating MaxScale storage
  Normal   Successful                          14m   KubeDB Ops-manager Operator  Successfully updated MaxScale storage
  Normal   Starting                            14m   KubeDB Ops-manager Operator  Resuming MariaDB database: demo/md-replication
  Normal   Successful                          14m   KubeDB Ops-manager Operator  Successfully resumed MariaDB database: demo/md-replication
  Normal   Successful                          14m   KubeDB Ops-manager Operator  Controller has Successfully expand the volume of MaxScale: demo/md-replication

```

Now, we are going to verify from the `Petset`, and the `Persistent Volumes` whether the volume of the database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get petset -n demo md-replication-mx -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"100Mi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                           STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-27e4f4b2-289b-44bb-97a2-729d9420f668   1Gi        RWO            Delete           Bound    demo/data-md-replication-2      longhorn       <unset>                          34m
pvc-2df9a141-5d32-4c92-b0ec-a8043975c2ae   1Gi        RWO            Delete           Bound    demo/data-md-replication-1      longhorn       <unset>                          34m
pvc-7609183e-f9a5-4177-b260-0d24796fb04c   1Gi        RWO            Delete           Bound    demo/data-md-replication-0      longhorn       <unset>                          34m
pvc-96449ed7-305e-4857-a2b6-6eda33c99207   100Mi      RWO            Delete           Bound    demo/data-md-replication-mx-2   longhorn       <unset>                          34m
pvc-c1424029-4a52-4ff4-9888-14d5e7b4fb61   100Mi      RWO            Delete           Bound    demo/data-md-replication-mx-0   longhorn       <unset>                          34m
pvc-d12d301c-58bd-4c59-bd5a-d9167df2b53d   100Mi      RWO            Delete           Bound    demo/data-md-replication-mx-1   longhorn       <unset>                          34m

```

The above output verifies that we have successfully expanded the volume of the MariaDB database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete mariadb -n demo md-replication
$ kubectl delete mariadbopsrequest -n demo maxscale-volume-expansion
$ kubectl delete ns demo
```
