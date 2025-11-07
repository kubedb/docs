---
title: Elasticsearch Restart
menu:
  docs_{{ .version }}:
    identifier: es-restart-elasticsearch
    name: Restart
    parent: es-elasticsearch-guides
    weight: 15
menu_name: docs_{{ .version }}
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart Elasticsearch

KubeDB supports restarting a Elasticsearch database using a `ElasticsearchOpsRequest`. Restarting can be
useful if some pods are stuck in a certain state or not functioning correctly.

This guide will demonstrate how to restart a Elasticsearch cluster using an OpsRequest.

---

## Before You Begin

- You need a running Kubernetes cluster and a properly configured `kubectl` command-line tool. If you don’t have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install the KubeDB CLI on your workstation and the KubeDB operator in your cluster by following the [installation steps](/docs/setup/README.md).

- For better isolation, this tutorial uses a separate namespace called `demo`:

```bash
kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/Elasticsearch](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/Elasticsearch) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Elasticsearch

In this section, we are going to deploy a Elasticsearch database using KubeDB.

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/update-version/elasticsearch.yaml
Elasticsearch.kubedb.com/Elasticsearch created
```
let's wait until all pods are in the `Running` state,

```shell
kubectl get pods -n demo
NAME        READY   STATUS    RESTARTS   AGE
es-demo-0   1/1     Running   0          6m28s
es-demo-1   1/1     Running   0          6m28s
es-demo-2   1/1     Running   0          6m28s
```



# Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: es-demo
  timeout: 10m
  apply: Always
```

Here,

- `spec.type` specifies the type of operation (Restart in this case). `Restart`  is used to perform a smart restart of the Elasticsearch cluster.

- `spec.databaseRef` references the Elasticsearch database. The OpsRequest must be created in the same namespace as the database.

- `spec.timeout` the maximum time the operator will wait for the operation to finish before marking it as failed.

- `spec.apply` determines whether to always apply the operation (Always) or  if the database phase is ready (IfReady).

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/restart.yaml
ElasticsearchOpsRequest.ops.kubedb.com/restart created
```

In a Elasticsearch cluster, all pods act as primary nodes. When you apply a restart OpsRequest, the KubeDB operator will restart the pods sequentially, one by one, to maintain cluster availability.

Let's watch the rolling restart process with:
```shell
NAME        READY   STATUS        RESTARTS   AGE
es-demo-0   1/1     Terminating   0          56m
es-demo-1   1/1     Running       0          55m
es-demo-2   1/1     Running       0          54m
```

```shell
NAME        READY   STATUS        RESTARTS    AGE
es-demo-0   1/1     Running        0          112s
es-demo-1   1/1     Terminating    0          55m
es-demo-2   1/1     Running        0          56m

```
```shell
NAME        READY   STATUS        RESTARTS   AGE
es-demo-0   1/1     Running       0          112s
es-demo-1   1/1     Running       0          42s
es-demo-2   1/1     Terminating   0          56m

```

```shell
$ kubectl get Elasticsearchopsrequest -n demo
NAME      TYPE      STATUS       AGE
restart   Restart   Successful   64m

$ kubectl get Elasticsearchopsrequest -n demo restart -oyaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"ElasticsearchOpsRequest","metadata":{"annotations":{},"name":"restart","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"es-demo"},"timeout":"10m","type":"Restart"}}
  creationTimestamp: "2025-11-06T08:21:19Z"
  generation: 1
  name: restart
  namespace: demo
  resourceVersion: "613519"
  uid: e5eb346e-7ca8-4f1d-9ca5-0869f49a8134
spec:
  apply: Always
  databaseRef:
    name: es-demo
  timeout: 10m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2025-11-06T08:21:20Z"
    message: Elasticsearch ops request is restarting nodes
    observedGeneration: 1
    reason: Restart
    status: "True"
    type: Restart
  - lastTransitionTime: "2025-11-06T08:21:28Z"
    message: pod exists; ConditionStatus:True; PodName:es-demo-0
    observedGeneration: 1
    status: "True"
    type: PodExists--es-demo-0
  - lastTransitionTime: "2025-11-06T08:21:28Z"
    message: create es client; ConditionStatus:True; PodName:es-demo-0
    observedGeneration: 1
    status: "True"
    type: CreateEsClient--es-demo-0
  - lastTransitionTime: "2025-11-06T08:21:28Z"
    message: evict pod; ConditionStatus:True; PodName:es-demo-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--es-demo-0
  - lastTransitionTime: "2025-11-06T08:22:53Z"
    message: create es client; ConditionStatus:True
    observedGeneration: 1
    status: "True"
    type: CreateEsClient
  - lastTransitionTime: "2025-11-06T08:21:58Z"
    message: pod exists; ConditionStatus:True; PodName:es-demo-1
    observedGeneration: 1
    status: "True"
    type: PodExists--es-demo-1
  - lastTransitionTime: "2025-11-06T08:21:58Z"
    message: create es client; ConditionStatus:True; PodName:es-demo-1
    observedGeneration: 1
    status: "True"
    type: CreateEsClient--es-demo-1
  - lastTransitionTime: "2025-11-06T08:21:58Z"
    message: evict pod; ConditionStatus:True; PodName:es-demo-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--es-demo-1
  - lastTransitionTime: "2025-11-06T08:22:28Z"
    message: pod exists; ConditionStatus:True; PodName:es-demo-2
    observedGeneration: 1
    status: "True"
    type: PodExists--es-demo-2
  - lastTransitionTime: "2025-11-06T08:22:28Z"
    message: create es client; ConditionStatus:True; PodName:es-demo-2
    observedGeneration: 1
    status: "True"
    type: CreateEsClient--es-demo-2
  - lastTransitionTime: "2025-11-06T08:22:28Z"
    message: evict pod; ConditionStatus:True; PodName:es-demo-2
    observedGeneration: 1
    status: "True"
    type: EvictPod--es-demo-2
  - lastTransitionTime: "2025-11-06T08:22:58Z"
    message: Successfully restarted all nodes
    observedGeneration: 1
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - lastTransitionTime: "2025-11-06T08:22:58Z"
    message: Successfully completed the modification process.
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful

```
**Verify Data Persistence**

After the restart, reconnect to the database and verify that the previously created database still exists:
Connect to the Cluster:

```bash
# Port-forward the service to local machine
$ kubectl port-forward -n demo svc/es-standalone 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```

```bash
# Get admin username & password from k8s secret
$ kubectl get secret -n demo es-standalone-admin-cred -o jsonpath='{.data.username}' | base64 -d
elastic
$ kubectl get secret -n demo es-standalone-admin-cred -o jsonpath='{.data.password}' | base64 -d
d9QpQKiTcLNZx_gA

# Check cluster health
$  curl -XGET -k -u "elastic:d9QpQKiTcLNZx_gA" "https://localhost:9200/_cluster/health?pretty"
{
  "cluster_name" : "es-demo",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 3,
  "number_of_data_nodes" : 3,
  "active_primary_shards" : 4,
  "active_shards" : 8,
  "relocating_shards" : 0,
  "initializing_shards" : 0,
  "unassigned_shards" : 0,
  "unassigned_primary_shards" : 0,
  "delayed_unassigned_shards" : 0,
  "number_of_pending_tasks" : 0,
  "number_of_in_flight_fetch" : 0,
  "task_max_waiting_in_queue_millis" : 0,
  "active_shards_percent_as_number" : 100.0
}

```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete  Elasticsearchopsrequest -n demo restart
kubectl delete Elasticsearch -n demo es-demo
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Detail concepts of [ElasticsearchopsRequest object](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md)
