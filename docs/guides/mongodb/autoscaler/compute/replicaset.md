---
title: MongoDB Replicaset Autoscaling
menu:
  docs_{{ .version }}:
    identifier: mg-auto-scaling-replicaset
    name: Replicaset
    parent: mg-compute-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Autoscaling the Compute Resource of a MongoDB Replicaset Database

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. cpu and memory of a MongoDB replicaset database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community, Enterprise and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- Install `Vertical Pod Autoscaler` from [here](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler#installation)

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
  - [MongoDBAutoscaler](/docs/guides/mongodb/concepts/autoscaler.md)
  - [MongoDBOpsRequest](/docs/guides/mongodb/concepts/opsrequest.md)
  - [Compute Resource Autoscaling Overview](/docs/guides/mongodb/autoscaler/compute/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Autoscaling of Replicaset Database

Here, we are going to deploy a `MongoDB` Replicaset using a supported version by `KubeDB` operator. Then we are going to apply `MongoDBAutoscaler` to set up autoscaling.

#### Deploy MongoDB Replicaset

In this section, we are going to deploy a MongoDB Replicaset database with version `4.2.3`.  Then, in the next section we will set up autoscaling for this database using `MongoDBAutoscaler` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-rs
  namespace: demo
spec:
  version: "4.2.3"
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
  podTemplate:
    spec:
      resources:
        requests:
          cpu: "200m"
          memory: "300Mi"
        limits:
          cpu: "200m"
          memory: "300Mi"
  terminationPolicy: WipeOut
```

Let's create the `MongoDB` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/autoscaling/compute/mg-rs.yaml
mongodb.kubedb.com/mg-rs created
```

Now, wait until `mg-rs` has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo
NAME    VERSION    STATUS    AGE
mg-rs   4.2.3      Ready     2m53s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo mg-rs-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "200m",
    "memory": "300Mi"
  },
  "requests": {
    "cpu": "200m",
    "memory": "300Mi"
  }
}
```

Let's check the MongoDB resources,
```bash
$ kubectl get mongodb -n demo mg-rs -o json | jq '.spec.podTemplate.spec.resources'
{
  "limits": {
    "cpu": "200m",
    "memory": "300Mi"
  },
  "requests": {
    "cpu": "200m",
    "memory": "300Mi"
  }
}
```

You can see from the above outputs that the resources are same as the one we have assigned while deploying the mongodb.

We are now ready to apply the `MongoDBAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute resource autoscaling using a MongoDBAutoscaler Object.

#### Create MongoDBAutoscaler Object

In order to set up compute resource autoscaling for this replicaset database, we have to create a `MongoDBAutoscaler` CRO with our desired configuration. Below is the YAML of the `MongoDBAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MongoDBAutoscaler
metadata:
  name: mg-as-rs
  namespace: demo
spec:
  databaseRef:
    name: mg-rs
  compute:
    replicaSet:
      trigger: "On"
      podLifeTimeThreshold: 5m
      minAllowed:
        cpu: 250m
        memory: 350Mi
      maxAllowed:
        cpu: 1
        memory: 1Gi
      controlledResources: ["cpu", "memory"]
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource scaling operation on `mg-rs` database.
- `spec.compute.replicaSet.trigger` specifies that compute autoscaling is enabled for this database.
- `spec.compute.replicaSet.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.replicaSet.minAllowed` specifies the minimum allowed resources for the database.
- `spec.compute.replicaSet.maxAllowed` specifies the maximum allowed resources for the database.
- `spec.compute.replicaSet.controlledResources` specifies the resources that are controlled by the autoscaler.

Let's create the `MongoDBAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/autoscaling/compute/mg-as-rs.yaml
mongodbautoscaler.autoscaling.kubedb.com/mg-as-rs created
```

#### Verify Autoscaling is set up successfully

Let's check that the `mongodbautoscaler` resource is created successfully,

```bash
$ kubectl get mongodbautoscaler -n demo
NAME        AGE
mg-as-rs    102s

$ kubectl describe mongodbautoscaler mg-as-rs -n demo
Name:         mg-as-rs
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         MongoDBAutoscaler
Metadata:
  Creation Timestamp:  2021-03-06T19:10:46Z
  Generation:          1
  Managed Fields:
    API Version:  autoscaling.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:compute:
          .:
          f:replicaSet:
            .:
            f:controlledResources:
            f:maxAllowed:
              .:
              f:cpu:
              f:memory:
            f:minAllowed:
              .:
              f:cpu:
              f:memory:
            f:podLifeTimeThreshold:
            f:trigger:
        f:databaseRef:
          .:
          f:name:
    Manager:         kubectl-client-side-apply
    Operation:       Update
    Time:            2021-03-06T19:10:46Z
  Resource Version:  839193
  Self Link:         /apis/autoscaling.kubedb.com/v1alpha1/namespaces/demo/mongodbautoscalers/mg-as-rs
  UID:               9be99253-7475-43fe-a68a-34eaec3225c6
