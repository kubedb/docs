---
title: PerconaXtraDB Restart
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-restart
    name: Restart
    parent: guides-perconaxtradb
    weight: 47
menu_name: docs_{{ .version }}
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart PerconaXtraDB

KubeDB supports restarting a PerconaXtraDB database using a `PerconaXtraDBOpsRequest`. Restarting can be
useful if some pods are stuck in a certain state or not functioning correctly.  

This guide will demonstrate how to restart a PerconaXtraDB cluster using an OpsRequest.

---

## Before You Begin

- You need a running Kubernetes cluster and a properly configured `kubectl` command-line tool. If you don’t have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install the KubeDB CLI on your workstation and the KubeDB operator in your cluster by following the [installation steps](/docs/setup/README.md).

- For better isolation, this tutorial uses a separate namespace called `demo`:

```bash
kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/PerconaXtraDB](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/PerconaXtraDB) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy PerconaXtraDB

In this section, we are going to deploy a PerconaXtraDB database using KubeDB.

```yaml
apiVersion: kubedb.com/v1
kind: PerconaXtraDB
metadata:
  name: pxc
  namespace: demo
spec:
  version: "8.0.40"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `PerconaXtraDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/restart/yamls/pxc.yaml
PerconaXtraDB.kubedb.com/PerconaXtraDB created
```
let's wait until all pods are in the `Running` state,

```shell
kubectl get pods -n demo
NAME    READY   STATUS    RESTARTS   AGE
pxc-0   2/2     Running   0          6m28s
pxc-1   2/2     Running   0          6m28s
pxc-2   2/2     Running   0          6m28s
```
let's check database is ready to accept connections,

```bash
$ kubectl get secrets -n demo pxc-auth -o jsonpath='{.data.\username}' | base64 -d
root⏎                                                                                         banusree@bonusree-datta-PC ~> kubectl get secrets -n demo pxc-auth -o jsonpath='{.data.\password}' | base64 -d
kP!VVJ2e~DUtcD*D⏎                                                                             banusree@bonusree-datta-PC ~> kubectl exec -it -n demo sample-pxc-0 -- mysql -u root --password='kP!VVJ2e~DUtcD*D'
Error from server (NotFound): pods "sample-pxc-0" not found
banusree@bonusree-datta-PC ~ [1]> kubectl exec -it -n demo pxc-0 -- mysql -u root --password='kP!VVJ2e~DUtcD*D'
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 651
Server version: 8.0.40-31.1 Percona XtraDB Cluster (GPL), Release rel31, Revision 4b32153, WSREP version 26.1.4.3

Copyright (c) 2009-2024 Percona LLC and/or its affiliates
Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> show databases;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
5 rows in set (0.01 sec)

mysql> CREATE DATABASE shastriya;
Query OK, 1 row affected (0.02 sec)

mysql> exit                      
Bye
```


# Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: pxc
  timeout: 3m
  apply: Always
```

Here,

- `spec.type` specifies the type of operation (Restart in this case).

- `spec.databaseRef` references the PerconaXtraDB database. The OpsRequest must be created in the same namespace as the database.

- `spec.timeout` the maximum time the operator will wait for the operation to finish before marking it as failed.

- `spec.apply` determines whether to always apply the operation (Always) or only if there are changes (IfReady).

Let's create the `PerconaXtraDBOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/restart/yamls/restart.yaml
PerconaXtraDBOpsRequest.ops.kubedb.com/restart created
```

In a PerconaXtraDB cluster, all pods act as primary nodes. When you apply a restart OpsRequest, the KubeDB operator will restart the pods sequentially, one by one, to maintain cluster availability.

Let's watch the rolling restart process with:
```shell
NAME    READY   STATUS        RESTARTS   AGE
pxc-0   2/2     Terminating   0          56m
pxc-1   2/2     Running       0          55m
pxc-2   2/2     Running       0          54m
```

```shell
NAME    READY   STATUS        RESTARTS    AGE
pxc-0   2/2     Running        0          112s
pxc-1   2/2     Terminating    0          55m
pxc-2   2/2     Running        0          56m

