---
title: Update Version Elasticsearch
menu:
  docs_{{ .version }}:
    identifier: es-updateversion-Elasticsearch
    name: Elasticsearch
    parent: es-updateversion-elasticsearch
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Update version of Elasticsearch

This guide will show you how to use `KubeDB` Ops-manager operator to update the version of `Elasticsearch` Combined or Topology cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
    - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)
    - [Updating Overview](/docs/guides/elasticsearch/update-version/elasticsearch.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/elasticsearch](/docs/examples/elasticsearch) directory of [kubedb/docs](https://github.com/kube/docs) repository.

## Prepare Elasticsearch

Now, we are going to deploy a `Elasticsearch` replicaset database with version `xpack-8.11.1`.

### Deploy Elasticsearch

In this section, we are going to deploy a Elasticsearch topology cluster. Then, in the next section we will update the version using `ElasticsearchOpsRequest` CRD. Below is the YAML of the `Elasticsearch` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-demo
  namespace: demo
spec:
  deletionPolicy: Delete
  enableSSL: true
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: local-path
  storageType: Durable
  version: xpack-9.1.3
 
```

Let's create the `Elasticsearch` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/update-version/Elasticsearch.yaml
Elasticsearch.kubedb.com/es-demo created
```

Now, wait until `es-demo` created has status `Ready`. i.e,

```bash
$ kubectl get es -n demo 
NAME      VERSION        STATUS   AGE
es-demo   xpack-9.1.3   Ready    9m10s

```

We are now ready to apply the `ElasticsearchOpsRequest` CR to update.

### update Elasticsearch Version

Here, we are going to update `Elasticsearch` from `xpack-9.1.3` to `xpack-9.1.4`.

#### Create ElasticsearchOpsRequest:

In order to update the version, we have to create a `ElasticsearchOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `ElasticsearchOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: es-demo-update
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: es-demo
  updateVersion:
    targetVersion: xpack-9.1.4
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `es-demo` Elasticsearch.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version of the database `xpack-8.16.4`.

> **Note:** If you want to update combined Elasticsearch, you just refer to the `Elasticsearch` combined object name in `spec.databaseRef.name`. To create a combined Elasticsearch, you can refer to the [Elasticsearch Combined](/docs/guides/elasticsearch/clustering/combined-cluster/index.md) guide.

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/update-version/update-version.yaml
Elasticsearchopsrequest.ops.kubedb.com/Elasticsearch-update-version created
```

#### Verify Elasticsearch version updated successfully

If everything goes well, `KubeDB` Ops-manager operator will update the image of `Elasticsearch` object and related `PetSets` and `Pods`.

Let's wait for `ElasticsearchOpsRequest` to be `Successful`.  Run the following command to watch `ElasticsearchOpsRequest` CR,

```bash
$ kubectl get Elasticsearchopsrequest -n demo
NAME                   TYPE            STATUS        AGE
Elasticsearch-update-version   UpdateVersion   Successful    2m6s
```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed to update the database version.

```bash
$ kubectl describe Elasticsearchopsrequest -n demo es-demo-update
Name:         es-demo-update
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2025-11-06T05:19:15Z
  Generation:          1
  Resource Version:    609353
  UID:                 722d8557-a6c6-4412-87d4-61faee8a3be2
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  es-demo
  Type:    UpdateVersion
  Update Version:
    Target Version:  xpack-9.1.4
Status:
  Conditions:
    Last Transition Time:  2025-11-06T05:19:15Z
    Message:               Elasticsearch ops request is updating database version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2025-11-06T05:19:18Z
    Message:               Successfully updated PetSets
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-11-06T05:19:23Z
    Message:               pod exists; ConditionStatus:True; PodName:es-demo-0
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-demo-0
    Last Transition Time:  2025-11-06T05:19:23Z
    Message:               create es client; ConditionStatus:True; PodName:es-demo-0
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-demo-0
    Last Transition Time:  2025-11-06T05:19:23Z
    Message:               disable shard allocation; ConditionStatus:True; PodName:es-demo-0
    Observed Generation:   1
    Status:                True
    Type:                  DisableShardAllocation--es-demo-0
    Last Transition Time:  2025-11-06T05:19:23Z
    Message:               evict pod; ConditionStatus:True; PodName:es-demo-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-demo-0
    Last Transition Time:  2025-11-06T05:21:03Z
    Message:               create es client; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient
    Last Transition Time:  2025-11-06T05:19:58Z
    Message:               re enable shard allocation; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ReEnableShardAllocation
    Last Transition Time:  2025-11-06T05:20:03Z
    Message:               pod exists; ConditionStatus:True; PodName:es-demo-1
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-demo-1
    Last Transition Time:  2025-11-06T05:20:03Z
    Message:               create es client; ConditionStatus:True; PodName:es-demo-1
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-demo-1
    Last Transition Time:  2025-11-06T05:20:03Z
    Message:               disable shard allocation; ConditionStatus:True; PodName:es-demo-1
    Observed Generation:   1
    Status:                True
    Type:                  DisableShardAllocation--es-demo-1
    Last Transition Time:  2025-11-06T05:20:03Z
    Message:               evict pod; ConditionStatus:True; PodName:es-demo-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-demo-1
    Last Transition Time:  2025-11-06T05:20:33Z
    Message:               pod exists; ConditionStatus:True; PodName:es-demo-2
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-demo-2
    Last Transition Time:  2025-11-06T05:20:33Z
    Message:               create es client; ConditionStatus:True; PodName:es-demo-2
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-demo-2
    Last Transition Time:  2025-11-06T05:20:33Z
    Message:               disable shard allocation; ConditionStatus:True; PodName:es-demo-2
    Observed Generation:   1
    Status:                True
    Type:                  DisableShardAllocation--es-demo-2
    Last Transition Time:  2025-11-06T05:20:33Z
    Message:               evict pod; ConditionStatus:True; PodName:es-demo-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-demo-2
    Last Transition Time:  2025-11-06T05:21:08Z
    Message:               Successfully updated all nodes
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2025-11-06T05:21:08Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                             Age   From                         Message
  ----     ------                                                             ----  ----                         -------
  Normal   PauseDatabase                                                      29m   KubeDB Ops-manager Operator  Pausing Elasticsearch demo/es-demo
  Warning  pod exists; ConditionStatus:True; PodName:es-demo-0                29m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-demo-0
  Warning  create es client; ConditionStatus:True; PodName:es-demo-0          29m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-demo-0
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-demo-0  29m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-demo-0
  Warning  evict pod; ConditionStatus:True; PodName:es-demo-0                 29m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-demo-0
  Warning  create es client; ConditionStatus:False                            29m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                             29m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                   29m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-demo-1                29m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-demo-1
  Warning  create es client; ConditionStatus:True; PodName:es-demo-1          29m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-demo-1
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-demo-1  29m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-demo-1
  Warning  evict pod; ConditionStatus:True; PodName:es-demo-1                 29m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-demo-1
  Warning  create es client; ConditionStatus:False                            29m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                             28m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                   28m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-demo-2                28m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-demo-2
  Warning  create es client; ConditionStatus:True; PodName:es-demo-2          28m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-demo-2
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-demo-2  28m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-demo-2
  Warning  evict pod; ConditionStatus:True; PodName:es-demo-2                 28m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-demo-2
  Warning  create es client; ConditionStatus:False                            28m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                             28m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                   28m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Normal   RestartPods                                                        28m   KubeDB Ops-manager Operator  Successfully updated all nodes
  Normal   ResumeDatabase                                                     28m   KubeDB Ops-manager Operator  Resuming Elasticsearch
  Normal   ResumeDatabase                                                     28m   KubeDB Ops-manager Operator  Resuming Elasticsearch demo/es-demo
  Normal   ResumeDatabase                                                     28m   KubeDB Ops-manager Operator  Successfully resumed Elasticsearch demo/es-demo
  Normal   Successful                                                         28m   KubeDB Ops-manager Operator  Successfully Updated Database

```

Now, we are going to verify whether the `Elasticsearch` and the related `PetSets` and their `Pods` have the new version image. Let's check,

```bash
$ kubectl get es -n demo es-demo -o=jsonpath='{.spec.version}{"\n"}'
xpack-9.1.4

$ kubectl get petset -n demo es-demo -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/elastic:9.1.4@sha256:e0b89e3ace47308fa5fa842823bc622add3733e47c1067cd1e6afed2cfd317ca

$ kubectl get pods -n demo es-demo-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/elastic:9.1.4

```

You can see from above, our `Elasticsearch` has been updated with the new version. So, the updateVersion process is successfully completed.

> **NOTE:** If you want to update Opensearch, you can follow the same steps as above but using `ElasticsearchOpsRequest` CRD. You can visit [OpenSearch ](/docs/guides/elasticsearch/quickstart/overview/opensearch) guide for more details.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete Elasticsearchopsrequest -n demo es-demo-update
kubectl delete es -n demo es-demo
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Detail concepts of [ElasticsearchOpsRequest object](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md).
- Detailed concept of [Elasticesearch Version](/docs/guides/elasticsearch/concepts/catalog/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
