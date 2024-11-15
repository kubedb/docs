---
title: MSSQLServer Availability Group Cluster Autoscaling
menu:
  docs_{{ .version }}:
    identifier: ms-autoscaling-cluster
    name: Cluster
    parent: ms-compute-autoscaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a MSSQLServer Availability Group Cluster Database

This guide will show you how to use `KubeDB` to auto-scale compute resources i.e. cpu and memory of a MSSQLServer cluster database.

## Before You Begin


- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.MSSQLServer=true` to ensure MSSQLServer CRD installation.

- To configure TLS/SSL in `MSSQLServer`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
  - [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md)
  - [MSSQLServerOpsRequest](/docs/guides/mssqlserver/concepts/opsrequest.md)
  - [Compute Resource Autoscaling Overview](/docs/guides/mssqlserver/autoscaler/compute/overview.md)
  - [MSSQLServerAutoscaler](/docs/guides/mssqlserver/concepts/autoscaler.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
## Autoscaling of MSSQLServer Availability Group Cluster

Here, we are going to deploy a `MSSQLServer` Availability Group Cluster using a supported version by `KubeDB` operator. Then we are going to apply `MSSQLServerAutoscaler` to set up autoscaling.

#### Deploy MSSQLServer Availability Group Cluster

First, an issuer needs to be created, even if TLS is not enabled for SQL Server. The issuer will be used to configure the TLS-enabled Wal-G proxy server, which is required for the SQL Server backup and restore operations.

### Create Issuer/ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,
```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=MSSQLServer/O=kubedb"
```
- Create a secret using the certificate files we have just generated,
```bash
$ kubectl create secret tls mssqlserver-ca --cert=ca.crt  --key=ca.key --namespace=demo 
secret/mssqlserver-ca created
```
Now, we are going to create an `Issuer` using the `mssqlserver-ca` secret that contains the ca-certificate we have just created. Below is the YAML of the `Issuer` CR that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
 name: mssqlserver-ca-issuer
 namespace: demo
spec:
 ca:
   secretName: mssqlserver-ca
```

Let’s create the `Issuer` CR we have shown above,
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/ag-cluster/mssqlserver-ca-issuer.yaml
issuer.cert-manager.io/mssqlserver-ca-issuer created
```

In this section, we are going to deploy a MSSQLServer Availability Group Cluster with version `2022-cu12`. Then, in the next section we will set up autoscaling for this database using `MSSQLServerAutoscaler` CRD. Below is the YAML of the `MSSQLServer` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssqlserver-ag-cluster
  namespace: demo
spec:
  version: "2022-cu12"
  replicas: 3
  topology:
    mode: AvailabilityGroup
    availabilityGroup:
      databases:
        - agdb1
        - agdb2
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation # Change it 
          resources:
            requests:
              cpu: "500m"
              memory: "1.5Gi"
            limits:
              cpu: "600m"
              memory: "1.6Gi"
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `MSSQLServer` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/autoscaler/compute/mssqlserver-ag-cluster.yaml
mssqlserver.kubedb.com/mssqlserver-ag-cluster created
```

Now, wait until `mssqlserver-ag-cluster` has status `Ready`. i.e,

```bash
$ kubectl get mssqlserver -n demo
NAME                     VERSION     STATUS   AGE
mssqlserver-ag-cluster   2022-cu12   Ready    8m27s
```

Let's check the MSSQLServer resources,
```bash
$ kubectl get ms -n demo mssqlserver-ag-cluster -o json | jq '.spec.podTemplate.spec.containers[] | select(.name == "mssql") | .resources'
{
  "limits": {
    "cpu": "600m",
    "memory": "1717986918400m"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1536Mi"
  }
}
```


Let's check the Pod containers resources, there are two containers here, first one with index 0 named `mssql` is the main container of mssqlserver. 

```bash
$ kubectl get pod -n demo mssqlserver-ag-cluster-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "600m",
    "memory": "1717986918400m"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1536Mi"
  }
}
$ kubectl get pod -n demo mssqlserver-ag-cluster-1 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "600m",
    "memory": "1717986918400m"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1536Mi"
  }
}
$ kubectl get pod -n demo mssqlserver-ag-cluster-2 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "600m",
    "memory": "1717986918400m"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1536Mi"
  }
}
```


You can see from the above outputs that the resources are same as the one we have assigned while deploying the mssqlserver.

We are now ready to apply the `MSSQLServerAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute resource autoscaling using a `MSSQLServerAutoscaler` Object.

#### Create MSSQLServerAutoscaler Object

In order to set up compute resource autoscaling for this database cluster, we have to create a `MSSQLServerAutoscaler` CRO with our desired configuration. Below is the YAML of the `MSSQLServerAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MSSQLServerAutoscaler
metadata:
  name: ms-as-compute
  namespace: demo
