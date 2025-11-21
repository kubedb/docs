---
title: Horizontal Scaling Topology Elasticsearch
menu:
  docs_{{ .version }}:
    identifier: es-horizontal-scaling-Topology
    name: Topology Cluster
    parent: es-horizontal-scalling-elasticsearch
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Elasticsearch Topology Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to scale the Elasticsearch Topology cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
    - [Topology](/docs/guides/elasticsearch/clustering/Topology-cluster/index.md)
    - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)
    - [Horizontal Scaling Overview](/docs/guides/elasticsearch/scaling/horizontal/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/Elasticsearch](/docs/examples/elasticsearch) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on Topology Cluster

Here, we are going to deploy a  `Elasticsearch` Topology cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare Elasticsearch Topology cluster

Now, we are going to deploy a `Elasticsearch` Topology cluster with version `xpack-8.11.1`.

### Deploy Elasticsearch Topology cluster

In this section, we are going to deploy a Elasticsearch Topology cluster. Then, in the next section we will scale the cluster using `ElasticsearchOpsRequest` CRD. Below is the YAML of the `Elasticsearch` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-hscale-topology
  namespace: demo
spec:
  enableSSL: true
  version: xpack-8.11.1
  storageType: Durable
  topology:
    master:
      replicas: 3
      storage:
        storageClassName: "standard"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      replicas: 3
      storage:
        storageClassName: "standard"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    ingest:
      replicas: 3
      storage:
        storageClassName: "standard"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
```

Let's create the `Elasticsearch` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/clustering/topology.yaml
Elasticsearch.kubedb.com/es-hscale-topology created
```

Now, wait until `es-hscale-topology` has status `Ready`. i.e,

```bash
$ kubectl get es -n demo
NAME                    VERSION       STATUS   AGE
es-hscale-topology     xpack-8.11.1   Ready    3m53s
```

Let's check the number of replicas has from Elasticsearch object, number of pods the petset have,

```bash
$ kubectl get elasticsearch -n demo es-hscale-topology -o json | jq '.spec.topology.master.replicas'
3
$ kubectl get elasticsearch -n demo es-hscale-topology -o json | jq '.spec.topology.ingest.replicas'
3
$ kubectl get elasticsearch -n demo es-hscale-topology -o json | jq '.spec.topology.data.replicas'
3
```

We can see from both command that the cluster has 3 replicas.

Also, we can verify the replicas of the Topology from an internal Elasticsearch command by exec into a replica.

Now lets check the number of replicas,

