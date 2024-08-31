---
title: Postgres Cluster Autoscaling
menu:
  docs_{{ .version }}:
    identifier: pg-auto-scaling-cluster
    name: Cluster
    parent: pg-compute-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a Postgres Cluster Database

This guide will show you how to use `KubeDB` to auto-scale compute resources i.e. cpu and memory of a Postgres cluster database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community, Ops-Manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
  - [Postgres](/docs/guides/postgres/concepts/postgres.md)
  - [PostgresOpsRequest](/docs/guides/postgres/concepts/opsrequest.md)
  - [Compute Resource Autoscaling Overview](/docs/guides/postgres/autoscaler/compute/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
## Autoscaling of Cluster Database

Here, we are going to deploy a `Postgres` Cluster using a supported version by `KubeDB` operator. Then we are going to apply `PostgresAutoscaler` to set up autoscaling.

#### Deploy Postgres Cluster

In this section, we are going to deploy a Postgres Cluster with version `16.1'`. Then, in the next section we will set up autoscaling for this database using `PostgresAutoscaler` CRD. Below is the YAML of the `Postgres` CR that we are going to create,
> If you want to autoscale Postgres `Standalone`, Just remove the `spec.Replicas` from the below yaml and rest of the steps are same.

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: ha-postgres
  namespace: demo
spec:
  version: "16.1"
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
      containers:
      - name: postgres
        resources:
          requests:
            cpu: "200m"
            memory: "512Mi"
          limits:
            cpu: "200m"
            memory: "512Mi"
  deletionPolicy: WipeOut
```

Let's create the `Postgres` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/autoscaler/compute/ha-postgres.yaml
postgres.kubedb.com/ha-postgres created
```

Now, wait until `ha-postgres` has status `Ready`. i.e,

```bash
$ kubectl get postgres -n demo
NAME             VERSION   STATUS   AGE
ha-postgres   16.1    Ready    14m
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo ha-postgres-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "200m",
    "memory": "512Mi"
  },
  "requests": {
    "cpu": "200m",
    "memory": "512Mi"
  }
}
```

Let's check the Postgres resources,
```bash
$ kubectl get postgres -n demo ha-postgres -o json | jq '.spec.podTemplate.spec.resources'
{
  "limits": {
    "cpu": "200m",
    "memory": "512Mi"
  },
  "requests": {
    "cpu": "200m",
    "memory": "512Mi"
  }
}
```

You can see from the above outputs that the resources are same as the one we have assigned while deploying the postgres.

We are now ready to apply the `PostgresAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute resource autoscaling using a PostgresAutoscaler Object.

#### Create PostgresAutoscaler Object

In order to set up compute resource autoscaling for this database cluster, we have to create a `PostgresAutoscaler` CRO with our desired configuration. Below is the YAML of the `PostgresAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: PostgresAutoscaler
metadata:
  name: pg-as-compute
  namespace: demo
spec:
  databaseRef:
    name: ha-postgres
  opsRequestOptions:
    timeout: 3m
    apply: IfReady
  compute:
    postgres:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 250m
        memory: 1Gi
      maxAllowed:
        cpu: 1
        memory: 1Gi
      containerControlledValues: "RequestsAndLimits"
      controlledResources: ["cpu", "memory"]
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource scaling operation on `ha-postgres` database.
- `spec.compute.postgres.trigger` specifies that compute autoscaling is enabled for this database.
- `spec.compute.postgres.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.postgres.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
  If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.postgres.minAllowed` specifies the minimum allowed resources for the database.
- `spec.compute.postgres.maxAllowed` specifies the maximum allowed resources for the database.
- `spec.compute.postgres.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.postgres.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions.apply` has two supported value : `IfReady` & `Always`.
  Use `IfReady` if you want to process the opsReq only when the database is Ready. And use `Always` if you want to process the execution of opsReq irrespective of the Database state.
- `spec.opsRequestOptions.timeout` specifies the maximum time for each step of the opsRequest(in seconds).
  If a step doesn't finish within the specified timeout, the ops request will result in failure.


Let's create the `PostgresAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/autoscaler/compute/pgas-compute.yaml
postgresautoscaler.autoscaling.kubedb.com/pgas-compute created
```

#### Verify Autoscaling is set up successfully

Let's check that the `postgresautoscaler` resource is created successfully,

```bash
$ kubectl get postgresautoscaler -n demo
NAME            AGE
pg-as-compute   5m56s

$ kubectl describe postgresautoscaler pg-as-compute -n demo
Name:         pg-as-compute
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         PostgresAutoscaler
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
          f:postgres:
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
        Memory:                  1Gi
      Pod Life Time Threshold:   5m0s
      Resource Diff Percentage:  20
      Trigger:                   On
  Database Ref:
    Name:  ha-postgres
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
      Vpa Object Name:    ha-postgres
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
      Container Name:     postgres
      Vpa Object Name:    ha-postgres
    Total Samples Count:  19
    Version:              v3
  Conditions:
    Last Transition Time:  2022-09-16T11:27:07Z
    Message:               Successfully created postgresOpsRequest demo/pgops-ha-postgres-6xc1kc
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
        Container Name:  postgres
        Lower Bound:
          Cpu:     250m
          Memory:  1Gi
        Target:
          Cpu:     250m
          Memory:  1Gi
        Uncapped Target:
          Cpu:     25m
          Memory:  262144k
        Upper Bound:
          Cpu:     1
          Memory:  1Gi
    Vpa Name:      ha-postgres
Events:            <none>

```
So, the `postgresautoscaler` resource is created successfully.

We can verify from the above output that `status.vpas` contains the `RecommendationProvided` condition to true. And in the same time, `status.vpas.recommendation.containerRecommendations` contain the actual generated recommendation.

Our autoscaler operator continuously watches the recommendation generated and creates an `postgresopsrequest` based on the recommendations, if the database pod resources are needed to scaled up or down.

Let's watch the `postgresopsrequest` in the demo namespace to see if any `postgresopsrequest` object is created. After some time you'll see that a `postgresopsrequest` will be created based on the recommendation.

```bash
$ kubectl get postgresopsrequest -n demo
NAME                          TYPE              STATUS       AGE
pgops-ha-postgres-6xc1kc   VerticalScaling   Progressing  7s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get postgresopsrequest -n demo
NAME                              TYPE              STATUS       AGE
pgops-vpa-ha-postgres-z43wc8   VerticalScaling   Successful   3m32s
```

We can see from the above output that the `PostgresOpsRequest` has succeeded. If we describe the `PostgresOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe postgresopsrequest -n demo pgops-vpa-ha-postgres-z43wc8
Name:         pgops-ha-postgres-6xc1kc
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PostgresOpsRequest
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
          f:postgres:
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
    Kind:                  PostgresAutoscaler
    Name:                  pg-as-compute
    UID:                   44bd46c3-bbc5-4c4a-aff4-00c7f84c6f58
  Resource Version:        846324
  UID:                     c2b30107-c6d3-44bb-adf3-135edc5d615b
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   ha-postgres
  Timeout:  2m0s
  Type:     VerticalScaling
  Vertical Scaling:
    Mariadb:
      Limits:
        Cpu:     250m
        Memory:  1Gi
      Requests:
        Cpu:     250m
        Memory:  1Gi
Status:
  Conditions:
    Last Transition Time:  2022-09-16T11:27:07Z
    Message:               Controller has started to Progress the PostgresOpsRequest: demo/pgops-ha-postgres-6xc1kc
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-09-16T11:30:42Z
    Message:               Successfully restarted Postgres pods for PostgresOpsRequest: demo/pgops-ha-postgres-6xc1kc
    Observed Generation:   1
    Reason:                SuccessfullyRestatedPetSet
    Status:                True
    Type:                  RestartPetSet
    Last Transition Time:  2022-09-16T11:30:47Z
    Message:               Vertical scale successful for PostgresOpsRequest: demo/pgops-ha-postgres-6xc1kc
    Observed Generation:   1
    Reason:                SuccessfullyPerformedVerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2022-09-16T11:30:47Z
    Message:               Controller has successfully scaled the Postgres demo/pgops-ha-postgres-6xc1kc
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason      Age    From                        Message
  ----    ------      ----   ----                        -------
  Normal  Starting    8m48s  KubeDB Enterprise Operator  Start processing for PostgresOpsRequest: demo/pgops-ha-postgres-6xc1kc
  Normal  Starting    8m48s  KubeDB Enterprise Operator  Pausing Postgres databse: demo/ha-postgres
  Normal  Successful  8m48s  KubeDB Enterprise Operator  Successfully paused Postgres database: demo/ha-postgres for PostgresOpsRequest: pgops-ha-postgres-6xc1kc
  Normal  Starting    8m43s  KubeDB Enterprise Operator  Restarting Pod: demo/ha-postgres-0
  Normal  Starting    7m33s  KubeDB Enterprise Operator  Restarting Pod: demo/ha-postgres-1
  Normal  Starting    6m23s  KubeDB Enterprise Operator  Restarting Pod: demo/ha-postgres-2
  Normal  Successful  5m13s  KubeDB Enterprise Operator  Successfully restarted Postgres pods for PostgresOpsRequest: demo/pgops-ha-postgres-6xc1kc
  Normal  Successful  5m8s   KubeDB Enterprise Operator  Vertical scale successful for PostgresOpsRequest: demo/pgops-ha-postgres-6xc1kc
  Normal  Starting    5m8s   KubeDB Enterprise Operator  Resuming Postgres database: demo/ha-postgres
  Normal  Successful  5m8s   KubeDB Enterprise Operator  Successfully resumed Postgres database: demo/ha-postgres
  Normal  Successful  5m8s   KubeDB Enterprise Operator  Controller has Successfully scaled the Postgres database: demo/ha-postgres
```

Now, we are going to verify from the Pod, and the Postgres yaml whether the resources of the cluster database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo ha-postgres-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "250m",
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "250m",
    "memory": "1Gi"
  }
}

$ kubectl get postgres -n demo ha-postgres -o json | jq '.spec.podTemplate.spec.resources'
{
  "limits": {
    "cpu": "250m",
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "250m",
    "memory": "1Gi"
  }
}
```


The above output verifies that we have successfully autoscaled the resources of the Postgres cluster database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete postgres -n demo ha-postgres
kubectl delete postgresautoscaler -n demo pg-as-compute
kubectl delete ns demo
```