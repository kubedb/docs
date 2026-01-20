---
title: Druid Topology Autoscaling
menu:
  docs_{{ .version }}:
    identifier: guides-druid-autoscaler-storage-guide
    name: Druid Storage Autoscaling
    parent: guides-druid-autoscaler-storage
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a Druid Topology Cluster

This guide will show you how to use `KubeDB` to autoscale the storage of a Druid Topology cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- Install Prometheus from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

- You must have a `StorageClass` that supports volume expansion.

- You should be familiar with the following `KubeDB` concepts:
    - [Druid](/docs/guides/druid/concepts/druid.md)
    - [DruidAutoscaler](/docs/guides/druid/concepts/druidautoscaler.md)
    - [DruidOpsRequest](/docs/guides/druid/concepts/druidopsrequest.md)
    - [Storage Autoscaling Overview](/docs/guides/druid/autoscaler/storage/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/druid](/docs/examples/druid) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Storage Autoscaling of Topology Cluster

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  28h
longhorn (default)     driver.longhorn.io      Delete          Immediate              true                   28h
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   28h
```

We can see from the output the `longhorn` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `Druid` topology using a supported version by `KubeDB` operator. Then we are going to apply `DruidAutoscaler` to set up autoscaling.

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/autoscaler/storage/yamls/deep-storage-config.yaml
secret/deep-storage-config created
```

### Deploy Druid Cluster

In this section, we are going to deploy a Druid topology cluster with monitoring enabled and with version `28.0.1`.  Then, in the next section we will set up autoscaling for this cluster using `DruidAutoscaler` CRD. Below is the YAML of the `Druid` CR that we are going to create,

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
    historicals:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
      storageType: Durable
    middleManagers:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
      storageType: Durable
    routers:
      replicas: 1
  deletionPolicy: Delete
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

Let's create the `Druid` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/autoscaler/storage/yamls/druid-cluster.yaml
druid.kubedb.com/druid-cluster created
```

Now, wait until `druid-cluster` has status `Ready`. i.e,

```bash
$ kubectl get dr -n demo -w
NAME          TYPE            VERSION   STATUS         AGE
druid-cluster    kubedb.com/v1alpha2    28.0.1          Provisioning   0s
druid-cluster    kubedb.com/v1alpha2    28.0.1          Provisioning   24s
.
.
druid-cluster    kubedb.com/v1alpha2   28.0.1           Ready          2m20s
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo druid-cluster-historicals -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"
$ kubectl get petset -n demo druid-cluster-middleManagers -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                                             STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-2c0ef2aa-0438-4d75-9cb2-c12a176bae6a   1Gi        RWO            Delete           Bound    demo/druid-cluster-base-task-dir-druid-cluster-middlemanagers-0   longhorn       <unset>                          95s
pvc-5f4cea5f-e0c8-4339-b67c-9cb8b02ba49d   1Gi        RWO            Delete           Bound    demo/druid-cluster-segment-cache-druid-cluster-historicals-0      longhorn       <unset>                          96s
```

You can see the petset for both historicals and middleManagers has 1GB storage, and the capacity of all the persistent volume is also 1GB.

We are now ready to apply the `DruidAutoscaler` CRO to set up storage autoscaling for this cluster(historicals and middleManagers).

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a DruidAutoscaler Object.

#### Create DruidAutoscaler Object

In order to set up vertical autoscaling for this topology cluster, we have to create a `DruidAutoscaler` CRO with our desired configuration. Below is the YAML of the `DruidAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: DruidAutoscaler
metadata:
  name: druid-storage-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: druid-cluster
  storage:
    historicals:
      expansionMode: "Offline"
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 100
    middleManagers:
      expansionMode: "Offline"
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 100
```

Here,

- `spec.clusterRef.name` specifies that we are performing vertical scaling operation on `druid-cluster` cluster.
- `spec.storage.historicals.trigger/spec.storage.middleManagers.trigger` specifies that storage autoscaling is enabled for historicals and middleManagers of topology cluster.
- `spec.storage.historicals.usageThreshold/spec.storage.middleManagers.usageThreshold` specifies storage usage threshold, if storage usage exceeds `60%` then storage autoscaling will be triggered.
- `spec.storage.historicals.scalingThreshold/spec.storage.historicals.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `100%` of the current amount.
- It has another field `spec.storage.historicals.expansionMode/spec.storage.middleManagers.expansionMode` to set the opsRequest volumeExpansionMode, which support two values: `Online` & `Offline`. Default value is `Online`.

