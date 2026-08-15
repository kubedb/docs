---
title: Move Etcd Raft Leadership
menu:
  docs_{{ .version }}:
    identifier: etcd-maintenance-move-leader
    name: Move Leader
    parent: etcd-maintenance
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Move the Raft Leadership of an Etcd Cluster

KubeDB supports transferring the Raft leadership of an etcd cluster to another member through an
`EtcdOpsRequest` of type `MoveLeader`. This is useful when you are about to disturb the node that
currently hosts the leader — draining it, restarting its pod, replacing the machine — and would
rather hand leadership over deliberately than let the cluster hold an election after the fact.

For the background on what a leadership transfer is and when to reach for it, see the
[Maintenance Operations Overview](/docs/guides/etcd/maintenance/overview.md#raft-leadership-transfer).

## Before You Begin

- At first, you need a Kubernetes cluster, and the `kubectl` command-line tool must be configured
  to communicate with your cluster. If you do not already have a cluster, you can create one
  using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install the `KubeDB` Provisioner and Ops-manager operators in your cluster following the steps
  [here](/docs/setup/README.md). etcd support is behind an alpha feature gate, so make sure the
  operators are installed with `Etcd=true` enabled (Helm value `featureGates.Etcd=true`).

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)
  - [Maintenance Operations Overview](/docs/guides/etcd/maintenance/overview.md)

- To keep things isolated, this tutorial uses a separate namespace called `demo`.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in
> [docs/examples/etcd](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd)
> folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Etcd

A leadership transfer only makes sense with more than one voting member, so we deploy a three
member cluster.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-cluster
  namespace: demo
spec:
  version: "3.6.4"
  replicas: 3
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

Let's create the `Etcd` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/maintenance/etcd.yaml
etcd.kubedb.com/etcd-cluster created
```

Wait until the cluster reports `Ready`,

```bash
$ kubectl get etcd -n demo
NAME           VERSION   STATUS   AGE
etcd-cluster   3.6.4     Ready    2m
```

The members are the ordinal pods of the PetSet, and each member's etcd `--name` is its pod name.
Those pod names are what you refer to in `spec.moveLeader.newLeader`.

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=etcd-cluster
NAME             READY   STATUS    RESTARTS   AGE
etcd-cluster-0   1/1     Running   0          2m
etcd-cluster-1   1/1     Running   0          2m
etcd-cluster-2   1/1     Running   0          2m
```

## Apply the MoveLeader OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-move-leader
  namespace: demo
spec:
  type: MoveLeader
  databaseRef:
    name: etcd-cluster
  moveLeader:
    newLeader: etcd-cluster-1
  timeout: 5m
```

- `spec.type` specifies the type of the ops request, here `MoveLeader`.
- `spec.databaseRef` holds the name of the `Etcd` object. It must live in the same namespace as
  the ops request.
- `spec.moveLeader.newLeader` names the member that should become the new leader. It accepts the
  member's pod name (which is also its etcd member name).
- The meaning of `spec.timeout` and `spec.apply` is described
  [here](/docs/guides/etcd/concepts/etcdopsrequest.md#spectimeout).

Let's create the `EtcdOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/maintenance/move-leader.yaml
etcdopsrequest.ops.kubedb.com/etcd-move-leader created
```

### Letting the operator pick the new leader

`spec.moveLeader.newLeader` is optional. If you leave it out, the operator picks the transferee
itself: it takes the first healthy **voting** member that is not the current leader. Learners are
never eligible — a learner does not vote and cannot become leader — so a member that is still
catching up after a scale-up is skipped.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-move-leader-auto
  namespace: demo
spec:
  type: MoveLeader
  databaseRef:
    name: etcd-cluster
  moveLeader: {}
  timeout: 5m
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/maintenance/move-leader-auto.yaml
etcdopsrequest.ops.kubedb.com/etcd-move-leader-auto created
```

Use the explicit form when you care *where* leadership lands (for example, moving it onto a
member in a different failure domain than the node you are about to drain), and the automatic
form when you only care that it moves *off* the current leader.

## What the operator does

`MoveLeader` is a pure control-plane operation on the Raft group. It does **not** pause the
`Etcd` object, does not touch the PetSet, and does not restart or evict any pod — a leadership
transfer is near instantaneous and does not interrupt the client API.

