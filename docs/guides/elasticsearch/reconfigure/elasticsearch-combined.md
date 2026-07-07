---
title: Reconfigure Elasticsearch Combined Cluster
menu:
  docs_{{ .version }}:
    identifier: es-reconfigure-combined
    name: Combined Cluster
    parent: es-reconfigure
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Elasticsearch Combined Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to reconfigure an Elasticsearch Combined cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
    - [Combined Cluster](/docs/guides/elasticsearch/clustering/combined-cluster/index.md)
    - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)
    - [Reconfigure Overview](/docs/guides/elasticsearch/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
kubectl create ns demo
```
namespace/demo created

> **Note:** YAML files used in this tutorial are stored in [docs/examples/elasticsearch](/docs/examples/elasticsearch) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

Now, we are going to deploy an `Elasticsearch` Combined cluster using a supported version by `KubeDB` operator. Then we are going to apply `ElasticsearchOpsRequest` to reconfigure its configuration.

### Prepare Elasticsearch Combined Cluster

Now, we are going to deploy an `Elasticsearch` combined cluster with version `xpack-8.19.9`.

### Deploy Elasticsearch

At first, we will create a secret with the `elasticsearch.yml` file containing required configuration settings.

**elasticsearch.yml:**

```yaml
indices.query.bool.max_clause_count: 2048
```

Let's create a k8s secret containing the above configuration where the file name will be the key and the file-content as the value:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: es-combined-custom-config
  namespace: demo
stringData:
  elasticsearch.yml: |-
    indices.query.bool.max_clause_count: 2048
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/reconfigure/es-combined-custom-config.yaml
```
secret/es-combined-custom-config created

In this section, we are going to create an Elasticsearch object specifying `spec.configuration.secretName` field to apply this custom configuration. Below is the YAML of the `Elasticsearch` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-combined
  namespace: demo
spec:
  version: xpack-8.19.9
  enableSSL: true
  replicas: 2
  configuration:
    secretName: es-combined-custom-config
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: WipeOut
```

Let's create the `Elasticsearch` CR we have shown above,

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/reconfigure/es-combined.yaml
```
elasticsearch.kubedb.com/es-combined created

Now, wait until `es-combined` has status `Ready`. i.e,

```bash
kubectl get es -n demo 
```
NAME          VERSION        STATUS   AGE
es-combined   xpack-8.19.9   Ready    20m

Now, we will check if the Elasticsearch has started with the custom configuration we have provided.

Exec into the Elasticsearch pod and query the cluster settings to see the configuration:

```bash
kubectl exec -it -n demo es-combined-0 -c elasticsearch -- curl -k -XGET "https://localhost:9200/_nodes/settings?filter_path=nodes.*.settings.indices&pretty" --user "elastic:X4gzeLWqUHKMoQT7" | grep max_clause_count
```
              "max_clause_count" : "2048"
              "max_clause_count" : "2048"

Here, we can see that our given configuration is applied to the Elasticsearch cluster for all nodes. `indices.query.bool.max_clause_count` is set to `2048` from the default value `1024`.

### Reconfigure using new config secret

Now we will reconfigure this cluster to set `indices.query.bool.max_clause_count` to `4096`.

Update our `elasticsearch.yml` file with the new configuration.

**elasticsearch.yml:**

```yaml
indices.query.bool.max_clause_count: 4096
```

Then, we will create a new secret with this configuration file.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: new-es-combined-custom-config
  namespace: demo
stringData:
  elasticsearch.yml: |-
    indices.query.bool.max_clause_count: 4096
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/reconfigure/new-es-combined-custom-config.yaml
```
secret/new-es-combined-custom-config created

#### Create ElasticsearchOpsRequest

Now, we will use this secret to replace the previous secret using an `ElasticsearchOpsRequest` CR. The `ElasticsearchOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: esops-reconfigure-combined
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: es-combined
  configuration:
    configSecret:
      name: new-es-combined-custom-config
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `es-combined` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.configSecret.name` specifies the name of the new secret.

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/reconfigure/es-reconfigure-update-combined.yaml
```
elasticsearchopsrequest.ops.kubedb.com/esops-reconfigure-combined created

#### Verify the new configuration is working

If everything goes well, `KubeDB` Ops-manager operator will update the `configSecret` of `Elasticsearch` object.

Let's wait for `ElasticsearchOpsRequest` to be `Successful`. Run the following command to watch `ElasticsearchOpsRequest` CR,

```bash
kubectl get elasticsearchopsrequests -n demo
```
NAME                         TYPE          STATUS       AGE
esops-reconfigure-combined   Reconfigure   Successful   73s

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
 kubectl describe elasticsearchopsrequest -n demo esops-reconfigure-combined
```
Name:         esops-reconfigure-combined
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2026-06-15T09:43:13Z
  Generation:          1
  Resource Version:    431415
  UID:                 25c9da3e-6e1f-4073-a98c-8df9061307e5
Spec:
  Apply:  IfReady
  Configuration:
    Config Secret:
      Name:   new-es-combined-custom-config
    Restart:  auto
  Database Ref:
    Name:       es-combined
  Max Retries:  1
  Timeout:      5m
  Type:         Reconfigure
Status:
  Conditions:
    Last Transition Time:  2026-06-15T09:43:13Z
    Message:               Elasticsearch ops request is Reconfiguring
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2026-06-15T09:43:18Z
    Message:               Successfully updated petSets
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2026-06-15T09:43:28Z
    Message:               pod exists; ConditionStatus:True; PodName:es-combined-0
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-combined-0
    Last Transition Time:  2026-06-15T09:43:28Z
    Message:               create es client; ConditionStatus:True; PodName:es-combined-0
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-combined-0
    Last Transition Time:  2026-06-15T09:43:29Z
    Message:               evict pod; ConditionStatus:True; PodName:es-combined-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-combined-0
    Last Transition Time:  2026-06-15T09:43:53Z
    Message:               create es client; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient
    Last Transition Time:  2026-06-15T09:43:43Z
    Message:               pod exists; ConditionStatus:True; PodName:es-combined-1
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-combined-1
    Last Transition Time:  2026-06-15T09:43:43Z
    Message:               create es client; ConditionStatus:True; PodName:es-combined-1
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-combined-1
    Last Transition Time:  2026-06-15T09:43:44Z
    Message:               evict pod; ConditionStatus:True; PodName:es-combined-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-combined-1
    Last Transition Time:  2026-06-15T09:43:58Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2026-06-15T09:43:59Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                         Age   From                         Message
  ----     ------                                                         ----  ----                         -------
  Normal   PauseDatabase                                                  98s   KubeDB Ops-manager Operator  Pausing Elasticsearch demo/es-combined
  Warning  pod exists; ConditionStatus:True; PodName:es-combined-0        85s   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-combined-0
  Warning  create es client; ConditionStatus:True; PodName:es-combined-0  85s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-combined-0
  Warning  evict pod; ConditionStatus:True; PodName:es-combined-0         84s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-combined-0
  Warning  create es client; ConditionStatus:False                        80s   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                         75s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-combined-1        70s   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-combined-1
  Warning  create es client; ConditionStatus:True; PodName:es-combined-1  70s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-combined-1
  Warning  evict pod; ConditionStatus:True; PodName:es-combined-1         69s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-combined-1
  Warning  create es client; ConditionStatus:False                        65s   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                         60s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Normal   RestartNodes                                                   55s   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Successful                                                     54s   KubeDB Ops-manager Operator  Successfully reconfigured all elasticsearch nodes.

Now let's exec into one of the instances and query the cluster settings to check the new configuration.

```bash
kubectl exec -it -n demo es-combined-0 -c elasticsearch -- curl -k -XGET "https://localhost:9200/_nodes/settings?filter_path=nodes.*.settings.indices&pretty" --user "elastic:X4gzeLWqUHKMoQT7" | grep max_clause_count
```
              "max_clause_count" : "4096"
              "max_clause_count" : "4096"

As we can see from the configuration of the ready Elasticsearch, the value of `indices.query.bool.max_clause_count` has been changed from `2048` to `4096`. So the reconfiguration of the cluster is successful.


### Reconfigure using apply config

Now we will reconfigure this cluster again to set `indices.query.bool.max_clause_count` to `8192`. This time we won't use a new secret. We will use the `applyConfig` field of the `ElasticsearchOpsRequest`. This will merge the new config into the existing secret.

#### Create ElasticsearchOpsRequest

Now, we will use the new configuration in the `applyConfig` field in the `ElasticsearchOpsRequest` CR. The `ElasticsearchOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: esops-reconfigure-apply-combined
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: es-combined
  configuration:
    applyConfig:
      elasticsearch.yml: |
        indices.query.bool.max_clause_count: 8192
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `es-combined` cluster.
- `spec.type` specifies that we are performing `Reconfigure` on Elasticsearch.
- `spec.configuration.applyConfig` specifies the new configuration that will be merged into the existing secret.

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/reconfigure/es-reconfigure-apply-combined.yaml
```
elasticsearchopsrequest.ops.kubedb.com/esops-reconfigure-apply-combined created

#### Verify the new configuration is working

If everything goes well, `KubeDB` Ops-manager operator will merge this new config with the existing configuration.

Let's wait for `ElasticsearchOpsRequest` to be `Successful`. Run the following command to watch `ElasticsearchOpsRequest` CR,

```bash
kubectl get elasticsearchopsrequests -n demo esops-reconfigure-apply-combined
```
NAME                               TYPE          STATUS       AGE
esops-reconfigure-apply-combined   Reconfigure   Successful   118s

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. Now let's exec into one of the instances and check the new configuration.

```bash
kubectl exec -it -n demo es-combined-0 -c elasticsearch -- curl -k -XGET "https://localhost:9200/_nodes/settings?filter_path=nodes.*.settings.indices&pretty" --user "elastic:X4gzeLWqUHKMoQT7" | grep max_clause_count
```
              "max_clause_count" : "8192"
              "max_clause_count" : "8192"

As we can see from the configuration of the ready Elasticsearch, the value of `indices.query.bool.max_clause_count` has been changed from `4096` to `8192`. So the reconfiguration of the database using the `applyConfig` field is successful.


## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete es -n demo es-combined
kubectl delete elasticsearchopsrequest -n demo esops-reconfigure-apply-combined esops-reconfigure-combined
kubectl delete secret -n demo es-combined-custom-config new-es-combined-custom-config
kubectl delete namespace demo
```

## Next Steps

- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Different Elasticsearch topology clustering modes [here](/docs/guides/elasticsearch/clustering/combined-cluster/index.md).
- Monitor your Elasticsearch database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
