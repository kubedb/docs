---
title: Druid Topology Autoscaling
menu:
  docs_{{ .version }}:
    identifier: guides-druid-autoscaler-guide
    name: Topology Cluster
    parent: guides-druid-autoscaler
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a Druid Topology Cluster

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. cpu and memory of a Druid topology cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
    - [Druid](/docs/guides/druid/concepts/druid.md)
    - [DruidAutoscaler](/docs/guides/druid/concepts/druidautoscaler.md)
    - [DruidOpsRequest](/docs/guides/druid/concepts/druidopsrequest.md)
    - [Compute Resource Autoscaling Overview](/docs/guides/druid/autoscaler/compute/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/druid](/docs/examples/druid) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Autoscaling of Topology Cluster

Here, we are going to deploy a `Druid` Topology Cluster using a supported version by `KubeDB` operator. Then we are going to apply `DruidAutoscaler` to set up autoscaling.

### Create External Dependency (Deep Storage)

Before proceeding further, we need to prepare deep storage, which is one of the external dependency of Druid and used for storing the segments. It is a storage mechanism that Apache Druid does not provide. **Amazon S3**, **Google Cloud Storage**, or **Azure Blob Storage**, **S3-compatible storage** (like **Minio**), or **HDFS** are generally convenient options for deep storage.

In this tutorial, we will run a `minio-server` as deep storage in our local `kind` cluster using `minio-operator` and create a bucket named `druid` in it, which the deployed druid database will use.

```bash
$ helm repo add minio https://operator.min.io/
$ helm repo update minio
$ helm upgrade --install --namespace "minio-operator" --create-namespace "minio-operator" minio/operator --set operator.replicaCount=1

$ helm upgrade --install --namespace "demo" --create-namespace druid-minio minio/tenant \
--set tenant.pools[0].servers=1 \
--set tenant.pools[0].volumesPerServer=1 \
--set tenant.pools[0].size=1Gi \
--set tenant.certificate.requestAutoCert=false \
--set tenant.buckets[0].name="druid" \
--set tenant.pools[0].name="default"

```

Now we need to create a `Secret` named `deep-storage-config`. It contains the necessary connection information using which the druid database will connect to the deep storage.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: deep-storage-config
  namespace: demo
stringData:
  druid.storage.type: "s3"
  druid.storage.bucket: "druid"
  druid.storage.baseKey: "druid/segments"
  druid.s3.accessKey: "minio"
  druid.s3.secretKey: "minio123"
  druid.s3.protocol: "http"
  druid.s3.enablePathStyleAccess: "true"
  druid.s3.endpoint.signingRegion: "us-east-1"
  druid.s3.endpoint.url: "http://myminio-hl.demo.svc.cluster.local:9000/"
```

Let’s create the `deep-storage-config` Secret shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/autoscaler/compute/yamls/deep-storage-config.yaml
secret/deep-storage-config created
```

Now, we are going to deploy a `Druid` combined cluster with version `28.0.1`.

### Deploy Druid Cluster

In this section, we are going to deploy a Druid Topology cluster with version `28.0.1`.  Then, in the next section we will set up autoscaling for this database using `DruidAutoscaler` CRD. Below is the YAML of the `Druid` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-cluster
  namespace: demo
spec:
  version: 28.0.1
  deepStorage:
    type: s3
    configSecret:
      name: deep-storage-config
  topology:
    routers:
      replicas: 1
  deletionPolicy: WipeOut
```

Let's create the `Druid` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/autoscaler/compute/yamls/druid-cluster.yaml
druid.kubedb.com/druid-cluster created
```

Now, wait until `druid-cluster` has status `Ready`. i.e,

```bash
$ kubectl get kf -n demo -w
NAME             TYPE                  VERSION    STATUS         AGE
druid-cluster    kubedb.com/v1alpha2   28.0.1     Provisioning   0s
druid-cluster    kubedb.com/v1alpha2   28.0.1     Provisioning   24s
.
.
druid-cluster    kubedb.com/v1alpha2   28.0.1     Ready          118s
```

## Druid Topology Autoscaler

