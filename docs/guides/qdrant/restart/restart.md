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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/restart/ops.yaml
qdrantopsrequest.ops.kubedb.com/qdops-restart created
```

Now the Ops-manager operator will restart the Qdrant pods one by one, waiting for each pod to come back to `Running` state before proceeding to the next.

```bash
$ kubectl get qdops -n demo qdops-restart
NAME            TYPE      STATUS       AGE
qdops-restart   Restart   Successful   3m25s
```

```bash
$ kubectl get qdops -n demo qdops-restart -o yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdops-restart
  namespace: demo
spec:
  apply: Always
  databaseRef:
    name: qdrant-sample
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2026-05-01T10:00:00Z"
    message: Qdrant ops request is restarting nodes
    observedGeneration: 1
    reason: Restart
    status: "True"
    type: Restart
  - lastTransitionTime: "2026-05-01T10:02:30Z"
    message: Successfully restarted all nodes
    observedGeneration: 1
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - lastTransitionTime: "2026-05-01T10:00:10Z"
    message: evict pod; ConditionStatus:True
    observedGeneration: 1
    status: "True"
    type: EvictPod
  - lastTransitionTime: "2026-05-01T10:01:05Z"
    message: check pod ready; ConditionStatus:True
    observedGeneration: 1
    status: "True"
    type: CheckPodReady
  - lastTransitionTime: "2026-05-01T10:02:30Z"
    message: Successfully completed the modification process.
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
kubectl delete qdrantopsrequest -n demo qdops-restart
kubectl delete qdrant -n demo qdrant-sample
kubectl delete ns demo
```
