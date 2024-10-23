---
title: Druid Topology Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: guides-druid-volume-expansion-guide
    name: Topology
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
    - [Topology](/docs/guides/druid/clustering/topology-cluster/index.md)
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
NAME                 PROVISIONER            RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
standard (default)   kubernetes.io/gce-pd   Delete          Immediate           true                   2m49s
```

We can see from the output the `standard` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

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
    configSecret:
      name: deep-storage-config
  topology:
    historicals:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    middleManagers:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
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
NAME                   CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS     CLAIM                                                      STORAGECLASS   REASON   AGE
pvc-ccf50adf179e4162   1Gi        RWO            Delete           Bound      demo/druid-cluster-data-druid-cluster-historicals-0        standard                106s
pvc-3f177a92721440bb   1Gi        RWO            Delete           Bound      demo/druid-cluster-data-druid-cluster-middleManagers-0     standard                106s
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
    historicals: 3Gi
    middleManagers: 2Gi
    mode: Online
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `druid-cluster`.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.historicals` specifies the desired volume size for historicals node.
- `spec.volumeExpansion.middleManagers` specifies the desired volume size for middleManagers node.

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
$ kubectl describe druidopsrequest -n demo dr-volume-exp   
Name:         dr-volume-exp
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         DruidOpsRequest
Metadata:
  Creation Timestamp:  2024-07-31T04:44:17Z
  Generation:          1
  Resource Version:    149682
  UID:                 e0e19d97-7150-463c-9a7d-53eff05ea6c4
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  druid-cluster
  Type:    VolumeExpansion
  Volume Expansion:
    Broker:      3Gi
    Controller:  2Gi
    Mode:        Online
Status:
  Conditions:
    Last Transition Time:  2024-07-31T04:44:17Z
    Message:               Druid ops-request has started to expand volume of druid nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2024-07-31T04:44:25Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2024-07-31T04:44:25Z
    Message:               is petset deleted; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPetsetDeleted
    Last Transition Time:  2024-07-31T04:44:45Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2024-07-31T04:44:50Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-07-31T04:44:50Z
    Message:               is pvc patched; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPvcPatched
    Last Transition Time:  2024-07-31T04:44:55Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2024-07-31T04:45:10Z
    Message:               successfully updated middleManagers node PVC sizes
    Observed Generation:   1
    Reason:                UpdateControllerNodePVCs
    Status:                True
    Type:                  UpdateControllerNodePVCs
    Last Transition Time:  2024-07-31T04:45:35Z
    Message:               successfully updated historicals node PVC sizes
    Observed Generation:   1
    Reason:                UpdateBrokerNodePVCs
    Status:                True
    Type:                  UpdateBrokerNodePVCs
    Last Transition Time:  2024-07-31T04:45:42Z
    Message:               successfully reconciled the Druid resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-31T04:45:47Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2024-07-31T04:45:47Z
    Message:               Successfully completed volumeExpansion for druid
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                   Age   From                         Message
  ----     ------                                   ----  ----                         -------
  Normal   Starting                                 116s  KubeDB Ops-manager Operator  Start processing for DruidOpsRequest: demo/dr-volume-exp
  Normal   Starting                                 116s  KubeDB Ops-manager Operator  Pausing Druid databse: demo/druid-cluster
  Normal   Successful                               116s  KubeDB Ops-manager Operator  Successfully paused Druid database: demo/druid-cluster for DruidOpsRequest: dr-volume-exp
  Warning  get pet set; ConditionStatus:True        108s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  is petset deleted; ConditionStatus:True  108s  KubeDB Ops-manager Operator  is petset deleted; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        103s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        98s   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  is petset deleted; ConditionStatus:True  98s   KubeDB Ops-manager Operator  is petset deleted; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        93s   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   OrphanPetSetPods                         88s   KubeDB Ops-manager Operator  successfully deleted the petSets with orphan propagation policy
  Warning  get pvc; ConditionStatus:True            83s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True     83s   KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            78s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    78s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            73s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True     73s   KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            68s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    68s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Normal   UpdateControllerNodePVCs                 63s   KubeDB Ops-manager Operator  successfully updated middleManagers node PVC sizes
  Warning  get pvc; ConditionStatus:True            58s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True     58s   KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            53s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    53s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            48s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True     48s   KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            43s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    43s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Normal   UpdateBrokerNodePVCs                     38s   KubeDB Ops-manager Operator  successfully updated historicals node PVC sizes
  Normal   UpdatePetSets                            31s   KubeDB Ops-manager Operator  successfully reconciled the Druid resources
  Warning  get pet set; ConditionStatus:True        26s   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        26s   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   ReadyPetSets                             26s   KubeDB Ops-manager Operator  PetSet is recreated
  Normal   Starting                                 26s   KubeDB Ops-manager Operator  Resuming Druid database: demo/druid-cluster
  Normal   Successful                               26s   KubeDB Ops-manager Operator  Successfully resumed Druid database: demo/druid-cluster for DruidOpsRequest: dr-volume-exp
```

Now, we are going to verify from the `Petset`, and the `Persistent Volumes` whether the volume of the database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get petset -n demo druid-cluster-historicals -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"3Gi"

$ kubectl get petset -n demo druid-cluster-middleManagers -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"2Gi"

$ kubectl get pv -n demo                                       
NAME                   CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS     CLAIM                                           STORAGECLASS   REASON   AGE
pvc-3f177a92721440bb   1Gi        RWO            Delete           Bound      demo/druid-cluster-data-druid-cluster-middleManagers-0    standard                5m25s
pvc-86ff354122324b1c   1Gi        RWO            Delete           Bound      demo/druid-cluster-data-druid-cluster-historicals-1        standard                4m51s
pvc-9fa35d773aa74bd0   1Gi        RWO            Delete           Bound      demo/druid-cluster-data-druid-cluster-middleManagers-1    standard                5m1s
pvc-ccf50adf179e4162   1Gi        RWO            Delete           Bound      demo/druid-cluster-data-druid-cluster-historicals-0        standard                5m30s
```

The above output verifies that we have successfully expanded the volume of the Druid.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete druidopsrequest -n demo dr-volume-exp
kubectl delete dr -n demo druid-cluster
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Druid object](/docs/guides/druid/concepts/druid.md).
- Different Druid topology clustering modes [here](/docs/guides/druid/clustering/_index.md).
- Monitor your Druid database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/druid/monitoring/using-prometheus-operator.md).
- 
[//]: # (- Monitor your Druid database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/druid/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
