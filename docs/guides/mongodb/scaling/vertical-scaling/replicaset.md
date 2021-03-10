---
title: Vertical Scaling MongoDB Replicaset
menu:
  docs_{{ .version }}:
    identifier: mg-vertical-scaling-replicaset
    name: Replicaset
    parent: mg-vertical-scaling
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Vertical Scale MongoDB Replicaset

This guide will show you how to use `KubeDB` Enterprise operator to update the resources of a MongoDB replicaset database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
  - [Replicaset](/docs/guides/mongodb/clustering/replicaset.md) 
  - [MongoDBOpsRequest](/docs/guides/mongodb/concepts/opsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/mongodb/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on Replicaset

Here, we are going to deploy a  `MongoDB` replicaset using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare MongoDB Replicaset Database

Now, we are going to deploy a `MongoDB` replicaset database with version `4.2.3`.

### Deploy MongoDB replicaset 

In this section, we are going to deploy a MongoDB replicaset database. Then, in the next section we will update the resources of the database using `MongoDBOpsRequest` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-replicaset
  namespace: demo
spec:
  version: "4.2.3"
  replicaSet: 
    name: "replicaset"
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

Let's create the `MongoDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/scaling/mg-replicaset.yaml
mongodb.kubedb.com/mg-replicaset created
```

Now, wait until `mg-replicaset` has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo
NAME            VERSION    STATUS    AGE
mg-replicaset   4.2.3      Ready     3m46s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo mg-replicaset-0 -o json | jq '.spec.containers[].resources'
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

You can see the Pod has the default resources which is assigned by Kubedb operator.

We are now ready to apply the `MongoDBOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of the replicaset database to meet the desired resources after scaling.

#### Create MongoDBOpsRequest

In order to update the resources of the database, we have to create a `MongoDBOpsRequest` CR with our desired resources. Below is the YAML of the `MongoDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-vscale-replicaset
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: mg-replicaset
  verticalScaling:
    replicaSet:
      requests:
        memory: "1.2Gi"
        cpu: "0.6"
      limits:
        memory: "1.2Gi"
        cpu: "0.6"
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `mops-vscale-replicaset` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.replicaSet` specifies the desired resources after scaling.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/scaling/vertical-scaling/mops-vscale-replicaset.yaml
mongodbopsrequest.ops.kubedb.com/mops-vscale-replicaset created
```

#### Verify MongoDB Replicaset resources updated successfully 

If everything goes well, `KubeDB` Enterprise operator will update the resources of `MongoDB` object and related `StatefulSets` and `Pods`.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                     TYPE              STATUS       AGE
mops-vscale-replicaset   VerticalScaling   Successful   3m56s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-vscale-replicaset
Name:         mops-vscale-replicaset
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2021-03-02T17:03:37Z
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
        f:databaseRef:
          .:
          f:name:
        f:type:
        f:verticalScaling:
          .:
          f:replicaSet:
            .:
            f:limits:
            f:requests:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2021-03-02T17:03:37Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:spec:
        f:verticalScaling:
          f:replicaSet:
            f:limits:
              f:cpu:
              f:memory:
            f:requests:
              f:cpu:
              f:memory:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-enterprise
    Operation:       Update
    Time:            2021-03-02T17:03:37Z
  Resource Version:  154015
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-vscale-replicaset
  UID:               5149dd51-7538-421d-b3ca-23dfa2b77b95
Spec:
  Database Ref:
    Name:  mg-replicaset
  Type:    VerticalScaling
  Vertical Scaling:
    Replica Set:
      Limits:
        Cpu:     0.6
        Memory:  1.2Gi
      Requests:
        Cpu:     0.6
        Memory:  1.2Gi
Status:
  Conditions:
    Last Transition Time:  2021-03-02T17:03:37Z
    Message:               MongoDB ops request is vertically scaling database
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2021-03-02T17:03:37Z
    Message:               Successfully updated StatefulSets Resources
    Observed Generation:   1
    Reason:                UpdateStatefulSetResources
    Status:                True
    Type:                  UpdateStatefulSetResources
    Last Transition Time:  2021-03-02T17:05:33Z
    Message:               Successfully Vertically Scaled Replicaset Resources
    Observed Generation:   1
    Reason:                UpdateReplicaSetResources
    Status:                True
    Type:                  UpdateReplicaSetResources
    Last Transition Time:  2021-03-02T17:05:33Z
    Message:               Successfully Vertically Scaled Database
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                      Age    From                        Message
  ----    ------                      ----   ----                        -------
  Normal  PauseDatabase               2m13s  KubeDB Enterprise Operator  Pausing MongoDB demo/mg-replicaset
  Normal  PauseDatabase               2m13s  KubeDB Enterprise Operator  Successfully paused MongoDB demo/mg-replicaset
  Normal  Starting                    2m13s  KubeDB Enterprise Operator  Updating Resources of StatefulSet: mg-replicaset
  Normal  UpdateStatefulSetResources  2m13s  KubeDB Enterprise Operator  Successfully updated StatefulSets Resources
  Normal  Starting                    2m13s  KubeDB Enterprise Operator  Updating Resources of StatefulSet: mg-replicaset
  Normal  UpdateStatefulSetResources  2m13s  KubeDB Enterprise Operator  Successfully updated StatefulSets Resources
  Normal  UpdateReplicaSetResources   17s    KubeDB Enterprise Operator  Successfully Vertically Scaled Replicaset Resources
  Normal  ResumeDatabase              17s    KubeDB Enterprise Operator  Resuming MongoDB demo/mg-replicaset
  Normal  ResumeDatabase              17s    KubeDB Enterprise Operator  Successfully resumed MongoDB demo/mg-replicaset
  Normal  Successful                  17s    KubeDB Enterprise Operator  Successfully Vertically Scaled Database
```

Now, we are going to verify from one of the Pod yaml whether the resources of the replicaset database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo mg-replicaset-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "600m",
    "memory": "1288490188800m"
  },
  "requests": {
    "cpu": "600m",
    "memory": "1288490188800m"
  }
}
```

The above output verifies that we have successfully scaled up the resources of the MongoDB replicaset database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo mg-replicaset
kubectl delete mongodbopsrequest -n demo mops-vscale-replicaset
```