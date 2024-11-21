---
title: Restart PgBouncer
menu:
  docs_{{ .version }}:
    identifier: pb-restart-details
    name: Restart PgBouncer
    parent: pb-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart PgBouncer

KubeDB supports restarting the PgBouncer via a PgBouncerOpsRequest. Restarting is useful if some pods are got stuck in some phase, or they are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/pgbouncer](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgbouncer) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prepare Postgres
Prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgbouncer/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.

## Deploy PgBouncer

In this section, we are going to deploy a PgBouncer using KubeDB.

```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  name: pgbouncer
  namespace: demo
spec:
  replicas: 1
  version: "1.18.0"
  database:
    syncUsers: true
    databaseName: "postgres"
    databaseRef:
      name: "ha-postgres"
      namespace: demo
  connectionPool:
    poolMode: session
    port: 5432
    reservePoolSize: 5
    maxClientConnections: 87
    defaultPoolSize: 2
    minPoolSize: 1
    authType: md5
  deletionPolicy: WipeOut
```

Let's create the `PgBouncer` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/restart/pgbouncer.yaml
pgbouncer.kubedb.com/pgbouncer created
```

## Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: restart-pgbouncer
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: pgbouncer
  timeout: 3m
  apply: Always
```

- `spec.type` specifies the Type of the ops Request
- `spec.databaseRef` holds the name of the PgBouncer.  The pgbouncer should be available in the same namespace as the opsRequest
- The meaning of `spec.timeout` & `spec.apply` fields will be found [here](/docs/guides/pgbouncer/concepts/opsrequest.md#spectimeout)

Let's create the `PgBouncerOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/restart/ops.yaml
pgbounceropsrequest.ops.kubedb.com/restart-pgbouncer created
```

Now the Ops-manager operator will restart the pods one by one.

```shell
$ kubectl get pbops -n demo
NAME                TYPE      STATUS       AGE
restart-pgbouncer   Restart   Successful   79s

$ kubectl get pbops -n demo -oyaml restart-pgbouncer
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"PgBouncerOpsRequest","metadata":{"annotations":{},"name":"restart-pgbouncer","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"pgbouncer"},"timeout":"3m","type":"Restart"}}
  creationTimestamp: "2024-07-18T06:49:50Z"
  generation: 1
  name: restart-pgbouncer
  namespace: demo
  resourceVersion: "94394"
  uid: 8d3387fc-0c21-4e14-8bed-857a7cdf5423
spec:
  apply: Always
  databaseRef:
    name: pgbouncer
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2024-07-18T06:49:50Z"
    message: PgBouncer ops-request has started to restart pgbouncer nodes
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
    message: Successfully Restarted PgBouncer nodes
    observedGeneration: 1
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - lastTransitionTime: "2024-07-18T06:49:58Z"
    message: get pod; ConditionStatus:True; PodName:pgbouncer-0
    observedGeneration: 1
    status: "True"
    type: GetPod--pgbouncer-0
  - lastTransitionTime: "2024-07-18T06:49:58Z"
    message: evict pod; ConditionStatus:True; PodName:pgbouncer-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--pgbouncer-0
  - lastTransitionTime: "2024-07-18T06:50:33Z"
    message: check pod running; ConditionStatus:True; PodName:pgbouncer-0
    observedGeneration: 1
    status: "True"
    type: CheckPodRunning--pgbouncer-0
  - lastTransitionTime: "2024-07-18T06:50:38Z"
    message: Controller has successfully restart the PgBouncer replicas
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
kubectl delete pgbounceropsrequest -n demo restart-pgbouncer
kubectl delete pgbouncer -n demo pgbouncer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [PgBouncer object](/docs/guides/pgbouncer/concepts/pgbouncer.md).
- Monitor your PgBouncer database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/pgbouncer/monitoring/using-prometheus-operator.md).
- Monitor your PgBouncer database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/pgbouncer/monitoring/using-builtin-prometheus.md).
- Detail concepts of [PgBouncer object](/docs/guides/pgbouncer/concepts/pgbouncer.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
