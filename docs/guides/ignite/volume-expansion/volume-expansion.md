---
title: Ignite Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: ig-volume-expansion-describe
    name: Expand Storage Volume
    parent: ig-volume-expansion
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Ignite Standalone Volume Expansion

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of a Ignite database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Ignite](/docs/guides/ignite/concepts/ignite.md)
  - [IgniteOpsRequest](/docs/guides/ignite/concepts/opsrequest.md)
  - [Volume Expansion Overview](/docs/guides/ignite/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: The yaml files used in this tutorial are stored in [docs/examples/Ignite](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/ignite) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Expand Volume of Ignite Database

Here, we are going to deploy a `Ignite` standalone using a supported version by `KubeDB` operator. Then we are going to apply `IgniteOpsRequest` to expand its volume.

### Prepare Ignite Standalone Database

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER            RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
standard (default)   kubernetes.io/gce-pd   Delete          Immediate           true                   2m49s
```

We can see from the output the `standard` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `Ignite` standalone database with version `2.17.0`.

#### Deploy Ignite standalone

In this section, we are going to deploy a Ignite standalone database with 1GB volume. Then, in the next section we will expand its volume to 2GB using `IgniteOpsRequest` CRD. Below is the YAML of the `Ignite` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Ignite
metadata:
  name: ig-standalone
  namespace: demo
spec:
  version: "2.17.0"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Let's create the `Ignite` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Ignite/volume-expansion/ig-standalone.yaml
Ignite.kubedb.com/ig-standalone created
```

Now, wait until `ig-standalone` has status `Ready`. i.e,

```bash
$ kubectl get ig -n demo
NAME            VERSION     STATUS    AGE
ig-standalone   2.17.0      Ready     2m53s
```

Let's check volume size from PetSet, and from the persistent volume,

```bash
$ kubectl get petset -n demo ig-standalone -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS   REASON   AGE
pvc-d0b07657-a012-4384-862a-b4e437774287   1Gi        RWO            Delete           Bound    demo/datadir-ig-standalone-0   standard                49s
```

You can see the PetSet has 1GB storage, and the capacity of the persistent volume is also 1GB.

We are now ready to apply the `IgniteOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the standalone database.

#### Create IgniteOpsRequest

In order to expand the volume of the database, we have to create a `IgniteOpsRequest` CR with our desired volume size. Below is the YAML of the `IgniteOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: igops-volume-exp-standalone
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: ig-standalone
  volumeExpansion:
    node: 2Gi
    mode: Online
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `ig-standalone` database.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.node` specifies the desired volume size.
- `spec.volumeExpansion.mode` specifies the desired volume expansion mode(`Online` or `Offline`).

During `Online` VolumeExpansion KubeDB expands volume without pausing database object, it directly updates the underlying PVC. And for `Offline` volume expansion, the database is paused. The Pods are deleted and PVC is updated. Then the database Pods are recreated with updated PVC.

Let's create the `IgniteOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ignite/volume-expansion/igops-volume-exp-standalone.yaml
igniteopsrequest.ops.kubedb.com/igops-volume-exp-standalone created
```

#### Verify Ignite Standalone volume expanded successfully

If everything goes well, `KubeDB` Ops-manager operator will update the volume size of `Ignite` object and related `StatefulSets` and `Persistent Volume`.

Let's wait for `IgniteOpsRequest` to be `Successful`. Run the following command to watch `IgniteOpsRequest` CR,

```bash
$ kubectl get igniteopsrequest -n demo
NAME                         TYPE              STATUS       AGE
igops-volume-exp-standalone   VolumeExpansion   Successful   75s
```

We can see from the above output that the `IgniteOpsRequest` has succeeded. If we describe the `IgniteOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe igniteopsrequest -n demo igops-volume-exp-standalone
  Name:         igops-volume-exp-standalone
  Namespace:    demo
  Labels:       <none>
  Annotations:  API Version:  ops.kubedb.com/v1alpha1
  Kind:         IgniteOpsRequest
  Metadata:
    Creation Timestamp:  2020-08-25T17:48:33Z
    Finalizers:
      kubedb.com
    Generation:        1
    Resource Version:  72899
    Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/igniteopsrequest/igops-volume-exp-standalone
    UID:               007fe35a-25f6-45e7-9e85-9add488b2622
  Spec:
    Database Ref:
      Name:  ig-standalone
    Type:    VolumeExpansion
    Volume Expansion:
      Node:  2Gi
  Status:
    Conditions:
      Last Transition Time:  2020-08-25T17:48:33Z
      Message:               Ignite ops request is being processed
      Observed Generation:   1
      Reason:                Scaling
      Status:                True
      Type:                  Scaling
      Last Transition Time:  2020-08-25T17:50:03Z
      Message:               Successfully updated Storage
      Observed Generation:   1
      Reason:                VolumeExpansion
      Status:                True
      Type:                  VolumeExpansion
      Last Transition Time:  2020-08-25T17:50:03Z
      Message:               Successfully Resumed Ignite: ig
      Observed Generation:   1
      Reason:                ResumeDatabase
      Status:                True
      Type:                  ResumeDatabase
      Last Transition Time:  2020-08-25T17:50:03Z
      Message:               Successfully completed the modification process
      Observed Generation:   1
      Reason:                Successful
      Status:                True
      Type:                  Successful
    Observed Generation:     1
    Phase:                   Successful
  Events:
    Type    Reason           Age   From                        Message
    ----    ------           ----  ----                        -------
    Normal  VolumeExpansion  29s   KubeDB Ops-manager operator  Successfully Updated Storage
    Normal  ResumeDatabase   29s   KubeDB Ops-manager operator  Resuming Ignite
    Normal  ResumeDatabase   29s   KubeDB Ops-manager operator  Successfully Resumed Ignite
    Normal  Successful       29s   KubeDB Ops-manager operator  Successfully Scaled Database
```

Now, we are going to verify from the `Statefulset`, and the `Persistent Volume` whether the volume of the database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get petset -n demo ig -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"2Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS   REASON   AGE
pvc-d0b07657-a012-4384-862a-b4e437774287   2Gi        RWO            Delete           Bound    demo/datadir-ig-0   standard                4m29s
```

The above output verifies that we have successfully expanded the volume of the Ignite database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete ig -n demo ig
kubectl delete igniteopsrequest -n demo igops-volume-exp-standalone
```
