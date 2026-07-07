---
title: DocumentDB Automatic Failover
menu:
  docs_{{ .version }}:
    identifier: dc-failure-disaster-recovery-failover
    name: Automatic Failover
    parent: dc-failure-disaster-recovery
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Automatic Failover in a DocumentDB Cluster

A `DocumentDB` cluster is self-healing. Each pod runs a `documentdb-coordinator` container that
participates in a **Raft** consensus group; the Raft leader's pod is labelled
`kubedb.com/role=primary` and runs the writable PostgreSQL engine, while the others are
`standby` replicas streaming from it. If the primary pod dies, the surviving coordinators detect
the loss, elect a new leader, promote the healthiest standby to primary, and re-label the pods —
with **no operator action and no OpsRequest required**. This guide forces a failover and shows
that committed data survives it.

> This is a cluster-only scenario; a standalone (`replicas: 1`) DocumentDB has no standby to fail
> over to.

## Before You Begin

- You need a Kubernetes cluster and the `kubectl` CLI configured to talk to it.
- Install KubeDB following the steps [here](/docs/setup/README.md).
- This tutorial uses a namespace called `demo` (`kubectl create ns demo`).
- Deploy a 3-replica `DocumentDB` cluster (`documentdb-cls-sample`) and wait for it to become
  `Ready`.

> Note: YAML files used in this tutorial are stored in [docs/examples/documentdb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/documentdb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Identify the current leader

The leader is the pod labelled `kubedb.com/role=primary`:

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=documentdb-cls-sample -L kubedb.com/role
```
NAME                      READY   STATUS    RESTARTS   AGE     ROLE
documentdb-cls-sample-0   2/2     Running   0          4m4s    primary
documentdb-cls-sample-1   2/2     Running   0          99s     standby
documentdb-cls-sample-2   2/2     Running   0          2m48s   standby

`documentdb-cls-sample-0` is the leader.

## Write a test document on the primary

```bash
PASS=$(kubectl get secret -n demo documentdb-cls-sample-auth -o jsonpath='{.data.password}' | base64 -d)
```

```bash
kubectl exec -n demo documentdb-cls-sample-0 -c documentdb -- \
    mongosh "mongodb://default_user:${PASS}@localhost:10260/?tls=true&tlsAllowInvalidCertificates=true" \
    --quiet --eval '
```
      db.getSiblingDB("failover").coll.insertOne({k:"before-failover", ts:new Date()});
      printjson(db.getSiblingDB("failover").coll.findOne({k:"before-failover"}));'
{
  _id: ObjectId('6a43e4bd2bb67b71d58563b1'),
  k: 'before-failover',
  ts: ISODate('2026-06-30T15:46:05.334Z')
}

## Force a failover

Simulate a node loss by force-deleting the leader pod:

```bash
kubectl delete pod -n demo documentdb-cls-sample-0 --grace-period=0 --force
```
pod "documentdb-cls-sample-0" force deleted from demo namespace

## Watch the re-election

Within a few seconds a new primary is elected. The database briefly reports `Critical` (it has
lost a quorum member) and returns to `Ready` once a new leader is serving:

```bash
# poll: kubectl get pods ... -L kubedb.com/role  +  kubectl get docdb
```
[t+0s ] db=Critical primary=''                       sample-0=0/2 (terminating)  sample-1=standby  sample-2=standby
[t+15s] db=Critical primary='documentdb-cls-sample-2' sample-0=2/2 (rejoining)    sample-1=standby  sample-2=primary
[t+45s] db=Ready    primary='documentdb-cls-sample-2' sample-0=standby            sample-1=standby  sample-2=primary

Final topology — `documentdb-cls-sample-2` is the new primary and the old leader has rejoined as
a standby:

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=documentdb-cls-sample -L kubedb.com/role
```
NAME                      READY   STATUS    RESTARTS   AGE     ROLE
documentdb-cls-sample-0   2/2     Running   0          56s     standby
documentdb-cls-sample-1   2/2     Running   0          2m37s   standby
documentdb-cls-sample-2   2/2     Running   0          3m46s   primary

The coordinator log on the new primary tells the whole story: the Raft leader change is
detected, the **healthiest** node (lowest LSN diff) is chosen, the PostgreSQL engine is
promoted, and the pod is re-labelled `primary`:

```bash
kubectl logs -n demo documentdb-cls-sample-2 -c documentdb-coordinator | grep -iE 'leader|elect|primary|promot'
```
on_Leader_change.go:71]  *** Raft Leader Changed **** Checking if I can run as primary*** My current Role is standby
on_Leader_change.go:350] Healthiest node detected documentdb-cls-sample-2 with LSN diff 0 bytes
ha_postgres.go:286]      Previous primary from this node is : documentdb-cls-sample-0
ha_postgres.go:288]      new elected primary is :documentdb-cls-sample-2.
ha_postgres.go:381]      I am the healthiest one and I am the primary.
ha_postgres.go:760]      This pod is now a  primary
exec_utils.go:159]       demo/documentdb-cls-sample-2 is promoted as primary
ha_postgres.go:800]      Successfully patched pod demo/documentdb-cls-sample-2 to role "primary" on attempt 1
health.go:209]           Timeline missmatch identified. proposing new leader timeline = 3

## Verify data continuity

Reconnect to the **new** primary and read the document written before the failover — it is
intact:

```bash
kubectl exec -n demo documentdb-cls-sample-2 -c documentdb -- \
    mongosh "mongodb://default_user:${PASS}@localhost:10260/?tls=true&tlsAllowInvalidCertificates=true" \
    --quiet --eval 'printjson(db.getSiblingDB("failover").coll.findOne({k:"before-failover"}));'
```
{
  _id: ObjectId('6a43e4bd2bb67b71d58563b1'),
  k: 'before-failover',
  ts: ISODate('2026-06-30T15:46:05.334Z')
}

The cluster is back to `Ready` with all conditions healthy:

```bash
kubectl get docdb -n demo documentdb-cls-sample
```
NAME                    NAMESPACE   VERSION        STATUS   AGE
documentdb-cls-sample   demo        pg17-0.109.0   Ready    13m

```bash
kubectl get docdb -n demo documentdb-cls-sample \
    -o jsonpath='{range .status.conditions[*]}{.type}={.status} :: {.message}{"\n"}{end}'
```
ProvisioningStarted=True :: The KubeDB operator has started the provisioning of DocumentDB: demo/documentdb-cls-sample
ReplicaReady=True :: All replicas are ready for DocumentDB demo/documentdb-cls-sample
AcceptingConnection=True :: The DocumentDB: demo/documentdb-cls-sample is accepting client requests.
Ready=True :: The DocumentDB: demo/documentdb-cls-sample is ready.
Provisioned=True :: The DocumentDB: demo/documentdb-cls-sample is successfully provisioned.

## Summary

Failover is automatic and fast: the Raft group elected a new leader and promoted the healthiest
standby within seconds, the operator never had to intervene, and the previously-committed write
survived the loss of the original primary. The old pod rejoined the cluster as a standby once it
was rescheduled.

## Cleaning Up

```bash
kubectl delete documentdb -n demo documentdb-cls-sample
kubectl delete ns demo
```

## Next Steps

- [Horizontal scaling](/docs/guides/documentdb/scaling/horizontal-scaling/) of a DocumentDB cluster.
- [Restart](/docs/guides/documentdb/restart/) a DocumentDB database.