Spec:
  Compute:
    Replica Set:
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     1
        Memory:  1Gi
      Min Allowed:
        Cpu:                    250m
        Memory:                 350Mi
      Pod Life Time Threshold:  5m0s
      Trigger:                  On
  Database Ref:
    Name:  mg-rs
Events:    <none>
```
So, the `mongodbautoscaler` resource is created successfully.

Now, lets verify that the vertical pod autoscaler (vpa) resource is created successfully,

```bash
$ kubectl get vpa -n demo
NAME        AGE
vpa-mg-rs   7s

$ kubectl describe vpa vpa-mg-rs -n demo
Name:         vpa-mg-rs
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.k8s.io/v1
Kind:         VerticalPodAutoscaler
Metadata:
  Creation Timestamp:  2021-03-06T19:10:46Z
  Generation:          2
  Managed Fields:
    API Version:  autoscaling.k8s.io/v1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:ownerReferences:
          .:
          k:{"uid":"9be99253-7475-43fe-a68a-34eaec3225c6"}:
            .:
            f:apiVersion:
            f:blockOwnerDeletion:
            f:controller:
            f:kind:
            f:name:
            f:uid:
      f:spec:
        .:
        f:resourcePolicy:
          .:
          f:containerPolicies:
        f:targetRef:
          .:
          f:apiVersion:
          f:kind:
          f:name:
        f:updatePolicy:
          .:
          f:updateMode:
      f:status:
    Manager:      kubedb-autoscaler
    Operation:    Update
    Time:         2021-03-06T19:10:46Z
    API Version:  autoscaling.k8s.io/v1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        f:conditions:
        f:recommendation:
          .:
          f:containerRecommendations:
    Manager:    recommender
    Operation:  Update
    Time:       2021-03-06T19:10:59Z
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  MongoDBAutoscaler
    Name:                  mg-as-rs
    UID:                   9be99253-7475-43fe-a68a-34eaec3225c6
  Resource Version:        839239
  Self Link:               /apis/autoscaling.k8s.io/v1/namespaces/demo/verticalpodautoscalers/vpa-mg-rs
  UID:                     fd2d9896-2eee-43df-85a6-1b968f8d2862
Spec:
  Resource Policy:
    Container Policies:
      Container Name:  mongodb
      Controlled Resources:
        cpu
        memory
      Controlled Values:  RequestsAndLimits
      Max Allowed:
        Cpu:     1
        Memory:  1Gi
      Min Allowed:
        Cpu:           250m
        Memory:        350Mi
      Container Name:  replication-mode-detector
      Mode:            Off
  Target Ref:
    API Version:  apps/v1
    Kind:         StatefulSet
    Name:         mg-rs
  Update Policy:
    Update Mode:  Off
Status:
  Conditions:
    Last Transition Time:  2021-03-06T07:21:58Z
    Status:                False
    Type:                  RecommendationProvided
  Recommendation:
Events:  <none>
```

So, we can verify from the above output that the `vpa` resource is created successfully. But you can see that the `RecommendationProvided` is false and also the `Recommendation` section of the `vpa` is empty. Let's wait some time and describe the vpa again.

```shell
$ kubectl describe vpa vpa-mg-rs -n demo
Name:         vpa-mg-rs
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.k8s.io/v1
Kind:         VerticalPodAutoscaler
Metadata:
  Creation Timestamp:  2021-03-06T19:10:46Z
  Generation:          2
  Managed Fields:
    API Version:  autoscaling.k8s.io/v1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:ownerReferences:
          .:
          k:{"uid":"9be99253-7475-43fe-a68a-34eaec3225c6"}:
            .:
            f:apiVersion:
            f:blockOwnerDeletion:
            f:controller:
            f:kind:
            f:name:
            f:uid:
      f:spec:
        .:
        f:resourcePolicy:
          .:
          f:containerPolicies:
        f:targetRef:
          .:
          f:apiVersion:
          f:kind:
          f:name:
        f:updatePolicy:
          .:
          f:updateMode:
      f:status:
    Manager:      kubedb-autoscaler
    Operation:    Update
    Time:         2021-03-06T19:10:46Z
    API Version:  autoscaling.k8s.io/v1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        f:conditions:
        f:recommendation:
          .:
          f:containerRecommendations:
    Manager:    recommender
    Operation:  Update
    Time:       2021-03-06T19:10:59Z
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  MongoDBAutoscaler
    Name:                  mg-as-rs
    UID:                   9be99253-7475-43fe-a68a-34eaec3225c6
  Resource Version:        839239
  Self Link:               /apis/autoscaling.k8s.io/v1/namespaces/demo/verticalpodautoscalers/vpa-mg-rs
  UID:                     fd2d9896-2eee-43df-85a6-1b968f8d2862
Spec:
  Resource Policy:
    Container Policies:
      Container Name:  mongodb
      Controlled Resources:
        cpu
        memory
      Controlled Values:  RequestsAndLimits
      Max Allowed:
        Cpu:     1
        Memory:  1Gi
      Min Allowed:
        Cpu:           250m
        Memory:        350Mi
      Container Name:  replication-mode-detector
      Mode:            Off
  Target Ref:
    API Version:  apps/v1
    Kind:         StatefulSet
    Name:         mg-rs
  Update Policy:
    Update Mode:  Off