Let's create the `DruidAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/autoscaler/storage/yamls/druid-storage-autoscaler.yaml
druidautoscaler.autoscaling.kubedb.com/druid-storage-autoscaler created
```

#### Storage Autoscaling is set up successfully

Let's check that the `druidautoscaler` resource is created successfully,

```bash
$ kubectl get druidautoscaler -n demo
NAME                       AGE
druid-storage-autoscaler   34s

$ kubectl describe druidautoscaler -n demo druid-storage-autoscaler 
Name:         druid-storage-autoscaler
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         DruidAutoscaler
Metadata:
  Creation Timestamp:  2024-10-25T09:52:37Z
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
        f:databaseRef:
        f:storage:
          .:
          f:historicals:
            .:
            f:expansionMode:
            f:scalingThreshold:
            f:trigger:
            f:usageThreshold:
          f:middleManagers:
            .:
            f:expansionMode:
            f:scalingThreshold:
            f:trigger:
            f:usageThreshold:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2024-10-25T09:52:37Z
    API Version:  autoscaling.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:ownerReferences:
          .:
          k:{"uid":"712730e8-41ef-4700-b184-825b30ecbc8c"}:
    Manager:    kubedb-autoscaler
    Operation:  Update
    Time:       2024-10-25T09:52:37Z
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  Druid
    Name:                  druid-cluster
    UID:                   712730e8-41ef-4700-b184-825b30ecbc8c
  Resource Version:        226662
  UID:                     57cbd906-a9b7-4649-bfe0-304840bb60c1
Spec:
  Database Ref:
    Name:  druid-cluster
  Ops Request Options:
    Apply:  IfReady
  Storage:
    Historicals:
      Expansion Mode:  Offline
      Scaling Rules:
        Applies Upto:     
        Threshold:        100pc
      Scaling Threshold:  100
      Trigger:            On
      Usage Threshold:    60
    Middle Managers:
      Expansion Mode:  Offline
      Scaling Rules:
        Applies Upto:     
        Threshold:        100pc
      Scaling Threshold:  100
      Trigger:            On
      Usage Threshold:    60
Events:                   <none>
```
So, the `druidautoscaler` resource is created successfully.

Now, for this demo, we are going to manually fill up the persistent volume to exceed the `usageThreshold` using `dd` command to see if storage autoscaling is working or not.

We are autoscaling volume for both historicals and middleManagers. So we need to fill up the persistent volume for both historicals and middleManagers.

1. Lets exec into the historicals pod and fill the cluster volume using the following commands:

```bash
$ kubectl exec -it -n demo druid-cluster-historicals-0 -- bash
bash-5.1$ df -h /druid/data/segments
Filesystem                                                Size       Used     Available   Use%  Mounted on
/dev/longhorn/pvc-d4ef15ef-b1af-4a1f-ad25-ad9bc990a2fb    973.4M     92.0K    957.3M      0%    /druid/data/segment

bash-5.1$ dd if=/dev/zero of=/druid/data/segments/file.img bs=600M count=1                                           
1+0 records in                                                                                                       
1+0 records out
629145600 bytes (600.0MB) copied, 46.709228 seconds, 12.8MB/s

