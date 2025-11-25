---
title: Horizontal Scaling Combined Elasticsearch
menu:
  docs_{{ .version }}:
    identifier: es-horizontal-scaling-combined
    name: Combined Cluster
    parent: es-horizontal-scalling-elasticsearch
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Elasticsearch Combined Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to scale the Elasticsearch combined cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
    - [Combined](/docs/guides/elasticsearch/clustering/combined-cluster/index.md)
    - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)
    - [Horizontal Scaling Overview](/docs/guides/elasticsearch/scaling/horizontal/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/Elasticsearch](/docs/examples/elasticsearch) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on Combined Cluster

Here, we are going to deploy a  `Elasticsearch` combined cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare Elasticsearch Combined cluster

Now, we are going to deploy a `Elasticsearch` combined cluster with version `xpack-9.1.4`.

### Deploy Elasticsearch combined cluster

In this section, we are going to deploy a Elasticsearch combined cluster. Then, in the next section we will scale the cluster using `ElasticsearchOpsRequest` CRD. Below is the YAML of the `Elasticsearch` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es
  namespace: demo
spec:
  version: xpack-9.1.4
  enableSSL: true
  replicas: 2
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Elasticsearch` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/overview/quickstart/elasticsearch/yamls/elasticsearch-v1.yaml
Elasticsearch.kubedb.com/es created
```

Now, wait until `es` has status `Ready`. i.e,

```bash
$ kubectl get es -n demo
NAME   VERSION       STATUS   AGE
es     xpack-9.1.4   Ready    3m53s
```

Let's check the number of replicas has from Elasticsearch object, number of pods the petset have,

```bash
$ kubectl get elasticsearch -n demo es -o json | jq '.spec.replicas'
2
$ kubectl get petsets -n demo es -o json | jq '.spec.replicas'
2

```

We can see from both command that the cluster has 2 replicas.

Also, we can verify the replicas of the combined from an internal Elasticsearch command by exec into a replica.

Now lets check the number of replicas,

```bash
$ kubectl get all,secret,pvc -n demo -l 'app.kubernetes.io/instance=es'
NAME       READY   STATUS    RESTARTS   AGE
pod/es-0   1/1     Running   0          5m
pod/es-1   1/1     Running   0          4m54s

NAME                TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/es          ClusterIP   10.43.72.228   <none>        9200/TCP   5m5s
service/es-master   ClusterIP   None           <none>        9300/TCP   5m5s
service/es-pods     ClusterIP   None           <none>        9200/TCP   5m5s

NAME                                    TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/es   kubedb.com/elasticsearch   9.1.4     5m2s

NAME                                    TYPE                       DATA   AGE
secret/es-apm-system-cred               kubernetes.io/basic-auth   2      5m4s
secret/es-auth                          kubernetes.io/basic-auth   2      5m8s
secret/es-beats-system-cred             kubernetes.io/basic-auth   2      5m4s
secret/es-ca-cert                       kubernetes.io/tls          2      5m9s
secret/es-client-cert                   kubernetes.io/tls          3      5m8s
secret/es-config                        Opaque                     1      5m8s
secret/es-http-cert                     kubernetes.io/tls          3      5m8s
secret/es-kibana-system-cred            kubernetes.io/basic-auth   2      5m4s
secret/es-logstash-system-cred          kubernetes.io/basic-auth   2      5m4s
secret/es-remote-monitoring-user-cred   kubernetes.io/basic-auth   2      5m4s
secret/es-transport-cert                kubernetes.io/tls          3      5m8s

NAME                              STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
persistentvolumeclaim/data-es-0   Bound    pvc-7c8cc17d-7427-4411-9262-f213e826540b   1Gi        RWO            standard     <unset>                 5m5s
persistentvolumeclaim/data-es-1   Bound    pvc-f2cf7ac9-b0c2-4c44-93dc-476cc06c25b4   1Gi        RWO            standard     <unset>                 4m59s

```

We can see from the above output that the Elasticsearch has 2 nodes.

We are now ready to apply the `ElasticsearchOpsRequest` CR to scale this cluster.

## Scale Up Replicas

Here, we are going to scale up the replicas of the combined cluster to meet the desired number of replicas after scaling.

#### Create ElasticsearchOpsRequest

In order to scale up the replicas of the combined cluster, we have to create a `ElasticsearchOpsRequest` CR with our desired replicas. Below is the YAML of the `ElasticsearchOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: esops-hscale-up-combined
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: es
  horizontalScaling:
    node: 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `es` cluster.
- `spec.type` specifies that we are performing `HorizontalScaling` on Elasticsearch.
- `spec.horizontalScaling.node` specifies the desired replicas after scaling.

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/scaling/horizontal/Elasticsearch-hscale-up-combined.yaml
Elasticsearchopsrequest.ops.kubedb.com/esops-hscale-up-combined created
```

