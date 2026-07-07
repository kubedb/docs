---
title: Restart HanaDB
menu:
  docs_{{ .version }}:
    identifier: guides-hanadb-restart-restart
    name: Restart
    parent: guides-hanadb-restart
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart HanaDB

This guide shows how to restart a HanaDB database using a `HanaDBOpsRequest` of type `Restart`. KubeDB
performs a **rolling restart**, one pod at a time. For a System Replication cluster the **primary is
restarted last** to minimize avoidable failovers.

> Note: The YAML files used in this tutorial are stored in [docs/examples/hanadb/restart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/restart) folder in the GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- Install the KubeDB Provisioner and Ops-manager operators following the steps [here](/docs/setup/README.md).
- Create a namespace:

```bash
kubectl create ns demo
```
namespace/demo created

## Deploy a HanaDB System Replication Cluster

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hanadb-cluster
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 2
  storageType: Durable
  topology:
    mode: SystemReplication
    systemReplication:
      replicationMode: fullsync
      operationMode: logreplay_readaccess
  podTemplate:
    spec:
      containers:
      - name: hanadb
        resources:
          requests:
            cpu: "1500m"
            memory: "8Gi"
          limits:
            cpu: "4"
            memory: "14Gi"
  storage:
    accessModes: ["ReadWriteOnce"]
    resources:
      requests:
        storage: 64Gi
    storageClassName: local-path
  deletionPolicy: WipeOut
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/restart/system-replication-ops.yaml
```
hanadb.kubedb.com/hanadb-cluster created

Wait until `hanadb-cluster` is `Ready`. Note the current pod ages and which pod is primary:

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=hanadb-cluster -L kubedb.com/role
```
NAME                       READY   STATUS    RESTARTS   AGE    ROLE
hanadb-cluster-0           2/2     Running   0          14m    primary
hanadb-cluster-1           2/2     Running   0          14m    secondary
hanadb-cluster-arbiter-0   1/1     Running   0          8m     arbiter

## Create a Restart HanaDBOpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hdbops-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: hanadb-cluster
  timeout: 30m
  apply: Always
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/restart/restart.yaml
```
hanadbopsrequest.ops.kubedb.com/hdbops-restart created

Here `spec.apply: Always` lets the restart proceed even if the database is not `Ready`, which is useful
for recovering an unhealthy database.

## Verify the Restart

```bash
kubectl get hdbops -n demo hdbops-restart
```
NAME             TYPE      STATUS       AGE
hdbops-restart   Restart   Successful   7m32s

```bash
kubectl describe hdbops -n demo hdbops-restart
```
...
Status:
  Conditions:
    Message:  HanaDBOpsRequest has started to restart HanaDB nodes
    Reason:   Restart
    Status:   True
    Type:     Restart
    Message:  Successfully paused database
    Reason:   DatabasePauseSucceeded
    Status:   True
    Type:     DatabasePauseSucceeded
    Message:  Successfully restarted HanaDB nodes
    Reason:   RestartNodes
    Status:   True
    Type:     RestartNodes
    Message:  Successfully completed restart for HanaDB.
    Reason:   Successful
    Status:   True
    Type:     Successful
  Phase:      Successful

The operator evicts the secondary (`hanadb-cluster-1`) first and the primary (`hanadb-cluster-0`) last.
The pods now show a fresh age and the database is back to `Ready`. Note that restarting the old primary
triggers a normal HANA SystemReplication takeover, so the `primary`/`secondary` roles may swap:

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=hanadb-cluster -L kubedb.com/role
```
NAME                       READY   STATUS    RESTARTS   AGE     ROLE
hanadb-cluster-0           2/2     Running   0          4m30s   secondary
hanadb-cluster-1           2/2     Running   0          7m16s   primary
hanadb-cluster-arbiter-0   1/1     Running   0          15m     arbiter

```bash
kubectl get hanadb.kubedb.com -n demo hanadb-cluster
```
NAME             VERSION   STATUS   AGE
hanadb-cluster   2.0.82    Ready    22m

## Cleaning Up

```bash
kubectl delete hdbops -n demo hdbops-restart
```

```bash
kubectl delete hanadb.kubedb.com -n demo hanadb-cluster
```

```bash
kubectl delete ns demo
```

## Next Steps

- [Vertically scale](/docs/guides/hanadb/scaling/vertical-scaling/vertical-scaling.md) a HanaDB.
- Review the [HanaDBOpsRequest CRD](/docs/guides/hanadb/concepts/opsrequest.md).