bash-5.1$ df -h /druid/data/segments                                      
Filesystem                                                 Size      Used      Available   Use%    Mounted on
/dev/longhorn/pvc-d4ef15ef-b1af-4a1f-ad25-ad9bc990a2fb     973.4M    600.1M    357.3M      63%     /druid/data/segments
```

2. Let's exec into the middleManagers pod and fill the cluster volume using the following commands:

```bash
$ kubectl exec -it -n demo druid-cluster-middleManagers-0 -- bash
druid@druid-cluster-middleManagers-0:~$ df -h /var/druid/task
Filesystem                                              Size       Used     Available   Use%    Mounted on
/dev/longhorn/pvc-2c0ef2aa-0438-4d75-9cb2-c12a176bae6a  973.4M     24.0K    957.4M      0%      /var/druid/task
druid@druid-cluster-middleManagers-0:~$ dd if=/dev/zero of=/var/druid/task/file.img bs=600M count=1
1+0 records in
1+0 records out
629145600 bytes (629 MB, 600 MiB) copied, 3.39618 s, 185 MB/s
druid@druid-cluster-middleManagers-0:~$ df -h /var/druid/task
Filesystem                                              Size      Used      Available   Use%  Mounted on
/dev/longhorn/pvc-2c0ef2aa-0438-4d75-9cb2-c12a176bae6a  973.4M    600.0M    357.4M      63%   /var/druid/task
```

So, from the above output we can see that the storage usage is 63% for both nodes, which exceeded the `usageThreshold` 60%.

There will be two `DruidOpsRequest` created for both historicals and middleManagers to expand the volume of the cluster for both nodes.
Let's watch the `druidopsrequest` in the demo namespace to see if any `druidopsrequest` object is created. After some time you'll see that a `druidopsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
$ watch kubectl get druidopsrequest -n demo
NAME                                                        TYPE              STATUS        AGE
druidopsrequest.ops.kubedb.com/drops-druid-cluster-gq9huj   VolumeExpansion   Progressing   46s
druidopsrequest.ops.kubedb.com/drops-druid-cluster-kbw4fd   VolumeExpansion   Successful    4m46s
```

Once ops request has succeeded. Let's wait for the other one to become successful.

```bash
$ kubectl get druidopsrequest -n demo 
NAME                                                        TYPE              STATUS       AGE
druidopsrequest.ops.kubedb.com/drops-druid-cluster-gq9huj   VolumeExpansion   Successful   3m18s
druidopsrequest.ops.kubedb.com/drops-druid-cluster-kbw4fd   VolumeExpansion   Successful   7m18s
```

We can see from the above output that the both `DruidOpsRequest` has succeeded. If we describe the `DruidOpsRequest` one by one we will get an overview of the steps that were followed to expand the volume of the cluster.

```bash
$ kubectl describe druidopsrequest -n demo drops-druid-cluster-kbw4fd
Name:         drops-druid-cluster-kbw4fd
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=druid-cluster
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=druids.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         DruidOpsRequest
Metadata:
  Creation Timestamp:  2024-10-25T09:57:14Z
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
          k:{"uid":"57cbd906-a9b7-4649-bfe0-304840bb60c1"}:
      f:spec:
        .:
        f:apply:
        f:databaseRef:
        f:type:
        f:volumeExpansion:
          .:
          f:historicals:
          f:mode:
    Manager:      kubedb-autoscaler
    Operation:    Update
    Time:         2024-10-25T09:57:14Z
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
    Time:         2024-10-25T10:00:20Z
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  DruidAutoscaler
    Name:                  druid-storage-autoscaler
    UID:                   57cbd906-a9b7-4649-bfe0-304840bb60c1
  Resource Version:        228016
  UID:                     1fa750bb-2db3-4684-a7cf-1b3047bc07af
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  druid-cluster
  Type:    VolumeExpansion
  Volume Expansion:
    Historicals:  2041405440
    Mode:         Offline
