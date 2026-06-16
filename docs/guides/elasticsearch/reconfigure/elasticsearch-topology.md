---
title: Reconfigure Elasticsearch Topology Cluster
menu:
  docs_{{ .version }}:
    identifier: es-reconfigure-topology
    name: Topology Cluster
    parent: es-reconfigure
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Elasticsearch Topology Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to reconfigure an Elasticsearch Topology cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
    - [Topology Cluster](/docs/guides/elasticsearch/clustering/topology-cluster/simple-dedicated-cluster/index.md)
    - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)
    - [Reconfigure Overview](/docs/guides/elasticsearch/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/elasticsearch](/docs/examples/elasticsearch) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

Now, we are going to deploy an `Elasticsearch` Topology cluster using a supported version by `KubeDB` operator. Then we are going to apply `ElasticsearchOpsRequest` to reconfigure its configuration.

### Prepare Elasticsearch Topology Cluster

Now, we are going to deploy an `Elasticsearch` topology cluster with version `xpack-8.19.9`.

### Deploy Elasticsearch

At first, we will create a secret with role-prefixed configuration files. In a topology cluster, you can target specific node roles by prefixing the filename with the role name (e.g., `master-elasticsearch.yml` applies only to master nodes, `data-elasticsearch.yml` applies only to data nodes).

**master-elasticsearch.yml:**

```yaml
cluster.max_shards_per_node: 2000
```

**data-elasticsearch.yml:**

```yaml
indices.query.bool.max_clause_count: 2048
```

**ingest-elasticsearch.yml:**

```yaml
http.max_content_length: 200mb
```

Here, `cluster.max_shards_per_node` is set to `2000` (default `1000`) for master nodes, `indices.query.bool.max_clause_count` is set to `2048` (default `1024`) for data nodes, and `http.max_content_length` is set to `200mb` (default `100mb`) for ingest nodes.

Let's create a k8s secret containing the above configuration where the file name will be the key and the file-content as the value:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: es-topology-custom-config
  namespace: demo
stringData:
  master-elasticsearch.yml: |-
    cluster.max_shards_per_node: 2000
  data-elasticsearch.yml: |-
    indices.query.bool.max_clause_count: 2048
  ingest-elasticsearch.yml: |-
    http.max_content_length: 200mb
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/reconfigure/es-topology-custom-config.yaml
secret/es-topology-custom-config created
```

In this section, we are going to create an Elasticsearch object specifying `spec.configuration.secretName` field to apply this custom configuration. Below is the YAML of the `Elasticsearch` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-topology
  namespace: demo
spec:
  version: xpack-8.19.9
  enableSSL: true
  configuration:
    secretName: es-topology-custom-config
  topology:
    master:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    data:
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    ingest:
      replicas: 1
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/reconfigure/es-topology.yaml
elasticsearch.kubedb.com/es-topology created
```

Now, wait until `es-topology` has status `Ready`. i.e,

```bash
$ kubectl get es -n demo -w
NAME          VERSION          STATUS         AGE
es-topology   xpack-8.19.9    Provisioning   0s
es-topology   xpack-8.19.9    Provisioning   24s
.
.
es-topology   xpack-8.19.9    Ready          92s
```

Now, we will check if the Elasticsearch has started with the custom configuration we have provided.

Exec into the master node and check the cluster setting:

```bash
$ kubectl exec -it -n demo es-topology-master-0 -c elasticsearch -- curl -k -XGET "https://localhost:9200/_nodes/settings?filter_path=nodes.*.settings.cluster&pretty" --user "elastic:$ELASTIC_USER_PASSWORD" | grep max_shards_per_node
          "max_shards_per_node" : "2000",
```

Exec into a data node and check the index settings:

```bash
$ kubectl exec -it -n demo es-topology-data-0 -c elasticsearch -- curl -k -XGET "https://localhost:9200/_nodes/settings?filter_path=nodes.*.settings.indices&pretty" --user "elastic:$ELASTIC_USER_PASSWORD" | grep max_clause_count
              "max_clause_count" : "2048"
              "max_clause_count" : "2048"
```

