---
title: MariaDB Cluster Autoscaling
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-autoscaling-compute-cluster
    name: Cluster
    parent: guides-mariadb-autoscaling-compute
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Autoscaling the Compute Resource of a MariaDB Cluster Database

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. cpu and memory of a MariaDB replicaset database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community, Enterprise and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- Install `Vertical Pod Autoscaler` from [here](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler#installation)

- You should be familiar with the following `KubeDB` concepts:
  - [MariaDB](/docs/guides/mariadb/concepts/mariadb)
  - [MariaDBAutoscaler](/docs/guides/mariadb/concepts/autoscaler)
  - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/opsrequest)
  - [Compute Resource Autoscaling Overview](/docs/guides/mariadb/autoscaler/compute/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
## Autoscaling of Cluster Database

Here, we are going to deploy a `MariaDB` Cluster using a supported version by `KubeDB` operator. Then we are going to apply `MariaDBAutoscaler` to set up autoscaling.

#### Deploy MariaDB Cluster

In this section, we are going to deploy a MariaDB Cluster with version `10.5.8`. Then, in the next section we will set up autoscaling for this database using `MariaDBAutoscaler` CRD. Below is the YAML of the `MariaDB` CR that we are going to create,
> If you want to autoscale MariaDB `Standalone`, Just remove the `spec.Replicas` from the below yaml and rest of the steps are same.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "10.5.8"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "topolvm-provisioner"
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

Let's create the `MariaDB` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/autoscaler/compute/cluster/examples/sample-mariadb.yaml
mariadb.kubedb.com/sample-mariadb created
```

Now, wait until `sample-mariadb` has status `Ready`. i.e,

```bash
$ kubectl get mariadb -n demo
NAME             VERSION   STATUS   AGE
sample-mariadb   10.5.8    Ready    14m
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo sample-mariadb-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "200m",
    "memory": "300Mi",
    "topolvm.cybozu.com/capacity": "1"
  },
  "requests": {
    "cpu": "200m",
    "memory": "300Mi",
    "topolvm.cybozu.com/capacity": "1"
  }
}
```

Let's check the MariaDB resources,
```bash
$ kubectl get mariadb -n demo sample-mariadb -o json | jq '.spec.podTemplate.spec.resources'
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

You can see from the above outputs that the resources are same as the one we have assigned while deploying the mariadb.

We are now ready to apply the `MariaDBAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute resource autoscaling using a MariaDBAutoscaler Object.

#### Create MariaDBAutoscaler Object

In order to set up compute resource autoscaling for this database cluster, we have to create a `MariaDBAutoscaler` CRO with our desired configuration. Below is the YAML of the `MariaDBAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MariaDBAutoscaler
metadata:
  name: mdas-compute
  namespace: demo
spec:
  databaseRef:
    name: sample-mariadb
  compute:
    mariadb:
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

- `spec.databaseRef.name` specifies that we are performing compute resource scaling operation on `sample-mariadb` database.
- `spec.compute.mariadb.trigger` specifies that compute autoscaling is enabled for this database.
- `spec.compute.mariadb.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.mariadb.minAllowed` specifies the minimum allowed resources for the database.
- `spec.compute.mariadb.maxAllowed` specifies the maximum allowed resources for the database.
- `spec.compute.mariadb.controlledResources` specifies the resources that are controlled by the autoscaler.

Let's create the `MariaDBAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/autoscaler/compute/cluster/examples/mdas-compute.yaml
mariadbautoscaler.autoscaling.kubedb.com/mdas-compute created
```

#### Verify Autoscaling is set up successfully

Let's check that the `mariadbautoscaler` resource is created successfully,

```bash
$ kubectl get mariadbautoscaler -n demo
NAME           AGE
mdas-compute   5m13s

$ kubectl describe mariadbautoscaler mdas-compute -n demo
Name:         mdas-compute
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         MariaDBAutoscaler
Metadata:
  Creation Timestamp:  2022-01-13T13:05:56Z
  Generation:          1
  ...
  Resource Version:  50664
  UID:               e2f7f6cc-f2b1-46b5-88b4-2767e1a04b68
Spec:
  Compute:
    Mariadb:
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     1
        Memory:  1Gi
      Min Allowed:
        Cpu:                    250m
        Memory:                 350Mi
      Pod Life Time Threshold:  5m
      Trigger:                  On
  Database Ref:
    Name:  sample-mariadb
Status:
  Conditions:
    Last Transition Time:  2022-01-13T13:07:05Z
    Message:               Successfully created mariaDBOpsRequest demo/mdops-vpa-sample-mariadb-z43wc8
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
Events:                    <none>

```
So, the `mariadbautoscaler` resource is created successfully.

Now, lets verify that the vertical pod autoscaler (vpa) resource is created successfully,

```bash
$ kubectl get vpa -n demo
NAME                 MODE   CPU    MEM     PROVIDED   AGE
vpa-sample-mariadb   Off    250m   350Mi   True       6m3s

$ kubectl describe vpa -n demo 
Name:         vpa-sample-mariadb
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.k8s.io/v1
Kind:         VerticalPodAutoscaler
Metadata:
  Creation Timestamp:  2022-01-13T13:05:56Z
  Generation:          2
  ...
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  MariaDBAutoscaler
    Name:                  mdas-compute
    UID:                   e2f7f6cc-f2b1-46b5-88b4-2767e1a04b68
  Resource Version:        50458
  UID:                     5c876135-fa94-4a80-ab60-d3eb2b3fc69f
