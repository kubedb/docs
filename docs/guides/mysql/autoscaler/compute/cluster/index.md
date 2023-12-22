---
title: MySQL Cluster Autoscaling
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-autoscaling-compute-cluster
    name: Cluster
    parent: guides-mysql-autoscaling-compute
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a MySQL Cluster Database

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. cpu and memory of a MySQL replicaset database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community, Ops-Manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
  - [MySQL](/docs/guides/mysql/concepts/mysqldatabase)
  - [MySQLAutoscaler](/docs/guides/mysql/concepts/autoscaler)
  - [MySQLOpsRequest](/docs/guides/mysql/concepts/opsrequest)
  - [Compute Resource Autoscaling Overview](/docs/guides/mysql/autoscaler/compute/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
## Autoscaling of Cluster Database

Here, we are going to deploy a `MySQL` Cluster using a supported version by `KubeDB` operator. Then we are going to apply `MySQLAutoscaler` to set up autoscaling.

#### Deploy MySQL Cluster

In this section, we are going to deploy a MySQL Cluster with version `10.6.16`. Then, in the next section we will set up autoscaling for this database using `MySQLAutoscaler` CRD. Below is the YAML of the `MySQL` CR that we are going to create,
> If you want to autoscale MySQL `Standalone`, Just remove the `spec.Replicas` from the below yaml and rest of the steps are same.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: sample-mysql
  namespace: demo
spec:
  version: "8.0.35"
  replicas: 3
  topology:
    mode: GroupReplication
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
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

Let's create the `MySQL` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/autoscaler/compute/cluster/examples/sample-mysql.yaml
mysql.kubedb.com/sample-mysql created
```

Now, wait until `sample-mysql` has status `Ready`. i.e,

```bash
$ kubectl get mysql -n demo
NAME             VERSION   STATUS   AGE
sample-mysql     8.0.35    Ready    14m
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo sample-mysql-0 -o json | jq '.spec.containers[].resources'
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

Let's check the MySQL resources,
```bash
$ kubectl get mysql -n demo sample-mysql -o json | jq '.spec.podTemplate.spec.resources'
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

You can see from the above outputs that the resources are same as the one we have assigned while deploying the mysql.

We are now ready to apply the `MySQLAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute resource autoscaling using a MySQLAutoscaler Object.

#### Create MySQLAutoscaler Object

In order to set up compute resource autoscaling for this database cluster, we have to create a `MySQLAutoscaler` CRO with our desired configuration. Below is the YAML of the `MySQLAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MySQLAutoscaler
metadata:
  name: my-as-compute
  namespace: demo
spec:
  databaseRef:
    name: sample-mysql
  opsRequestOptions:
    timeout: 3m
    apply: IfReady
  compute:
    mysql:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 250m
        memory: 400Mi
      maxAllowed:
        cpu: 1
        memory: 1Gi
      containerControlledValues: "RequestsAndLimits"
      controlledResources: ["cpu", "memory"]
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource scaling operation on `sample-mysql` database.
- `spec.compute.mysql.trigger` specifies that compute autoscaling is enabled for this database.
- `spec.compute.mysql.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.mysql.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.mysql.minAllowed` specifies the minimum allowed resources for the database.
- `spec.compute.mysql.maxAllowed` specifies the maximum allowed resources for the database.
- `spec.compute.mysql.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.mysql.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions.apply` has two supported value : `IfReady` & `Always`.
Use `IfReady` if you want to process the opsReq only when the database is Ready. And use `Always` if you want to process the execution of opsReq irrespective of the Database state.
- `spec.opsRequestOptions.timeout` specifies the maximum time for each step of the opsRequest(in seconds).
If a step doesn't finish within the specified timeout, the ops request will result in failure.


Let's create the `MySQLAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/autoscaler/compute/cluster/examples/my-as-compute.yaml
mysqlautoscaler.autoscaling.kubedb.com/my-as-compute created
```

#### Verify Autoscaling is set up successfully

Let's check that the `mysqlautoscaler` resource is created successfully,

```bash
$ kubectl get mysqlautoscaler -n demo
NAME            AGE
my-as-compute   5m56s

$ kubectl describe mysqlautoscaler my-as-compute -n demo
Name:         my-as-compute
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         MySQLAutoscaler
Metadata:
  Creation Timestamp:  2022-09-16T11:26:58Z
  Generation:          1
  Managed Fields:
    ...
Spec:
  Compute:
    MySQL:
      Container Controlled Values:  RequestsAndLimits
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     1
        Memory:  1Gi
      Min Allowed:
        Cpu:                     250m
        Memory:                  400Mi
      Pod Life Time Threshold:   5m0s
      Resource Diff Percentage:  20
      Trigger:                   On
  Database Ref:
    Name:  sample-mysql
  Ops Request Options:
    Apply:    IfReady
    Timeout:  3m0s
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              0
        Weight:             10000
        Index:              46
        Weight:             555
      Reference Timestamp:  2022-09-16T00:00:00Z
      Total Weight:         2.648440345821337
    First Sample Start:     2022-09-16T11:26:48Z
    Last Sample Start:      2022-09-16T11:32:52Z
    Last Update Time:       2022-09-16T11:33:02Z
    Memory Histogram:
      Bucket Weights:
        Index:              1
        Weight:             10000
      Reference Timestamp:  2022-09-17T00:00:00Z
      Total Weight:         1.391848625060675
    Ref:
      Container Name:     md-coordinator
      Vpa Object Name:    sample-mysql
    Total Samples Count:  19
    Version:              v3
    Cpu Histogram:
      Bucket Weights:
        Index:              0
        Weight:             10000
        Index:              3
        Weight:             556
      Reference Timestamp:  2022-09-16T00:00:00Z
      Total Weight:         2.648440345821337
    First Sample Start:     2022-09-16T11:26:48Z
    Last Sample Start:      2022-09-16T11:32:52Z
    Last Update Time:       2022-09-16T11:33:02Z
    Memory Histogram:
      Reference Timestamp:  2022-09-17T00:00:00Z
    Ref:
      Container Name:     mysql
      Vpa Object Name:    sample-mysql
    Total Samples Count:  19
    Version:              v3
  Conditions:
    Last Transition Time:  2022-09-16T11:27:07Z
    Message:               Successfully created mysqlDBOpsRequest demo/myops-sample-mysql-6xc1kc
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2022-09-16T11:27:02Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  mysql
        Lower Bound:
          Cpu:     250m
          Memory:  400Mi
        Target:
          Cpu:     250m
          Memory:  400Mi
        Uncapped Target:
          Cpu:     25m
          Memory:  262144k
        Upper Bound:
          Cpu:     1
          Memory:  1Gi
    Vpa Name:      sample-mysql
Events:            <none>

```
So, the `mysqlautoscaler` resource is created successfully.

We can verify from the above output that `status.vpas` contains the `RecommendationProvided` condition to true. And in the same time, `status.vpas.recommendation.containerRecommendations` contain the actual generated recommendation.

Our autoscaler operator continuously watches the recommendation generated and creates an `mysqlopsrequest` based on the recommendations, if the database pod resources are needed to scaled up or down. 

Let's watch the `mysqlopsrequest` in the demo namespace to see if any `mysqlopsrequest` object is created. After some time you'll see that a `mysqlopsrequest` will be created based on the recommendation.

```bash
$ kubectl get mysqlopsrequest -n demo
NAME                          TYPE              STATUS       AGE
myops-sample-mysql-6xc1kc   VerticalScaling   Progressing  7s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get mysqlopsrequest -n demo
NAME                              TYPE              STATUS       AGE
myops-vpa-sample-mysql-z43wc8   VerticalScaling   Successful   3m32s
```

We can see from the above output that the `MySQLOpsRequest` has succeeded. If we describe the `MySQLOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe mysqlopsrequest -n demo myops-vpa-sample-mysql-z43wc8
Name:         myops-sample-mysql-6xc1kc
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MySQLOpsRequest
Metadata:
  Creation Timestamp:  2022-09-16T11:27:07Z
  Generation:          1
  Managed Fields:
   ...
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  MySQLAutoscaler
    Name:                  my-as-compute
    UID:                   44bd46c3-bbc5-4c4a-aff4-00c7f84c6f58
  Resource Version:        846324
  UID:                     c2b30107-c6d3-44bb-adf3-135edc5d615b
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   sample-mysql
  Timeout:  2m0s
  Type:     VerticalScaling
  Vertical Scaling:
    MySQL:
      Limits:
        Cpu:     250m
        Memory:  400Mi
      Requests:
        Cpu:     250m
        Memory:  400Mi
Status:
  Conditions:
    Last Transition Time:  2022-09-16T11:27:07Z
    Message:               Controller has started to Progress the MySQLOpsRequest: demo/myops-sample-mysql-6xc1kc
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-09-16T11:30:42Z
    Message:               Successfully restarted MySQL pods for MySQLOpsRequest: demo/myops-sample-mysql-6xc1kc
    Observed Generation:   1
    Reason:                SuccessfullyRestatedStatefulSet
    Status:                True
    Type:                  RestartStatefulSet
    Last Transition Time:  2022-09-16T11:30:47Z
    Message:               Vertical scale successful for MySQLOpsRequest: demo/myops-sample-mysql-6xc1kc
    Observed Generation:   1
    Reason:                SuccessfullyPerformedVerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2022-09-16T11:30:47Z
    Message:               Controller has successfully scaled the MySQL demo/myops-sample-mysql-6xc1kc
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason      Age    From                        Message
  ----    ------      ----   ----                        -------
  Normal  Starting    8m48s  KubeDB Enterprise Operator  Start processing for MySQLOpsRequest: demo/myops-sample-mysql-6xc1kc
  Normal  Starting    8m48s  KubeDB Enterprise Operator  Pausing MySQL databse: demo/sample-mysql
  Normal  Successful  8m48s  KubeDB Enterprise Operator  Successfully paused MySQL database: demo/sample-mysql for MySQLOpsRequest: myops-sample-mysql-6xc1kc
  Normal  Starting    8m43s  KubeDB Enterprise Operator  Restarting Pod: demo/sample-mysql-0
  Normal  Starting    7m33s  KubeDB Enterprise Operator  Restarting Pod: demo/sample-mysql-1
  Normal  Starting    6m23s  KubeDB Enterprise Operator  Restarting Pod: demo/sample-mysql-2
  Normal  Successful  5m13s  KubeDB Enterprise Operator  Successfully restarted MySQL pods for MySQLOpsRequest: demo/myops-sample-mysql-6xc1kc
  Normal  Successful  5m8s   KubeDB Enterprise Operator  Vertical scale successful for MySQLOpsRequest: demo/myops-sample-mysql-6xc1kc
  Normal  Starting    5m8s   KubeDB Enterprise Operator  Resuming MySQL database: demo/sample-mysql
  Normal  Successful  5m8s   KubeDB Enterprise Operator  Successfully resumed MySQL database: demo/sample-mysql
  Normal  Successful  5m8s   KubeDB Enterprise Operator  Controller has Successfully scaled the MySQL database: demo/sample-mysql
```

Now, we are going to verify from the Pod, and the MySQL yaml whether the resources of the replicaset database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo sample-mysql-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "250m",
    "memory": "400Mi"
  },
  "requests": {
    "cpu": "250m",
    "memory": "400Mi"
  }
}

$ kubectl get mysql -n demo sample-mysql -o json | jq '.spec.podTemplate.spec.resources'
{
  "limits": {
    "cpu": "250m",
    "memory": "400Mi"
  },
  "requests": {
    "cpu": "250m",
    "memory": "400Mi"
  }
}
```


The above output verifies that we have successfully autoscaled the resources of the MySQL replicaset database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mysql -n demo sample-mysql
kubectl delete mysqlautoscaler -n demo my-as-compute
kubectl delete mysqlopsrequest -n demo myops-vpa-sample-mysql-z43wc8 
kubectl delete mysqlopsrequest -n demo myops-sample-mysql-6xc1kc
kubectl delete ns demo
```