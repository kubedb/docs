---
title: Druid Topology Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: guides-druid-volume-expansion-guide
    name: Druid Volume Expansion
    parent: guides-druid-volume-expansion
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Druid Topology Volume Expansion

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of a Druid Topology Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Druid](/docs/guides/druid/concepts/druid.md)
    - [Topology](/docs/guides/druid/clustering/overview/index.md)
    - [DruidOpsRequest](/docs/guides/druid/concepts/druidopsrequest.md)
    - [Volume Expansion Overview](/docs/guides/druid/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: The yaml files used in this tutorial are stored in [docs/examples/druid](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/druid) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Expand Volume of Topology Druid Cluster

Here, we are going to deploy a `Druid` topology using a supported version by `KubeDB` operator. Then we are going to apply `DruidOpsRequest` to expand its volume.

### Prepare Druid Topology Cluster

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  28h
longhorn (default)     driver.longhorn.io      Delete          Immediate              true                   27h
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   27h
```

We can see from the output the `longhorn` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/volume-expansion/yamls/deep-storage-config.yaml
secret/deep-storage-config created
```

Now, we are going to deploy a `Druid` combined cluster with version `28.0.1`.

### Deploy Druid

In this section, we are going to deploy a Druid topology cluster for historicals and middleManagers with 1GB volume. Then, in the next section we will expand its volume to 2GB using `DruidOpsRequest` CRD. Below is the YAML of the `Druid` CR that we are going to create,

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
```

Let's create the `Druid` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/volume-expansion/yamls/druid-topology.yaml
druid.kubedb.com/druid-cluster created
```

Now, wait until `druid-cluster` has status `Ready`. i.e,

```bash
$ kubectl get dr -n demo -w
NAME            TYPE                  VERSION   STATUS         AGE
druid-cluster   kubedb.com/v1alpha2   28.0.1    Provisioning   0s
druid-cluster   kubedb.com/v1alpha2   28.0.1    Provisioning   9s
.
.
druid-cluster   kubedb.com/v1alpha2   28.0.1    Ready          3m26s
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo druid-cluster-historicals -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get petset -n demo druid-cluster-middleManagers -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo                                       
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                                             STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-0bf49077-1c7a-4943-bb17-1dffd1626dcd   1Gi        RWO            Delete           Bound    demo/druid-cluster-segment-cache-druid-cluster-historicals-0      longhorn       <unset>                          10m
pvc-59ed4914-53b3-4f18-a6aa-7699c2b738e2   1Gi        RWO            Delete           Bound    demo/druid-cluster-base-task-dir-druid-cluster-middlemanagers-0   longhorn       <unset>                          10m
```

You can see the petsets have 1GB storage, and the capacity of all the persistent volumes are also 1GB.

We are now ready to apply the `DruidOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the druid topology cluster.

#### Create DruidOpsRequest

In order to expand the volume of the database, we have to create a `DruidOpsRequest` CR with our desired volume size. Below is the YAML of the `DruidOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: dr-volume-exp
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: druid-cluster
  volumeExpansion:
    historicals: 2Gi
    middleManagers: 2Gi
    mode: Offline
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `druid-cluster`.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.historicals` specifies the desired volume size for historicals node.
- `spec.volumeExpansion.middleManagers` specifies the desired volume size for middleManagers node.
- `spec.volumeExpansion.mode` specifies the desired volume expansion mode(`Online` or `Offline`).

During `Online` VolumeExpansion KubeDB expands volume without pausing database object, it directly updates the underlying PVC. And for `Offline` volume expansion, the database is paused. The Pods are deleted and PVC is updated. Then the database Pods are recreated with updated PVC.

> If you want to expand the volume of only one node, you can specify the desired volume size for that node only.

Let's create the `DruidOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/volume-expansion/yamls/druid-volume-expansion-topology.yaml
druidopsrequest.ops.kubedb.com/dr-volume-exp created
```

#### Verify Druid Topology volume expanded successfully

If everything goes well, `KubeDB` Ops-manager operator will update the volume size of `Druid` object and related `PetSets` and `Persistent Volumes`.

Let's wait for `DruidOpsRequest` to be `Successful`.  Run the following command to watch `DruidOpsRequest` CR,

```bash
$ kubectl get druidopsrequest -n demo
NAME                     TYPE              STATUS       AGE
dr-volume-exp            VolumeExpansion   Successful   3m1s
```

We can see from the above output that the `DruidOpsRequest` has succeeded. If we describe the `DruidOpsRequest` we will get an overview of the steps that were followed to expand the volume of druid.

```bash
$  kubectl describe druidopsrequest -n demo dr-volume-exp   
Name:         dr-volume-exp
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         DruidOpsRequest
Metadata:
  Creation Timestamp:  2024-10-25T09:22:02Z
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
        f:type:
        f:volumeExpansion:
          .:
          f:historicals:
          f:middleManagers:
          f:mode:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2024-10-25T09:22:02Z
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
    Time:            2024-10-25T09:24:35Z
  Resource Version:  221378
  UID:               2407cfa7-8d3b-463e-abf7-1910249009bd
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  druid-cluster
  Type:    VolumeExpansion
  Volume Expansion:
    Historicals:      2Gi
    Middle Managers:  2Gi
    Mode:             Offline
Status:
  Conditions:
    Last Transition Time:  2024-10-25T09:22:02Z
    Message:               Druid ops-request has started to expand volume of druid nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2024-10-25T09:22:10Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2024-10-25T09:22:10Z
    Message:               is pet set deleted; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPetSetDeleted
    Last Transition Time:  2024-10-25T09:22:30Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2024-10-25T09:22:35Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2024-10-25T09:22:35Z
    Message:               is ops req patched; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsOpsReqPatched
    Last Transition Time:  2024-10-25T09:22:35Z
    Message:               create pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreatePod
    Last Transition Time:  2024-10-25T09:22:40Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-10-25T09:22:40Z
    Message:               is pvc patched; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPvcPatched
    Last Transition Time:  2024-10-25T09:23:50Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2024-10-25T09:23:00Z
    Message:               create; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  Create
    Last Transition Time:  2024-10-25T09:23:08Z
    Message:               is druid running; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  IsDruidRunning
    Last Transition Time:  2024-10-25T09:23:20Z
    Message:               successfully updated middleManagers node PVC sizes
    Observed Generation:   1
    Reason:                UpdateMiddleManagersNodePVCs
    Status:                True
    Type:                  UpdateMiddleManagersNodePVCs
    Last Transition Time:  2024-10-25T09:24:15Z
    Message:               successfully updated historicals node PVC sizes
    Observed Generation:   1
    Reason:                UpdateHistoricalsNodePVCs
    Status:                True
    Type:                  UpdateHistoricalsNodePVCs
    Last Transition Time:  2024-10-25T09:24:30Z
    Message:               successfully reconciled the Druid resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-25T09:24:35Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2024-10-25T09:24:35Z
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
  Normal   Starting                                  10m    KubeDB Ops-manager Operator  Start processing for DruidOpsRequest: demo/dr-volume-exp
  Normal   Starting                                  10m    KubeDB Ops-manager Operator  Pausing Druid databse: demo/druid-cluster
  Normal   Successful                                10m    KubeDB Ops-manager Operator  Successfully paused Druid database: demo/druid-cluster for DruidOpsRequest: dr-volume-exp
  Warning  get pet set; ConditionStatus:True         10m    KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  is pet set deleted; ConditionStatus:True  10m    KubeDB Ops-manager Operator  is pet set deleted; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True         10m    KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True         10m    KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  is pet set deleted; ConditionStatus:True  10m    KubeDB Ops-manager Operator  is pet set deleted; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True         10m    KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   OrphanPetSetPods                          9m59s  KubeDB Ops-manager Operator  successfully deleted the petSets with orphan propagation policy
  Warning  get pod; ConditionStatus:True             9m54s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  is ops req patched; ConditionStatus:True  9m54s  KubeDB Ops-manager Operator  is ops req patched; ConditionStatus:True
  Warning  create pod; ConditionStatus:True          9m54s  KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             9m49s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             9m49s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True      9m49s  KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False    9m49s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True             9m44s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             9m44s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             9m39s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             9m39s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             9m34s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             9m34s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             9m29s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             9m29s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True     9m29s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create; ConditionStatus:True              9m29s  KubeDB Ops-manager Operator  create; ConditionStatus:True
  Warning  is ops req patched; ConditionStatus:True  9m29s  KubeDB Ops-manager Operator  is ops req patched; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             9m24s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  is druid running; ConditionStatus:False   9m21s  KubeDB Ops-manager Operator  is druid running; ConditionStatus:False
  Warning  get pod; ConditionStatus:True             9m19s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             9m14s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Normal   UpdateMiddleManagersNodePVCs              9m9s   KubeDB Ops-manager Operator  successfully updated middleManagers node PVC sizes
  Warning  get pod; ConditionStatus:True             9m4s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  is ops req patched; ConditionStatus:True  9m4s   KubeDB Ops-manager Operator  is ops req patched; ConditionStatus:True
  Warning  create pod; ConditionStatus:True          9m4s   KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             8m59s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             8m59s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True      8m59s  KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False    8m59s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True             8m54s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc:
