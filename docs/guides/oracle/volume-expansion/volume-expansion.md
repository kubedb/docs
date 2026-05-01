---
title: Expand Oracle Volume
menu:
  docs_{{ .version }}:
    identifier: oracle-volume-expansion-cluster
    name: Cluster
    parent: oracle-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Expand Oracle Volume

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of a Oracle database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Oracle](/docs/guides/oracle/concepts/oracle.md)
  - [OracleOpsRequest](/docs/guides/oracle/concepts/opsrequest.md)
  - [Volume Expansion Overview](/docs/guides/oracle/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/oracle/volume-expansion/yamls](/docs/guides/oracle/volume-expansion/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Expand Volume of Oracle Database

Here, we are going to deploy a `Oracle` cluster using a supported version by `KubeDB` operator. Then we are going to apply `OracleOpsRequest` to expand its volume.

### Prepare Oracle Database

At first verify that your cluster has a storage class that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
standard (default)     rancher.io/local-path   Delete          WaitForFirstConsumer   false                  5m
standard-expandable    kubernetes.io/gce-pd    Delete          Immediate           true                   5m
```

We can see the `standard-expandable` storage class has `ALLOWVOLUMEEXPANSION` field as `true`. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `Oracle` cluster with version `1.17.0`.

#### Deploy Oracle

In this section, we are going to deploy a Oracle cluster with 1Gi volume. Then, in the next section we will expand its volume to 3Gi using `OracleOpsRequest` CRD. Below is the YAML of the `Oracle` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle-sample
  namespace: demo
spec:
  version: "21.3.0"
  replicas: 3
  storage:
    storageClassName: "standard-expandable"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Oracle` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/volume-expansion/yamls/oracle.yaml
oracle.kubedb.com/oracle-sample created
```

Now, wait until `oracle-sample` has status `Ready`:

```bash
$ kubectl get oracle -n demo
NAME             VERSION   STATUS   AGE
oracle-sample    1.17.0    Ready    3m47s
```

Let's check volume size from the PetSet and from the persistent volumes:

```bash
$ kubectl get petset -n demo oracle-sample -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pvc -n demo
NAME                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS           AGE
oracle-sample-oracle-sample-0 Bound    pvc-a1b2c3d4-e5f6-7890-abcd-ef1234567890   1Gi        RWO            standard-expandable    4m
oracle-sample-oracle-sample-1 Bound    pvc-b2c3d4e5-f6a7-8901-bcde-f01234567891   1Gi        RWO            standard-expandable    3m
oracle-sample-oracle-sample-2 Bound    pvc-c3d4e5f6-a7b8-9012-cdef-012345678902   1Gi        RWO            standard-expandable    2m
```

You can see the PetSet has 1Gi storage, and the capacity of the persistent volumes is also 1Gi.

We are now ready to apply the `OracleOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the Oracle cluster.

#### Create OracleOpsRequest

In order to expand the volume of the database, we have to create a `OracleOpsRequest` CR with our desired volume size. Below is the YAML of the `OracleOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: qdops-vol-exp
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: oracle-sample
  type: VolumeExpansion
  volumeExpansion:
    mode: Online
    node: 3Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `oracle-sample` Oracle database.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.node` specifies the desired volume size.
- `spec.volumeExpansion.mode` specifies the desired volume expansion mode (`Online` or `Offline`).

> Note: If the StorageClass doesn't support `Online` volume expansion, try offline volume expansion by using `spec.volumeExpansion.mode: "Offline"`.

During `Online` VolumeExpansion KubeDB expands volume without pausing the database object; it directly updates the underlying PVC. For `Offline` volume expansion, the database is paused, the Pods are deleted, the PVC is updated, and then the database Pods are recreated with the updated PVC.

Let's create the `OracleOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/volume-expansion/yamls/qdops-vol-exp.yaml
oracleopsrequest.ops.kubedb.com/qdops-vol-exp created
```

#### Verify Oracle volume expanded successfully

If everything goes well, `KubeDB` Ops-manager operator will update the volume size of the `Oracle` object and related `PetSet` and `Persistent Volumes`.

Let's wait for `OracleOpsRequest` to be `Successful`. Run the following command to watch `OracleOpsRequest` CR,

```bash
$ kubectl get oracleopsrequest -n demo
NAME             TYPE              STATUS       AGE
qdops-vol-exp    VolumeExpansion   Successful   10m
```

We can see from the above output that the `OracleOpsRequest` has succeeded. If we describe the `OracleOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe oracleopsrequest qdops-vol-exp -n demo
Name:         qdops-vol-exp
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         OracleOpsRequest
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  oracle-sample
  Type:    VolumeExpansion
  Volume Expansion:
    Mode:  Online
    Node:  3Gi
Status:
  Conditions:
    Last Transition Time:  2026-05-01T10:04:19Z
    Message:               Oracle ops request is expanding volume of database
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2026-05-01T10:05:12Z
    Message:               Online Volume Expansion performed successfully in Oracle pods for OracleOpsRequest: demo/qdops-vol-exp
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2026-05-01T10:06:08Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2026-05-01T10:06:52Z
    Message:               Successfully Expanded Volume.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason             Age   From                              Message
  ----    ------             ----  ----                              -------
  Normal  PauseDatabase      12m   KubeDB Ops-manager Operator  Pausing Oracle demo/oracle-sample
  Normal  PauseDatabase      12m   KubeDB Ops-manager Operator  Successfully paused Oracle demo/oracle-sample
  Normal  VolumeExpansion    11m   KubeDB Ops-manager Operator  Online Volume Expansion performed successfully in Oracle pods for OracleOpsRequest: demo/qdops-vol-exp
  Normal  ResumeDatabase     11m   KubeDB Ops-manager Operator  Resuming Oracle demo/oracle-sample
  Normal  ResumeDatabase     11m   KubeDB Ops-manager Operator  Successfully resumed Oracle demo/oracle-sample
  Normal  ReadyPetSets       10m   KubeDB Ops-manager Operator  PetSet is recreated
  Normal  Successful         10m   KubeDB Ops-manager Operator  Successfully Expanded Volume
```

Now, we are going to verify from the `PetSet` and `Persistent Volumes` whether the volume of the Oracle database has expanded to meet the desired state:

```bash
$ kubectl get petset -n demo oracle-sample -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"3Gi"

$ kubectl get pvc -n demo
NAME                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS           AGE
oracle-sample-oracle-sample-0 Bound    pvc-a1b2c3d4-e5f6-7890-abcd-ef1234567890   3Gi        RWO            standard-expandable    14m
oracle-sample-oracle-sample-1 Bound    pvc-b2c3d4e5-f6a7-8901-bcde-f01234567891   3Gi        RWO            standard-expandable    13m
oracle-sample-oracle-sample-2 Bound    pvc-c3d4e5f6-a7b8-9012-cdef-012345678902   3Gi        RWO            standard-expandable    12m
```

The above output verifies that we have successfully expanded the volume of the Oracle database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete oracle -n demo oracle-sample
oracle.kubedb.com "oracle-sample" deleted

$ kubectl delete oracleopsrequest -n demo qdops-vol-exp
oracleopsrequest.ops.kubedb.com "qdops-vol-exp" deleted
```