Status:
  Conditions:
    Last Transition Time:  2024-10-25T09:57:14Z
    Message:               Druid ops-request has started to expand volume of druid nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2024-10-25T09:57:22Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2024-10-25T09:57:22Z
    Message:               is pet set deleted; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPetSetDeleted
    Last Transition Time:  2024-10-25T09:57:32Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2024-10-25T09:57:37Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2024-10-25T09:57:37Z
    Message:               is ops req patched; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsOpsReqPatched
    Last Transition Time:  2024-10-25T09:57:37Z
    Message:               create pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreatePod
    Last Transition Time:  2024-10-25T09:57:42Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-10-25T09:57:42Z
    Message:               is pvc patched; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPvcPatched
    Last Transition Time:  2024-10-25T09:59:27Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2024-10-25T09:59:27Z
    Message:               create; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  Create
    Last Transition Time:  2024-10-25T09:59:35Z
    Message:               is druid running; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  IsDruidRunning
    Last Transition Time:  2024-10-25T09:59:57Z
    Message:               successfully updated historicals node PVC sizes
    Observed Generation:   1
    Reason:                UpdateHistoricalsNodePVCs
    Status:                True
    Type:                  UpdateHistoricalsNodePVCs
    Last Transition Time:  2024-10-25T10:00:15Z
    Message:               successfully reconciled the Druid resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-25T10:00:20Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2024-10-25T10:00:20Z
    Message:               Successfully completed volumeExpansion for Druid
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                    Age    From                         Message
  ----     ------                                    ----   ----                         -------
  Normal   Starting                                  8m29s  KubeDB Ops-manager Operator  Start processing for DruidOpsRequest: demo/drops-druid-cluster-kbw4fd
  Normal   Starting                                  8m29s  KubeDB Ops-manager Operator  Pausing Druid databse: demo/druid-cluster
  Normal   Successful                                8m29s  KubeDB Ops-manager Operator  Successfully paused Druid database: demo/druid-cluster for DruidOpsRequest: drops-druid-cluster-kbw4fd
  Warning  get pet set; ConditionStatus:True         8m21s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  is pet set deleted; ConditionStatus:True  8m21s  KubeDB Ops-manager Operator  is pet set deleted; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True         8m16s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   OrphanPetSetPods                          8m11s  KubeDB Ops-manager Operator  successfully deleted the petSets with orphan propagation policy
  Warning  get pod; ConditionStatus:True             8m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  is ops req patched; ConditionStatus:True  8m6s   KubeDB Ops-manager Operator  is ops req patched; ConditionStatus:True
  Warning  create pod; ConditionStatus:True          8m6s   KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             8m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             8m1s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True      8m1s   KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False    8m1s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True             7m56s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             7m56s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             7m51s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             7m51s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             7m46s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             7m46s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             7m41s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             7m41s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             7m36s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             7m36s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             7m31s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             7m31s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             7m26s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             7m26s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             7m21s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             7m21s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             7m16s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             7m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             7m11s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             7m11s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             7m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             7m6s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             7m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             7m1s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             6m56s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             6m56s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             6m51s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             6m51s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             6m46s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             6m46s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             6m41s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             6m41s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             6m36s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             6m36s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             6m31s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             6m31s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             6m26s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             6m26s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             6m21s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             6m21s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             6m16s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             6m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True     6m16s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create; ConditionStatus:True              6m16s  KubeDB Ops-manager Operator  create; ConditionStatus:True
  Warning  is ops req patched; ConditionStatus:True  6m16s  KubeDB Ops-manager Operator  is ops req patched; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             6m11s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  is druid running; ConditionStatus:False   6m8s   KubeDB Ops-manager Operator  is druid running; ConditionStatus:False
  Warning  get pod; ConditionStatus:True             6m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             6m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             5m56s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             5m51s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Normal   UpdateHistoricalsNodePVCs                 5m46s  KubeDB Ops-manager Operator  successfully updated historicals node PVC sizes
  Normal   UpdatePetSets                             5m28s  KubeDB Ops-manager Operator  successfully reconciled the Druid resources
  Warning  get pet set; ConditionStatus:True         5m23s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   ReadyPetSets                              5m23s  KubeDB Ops-manager Operator  PetSet is recreated
  Normal   Starting                                  5m23s  KubeDB Ops-manager Operator  Resuming Druid database: demo/druid-cluster
  Normal   Successful                                5m23s  KubeDB Ops-manager Operator  Successfully resumed Druid database: demo/druid-cluster for DruidOpsRequest: drops-druid-cluster-kbw4fd
  Normal   UpdatePetSets                             5m18s  KubeDB Ops-manager Operator  successfully reconciled the Druid resources
  Normal   UpdatePetSets                             5m8s   KubeDB Ops-manager Operator  successfully reconciled the Druid resources
  Normal   UpdatePetSets                             4m57s  KubeDB Ops-manager Operator  successfully reconciled the Druid resources