```bash
$ kubectl get all,secret,pvc -n demo -l 'app.kubernetes.io/instance=es-hscale-topology'
NAME                              READY   STATUS    RESTARTS   AGE
pod/es-hscale-topology-data-0     1/1     Running   0          27m
pod/es-hscale-topology-data-1     1/1     Running   0          25m
pod/es-hscale-topology-data-2     1/1     Running   0          24m
pod/es-hscale-topology-ingest-0   1/1     Running   0          27m
pod/es-hscale-topology-ingest-1   1/1     Running   0          25m
pod/es-hscale-topology-ingest-2   1/1     Running   0          24m
pod/es-hscale-topology-master-0   1/1     Running   0          27m
pod/es-hscale-topology-master-1   1/1     Running   0          25m
pod/es-hscale-topology-master-2   1/1     Running   0          24m

NAME                                TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/es-hscale-topology          ClusterIP   10.43.33.118   <none>        9200/TCP   27m
service/es-hscale-topology-master   ClusterIP   None           <none>        9300/TCP   27m
service/es-hscale-topology-pods     ClusterIP   None           <none>        9200/TCP   27m

NAME                                                    TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/es-hscale-topology   kubedb.com/elasticsearch   8.11.1    27m

NAME                                                    TYPE                       DATA   AGE
secret/es-hscale-topology-apm-system-cred               kubernetes.io/basic-auth   2      27m
secret/es-hscale-topology-auth                          kubernetes.io/basic-auth   2      27m
secret/es-hscale-topology-beats-system-cred             kubernetes.io/basic-auth   2      27m
secret/es-hscale-topology-ca-cert                       kubernetes.io/tls          2      27m
secret/es-hscale-topology-client-cert                   kubernetes.io/tls          3      27m
secret/es-hscale-topology-config                        Opaque                     1      27m
secret/es-hscale-topology-http-cert                     kubernetes.io/tls          3      27m
secret/es-hscale-topology-kibana-system-cred            kubernetes.io/basic-auth   2      27m
secret/es-hscale-topology-logstash-system-cred          kubernetes.io/basic-auth   2      27m
secret/es-hscale-topology-remote-monitoring-user-cred   kubernetes.io/basic-auth   2      27m
secret/es-hscale-topology-transport-cert                kubernetes.io/tls          3      27m

NAME                                                     STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
persistentvolumeclaim/data-es-hscale-topology-data-0     Bound    pvc-ce9ce1ec-a2db-43c8-9d40-d158f53f25fe   1Gi        RWO            standard     <unset>                 27m
persistentvolumeclaim/data-es-hscale-topology-data-1     Bound    pvc-babfc22c-1e29-44e3-a094-8fa48876db68   1Gi        RWO            standard     <unset>                 25m
persistentvolumeclaim/data-es-hscale-topology-data-2     Bound    pvc-c0e64663-1cc4-420c-85b9-4f643c76f006   1Gi        RWO            standard     <unset>                 24m
persistentvolumeclaim/data-es-hscale-topology-ingest-0   Bound    pvc-3de6c8f6-17aa-43d8-8c10-8cbd2dc543aa   1Gi        RWO            standard     <unset>                 27m
persistentvolumeclaim/data-es-hscale-topology-ingest-1   Bound    pvc-d990c570-c687-4192-ad2e-bad127b7b5db   1Gi        RWO            standard     <unset>                 25m
persistentvolumeclaim/data-es-hscale-topology-ingest-2   Bound    pvc-4540c342-811a-4b82-970e-0e6d29e80e9b   1Gi        RWO            standard     <unset>                 24m
persistentvolumeclaim/data-es-hscale-topology-master-0   Bound    pvc-902a0ebb-b6fb-4106-8220-f137972a84be   1Gi        RWO            standard     <unset>                 27m
persistentvolumeclaim/data-es-hscale-topology-master-1   Bound    pvc-f97215e6-1a91-4e77-8bfb-78d907828e51   1Gi        RWO            standard     <unset>                 25m
persistentvolumeclaim/data-es-hscale-topology-master-2   Bound    pvc-a9160094-c08e-4d40-b4ea-ec5681f8be30   1Gi        RWO            standard     <unset>                 24m

```

We can see from the above output that the Elasticsearch has 2 nodes.

We are now ready to apply the `ElasticsearchOpsRequest` CR to scale this cluster.


### Scale Down Replicas

Here, we are going to scale down the replicas of the Elasticsearch Topology cluster to meet the desired number of replicas after scaling.

#### Create ElasticsearchOpsRequest

In order to scale down the replicas of the Elasticsearch Topology cluster, we have to create a `ElasticsearchOpsRequest` CR with our desired replicas. Below is the YAML of the `ElasticsearchOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: esops-hscale-down-topology
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: es-hscale-topology
  horizontalScaling:
    topology:
      master: 2
      ingest: 2
      data: 2
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down operation on `es-hscale-topology` cluster.
- `spec.type` specifies that we are performing `HorizontalScaling` on Elasticsearch.
- `verticalScaling.topology` - specifies the desired node resources for different type of node of the Elasticsearch running in cluster topology mode (ie. `Elasticsearch.spec.topology` is `not empty`).
  - `topology.master` - specifies the desired resources for the master nodes. It takes input same as the k8s [resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-types).
  - `topology.data` - specifies the desired node resources for the data nodes. It takes input  same as the k8s [resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-types).
  - `topology.ingest` - specifies the desired node resources for the ingest nodes. It takes input  same as the k8s [resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-types).

> Note: It is recommended not to use resources below the default one; `cpu: 500m, memory: 1Gi`.


Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/scaling/horizontal/Elasticsearch-hscale-down-Topology.yaml
Elasticsearchopsrequest.ops.kubedb.com/esops-hscale-down-topology created
```

#### Verify Topology cluster replicas scaled down successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Elasticsearch` object and related `PetSets` and `Pods`.

Let's wait for `ElasticsearchOpsRequest` to be `Successful`. Run the following command to watch `ElasticsearchOpsRequest` CR,