Exec into the ingest node and check the HTTP settings:

```bash
$ kubectl exec -it -n demo es-topology-ingest-0 -c elasticsearch -- curl -k -XGET "https://localhost:9200/_nodes/settings?filter_path=nodes.*.settings.http&pretty" --user "elastic:$ELASTIC_USER_PASSWORD" | grep max_content_length
          "max_content_length" : "200mb",
```

Here, we can see that our given configurations are applied to the respective node roles.

### Reconfigure using new config secret

Now we will reconfigure this cluster to update `cluster.max_shards_per_node` to `3000` for master nodes, `indices.query.bool.max_clause_count` to `4096` for data nodes, and `http.max_content_length` to `300mb` for ingest nodes.

Update our configuration files with the new values.

**master-elasticsearch.yml:**

```yaml
cluster.max_shards_per_node: 3000
```

**data-elasticsearch.yml:**

```yaml
indices.query.bool.max_clause_count: 4096
```

**ingest-elasticsearch.yml:**

```yaml
http.max_content_length: 300mb
```

Then, we will create a new secret with these configuration files.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: new-es-topology-custom-config
  namespace: demo
stringData:
  master-elasticsearch.yml: |-
    cluster.max_shards_per_node: 3000
  data-elasticsearch.yml: |-
    indices.query.bool.max_clause_count: 4096
  ingest-elasticsearch.yml: |-
    http.max_content_length: 300mb
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/reconfigure/new-es-topology-custom-config.yaml
secret/new-es-topology-custom-config created
```

#### Create ElasticsearchOpsRequest

Now, we will use this secret to replace the previous secret using an `ElasticsearchOpsRequest` CR. The `ElasticsearchOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: esops-reconfigure-topology
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: es-topology
  configuration:
    configSecret:
      name: new-es-topology-custom-config
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `es-topology` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.configSecret.name` specifies the name of the new secret.

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/reconfigure/es-reconfigure-update-topology.yaml
elasticsearchopsrequest.ops.kubedb.com/esops-reconfigure-topology created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Ops-manager operator will update the `configSecret` of `Elasticsearch` object.

Let's wait for `ElasticsearchOpsRequest` to be `Successful`. Run the following command to watch `ElasticsearchOpsRequest` CR,

```bash
$ kubectl get elasticsearchopsrequests -n demo
NAME                         TYPE          STATUS       AGE
esops-reconfigure-topology   Reconfigure   Successful   4m55s
```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe elasticsearchopsrequest -n demo esops-reconfigure-topology
Name:         esops-reconfigure-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2026-06-15T11:13:59Z
  Generation:          1
  Resource Version:    434719
  UID:                 a050a578-420c-4bb8-a390-3b32b369ec65
Spec:
  Apply:  IfReady
  Configuration:
    Config Secret:
      Name:   new-es-topology-custom-config
    Restart:  auto
  Database Ref:
    Name:       es-topology
  Max Retries:  1
  Timeout:      5m
  Type:         Reconfigure
Status:
  Conditions:
    Last Transition Time:  2026-06-15T11:13:59Z
    Message:               Elasticsearch ops request is Reconfiguring
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2026-06-15T11:14:09Z
    Message:               Successfully updated petSets
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2026-06-15T11:14:19Z
    Message:               pod exists; ConditionStatus:True; PodName:es-topology-ingest-0
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-topology-ingest-0
    Last Transition Time:  2026-06-15T11:14:19Z
    Message:               create es client; ConditionStatus:True; PodName:es-topology-ingest-0
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-topology-ingest-0
    Last Transition Time:  2026-06-15T11:14:19Z
    Message:               evict pod; ConditionStatus:True; PodName:es-topology-ingest-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-topology-ingest-0
    Last Transition Time:  2026-06-15T11:14:59Z
    Message:               create es client; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient
    Last Transition Time:  2026-06-15T11:14:34Z
    Message:               pod exists; ConditionStatus:True; PodName:es-topology-data-0
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-topology-data-0
    Last Transition Time:  2026-06-15T11:14:34Z
    Message:               create es client; ConditionStatus:True; PodName:es-topology-data-0
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-topology-data-0
    Last Transition Time:  2026-06-15T11:14:34Z
    Message:               evict pod; ConditionStatus:True; PodName:es-topology-data-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-topology-data-0
    Last Transition Time:  2026-06-15T11:14:49Z
    Message:               pod exists; ConditionStatus:True; PodName:es-topology-data-1
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-topology-data-1
    Last Transition Time:  2026-06-15T11:14:49Z
    Message:               create es client; ConditionStatus:True; PodName:es-topology-data-1
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-topology-data-1
    Last Transition Time:  2026-06-15T11:14:49Z
    Message:               evict pod; ConditionStatus:True; PodName:es-topology-data-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-topology-data-1
    Last Transition Time:  2026-06-15T11:15:04Z
    Message:               pod exists; ConditionStatus:True; PodName:es-topology-master-0
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-topology-master-0
    Last Transition Time:  2026-06-15T11:15:04Z
    Message:               create es client; ConditionStatus:True; PodName:es-topology-master-0
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-topology-master-0
    Last Transition Time:  2026-06-15T11:15:04Z
    Message:               evict pod; ConditionStatus:True; PodName:es-topology-master-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-topology-master-0
    Last Transition Time:  2026-06-15T11:15:14Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2026-06-15T11:15:15Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                Age    From                         Message
  ----     ------                                                                ----   ----                         -------
  Normal   PauseDatabase                                                         4m47s  KubeDB Ops-manager Operator  Pausing Elasticsearch demo/es-topology
  Warning  pod exists; ConditionStatus:True; PodName:es-topology-ingest-0        4m29s  KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-topology-ingest-0
  Warning  create es client; ConditionStatus:True; PodName:es-topology-ingest-0  4m29s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-topology-ingest-0
  Warning  evict pod; ConditionStatus:True; PodName:es-topology-ingest-0         4m29s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-topology-ingest-0
  Warning  create es client; ConditionStatus:False                               4m24s  KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                                4m19s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-topology-data-0          4m14s  KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-topology-data-0
  Warning  create es client; ConditionStatus:True; PodName:es-topology-data-0    4m14s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-topology-data-0
  Warning  evict pod; ConditionStatus:True; PodName:es-topology-data-0           4m14s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-topology-data-0
  Warning  create es client; ConditionStatus:False                               4m9s   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                                4m4s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-topology-data-1          3m59s  KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-topology-data-1
  Warning  create es client; ConditionStatus:True; PodName:es-topology-data-1    3m59s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-topology-data-1
  Warning  evict pod; ConditionStatus:True; PodName:es-topology-data-1           3m59s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-topology-data-1
  Warning  create es client; ConditionStatus:False                               3m54s  KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                                3m49s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-topology-master-0        3m44s  KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-topology-master-0
  Warning  create es client; ConditionStatus:True; PodName:es-topology-master-0  3m44s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-topology-master-0
  Warning  evict pod; ConditionStatus:True; PodName:es-topology-master-0         3m44s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-topology-master-0
  Warning  create es client; ConditionStatus:True                                3m39s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Normal   RestartNodes                                                          3m34s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Successful                                                            3m33s  KubeDB Ops-manager Operator  Successfully reconfigured all elasticsearch nodes.
```

