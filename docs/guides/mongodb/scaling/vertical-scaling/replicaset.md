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

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Vertical Scale MongoDB Replicaset

This guide will show you how to use `KubeDB` Enterprise operator to update the resources of a MongoDB replicaset database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here]().

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

Now, we are going to deploy a `MongoDB` replicaset database with version `3.6.8`.

### Deploy MongoDB replicaset 

In this section, we are going to deploy a MongoDB replicaset database. Then, in the next section we will update the resources of the database using `MongoDBOpsRequest` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-replicaset
  namespace: demo
spec:
  version: "3.6.8-v1"
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

Now, wait until `mg-replicaset` has status `Running`. i.e,

```bash
$ kubectl get mg -n demo                                                                                                                                             20:05:47
  NAME            VERSION    STATUS    AGE
  mg-replicaset   3.6.8-v1   Running   3m46s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo mg-replicaset-0 -o json | jq '.spec.containers[].resources'
{}
```

You can see the Pod has empty resources that means the scheduler will choose a random node to place the container of the Pod on by default.

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
        memory: "150Mi"
        cpu: "0.1"
      limits:
        memory: "250Mi"
        cpu: "0.2"
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
 Annotations:  API Version:  ops.kubedb.com/v1alpha1
 Kind:         MongoDBOpsRequest
 Metadata:
   Creation Timestamp:  2020-08-25T06:05:39Z
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
           f:replicaSet:
             .:
             f:limits:
               .:
               f:memory:
             f:requests:
               .:
               f:memory:
     Manager:      kubectl
     Operation:    Update
     Time:         2020-08-25T06:05:39Z
     API Version:  ops.kubedb.com/v1alpha1
     Fields Type:  FieldsV1
     fieldsV1:
       f:metadata:
         f:finalizers:
       f:spec:
         f:verticalScaling:
           f:replicaSet:
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
     Time:            2020-08-25T06:09:19Z
   Resource Version:  5034763
   Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-vscale-replicaset
   UID:               65c44484-9216-4767-a42a-20eb04391a9e
 Spec:
   Database Ref:
     Name:  mg-replicaset
   Type:    VerticalScaling
   Vertical Scaling:
     ReplicaSet:
       Limits:
         Cpu:     0.2
         Memory:  250Mi
       Requests:
         Cpu:     0.1
         Memory:  150Mi
 Status:
   Conditions:
     Last Transition Time:  2020-08-25T06:05:39Z
     Message:               MongoDB ops request is being processed
     Observed Generation:   1
     Reason:                Scaling
     Status:                True
     Type:                  Scaling
     Last Transition Time:  2020-08-25T06:05:39Z
     Message:               Successfully halted mongodb: mg-replicaset
     Observed Generation:   1
     Reason:                HaltDatabase
     Status:                True
     Type:                  HaltDatabase
     Last Transition Time:  2020-08-25T06:05:39Z
     Message:               Successfully updated StatefulSets Resources
     Observed Generation:   1
     Reason:                UpdateStatefulSetResources
     Status:                True
     Type:                  UpdateStatefulSetResources
     Last Transition Time:  2020-08-25T06:09:19Z
     Message:               Successfully updated ReplicaSet resources
     Observed Generation:   1
     Reason:                UpdateReplicaSetResources
     Status:                True
     Type:                  UpdateReplicaSetResources
     Last Transition Time:  2020-08-25T06:09:19Z
     Message:               Successfully Resumed mongodb: mg-replicaset
     Observed Generation:   1
     Reason:                ResumeDatabase
     Status:                True
     Type:                  ResumeDatabase
     Last Transition Time:  2020-08-25T06:09:19Z
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
   Normal  HaltDatabase               4m38s  KubeDB Enterprise Operator  Pausing Mongodb mg-replicaset in Namespace demo
   Normal  HaltDatabase               4m38s  KubeDB Enterprise Operator  Successfully Halted Mongodb mg-replicaset in Namespace demo
   Normal  Starting                    4m38s  KubeDB Enterprise Operator  Updating Resources of StatefulSet: mg-replicaset
   Normal  UpdateStatefulSetResources  4m38s  KubeDB Enterprise Operator  Successfully updated StatefulSets Resources
   Normal  UpdateReplicaSetResources   4m38s  KubeDB Enterprise Operator  Updating ReplicaSet Resources
   Normal  UpdateReplicaSetResources   3m38s  KubeDB Enterprise Operator  Successfully Updated Resources of Pod mg-replicaset-1
   Normal  UpdateReplicaSetResources   2m18s  KubeDB Enterprise Operator  Successfully Updated Resources of Pod mg-replicaset-2
   Normal  UpdateReplicaSetResources   58s    KubeDB Enterprise Operator  Successfully Updated Resources of Pod (master): mg-replicaset-0
   Normal  UpdateReplicaSetResources   58s    KubeDB Enterprise Operator  Successfully Updated ReplicaSet Resources
   Normal  ResumeDatabase              58s    KubeDB Enterprise Operator  Resuming MongoDB
   Normal  ResumeDatabase              58s    KubeDB Enterprise Operator  Successfully Resumed mongodb
   Normal  Successful                  58s    KubeDB Enterprise Operator  Successfully Scaled Database
```

Now, we are going to verify from one of the Pod yaml whether the resources of the replicaset database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo mg-replicaset-0 -o json | jq '.spec.containers[].resources'
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

The above output verifies that we have successfully scaled up the resources of the MongoDB replicaset database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo mg-replicaset
kubectl delete mongodbopsrequest -n demo mops-vscale-replicaset
```