Spec:
  Resource Policy:
    Container Policies:
      Container Name:  mariadb
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
      Container Name:  exporter
      Mode:            Off
      Container Name:  md-coordinator
      Mode:            Off
  Target Ref:
    API Version:  apps/v1
    Kind:         StatefulSet
    Name:         sample-mariadb
  Update Policy:
    Update Mode:  Off
Status:
  Conditions:
    Last Transition Time:  2022-01-13T13:06:13Z
    Status:                False
    Type:                  RecommendationProvided
  Recommendation:
Events:          <none>
```

So, we can verify from the above output that the `vpa` resource is created successfully. But you can see that the `RecommendationProvided` is false and also the `Recommendation` section of the `vpa` is empty. Let's wait some time and describe the vpa again.

```shell
$ kubectl describe vpa vpa-sample-mariadb -n demo
Name:         vpa-sample-mariadb
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.k8s.io/v1
Kind:         VerticalPodAutoscaler
Metadata:
  Creation Timestamp:  2021-03-06T19:10:46Z
  Generation: ...
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  MariaDBAutoscaler
    Name:                  mg-as-rs
    UID:                   9be99253-7475-43fe-a68a-34eaec3225c6
  Resource Version:        839239
  Self Link:               /apis/autoscaling.k8s.io/v1/namespaces/demo/verticalpodautoscalers/vpa-sample-mariadb
  UID:                     fd2d9896-2eee-43df-85a6-1b968f8d2862
Spec:
  Resource Policy:
    Container Policies:
      Container Name:  mariadb
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
    Name:         sample-mariadb
  Update Policy:
    Update Mode:  Off
Status:
  Conditions:
    Last Transition Time:  2021-03-06T19:10:59Z
    Status:                True
    Type:                  RecommendationProvided
  Recommendation:
    Container Recommendations:
      Container Name:  mariadb
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

As you can see from the output the vpa has generated a recommendation for our database. Our autoscaler operator continuously watches the recommendation generated and creates an `mariadbopsrequest` based on the recommendations, if the database pods are needed to scaled up or down. If you see that the `RecommendationProvided` is false and also the `Recommendation` section of the `vpa` is empty then wait couple of minutes and describe the vpa again.

Let's watch the `mariadbopsrequest` in the demo namespace to see if any `mariadbopsrequest` object is created. After some time you'll see that a `mariadbopsrequest` will be created based on the recommendation.

```bash
$ kubectl get mariadbopsrequest -n demo
NAME                              TYPE              STATUS       AGE
mdops-vpa-sample-mariadb-z43wc8   VerticalScaling   Progressing  11s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get mariadbopsrequest -n demo
NAME                              TYPE              STATUS       AGE
mdops-vpa-sample-mariadb-z43wc8   VerticalScaling   Successful   2m32s
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. If we describe the `MariaDBOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe mariadbopsrequest -n demo mdops-vpa-sample-mariadb-z43wc8
Name:         mdops-vpa-sample-mariadb-z43wc8
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=sample-mariadb
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=mariadbs.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MariaDBOpsRequest
Metadata:
  Creation Timestamp:  2022-01-13T13:07:05Z
  Generation:          1
  ...
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  MariaDBAutoscaler
    Name:                  mdas-compute
    UID:                   e2f7f6cc-f2b1-46b5-88b4-2767e1a04b68
  Resource Version:        51793
  UID:                     15338f3d-b394-4276-bbd0-52bbf771d06b
Spec:
  Database Ref:
    Name:  sample-mariadb
  Type:    VerticalScaling
  Vertical Scaling:
    Mariadb:
      Limits:
        Cpu:     250m
        Memory:  350Mi
      Requests:
        Cpu:     250m
        Memory:  350Mi
Status:
  Conditions:
    Last Transition Time:  2022-01-13T13:07:05Z
    Message:               Controller has started to Progress the MariaDBOpsRequest: demo/mdops-vpa-sample-mariadb-z43wc8
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-01-13T13:07:05Z
    Message:               Vertical scaling started in MariaDB: demo/sample-mariadb for MariaDBOpsRequest: mdops-vpa-sample-mariadb-z43wc8
    Observed Generation:   1
    Reason:                VerticalScalingStarted
    Status:                True
    Type:                  Scaling
    Last Transition Time:  2022-01-13T13:11:11Z
    Message:               Vertical scaling performed successfully in MariaDB: demo/sample-mariadb for MariaDBOpsRequest: mdops-vpa-sample-mariadb-z43wc8
    Observed Generation:   1
    Reason:                SuccessfullyPerformedVerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2022-01-13T13:11:11Z
    Message:               Controller has successfully scaled the MariaDB demo/mdops-vpa-sample-mariadb-z43wc8
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     3
  Phase:                   Successful
...
```

Now, we are going to verify from the Pod, and the MariaDB yaml whether the resources of the replicaset database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo sample-mariadb-0 -o json | jq '.spec.containers[].resources'
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

$ kubectl get mariadb -n demo sample-mariadb -o json | jq '.spec.podTemplate.spec.resources'
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


The above output verifies that we have successfully auto scaled the resources of the MariaDB replicaset database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mariadb -n demo sample-mariadb
kubectl delete mariadbautoscaler -n demo mdas-compute
kubectl delete ns demo
```