```

```bash
$ kubectl describe druidopsrequest -n demo drops-druid-cluster-gq9huj 
Name:         drops-druid-cluster-gq9huj
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=druid-cluster
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=druids.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         DruidOpsRequest
Metadata:
  Creation Timestamp:  2024-10-25T10:01:14Z
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
          k:{"uid":"57cbd906-a9b7-4649-bfe0-304840bb60c1"}:
      f:spec:
        .:
        f:apply:
        f:databaseRef:
        f:type:
        f:volumeExpansion:
          .:
          f:middleManagers:
          f:mode:
    Manager:      kubedb-autoscaler
    Operation:    Update
    Time:         2024-10-25T10:01:14Z
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
    Time:         2024-10-25T10:04:12Z
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  DruidAutoscaler
    Name:                  druid-storage-autoscaler
    UID:                   57cbd906-a9b7-4649-bfe0-304840bb60c1
  Resource Version:        228783
  UID:                     3b97380c-e867-467f-b366-4b50c7cd7d6d
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  druid-cluster
  Type:    VolumeExpansion
  Volume Expansion:
    Middle Managers:  2041405440
    Mode:             Offline
Status:
  Conditions:
    Last Transition Time:  2024-10-25T10:01:14Z
    Message:               Druid ops-request has started to expand volume of druid nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2024-10-25T10:01:22Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2024-10-25T10:01:22Z
    Message:               is pet set deleted; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPetSetDeleted
    Last Transition Time:  2024-10-25T10:01:32Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2024-10-25T10:01:37Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2024-10-25T10:01:37Z
    Message:               is ops req patched; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsOpsReqPatched
    Last Transition Time:  2024-10-25T10:01:37Z
    Message:               create pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreatePod
    Last Transition Time:  2024-10-25T10:01:42Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-10-25T10:01:42Z
    Message:               is pvc patched; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPvcPatched
    Last Transition Time:  2024-10-25T10:03:32Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2024-10-25T10:03:32Z
    Message:               create; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  Create
    Last Transition Time:  2024-10-25T10:03:40Z
    Message:               is druid running; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  IsDruidRunning
    Last Transition Time:  2024-10-25T10:03:52Z
    Message:               successfully updated middleManagers node PVC sizes
    Observed Generation:   1
    Reason:                UpdateMiddleManagersNodePVCs
    Status:                True
    Type:                  UpdateMiddleManagersNodePVCs
    Last Transition Time:  2024-10-25T10:04:07Z
    Message:               successfully reconciled the Druid resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-25T10:04:12Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2024-10-25T10:04:12Z
    Message:               Successfully completed volumeExpansion for Druid
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                    Age    From                         Message
  ----     ------                                    ----   ----                         -------
  Normal   Starting                                  5m33s  KubeDB Ops-manager Operator  Start processing for DruidOpsRequest: demo/drops-druid-cluster-gq9huj
  Normal   Starting                                  5m33s  KubeDB Ops-manager Operator  Pausing Druid databse: demo/druid-cluster
  Normal   Successful                                5m33s  KubeDB Ops-manager Operator  Successfully paused Druid database: demo/druid-cluster for DruidOpsRequest: drops-druid-cluster-gq9huj
  Warning  get pet set; ConditionStatus:True         5m25s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  is pet set deleted; ConditionStatus:True  5m25s  KubeDB Ops-manager Operator  is pet set deleted; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True         5m20s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   OrphanPetSetPods                          5m15s  KubeDB Ops-manager Operator  successfully deleted the petSets with orphan propagation policy
  Warning  get pod; ConditionStatus:True             5m10s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  is ops req patched; ConditionStatus:True  5m10s  KubeDB Ops-manager Operator  is ops req patched; ConditionStatus:True
  Warning  create pod; ConditionStatus:True          5m10s  KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             5m5s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             5m5s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True      5m5s   KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False    5m5s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True             5m     KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             5m     KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m55s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m55s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m50s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m50s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m45s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m45s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m40s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m40s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m35s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m35s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m30s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m30s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m25s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m25s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m20s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m20s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m15s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m15s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m10s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m10s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m5s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m5s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m     KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m     KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             3m55s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             3m55s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             3m50s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             3m50s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             3m45s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             3m45s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             3m40s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             3m40s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             3m35s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             3m35s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             3m30s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             3m30s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             3m25s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             3m25s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             3m20s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             3m20s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             3m15s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             3m15s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True     3m15s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create; ConditionStatus:True              3m15s  KubeDB Ops-manager Operator  create; ConditionStatus:True
  Warning  is ops req patched; ConditionStatus:True  3m15s  KubeDB Ops-manager Operator  is ops req patched; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             3m10s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  is druid running; ConditionStatus:False   3m7s   KubeDB Ops-manager Operator  is druid running; ConditionStatus:False
  Warning  get pod; ConditionStatus:True             3m5s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             3m     KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Normal   UpdateMiddleManagersNodePVCs              2m55s  KubeDB Ops-manager Operator  successfully updated middleManagers node PVC sizes
  Normal   UpdatePetSets                             2m40s  KubeDB Ops-manager Operator  successfully reconciled the Druid resources
  Warning  get pet set; ConditionStatus:True         2m35s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   ReadyPetSets                              2m35s  KubeDB Ops-manager Operator  PetSet is recreated
  Normal   Starting                                  2m35s  KubeDB Ops-manager Operator  Resuming Druid database: demo/druid-cluster
  Normal   Successful                                2m35s  KubeDB Ops-manager Operator  Successfully resumed Druid database: demo/druid-cluster for DruidOpsRequest: drops-druid-cluster-gq9huj
