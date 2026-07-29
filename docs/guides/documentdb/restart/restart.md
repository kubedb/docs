---
title: Restart DocumentDB
menu:
  docs_{{ .version }}:
    identifier: dc-restart-details
    name: Restart DocumentDB
    parent: dc-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart DocumentDB

KubeDB supports restarting every pod of a `DocumentDB` database through a
`DocumentDBOpsRequest` of type `Restart`. This is useful after a node-level change, to pick up
a rotated certificate, or simply to clear transient state without deleting the database. The
operator drains and recreates the pods one at a time, always keeping a Raft leader available,
so the MongoDB wire endpoint (port `10260`) stays serviceable throughout.

## Before You Begin

- You need a Kubernetes cluster and the `kubectl` CLI configured to talk to it. If you do not
  have one, create it with [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).
- Install KubeDB following the steps [here](/docs/setup/README.md).
- This tutorial uses a namespace called `demo`:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/documentdb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/documentdb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy a DocumentDB cluster

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: documentdb-cls-sample
  namespace: demo
spec:
  version: 'pg17-0.109.0'
  storageType: Durable
  deletionPolicy: Delete
  replicas: 3
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
```

```bash
$ kubectl apply -f cluster.yaml
documentdb.kubedb.com/documentdb-cls-sample created
```

The cluster has three pods — one Raft `primary` and two `standby` replicas. Each pod runs `2/2`
containers: the `documentdb` engine and the `documentdb-coordinator` (the Raft member that
participates in leader election):

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=documentdb-cls-sample -L kubedb.com/role
NAME                      READY   STATUS    RESTARTS   AGE     ROLE
documentdb-cls-sample-0   2/2     Running   0          2m23s   primary
documentdb-cls-sample-1   2/2     Running   0          2m      standby
documentdb-cls-sample-2   2/2     Running   0          93s     standby
```

## Create the Restart OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DocumentDBOpsRequest
metadata:
  name: documentdb-cls-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: documentdb-cls-sample
```

- `spec.type` specifies the type of the OpsRequest.
- `spec.databaseRef` holds the name of the `DocumentDB` (it must be in the same namespace).

```bash
$ kubectl apply -f cluster-restart.yaml
documentdbopsrequest.ops.kubedb.com/documentdb-cls-restart created
```

Watch the OpsRequest until it reports `Successful` (`dcops` is the short name for
`DocumentDBOpsRequest`; `docdb` is the short name for `DocumentDB`):

```bash
$ kubectl get dcops -n demo documentdb-cls-restart -w
NAME                     TYPE      STATUS        AGE
documentdb-cls-restart   Restart   Progressing   20s
documentdb-cls-restart   Restart   Successful    3m52s
```

## What happened

The operator restarts the **standbys first**, then transfers Raft leadership off the current
primary (a controlled failover) before restarting it last, so a writable leader is always
present. The status conditions tell the whole story:

```bash
$ kubectl get dcops -n demo documentdb-cls-restart \
    -o jsonpath='{range .status.conditions[*]}{.type}={.status} :: {.message}{"\n"}{end}'
Restart=True :: DocumentDB ops request is restarting pods
ResumePGCoordinator=True :: successfully resumed documentdb-coordinator
SetRaftKeyOpsRequestProgressing=True :: Successfully Set Raft Key OpsRequestProgressing
EvictPod=True :: evict pod; ConditionStatus:True
GetPrimary=True :: get primary; ConditionStatus:True
TransferLeader=True :: transfer leader; ConditionStatus:True
TransferLeaderForFailover=True :: transfer leader for failover; ConditionStatus:True
CheckIsMaster=True :: check is master; ConditionStatus:True
FailoverDone=True :: failover is done successfully
RestartNodes=True :: Successfully restarted all nodes
Successful=True :: Successfully completed the modification process.
UnsetRaftKeyOpsRequestProgressing=True :: Successfully Unset Raft Key OpsRequestProgressing
```

## After the restart

All three pods are freshly recreated and back to `2/2 Running`. Because leadership was
transferred during the rolling restart, the `primary` role has moved to a different pod — this
is expected and harmless:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=documentdb-cls-sample -L kubedb.com/role
NAME                      READY   STATUS    RESTARTS   AGE     ROLE
documentdb-cls-sample-0   2/2     Running   0          66s     standby
documentdb-cls-sample-1   2/2     Running   0          2m51s   primary
documentdb-cls-sample-2   2/2     Running   0          2m1s    standby
```

The database answers the MongoDB wire protocol immediately after the restart:

```bash
$ PASS=$(kubectl get secret -n demo documentdb-cls-sample-auth -o jsonpath='{.data.password}' | base64 -d)
$ kubectl exec -n demo documentdb-cls-sample-0 -c documentdb -- \
    mongosh "mongodb://default_user:${PASS}@localhost:10260/?tls=true&tlsAllowInvalidCertificates=true" \
    --quiet --eval 'db.runCommand({ ping: 1 })'
{ ok: 1 }
```

## Standalone

The method of restarting a standalone (`replicas: 1`) and a cluster database is exactly the
same — point `spec.databaseRef.name` at the standalone instance (`documentdb-sa-sample`).

> [!NOTE]
> On the build used to capture this guide (`pg17-0.109.0`), standalone instances did not finish
> bootstrapping: the standalone PetSet is rendered without the `documentdb-coordinator` sidecar,
> so the internal PostgreSQL is never initialized and the database never reaches `Ready`.
> Because KubeDB admits OpsRequests only against a `Ready` database, the standalone variant
> could not be exercised live (a `Restart` request stayed `Pending`); the cluster procedure
> above applies verbatim once a standalone instance is healthy.

## Cleaning Up

```bash
kubectl delete documentdbopsrequest -n demo documentdb-cls-restart
kubectl delete documentdb -n demo documentdb-cls-sample
kubectl delete ns demo
```

## Next Steps

- [Vertical scaling](/docs/guides/documentdb/scaling/vertical-scaling/) of a DocumentDB cluster.
- [Horizontal scaling](/docs/guides/documentdb/scaling/horizontal-scaling/) of a DocumentDB cluster.
