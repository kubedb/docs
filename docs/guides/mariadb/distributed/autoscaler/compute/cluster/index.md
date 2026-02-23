---
title: Distributed MariaDB Cluster Compute Autoscaling
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-distributed-autoscaling-compute-cluster
    name: Cluster
    parent: guides-mariadb-distributed-autoscaling-compute
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a Distributed MariaDB Cluster

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. cpu and memory of a distributed MariaDB Galera cluster deployed across multiple Kubernetes clusters.

## Before You Begin

- At first, you need to have a multi-cluster Kubernetes setup with OCM and KubeSlice configured. Follow the [Distributed MariaDB Overview](/docs/guides/mariadb/distributed/overview) guide to set up the required infrastructure.

- Install `KubeDB` Community, Ops-Manager and Autoscaler operator in your hub cluster following the steps [here](/docs/setup/README.md). Make sure to enable OCM support:
  ```bash
  --set petset.features.ocm.enabled=true
  ```

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- Install Prometheus in each spoke cluster. The autoscaler queries the per-cluster Prometheus endpoint (configured in the `PlacementPolicy`) to collect resource usage metrics. You can install it from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

- You should be familiar with the following `KubeDB` concepts:
  - [MariaDB](/docs/guides/mariadb/concepts/mariadb)
  - [MariaDBAutoscaler](/docs/guides/mariadb/concepts/autoscaler)
  - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/opsrequest)
  - [Distributed MariaDB Compute Autoscaling Overview](/docs/guides/mariadb/distributed/autoscaler/compute/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Autoscaling of Distributed Cluster Database

Here, we are going to deploy a distributed `MariaDB` Galera cluster using a supported version by `KubeDB` operator. Then we are going to apply `MariaDBAutoscaler` to set up autoscaling.

### Deploy PlacementPolicy

For distributed MariaDB autoscaling, the `PlacementPolicy` must include a `monitoring.prometheus.url` for each spoke cluster. The autoscaler uses these endpoints to query resource usage metrics from Prometheus instances running in the respective clusters.

Below is the YAML of the `PlacementPolicy` that we are going to create. It distributes 4 replicas across two clusters and provides the Prometheus endpoint for each:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  labels:
    app.kubernetes.io/managed-by: Helm
  name: distributed-mariadb
spec:
  clusterSpreadConstraint:
    distributionRules:
      - clusterName: demo-controller
        monitoring:
          prometheus:
            url: http://prometheus-operated.monitoring.svc.cluster.local:9090
        replicaIndices:
          - 0
          - 2
      - clusterName: demo-worker
        monitoring:
          prometheus:
            url: http://prometheus-operated.monitoring.svc.cluster.local:9090
        replicaIndices:
          - 1
    slice:
      projectNamespace: kubeslice-demo-distributed-mariadb
      sliceName: demo-slice
  nodeSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: ScheduleAnyway
  zoneSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: ScheduleAnyway
```

Here,

- `spec.clusterSpreadConstraint.distributionRules[].monitoring.prometheus.url` specifies the Prometheus endpoint for the corresponding spoke cluster. The autoscaler uses this URL to scrape CPU and memory usage metrics for pods running in that cluster.
- `spec.clusterSpreadConstraint.distributionRules[].replicaIndices` specifies which MariaDB replica indices are scheduled on that cluster. Here `demo-controller` hosts replicas `0` and `2`, and `demo-worker` hosts replica `1`.

Apply the `PlacementPolicy` on the hub (`demo-controller`) cluster:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/distributed/autoscaler/compute/cluster/examples/placement-policy.yaml --context demo-controller
placementpolicy.apps.k8s.appscode.com/distributed-mariadb created
```

> **Note:** Update the `monitoring.prometheus.url` values to match the actual Prometheus service endpoints in each of your spoke clusters.

### Deploy Distributed MariaDB Cluster

In this section, we are going to deploy a distributed MariaDB Galera cluster with version `11.5.2`. Then, in the next section we will set up autoscaling for this database using `MariaDBAutoscaler` CRD.

Below is the YAML of the `MariaDB` CR that we are going to create. Note that `spec.distributed` is set to `true` and the `PlacementPolicy` is referenced via `spec.podTemplate.spec.podPlacementPolicy`:

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "11.5.2"
  distributed: true
  replicas: 3
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
      podPlacementPolicy:
        name: distributed-mariadb
      containers:
      - name: mariadb
        resources:
          requests:
            cpu: "200m"
            memory: "300Mi"
          limits:
            cpu: "200m"
            memory: "300Mi"
  deletionPolicy: WipeOut
```

Let's create the `MariaDB` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/distributed/autoscaler/compute/cluster/examples/sample-mariadb.yaml --context demo-controller
mariadb.kubedb.com/sample-mariadb created
```

Now, wait until `sample-mariadb` has status `Ready`. i.e,

```bash
$ kubectl get mariadb -n demo --context demo-controller
NAME             VERSION   STATUS   AGE
sample-mariadb   11.5.2   Ready    14m
```

The pods are distributed across clusters as defined by the `PlacementPolicy`:

```bash
$ kubectl get pod -n demo --context demo-controller
NAME               READY   STATUS    RESTARTS   AGE
sample-mariadb-0   3/3     Running   0          14m
sample-mariadb-2   3/3     Running   0          14m

$ kubectl get pod -n demo --context demo-worker
NAME               READY   STATUS    RESTARTS   AGE
sample-mariadb-1   3/3     Running   0          14m
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo sample-mariadb-0 -o json --context demo-worker | jq '.spec.containers[].resources'
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

Let's check the MariaDB resources,
```bash
$ kubectl get mariadb -n demo sample-mariadb -o json --context demo-controller | jq '.spec.podTemplate.spec.containers[] | select(.name == "mariadb") | .resources'
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

In order to set up compute resource autoscaling for this distributed database cluster, we have to create a `MariaDBAutoscaler` CRO with our desired configuration. Below is the YAML of the `MariaDBAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MariaDBAutoscaler
metadata:
  name: md-as-compute
  namespace: demo
spec:
  databaseRef:
    name: sample-mariadb
  opsRequestOptions:
    timeout: 3m
    apply: IfReady
  compute:
    mariadb:
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

- `spec.databaseRef.name` specifies that we are performing compute resource scaling operation on `sample-mariadb` database.
- `spec.compute.mariadb.trigger` specifies that compute autoscaling is enabled for this database.
- `spec.compute.mariadb.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.mariadb.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.mariadb.minAllowed` specifies the minimum allowed resources for the database.
- `spec.compute.mariadb.maxAllowed` specifies the maximum allowed resources for the database.
- `spec.compute.mariadb.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.mariadb.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions.apply` has two supported value : `IfReady` & `Always`.
Use `IfReady` if you want to process the opsReq only when the database is Ready. And use `Always` if you want to process the execution of opsReq irrespective of the Database state.
- `spec.opsRequestOptions.timeout` specifies the maximum time for each step of the opsRequest(in seconds).
If a step doesn't finish within the specified timeout, the ops request will result in failure.

> **Note:** The autoscaler collects resource metrics for each pod by querying the Prometheus endpoint of the spoke cluster where that pod is scheduled, as configured in the `PlacementPolicy`.

Let's create the `MariaDBAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/distributed/autoscaler/compute/cluster/examples/mdas-compute.yaml --context demo-controller
mariadbautoscaler.autoscaling.kubedb.com/mdas-compute created
```

#### Verify Autoscaling is set up successfully

Let's check that the `mariadbautoscaler` resource is created successfully,

```bash
$ kubectl get mariadbautoscaler -n demo --context demo-controller
NAME            AGE
md-as-compute   5m56s

$ kubectl describe mariadbautoscaler md-as-compute -n demo --context demo-controller
Name:         md-as-compute
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         MariaDBAutoscaler
Metadata:
  Creation Timestamp:  2022-09-16T11:26:58Z
  Generation:          1
  ...
Spec:
  Compute:
    Mariadb:
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
    Name:  sample-mariadb
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
      Vpa Object Name:    sample-mariadb
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
      Container Name:     mariadb
      Vpa Object Name:    sample-mariadb
    Total Samples Count:  19
    Version:              v3
  Conditions:
    Last Transition Time:  2022-09-16T11:27:07Z
    Message:               Successfully created mariaDBOpsRequest demo/mdops-sample-mariadb-6xc1kc
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
        Container Name:  mariadb
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
    Vpa Name:      sample-mariadb
Events:            <none>

```
So, the `mariadbautoscaler` resource is created successfully.

We can verify from the above output that `status.vpas` contains the `RecommendationProvided` condition to true. And in the same time, `status.vpas.recommendation.containerRecommendations` contain the actual generated recommendation.

Our autoscaler operator continuously watches the recommendation generated and creates an `mariadbopsrequest` based on the recommendations, if the database pod resources are needed to scaled up or down.

Let's watch the `mariadbopsrequest` in the demo namespace to see if any `mariadbopsrequest` object is created. After some time you'll see that a `mariadbopsrequest` will be created based on the recommendation.

```bash
$ kubectl get mariadbopsrequest -n demo --context demo-controller
NAME                          TYPE              STATUS       AGE
mdops-sample-mariadb-6xc1kc   VerticalScaling   Progressing  7s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get mariadbopsrequest -n demo --context demo-controller
NAME                              TYPE              STATUS       AGE
mdops-vpa-sample-mariadb-z43wc8   VerticalScaling   Successful   3m32s
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. If we describe the `MariaDBOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe mariadbopsrequest -n demo mdops-vpa-sample-mariadb-z43wc8 --context demo-controller
Name:         mdops-sample-mariadb-6xc1kc
Namespace:    demo
...
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   sample-mariadb
  Timeout:  2m0s
  Type:     VerticalScaling
  Vertical Scaling:
    Mariadb:
      Limits:
        Cpu:     250m
        Memory:  400Mi
      Requests:
        Cpu:     250m
        Memory:  400Mi
Status:
  Conditions:
    ...
    Last Transition Time:  2022-09-16T11:30:47Z
    Message:               Vertical scale successful for MariaDBOpsRequest: demo/mdops-sample-mariadb-6xc1kc
    Observed Generation:   1
    Reason:                SuccessfullyPerformedVerticalScaling
    Status:                True
    Type:                  VerticalScaling
    ...
  Phase:  Successful
```

Now, we are going to verify from the Pod, and the MariaDB yaml whether the resources of the distributed cluster database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo sample-mariadb-0 -o json --context demo-worker | jq '.spec.containers[].resources'
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

$ kubectl get mariadb -n demo sample-mariadb -o json --context demo-controller | jq '.spec.podTemplate.spec.containers[] | select(.name == "mariadb") | .resources'
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


The above output verifies that we have successfully autoscaled the resources of the distributed MariaDB cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mariadb -n demo sample-mariadb --context demo-controller
kubectl delete mariadbautoscaler -n demo md-as-compute --context demo-controller
kubectl delete placementpolicy distributed-mariadb --context demo-controller
kubectl delete ns demo --context demo-controller
kubectl delete ns demo --context demo-worker
```
