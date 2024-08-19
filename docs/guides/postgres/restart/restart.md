---
title: Restart Postgres
menu:
  docs_{{ .version }}:
    identifier: pg-restart-details
    name: Restart Postgres
    parent: pg-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart Postgres

KubeDB supports restarting the Postgres database via a PostgresOpsRequest. Restarting is useful if some pods are got stuck in some phase, or they are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/postgres](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/postgres) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Postgres

In this section, we are going to deploy a Postgres database using KubeDB.

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: ha-postgres
  namespace: demo
spec:
  replicas: 3
  storageType: Durable
  deletionPolicy: WipeOut
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  version: "13.13"
```

Let's create the `Postgres` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/restart/postgres.yaml
postgres.kubedb.com/ha-postgres created
```

## Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: ha-postgres
  timeout: 3m
  apply: Always
```

- `spec.type` specifies the Type of the ops Request
- `spec.databaseRef` holds the name of the Postgres database.  The db should be available in the same namespace as the opsRequest
- The meaning of `spec.timeout` & `spec.apply` fields can be found [here](/docs/guides/postgres/concepts/opsrequest.md)

> Note: The method of restarting the standalone & cluster mode db is exactly same as above. All you need, is to specify the corresponding Postgres name in `spec.databaseRef.name` section.

Let's create the `PostgresOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/restart/ops.yaml
postgresopsrequest.ops.kubedb.com/restart created
```

Now the Ops-manager operator will first restart the general secondary pods and lastly will restart the Primary pod of the database.
> Note: This will not restart the arbiter pod if you have one. Arbiter pod doesn't have any data related to your database. So you can ignore restarting this pod because no restart is necessary for arbiter pod but if you want so, just kubectl delete the arbiter pod (dbName-arbiter-0) in order to restart it.

```shell
$ kubectl get pgops -n demo restart 
NAME      TYPE      STATUS       AGE
restart   Restart   Successful   3m25s


$ kubectl get pgops -n demo restart -oyaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"PostgresOpsRequest","metadata":{"annotations":{},"name":"restart","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"ha-postgres"},"timeout":"3m","type":"Restart"}}
  creationTimestamp: "2024-08-16T10:24:22Z"
  generation: 1
  name: restart
  namespace: demo
  resourceVersion: "568540"
  uid: dc829c3c-81fb-4da3-b83d-a2c2f09fa73b
spec:
  apply: Always
  databaseRef:
    name: ha-postgres
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2024-08-16T10:24:22Z"
    message: Postgres ops request is restarting nodes
    observedGeneration: 1
    reason: Restart
    status: "True"
    type: Restart
  - lastTransitionTime: "2024-08-16T10:24:22Z"
    message: successfully resumed pg-coordinator
    observedGeneration: 1
    reason: ResumePGCoordinator
    status: "True"
    type: ResumePGCoordinator
  - lastTransitionTime: "2024-08-16T10:26:11Z"
    message: Successfully restarted all nodes
    observedGeneration: 1
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - lastTransitionTime: "2024-08-16T10:24:31Z"
    message: evict pod; ConditionStatus:True
    observedGeneration: 1
    status: "True"
    type: EvictPod
  - lastTransitionTime: "2024-08-16T10:24:31Z"
    message: check pod ready; ConditionStatus:False; PodName:ha-postgres-1
    observedGeneration: 1
    status: "False"
    type: CheckPodReady--ha-postgres-1
  - lastTransitionTime: "2024-08-16T10:25:05Z"
    message: check pod ready; ConditionStatus:True
    observedGeneration: 1
    status: "True"
    type: CheckPodReady
  - lastTransitionTime: "2024-08-16T10:25:05Z"
    message: check replica func; ConditionStatus:False; PodName:ha-postgres-1
    observedGeneration: 1
    status: "False"
    type: CheckReplicaFunc--ha-postgres-1
  - lastTransitionTime: "2024-08-16T10:25:10Z"
    message: check replica func; ConditionStatus:True
    observedGeneration: 1
    status: "True"
    type: CheckReplicaFunc
  - lastTransitionTime: "2024-08-16T10:25:10Z"
    message: get primary; ConditionStatus:True
    observedGeneration: 1
    status: "True"
    type: GetPrimary
  - lastTransitionTime: "2024-08-16T10:25:11Z"
    message: transfer leader; ConditionStatus:True
    observedGeneration: 1
    status: "True"
    type: TransferLeader
  - lastTransitionTime: "2024-08-16T10:25:16Z"
    message: transfer leader for failover; ConditionStatus:True
    observedGeneration: 1
    status: "True"
    type: TransferLeaderForFailover
  - lastTransitionTime: "2024-08-16T10:25:16Z"
    message: check is master; ConditionStatus:True
    observedGeneration: 1
    status: "True"
    type: CheckIsMaster
  - lastTransitionTime: "2024-08-16T10:25:16Z"
    message: failover is done successfully
    observedGeneration: 1
    reason: FailoverDone
    status: "True"
    type: FailoverDone
  - lastTransitionTime: "2024-08-16T10:25:16Z"
    message: update ops request; ConditionStatus:True
    observedGeneration: 1
    status: "True"
    type: UpdateOpsRequest
  - lastTransitionTime: "2024-08-16T10:25:16Z"
    message: check pod ready; ConditionStatus:False; PodName:ha-postgres-0
    observedGeneration: 1
    status: "False"
    type: CheckPodReady--ha-postgres-0
  - lastTransitionTime: "2024-08-16T10:26:00Z"
    message: check replica func; ConditionStatus:False; PodName:ha-postgres-0
    observedGeneration: 1
    status: "False"
    type: CheckReplicaFunc--ha-postgres-0
  - lastTransitionTime: "2024-08-16T10:26:11Z"
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
kubectl delete postgresopsrequest -n demo restart
kubectl delete postgres -n demo ha-postgres
kubectl delete ns demo
```

## Next Steps

- Learn about [backup and restore](/docs/guides/postgres/backup/overview/index.md) PostgreSQL database using Stash.
- Learn about initializing [PostgreSQL with Script](/docs/guides/postgres/initialization/script_source.md).
- Learn about [custom PostgresVersions](/docs/guides/postgres/custom-versions/setup.md).
- Want to setup PostgreSQL cluster? Check how to [configure Highly Available PostgreSQL Cluster](/docs/guides/postgres/clustering/ha_cluster.md)
- Monitor your PostgreSQL database with KubeDB using [built-in Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md).
- Monitor your PostgreSQL database with KubeDB using [Prometheus operator](/docs/guides/postgres/monitoring/using-prometheus-operator.md).
- Detail concepts of [Postgres object](/docs/guides/postgres/concepts/postgres.md).
- Use [private Docker registry](/docs/guides/postgres/private-registry/using-private-registry.md) to deploy PostgreSQL with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
