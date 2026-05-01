---
title: Restart Oracle
menu:
  docs_{{ .version }}:
    identifier: oracle-restart-overview
    name: Restart Oracle
    parent: oracle-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart Oracle

KubeDB supports restarting the Oracle database via a `OracleOpsRequest`. Restarting is useful if some pods are stuck in a non-running phase or are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Oracle](/docs/guides/oracle/concepts/oracle.md)
  - [OracleOpsRequest](/docs/guides/oracle/concepts/opsrequest.md)

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/oracle/restart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/oracle/restart) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Oracle

In this section, we are going to deploy a Oracle database using KubeDB.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle-sample
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

Let's create the `Oracle` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/restart/oracle.yaml
oracle.kubedb.com/oracle-sample created
```

Now, wait until `oracle-sample` has status `Ready`:

```bash
$ kubectl get oracle -n demo
NAME             VERSION   STATUS   AGE
oracle-sample    1.17.0    Ready    3m47s
```

## Apply Restart OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: qdops-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: oracle-sample
  timeout: 3m
  apply: Always
```

- `spec.type` specifies the type of the OpsRequest.
- `spec.databaseRef` holds the name of the Oracle database. The db should be available in the same namespace as the OpsRequest.
- The meaning of `spec.timeout` & `spec.apply` fields can be found [here](/docs/guides/oracle/concepts/opsrequest.md).

Let's create the `OracleOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/restart/ops.yaml
oracleopsrequest.ops.kubedb.com/qdops-restart created
```

Now the Ops-manager operator will restart the Oracle pods one by one, waiting for each pod to come back to `Running` state before proceeding to the next.

```bash
$ kubectl get qdops -n demo qdops-restart
NAME            TYPE      STATUS       AGE
qdops-restart   Restart   Successful   3m25s
```

```bash
$ kubectl get qdops -n demo qdops-restart -o yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: qdops-restart
  namespace: demo
spec:
  apply: Always
  databaseRef:
    name: oracle-sample
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2026-05-01T10:00:00Z"
    message: Oracle ops request is restarting nodes
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
kubectl delete oracleopsrequest -n demo qdops-restart
kubectl delete oracle -n demo oracle-sample
kubectl delete ns demo
```