spec:
  databaseRef:
    name: mssqlserver-ag-cluster
  opsRequestOptions:
    timeout: 5m
    apply: IfReady
  compute:
    mssqlserver:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 10
      minAllowed:
        cpu: 800m
        memory: 2Gi
      maxAllowed:
        cpu: 1
        memory: 3Gi
      containerControlledValues: "RequestsAndLimits"
      controlledResources: ["cpu", "memory"]
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource scaling operation on `mssqlserver-ag-cluster` database.
- `spec.compute.mssqlserver.trigger` specifies that compute autoscaling is enabled for this database.
- `spec.compute.mssqlserver.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.mssqlserver.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
  If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.mssqlserver.minAllowed` specifies the minimum allowed resources for the database.
- `spec.compute.mssqlserver.maxAllowed` specifies the maximum allowed resources for the database.
- `spec.compute.mssqlserver.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.mssqlserver.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions.apply` has two supported value : `IfReady` & `Always`.
  Use `IfReady` if you want to process the opsReq only when the database is Ready. And use `Always` if you want to process the execution of opsReq irrespective of the Database state.
- `spec.opsRequestOptions.timeout` specifies the maximum time for each step of the opsRequest(in seconds).
  If a step doesn't finish within the specified timeout, the ops request will result in failure.


Let's create the `MSSQLServerAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/autoscaler/compute/ms-as-compute.yaml
mssqlserverautoscaler.autoscaling.kubedb.com/ms-as-compute created
```

#### Verify Autoscaling is set up successfully

Let's check that the `mssqlserverautoscaler` resource is created successfully,

```bash
$ kubectl get mssqlserverautoscaler -n demo
NAME            AGE
ms-as-compute   16s

$ kubectl describe mssqlserverautoscaler ms-as-compute -n demo
Name:         ms-as-compute
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         MSSQLServerAutoscaler
Metadata:
  Creation Timestamp:  2024-10-25T15:02:58Z
  Generation:          1
  Resource Version:    106200
  UID:                 cc34737b-2e42-4b94-bcc4-cfcac98eb6a6
Spec:
  Compute:
    Mssqlserver:
      Container Controlled Values:  RequestsAndLimits
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     1
        Memory:  3Gi
      Min Allowed:
        Cpu:                     800m
        Memory:                  2Gi
      Pod Life Time Threshold:   5m
      Resource Diff Percentage:  10
      Trigger:                   On
  Database Ref:
    Name:  mssqlserver-ag-cluster
  Ops Request Options:
    Apply:    IfReady
    Timeout:  5m
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              0
        Weight:             524
        Index:              20
        Weight:             456
        Index:              28
        Weight:             2635
        Index:              34
        Weight:             455
        Index:              35
        Weight:             10000
        Index:              36
        Weight:             6980
      Reference Timestamp:  2024-10-25T15:10:00Z
      Total Weight:         2.465794209092962
    First Sample Start:     2024-10-25T15:03:11Z
    Last Sample Start:      2024-10-25T15:13:21Z
    Last Update Time:       2024-10-25T15:13:34Z
    Memory Histogram:
      Bucket Weights:
        Index:              36
        Weight:             10000
        Index:              37
        Weight:             5023
        Index:              39
        Weight:             5710
        Index:              40
        Weight:             2918
      Reference Timestamp:  2024-10-25T15:15:00Z
      Total Weight:         2.8324869288693995
    Ref:
      Container Name:     mssql
      Vpa Object Name:    mssqlserver-ag-cluster
    Total Samples Count:  30
    Version:              v3
    Cpu Histogram:
      Bucket Weights:
        Index:              0
        Weight:             10000
        Index:              1
        Weight:             3741
        Index:              2
        Weight:             1924
      Reference Timestamp:  2024-10-25T15:10:00Z
      Total Weight:         2.033798492571757
    First Sample Start:     2024-10-25T15:03:11Z
    Last Sample Start:      2024-10-25T15:12:22Z
    Last Update Time:       2024-10-25T15:12:34Z
    Memory Histogram:
      Bucket Weights:
        Index:              3
        Weight:             1357
        Index:              4
        Weight:             10000
      Reference Timestamp:  2024-10-25T15:15:00Z
      Total Weight:         2.8324869288693995
    Ref:
      Container Name:     mssql-coordinator
      Vpa Object Name:    mssqlserver-ag-cluster
    Total Samples Count:  26
    Version:              v3
  Conditions:
    Last Transition Time:  2024-10-25T15:10:27Z
    Message:               Successfully created MSSQLServerOpsRequest demo/msops-mssqlserver-ag-cluster-v5xep9
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2024-10-25T15:03:34Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  mssql
        Lower Bound:
          Cpu:     844m
          Memory:  2Gi
        Target:
          Cpu:     1
          Memory:  2Gi
        Uncapped Target:
          Cpu:     1168m
          Memory:  1389197403
        Upper Bound:
          Cpu:           1
          Memory:        3Gi
        Container Name:  mssql-coordinator
        Lower Bound:
          Cpu:     50m
          Memory:  131072k
        Target:
          Cpu:     50m
          Memory:  131072k
        Uncapped Target:
          Cpu:     50m
          Memory:  131072k
        Upper Bound:
          Cpu:     4992m
          Memory:  9063982612
    Vpa Name:      mssqlserver-ag-cluster
