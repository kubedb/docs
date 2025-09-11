---
title: Hazelcast Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: hz-volume-expansion-describe
    name: Expand Storage Volume
    parent: hz-volume-expansion
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Hazelcast Volume Expansion

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of a Hazelcast database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Hazelcast](/docs/guides/hazelcast/concepts/hazelcast.md)
  - [HazelcastOpsRequest](/docs/guides/hazelcast/concepts/hazelcast-opsrequest.md)
  - [Volume Expansion Overview](/docs/guides/hazelcast/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/hazelcast](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hazelcast) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Expand Volume of Hazelcast Database

Here, we are going to expand the volume of a Hazelcast database.

### Deploy Hazelcast Database

In this section, we are going to deploy a Hazelcast database with 1Gi storage and then we will expand its storage. Below is the YAML of the `Hazelcast` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Hazelcast
metadata:
  name: hz-prod
  namespace: demo
spec:
  deletionPolicy: WipeOut
  licenseSecret:
    name: hz-license-key
  replicas: 2
  version: 5.5.2
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi

```

Let's create the `Hazelcast` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/volume-expansion/hazelcast.yaml
hazelcast.kubedb.com/hz-prod created
```

Now, wait until `hz-prod` has status `Ready`. i.e,

```bash
$ kubectl get hz -n demo
NAME             TYPE            VERSION   STATUS   AGE
hz-prod    kubedb.com/v1alpha2   5.2.2     Ready    3m
```

Let's check volume size from statefulset, and from the persistent volume,

```bash
$ kubectl get statefulset -n demo hz-prod -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -o json | jq '.items[].spec.capacity.storage'
"1Gi"
"1Gi"
```

You can see the statefulset has 1Gi storage, and the capacity of all the persistent volumes are 1Gi.

We are now ready to apply the `HazelcastOpsRequest` CR to expand the volume of this database.

### Expanding Storage Size

Here, we are going to expand the volume of the database.

#### Create HazelcastOpsRequest

In order to expand the volume of the database, we have to create a `HazelcastOpsRequest` CR with our desired size. Below is the YAML of the `HazelcastOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hz-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: hz-prod
  volumeExpansion:
    hazelcast: 2Gi
    mode: Online
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `hz-prod` database.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.hazelcast` specifies the desired size of the volume after expansion.

Let's create the `HazelcastOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/volume-expansion/ops.yaml
hazelcastopsrequest.ops.kubedb.com/hz-volume-expansion created
```

#### Verify Hazelcast volume expanded successfully

If everything goes well, `KubeDB` Ops-manager operator will expand the volume of `Hazelcast` object and related `statefulsets` and `Pods`.

Let's wait for `HazelcastOpsRequest` to be `Successful`. Run the following command to watch `HazelcastOpsRequest` CR,

```bash
$ kubectl get hazelcastopsrequest -n demo
NAME                  TYPE              STATUS       AGE
hz-volume-expansion   VolumeExpansion   Successful   3m2s
```

We can see from the above output that the `HazelcastOpsRequest` has succeeded. If we describe the `HazelcastOpsRequest` we will get an overview of the steps that were followed to expand the database volume.

```bash
$ kubectl describe hazelcastopsrequest -n demo hz-volume-expansion
Name:         hz-volume-expansion
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         HazelcastOpsRequest
Metadata:
  Creation Timestamp:  2025-08-19T13:11:56Z
  Generation:          1
  Resource Version:    5498537
  UID:                 1141abd1-aa25-4beb-837b-69a6d577920f
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  hz-prod
  Type:    VolumeExpansion
  Volume Expansion:
    Hazelcast:  2Gi
    Mode:       Online
Status:
  Conditions:
    Last Transition Time:  2025-08-19T13:11:56Z
    Message:               Hazelcast ops-request has started to expand volume of hazelcast nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2025-08-19T13:12:29Z
    Message:               successfully deleted the statefulSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanStatefulSetPods
    Status:                True
    Type:                  OrphanStatefulSetPods
    Last Transition Time:  2025-08-19T13:12:09Z
    Message:               get statefulset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetStatefulset
    Last Transition Time:  2025-08-19T13:12:09Z
    Message:               delete statefulset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeleteStatefulset
    Last Transition Time:  2025-08-19T13:15:39Z
    Message:               successfully updated PVC sizes
    Observed Generation:   1
    Reason:                VolumeExpansionSucceeded
    Status:                True
    Type:                  VolumeExpansionSucceeded
    Last Transition Time:  2025-08-19T13:12:39Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2025-08-19T13:12:39Z
    Message:               patch pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPvc
    Last Transition Time:  2025-08-19T13:15:29Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2025-08-19T13:15:49Z
    Message:               successfully reconciled the Hazelcast resources
    Observed Generation:   1
    Reason:                UpdateStatefulSets
    Status:                True
    Type:                  UpdateStatefulSets
    Last Transition Time:  2025-08-19T13:16:09Z
    Message:               StatefulSet is recreated
    Observed Generation:   1
    Reason:                ReadyStatefulSets
    Status:                True
    Type:                  ReadyStatefulSets
    Last Transition Time:  2025-08-19T13:15:59Z
    Message:               get stateful set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetStatefulSet
    Last Transition Time:  2025-08-19T13:15:59Z
    Message:               Successfully completed volumeExpansion for Hazelcast
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:                    <none>
```

Now, we are going to verify from the `statefulset`, and the `PersistentVolume` whether the volume of the database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get statefulset -n demo hz-prod -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"2Gi"

$ kubectl get pv -o json | jq '.items[].spec.capacity.storage'
"2Gi"
"2Gi"
```

The above output verifies that we have successfully expanded the volume of the Hazelcast database.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete hazelcastopsrequest -n demo hz-volume-expansion
kubectl delete hazelcast -n demo hz-prod
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Hazelcast object](/docs/guides/hazelcast/concepts/hazelcast.md).
- Monitor your Hazelcast database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/hazelcast/monitoring/prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
