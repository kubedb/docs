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
  name: es
  namespace: demo
spec:
  version: "8.0.40"
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Elasticsearch` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/restart/yamls/es.yaml
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
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"ElasticsearchOpsRequest","metadata":{"annotations":{},"name":"restart","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"es"},"timeout":"3m","type":"Restart"}}
  creationTimestamp: "2025-10-17T05:45:40Z"
  generation: 1
  name: restart
  namespace: demo
  resourceVersion: "22350"
  uid: c6ef7130-9a31-4f64-ae49-1b4e332f0817
spec:
  apply: Always
  databaseRef:
    name: es
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2025-10-17T05:45:41Z"
    message: 'Controller has started to Progress the ElasticsearchOpsRequest: demo/restart'
    observedGeneration: 1
    reason: Running
    status: "True"
    type: Running
  - lastTransitionTime: "2025-10-17T05:45:49Z"
    message: evict pod; ConditionStatus:True; PodName:es-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--es-0
  - lastTransitionTime: "2025-10-17T05:45:49Z"
    message: get pod; ConditionStatus:True; PodName:es-0
    observedGeneration: 1
    status: "True"
    type: GetPod--es-0
  - lastTransitionTime: "2025-10-17T05:46:59Z"
    message: evict pod; ConditionStatus:True; PodName:es-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--es-1
  - lastTransitionTime: "2025-10-17T05:46:59Z"
    message: get pod; ConditionStatus:True; PodName:es-1
    observedGeneration: 1
    status: "True"
    type: GetPod--es-1
  - lastTransitionTime: "2025-10-17T05:48:09Z"
    message: evict pod; ConditionStatus:True; PodName:es-2
    observedGeneration: 1
    status: "True"
    type: EvictPod--es-2
  - lastTransitionTime: "2025-10-17T05:48:09Z"
    message: get pod; ConditionStatus:True; PodName:es-2
    observedGeneration: 1
    status: "True"
    type: GetPod--es-2
  - lastTransitionTime: "2025-10-17T05:49:19Z"
    message: 'Successfully started Elasticsearch pods for ElasticsearchOpsRequest:
      demo/restart '
    observedGeneration: 1
    reason: RestartPodsSucceeded
    status: "True"
    type: Restart
  - lastTransitionTime: "2025-10-17T05:49:19Z"
    message: Controller has successfully restart the Elasticsearch replicas
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful

```
**Verify Data Persistence**

After the restart, reconnect to the database and verify that the previously created database still exists:

```bash
$ kubectl exec -it -n demo es-0 -- mysql -u root --password='kP!VVJ2e~DUtcD*D'
Defaulted container "Elasticsearch" out of: Elasticsearch, px-coordinator, px-init (init)
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 112
Server version: 8.0.40-31.1 Percona XtraDB Cluster (GPL), Release rel31, Revision 4b32153, WSREP version 26.1.4.3

Copyright (c) 2009-2024 Percona LLC and/or its affiliates
Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> show databases;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| shastriya          |
| sys                |
+--------------------+
6 rows in set (0.02 sec)

mysql> exit
Bye
```
## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete  Elasticsearchopsrequest -n demo restart
kubectl delete Elasticsearch -n demo Elasticsearch
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/Elasticsearch/index.md).
- Detail concepts of [ElasticsearchopsRequest object](/docs/guides/elasticsearch/concepts/opsrequest/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md)..