```

Now, we are going to verify from the `Petset`, and the `Persistent Volume` whether the volume of the topology cluster has expanded to meet the desired state, Let's check,

```bash
$ kubectl get petset -n demo druid-cluster-historicals -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"2041405440"
$ kubectl get petset -n demo druid-cluster-middleManagers -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"2041405440"
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                                             STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-2c0ef2aa-0438-4d75-9cb2-c12a176bae6a   1948Mi     RWO            Delete           Bound    demo/druid-cluster-base-task-dir-druid-cluster-middlemanagers-0   longhorn       <unset>                          19m
pvc-5f4cea5f-e0c8-4339-b67c-9cb8b02ba49d   1948Mi     RWO            Delete           Bound    demo/druid-cluster-segment-cache-druid-cluster-historicals-0      longhorn       <unset>                          19m
```

The above output verifies that we have successfully autoscaled the volume of the Druid topology cluster for both historicals and middleManagers.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete druidopsrequests -n demo drops-druid-cluster-gq9huj drops-druid-cluster-kbw4fd
kubectl delete druidutoscaler -n demo druid-storage-autoscaler
kubectl delete dr -n demo druid-cluster
```

## Next Steps

- Detail concepts of [Druid object](/docs/guides/druid/concepts/druid.md).
- Different Druid topology clustering modes [here](/docs/guides/druid/clustering/_index.md).
- Monitor your Druid database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/druid/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Druid database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/druid/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
