---
title: Vertical Scaling Druid Cluster
menu:
  docs_{{ .version }}:
    identifier: guides-druid-scaling-vertical-scaling-guide
    name: Druid Vertical Scaling
    parent: guides-druid-scaling-vertical-scaling
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale Druid Topology Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to update the resources of a Druid topology cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Druid](/docs/guides/druid/concepts/druid.md)
    - [Topology](/docs/guides/druid/clustering/overview/index.md)
    - [DruidOpsRequest](/docs/guides/druid/concepts/druidopsrequest.md)
    - [Vertical Scaling Overview](/docs/guides/druid/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/druid](/docs/examples/druid) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on Topology Cluster

Here, we are going to deploy a `Druid` topology cluster using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare Druid Topology Cluster

Now, we are going to deploy a `Druid` topology cluster database with version `28.0.1`.

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/scaling/vertical-scaling/yamls/deep-storage-config.yaml
secret/deep-storage-config created
```

### Deploy Druid Cluster

In this section, we are going to deploy a Druid topology cluster. Then, in the next section we will update the resources of the database using `DruidOpsRequest` CRD. Below is the YAML of the `Druid` CR that we are going to create,

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
    configuration:
      secretName: deep-storage-config
  topology:
    routers:
      replicas: 1
  deletionPolicy: Delete
```

Let's create the `Druid` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/druid/scaling/vertical-scaling/yamls/druid-cluster.yaml
druid.kubedb.com/druid-cluster created
```

Now, wait until `druid-cluster` has status `Ready`. i.e,

```bash
$ kubectl get dr -n demo -w
NAME             TYPE                  VERSION    STATUS         AGE
druid-cluster    kubedb.com/v1aplha2   28.0.1     Provisioning   0s
druid-cluster    kubedb.com/v1aplha2   28.0.1     Provisioning   24s
.
.
druid-cluster    kubedb.com/v1aplha2   28.0.1     Ready          92s
```

Let's check the Pod containers resources for both `coordinators` and `historicals` of the Druid topology cluster. Run the following command to get the resources of the `coordinators` and `historicals` containers of the Druid topology cluster

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
```

```bash
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
This is the default resources of the Druid topology cluster set by the `KubeDB` operator.

We are now ready to apply the `DruidOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of the topology cluster to meet the desired resources after scaling.

#### Create DruidOpsRequest

In order to update the resources of the database, we have to create a `DruidOpsRequest` CR with our desired resources. Below is the YAML of the `DruidOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: druid-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: druid-cluster
  verticalScaling:
    coordinators:
      resources:
        requests:
          memory: "1.2Gi"
          cpu: "0.6"
        limits:
          memory: "1.2Gi"
          cpu: "0.6"
    historicals:
      resources:
        requests:
          memory: "1.1Gi"
          cpu: "0.6"
        limits:
          memory: "1.1Gi"
          cpu: "0.6"
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `druid-cluster` cluster.
- `spec.type` specifies that we are performing `VerticalScaling` on druid.
- `spec.VerticalScaling.coordinators` specifies the desired resources of `coordinators` node after scaling.
- `spec.VerticalScaling.historicals` specifies the desired resources of `historicals` node after scaling.

> **Note:** Similarly you can scale other druid nodes vertically by specifying the following fields:
> - For `overlords` use `spec.verticalScaling.overlords`.
> - For `brokers` use `spec.verticalScaling.brokers`.
> - For `middleManagers` use `spec.verticalScaling.middleManagers`.
> - For `routers` use `spec.verticalScaling.routers`.

Let's create the `DruidOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/scaling/vertical-scaling/yamls/druid-vscale.yaml
druidopsrequest.ops.kubedb.com/druid-vscale created
```

#### Verify Druid cluster resources have been updated successfully

If everything goes well, `KubeDB` Ops-manager operator will update the resources of `Druid` object and related `PetSets` and `Pods`.

Let's wait for `DruidOpsRequest` to be `Successful`.  Run the following command to watch `DruidOpsRequest` CR,

```bash
$ kubectl get druidopsrequest -n demo
NAME            TYPE              STATUS       AGE
druid-vscale    VerticalScaling   Successful   3m56s
```

We can see from the above output that the `DruidOpsRequest` has succeeded. If we describe the `DruidOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe druidopsrequest -n demo druid-vscale
Name:         druid-vscale
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         DruidOpsRequest
Metadata:
  Creation Timestamp:  2024-10-21T12:53:55Z
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
        f:apply:
        f:databaseRef:
        f:timeout:
        f:type:
        f:verticalScaling:
          .:
          f:coordinators:
            .:
            f:resources:
              .:
              f:limits:
                .:
                f:cpu:
                f:memory:
              f:requests:
                .:
                f:cpu:
                f:memory:
          f:historicals:
            .:
            f:resources:
              .:
              f:limits:
                .:
                f:cpu:
                f:memory:
              f:requests:
                .:
                f:cpu:
                f:memory:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2024-10-21T12:53:55Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-ops-manager
    Operation:       Update
    Subresource:     status
    Time:            2024-10-21T12:54:23Z
  Resource Version:  102002
  UID:               fe8bb22f-02e8-4a10-9a78-fc211371d581
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   druid-cluster
  Timeout:  5m
  Type:     VerticalScaling
  Vertical Scaling:
    Coordinators:
      Resources:
        Limits:
          Cpu:     0.6
          Memory:  1.2Gi
        Requests:
          Cpu:     0.6
          Memory:  1.2Gi
    Historicals:
      Resources:
        Limits:
          Cpu:     0.6
          Memory:  1.1Gi
        Requests:
          Cpu:     0.6
          Memory:  1.1Gi
