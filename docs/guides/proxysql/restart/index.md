---
title: Restart 
menu:
  docs_{{ .version }}:
    identifier:  guides-proxysql-restart
    name: Restart ProxySQL
    parent: guides-proxysql
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart ProxySQL

KubeDB supports restarting the ProxySQL database via a `ProxySQLOpsRequest`. Restarting is useful if some pods are stuck in an unexpected phase or not functioning correctly. This tutorial will show you how to use a `ProxySQLOpsRequest` to restart ProxySQL.

## Before You Begin

- You need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install KubeDB CLI on your workstation and the KubeDB operator in your cluster by following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
> Note: YAML files used in this tutorial are stored in the [docs/examples/proxysql](https://github.com/kubedb/docs/tree/{{
< param "info.version" >}}/docs/examples/proxysql) folder in the GitHub repository kubedb/docs.

## Prepare MySQL Backend

In this tutorial we are going to set up ProxySQL using KubeDB for a MySQL Group Replication. We will use KubeDB to set up our MySQL servers.

We need to apply the following yaml to create our MySQL Group Replication
`Note`: If your `KubeDB version` is less or equal to `v2024.6.4`, You have to use `kubedb.com/v1alpha2` apiVersion.

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-server
  namespace: demo
spec:
  version: "8.4.3"
  replicas: 3
  topology:
    mode: GroupReplication
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/backends/mysql/examples/sample-mysql.yaml
mysql.kubedb.com/mysql-server created
```

Let's wait for the MySQL to be Ready.

```bash
$ kubectl get my -n demo 
NAME           VERSION   STATUS   AGE
mysql-server   8.4.3    Ready    7m6s
```
> Here you can use MariaDB or PerconXtraDB as well as backend. Have a look at other [ProxySQL backend examples](https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/backends/)

Now we are ready to deploy and test our ProxySQL server.

## Deploy ProxySQL Server

With the following yaml we are going to create our desired ProxySQL server.

`Note`: If your `KubeDB version` is less or equal to `v2024.6.4`, You have to use `kubedb.com/v1alpha2` apiVersion.

```yaml
apiVersion: kubedb.com/v1
kind: ProxySQL
metadata:
  name: mysql-proxy
  namespace: demo
spec:
  version: "2.7.3-debian"
  replicas: 3
  syncUsers: true
  backend:
    name: mysql-server
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/backends/mysql/examples/sample-proxysql.yaml
proxysql.kubedb.com/mysql-proxy created
```


Let's wait for the ProxySQL to be Ready.

```bash
$ kubectl get proxysql -n demo
NAME          VERSION        STATUS   AGE
mysql-proxy   2.7.3-debian   Ready    3m45s
``` 
## Apply Restart opsRequest
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  type: Restart
  proxyRef:
    name: mysql-proxy
  timeout: 3m
  apply: Always
```
- `spec.type` specifies the type of operation (Restart in this case).

- `spec.proxyRef` references the ProxySQL database. The OpsRequest must be created in the same namespace as the database.

- `spec.timeout` the maximum time the operator will wait for the operation to finish before marking it as failed.

- `spec.apply` determines whether to always apply the operation (Always) or only if there are changes (IfReady).


Let's create the `ProxySQLOpsRequest` CR we have shown above,

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/restart/examples/restart.yaml
proxysql.kubedb.com/restart created
```
let's see how the pods are restarting

```shell
kubectl get pods -n demo -w
NAME             READY   STATUS    RESTARTS   AGE
mysql-proxy-0    1/1     Running   0          46s
mysql-proxy-1    1/1     Running   0          21s
mysql-proxy-2    1/1     Running   0          20s
mysql-server-0   2/2     Running   0          12m
mysql-server-1   2/2     Running   0          11m
mysql-server-2   2/2     Running   0          11m
mysql-proxy-0    1/1     Running   0          3m23s
mysql-proxy-0    1/1     Terminating   0          3m23s
mysql-proxy-0    1/1     Terminating   0          3m23s

