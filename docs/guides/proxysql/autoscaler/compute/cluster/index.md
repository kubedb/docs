---
title: ProxySQL Cluster Autoscaling
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-autoscaling-compute-cluster
    name: Demo
    parent: guides-proxysql-autoscaling-compute
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a ProxySQL Cluster Database

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. cpu and memory of a ProxySQL replicaset database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community, Ops-Manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
  - [ProxySQL](/docs/guides/proxysql/concepts/proxysql)
  - [ProxySQLAutoscaler](/docs/guides/proxysql/concepts/autoscaler)
  - [ProxySQLOpsRequest](/docs/guides/proxysql/concepts/opsrequest)
  - [Compute Resource Autoscaling Overview](/docs/guides/proxysql/autoscaler/compute/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
### Prepare MySQL backend

We need a mysql backend for the proxysql server. So we are creating one with the below yaml.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql-server
  namespace: demo
spec:
  version: "5.7.44"
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
  terminationPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/autoscaler/cluster/examples/sample-mysql.yaml
mysql.kubedb.com/mysql-server created
```

Let's wait for the MySQL to be Ready. 

```bash
$ kubectl get mysql -n demo 
NAME           VERSION   STATUS   AGE
mysql-server   5.7.44    Ready    3m51s
```

## Autoscaling of ProxySQL Cluster

Here, we are going to deploy a `ProxySQL` Cluster using a supported version by `KubeDB` operator. Then we are going to apply `ProxySQLAutoscaler` to set up autoscaling.

### Deploy ProxySQL Cluster

In this section, we are going to deploy a ProxySQL Cluster with version `2.3.2-debian`. Then, in the next section we will set up autoscaling for this database using `ProxySQLAutoscaler` CRD. Below is the YAML of the `ProxySQL` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ProxySQL
metadata:
  name: proxy-server
  namespace: demo
spec:
  version: "2.3.2-debian"
  replicas: 3
  backend:
    name: mysql-server
  syncUsers: true
  terminationPolicy: WipeOut
  podTemplate:
    spec:
      resources:
        limits:
          cpu: 200m
          memory: 300Mi
        requests:
          cpu: 200m
          memory: 300Mi
```

Let's create the `ProxySQL` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/autoscaler/compute/cluster/examples/sample-proxysql.yaml
proxysql.kubedb.com/proxy-server created
```

Now, wait until `proxy-server` has status `Ready`. i.e,

```bash
$ kubectl get proxysql -n demo
NAME             VERSION       STATUS   AGE
proxy-server   2.3.2-debian    Ready    4m
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo proxy-server-0 -o json | jq '.spec.containers[].resources'
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

Let's check the ProxySQL resources,
```bash
$ kubectl get proxysql -n demo proxy-server -o json | jq '.spec.podTemplate.spec.resources'
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

You can see from the above outputs that the resources are same as the one we have assigned while deploying the proxysql.

We are now ready to apply the `ProxySQLAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute resource autoscaling using a ProxySQLAutoscaler Object.

#### Create ProxySQLAutoscaler Object

In order to set up compute resource autoscaling for this proxysql cluster, we have to create a `ProxySQLAutoscaler` CRO with our desired configuration. Below is the YAML of the `ProxySQLAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: ProxySQLAutoscaler
metadata:
  name: proxy-as-compute
  namespace: demo
spec:
  proxyRef:
    name: proxy-server
  opsRequestOptions:
    timeout: 3m
    apply: IfReady
  compute:
    proxysql:
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

- `spec.proxyRef.name` specifies that we are performing compute resource scaling operation on `proxy-server` proxysql server.
- `spec.compute.proxysql.trigger` specifies that compute autoscaling is enabled for this proxysql server.
- `spec.compute.proxysql.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.proxysql.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.proxysql.minAllowed` specifies the minimum allowed resources for the proxysql server.
- `spec.compute.proxysql.maxAllowed` specifies the maximum allowed resources for the proxysql server.
- `spec.compute.proxysql.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.proxysql.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions.apply` has two supported value : `IfReady` & `Always`.
Use `IfReady` if you want to process the opsReq only when the proxysql server is Ready. And use `Always` if you want to process the execution of opsReq irrespective of the proxysql server state.
- `spec.opsRequestOptions.timeout` specifies the maximum time for each step of the opsRequest(in seconds).
If a step doesn't finish within the specified timeout, the ops request will result in failure.


Let's create the `ProxySQLAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/autoscaler/compute/cluster/examples/proxy-as-compute.yaml
proxysqlautoscaler.autoscaling.kubedb.com/proxy-as-compute created
```

#### Verify Autoscaling is set up successfully

Let's check that the `proxysqlautoscaler` resource is created successfully,

```bash
$ kubectl get proxysqlautoscaler -n demo
NAME               AGE
proxy-as-compute   5m56s

$ kubectl describe proxysqlautoscaler proxy-as-compute -n demo
Name:         proxy-as-compute
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         ProxySQLAutoscaler
Metadata:
  Creation Timestamp:  2022-09-16T11:26:58Z
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
          f:proxysql:
            .:
            f:containerControlledValues:
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
            f:resourceDiffPercentage:
            f:trigger:
        f:databaseRef:
          .:
          f:name:
        f:opsRequestOptions:
          .:
          f:apply:
          f:timeout:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2022-09-16T11:26:58Z
    API Version:  autoscaling.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:checkpoints:
        f:conditions:
        f:vpas:
    Manager:         kubedb-autoscaler
    Operation:       Update
    Subresource:     status
    Time:            2022-09-16T11:27:07Z
  Resource Version:  846645
  UID:               44bd46c3-bbc5-4c4a-aff4-00c7f84c6f58
Spec:
  Compute:
    Proxysql:
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
  Proxy Ref:
    Name:  proxy-server
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
      Vpa Object Name:    proxy-server
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
      Container Name:     proxysql
      Vpa Object Name:    proxy-server
    Total Samples Count:  19
    Version:              v3
  Conditions:
    Last Transition Time:  2022-09-16T11:27:07Z
    Message:               Successfully created proxySQLOpsRequest demo/prxops-proxy-server-6xc1kc
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
        Container Name:  proxysql
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
    Vpa Name:      proxy-server
Events:            <none>

```
So, the `proxysqlautoscaler` resource is created successfully.

We can verify from the above output that `status.vpas` contains the `RecommendationProvided` condition to true. And in the same time, `status.vpas.recommendation.containerRecommendations` contain the actual generated recommendation.

Our autoscaler operator continuously watches the recommendation generated and creates an `proxysqlopsrequest` based on the recommendations, if the database pod resources are needed to scaled up or down. 

Let's watch the `proxysqlopsrequest` in the demo namespace to see if any `proxysqlopsrequest` object is created. After some time you'll see that a `proxysqlopsrequest` will be created based on the recommendation.

```bash
$ kubectl get proxysqlopsrequest -n demo
NAME                          TYPE              STATUS       AGE
prxops-proxy-server-6xc1kc   VerticalScaling   Progressing  7s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get proxysqlopsrequest -n demo
NAME                              TYPE              STATUS       AGE
prxops-vpa-proxy-server-z43wc8   VerticalScaling   Successful   3m32s
```

We can see from the above output that the `ProxySQLOpsRequest` has succeeded. If we describe the `ProxySQLOpsRequest` we will get an overview of the steps that were followed to scale the proxysql server.

```bash
$ kubectl describe proxysqlopsrequest -n demo prxops-vpa-proxy-server-z43wc8
Name:         prxops-proxy-server-6xc1kc
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ProxySQLOpsRequest
Metadata:
  Creation Timestamp:  2022-09-16T11:27:07Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:ownerReferences:
          .:
          k:{"uid":"44bd46c3-bbc5-4c4a-aff4-00c7f84c6f58"}:
      f:spec:
        .:
        f:apply:
        f:databaseRef:
          .:
          f:name:
        f:timeout:
        f:type:
        f:verticalScaling:
          .:
          f:proxysql:
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
    Time:         2022-09-16T11:27:07Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:      kubedb-ops-manager
    Operation:    Update
    Subresource:  status
    Time:         2022-09-16T11:27:07Z
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  ProxySQLAutoscaler
    Name:                  proxy-as-compute
    UID:                   44bd46c3-bbc5-4c4a-aff4-00c7f84c6f58
  Resource Version:        846324
  UID:                     c2b30107-c6d3-44bb-adf3-135edc5d615b
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   proxy-server
  Timeout:  2m0s
  Type:     VerticalScaling
  Vertical Scaling:
    Proxysql:
      Limits:
        Cpu:     250m
        Memory:  400Mi
      Requests:
        Cpu:     250m
        Memory:  400Mi
Status:
  Conditions:
    Last Transition Time:  2022-09-16T11:27:07Z
    Message:               Controller has started to Progress the ProxySQLOpsRequest: demo/prxops-proxy-server-6xc1kc
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-09-16T11:30:42Z
    Message:               Successfully restarted ProxySQL pods for ProxySQLOpsRequest: demo/prxops-proxy-server-6xc1kc
    Observed Generation:   1
    Reason:                SuccessfullyRestatedStatefulSet
    Status:                True
    Type:                  RestartStatefulSet
    Last Transition Time:  2022-09-16T11:30:47Z
    Message:               Vertical scale successful for ProxySQLOpsRequest: demo/prxops-proxy-server-6xc1kc
    Observed Generation:   1
    Reason:                SuccessfullyPerformedVerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2022-09-16T11:30:47Z
    Message:               Controller has successfully scaled the ProxySQL demo/prxops-proxy-server-6xc1kc
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason      Age    From                        Message
  ----    ------      ----   ----                        -------
  Normal  Starting    8m48s  KubeDB Enterprise Operator  Start processing for ProxySQLOpsRequest: demo/prxops-proxy-server-6xc1kc
  Normal  Starting    8m48s  KubeDB Enterprise Operator  Pausing ProxySQL databse: demo/proxy-server
  Normal  Successful  8m48s  KubeDB Enterprise Operator  Successfully paused ProxySQL database: demo/proxy-server for ProxySQLOpsRequest: prxops-proxy-server-6xc1kc
  Normal  Starting    8m43s  KubeDB Enterprise Operator  Restarting Pod: demo/proxy-server-0
  Normal  Starting    7m33s  KubeDB Enterprise Operator  Restarting Pod: demo/proxy-server-1
  Normal  Starting    6m23s  KubeDB Enterprise Operator  Restarting Pod: demo/proxy-server-2
  Normal  Successful  5m13s  KubeDB Enterprise Operator  Successfully restarted ProxySQL pods for ProxySQLOpsRequest: demo/prxops-proxy-server-6xc1kc
  Normal  Successful  5m8s   KubeDB Enterprise Operator  Vertical scale successful for ProxySQLOpsRequest: demo/prxops-proxy-server-6xc1kc
  Normal  Starting    5m8s   KubeDB Enterprise Operator  Resuming ProxySQL database: demo/proxy-server
  Normal  Successful  5m8s   KubeDB Enterprise Operator  Successfully resumed ProxySQL database: demo/proxy-server
  Normal  Successful  5m8s   KubeDB Enterprise Operator  Controller has Successfully scaled the ProxySQL database: demo/proxy-server
```

Now, we are going to verify from the Pod, and the ProxySQL yaml whether the resources of the replicaset database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo proxy-server-0 -o json | jq '.spec.containers[].resources'
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

$ kubectl get proxysql -n demo proxy-server -o json | jq '.spec.podTemplate.spec.resources'
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


The above output verifies that we have successfully autoscaled the resources of the ProxySQL replicaset database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete proxysql -n demo proxy-server
kubectl delete proxysqlautoscaler -n demo proxy-as-compute
kubectl delete ns demo
```