The steps are:

1. Resolve the current leader. Every member reports the ID of the member it believes is the
   leader, and that ID is mapped back onto a pod through the member list.
2. Resolve the transferee from `spec.moveLeader.newLeader`, or pick a healthy voting member if it
   is empty. If the request names a member that does not exist (or names a learner), the request
   fails immediately with an error — that is a mistake in the spec, not a transient condition
   worth retrying.
3. If the requested member is already the leader, the operator treats the request as satisfied
   and does nothing.
4. Issue the transfer. etcd only serves `MoveLeader` on the leader itself, so the operator opens
   a client pinned to the current leader's own endpoint to make the call.
5. Verify. The operator re-reads the leader until the cluster agrees that the requested member is
   now the leader, and only then marks the step done.

## Progress and status

```bash
$ kubectl get etcdopsrequest -n demo
NAME               TYPE         STATUS       AGE
etcd-move-leader   MoveLeader   Successful   35s
```

You can follow the individual steps through the ops request's conditions and events:

```bash
$ kubectl describe etcdopsrequest -n demo etcd-move-leader
```

The full object looks like this once it has finished:

```yaml
$ kubectl get etcdopsrequest -n demo etcd-move-leader -o yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  creationTimestamp: "2026-02-11T11:04:18Z"
  generation: 1
  name: etcd-move-leader
  namespace: demo
  resourceVersion: "61240"
  uid: 3f7a1c02-9d55-4a1e-8f6b-c0a1d2e3f455
spec:
  apply: IfReady
  databaseRef:
    name: etcd-cluster
  maxRetries: 1
  moveLeader:
    newLeader: etcd-cluster-1
  timeout: 5m
  type: MoveLeader
status:
  conditions:
  - lastTransitionTime: "2026-02-11T11:04:18Z"
    message: Moving the etcd raft leadership
    observedGeneration: 1
    reason: Running
    status: "True"
    type: Running
  - lastTransitionTime: "2026-02-11T11:04:24Z"
    message: Successfully moved the etcd raft leadership
    observedGeneration: 1
    reason: EtcdLeaderMoved
    status: "True"
    type: EtcdLeaderMoved
  - lastTransitionTime: "2026-02-11T11:04:25Z"
    message: Successfully moved the etcd raft leadership
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful
```

The condition to look for is **`EtcdLeaderMoved`**. It flips to `True` once the transfer has been
issued *and* the cluster has confirmed that the requested member holds the leadership; the ops
request is then marked `Successful`.

If the transfer has to be retried — for example because the cluster was momentarily without a
leader — you will also see interim conditions from the individual sub-steps, such as `GetLeader`
and `VerifyLeader`, with messages of the shape `get leader; ConditionStatus:False`. Those are
progress markers, not failures.

## Verify the new leader

The `EtcdLeaderMoved` condition is the operator's own confirmation, but you can check
independently from inside a member pod using `etcdctl`:

```bash
$ kubectl exec -it -n demo etcd-cluster-0 -c etcd -- etcdctl \
    --endpoints=http://etcd-cluster-0.etcd-cluster-pods.demo.svc:2379,http://etcd-cluster-1.etcd-cluster-pods.demo.svc:2379,http://etcd-cluster-2.etcd-cluster-pods.demo.svc:2379 \
    endpoint status -w table
```

The `IS LEADER` column of that table will show `true` for `etcd-cluster-1`.

> If the cluster has TLS or authentication enabled, add the corresponding `etcdctl` flags
> (`--cacert`/`--cert`/`--key`, and `--user`) and use the `https` scheme. See
> [Etcd CRD](/docs/guides/etcd/concepts/etcd.md) for where KubeDB mounts those secrets.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete etcdopsrequest -n demo etcd-move-leader
kubectl delete etcd -n demo etcd-cluster
kubectl delete ns demo
```

## Next Steps

- [Compact the keyspace history of an Etcd cluster](/docs/guides/etcd/maintenance/compact.md).
- [Defragment the member backends of an Etcd cluster](/docs/guides/etcd/maintenance/defragment.md).
- Detail concepts of the [Etcd object](/docs/guides/etcd/concepts/etcd.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
