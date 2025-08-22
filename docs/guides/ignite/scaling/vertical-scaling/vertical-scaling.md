---
title: Vertical Scaling Ignite
menu:
  docs_{{ .version }}:
    identifier: ig-vertical-scaling-ops
    name: Scale Vertically
    parent: ig-vertical-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale Ignite Standalone

This guide will show you how to use `KubeDB` Ops-manager operator to update the resources of a Ignite standalone database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Ignite](/docs/guides/ignite/concepts/ignite.md)
  - [IgniteOpsRequest](/docs/guides/ignite/concepts/opsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/ignite/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/ignite](/docs/examples/ignite) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on Standalone

Here, we are going to deploy a  `Ignite` standalone using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare Ignite Standalone Database

Now, we are going to deploy a `Ignite` standalone database with version `2.17.0`.

### Deploy Ignite standalone 

In this section, we are going to deploy a Ignite standalone database. Then, in the next section we will update the resources of the database using `IgniteOpsRequest` CRD. Below is the YAML of the `Ignite` CR that we are going to create,

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ignite/scaling/mg-standalone.yaml
ignite.kubedb.com/ig-standalone created
```

Now, wait until `ig-standalone` has status `Ready`. i.e,

```bash
$ kubectl get ig -n demo
NAME            VERSION    STATUS    AGE
ig-standalone   2.17.0      Ready     5m56s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo ig-standalone-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "500m",
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

You can see the Pod has default resources which is assigned by the KubeDB operator.

We are now ready to apply the `IgniteOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of the standalone database to meet the desired resources after scaling.

#### Create IgniteOpsRequest

In order to update the resources of the database, we have to create a `IgniteOpsRequest` CR with our desired resources. Below is the YAML of the `IgniteOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: igops-vscale-standalone
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: ig-standalone
  verticalScaling:
    node:
      resources:
        requests:
          memory: "2Gi"
          cpu: "1"
        limits:
          memory: "2Gi"
          cpu: "1"
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `mops-vscale-standalone` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.standalone` specifies the desired resources after scaling.
- Have a look [here](/docs/guides/ignite/concepts/opsrequest.md#spectimeout) on the respective sections to understand the `timeout` & `apply` fields.

Let's create the `IgniteOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ignite/scaling/vertical-scaling/igops-vscale-standalone.yaml
igniteopsrequest.ops.kubedb.com/igops-vscale-standalone created
```

#### Verify Ignite Standalone resources updated successfully 

If everything goes well, `KubeDB` Ops-manager operator will update the resources of `Ignite` object and related `StatefulSets` and `Pods`.

Let's wait for `IgniteOpsRequest` to be `Successful`.  Run the following command to watch `IgniteOpsRequest` CR,

```bash
$ kubectl get igniteopsrequest -n demo
Every 2.0s: kubectl get igniteopsrequest -n demo
NAME                      TYPE              STATUS       AGE
igops-vscale-standalone   VerticalScaling   Successful   108s
```

We can see from the above output that the `IgniteOpsRequest` has succeeded. If we describe the `IgniteOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe igniteopsrequest -n demo igops-vscale-standalone
Name:         igops-vscale-standalone
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         IgniteOpsRequest
Metadata:
  Creation Timestamp:  2022-10-26T10:54:01Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:apply:
        f:databaseRef:
        f:readinessCriteria:
          .:
          f:objectsCountDiffPercentage:
          f:oplogMaxLagSeconds:
        f:timeout:
        f:type:
        f:verticalScaling:
          .:
          f:standalone:
            .:
            f:limits:
              .:
              f:cpu:
              f:memory:
            f:requests:
              .:
              f:cpu:
              f:memory:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2022-10-26T10:54:01Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-ops-manager
    Operation:       Update
    Subresource:     status
    Time:            2022-10-26T10:54:52Z
  Resource Version:  613933
  UID:               c3bf9c3d-cf96-49ae-877f-a895e0b1d280
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  mg-standalone
  Readiness Criteria:
    Objects Count Diff Percentage:  10
    Oplog Max Lag Seconds:          20
  Timeout:                          5m
  Type:                             VerticalScaling
  Vertical Scaling:
    Standalone:
      Limits:
        Cpu:     1
        Memory:  2Gi
      Requests:
        Cpu:     1
        Memory:  2Gi
Status:
  Conditions:
    Last Transition Time:  2022-10-26T10:54:21Z
    Message:               Ignite ops request is vertically scaling database
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2022-10-26T10:54:51Z
    Message:               Successfully Vertically Scaled Standalone Resources
    Observed Generation:   1
    Reason:                UpdateStandaloneResources
    Status:                True
    Type:                  UpdateStandaloneResources
    Last Transition Time:  2022-10-26T10:54:52Z
    Message:               Successfully Vertically Scaled Database
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                     Age   From                         Message
  ----    ------                     ----  ----                         -------
  Normal  PauseDatabase              34s   KubeDB Ops-manager Operator  Pausing Ignite demo/mg-standalone
  Normal  PauseDatabase              34s   KubeDB Ops-manager Operator  Successfully paused Ignite demo/mg-standalone
  Normal  Starting                   34s   KubeDB Ops-manager Operator  Updating Resources of StatefulSet: mg-standalone
  Normal  UpdateStandaloneResources  34s   KubeDB Ops-manager Operator  Successfully updated standalone Resources
  Normal  Starting                   34s   KubeDB Ops-manager Operator  Updating Resources of StatefulSet: mg-standalone
  Normal  UpdateStandaloneResources  34s   KubeDB Ops-manager Operator  Successfully updated standalone Resources
  Normal  UpdateStandaloneResources  4s    KubeDB Ops-manager Operator  Successfully Vertically Scaled Standalone Resources
  Normal  UpdateStandaloneResources  4s    KubeDB Ops-manager Operator  Successfully Vertically Scaled Standalone Resources
  Normal  ResumeDatabase             4s    KubeDB Ops-manager Operator  Resuming Ignite demo/mg-standalone
  Normal  ResumeDatabase             3s    KubeDB Ops-manager Operator  Successfully resumed Ignite demo/mg-standalone
  Normal  Successful                 3s    KubeDB Ops-manager Operator  Successfully Vertically Scaled Database

```

Now, we are going to verify from the Pod yaml whether the resources of the standalone database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo ig-standalone-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "1",
    "memory": "2Gi"
  }
}
```

The above output verifies that we have successfully scaled up the resources of the Ignite standalone database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete ig -n demo ig-standalone
kubectl delete igniteopsrequest -n demo igops-vscale-standalone
```