Status:
  Conditions:
    Last Transition Time:  2021-03-06T19:10:59Z
    Status:                True
    Type:                  RecommendationProvided
  Recommendation:
    Container Recommendations:
      Container Name:  mongodb
      Lower Bound:
        Cpu:     250m
        Memory:  350Mi
      Target:
        Cpu:     250m
        Memory:  350Mi
      Uncapped Target:
        Cpu:     182m
        Memory:  262144k
      Upper Bound:
        Cpu:     1
        Memory:  1Gi
Events:          <none>
```

As you can see from the output the vpa has generated a recommendation for our database. Our autoscaler operator continuously watches the recommendation generated and creates an `mongodbopsrequest` based on the recommendations, if the database pods are needed to scaled up or down.

Let's watch the `mongodbopsrequest` in the demo namespace to see if any `mongodbopsrequest` object is created. After some time you'll see that a `mongodbopsrequest` will be created based on the recommendation.

```bash
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                    TYPE              STATUS       AGE
mops-vpa-mg-rs-l3ulmr   VerticalScaling   Progressing  10s
```

Let's wait for the ops request to become successful.

```bash
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                    TYPE              STATUS       AGE
mops-vpa-mg-rs-l3ulmr   VerticalScaling   Successful   68s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-vpa-mg-rs-l3ulmr
Name:         mops-vpa-mg-rs-l3ulmr
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=mg-rs
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=mongodbs.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2021-03-07T15:55:12Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:labels:
          .:
          f:app.kubernetes.io/component:
          f:app.kubernetes.io/instance:
          f:app.kubernetes.io/managed-by:
          f:app.kubernetes.io/name:
        f:ownerReferences:
      f:spec:
        .:
        f:configuration:
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
              f:cpu:
              f:memory:
            f:requests:
              .:
              f:cpu:
              f:memory:
    Manager:      kubedb-autoscaler
    Operation:    Update
    Time:         2021-03-07T15:55:12Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:    kubedb-enterprise
    Operation:  Update
    Time:       2021-03-07T15:55:12Z
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  MongoDBAutoscaler
    Name:                  mg-as-rs
    UID:                   9be99253-7475-43fe-a68a-34eaec3225c6
  Resource Version:        868114
  Self Link:               /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-vpa-mg-rs-l3ulmr
  UID:                     2029223c-25d3-4ecb-b71f-ff0f93a2e4b6
Spec:
  Configuration:
  Database Ref:
    Name:  mg-rs
  Type:    VerticalScaling
  Vertical Scaling:
    Replica Set:
      Limits:
        Cpu:     250m
        Memory:  350Mi
      Requests:
        Cpu:     250m
        Memory:  350Mi
Status:
  Conditions:
    Last Transition Time:  2021-03-07T15:55:12Z
    Message:               MongoDB ops request is vertically scaling database
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2021-03-07T15:55:12Z
    Message:               Successfully updated StatefulSets Resources
    Observed Generation:   1
    Reason:                UpdateStatefulSetResources
    Status:                True
    Type:                  UpdateStatefulSetResources
    Last Transition Time:  2021-03-07T15:57:22Z
    Message:               Successfully Vertically Scaled Replicaset Resources
    Observed Generation:   1
    Reason:                UpdateReplicaSetResources
    Status:                True
    Type:                  UpdateReplicaSetResources
    Last Transition Time:  2021-03-07T15:57:22Z
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
  Normal  Starting                    2m21s  KubeDB Enterprise Operator  Updating Resources of StatefulSet: mg-rs
  Normal  UpdateStatefulSetResources  2m21s  KubeDB Enterprise Operator  Successfully updated StatefulSets Resources
  Normal  UpdateReplicaSetResources   11s    KubeDB Enterprise Operator  Successfully Vertically Scaled Replicaset Resources
  Normal  ResumeDatabase              11s    KubeDB Enterprise Operator  Resuming MongoDB demo/mg-rs
  Normal  ResumeDatabase              11s    KubeDB Enterprise Operator  Successfully resumed MongoDB demo/mg-rs
  Normal  Successful                  11s    KubeDB Enterprise Operator  Successfully Vertically Scaled Database
```

Now, we are going to verify from the Pod, and the MongoDB yaml whether the resources of the replicaset database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo mg-rs-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "250m",
    "memory": "350Mi"
  },
  "requests": {
    "cpu": "250m",
    "memory": "350Mi"
  }
}

$ kubectl get mongodb -n demo mg-rs -o json | jq '.spec.podTemplate.spec.resources'
{
  "limits": {
    "cpu": "250m",
    "memory": "350Mi"
  },
  "requests": {
    "cpu": "250m",
    "memory": "350Mi"
  }
}
```


The above output verifies that we have successfully auto scaled the resources of the MongoDB replicaset database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo mg-rs
kubectl delete mongodbautoscaler -n demo mg-as-rs
```