Now let's exec into a master node, a data node, and the ingest node to verify the new configuration.

```bash
$ kubectl exec -it -n demo es-topology-master-0 -c elasticsearch -- curl -k -XGET "https://localhost:9200/_nodes/settings?filter_path=nodes.*.settings.cluster&pretty" --user "elastic:$ELASTIC_USER_PASSWORD" | grep max_shards_per_node
          "max_shards_per_node" : "3000",

$ kubectl exec -it -n demo es-topology-data-0 -c elasticsearch -- curl -k -XGET "https://localhost:9200/_nodes/settings?filter_path=nodes.*.settings.indices&pretty" --user "elastic:$ELASTIC_USER_PASSWORD" | grep max_clause_count
              "max_clause_count" : "4096"
              "max_clause_count" : "4096"

$ kubectl exec -it -n demo es-topology-ingest-0 -c elasticsearch -- curl -k -XGET "https://localhost:9200/_nodes/settings?filter_path=nodes.*.settings.http&pretty" --user "elastic:$ELASTIC_USER_PASSWORD" | grep max_content_length
          "max_content_length" : "300mb",
```

As we can see, the values have been updated on their respective node roles. So the reconfiguration of the cluster is successful.


### Reconfigure using apply config

Now we will reconfigure this cluster again using the `applyConfig` field. This will merge the new config into the existing secret without requiring a new secret.