Let's check the Druid resources for coordinators and historicals,

```bash
$ kubectl get druid -n demo druid-cluster -o json | jq '.spec.topology.coordinators.podTemplate.spec.containers[].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}

$ kubectl get druid -n demo druid-cluster -o json | jq '.spec.topology.historicals.podTemplate.spec.containers[].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

Let's check the coordinators and historicals Pod containers resources,

```bash
$ kubectl get pod -n demo druid-cluster-coordinators-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}

$ kubectl get pod -n demo druid-cluster-historicals-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

You can see from the above outputs that the resources for coordinators and historicals are same as the one we have assigned while deploying the druid.

We are now ready to apply the `DruidAutoscaler` CRO to set up autoscaling for these coordinators and historicals nodes.

### Compute Resource Autoscaling

Here, we are going to set up compute resource autoscaling using a DruidAutoscaler Object.

#### Create DruidAutoscaler Object

In order to set up compute resource autoscaling for this topology cluster, we have to create a `DruidAutoscaler` CRO with our desired configuration. Below is the YAML of the `DruidAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: DruidAutoscaler
metadata:
  name: druid-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: druid-quickstart
  compute:
    coordinators:
      trigger: "On"
      podLifeTimeThreshold: 1m
      minAllowed:
        cpu: 600m
        memory: 2Gi
      maxAllowed:
        cpu: 1000m
        memory: 5Gi
      resourceDiffPercentage: 20
      controlledResources: ["cpu", "memory"]
    historicals:
      trigger: "On"
      podLifeTimeThreshold: 1m
      minAllowed:
        cpu: 600m
        memory: 2Gi
      maxAllowed:
        cpu: 1000m
        memory: 5Gi
      resourceDiffPercentage: 20
      controlledResources: [ "cpu", "memory"]
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource scaling operation on `druid-cluster` cluster.
- `spec.compute.coordinators.trigger` specifies that compute autoscaling is enabled for this node.
- `spec.compute.coordinators.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.coordinators.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%. If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.coordinators.minAllowed` specifies the minimum allowed resources for the cluster.
- `spec.compute.coordinators.maxAllowed` specifies the maximum allowed resources for the cluster.
- `spec.compute.coordinators.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.coordinators.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.compute.historicals` can be configured the same way shown above. 
- `spec.opsRequestOptions` contains the options to pass to the created OpsRequest. It has 2 fields.
    - `timeout` specifies the timeout for the OpsRequest.
    - `apply` specifies when the OpsRequest should be applied. The default is "IfReady".

> **Note:** You can also configure autoscaling configurations for all other nodes as well. You can apply autoscaler for each node in separate YAML or combinedly in one a YAML as shown above. 

Let's create the `DruidAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/autoscaler/compute/yamls/druid-autoscaler.yaml
druidautoscaler.autoscaling.kubedb.com/druid-autoscaler created
```

#### Verify Autoscaling is set up successfully

Let's check that the `druidautoscaler` resource is created successfully,

```bash
$ kubectl describe druidautoscaler druid-autoscaler -n demo
 kubectl describe druidautoscaler druid-autoscaler -n demo
Name:         druid-autoscaler
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         DruidAutoscaler
Metadata:
  Creation Timestamp:  2024-10-24T10:04:22Z
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
          f:coordinators:
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
            f:resourceDiffPercentage:
            f:trigger:
          f:historicals:
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
            f:resourceDiffPercentage:
            f:trigger:
        f:databaseRef:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2024-10-24T10:04:22Z
    API Version:  autoscaling.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:ownerReferences:
          .:
          k:{"uid":"c2a5c29d-3589-49d8-bc18-585b9c05bf8d"}:
    Manager:      kubedb-autoscaler
    Operation:    Update
    Time:         2024-10-24T10:04:22Z
    API Version:  autoscaling.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:checkpoints:
        f:conditions:
        f:vpas:
    Manager:      kubedb-autoscaler
    Operation:    Update
    Subresource:  status
    Time:         2024-10-24T10:16:20Z
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  Druid
    Name:                  druid-cluster
    UID:                   c2a5c29d-3589-49d8-bc18-585b9c05bf8d
  Resource Version:        274969
  UID:                     069fbdd7-87ad-4fd7-acc7-9753fa188312