#### Verify Combined cluster replicas scaled up successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Elasticsearch` object and related `PetSets` and `Pods`.

Let's wait for `ElasticsearchOpsRequest` to be `Successful`. Run the following command to watch `ElasticsearchOpsRequest` CR,

```bash
$ kubectl get Elasticsearchopsrequest -n demo
NAME          TYPE                STATUS       AGE
esops-hscale-up-combined   HorizontalScaling   Successful   2m42s
```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe Elasticsearchopsrequests -n demo esops-hscale-up-combined
Name:         esops-hscale-up-combined
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2025-11-13T10:25:18Z
  Generation:          1
  Resource Version:    810747
  UID:                 29134aef-1379-4e4f-91c8-23b1cf74c784
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  es
  Horizontal Scaling:
    Node:  3
  Type:    HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2025-11-13T10:25:58Z
    Message:               Elasticsearch ops request is horizontally scaling the nodes.
    Observed Generation:   1
    Reason:                HorizontalScale
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2025-11-13T10:26:06Z
    Message:               patch pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetSet
    Last Transition Time:  2025-11-13T10:26:26Z
    Message:               is node in cluster; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsNodeInCluster
    Last Transition Time:  2025-11-13T10:26:31Z
    Message:               ScaleUp es nodes
    Observed Generation:   1
    Reason:                HorizontalScaleCombinedNode
    Status:                True
    Type:                  HorizontalScaleCombinedNode
    Last Transition Time:  2025-11-13T10:26:36Z
    Message:               successfully updated Elasticsearch CR
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2025-11-13T10:26:36Z
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
  Normal   PauseDatabase                              2m54s  KubeDB Ops-manager Operator  Pausing Elasticsearch demo/es
  Warning  patch pet set; ConditionStatus:True        2m46s  KubeDB Ops-manager Operator  patch pet set; ConditionStatus:True
  Warning  is node in cluster; ConditionStatus:False  2m41s  KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:False
  Warning  is node in cluster; ConditionStatus:True   2m26s  KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:True
  Normal   HorizontalScaleCombinedNode                2m21s  KubeDB Ops-manager Operator  ScaleUp es nodes
  Normal   UpdateDatabase                             2m16s  KubeDB Ops-manager Operator  successfully updated Elasticsearch CR
  Normal   ResumeDatabase                             2m16s  KubeDB Ops-manager Operator  Resuming Elasticsearch demo/es
  Normal   ResumeDatabase                             2m16s  KubeDB Ops-manager Operator  Successfully resumed Elasticsearch demo/es
  Normal   Successful                                 2m16s  KubeDB Ops-manager Operator  Successfully Horizontally Scaled Database
bonusree@bonusree-HP-ProBook-450-G4 ~> 

```

Now, we are going to verify the number of replicas this cluster has from the Elasticsearch object, number of pods the petset have,

```bash
$ kubectl get Elasticsearch -n demo es -o json | jq '.spec.replicas'
3

$ kubectl get petset -n demo es -o json | jq '.spec.replicas'
3
```



From all the above outputs we can see that the brokers of the combined Elasticsearch is `3`. That means we have successfully scaled up the replicas of the Elasticsearch combined cluster.

### Scale Down Replicas

Here, we are going to scale down the replicas of the Elasticsearch combined cluster to meet the desired number of replicas after scaling.

#### Create ElasticsearchOpsRequest

In order to scale down the replicas of the Elasticsearch combined cluster, we have to create a `ElasticsearchOpsRequest` CR with our desired replicas. Below is the YAML of the `ElasticsearchOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: esops-hscale-down-combined
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: es
  horizontalScaling:
    node: 2
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down operation on `es` cluster.
- `spec.type` specifies that we are performing `HorizontalScaling` on Elasticsearch.
- `spec.horizontalScaling.node` specifies the desired replicas after scaling.

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/scaling/horizontal/Elasticsearch-hscale-down-combined.yaml
Elasticsearchopsrequest.ops.kubedb.com/esops-hscale-down-combined created
```

#### Verify Combined cluster replicas scaled down successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Elasticsearch` object and related `PetSets` and `Pods`.

Let's wait for `ElasticsearchOpsRequest` to be `Successful`. Run the following command to watch `ElasticsearchOpsRequest` CR,

