---
title: Elasticsearch Restart
menu:
  docs_{{ .version }}:
    identifier: es-restart-elasticsearch
    name: Restart
    parent: es-elasticsearch-guides
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---
> New to KubeDB? Please start [here](/docs/README.md).

# Restart Elasticsearch

KubeDB supports restarting an Elasticsearch database using a `ElasticsearchOpsRequest`. Restarting can be
useful if some pods are stuck in a certain state or not functioning correctly.

This guide will demonstrate how to restart an Elasticsearch cluster using an OpsRequest.

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
  name: es
  namespace: demo
spec:
  version: xpack-8.2.3
  enableSSL: true
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Elasticsearch` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/quickstart/overview/elasticsearch/yamls/elasticsearch-v1.yaml
Elasticsearch.kubedb.com/Elasticsearch created
```
let's wait until all pods are in the `Running` state,

```shell
kubectl get pods -n demo
NAME    READY   STATUS    RESTARTS   AGE
es-0   2/2     Running   0          6m28s
es-1   2/2     Running   0          6m28s
es-2   2/2     Running   0          6m28s
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
    name: es
  timeout: 3m
  apply: Always
```

Here,

- `spec.type` specifies the type of operation (Restart in this case).

- `spec.databaseRef` references the Elasticsearch database. The OpsRequest must be created in the same namespace as the database.

- `spec.timeout` the maximum time the operator will wait for the operation to finish before marking it as failed.

- `spec.apply` determines whether to always apply the operation (Always) or  if the database phase is ready (IfReady).

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/restart/yamls/restart.yaml
ElasticsearchOpsRequest.ops.kubedb.com/restart created
```

In a Elasticsearch cluster, all pods act as primary nodes. When you apply a restart OpsRequest, the KubeDB operator will restart the pods sequentially, one by one, to maintain cluster availability.

Let's watch the rolling restart process with:
```shell
NAME    READY   STATUS        RESTARTS   AGE
es-0   2/2     Terminating   0          56m
es-1   2/2     Running       0          55m
es-2   2/2     Running       0          54m
```

```shell
NAME    READY   STATUS        RESTARTS    AGE
es-0   2/2     Running        0          112s
es-1   2/2     Terminating    0          55m
es-2   2/2     Running        0          56m

```
```shell
NAME    READY   STATUS        RESTARTS   AGE
es-0   2/2     Running       0          112s
es-1   2/2     Running       0          42s
es-2   2/2     Terminating   0          56m

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
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"ElasticsearchOpsRequest","metadata":{"annotations":{},"name":"restart","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"es-quickstart"},"timeout":"3m","type":"Restart"}}
  creationTimestamp: "2025-11-11T05:02:36Z"
  generation: 1
  name: restart
  namespace: demo
  resourceVersion: "749630"
  uid: 52fe9376-cef4-4171-9ca7-8a0d1be902fb
spec:
  apply: Always
  databaseRef:
    name: es-quickstart
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2025-11-11T05:02:36Z"
    message: Elasticsearch ops request is restarting nodes
    observedGeneration: 1
    reason: Restart
    status: "True"
    type: Restart
  - lastTransitionTime: "2025-11-11T05:02:44Z"
    message: pod exists; ConditionStatus:True; PodName:es-quickstart-0
    observedGeneration: 1
    status: "True"
    type: PodExists--es-quickstart-0
  - lastTransitionTime: "2025-11-11T05:02:44Z"
    message: create es client; ConditionStatus:True; PodName:es-quickstart-0
    observedGeneration: 1
    status: "True"
    type: CreateEsClient--es-quickstart-0
  - lastTransitionTime: "2025-11-11T05:02:44Z"
    message: evict pod; ConditionStatus:True; PodName:es-quickstart-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--es-quickstart-0
  - lastTransitionTime: "2025-11-11T05:03:55Z"
    message: create es client; ConditionStatus:True
    observedGeneration: 1
    status: "True"
    type: CreateEsClient
  - lastTransitionTime: "2025-11-11T05:03:09Z"
    message: pod exists; ConditionStatus:True; PodName:es-quickstart-1
    observedGeneration: 1
    status: "True"
    type: PodExists--es-quickstart-1
  - lastTransitionTime: "2025-11-11T05:03:09Z"
    message: create es client; ConditionStatus:True; PodName:es-quickstart-1
    observedGeneration: 1
    status: "True"
    type: CreateEsClient--es-quickstart-1
  - lastTransitionTime: "2025-11-11T05:03:09Z"
    message: evict pod; ConditionStatus:True; PodName:es-quickstart-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--es-quickstart-1
  - lastTransitionTime: "2025-11-11T05:03:34Z"
    message: pod exists; ConditionStatus:True; PodName:es-quickstart-2
    observedGeneration: 1
    status: "True"
    type: PodExists--es-quickstart-2
  - lastTransitionTime: "2025-11-11T05:03:34Z"
    message: create es client; ConditionStatus:True; PodName:es-quickstart-2
    observedGeneration: 1
    status: "True"
    type: CreateEsClient--es-quickstart-2
  - lastTransitionTime: "2025-11-11T05:03:34Z"
    message: evict pod; ConditionStatus:True; PodName:es-quickstart-2
    observedGeneration: 1
    status: "True"
    type: EvictPod--es-quickstart-2
  - lastTransitionTime: "2025-11-11T05:03:59Z"
    message: Successfully restarted all nodes
    observedGeneration: 1
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - lastTransitionTime: "2025-11-11T05:03:59Z"
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
Let's port-forward the port `9200` to local machine:

```bash
$ kubectl port-forward -n demo svc/es-quickstart 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```

Now, our Elasticsearch cluster is accessible at `localhost:9200`.

**Connection information:**

- Address: `localhost:9200`
- Username:

```bash
$ kubectl get secret -n demo es-quickstart-elastic-cred -o jsonpath='{.data.username}' | base64 -d
elastic
```

- Password:

```bash
$ kubectl get secret -n demo es-quickstart-elastic-cred -o jsonpath='{.data.password}' | base64 -d
vIHoIfHn=!Z8F4gP
```

Now let's check the health of our Elasticsearch database.

```bash
$ curl -XGET -k -u 'elastic:vIHoIfHn=!Z8F4gP' "https://localhost:9200/_cluster/health?pretty"

{
  "cluster_name" : "es-quickstart",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 3,
  "number_of_data_nodes" : 3,
  "active_primary_shards" : 3,
  "active_shards" : 6,
  "relocating_shards" : 0,
  "initializing_shards" : 0,
  "unassigned_shards" : 0,
  "delayed_unassigned_shards" : 0,
  "number_of_pending_tasks" : 0,
  "number_of_in_flight_fetch" : 0,
  "task_max_waiting_in_queue_millis" : 0,
  "active_shards_percent_as_number" : 100.0
}
```

From the health information above, we can see that our Elasticsearch cluster's status is `green` which means the cluster is healthy.


## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete  Elasticsearchopsrequest -n demo restart
kubectl delete Elasticsearch -n demo es
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Detail concepts of [ElasticsearchopsRequest object](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md)
