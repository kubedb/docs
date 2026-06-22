---
title: Restart Qdrant
menu:
  docs_{{ .version }}:
    identifier: qdrant-restart-overview
    name: Restart Qdrant
    parent: qdrant-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart Qdrant

KubeDB supports restarting the Qdrant database via a `QdrantOpsRequest`. Restarting is useful if some pods are stuck in a non-running phase or are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/qdrant/restart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/qdrant/restart) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Qdrant

In this section, we are going to deploy a Qdrant database using KubeDB.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Qdrant` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/restart/qdrant.yaml
qdrant.kubedb.com/qdrant-sample created
```

Now, wait until `qdrant-sample` has status `Ready`:

```bash
$ kubectl get qdrant -n demo
NAME             VERSION   STATUS   AGE
qdrant-sample    1.17.0    Ready    3m47s
```

## Apply Restart OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdops-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: qdrant-sample
  timeout: 3m
  apply: Always
```

- `spec.type` specifies the type of the OpsRequest.
- `spec.databaseRef` holds the name of the Qdrant database. The db should be available in the same namespace as the OpsRequest.
- The meaning of `spec.timeout` & `spec.apply` fields can be found [here](/docs/guides/qdrant/concepts/opsrequest.md).

Let's create the `QdrantOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/restart/ops-request.yaml
qdrantopsrequest.ops.kubedb.com/qdops-restart created
```

Now the Ops-manager operator will restart the Qdrant pods one by one, waiting for each pod to come back to `Running` state before proceeding to the next.

```bash
$ kubectl get qdops -n demo qdops-restart
NAME            TYPE      STATUS       AGE
qdops-restart   Restart   Successful   66s
```

```bash
$ kubectl get qdops -n demo qdops-restart -o yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"QdrantOpsRequest","metadata":{"annotations":{},"name":"qdops-restart","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"qdrant-sample"},"timeout":"3m","type":"Restart"}}
  creationTimestamp: "2026-05-15T05:43:05Z"
  generation: 1
  name: qdops-restart
  namespace: demo
  resourceVersion: "3357120"
  uid: f90d628f-1db2-4fbc-a20d-e449ab7215a8
spec:
  apply: Always
  databaseRef:
    name: qdrant-sample
  maxRetries: 1
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2026-05-15T05:43:05Z"
    message: Qdrant ops-request has started to restart Qdrant nodes
    observedGeneration: 1
    reason: Restart
    status: "True"
    type: Restart
  - lastTransitionTime: "2026-05-15T05:44:07Z"
    message: Successfully Restarted Qdrant nodes
    observedGeneration: 1
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - lastTransitionTime: "2026-05-15T05:43:17Z"
    message: get pod; ConditionStatus:True; PodName:qdrant-sample-0
    observedGeneration: 1
    status: "True"
    type: GetPod--qdrant-sample-0
  - lastTransitionTime: "2026-05-15T05:43:18Z"
    message: evict pod; ConditionStatus:True; PodName:qdrant-sample-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--qdrant-sample-0
  - lastTransitionTime: "2026-05-15T05:43:32Z"
    message: running pod; ConditionStatus:True; PodName:qdrant-sample-0
    observedGeneration: 1
    status: "True"
    type: RunningPod--qdrant-sample-0
  - lastTransitionTime: "2026-05-15T05:43:37Z"
    message: get pod; ConditionStatus:True; PodName:qdrant-sample-1
    observedGeneration: 1
    status: "True"
    type: GetPod--qdrant-sample-1
  - lastTransitionTime: "2026-05-15T05:43:38Z"
    message: evict pod; ConditionStatus:True; PodName:qdrant-sample-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--qdrant-sample-1
  - lastTransitionTime: "2026-05-15T05:43:42Z"
    message: running pod; ConditionStatus:True; PodName:qdrant-sample-1
    observedGeneration: 1
    status: "True"
    type: RunningPod--qdrant-sample-1
  - lastTransitionTime: "2026-05-15T05:43:47Z"
    message: get pod; ConditionStatus:True; PodName:qdrant-sample-2
    observedGeneration: 1
    status: "True"
    type: GetPod--qdrant-sample-2
  - lastTransitionTime: "2026-05-15T05:43:48Z"
    message: evict pod; ConditionStatus:True; PodName:qdrant-sample-2
    observedGeneration: 1
    status: "True"
    type: EvictPod--qdrant-sample-2
  - lastTransitionTime: "2026-05-15T05:44:02Z"
    message: running pod; ConditionStatus:True; PodName:qdrant-sample-2
    observedGeneration: 1
    status: "True"
    type: RunningPod--qdrant-sample-2
  - lastTransitionTime: "2026-05-15T05:44:08Z"
    message: Controller has successfully restarted the Qdrant replicas
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful
```

## Next Steps

- Learn about [backup and restore](/docs/guides/qdrant/backup/overview/index.md) Qdrant using KubeStash.
- Detail concepts of [Qdrant object](/docs/guides/qdrant/concepts/qdrant.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete qdrantopsrequest -n demo qdops-restart
qdrantopsrequest.ops.kubedb.com "qdops-restart" deleted

$ kubectl delete qdrant -n demo qdrant-sample
qdrant.kubedb.com "qdrant-sample" deleted

$ kubectl delete ns demo
namespace "demo" deleted
```