Status:
  Conditions:
    Last Transition Time:  2024-10-21T12:53:55Z
    Message:               Druid ops-request has started to vertically scale the Druid nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2024-10-21T12:53:58Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-21T12:54:23Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-10-21T12:54:03Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-coordinators-0
    Last Transition Time:  2024-10-21T12:54:03Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-coordinators-0
    Last Transition Time:  2024-10-21T12:54:08Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-coordinators-0
    Last Transition Time:  2024-10-21T12:54:13Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-historicals-0
    Last Transition Time:  2024-10-21T12:54:13Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-historicals-0
    Last Transition Time:  2024-10-21T12:54:18Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-historicals-0
    Last Transition Time:  2024-10-21T12:54:23Z
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
  Normal   Starting                                                                       67s   KubeDB Ops-manager Operator  Start processing for DruidOpsRequest: demo/druid-vscale
  Normal   Starting                                                                       67s   KubeDB Ops-manager Operator  Pausing Druid databse: demo/druid-cluster
  Normal   Successful                                                                     67s   KubeDB Ops-manager Operator  Successfully paused Druid database: demo/druid-cluster for DruidOpsRequest: druid-vscale
  Normal   UpdatePetSets                                                                  64s   KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0            59s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0          59s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0  54s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0             49s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0           49s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0   44s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Normal   RestartPods                                                                    39s   KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                                       39s   KubeDB Ops-manager Operator  Resuming Druid database: demo/druid-cluster
  Normal   Successful                                                                     39s   KubeDB Ops-manager Operator  Successfully resumed Druid database: demo/druid-cluster for DruidOpsRequest: druid-vscale
```
Now, we are going to verify from one of the Pod yaml whether the resources of the topology cluster has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo druid-cluster-coordinators-0 -o json | jq '.spec.containers[].resources'
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
$ kubectl get pod -n demo druid-cluster-historicals-1 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "600m",
    "memory": "1181116006400m"
  },
  "requests": {
    "cpu": "600m",
    "memory": "1181116006400m"
  }
}
```

The above output verifies that we have successfully scaled up the resources of the Druid topology cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete dr -n demo druid-cluster
kubectl delete druidopsrequest -n demo druid-vscale
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Druid object](/docs/guides/druid/concepts/druid.md).
- Different Druid topology clustering modes [here](/docs/guides/druid/clustering/_index.md).
- Monitor your Druid database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/druid/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Druid database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/druid/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
