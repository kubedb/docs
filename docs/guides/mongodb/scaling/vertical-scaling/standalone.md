---
title: Vertical Scaling Standalone MongoDB
menu:
  docs_{{ .version }}:
    identifier: mg-vertical-scaling-standalone
    name: Standalone
    parent: mg-vertical-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Vertical Scale MongoDB Standalone

This guide will show you how to use `KubeDB` Enterprise operator to update the resources of a MongoDB standalone database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here]().

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/concepts/databases/mongodb.md)
  - [MongoDBOpsRequest](/docs/concepts/day-2-operations/mongodbopsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/mongodb/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on Standalone

Here, we are going to deploy a  `MongoDB` standalone using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare MongoDB Standalone Database

Now, we are going to deploy a `MongoDB` standalone database with version `3.6.8`.

### Deploy MongoDB standalone 

In this section, we are going to deploy a MongoDB standalone database. Then, in the next section we will update the resources of the database using `MongoDBOpsRequest` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-standalone
  namespace: demo
spec:
  version: "3.6.8-v1"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Let's create the `MongoDB` CR we have shown above,

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/scaling/mg-standalone.yaml
mongodb.kubedb.com/mg-standalone created
```

Now, wait until `mg-standalone` has status `Running`. i.e,

```console
$ kubectl get mg -n demo                                                                                                                                             20:05:47
  NAME            VERSION    STATUS    AGE
  mg-standalone   3.6.8-v1   Running   5m56s
```

Let's check the Pod containers resources,

```console
$ kubectl get pod -n demo mg-standalone-0 -o json | jq '.spec.containers[].resources'
{}
```

You can see the Pod has empty resources that means the scheduler will choose a random node to place the container of the Pod on by default.

We are now ready to apply the `MongoDBOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of the standalone database to meet the desired resources after scaling.

#### Create MongoDBOpsRequest

In order to update the resources of the database, we have to create a `MongoDBOpsRequest` CR with our desired resources. Below is the YAML of the `MongoDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-vscale-standalone
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: mg-standalone
  verticalScaling:
    standalone:
      requests:
        memory: "150Mi"
        cpu: "0.1"
      limits:
        memory: "250Mi"
        cpu: "0.2"
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `mops-vscale-standalone` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.standalone` specifies the desired resources after scaling.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/scaling/vertical-scaling/mops-vscale-standalone.yaml
mongodbopsrequest.ops.kubedb.com/mops-vscale-standalone created
```

#### Verify MongoDB Standalone resources updated successfully 

If everything goes well, `KubeDB` Enterprise operator will update the resources of `MongoDB` object and related `StatefulSets` and `Pods`.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```console
$ kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                     TYPE              STATUS       AGE
mops-vscale-standalone   VerticalScaling   Successful   108s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to scale the database.

```console
$ kubectl describe mongodbopsrequest -n demo mops-vscale-standalone
  Name:         mops-vscale-standalone
  Namespace:    demo
  Labels:       <none>
  Annotations:  API Version:  ops.kubedb.com/v1alpha1
  Kind:         MongoDBOpsRequest
  Metadata:
    Creation Timestamp:  2020-08-25T05:12:17Z
    Finalizers:
      kubedb.com
    Generation:  1
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
          f:databaseRef:
            .:
            f:name:
          f:type:
          f:verticalScaling:
            .:
            f:standalone:
              .:
              f:limits:
                .:
                f:memory:
              f:requests:
                .:
                f:memory:
      Manager:      kubectl
      Operation:    Update
      Time:         2020-08-25T05:12:17Z
      API Version:  ops.kubedb.com/v1alpha1
      Fields Type:  FieldsV1
      fieldsV1:
        f:metadata:
          f:finalizers:
        f:spec:
          f:verticalScaling:
            f:standalone:
              f:limits:
                f:cpu:
              f:requests:
                f:cpu:
        f:status:
          .:
          f:conditions:
          f:observedGeneration:
          f:phase:
      Manager:         kubedb-enterprise
      Operation:       Update
      Time:            2020-08-25T05:12:57Z
    Resource Version:  4991494
    Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-vscale-standalone
    UID:               99935637-43f4-45a2-8538-9b3a63fb3ca1
  Spec:
    Database Ref:
      Name:  mg-standalone
    Type:    VerticalScaling
    Vertical Scaling:
      Standalone:
        Limits:
          Cpu:     0.2
          Memory:  250Mi
        Requests:
          Cpu:     0.1
          Memory:  150Mi
  Status:
    Conditions:
      Last Transition Time:  2020-08-25T05:12:17Z
      Message:               MongoDB ops request is being processed
      Observed Generation:   1
      Reason:                Scaling
      Status:                True
      Type:                  Scaling
      Last Transition Time:  2020-08-25T05:12:17Z
      Message:               Successfully paused mongodb: mg-standalone
      Observed Generation:   1
      Reason:                PauseDatabase
      Status:                True
      Type:                  PauseDatabase
      Last Transition Time:  2020-08-25T05:12:17Z
      Message:               Successfully updated StatefulSets Resources
      Observed Generation:   1
      Reason:                UpdateStatefulSetResources
      Status:                True
      Type:                  UpdateStatefulSetResources
      Last Transition Time:  2020-08-25T05:12:57Z
      Message:               Successfully updated Standalone resources
      Observed Generation:   1
      Reason:                UpdateStandaloneResources
      Status:                True
      Type:                  UpdateStandaloneResources
      Last Transition Time:  2020-08-25T05:12:57Z
      Message:               Successfully Resumed mongodb: mg-standalone
      Observed Generation:   1
      Reason:                ResumeDatabase
      Status:                True
      Type:                  ResumeDatabase
      Last Transition Time:  2020-08-25T05:12:57Z
      Message:               Successfully completed the modification process
      Observed Generation:   1
      Reason:                Successful
      Status:                True
      Type:                  Successful
    Observed Generation:     1
    Phase:                   Successful
  Events:
    Type    Reason                      Age    From                        Message
    ----    ------                      ----   ----                        -------
    Normal  PauseDatabase               6m12s  KubeDB Enterprise Operator  Pausing Mongodb mg-standalone in Namespace demo
    Normal  PauseDatabase               6m12s  KubeDB Enterprise Operator  Successfully Paused Mongodb mg-standalone in Namespace demo
    Normal  Starting                    6m12s  KubeDB Enterprise Operator  Updating Resources of StatefulSet: mg-standalone
    Normal  UpdateStatefulSetResources  6m12s  KubeDB Enterprise Operator  Successfully updated StatefulSets Resources
    Normal  UpdateStandaloneResources   6m12s  KubeDB Enterprise Operator  Updating Standalone Resources
    Normal  UpdateStandaloneResources   5m32s  KubeDB Enterprise Operator  Successfully Updated Resources of Pod (master): mg-standalone-0
    Normal  UpdateStandaloneResources   5m32s  KubeDB Enterprise Operator  Successfully Updated Standalone Resources
    Normal  ResumeDatabase              5m32s  KubeDB Enterprise Operator  Resuming MongoDB
    Normal  ResumeDatabase              5m32s  KubeDB Enterprise Operator  Successfully Resumed mongodb
    Normal  Successful                  5m32s  KubeDB Enterprise Operator  Successfully Scaled Database
```

Now, we are going to verify from the Pod yaml whether the resources of the standalone database has updated to meet up the desired state, Let's check,

```console
$ kubectl get pod -n demo mg-standalone-0 -o json | jq '.spec.containers[].resources'
  {
    "limits": {
      "cpu": "200m",
      "memory": "250Mi"
    },
    "requests": {
      "cpu": "100m",
      "memory": "150Mi"
    }
  }
```

The above output verifies that we have successfully scaled up the resources of the MongoDB standalone database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```console
kubectl delete mg -n demo mg-standalone
kubectl delete mongodbopsrequest -n demo mops-vscale-standalone
```