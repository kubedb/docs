---
title: SingleStore Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-volume-expansion-volume-expansion
    name: SingleStore Volume Expansion
    parent: guides-sdb-volume-expansion
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# SingleStore Volume Expansion

This guide will show you how to use `KubeDB` Enterprise operator to expand the volume of a SingleStore.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [SingleStore](/docs/guides/singlestore/concepts/singlestore.md)
  - [SingleStoreOpsRequest](/docs/guides/singlestore/concepts/opsrequest.md)
  - [Volume Expansion Overview](/docs/guides/singlestore/volume-expansion/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Expand Volume of SingleStore

Here, we are going to deploy a  `SingleStore` cluster using a supported version by `KubeDB` operator. Then we are going to apply `SingleStoreOpsRequest` to expand its volume. The process of expanding SingleStore `standalone` is same as SingleStore cluster.

### Create SingleStore License Secret

We need SingleStore License to create SingleStore Database. So, Ensure that you have acquired a license and then simply pass the license by secret.

```bash
$ kubectl create secret generic -n demo license-secret \
                --from-literal=username=license \
                --from-literal=password='your-license-set-here'
secret/license-secret created
```

### Prepare SingleStore Database

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageClass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  6d2h
longhorn (default)     driver.longhorn.io      Delete          Immediate              true                   3d21h
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   42m
```

Here, we will use `longhorn` storageClass for this tuitorial.

Now, we are going to deploy a `SingleStore` database of 3 replicas with version `8.7.10`.

### Deploy SingleStore

In this section, we are going to deploy a SingleStore Cluster with 1GB volume for `aggregator` nodes and 10GB volume for `leaf` nodes. Then, in the next section we will expand its volume to 2GB using `SingleStoreOpsRequest` CRD. Below is the YAML of the `SingleStore` CR that we are going to create,

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
    name: license-secret
  storageType: Durable
  deletionPolicy: WipeOut
```

Let's create the `SingleStore` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/volume-expansion/volume-expansion/example/sample-sdb.yaml
singlestore.kubedb.com/sample-sdb created
```

Now, wait until `sample-sdb` has status `Ready`. i.e,

```bash
$ kubectl get sdb -n demo
NAME         TYPE                  VERSION   STATUS   AGE
sample-sdb   kubedb.com/v1alpha2   8.7.10    Ready    4m25s

```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo sample-sdb-aggregator -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get petset -n demo sample-sdb-leaf -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"10Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                               STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-41cb892c-99fc-4211-a8c2-4e6f8a16c661   10Gi       RWO            Delete           Bound    demo/data-sample-sdb-leaf-0         longhorn       <unset>                          90s
pvc-6e241724-6577-408e-b8de-9569d7d785c4   10Gi       RWO            Delete           Bound    demo/data-sample-sdb-leaf-1         longhorn       <unset>                          75s
pvc-95ecc525-540b-4496-bf14-bfac901d73c4   1Gi        RWO            Delete           Bound    demo/data-sample-sdb-aggregator-0   longhorn       <unset>                          94s


```

You can see the `aggregator` petset has 1GB storage, and the capacity of all the `aggregator` persistent volumes are also 1GB.

You can see the `leaf` petset has 10GB storage, and the capacity of all the `leaf` persistent volumes are also 10GB.

We are now ready to apply the `SingleStoreOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the SingleStore cluster.

#### Create SingleStoreOpsRequest

In order to expand the volume of the database, we have to create a `SingleStoreOpsRequest` CR with our desired volume size. Below is the YAML of the `SingleStoreOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: md-online-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion  
  databaseRef:
    name: sample-mariadb
  volumeExpansion:   
    mode: "Online"
    mariadb: 2Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `sample-mariadb` database.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.mariadb` specifies the desired volume size.
- `spec.volumeExpansion.mode` specifies the desired volume expansion mode (`Online` or `Offline`). Storageclass `topolvm-provisioner` supports `Online` volume expansion.

> **Note:** If the Storageclass you are using doesn't support `Online` Volume Expansion, Try offline volume expansion by using `spec.volumeExpansion.mode:"Offline"`.

Let's create the `MariaDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/volume-expansion/volume-expansion/example/online-volume-expansion.yaml
mariadbopsrequest.ops.kubedb.com/md-online-volume-expansion created
```

#### Verify MariaDB volume expanded successfully

If everything goes well, `KubeDB` Enterprise operator will update the volume size of `MariaDB` object and related `PetSets` and `Persistent Volumes`.

Let's wait for `MariaDBOpsRequest` to be `Successful`.  Run the following command to watch `MariaDBOpsRequest` CR,

```bash
$ kubectl get mariadbopsrequest -n demo
NAME                         TYPE              STATUS       AGE
md-online-volume-expansion   VolumeExpansion   Successful   96s
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. If we describe the `MariaDBOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe mariadbopsrequest -n demo md-online-volume-expansion
Name:         md-online-volume-expansion
Namespace:    demo
Labels:       <none>
Annotations:  API Version:  ops.kubedb.com/v1alpha1
Kind:         MariaDBOpsRequest
Metadata:
  UID:               09a119aa-4f2a-4cb4-b620-2aa3a514df11
Spec:
  Database Ref:
    Name:  sample-mariadb
  Type:    VolumeExpansion
  Volume Expansion:
    Mariadb:  2Gi
    Mode:     Online
Status:
  Conditions:
    Last Transition Time:  2022-01-07T06:38:29Z
    Message:               Controller has started to Progress the MariaDBOpsRequest: demo/md-online-volume-expansion
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-01-07T06:39:49Z
    Message:               Online Volume Expansion performed successfully in MariaDB pod for MariaDBOpsRequest: demo/md-online-volume-expansion
    Observed Generation:   1
    Reason:                SuccessfullyVolumeExpanded
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2022-01-07T06:39:49Z
    Message:               Controller has successfully expand the volume of MariaDB demo/md-online-volume-expansion
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     3
  Phase:                   Successful
Events:
  Type    Reason      Age   From                        Message
  ----    ------      ----  ----                        -------
  Normal  Starting    2m1s  KubeDB Enterprise Operator  Start processing for MariaDBOpsRequest: demo/md-online-volume-expansion
  Normal  Starting    2m1s  KubeDB Enterprise Operator  Pausing MariaDB databse: demo/sample-mariadb
  Normal  Successful  2m1s  KubeDB Enterprise Operator  Successfully paused MariaDB database: demo/sample-mariadb for MariaDBOpsRequest: md-online-volume-expansion
  Normal  Successful  41s   KubeDB Enterprise Operator  Online Volume Expansion performed successfully in MariaDB pod for MariaDBOpsRequest: demo/md-online-volume-expansion
  Normal  Starting    41s   KubeDB Enterprise Operator  Updating MariaDB storage
  Normal  Successful  41s   KubeDB Enterprise Operator  Successfully Updated MariaDB storage
  Normal  Starting    41s   KubeDB Enterprise Operator  Resuming MariaDB database: demo/sample-mariadb
  Normal  Successful  41s   KubeDB Enterprise Operator  Successfully resumed MariaDB database: demo/sample-mariadb
  Normal  Successful  41s   KubeDB Enterprise Operator  Controller has Successfully expand the volume of MariaDB: demo/sample-mariadb
  
```

Now, we are going to verify from the `Petset`, and the `Persistent Volumes` whether the volume of the database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get sts -n demo sample-mariadb -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"2Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS          REASON   AGE
pvc-331335d1-c8e0-4b73-9dab-dae57920e997   2Gi        RWO            Delete           Bound    demo/data-sample-mariadb-0   topolvm-provisioner            12m
pvc-b90179f8-c40a-4273-ad77-74ca8470b782   2Gi        RWO            Delete           Bound    demo/data-sample-mariadb-1   topolvm-provisioner            12m
pvc-f72411a4-80d5-4d32-b713-cb30ec662180   2Gi        RWO            Delete           Bound    demo/data-sample-mariadb-2   topolvm-provisioner            12m
```

The above output verifies that we have successfully expanded the volume of the MariaDB database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete mariadb -n demo sample-mariadb
$ kubectl delete mariadbopsrequest -n demo md-online-volume-expansion
```