```
```shell
NAME    READY   STATUS        RESTARTS   AGE
pxc-0   2/2     Running       0          112s
pxc-1   2/2     Running       0          42s
pxc-2   2/2     Terminating   0          56m

```
> Note: The arbiter pod (if any) is not restarted by the operator. The arbiter doesn’t store any database data, so it doesn’t require a restart. If you want to restart it manually, simply run kubectl delete pod <db-name>-arbiter-0.

```shell
$ kubectl get PerconaXtraDBopsrequest -n demo
NAME      TYPE      STATUS       AGE
restart   Restart   Successful   64m

$ kubectl get PerconaXtraDBopsrequest -n demo restart -oyaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"PerconaXtraDBOpsRequest","metadata":{"annotations":{},"name":"restart","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"pxc"},"timeout":"3m","type":"Restart"}}
  creationTimestamp: "2025-10-17T05:45:40Z"
  generation: 1
  name: restart
  namespace: demo
  resourceVersion: "22350"
  uid: c6ef7130-9a31-4f64-ae49-1b4e332f0817
spec:
  apply: Always
  databaseRef:
    name: pxc
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2025-10-17T05:45:41Z"
    message: 'Controller has started to Progress the PerconaXtraDBOpsRequest: demo/restart'
    observedGeneration: 1
    reason: Running
    status: "True"
    type: Running
  - lastTransitionTime: "2025-10-17T05:45:49Z"
    message: evict pod; ConditionStatus:True; PodName:pxc-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--pxc-0
  - lastTransitionTime: "2025-10-17T05:45:49Z"
    message: get pod; ConditionStatus:True; PodName:pxc-0
    observedGeneration: 1
    status: "True"
    type: GetPod--pxc-0
  - lastTransitionTime: "2025-10-17T05:46:59Z"
    message: evict pod; ConditionStatus:True; PodName:pxc-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--pxc-1
  - lastTransitionTime: "2025-10-17T05:46:59Z"
    message: get pod; ConditionStatus:True; PodName:pxc-1
    observedGeneration: 1
    status: "True"
    type: GetPod--pxc-1
  - lastTransitionTime: "2025-10-17T05:48:09Z"
    message: evict pod; ConditionStatus:True; PodName:pxc-2
    observedGeneration: 1
    status: "True"
    type: EvictPod--pxc-2
  - lastTransitionTime: "2025-10-17T05:48:09Z"
    message: get pod; ConditionStatus:True; PodName:pxc-2
    observedGeneration: 1
    status: "True"
    type: GetPod--pxc-2
  - lastTransitionTime: "2025-10-17T05:49:19Z"
    message: 'Successfully started PerconaXtraDB pods for PerconaXtraDBOpsRequest:
      demo/restart '
    observedGeneration: 1
    reason: RestartPodsSucceeded
    status: "True"
    type: Restart
  - lastTransitionTime: "2025-10-17T05:49:19Z"
    message: Controller has successfully restart the PerconaXtraDB replicas
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful

```
**Verify Data Persistence**

After the restart, reconnect to the database and verify that the previously created database still exists:

```bash
$ kubectl exec -it -n demo pxc-0 -- mysql -u root --password='kP!VVJ2e~DUtcD*D'
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 112
Server version: 8.0.40-31.1 Percona XtraDB Cluster (GPL), Release rel31, Revision 4b32153, WSREP version 26.1.4.3

Copyright (c) 2009-2024 Percona LLC and/or its affiliates
Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> show databases;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| shastriya          |
| sys                |
+--------------------+
6 rows in set (0.02 sec)

mysql> exit
Bye
```
## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete  PerconaXtraDBopsrequest -n demo restart
kubectl delete PerconaXtraDB -n demo PerconaXtraDB
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [PerconaXtraDB object](/docs/guides/percona-xtradb/concepts/perconaxtradb/index.md).
- Detail concepts of [PerconaXtraDBopsRequest object](/docs/guides/percona-xtradb/concepts/opsrequest/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md)..
