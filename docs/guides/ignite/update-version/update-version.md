---
title: Updating Ignite Cluster
menu:
  docs_{{ .version }}:
    identifier: ig-cluster-update-version
    name: Update Version
    parent: ig-update-version
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Update Version of Ignite Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to update the version of `Ignite` Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Ignite](/docs/guides/ignite/concepts/ignite/index.md)
    - [IgniteOpsRequest](/docs/guides/ignite/concepts/opsrequest/index.md)
    - [Updating Overview](/docs/guides/ignite/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
kubectl create ns demo
```
namespace/demo created

> **Note:** YAML files used in this tutorial are stored in [docs/examples/ignite](/docs/examples/ignite) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Prepare Ignite Cluster

Now, we are going to deploy an `Ignite` cluster with version `2.16.0`.

### Deploy Ignite

In this section, we are going to deploy an Ignite cluster. Then, in the next section we will update the version of the database using `IgniteOpsRequest` CRD. Below is the YAML of the `Ignite` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Ignite
metadata:
  name: ignite-quickstart
  namespace: demo
spec:
  version: "2.16.0"
  replicas: 3
  storage:
    resources:
      requests:
        storage: "1Gi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: "WipeOut"
```

Let's create the `Ignite` CR we have shown above,

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ignite/update-version/ignite.yaml
```
ignite.kubedb.com/ignite-quickstart created

Now, wait until `ignite-quickstart` has status `Ready`. i.e,

```bash
kubectl get ignite -n demo
```
NAME                VERSION    STATUS    AGE
ignite-quickstart   2.16.0     Ready     109s

We are now ready to apply the `IgniteOpsRequest` CR to update this database.

### Update Ignite Version

Here, we are going to update `Ignite` cluster from `2.16.0` to `2.17.0`.

#### Create IgniteOpsRequest:

In order to update the version of the cluster, we have to create an `IgniteOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `IgniteOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: upgrade-topology
  namespace: demo
spec:
  databaseRef:
    name: ignite-quickstart
  type: UpdateVersion
  updateVersion:
    targetVersion: 2.17.0
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `ignite-quickstart` Ignite database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version of the database `2.17.0`.
- Have a look [here](/docs/guides/ignite/concepts/opsrequest/index.md#spectimeout) on the respective sections to understand the `readinessCriteria`, `timeout` & `apply` fields.

Let's create the `IgniteOpsRequest` CR we have shown above,

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ignite/update-version/ig-version-upgrade-ops.yaml
```
igniteopsrequest.ops.kubedb.com/upgrade-topology created

#### Verify Ignite version updated successfully

If everything goes well, `KubeDB` Ops-manager operator will update the image of `Ignite` object and related `PetSets` and `Pods`.

Let's wait for `IgniteOpsRequest` to be `Successful`. Run the following command to watch `IgniteOpsRequest` CR,

```bash
kubectl get igniteopsrequest -n demo
```
Every 2.0s: kubectl get igniteopsrequest -n demo
NAME                TYPE            STATUS       AGE
upgrade-topology    UpdateVersion   Successful   84s

We can see from the above output that the `IgniteOpsRequest` has succeeded. If we describe the `IgniteOpsRequest` we will get an overview of the steps that were followed to update the database version.

```bash
kubectl describe igniteopsrequest -n demo upgrade-topology
```
Name:         upgrade-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         IgniteOpsRequest
Metadata:
  Creation Timestamp:  2024-10-23T10:46:27Z
  Generation:          1
  Resource Version:    1112190
  UID:                 6a1baef3-74cb-4a44-9b8f-f4fa49a4cfca
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   ignite-quickstart
  Timeout:  5m
  Type:     UpdateVersion
  Update Version:
    Target Version:  2.17.0
Status:
  Conditions:
    Last Transition Time:  2024-10-23T10:46:27Z
    Message:               Ignite ops-request has started to update version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2024-10-23T10:46:35Z
    Message:               successfully reconciled the Ignite with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-23T10:49:25Z
    Message:               Successfully Restarted Ignite nodes
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-10-23T10:46:40Z
    Message:               get pod; ConditionStatus:True; PodName:ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ignite-quickstart-0
    Last Transition Time:  2024-10-23T10:46:40Z
    Message:               evict pod; ConditionStatus:True; PodName:ignite-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ignite-quickstart-0
    Last Transition Time:  2024-10-23T10:46:45Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2024-10-23T10:47:25Z
    Message:               get pod; ConditionStatus:True; PodName:ignite-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ignite-quickstart-1
    Last Transition Time:  2024-10-23T10:47:25Z
    Message:               evict pod; ConditionStatus:True; PodName:ignite-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ignite-quickstart-1
    Last Transition Time:  2024-10-23T10:48:05Z
    Message:               get pod; ConditionStatus:True; PodName:ignite-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ignite-quickstart-2
    Last Transition Time:  2024-10-23T10:48:05Z
    Message:               evict pod; ConditionStatus:True; PodName:ignite-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ignite-quickstart-2
    Last Transition Time:  2024-10-23T10:49:25Z
    Message:               Successfully updated Ignite version
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                         Age    From                         Message
  ----     ------                                                         ----   ----                         -------
  Normal   Starting                                                       10m    KubeDB Ops-manager Operator  Start processing for IgniteOpsRequest: demo/upgrade-topology
  Normal   Starting                                                       10m    KubeDB Ops-manager Operator  Pausing Ignite database: demo/ignite-quickstart
  Normal   Successful                                                     10m    KubeDB Ops-manager Operator  Successfully paused Ignite database: demo/ignite-quickstart for IgniteOpsRequest: upgrade-topology
  Normal   UpdatePetSets                                                  10m    KubeDB Ops-manager Operator  successfully reconciled the Ignite with updated version
  Warning  get pod; ConditionStatus:True; PodName:ignite-quickstart-0    10m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ignite-quickstart-0
  Warning  evict pod; ConditionStatus:True; PodName:ignite-quickstart-0  10m    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ignite-quickstart-0
  Warning  running pod; ConditionStatus:False                            10m    KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:ignite-quickstart-1    9m25s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ignite-quickstart-1
  Warning  evict pod; ConditionStatus:True; PodName:ignite-quickstart-1  9m25s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ignite-quickstart-1
  Warning  get pod; ConditionStatus:True; PodName:ignite-quickstart-2    8m45s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ignite-quickstart-2
  Warning  evict pod; ConditionStatus:True; PodName:ignite-quickstart-2  8m45s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ignite-quickstart-2
  Normal   RestartPods                                                    7m25s  KubeDB Ops-manager Operator  Successfully Restarted Ignite nodes
  Normal   Starting                                                       7m25s  KubeDB Ops-manager Operator  Resuming Ignite database: demo/ignite-quickstart
  Normal   Successful                                                     7m25s  KubeDB Ops-manager Operator  Successfully updated Ignite version

Now, we are going to verify whether the `Ignite` and the related `PetSets` and their `Pods` have the new version image. Let's check,

```bash
kubectl get ignite -n demo ignite-quickstart -o=jsonpath='{.spec.version}{"\n"}'
```
2.17.0

```bash
kubectl get petset -n demo ignite-quickstart -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
```
ghcr.io/appscode-images/ignite:2.17.0

```bash
kubectl get pods -n demo ignite-quickstart-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
```
ghcr.io/appscode-images/ignite:2.17.0

You can see from above, our `Ignite` cluster has been updated with the new version. So, the updateVersion process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete ignite -n demo ignite-quickstart
kubectl delete igniteopsrequest -n demo upgrade-topology
```