```bash
$ kubectl get Elasticsearchopsrequest -n demo
NAME                         TYPE                STATUS       AGE
esops-hscale-down-Topology   HorizontalScaling   Successful   76s
```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe Elasticsearchopsrequests -n demo esops-hscale-down-topology
Name:         esops-hscale-down-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2025-11-17T12:01:29Z
  Generation:          1
  Resource Version:    11617
  UID:                 4b4f9728-b31e-4336-a95c-cf34d97d8b4a
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  es-hscale-topology
  Horizontal Scaling:
    Topology:
      Data:    2
      Ingest:  2
      Master:  2
  Type:        HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2025-11-17T12:01:29Z
    Message:               Elasticsearch ops request is horizontally scaling the nodes.
    Observed Generation:   1
    Reason:                HorizontalScale
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2025-11-17T12:01:37Z
    Message:               create es client; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient
    Last Transition Time:  2025-11-17T12:01:37Z
    Message:               patch pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetSet
    Last Transition Time:  2025-11-17T12:01:42Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2025-11-17T12:01:42Z
    Message:               delete pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePvc
    Last Transition Time:  2025-11-17T12:02:27Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2025-11-17T12:01:52Z
    Message:               ScaleDown es-hscale-topology-ingest nodes
    Observed Generation:   1
    Reason:                HorizontalScaleIngestNode
    Status:                True
    Type:                  HorizontalScaleIngestNode
    Last Transition Time:  2025-11-17T12:01:57Z
    Message:               exclude node allocation; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ExcludeNodeAllocation
    Last Transition Time:  2025-11-17T12:01:57Z
    Message:               get used data nodes; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetUsedDataNodes
    Last Transition Time:  2025-11-17T12:01:57Z
    Message:               move data; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  MoveData
    Last Transition Time:  2025-11-17T12:02:12Z
    Message:               delete node allocation exclusion; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeleteNodeAllocationExclusion
    Last Transition Time:  2025-11-17T12:02:12Z
    Message:               ScaleDown es-hscale-topology-data nodes
    Observed Generation:   1
    Reason:                HorizontalScaleDataNode
    Status:                True
    Type:                  HorizontalScaleDataNode
    Last Transition Time:  2025-11-17T12:02:18Z
    Message:               get voting config exclusion; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetVotingConfigExclusion
    Last Transition Time:  2025-11-17T12:02:32Z
    Message:               delete voting config exclusion; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeleteVotingConfigExclusion
    Last Transition Time:  2025-11-17T12:02:32Z
    Message:               ScaleDown es-hscale-topology-master nodes
    Observed Generation:   1
    Reason:                HorizontalScaleMasterNode
    Status:                True
    Type:                  HorizontalScaleMasterNode
    Last Transition Time:  2025-11-17T12:02:37Z
    Message:               successfully updated Elasticsearch CR
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2025-11-17T12:02:38Z
    Message:               Successfully Horizontally Scaled.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                  Age   From                         Message
  ----     ------                                                  ----  ----                         -------
  Normal   PauseDatabase                                           101s  KubeDB Ops-manager Operator  Pausing Elasticsearch demo/es-hscale-topology
  Warning  create es client; ConditionStatus:True                  93s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  patch pet set; ConditionStatus:True                     93s   KubeDB Ops-manager Operator  patch pet set; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                           88s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True                        88s   KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:False                          88s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:False
  Warning  get pod; ConditionStatus:True                           83s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True                        83s   KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                           83s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  create es client; ConditionStatus:True                  78s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Normal   HorizontalScaleIngestNode                               78s   KubeDB Ops-manager Operator  ScaleDown es-hscale-topology-ingest nodes
  Warning  create es client; ConditionStatus:True                  73s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  exclude node allocation; ConditionStatus:True           73s   KubeDB Ops-manager Operator  exclude node allocation; ConditionStatus:True
  Warning  get used data nodes; ConditionStatus:True               73s   KubeDB Ops-manager Operator  get used data nodes; ConditionStatus:True
  Warning  move data; ConditionStatus:True                         73s   KubeDB Ops-manager Operator  move data; ConditionStatus:True
  Warning  patch pet set; ConditionStatus:True                     73s   KubeDB Ops-manager Operator  patch pet set; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                           68s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True                        68s   KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:False                          68s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:False
  Warning  get pod; ConditionStatus:True                           63s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True                        63s   KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                           63s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  create es client; ConditionStatus:True                  58s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  delete node allocation exclusion; ConditionStatus:True  58s   KubeDB Ops-manager Operator  delete node allocation exclusion; ConditionStatus:True
  Normal   HorizontalScaleDataNode                                 58s   KubeDB Ops-manager Operator  ScaleDown es-hscale-topology-data nodes
  Warning  create es client; ConditionStatus:True                  53s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  get voting config exclusion; ConditionStatus:True       52s   KubeDB Ops-manager Operator  get voting config exclusion; ConditionStatus:True
  Warning  patch pet set; ConditionStatus:True                     52s   KubeDB Ops-manager Operator  patch pet set; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                           48s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True                        48s   KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:False                          48s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:False
  Warning  get pod; ConditionStatus:True                           43s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True                        43s   KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                           43s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  create es client; ConditionStatus:True                  38s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  delete voting config exclusion; ConditionStatus:True    38s   KubeDB Ops-manager Operator  delete voting config exclusion; ConditionStatus:True
  Normal   HorizontalScaleMasterNode                               38s   KubeDB Ops-manager Operator  ScaleDown es-hscale-topology-master nodes
  Normal   UpdateDatabase                                          33s   KubeDB Ops-manager Operator  successfully updated Elasticsearch CR
  Normal   ResumeDatabase                                          33s   KubeDB Ops-manager Operator  Resuming Elasticsearch demo/es-hscale-topology
  Normal   ResumeDatabase                                          33s   KubeDB Ops-manager Operator  Successfully resumed Elasticsearch demo/es-hscale-topology
  Normal   Successful                                              33s   KubeDB Ops-manager Operator  Successfully Horizontally Scaled Database