Spec:
  Compute:
    Coordinators:
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     1000m
        Memory:  5Gi
      Min Allowed:
        Cpu:                     600m
        Memory:                  2Gi
      Pod Life Time Threshold:   1m
      Resource Diff Percentage:  20
      Trigger:                   On
    Historicals:
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     1000m
        Memory:  5Gi
      Min Allowed:
        Cpu:                     600m
        Memory:                  2Gi
      Pod Life Time Threshold:   1m
      Resource Diff Percentage:  20
      Trigger:                   On
  Database Ref:
    Name:  druid-cluster
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              0
        Weight:             10000
        Index:              5
        Weight:             490
      Reference Timestamp:  2024-10-24T10:05:00Z
      Total Weight:         2.871430450948392
    First Sample Start:     2024-10-24T10:05:07Z
    Last Sample Start:      2024-10-24T10:16:03Z
    Last Update Time:       2024-10-24T10:16:20Z
    Memory Histogram:
      Bucket Weights:
        Index:              25
        Weight:             3648
        Index:              29
        Weight:             10000
      Reference Timestamp:  2024-10-24T10:10:00Z
      Total Weight:         3.3099198846728424
    Ref:
      Container Name:     druid
      Vpa Object Name:    druid-cluster-historicals
    Total Samples Count:  12
    Version:              v3
    Cpu Histogram:
      Bucket Weights:
        Index:              0
        Weight:             3040
        Index:              1
        Weight:             10000
        Index:              2
        Weight:             3278
        Index:              14
        Weight:             1299
      Reference Timestamp:  2024-10-24T10:10:00Z
      Total Weight:         1.0092715955023177
    First Sample Start:     2024-10-24T10:04:53Z
    Last Sample Start:      2024-10-24T10:14:03Z
    Last Update Time:       2024-10-24T10:14:20Z
    Memory Histogram:
      Bucket Weights:
        Index:              24
        Weight:             10000
        Index:              27
        Weight:             8706
      Reference Timestamp:  2024-10-24T10:10:00Z
      Total Weight:         3.204567438391289
    Ref:
      Container Name:     druid
      Vpa Object Name:    druid-cluster-coordinators
    Total Samples Count:  10
    Version:              v3
  Conditions:
    Last Transition Time:  2024-10-24T10:07:19Z
    Message:               Successfully created druidOpsRequest demo/drops-druid-cluster-coordinators-g02xtu
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2024-10-24T10:05:19Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  druid
        Lower Bound:
          Cpu:     600m
          Memory:  2Gi
        Target:
          Cpu:     600m
          Memory:  2Gi
        Uncapped Target:
          Cpu:     100m
          Memory:  764046746
        Upper Bound:
          Cpu:     1
          Memory:  5Gi
    Vpa Name:      druid-cluster-historicals
    Conditions:
      Last Transition Time:  2024-10-24T10:06:19Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  druid
        Lower Bound:
          Cpu:     600m
          Memory:  2Gi
        Target:
          Cpu:     600m
          Memory:  2Gi
        Uncapped Target:
          Cpu:     100m
          Memory:  671629701
        Upper Bound:
          Cpu:     1
          Memory:  5Gi
    Vpa Name:      druid-cluster-coordinators
Events:            <none>
```
So, the `druidautoscaler` resource is created successfully.

you can see in the `Status.VPAs.Recommendation` section, that recommendation has been generated for our database. Our autoscaler operator continuously watches the recommendation generated and creates an `druidopsrequest` based on the recommendations, if the database pods resources are needed to scaled up or down.

Let's watch the `druidopsrequest` in the demo namespace to see if any `druidopsrequest` object is created. After some time you'll see that a `druidopsrequest` will be created based on the recommendation.

```bash
$ watch kubectl get druidopsrequest -n demo
Every 2.0s: kubectl get druidopsrequest -n demo
NAME                                      TYPE              STATUS        AGE
drops-druid-cluster-coordinators-g02xtu   VerticalScaling   Progressing   8m
drops-druid-cluster-historicals-g3oqje    VerticalScaling   Progressing   8m

