---
title: Updating RabbitMQ Cluster
menu:
  docs_{{ .version }}:
    identifier: rm-cluster-update-version
    name: Update Version
    parent: rm-update-version
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# update version of RabbitMQ Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to update the version of `RabbitMQ` Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [RabbitMQ](/docs/guides/rabbitmq/concepts/rabbitmq.md)
    - [RabbitMQOpsRequest](/docs/guides/rabbitmq/concepts/opsrequest.md)
    - [Updating Overview](/docs/guides/rabbitmq/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/rabbitmq](/docs/examples/rabbitmq) directory of [kubedb/docs](https://github.com/kube/docs) repository.

## Prepare RabbitMQ cluster

Now, we are going to deploy a `RabbitMQ` cluster with version `3.12.12`.

### Deploy RabbitMQ

In this section, we are going to deploy a RabbitMQ cluster. Then, in the next section we will update the version of the database using `RabbitMQOpsRequest` CRD. Below is the YAML of the `RabbitMQ` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rm-cluster
  namespace: demo
spec:
  version: "3.12.12"
  replicas: 3
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/RabbitMQ/update-version/rm-cluster.yaml
rabbitmq.kubedb.com/rm-cluster created
```

Now, wait until `rm-cluster` created has status `Ready`. i.e,

```bash
$ kubectl get rm -n demo                                                                                                                                             
NAME            VERSION    STATUS    AGE
rm-cluster      3.12.12   Ready     109s
```

We are now ready to apply the `RabbitMQOpsRequest` CR to update this database.

### update RabbitMQ Version

Here, we are going to update `RabbitMQ` cluster from `3.12.12` to `3.13.2`.

#### Create RabbitMQOpsRequest:

In order to update the version of the cluster, we have to create a `RabbitMQOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `RabbitMQOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RabbitMQOpsRequest
metadata:
  name: rm-cluster-update
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: rm-cluster
  updateVersion:
    targetVersion: 3.13.2
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `rm-cluster` RabbitMQ database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version of the database `3.13.2`.
- Have a look [here](/docs/guides/rabbitmq/concepts/opsrequest.md#spectimeout) on the respective sections to understand the `readinessCriteria`, `timeout` & `apply` fields.

Let's create the `RabbitMQOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/update-version/rmops-cluster-update .yaml
rabbitmqopsrequest.ops.kubedb.com/rmops-cluster-update created
```

#### Verify RabbitMQ version updated successfully

If everything goes well, `KubeDB` Ops-manager operator will update the image of `RabbitMQ` object and related `PetSets` and `Pods`.

Let's wait for `RabbitMQOpsRequest` to be `Successful`.  Run the following command to watch `RabbitMQOpsRequest` CR,

```bash
$ kubectl get rabbitmqopsrequest -n demo
Every 2.0s: kubectl get rabbitmqopsrequest -n demo
NAME                      TYPE            STATUS       AGE
rmops-cluster-update      UpdateVersion   Successful   84s
```

We can see from the above output that the `RabbitMQOpsRequest` has succeeded. If we describe the `RabbitMQOpsRequest` we will get an overview of the steps that were followed to update the database version.

```bash
$ kubectl describe rabbitmqopsrequest -n demo rmops-cluster-update
Name:         rmops-cluster-update
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RabbitMQOpsRequest
Metadata:
  Creation Timestamp:  2022-10-26T10:19:55Z
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
        f:updateVersion:
          .:
          f:targetVersion:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2022-10-26T10:19:55Z
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
    Time:            2022-10-26T10:23:09Z
  Resource Version:  607814
  UID:               38053605-47bd-4d94-9f53-ce9474ad0a98
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  rm-cluster
  Readiness Criteria:
    Objects Count Diff Percentage:  10
    Oplog Max Lag Seconds:          20
  Timeout:                          5m
  Type:                             UpdateVersion
  UpdateVersion:
    Target Version:  4.4.26
Status:
  Conditions:
    Last Transition Time:  2022-10-26T10:21:20Z
    Message:               RabbitMQ ops request is update-version database version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2022-10-26T10:21:39Z
    Message:               Successfully updated statefulsets update strategy type
    Observed Generation:   1
    Reason:                UpdateStatefulSets
    Status:                True
    Type:                  UpdateStatefulSets
    Last Transition Time:  2022-10-26T10:23:09Z
    Message:               Successfully Updated Standalone Image
    Observed Generation:   1
    Reason:                UpdateStandaloneImage
    Status:                True
    Type:                  UpdateStandaloneImage
    Last Transition Time:  2022-10-26T10:23:09Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                 Age    From                         Message
  ----    ------                 ----   ----                         -------
  Normal  PauseDatabase          2m27s  KubeDB Ops-manager Operator  Pausing RabbitMQ demo/rm-cluster
  Normal  PauseDatabase          2m27s  KubeDB Ops-manager Operator  Successfully paused RabbitMQ demo/rm-cluster
  Normal  Updating               2m27s  KubeDB Ops-manager Operator  Updating StatefulSets
  Normal  Updating               2m8s   KubeDB Ops-manager Operator  Successfully Updated StatefulSets
  Normal  UpdateStandaloneImage  38s    KubeDB Ops-manager Operator  Successfully Updated Standalone Image
  Normal  ResumeDatabase         38s    KubeDB Ops-manager Operator  Resuming RabbitMQ demo/rm-cluster
  Normal  ResumeDatabase         38s    KubeDB Ops-manager Operator  Successfully resumed RabbitMQ demo/rm-cluster
  Normal  Successful             38s    KubeDB Ops-manager Operator  Successfully Updated Database
```

Now, we are going to verify whether the `RabbitMQ` and the related `PetSets` and their `Pods` have the new version image. Let's check,

```bash
$ kubectl get rm -n demo rm-cluster -o=jsonpath='{.spec.version}{"\n"}'
3.13.2

$ kubectl get petset -n demo rm-cluster -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/rabbitmq:3.13.2-management-alpine

$ kubectl get pods -n demo rm-cluster-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/rabbitmq:3.13.2-management-alpine
```

You can see from above, our `RabbitMQ` cluster has been updated with the new version. So, the updateVersion process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete rm -n demo rm-cluster
kubectl delete rabbitmqopsrequest -n demo rmops-update-cluster
```