```

Now, we are going to verify the number of replicas this cluster has from the Elasticsearch object, number of pods the petset have,

```bash
$ kubectl get elasticsearch -n demo es-hscale-topology -o json | jq '.spec.topology.master.replicas'
2
$ kubectl get elasticsearch -n demo es-hscale-topology -o json | jq '.spec.topology.data.replicas'
2
$ kubectl get elasticsearch -n demo es-hscale-topology -o json | jq '.spec.topology.ingest.replicas'
2
```
**Only ingest nodes after scaling down:**
```bash
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: esops-ingest-hscale-down-topology
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: es-hscale-topology
  horizontalScaling:
    topology:
      ingest: 2
```
From all the above outputs we can see that the replicas of the Topology cluster is `2`. That means we have successfully scaled down the replicas of the Elasticsearch Topology cluster.



## Scale Up Replicas

Here, we are going to scale up the replicas of the Topology cluster to meet the desired number of replicas after scaling.

#### Create ElasticsearchOpsRequest

In order to scale up the replicas of the Topology cluster, we have to create a `ElasticsearchOpsRequest` CR with our desired replicas. Below is the YAML of the `ElasticsearchOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: esops-hscale-up-topology
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: es-hscale-topology
  horizontalScaling:
    topology:
      master: 3
      ingest: 3
      data: 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `es-hscale-topology` cluster.
- `spec.type` specifies that we are performing `HorizontalScaling` on Elasticsearch.
- `verticalScaling.topology` - specifies the desired node resources for different type of node of the Elasticsearch running in cluster topology mode (ie. `Elasticsearch.spec.topology` is `not empty`).
    - `topology.master` - specifies the desired resources for the master nodes. It takes input same as the k8s [resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-types).
    - `topology.data` - specifies the desired node resources for the data nodes. It takes input  same as the k8s [resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-types).
    - `topology.ingest` - specifies the desired node resources for the ingest nodes. It takes input  same as the k8s [resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-types).

> Note: It is recommended not to use resources below the default one; `cpu: 500m, memory: 1Gi`.

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/scaling/horizontal/Elasticsearch-hscale-up-Topology.yaml
Elasticsearchopsrequest.ops.kubedb.com/esops-hscale-up-topology created
```

#### Verify Topology cluster replicas scaled up successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Elasticsearch` object and related `PetSets` and `Pods`.

Let's wait for `ElasticsearchOpsRequest` to be `Successful`. Run the following command to watch `ElasticsearchOpsRequest` CR,

