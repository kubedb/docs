---
title: Horizontal Scaling DocumentDB
menu:
  docs_{{ .version }}:
    identifier: guides-documentdb-scaling-horizontal-details
    name: Horizontal Scaling
    parent: guides-documentdb-scaling-horizontal
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scaling of a DocumentDB Cluster

Horizontal scaling changes the **number of replicas** in a `DocumentDB` cluster. KubeDB drives
this through a `DocumentDBOpsRequest` of type `HorizontalScaling`. Because a DocumentDB cluster
forms a Raft group (managed by the `documentdb-coordinator` container in each pod), scaling is
not just a matter of adding or removing Kubernetes pods — the operator also **adds or removes
Raft members** so the consensus group always reflects the live set of replicas.

> Horizontal scaling applies to the clustered topology only; a standalone (`replicas: 1`)
> DocumentDB has no replicas to scale.

## Before You Begin

- You need a Kubernetes cluster and the `kubectl` CLI configured to talk to it.
- Install KubeDB following the steps [here](/docs/setup/README.md).
- This tutorial uses a namespace called `demo` (`kubectl create ns demo`).
- Deploy a 3-replica `DocumentDB` cluster (`documentdb-cls-sample`) and wait for it to become
  `Ready` before proceeding.

> Note: YAML files used in this tutorial are stored in [docs/examples/documentdb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/documentdb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Starting point: 3 replicas

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=documentdb-cls-sample -L kubedb.com/role
```
NAME                      READY   STATUS    RESTARTS   AGE     ROLE
documentdb-cls-sample-0   2/2     Running   0          97s     standby
documentdb-cls-sample-1   2/2     Running   0          3m22s   primary
documentdb-cls-sample-2   2/2     Running   0          2m32s   standby

```bash
kubectl get docdb -n demo documentdb-cls-sample -o jsonpath='{.spec.replicas}'
```
3

## Scale up: 3 → 5

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DocumentDBOpsRequest
metadata:
  name: documentdb-cls-hscale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: documentdb-cls-sample
  horizontalScaling:
    replicas: 5
```

```bash
kubectl apply -f cluster-hscale-up.yaml
```
documentdbopsrequest.ops.kubedb.com/documentdb-cls-hscale-up created

```bash
kubectl get dcops -n demo documentdb-cls-hscale-up
```
NAME                       TYPE                STATUS       AGE
documentdb-cls-hscale-up   HorizontalScaling   Successful   3m31s

Two new pods are provisioned (`-3`, `-4`) and **joined to the Raft group** as standbys. The
status conditions show the new members being added via the coordinator:

```bash
kubectl get dcops -n demo documentdb-cls-hscale-up \
    -o jsonpath='{range .status.conditions[*]}{.type}={.status} :: {.message}{"\n"}{end}'
```
Running=True :: DocumentDB ops request is horizontally scaling database
GetCurrentLeader--documentdb-cls-sample-0=True :: get current leader; ConditionStatus:True
AddRaftNode--documentdb-cls-sample-3=True :: add raft node; ConditionStatus:True; PodName:documentdb-cls-sample-3
PatchPetset=True :: patch petset; ConditionStatus:True
AddRaftNode--documentdb-cls-sample-4=True :: add raft node; ConditionStatus:True; PodName:documentdb-cls-sample-4
HorizontalScaleUp=True :: Successfully Horizontally Scaled Up
Successful=True :: Successfully Horizontally Scaled DocumentDB

The cluster now runs five pods — one primary and four standbys:

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=documentdb-cls-sample -L kubedb.com/role
```
NAME                      READY   STATUS    RESTARTS   AGE     ROLE
documentdb-cls-sample-0   2/2     Running   0          5m7s    standby
documentdb-cls-sample-1   2/2     Running   0          6m52s   primary
documentdb-cls-sample-2   2/2     Running   0          6m2s    standby
documentdb-cls-sample-3   2/2     Running   0          2m56s   standby
documentdb-cls-sample-4   2/2     Running   0          106s    standby

## Scale down: 5 → 3

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DocumentDBOpsRequest
metadata:
  name: documentdb-cls-hscale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: documentdb-cls-sample
  horizontalScaling:
    replicas: 3
```

```bash
kubectl apply -f cluster-hscale-down.yaml
```
documentdbopsrequest.ops.kubedb.com/documentdb-cls-hscale-down created

```bash
kubectl get dcops -n demo documentdb-cls-hscale-down
```
NAME                         TYPE                STATUS       AGE
documentdb-cls-hscale-down   HorizontalScaling   Successful   2m43s

On the way down the operator first **removes the surplus Raft members**, then deletes their
pods and their PVCs, so no orphaned storage is left behind:

```bash
kubectl get dcops -n demo documentdb-cls-hscale-down \
    -o jsonpath='{range .status.conditions[*]}{.type}={.status} :: {.message}{"\n"}{end}'
```
Running=True :: DocumentDB ops request is horizontally scaling database
GetCurrentRaftLeader--documentdb-cls-sample-0=True :: get current raft leader; ConditionStatus:True
RemoveRaftNode--documentdb-cls-sample-4=True :: remove raft node; ConditionStatus:True; PodName:documentdb-cls-sample-4
PatchPetset=True :: patch petset; ConditionStatus:True
DeletePvc--documentdb-cls-sample-4=True :: delete pvc; ConditionStatus:True; PodName:documentdb-cls-sample-4
RemoveRaftNode--documentdb-cls-sample-3=True :: remove raft node; ConditionStatus:True; PodName:documentdb-cls-sample-3
DeletePvc--documentdb-cls-sample-3=True :: delete pvc; ConditionStatus:True; PodName:documentdb-cls-sample-3
HorizontalScaleDown=True :: Successfully Horizontally Scaled Down
Successful=True :: Successfully Horizontally Scaled DocumentDB

Back to the original three-pod topology, still fully serviceable:

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=documentdb-cls-sample -L kubedb.com/role
```
NAME                      READY   STATUS    RESTARTS   AGE     ROLE
documentdb-cls-sample-0   2/2     Running   0          8m1s    standby
documentdb-cls-sample-1   2/2     Running   0          9m46s   primary
documentdb-cls-sample-2   2/2     Running   0          8m56s   standby

```bash
PASS=$(kubectl get secret -n demo documentdb-cls-sample-auth -o jsonpath='{.data.password}' | base64 -d)
```

```bash
kubectl exec -n demo documentdb-cls-sample-1 -c documentdb -- \
    mongosh "mongodb://default_user:${PASS}@localhost:10260/?tls=true&tlsAllowInvalidCertificates=true" \
    --quiet --eval 'db.runCommand({ ping: 1 })'
```
{ ok: 1 }

## Key takeaway

Raft membership grows and shrinks **in lockstep** with the replica count. On scale-up the
coordinator runs `AddRaftNode` for each new pod before it counts toward the quorum; on
scale-down it runs `RemoveRaftNode` (and cleans up the PVC) before the pod disappears. The
leader is never disrupted, so writes through the MongoDB endpoint continue uninterrupted in
both directions.

## Cleaning Up

```bash
kubectl delete documentdbopsrequest -n demo documentdb-cls-hscale-up documentdb-cls-hscale-down
kubectl delete documentdb -n demo documentdb-cls-sample
kubectl delete ns demo
```

## Next Steps

- [Vertical scaling](/docs/guides/documentdb/scaling/vertical-scaling/) of a DocumentDB cluster.
- [Compute autoscaling](/docs/guides/documentdb/autoscaler/compute/) of a DocumentDB cluster.
