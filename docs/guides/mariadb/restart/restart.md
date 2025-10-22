---
title: Restart MariaDB
menu:
  docs_{{ .version }}:
    identifier: mariadb-restart
    name: Restart MariaDB
    parent: mariadb-restart-details
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart MariaDB

KubeDB supports restarting the MariaDB database via a `MariaDBOpsRequest`. Restarting is useful if some pods are got stuck in some phase, or they are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
  $ kubectl create ns demo
  namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/MariaDB](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mariadb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy MariaDB

In this section, we are going to deploy a MariaDB database using KubeDB.

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: mariadb
  namespace: demo
spec:
  version: "10.5.23"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: local-path
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `MariaDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/MariaDB/restart/MariaDB.yaml
MariaDB.kubedb.com/mariadb created
```

## Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: mariadb
  timeout: 3m
  apply: Always
```

- `spec.type` specifies the Type of the ops Request
- `spec.databaseRef` holds the name of the MariaDB database.  The db should be available in the same namespace as the opsRequest
- The `spec.timeout` field specifies the maximum amount of time the operator will wait for the operation to complete before marking it as failed. 
- The `spec.apply` field determines whether the operation should always be applied (Always) or only when there are changes (IfReady).


> Note: The method of restarting the standalone & cluster mode db is exactly same as above. All you need, is to specify the corresponding Postgres name in `spec.databaseRef.name` section.

Let's create the `PostgresOpsRequest` CR we have shown above,


```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/restart/ops.yaml
MariaDBopsrequest.ops.kubedb.com/restart created
```

In `MariaDB`, all pods act as primary, so the Ops-manager operator will restart the pods one by one in sequence.
> Note: This will not restart the arbiter pod if you have one. Arbiter pod doesn't have any data related to your database. So you can ignore restarting this pod because no restart is necessary for arbiter pod but if you want so, just kubectl delete the arbiter pod (dbName-arbiter-0) in order to restart it.

```shell
$ kubectl get mariaops -n demo restart 
NAME      TYPE      STATUS       AGE
restart   Restart   Successful   3m25s


$ kubectl get mariaops -n demo restart -oyaml
kubectl get mariaops -n demo restart -oyaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"MariaDBOpsRequest","metadata":{"annotations":{},"name":"restart","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"mariadb"},"timeout":"3m","type":"Restart"}}
  creationTimestamp: "2025-10-14T12:31:09Z"
  generation: 1
  name: restart
  namespace: demo
  resourceVersion: "1004302"
  uid: 4e135330-46c4-41cd-aff6-bf2ddb018911
spec:
  apply: Always
  databaseRef:
    name: mariadb
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2025-10-14T12:31:09Z"
    message: 'Controller has started to Progress the MariaDBOpsRequest: demo/restart'
    observedGeneration: 1
    reason: Running
    status: "True"
    type: Running
  - lastTransitionTime: "2025-10-14T12:31:17Z"
    message: evict pod; ConditionStatus:True; PodName:mariadb-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--mariadb-0
  - lastTransitionTime: "2025-10-14T12:31:17Z"
    message: get pod; ConditionStatus:True; PodName:mariadb-0
    observedGeneration: 1
    status: "True"
    type: GetPod--mariadb-0
  - lastTransitionTime: "2025-10-14T12:32:27Z"
    message: evict pod; ConditionStatus:True; PodName:mariadb-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--mariadb-1
  - lastTransitionTime: "2025-10-14T12:32:27Z"
    message: get pod; ConditionStatus:True; PodName:mariadb-1
    observedGeneration: 1
    status: "True"
    type: GetPod--mariadb-1
  - lastTransitionTime: "2025-10-14T12:33:37Z"
    message: evict pod; ConditionStatus:True; PodName:mariadb-2
    observedGeneration: 1
    status: "True"
    type: EvictPod--mariadb-2
  - lastTransitionTime: "2025-10-14T12:33:37Z"
    message: get pod; ConditionStatus:True; PodName:mariadb-2
    observedGeneration: 1
    status: "True"
    type: GetPod--mariadb-2
  - lastTransitionTime: "2025-10-14T12:34:47Z"
    message: 'Successfully started MariaDB pods for MariaDBOpsRequest: demo/restart '
    observedGeneration: 1
    reason: RestartPodsSucceeded
    status: "True"
    type: Restart
  - lastTransitionTime: "2025-10-14T12:34:47Z"
    message: Controller has successfully restart the MariaDB replicas
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
kubectl delete  mariaops -n demo restart
kubectl delete mariadb -n demo mariadb
kubectl delete ns demo
```

## Next Steps

- Learn about [backup and restore](/docs/guides/mariadb/backup/kubestash/overview/index.md) MariaDBQL database using Stash.
- Learn about initializing [MariaDBQL with Script](/docs/guides/mariadb/initialization/using-script/index.md)
- Detail concepts of [MariaDB object](/docs/guides/mariadb/concepts/mariadb/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
