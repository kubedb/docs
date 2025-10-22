---
title: Restart mysql
menu:
  docs_{{ .version }}:
    identifier: mysql-restart
    name: Restart mysql
    parent: mysql-restart-details
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart MySQL

KubeDB supports restarting the MySQL database via a `MySQLOpsRequest`. Restarting is useful if some pods are got stuck in some phase, or they are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
  $ kubectl create ns demo
  namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/mysql](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mysql) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy MySQL

In this section, we are going to deploy a MySQL database using KubeDB.

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql
  namespace: demo
spec:
  version: "8.2.0"
  replicas: 3
  topology:
    mode: GroupReplication
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `mysql` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mysql/restart/mysql.yaml
mysql.kubedb.com/mysql created
```

## Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: mysql
  timeout: 3m
  apply: Always
```

- `spec.type` specifies the Type of the ops Request
- `spec.databaseRef` holds the name of the MySQL database.  The db should be available in the same namespace as the opsRequest
- The `spec.timeout` field specifies the maximum amount of time the operator will wait for the operation to complete before marking it as failed.
- The `spec.apply` field determines whether the operation should always be applied (Always) or only when there are changes (IfReady).

Let's create the `MySQLOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mysql/restart/restart.yaml
MySQLOpsRequest.ops.kubedb.com/restart created
```
In MySQL, pods follow a `primary-standby` architecture:

- `Standby` pods are restarted **first**, one by one.
- The `primary` pod is restarted last.
- During the primary pod restart, one of the standby pods is automatically promoted to primary to ensure continuous availability.

> Note: This will not restart the arbiter pod if you have one. Arbiter pod doesn't have any data related to your database. So you can ignore restarting this pod because no restart is necessary for arbiter pod but if you want so, just kubectl delete the arbiter pod (dbName-arbiter-0) in order to restart it.

```shell
$ kubectl get myops -n demo
NAME      TYPE      STATUS       AGE
restart   Restart   Successful   64m

$ kubectl get myops -n demo restart -oyaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"MySQLOpsRequest","metadata":{"annotations":{},"name":"restart","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"mysql"},"timeout":"3m","type":"Restart"}}
  creationTimestamp: "2025-10-15T11:19:36Z"
  generation: 1
  name: restart
  namespace: demo
  resourceVersion: "1031419"
  uid: 9e63a9aa-14da-432d-a972-d4e8561dcba5
spec:
  apply: Always
  databaseRef:
    name: mysql
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2025-10-15T11:19:36Z"
    message: 'Controller has started to Progress the MySQLOpsRequest: demo/restart'
    observedGeneration: 1
    reason: Running
    status: "True"
    type: Running
  - lastTransitionTime: "2025-10-15T11:19:44Z"
    message: evict pod; ConditionStatus:True
    observedGeneration: 1
    status: "True"
    type: EvictPod
  - lastTransitionTime: "2025-10-15T11:21:54Z"
    message: is pod ready; ConditionStatus:True
    observedGeneration: 1
    status: "True"
    type: IsPodReady
  - lastTransitionTime: "2025-10-15T11:21:59Z"
    message: is join in cluster; ConditionStatus:True
    observedGeneration: 1
    status: "True"
    type: IsJoinInCluster
  - lastTransitionTime: "2025-10-15T11:21:19Z"
    message: switch primary; ConditionStatus:True
    observedGeneration: 1
    status: "True"
    type: SwitchPrimary
  - lastTransitionTime: "2025-10-15T11:21:59Z"
    message: 'Successfully started MySQL pods for MySQLOpsRequest: demo/restart '
    observedGeneration: 1
    reason: RestartPodsSucceeded
    status: "True"
    type: Restart
  - lastTransitionTime: "2025-10-15T11:21:59Z"
    message: Controller has successfully restart the MySQL replicas
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
kubectl delete  myops -n demo restart
kubectl delete mysql -n demo mysql
kubectl delete ns demo
```

## Next Steps

- Learn about [backup and restore](/docs/guides/mysql/backup/kubestash/overview/index.md) mysqlQL database using Stash.
- Learn about initializing [mysqlQL with Script](/docs/guides/mysql/initialization/index.md)
- Detail concepts of [mysql object](/docs/guides/mysql/concepts/mysqldatabase/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