```
Progressing
Let's wait for the ops request to become successful.

```bash
$ kubectl get druidopsrequest -n demo
NAME                                      TYPE              STATUS       AGE
drops-druid-cluster-coordinators-g02xtu   VerticalScaling   Successful   12m
drops-druid-cluster-historicals-g3oqje    VerticalScaling   Successful   13m
```

We can see from the above output that the `DruidOpsRequest` has succeeded. If we describe the `DruidOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe druidopsrequests -n demo drops-druid-cluster-coordinators-f6qbth 
Name:         drops-druid-cluster-coordinators-g02xtu
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=druid-cluster
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=druids.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         DruidOpsRequest
Metadata:
  Creation Timestamp:  2024-10-24T10:07:19Z
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
          .:
          k:{"uid":"069fbdd7-87ad-4fd7-acc7-9753fa188312"}:
      f:spec:
        .:
        f:apply:
        f:databaseRef:
        f:type:
        f:verticalScaling:
          .:
          f:coordinators:
            .:
            f:resources:
              .:
              f:limits:
                .:
                f:memory:
              f:requests:
                .:
                f:cpu:
                f:memory:
    Manager:      kubedb-autoscaler
    Operation:    Update
    Time:         2024-10-24T10:07:19Z
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
    Time:         2024-10-24T10:07:43Z
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  DruidAutoscaler
    Name:                  druid-autoscaler
    UID:                   069fbdd7-87ad-4fd7-acc7-9753fa188312
  Resource Version:        273990
  UID:                     d14d964b-f4ae-4570-a296-38e91c802473
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  druid-cluster
  Type:    VerticalScaling
  Vertical Scaling:
    Coordinators:
      Resources:
        Limits:
          Memory:  2Gi
        Requests:
          Cpu:     600m
          Memory:  2Gi
Status:
  Conditions:
    Last Transition Time:  2024-10-24T10:07:19Z
    Message:               Druid ops-request has started to vertically scale the Druid nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2024-10-24T10:07:28Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-24T10:07:43Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-10-24T10:07:33Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-coordinators-0
    Last Transition Time:  2024-10-24T10:07:33Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-coordinators-0
    Last Transition Time:  2024-10-24T10:07:38Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-coordinators-0
    Last Transition Time:  2024-10-24T10:07:43Z
    Message:               Successfully completed the vertical scaling for RabbitMQ
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                         Age   From                         Message
  ----     ------                                                                         ----  ----                         -------
  Normal   Starting                                                                       12m   KubeDB Ops-manager Operator  Start processing for DruidOpsRequest: demo/drops-druid-cluster-coordinators-g02xtu
  Normal   Starting                                                                       12m   KubeDB Ops-manager Operator  Pausing Druid databse: demo/druid-cluster
  Normal   Successful                                                                     12m   KubeDB Ops-manager Operator  Successfully paused Druid database: demo/druid-cluster for DruidOpsRequest: drops-druid-cluster-coordinators-g02xtu
  Normal   UpdatePetSets                                                                  12m   KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0            12m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0          12m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0  12m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Normal   RestartPods                                                                    12m   KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                                       12m   KubeDB Ops-manager Operator  Resuming Druid database: demo/druid-cluster
  Normal   Successful                                                                     12m   KubeDB Ops-manager Operator  Successfully resumed Druid database: demo/druid-cluster for DruidOpsRequest: drops-druid-cluster-coordinators-g02xtu