```bash
$  kubectl get Elasticsearchopsrequest -n demo
NAME                       TYPE                STATUS       AGE
esops-hscale-up-topology   HorizontalScaling   Successful   13m
```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe Elasticsearchopsrequests -n demo esops-hscale-up-topology
Name:         esops-hscale-up-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2025-11-17T12:12:44Z
  Generation:          1
  Resource Version:    12241
  UID:                 5342e779-62bc-4fe1-b91c-21b30c30cd39
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  es-hscale-topology
  Horizontal Scaling:
    Topology:
      Data:    3
      Ingest:  3
      Master:  3
  Type:        HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2025-11-17T12:12:44Z
    Message:               Elasticsearch ops request is horizontally scaling the nodes.
    Observed Generation:   1
    Reason:                HorizontalScale
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2025-11-17T12:12:52Z
    Message:               patch pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetSet
    Last Transition Time:  2025-11-17T12:13:58Z
    Message:               is node in cluster; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsNodeInCluster
    Last Transition Time:  2025-11-17T12:13:12Z
    Message:               ScaleUp es-hscale-topology-ingest nodes
    Observed Generation:   1
    Reason:                HorizontalScaleIngestNode
    Status:                True
    Type:                  HorizontalScaleIngestNode
    Last Transition Time:  2025-11-17T12:13:37Z
    Message:               ScaleUp es-hscale-topology-data nodes
    Observed Generation:   1
    Reason:                HorizontalScaleDataNode
    Status:                True
    Type:                  HorizontalScaleDataNode
    Last Transition Time:  2025-11-17T12:14:02Z
    Message:               ScaleUp es-hscale-topology-master nodes
    Observed Generation:   1
    Reason:                HorizontalScaleMasterNode
    Status:                True
    Type:                  HorizontalScaleMasterNode
    Last Transition Time:  2025-11-17T12:14:07Z
    Message:               successfully updated Elasticsearch CR
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2025-11-17T12:14:08Z
    Message:               Successfully Horizontally Scaled.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                     Age    From                         Message
  ----     ------                                     ----   ----                         -------
  Normal   PauseDatabase                              6m15s  KubeDB Ops-manager Operator  Pausing Elasticsearch demo/es-hscale-topology
  Warning  patch pet set; ConditionStatus:True        6m7s   KubeDB Ops-manager Operator  patch pet set; ConditionStatus:True
  Warning  is node in cluster; ConditionStatus:False  6m2s   KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:False
  Warning  is node in cluster; ConditionStatus:True   5m52s  KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:True
  Normal   HorizontalScaleIngestNode                  5m47s  KubeDB Ops-manager Operator  ScaleUp es-hscale-topology-ingest nodes
  Warning  patch pet set; ConditionStatus:True        5m42s  KubeDB Ops-manager Operator  patch pet set; ConditionStatus:True
  Warning  is node in cluster; ConditionStatus:False  5m37s  KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:False
  Warning  is node in cluster; ConditionStatus:True   5m27s  KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:True
  Normal   HorizontalScaleDataNode                    5m22s  KubeDB Ops-manager Operator  ScaleUp es-hscale-topology-data nodes
  Warning  patch pet set; ConditionStatus:True        5m17s  KubeDB Ops-manager Operator  patch pet set; ConditionStatus:True
  Warning  is node in cluster; ConditionStatus:False  5m12s  KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:False
  Warning  is node in cluster; ConditionStatus:True   5m1s   KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:True
  Normal   HorizontalScaleMasterNode                  4m57s  KubeDB Ops-manager Operator  ScaleUp es-hscale-topology-master nodes
  Normal   UpdateDatabase                             4m52s  KubeDB Ops-manager Operator  successfully updated Elasticsearch CR
  Normal   ResumeDatabase                             4m52s  KubeDB Ops-manager Operator  Resuming Elasticsearch demo/es-hscale-topology
  Normal   ResumeDatabase                             4m52s  KubeDB Ops-manager Operator  Successfully resumed Elasticsearch demo/es-hscale-topology
  Normal   Successful                                 4m51s  KubeDB Ops-manager Operator  Successfully Horizontally Scaled Database
```

Now, we are going to verify the number of replicas this cluster has from the Elasticsearch object, number of pods the petset have,

```bash
$ kubectl get elasticsearch -n demo es-hscale-topology -o json | jq '.spec.topology.master.replicas'
3
$ kubectl get elasticsearch -n demo es-hscale-topology -o json | jq '.spec.topology.data.replicas'
3
$ kubectl get elasticsearch -n demo es-hscale-topology -o json | jq '.spec.topology.ingest.replicas'
3
```

From all the above outputs we can see that the brokers of the Topology Elasticsearch is `3`. That means we have successfully scaled up the replicas of the Elasticsearch Topology cluster.


**Only ingest nodes after scaling up:**
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: esops-ingest-hscale-up-topology
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: es-hscale-topology
  horizontalScaling:
    topology:
      ingest: 3
```


## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete es -n demo es-hscale-topology
kubectl delete Elasticsearchopsrequest -n demo esops-hscale-down-topology,esops-hscale-up-topology,esops-ingest-hscale-up-topology,esops-ingest-hscale-down-topology
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Different Elasticsearch topology clustering modes [here](/docs/guides/elasticsearch/clustering/_index.md).
- Monitor your Elasticsearch with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
