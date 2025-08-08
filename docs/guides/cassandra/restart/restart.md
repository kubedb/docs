---
title: Restart Cassandra
menu:
  docs_{{ .version }}:
    identifier: cas-restart-details
    name: Restart Cassandra
    parent: cas-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart Cassandra

KubeDB supports restarting the Cassandra database via a CassandraOpsRequest. Restarting is useful if some pods are got stuck in some phase, or they are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/cassandra](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/cassandra) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Cassandra

In this section, we are going to deploy a Cassandra database using KubeDB.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:
  name: cassandra-prod
  namespace: demo
spec:
  version: 5.0.3
  configSecret:
    name: cas-configuration
  topology:
    rack:
      - name: r0
        replicas: 2
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
        storageType: Durable
  deletionPolicy: WipeOut
```

Let's create the `Cassandra` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/restart/cassandra.yaml
cassandra.kubedb.com/cassandra-prod created
```

## Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: cassandra-prod
  timeout: 3m
  apply: Always
```

- `spec.type` specifies the Type of the ops Request
- `spec.databaseRef` holds the name of the Cassandra CR. It should be available in the same namespace as the opsRequest
- The meaning of `spec.timeout` & `spec.apply` fields will be found [here](/docs/guides/cassandra/concepts/cassandraopsrequest.md#spectimeout)

> Note: The method of restarting the combined node is exactly same as above. All you need, is to specify the corresponding Cassandra name in `spec.databaseRef.name` section.

Let's create the `CassandraOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/restart/ops.yaml
cassandraopsrequest.ops.kubedb.com/restart created
```

Now the Ops-manager operator will first restart the controller pods, then broker of the referenced cassandra.

```shell
$ kubectl get casops -n demo
NAME      TYPE      STATUS       AGE
restart   Restart   Successful   119s

$ kubectl get casops -n demo restart -oyaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"CassandraOpsRequest","metadata":{"annotations":{},"name":"restart","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"cassandra-prod"},"timeout":"3m","type":"Restart"}}
  creationTimestamp: "2025-07-26T10:12:10Z"
  generation: 1
  name: restart
  namespace: demo
  resourceVersion: "24434"
  uid: 956a374e-1d6f-4f68-828f-cfed4410b175
spec:
  apply: Always
  databaseRef:
    name: cassandra-prod
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2025-07-26T10:12:10Z"
    message: Cassandra ops-request has started to restart cassandra nodes
    observedGeneration: 1
    reason: Restart
    status: "True"
    type: Restart
  - lastTransitionTime: "2025-07-26T10:12:18Z"
    message: get pod; ConditionStatus:True; PodName:cassandra-prod-controller-0
    observedGeneration: 1
    status: "True"
    type: GetPod--cassandra-prod-controller-0
  - lastTransitionTime: "2025-07-26T10:12:18Z"
    message: evict pod; ConditionStatus:True; PodName:cassandra-prod-controller-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--cassandra-prod-controller-0
  - lastTransitionTime: "2025-07-26T10:12:23Z"
    message: check pod running; ConditionStatus:True; PodName:cassandra-prod-controller-0
    observedGeneration: 1
    status: "True"
    type: CheckPodRunning--cassandra-prod-controller-0
  - lastTransitionTime: "2025-07-26T10:12:28Z"
    message: get pod; ConditionStatus:True; PodName:cassandra-prod-controller-1
    observedGeneration: 1
    status: "True"
    type: GetPod--cassandra-prod-controller-1
  - lastTransitionTime: "2025-07-26T10:12:28Z"
    message: evict pod; ConditionStatus:True; PodName:cassandra-prod-controller-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--cassandra-prod-controller-1
  - lastTransitionTime: "2025-07-26T10:12:38Z"
    message: check pod running; ConditionStatus:True; PodName:cassandra-prod-controller-1
    observedGeneration: 1
    status: "True"
    type: CheckPodRunning--cassandra-prod-controller-1
  - lastTransitionTime: "2025-07-26T10:12:43Z"
    message: get pod; ConditionStatus:True; PodName:cassandra-prod-broker-0
    observedGeneration: 1
    status: "True"
    type: GetPod--cassandra-prod-broker-0
  - lastTransitionTime: "2025-07-26T10:12:43Z"
    message: evict pod; ConditionStatus:True; PodName:cassandra-prod-broker-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--cassandra-prod-broker-0
  - lastTransitionTime: "2025-07-26T10:13:18Z"
    message: check pod running; ConditionStatus:True; PodName:cassandra-prod-broker-0
    observedGeneration: 1
    status: "True"
    type: CheckPodRunning--cassandra-prod-broker-0
  - lastTransitionTime: "2025-07-26T10:13:23Z"
    message: get pod; ConditionStatus:True; PodName:cassandra-prod-broker-1
    observedGeneration: 1
    status: "True"
    type: GetPod--cassandra-prod-broker-1
  - lastTransitionTime: "2025-07-26T10:13:23Z"
    message: evict pod; ConditionStatus:True; PodName:cassandra-prod-broker-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--cassandra-prod-broker-1
  - lastTransitionTime: "2025-07-26T10:13:28Z"
    message: check pod running; ConditionStatus:True; PodName:cassandra-prod-broker-1
    observedGeneration: 1
    status: "True"
    type: CheckPodRunning--cassandra-prod-broker-1
  - lastTransitionTime: "2025-07-26T10:13:33Z"
    message: Successfully Restarted Cassandra nodes
    observedGeneration: 1
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - lastTransitionTime: "2025-07-26T10:13:33Z"
    message: Controller has successfully restart the Cassandra replicas
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
kubectl delete cassandraopsrequest -n demo restart
kubectl delete cassandra -n demo cassandra-prod
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Cassandra object](/docs/guides/cassandra/concepts/cassandra.md).
- Monitor your Cassandra database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/cassandra/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Cassandra database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/cassandra/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