```bash
$ kubectl get Elasticsearchopsrequest -n demo
NAME                         TYPE                STATUS       AGE
esops-hscale-down-combined   HorizontalScaling   Successful   76s
```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$  kubectl describe Elasticsearchopsrequests -n demo esops-hscale-down-combined
Name:         esops-hscale-down-combined
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2025-11-13T10:46:22Z
  Generation:          1
  Resource Version:    811301
  UID:                 558530d7-5d02-4757-b459-476129b411d6
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  es
  Horizontal Scaling:
    Node:  2
  Type:    HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2025-11-13T10:46:22Z
    Message:               Elasticsearch ops request is horizontally scaling the nodes.
    Observed Generation:   1
    Reason:                HorizontalScale
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2025-11-13T10:46:30Z
    Message:               create es client; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient
    Last Transition Time:  2025-11-13T10:46:30Z
    Message:               get voting config exclusion; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetVotingConfigExclusion
    Last Transition Time:  2025-11-13T10:46:31Z
    Message:               exclude node allocation; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ExcludeNodeAllocation
    Last Transition Time:  2025-11-13T10:46:31Z
    Message:               get used data nodes; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetUsedDataNodes
    Last Transition Time:  2025-11-13T10:46:31Z
    Message:               move data; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  MoveData
    Last Transition Time:  2025-11-13T10:46:31Z
    Message:               patch pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetSet
    Last Transition Time:  2025-11-13T10:46:35Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2025-11-13T10:46:35Z
    Message:               delete pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePvc
    Last Transition Time:  2025-11-13T10:46:40Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2025-11-13T10:46:45Z
    Message:               delete voting config exclusion; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeleteVotingConfigExclusion
    Last Transition Time:  2025-11-13T10:46:45Z
    Message:               delete node allocation exclusion; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeleteNodeAllocationExclusion
    Last Transition Time:  2025-11-13T10:46:45Z
    Message:               ScaleDown es nodes
    Observed Generation:   1
    Reason:                HorizontalScaleCombinedNode
    Status:                True
    Type:                  HorizontalScaleCombinedNode
    Last Transition Time:  2025-11-13T10:46:51Z
    Message:               successfully updated Elasticsearch CR
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2025-11-13T10:46:51Z
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
  Normal   PauseDatabase                                           112s  KubeDB Ops-manager Operator  Pausing Elasticsearch demo/es
  Warning  create es client; ConditionStatus:True                  104s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  get voting config exclusion; ConditionStatus:True       104s  KubeDB Ops-manager Operator  get voting config exclusion; ConditionStatus:True
  Warning  exclude node allocation; ConditionStatus:True           103s  KubeDB Ops-manager Operator  exclude node allocation; ConditionStatus:True
  Warning  get used data nodes; ConditionStatus:True               103s  KubeDB Ops-manager Operator  get used data nodes; ConditionStatus:True
  Warning  move data; ConditionStatus:True                         103s  KubeDB Ops-manager Operator  move data; ConditionStatus:True
  Warning  patch pet set; ConditionStatus:True                     103s  KubeDB Ops-manager Operator  patch pet set; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                           99s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True                        99s   KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:False                          99s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:False
  Warning  get pod; ConditionStatus:True                           94s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True                        94s   KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                           94s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  create es client; ConditionStatus:True                  89s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  delete voting config exclusion; ConditionStatus:True    89s   KubeDB Ops-manager Operator  delete voting config exclusion; ConditionStatus:True
  Warning  delete node allocation exclusion; ConditionStatus:True  89s   KubeDB Ops-manager Operator  delete node allocation exclusion; ConditionStatus:True
  Normal   HorizontalScaleCombinedNode                             89s   KubeDB Ops-manager Operator  ScaleDown es nodes
  Normal   UpdateDatabase                                          83s   KubeDB Ops-manager Operator  successfully updated Elasticsearch CR
  Normal   ResumeDatabase                                          83s   KubeDB Ops-manager Operator  Resuming Elasticsearch demo/es
  Normal   ResumeDatabase                                          83s   KubeDB Ops-manager Operator  Successfully resumed Elasticsearch demo/es
  Normal   Successful                                              83s   KubeDB Ops-manager Operator  Successfully Horizontally Scaled Database
```

Now, we are going to verify the number of replicas this cluster has from the Elasticsearch object, number of pods the petset have,

```bash
$ kubectl get Elasticsearch -n demo es -o json | jq '.spec.replicas' 
2

$ kubectl get petset -n demo es -o json | jq '.spec.replicas'
2
```


From all the above outputs we can see that the replicas of the combined cluster is `2`. That means we have successfully scaled down the replicas of the Elasticsearch combined cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete es -n demo es
kubectl delete Elasticsearchopsrequest -n demo esops-hscale-up-combined esops-hscale-down-combined
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Different Elasticsearch topology clustering modes [here](/docs/guides/elasticsearch/clustering/_index.md).
- Monitor your Elasticsearch with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
