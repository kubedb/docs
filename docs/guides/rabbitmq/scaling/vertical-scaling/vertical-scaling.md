---
title: Vertical Scaling RabbitMQ
menu:
  docs_{{ .version }}:
    identifier: rm-vertical-scaling-ops
    name: rabbitmq-vertical-scaling
    parent: rm-vertical-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale RabbitMQ Standalone

This guide will show you how to use `KubeDB` Ops-manager operator to update the resources of a RabbitMQ standalone database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [RabbitMQ](/docs/guides/rabbitmq/concepts/rabbitmq.md)
  - [RabbitMQOpsRequest](/docs/guides/rabbitmq/concepts/opsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/rabbitmq/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/rabbitmq](/docs/examples/rabbitmq) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on Standalone

Here, we are going to deploy a  `RabbitMQ` standalone using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare RabbitMQ Standalone Database

Now, we are going to deploy a `RabbitMQ` standalone database with version `3.13.2`.

### Deploy RabbitMQ standalone 

In this section, we are going to deploy a RabbitMQ standalone database. Then, in the next section we will update the resources of the database using `RabbitMQOpsRequest` CRD. Below is the YAML of the `RabbitMQ` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rm-standalone
  namespace: demo
spec:
  version: "3.13.2"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Let's create the `RabbitMQ` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/scaling/mg-standalone.yaml
rabbitmq.kubedb.com/rm-standalone created
```

Now, wait until `mg-standalone` has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo
NAME            VERSION    STATUS    AGE
rm-standalone   3.13.2      Ready     5m56s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo rm-standalone-0 -o json | jq '.spec.containers[].resources'
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

You can see the Pod has default resources which is assigned by the Kubedb operator.

We are now ready to apply the `RabbitMQOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of the standalone database to meet the desired resources after scaling.

#### Create RabbitMQOpsRequest

In order to update the resources of the database, we have to create a `RabbitMQOpsRequest` CR with our desired resources. Below is the YAML of the `RabbitMQOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RabbitMQOpsRequest
metadata:
  name: rmops-vscale-standalone
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: rm-standalone
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
- Have a look [here](/docs/guides/rabbitmq/concepts/opsrequest.md#spectimeout) on the respective sections to understand the `timeout` & `apply` fields.

Let's create the `RabbitMQOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/scaling/vertical-scaling/rmops-vscale-standalone.yaml
rabbitmqopsrequest.ops.kubedb.com/rmops-vscale-standalone created
```

#### Verify RabbitMQ Standalone resources updated successfully 

If everything goes well, `KubeDB` Ops-manager operator will update the resources of `RabbitMQ` object and related `StatefulSets` and `Pods`.

Let's wait for `RabbitMQOpsRequest` to be `Successful`.  Run the following command to watch `RabbitMQOpsRequest` CR,

```bash
$ kubectl get RabbitMQopsrequest -n demo
Every 2.0s: kubectl get RabbitMQopsrequest -n demo
NAME                     TYPE              STATUS       AGE
mops-vscale-standalone   VerticalScaling   Successful   108s
```

We can see from the above output that the `RabbitMQOpsRequest` has succeeded. If we describe the `RabbitMQOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe rabbitmqopsrequest -n demo rmops-vscale-standalone
Name:         rmops-vscale-standalone
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RabbitMQOpsRequest
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
    Message:               RabbitMQ ops request is vertically scaling database
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
  Normal  PauseDatabase              34s   KubeDB Ops-manager Operator  Pausing RabbitMQ demo/mg-standalone
  Normal  PauseDatabase              34s   KubeDB Ops-manager Operator  Successfully paused RabbitMQ demo/mg-standalone
  Normal  Starting                   34s   KubeDB Ops-manager Operator  Updating Resources of StatefulSet: mg-standalone
  Normal  UpdateStandaloneResources  34s   KubeDB Ops-manager Operator  Successfully updated standalone Resources
  Normal  Starting                   34s   KubeDB Ops-manager Operator  Updating Resources of StatefulSet: mg-standalone
  Normal  UpdateStandaloneResources  34s   KubeDB Ops-manager Operator  Successfully updated standalone Resources
  Normal  UpdateStandaloneResources  4s    KubeDB Ops-manager Operator  Successfully Vertically Scaled Standalone Resources
  Normal  UpdateStandaloneResources  4s    KubeDB Ops-manager Operator  Successfully Vertically Scaled Standalone Resources
  Normal  ResumeDatabase             4s    KubeDB Ops-manager Operator  Resuming RabbitMQ demo/mg-standalone
  Normal  ResumeDatabase             3s    KubeDB Ops-manager Operator  Successfully resumed RabbitMQ demo/mg-standalone
  Normal  Successful                 3s    KubeDB Ops-manager Operator  Successfully Vertically Scaled Database

```

Now, we are going to verify from the Pod yaml whether the resources of the standalone database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo rm-standalone-0 -o json | jq '.spec.containers[].resources'
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

The above output verifies that we have successfully scaled up the resources of the RabbitMQ standalone database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete rm -n demo rm-standalone
kubectl delete rabbitmqopsrequest -n demo rmops-vscale-standalone
```