Events:            <none>
```
So, the `mssqlserverautoscaler` resource is created successfully.

We can verify from the above output that `status.vpas` contains the `RecommendationProvided` condition to true. And in the same time, `status.vpas.recommendation.containerRecommendations` contain the actual generated recommendation.

Our autoscaler operator continuously watches the recommendation generated and creates an `mssqlserveropsrequest` based on the recommendations, if the database pod resources are needed to scaled up or down.

Let's watch the `mssqlserveropsrequest` in the demo namespace to see if any `mssqlserveropsrequest` object is created. After some time you'll see that a `mssqlserveropsrequest` will be created based on the recommendation.

```bash
$ kubectl get mssqlserveropsrequest -n demo
NAME                          TYPE              STATUS       AGE
msops-mssqlserver-ag-cluster-6xc1kc   VerticalScaling   Progressing  7s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get mssqlserveropsrequest -n demo
NAME                                  TYPE              STATUS       AGE
msops-mssqlserver-ag-cluster-8li26q   VerticalScaling   Successful   11m
```

We can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe msops -n demo msops-mssqlserver-ag-cluster-8li26q
Name:         msops-mssqlserver-ag-cluster-8li26q
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=mssqlserver-ag-cluster
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=mssqlservers.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2024-10-25T15:04:27Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  MSSQLServerAutoscaler
    Name:                  ms-as-compute
    UID:                   cc34737b-2e42-4b94-bcc4-cfcac98eb6a6
  Resource Version:        105300
  UID:                     b2f29a6a-f4cf-4c97-871c-f203e08af320
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   mssqlserver-ag-cluster
  Timeout:  5m0s
  Type:     VerticalScaling
  Vertical Scaling:
    Mssqlserver:
      Resources:
        Limits:
          Cpu:     960m
          Memory:  2290649225
        Requests:
          Cpu:     800m
          Memory:  2Gi
Status:
  Conditions:
    Last Transition Time:  2024-10-25T15:04:27Z
    Message:               MSSQLServer ops-request has started to vertically scaling the MSSQLServer nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2024-10-25T15:04:30Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-10-25T15:04:30Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-25T15:04:35Z
    Message:               get pod; ConditionStatus:True; PodName:mssqlserver-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssqlserver-ag-cluster-0
    Last Transition Time:  2024-10-25T15:04:35Z
    Message:               evict pod; ConditionStatus:True; PodName:mssqlserver-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssqlserver-ag-cluster-0
    Last Transition Time:  2024-10-25T15:05:15Z
    Message:               check pod running; ConditionStatus:True; PodName:mssqlserver-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssqlserver-ag-cluster-0
    Last Transition Time:  2024-10-25T15:05:20Z
    Message:               get pod; ConditionStatus:True; PodName:mssqlserver-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssqlserver-ag-cluster-1
    Last Transition Time:  2024-10-25T15:05:20Z
    Message:               evict pod; ConditionStatus:True; PodName:mssqlserver-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssqlserver-ag-cluster-1
    Last Transition Time:  2024-10-25T15:05:55Z
    Message:               check pod running; ConditionStatus:True; PodName:mssqlserver-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssqlserver-ag-cluster-1
    Last Transition Time:  2024-10-25T15:06:00Z
    Message:               get pod; ConditionStatus:True; PodName:mssqlserver-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssqlserver-ag-cluster-2
    Last Transition Time:  2024-10-25T15:06:00Z
    Message:               evict pod; ConditionStatus:True; PodName:mssqlserver-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssqlserver-ag-cluster-2
    Last Transition Time:  2024-10-25T15:06:35Z
    Message:               check pod running; ConditionStatus:True; PodName:mssqlserver-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssqlserver-ag-cluster-2
    Last Transition Time:  2024-10-25T15:06:40Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-10-25T15:06:40Z
    Message:               Successfully completed the VerticalScaling for MSSQLServer
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
```

Now, we are going to verify from the Pod, and the MSSQLServer yaml whether the resources of the cluster database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo mssqlserver-ag-cluster-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "960m",
    "memory": "2290649225"
  },
  "requests": {
    "cpu": "800m",
    "memory": "2Gi"
  }
}

$ kubectl get ms -n demo mssqlserver-ag-cluster -o json | jq '.spec.podTemplate.spec.containers[] | select(.name == "mssql") | .resources'
{
  "limits": {
    "cpu": "960m",
    "memory": "2290649225"
  },
  "requests": {
    "cpu": "800m",
    "memory": "2Gi"
  }
}
```


The above output verifies that we have successfully autoscaled the resources of the MSSQLServer cluster.


### Autoscaling for Standalone MSSQLServer
Autoscaling for Standalone MSSQLServer is exactly same as cluster mode. Just refer the standalone mssqlserver in `databaseRef` field of `MSSQLServerAutoscaler` spec.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mssqlserver -n demo mssqlserver-ag-cluster
kubectl delete mssqlserverautoscaler -n demo ms-as-compute
kubectl delete issuer -n demo mssqlserver-ca-issuer
kubectl delete secret -n demo mssqlserver-ca
kubectl delete ns demo
```