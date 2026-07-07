---
title: Restart Weaviate
menu:
  docs_{{ .version }}:
    identifier: weaviate-restart-overview
    name: Restart Weaviate
    parent: weaviate-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart Weaviate

KubeDB supports restarting the Weaviate database via a `WeaviateOpsRequest`. Restarting is useful if some pods are stuck in a non-running phase or are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [Weaviate Quickstart](/docs/guides/weaviate/quickstart/quickstart.md)

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
kubectl create ns demo
```
namespace/demo created

> **Note:** YAML files used in this tutorial are stored in [docs/examples/weaviate/restart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/weaviate/restart) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Weaviate

In this section, we are going to deploy a Weaviate database using KubeDB.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-sample
  namespace: demo
spec:
  version: 1.33.1
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Weaviate` CR we have shown above,

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/restart/weaviate.yaml
```
weaviate.kubedb.com/weaviate-sample created

Now, wait until `weaviate-sample` has status `Ready`:

```bash
kubectl get weaviate -n demo
```
NAME              TYPE                  VERSION   STATUS   AGE
weaviate-sample   kubedb.com/v1alpha2   1.33.1    Ready    5m

## Apply Restart OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: weaviate-sample
  timeout: 3m
  apply: Always
```

- `spec.type` specifies the type of the OpsRequest.
- `spec.databaseRef` holds the name of the Weaviate database. The db should be available in the same namespace as the OpsRequest.
- `spec.timeout` is the maximum time the ops-manager waits for the request to complete before it is marked as failed.
- `spec.apply` controls when the operation is applied. `Always` applies it even if the database is not `Ready`; `IfReady` applies it only when the database is `Ready`.

Let's create the `WeaviateOpsRequest` CR we have shown above,

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/restart/ops-request.yaml
```
weaviateopsrequest.ops.kubedb.com/restart created

Now the Ops-manager operator will restart the Weaviate pods one by one, waiting for each pod to come back to `Running` state before proceeding to the next.

```bash
kubectl get weaviateopsrequest -n demo restart
```
NAME      TYPE      STATUS       AGE
restart   Restart   Successful   92s

```bash
kubectl get weaviateopsrequest -n demo restart -o yaml
```
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  creationTimestamp: "2026-06-30T17:32:30Z"
  generation: 1
  name: restart
  namespace: demo
  resourceVersion: "62408"
  uid: f47ef99e-0a20-46d8-abf6-4dc40c65e2ff
spec:
  apply: Always
  databaseRef:
    name: weaviate-sample
  maxRetries: 1
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2026-06-30T17:32:30Z"
    message: Weaviate ops-request has started to restart ops nodes
    observedGeneration: 1
    reason: Restart
    status: "True"
    type: Restart
  - lastTransitionTime: "2026-06-30T17:32:42Z"
    message: get pod; ConditionStatus:True; PodName:weaviate-sample-0
    observedGeneration: 1
    status: "True"
    type: GetPod--weaviate-sample-0
  - lastTransitionTime: "2026-06-30T17:32:43Z"
    message: evict pod; ConditionStatus:True; PodName:weaviate-sample-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--weaviate-sample-0
  - lastTransitionTime: "2026-06-30T17:32:57Z"
    message: running pod; ConditionStatus:True; PodName:weaviate-sample-0
    observedGeneration: 1
    status: "True"
    type: RunningPod--weaviate-sample-0
  - lastTransitionTime: "2026-06-30T17:33:02Z"
    message: get pod; ConditionStatus:True; PodName:weaviate-sample-1
    observedGeneration: 1
    status: "True"
    type: GetPod--weaviate-sample-1
  - lastTransitionTime: "2026-06-30T17:33:02Z"
    message: evict pod; ConditionStatus:True; PodName:weaviate-sample-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--weaviate-sample-1
  - lastTransitionTime: "2026-06-30T17:33:17Z"
    message: running pod; ConditionStatus:True; PodName:weaviate-sample-1
    observedGeneration: 1
    status: "True"
    type: RunningPod--weaviate-sample-1
  - lastTransitionTime: "2026-06-30T17:33:22Z"
    message: get pod; ConditionStatus:True; PodName:weaviate-sample-2
    observedGeneration: 1
    status: "True"
    type: GetPod--weaviate-sample-2
  - lastTransitionTime: "2026-06-30T17:33:22Z"
    message: evict pod; ConditionStatus:True; PodName:weaviate-sample-2
    observedGeneration: 1
    status: "True"
    type: EvictPod--weaviate-sample-2
  - lastTransitionTime: "2026-06-30T17:33:37Z"
    message: running pod; ConditionStatus:True; PodName:weaviate-sample-2
    observedGeneration: 1
    status: "True"
    type: RunningPod--weaviate-sample-2
  - lastTransitionTime: "2026-06-30T17:33:42Z"
    message: Successfully Restarted Weaviate nodes
    observedGeneration: 1
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - lastTransitionTime: "2026-06-30T17:33:43Z"
    message: Controller has successfully restart the Weaviate replicas
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful

## Next Steps

- Detail concepts of [Weaviate object](/docs/guides/weaviate/concepts/weaviate.md).
- Learn about [Reconfigure](/docs/guides/weaviate/reconfigure/reconfigure.md) and other day-2 operations.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete weaviateopsrequest -n demo restart
```

```bash
kubectl delete weaviate -n demo weaviate-sample
```

```bash
kubectl delete ns demo
```