```

```shell
NAME             READY   STATUS    RESTARTS   AGE
mysql-proxy-0    1/1     Running   0          6s
mysql-proxy-1    1/1     Running   0          3m35s
mysql-proxy-2    1/1     Running   0          3m34s
mysql-server-0   2/2     Running   0          15m
mysql-server-1   2/2     Running   0          14m
mysql-server-2   2/2     Running   0          14m
mysql-proxy-2    1/1     Running   0          3m48s
mysql-proxy-2   1/1     Terminating   0          3m48s
mysql-proxy-2    1/1     Terminating   0          3m48s

```
After some time all the pods will be restarted successfully.
```shell

NAME             READY   STATUS    RESTARTS   AGE
mysql-proxy-0    1/1     Running   0          114s
mysql-proxy-1    1/1     Running   0          74s
mysql-proxy-2    1/1     Running   0          34s
mysql-server-0   2/2     Running   0          17m
mysql-server-1   2/2     Running   0          16m
mysql-server-2   2/2     Running   0          16m

```
Now let's check the status of our `ProxySQLOpsRequest` and the Yaml output of the created `ProxySQLOpsRequest` CR.
```shell
$ kubectl get Proxysqlopsrequest -n demo 
NAME      TYPE      STATUS       AGE
restart   Restart   Successful   31m

$ kubectl get Proxysqlopsrequest -n demo restart -oyaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"ProxySQLOpsRequest","metadata":{"annotations":{},"name":"restart","namespace":"demo"},"spec":{"apply":"Always","proxyRef":{"name":"mysql-proxy"},"timeout":"3m","type":"Restart"}}
  creationTimestamp: "2025-10-20T06:38:39Z"
  generation: 1
  name: restart
  namespace: demo
  resourceVersion: "104283"
  uid: e76276f1-6682-41f9-a5a2-81ed54903556
spec:
  apply: Always
  proxyRef:
    name: mysql-proxy
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2025-10-20T06:38:39Z"
    message: 'Controller has started to Progress the ProxySQLOpsRequest: demo/restart'
    observedGeneration: 1
    reason: Running
    status: "True"
    type: Running
  - lastTransitionTime: "2025-10-20T06:38:47Z"
    message: evict pod; ConditionStatus:True; PodName:mysql-proxy-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--mysql-proxy-0
  - lastTransitionTime: "2025-10-20T06:38:47Z"
    message: get pod; ConditionStatus:True; PodName:mysql-proxy-0
    observedGeneration: 1
    status: "True"
    type: GetPod--mysql-proxy-0
  - lastTransitionTime: "2025-10-20T06:39:27Z"
    message: evict pod; ConditionStatus:True; PodName:mysql-proxy-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--mysql-proxy-1
  - lastTransitionTime: "2025-10-20T06:39:27Z"
    message: get pod; ConditionStatus:True; PodName:mysql-proxy-1
    observedGeneration: 1
    status: "True"
    type: GetPod--mysql-proxy-1
  - lastTransitionTime: "2025-10-20T06:40:07Z"
    message: evict pod; ConditionStatus:True; PodName:mysql-proxy-2
    observedGeneration: 1
    status: "True"
    type: EvictPod--mysql-proxy-2
  - lastTransitionTime: "2025-10-20T06:40:07Z"
    message: get pod; ConditionStatus:True; PodName:mysql-proxy-2
    observedGeneration: 1
    status: "True"
    type: GetPod--mysql-proxy-2
  - lastTransitionTime: "2025-10-20T06:40:47Z"
    message: 'Successfully started ProxySQL pods for ProxySQLOpsRequest: demo/restart '
    observedGeneration: 1
    reason: RestartPodsSucceeded
    status: "True"
    type: Restart
  - lastTransitionTime: "2025-10-20T06:40:47Z"
    message: Controller has successfully restart the ProxySQL replicas
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
$ kubectl exec -it -n demo mysql-proxy-0 -- bash

proxysql@mysql-proxy-0:/$  mysql -uadmin -padmin -h127.0.0.1 -P6032 --prompt="ProxySQLAdmin > " 
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 120
Server version: 8.0.27 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.
```
## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete  Proxysqlopsrequest -n demo restart
kubectl delete ProxySQL -n demo mysql-proxy
kubectl delete MySQL -n demo mysql-server
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [ProxySQL object](/docs/guides/proxysql/concepts/proxysql/index.md).
- Detail concepts of [ProxySQL object](/docs/guides/proxysql/concepts/opsrequest/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md)..