#### Create ElasticsearchOpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: esops-reconfigure-apply-topology
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: es-topology
  configuration:
    applyConfig:
      master-elasticsearch.yml: |
        cluster.max_shards_per_node: 4000
      data-elasticsearch.yml: |
        indices.query.bool.max_clause_count: 8192
      ingest-elasticsearch.yml: |
        http.max_content_length: 400mb
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `es-topology` cluster.
- `spec.type` specifies that we are performing `Reconfigure` on Elasticsearch.
- `spec.configuration.applyConfig` specifies the new configuration that will be merged into the existing secret.

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/reconfigure/es-reconfigure-apply-topology.yaml
elasticsearchopsrequest.ops.kubedb.com/esops-reconfigure-apply-topology created
```

#### Verify the new configuration is working

Let's wait for `ElasticsearchOpsRequest` to be `Successful`.

```bash
$  kubectl get elasticsearchopsrequests -n demo esops-reconfigure-apply-topology
NAME                               TYPE          STATUS       AGE
esops-reconfigure-apply-topology   Reconfigure   Successful   6m52s
```

Now let's verify the updated values on each node role.

```bash
$ kubectl exec -it -n demo es-topology-master-0 -c elasticsearch -- curl -k -XGET "https://localhost:9200/_nodes/settings?filter_path=nodes.*.settings.cluster&pretty" --user "elastic:$ELASTIC_USER_PASSWORD" | grep max_shards_per_node
          "max_shards_per_node" : "4000",

$ kubectl exec -it -n demo es-topology-data-0 -c elasticsearch -- curl -k -XGET "https://localhost:9200/_nodes/settings?filter_path=nodes.*.settings.indices&pretty" --user "elastic:$ELASTIC_USER_PASSWORD" | grep max_clause_count
              "max_clause_count" : "8192"
              "max_clause_count" : "8192"

$ kubectl exec -it -n demo es-topology-ingest-0 -c elasticsearch -- curl -k -XGET "https://localhost:9200/_nodes/settings?filter_path=nodes.*.settings.http&pretty" --user "elastic:X4am_*ihVy~M)m0j" | grep max_content_length
          "max_content_length" : "400mb",
```

As we can see, `cluster.max_shards_per_node` has been changed from `3000` to `4000` on master nodes, `indices.query.bool.max_clause_count` has been changed from `4096` to `8192` on data nodes, and `http.max_content_length` has been changed from `300mb` to `400mb` on ingest nodes. So the reconfiguration using the `applyConfig` field is successful.


## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete es -n demo es-topology
kubectl delete elasticsearchopsrequest -n demo esops-reconfigure-apply-topology esops-reconfigure-topology
kubectl delete secret -n demo es-topology-custom-config new-es-topology-custom-config
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Different Elasticsearch topology clustering modes [here](/docs/guides/elasticsearch/clustering/topology-cluster/hot-warm-cold-cluster/index.md).
- Monitor your Elasticsearch database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
