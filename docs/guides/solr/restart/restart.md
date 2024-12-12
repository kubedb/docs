---
title: Restart Solr
menu:
  docs_{{ .version }}:
    identifier: sl-restart-details
    name: Restart Solr
    parent: sl-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart Solr

KubeDB supports restarting the Solr database via a `SolrOpsRequest`. Restarting is useful if some pods are got stuck in some phase, or they are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/Solr](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/solr) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Solr

In this section, we are going to deploy a Solr database using KubeDB.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-cluster
  namespace: demo
spec:
  version: 9.6.1
  zookeeperRef:
    name: zoo
    namespace: demo
  topology:
    overseer:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    coordinator:
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi

```

- `spec.topology` represents the specification for Solr topology.
    - `data` denotes the data node of solr topology.
    - `overseer` denotes the controller node of solr topology.
    - `coordinator` denotes the controller node of solr topology

Let's create the `Solr` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Sslr/restart/solr-cluster.yaml
solr.kubedb.com/solr-cluster created
```

## Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: solr-cluster
  type: Restart
```

- `spec.type` specifies the Type of the ops Request
- `spec.databaseRef` holds the name of the Solr CR. It should be available in the same namespace as the opsRequest
- The meaning of `spec.timeout` & `spec.apply` fields will be found [here](/docs/guides/solr/concepts/solropsrequests.md#spectimeout)

> Note: The method of restarting the combined node is exactly same as above. All you need, is to specify the corresponding Solr name in `spec.databaseRef.name` section.

Let's create the `SolrOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/restart/ops.yaml
solropsrequest.ops.kubedb.com/restart created
```

Now the Ops-manager operator will first restart the controller pods, then broker of the referenced Solr.

```shell
$ kubectl get slops -n demo
NAME      TYPE      STATUS       AGE
restart   Restart   Successful   2m34s
````

```bash
$ kubectl get slops -n demo restart -oyaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"SolrOpsRequest","metadata":{"annotations":{},"name":"restart","namespace":"demo"},"spec":{"apply":"IfReady","databaseRef":{"name":"solr-cluster"},"type":"Restart"}}
  creationTimestamp: "2024-11-06T06:11:55Z"
  generation: 1
  name: restart
  namespace: demo
  resourceVersion: "1746799"
  uid: de3f03f9-512e-44d2-b29f-6c084c5d993b
spec:
  apply: IfReady
  databaseRef:
    name: solr-cluster
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2024-11-06T06:11:55Z"
    message: Solr ops-request has started to restart solr nodes
    observedGeneration: 1
    reason: Restart
    status: "True"
    type: Restart
  - lastTransitionTime: "2024-11-06T06:14:23Z"
    message: Successfully Restarted Solr nodes
    observedGeneration: 1
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - lastTransitionTime: "2024-11-06T06:12:03Z"
    message: get pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
    observedGeneration: 1
    status: "True"
    type: GetPod--solr-cluster-overseer-0
  - lastTransitionTime: "2024-11-06T06:12:03Z"
    message: evict pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--solr-cluster-overseer-0
  - lastTransitionTime: "2024-11-06T06:12:08Z"
    message: running pod; ConditionStatus:False
    observedGeneration: 1
    status: "False"
    type: RunningPod
  - lastTransitionTime: "2024-11-06T06:12:48Z"
    message: get pod; ConditionStatus:True; PodName:solr-cluster-data-0
    observedGeneration: 1
    status: "True"
    type: GetPod--solr-cluster-data-0
  - lastTransitionTime: "2024-11-06T06:12:48Z"
    message: evict pod; ConditionStatus:True; PodName:solr-cluster-data-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--solr-cluster-data-0
  - lastTransitionTime: "2024-11-06T06:13:38Z"
    message: get pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
    observedGeneration: 1
    status: "True"
    type: GetPod--solr-cluster-coordinator-0
  - lastTransitionTime: "2024-11-06T06:13:38Z"
    message: evict pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--solr-cluster-coordinator-0
  - lastTransitionTime: "2024-11-06T06:14:23Z"
    message: Controller has successfully restart the Solr replicas
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete solropsrequest -n demo restart
kubectl delete solr -n demo solr-cluster
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Solr object](/docs/guides/solr/concepts/solr.md).
- Different Solr topology clustering modes [here](/docs/guides/solr/clustering/topology_cluster.md).
- Monitor your Solr database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/solr/monitoring/prometheus-operator.md).

- Monitor your Solr database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/solr/monitoring/prometheus-builtin.md)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
