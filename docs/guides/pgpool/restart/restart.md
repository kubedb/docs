---
title: Restart Pgpool
menu:
  docs_{{ .version }}:
    identifier: pp-restart-details
    name: Restart Pgpool
    parent: pp-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart Pgpool

KubeDB supports restarting the Pgpool via a PgpoolOpsRequest. Restarting is useful if some pods are got stuck in some phase, or they are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/pgpool](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgpool) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prepare Postgres
Prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgpool/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.

## Deploy Pgpool

In this section, we are going to deploy a Pgpool using KubeDB.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pgpool
  namespace: demo
spec:
  version: "4.4.5"
  replicas: 1
  postgresRef:
    name: ha-postgres
    namespace: demo
  deletionPolicy: WipeOut
```

Let's create the `Pgpool` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/restart/pgpool.yaml
pgpool.kubedb.com/pgpool created
```

## Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgpoolOpsRequest
metadata:
  name: restart-pgpool
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: pgpool
  timeout: 3m
  apply: Always
```

- `spec.type` specifies the Type of the ops Request
- `spec.databaseRef` holds the name of the Pgpool.  The pgpool should be available in the same namespace as the opsRequest
- The meaning of `spec.timeout` & `spec.apply` fields will be found [here](/docs/guides/pgpool/concepts/opsrequest.md#spectimeout)

Let's create the `PgpoolOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/restart/ops.yaml
pgpoolopsrequest.ops.kubedb.com/restart-pgpool created
```

Now the Ops-manager operator will restart the pods one by one.

```shell
$ kubectl get ppops -n demo
NAME             TYPE      STATUS       AGE
restart-pgpool   Restart   Successful   79s

$ kubectl get ppops -n demo -oyaml restart-pgpool
apiVersion: ops.kubedb.com/v1alpha1
kind: PgpoolOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"PgpoolOpsRequest","metadata":{"annotations":{},"name":"restart-pgpool","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"pgpool"},"timeout":"3m","type":"Restart"}}
  creationTimestamp: "2024-07-18T06:49:50Z"
  generation: 1
  name: restart-pgpool
  namespace: demo
  resourceVersion: "94394"
  uid: 8d3387fc-0c21-4e14-8bed-857a7cdf5423
spec:
  apply: Always
  databaseRef:
    name: pgpool
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2024-07-18T06:49:50Z"
    message: Pgpool ops-request has started to restart pgpool nodes
    observedGeneration: 1
    reason: Restart
    status: "True"
    type: Restart
  - lastTransitionTime: "2024-07-18T06:49:53Z"
    message: Successfully paused database
    observedGeneration: 1
    reason: DatabasePauseSucceeded
    status: "True"
    type: DatabasePauseSucceeded
  - lastTransitionTime: "2024-07-18T06:50:38Z"
    message: Successfully Restarted Pgpool nodes
    observedGeneration: 1
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - lastTransitionTime: "2024-07-18T06:49:58Z"
    message: get pod; ConditionStatus:True; PodName:pgpool-0
    observedGeneration: 1
    status: "True"
    type: GetPod--pgpool-0
  - lastTransitionTime: "2024-07-18T06:49:58Z"
    message: evict pod; ConditionStatus:True; PodName:pgpool-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--pgpool-0
  - lastTransitionTime: "2024-07-18T06:50:33Z"
    message: check pod running; ConditionStatus:True; PodName:pgpool-0
    observedGeneration: 1
    status: "True"
    type: CheckPodRunning--pgpool-0
  - lastTransitionTime: "2024-07-18T06:50:38Z"
    message: Controller has successfully restart the Pgpool replicas
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful
```


## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete pgpoolopsrequest -n demo restart-pgpool
kubectl delete pgpool -n demo pgpool
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Pgpool object](/docs/guides/pgpool/concepts/pgpool.md).
- Monitor your Pgpool database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/pgpool/monitoring/using-prometheus-operator.md).
- Monitor your Pgpool database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/pgpool/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Pgpool object](/docs/guides/pgpool/concepts/pgpool.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