```

Let's describe the other `DruidOpsRequest` created for scaling of historicals. 

```bash
$ kubectl describe druidopsrequests -n demo drops-druid-cluster-historicals-g3oqje
Name:         drops-druid-cluster-historicals-g3oqje
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=druid-cluster
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=druids.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         DruidOpsRequest
Metadata:
  Creation Timestamp:  2024-10-24T10:06:19Z
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
          .:
          k:{"uid":"069fbdd7-87ad-4fd7-acc7-9753fa188312"}:
      f:spec:
        .:
        f:apply:
        f:databaseRef:
        f:type:
        f:verticalScaling:
          .:
          f:historicals:
            .:
            f:resources:
              .:
              f:limits:
                .:
                f:memory:
              f:requests:
                .:
                f:cpu:
                f:memory:
    Manager:      kubedb-autoscaler
    Operation:    Update
    Time:         2024-10-24T10:06:19Z
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
    Time:         2024-10-24T10:06:37Z
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  DruidAutoscaler
    Name:                  druid-autoscaler
    UID:                   069fbdd7-87ad-4fd7-acc7-9753fa188312
  Resource Version:        273770
  UID:                     fc13624c-42d4-4b03-9448-80f451b1a888
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  druid-cluster
  Type:    VerticalScaling
  Vertical Scaling:
    Historicals:
      Resources:
        Limits:
          Memory:  2Gi
        Requests:
          Cpu:     600m
          Memory:  2Gi
Status:
  Conditions:
    Last Transition Time:  2024-10-24T10:06:19Z
    Message:               Druid ops-request has started to vertically scale the Druid nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2024-10-24T10:06:22Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-24T10:06:37Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-10-24T10:06:27Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-historicals-0
    Last Transition Time:  2024-10-24T10:06:27Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-historicals-0
    Last Transition Time:  2024-10-24T10:06:32Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-historicals-0
    Last Transition Time:  2024-10-24T10:06:37Z
    Message:               Successfully completed the vertical scaling for RabbitMQ
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                        Age   From                         Message
  ----     ------                                                                        ----  ----                         -------
  Normal   Starting                                                                      16m   KubeDB Ops-manager Operator  Start processing for DruidOpsRequest: demo/drops-druid-cluster-historicals-g3oqje
  Normal   Starting                                                                      16m   KubeDB Ops-manager Operator  Pausing Druid databse: demo/druid-cluster
  Normal   Successful                                                                    16m   KubeDB Ops-manager Operator  Successfully paused Druid database: demo/druid-cluster for DruidOpsRequest: drops-druid-cluster-historicals-g3oqje
  Normal   UpdatePetSets                                                                 16m   KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0            16m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0          16m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0  16m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Normal   RestartPods                                                                   16m   KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                                      16m   KubeDB Ops-manager Operator  Resuming Druid database: demo/druid-cluster
  Normal   Successful                                                                    16m   KubeDB Ops-manager Operator  Successfully resumed Druid database: demo/druid-cluster for DruidOpsRequest: drops-druid-cluster-historicals-g3oqje

```

Now, we are going to verify from the Pod, and the Druid yaml whether the resources of the coordinators and historicals node has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo druid-cluster-coordinators-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1536Mi"
  },
  "requests": {
    "cpu": "600m",
    "memory": "1536Mi"
  }
}

$ kubectl get pod -n demo druid-cluster-historicals-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "600m",
    "memory": "2Gi"
  }
}

$ kubectl get druid -n demo druid-cluster -o json | jq '.spec.topology.coordinators.podTemplate.spec.containers[].resources'
{
  "limits": {
    "memory": "1536Mi"
  },
  "requests": {
    "cpu": "600m",
    "memory": "1536Mi"
  }
}

$ kubectl get druid -n demo druid-cluster -o json | jq '.spec.topology.historicals.podTemplate.spec.containers[].resources'
{
  "limits": {
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "600m",
    "memory": "2Gi"
  }
}
```

The above output verifies that we have successfully auto scaled the resources of the Druid topology cluster for coordinators and historicals.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete druidopsrequest -n demo drops-druid-cluster-coordinators-g02xtu drops-druid-cluster-historicals-g3oqje
kubectl delete druidautoscaler -n demo druid-autoscaler
kubectl delete dr -n demo druid-cluster
kubectl delete ns demo
```
## Next Steps

- Detail concepts of [Druid object](/docs/guides/druid/concepts/druid.md).
- Different Druid topology clustering modes [here](/docs/guides/druid/clustering/_index.md).
- Monitor your Druid database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/druid/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Druid database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/druid/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
