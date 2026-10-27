---
title: Updating Etcd Cluster
menu:
  docs_{{ .version }}:
    identifier: etcd-update-version-ops
    name: Update Version
    parent: etcd-update-version
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Update Version of an Etcd Cluster

This guide will show you how to use the `KubeDB` Ops-manager operator to update the version of an
`Etcd` cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be
  configured to communicate with your cluster. If you do not already have a cluster, you can create
  one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps
  [here](/docs/setup/README.md). Etcd support is an **alpha** feature, so the operators must be
  installed with the `Etcd` feature gate turned on (`--set featureGates.Etcd=true`).

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdVersion](/docs/guides/etcd/concepts/etcdversion.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)
  - [Update Version Overview](/docs/guides/etcd/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this
tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in
> [docs/examples/etcd/update-version](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd/update-version)
> directory of the [kubedb/docs](https://github.com/kubedb/docs) repository.

## Check the Available Versions

The versions you may update to are the `EtcdVersion` objects in the catalog:

```bash
$ kubectl get etcdversions
NAME     VERSION   DB_IMAGE                                DEPRECATED   AGE
3.5.21   3.5.21    ghcr.io/appscode-images/etcd:v3.5.21                 3d
3.6.4    3.6.4     ghcr.io/appscode-images/etcd:v3.6.4                  3d
```

Before you pick a target, check it against the upgrade-path rules — **same major version, no
downgrade, at most one minor step, target not deprecated**. See
[The Upgrade Path Is Validated](/docs/guides/etcd/update-version/overview.md#the-upgrade-path-is-validated--read-this-first).
`3.5.21 → 3.6.4` below is a single minor step, so it is allowed.

## Prepare the Etcd Cluster

Below is the YAML of the `Etcd` CR that we are going to create, on version `3.5.21`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-cluster
  namespace: demo
spec:
  version: "3.5.21"
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/update-version/etcd.yaml
etcd.kubedb.com/etcd-cluster created
```

Now, wait until `etcd-cluster` has status `Ready`. i.e,

```bash
$ kubectl get etcd -n demo
NAME           VERSION   STATUS   AGE
etcd-cluster   3.5.21    Ready    4m
```

We are now ready to apply an `EtcdOpsRequest` to update this cluster.

## Update the Etcd Version

### Create EtcdOpsRequest

Below is the YAML of the `EtcdOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: etcd-cluster
  updateVersion:
    targetVersion: "3.6.4"
  timeout: 20m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are updating the `etcd-cluster` database.
- `spec.type` specifies that we are performing `UpdateVersion`.
- `spec.updateVersion.targetVersion` is the **name of an `EtcdVersion` object** in the catalog, not an
  arbitrary image tag.
- `spec.timeout` bounds each step. Every member has to restart and rejoin the quorum in sequence, so
  give it room on a larger cluster.

Let's create the `EtcdOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/update-version/etcdops-update-version.yaml
etcdopsrequest.ops.kubedb.com/etcd-update-version created
```

### Verify the Etcd Version Updated Successfully

Let's wait for the `EtcdOpsRequest` to become `Successful`:

```bash
$ watch kubectl get etcdopsrequest -n demo
Every 2.0s: kubectl get etcdopsrequest -n demo
NAME                  TYPE            STATUS       AGE
etcd-update-version   UpdateVersion   Successful   3m41s
```

If we describe the `EtcdOpsRequest`, we get an overview of the steps that were followed:

```bash
$ kubectl describe etcdopsrequest -n demo etcd-update-version
Name:         etcd-update-version
Namespace:    demo
API Version:  ops.kubedb.com/v1alpha1
Kind:         EtcdOpsRequest
Metadata:
  Creation Timestamp:  2026-08-15T13:02:55Z
  Generation:          1
  Resource Version:    171338
  UID:                 7d10f2b6-58a1-4b70-9f2a-0f1a7c3e2b44
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   etcd-cluster
  Timeout:  20m
  Type:     UpdateVersion
  Update Version:
    Target Version:  3.6.4
Status:
  Conditions:
    Last Transition Time:  2026-08-15T13:02:55Z
    Message:               Version Update is in progress
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2026-08-15T13:03:05Z
    Message:               Successfully updated the petset image
    Observed Generation:   1
    Reason:                UpdateEtcdPetSet
    Status:                True
    Type:                  UpdateEtcdPetSet
    Last Transition Time:  2026-08-15T13:03:05Z
    Message:               Successfully updated the Etcd version
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2026-08-15T13:03:45Z
    Message:               check pod ready; ConditionStatus:True; PodName:etcd-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--etcd-cluster-1
    Last Transition Time:  2026-08-15T13:03:50Z
    Message:               etcd cluster healthy; ConditionStatus:True; PodName:etcd-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  EtcdClusterHealthy--etcd-cluster-1
    Last Transition Time:  2026-08-15T13:04:30Z
    Message:               check pod ready; ConditionStatus:True; PodName:etcd-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--etcd-cluster-2
    Last Transition Time:  2026-08-15T13:04:35Z
    Message:               etcd cluster healthy; ConditionStatus:True; PodName:etcd-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  EtcdClusterHealthy--etcd-cluster-2
    Last Transition Time:  2026-08-15T13:05:15Z
    Message:               check pod ready; ConditionStatus:True; PodName:etcd-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--etcd-cluster-0
    Last Transition Time:  2026-08-15T13:06:30Z
    Message:               Successfully restarted all the etcd members on the new version
    Observed Generation:   1
    Reason:                RestartEtcdPods
    Status:                True
    Type:                  RestartEtcdPods
    Last Transition Time:  2026-08-15T13:06:36Z
    Message:               Successfully updated the etcd version
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason            Age    From                         Message
  ----    ------            ----   ----                         -------
  Normal  PauseDatabase     3m41s  KubeDB Ops-manager Operator  Pausing Etcd demo/etcd-cluster
  Normal  UpdateEtcdPetSet  3m31s  KubeDB Ops-manager Operator  Successfully updated the petset image
  Normal  UpdateDatabase    3m31s  KubeDB Ops-manager Operator  Successfully updated the Etcd version
  Normal  RestartEtcdPods   6s     KubeDB Ops-manager Operator  Successfully restarted all the etcd members on the new version
  Normal  ResumeDatabase    3s     KubeDB Ops-manager Operator  Successfully resumed Etcd demo/etcd-cluster
  Normal  Successful        3s     KubeDB Ops-manager Operator  Successfully updated the etcd version
```

The `CheckPodReady--<pod>` / `EtcdClusterHealthy--<pod>` pairs are the sequencing gate: the operator
does not evict the next member until the previous one is `Ready` *and* the cluster has a healthy
quorum again. The followers (`etcd-cluster-1`, `etcd-cluster-2` in this run) are rolled before the
Raft leader (`etcd-cluster-0` here); if the `MoveLeader` step before the leader's restart has to
wait, you will also see a `MoveLeader--<pod>` condition.

Now let's verify the version:

```bash
$ kubectl get etcd -n demo etcd-cluster -o jsonpath='{.spec.version}'
3.6.4

$ kubectl get etcd -n demo
NAME           VERSION   STATUS   AGE
etcd-cluster   3.6.4     Ready    21m

$ kubectl get pod -n demo etcd-cluster-0 -o jsonpath='{.spec.containers[?(@.name=="etcd")].image}'
ghcr.io/appscode-images/etcd:v3.6.4@sha256:...
```

> The image on the `PetSet` is resolved to an immutable digest, so every member is guaranteed to run
> the exact same build even if the tag is later re-pushed.

And from etcd itself:

```bash
$ kubectl exec -n demo etcd-cluster-0 -c etcd -- \
    etcdctl --endpoints=http://127.0.0.1:2379 endpoint status --cluster -w table
+---------------------------------------------------------------------+------------------+---------+---------+-----------+-----------+------------+
|                              ENDPOINT                               |        ID        | VERSION | DB SIZE | IS LEADER | RAFT TERM | RAFT INDEX |
+---------------------------------------------------------------------+------------------+---------+---------+-----------+-----------+------------+
| http://etcd-cluster-0.etcd-cluster-pods.demo.svc.cluster.local:2379 | 1a2b3c4d5e6f7a8b |   3.6.4 |   20 kB |     false |         4 |         62 |
| http://etcd-cluster-1.etcd-cluster-pods.demo.svc.cluster.local:2379 | 2b3c4d5e6f7a8b9c |   3.6.4 |   20 kB |      true |         4 |         62 |
| http://etcd-cluster-2.etcd-cluster-pods.demo.svc.cluster.local:2379 | 3c4d5e6f7a8b9c0d |   3.6.4 |   20 kB |     false |         4 |         62 |
+---------------------------------------------------------------------+------------------+---------+---------+-----------+-----------+------------+
```

Note that leadership has moved to `etcd-cluster-1` — that is the `MoveLeader` call the operator made
before restarting the old leader, and it is expected.

## What a Rejected Upgrade Path Looks Like

The upgrade path is validated by the Ops-manager operator **at execution time**, not by the admission
webhook. So a request with an unsupported target is accepted by `kubectl apply` and then fails:

```bash
$ kubectl get etcdopsrequest -n demo etcd-bad-update
NAME              TYPE            STATUS   AGE
etcd-bad-update   UpdateVersion   Failed   35s

$ kubectl describe etcdopsrequest -n demo etcd-bad-update
...
Events:
  Type     Reason   Age  From                         Message
  ----     ------   ---- ----                         -------
  Warning  Failed   30s  KubeDB Ops-manager Operator  etcd can not be updated from 3.5.21 to 3.7.0 in one step: only a single minor version at a time is supported (upgrade to 3.6.x first)
```

Other messages you may see from the same check:

- `etcd can not be updated from <a> to <b>: a major version change is not supported`
- `etcd can not be updated from <a> to <b>: downgrades are not supported`
- `EtcdOpsRequest demo/<name>: can not update to the deprecated version <target>`

The database is untouched in all of these cases — validation runs before the `PetSet` is patched.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete etcdopsrequest -n demo etcd-update-version
kubectl delete etcd -n demo etcd-cluster
kubectl delete ns demo
```

## Next Steps

- Scale the cluster [vertically](/docs/guides/etcd/scaling/vertical-scaling/vertical-scaling.md) or
  [horizontally](/docs/guides/etcd/scaling/horizontal-scaling/horizontal-scaling.md).
- [Expand the volume](/docs/guides/etcd/volume-expansion/volume-expansion.md) of an Etcd cluster.
- Detail concepts of the [EtcdVersion](/docs/guides/etcd/concepts/etcdversion.md) and
  [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md) objects.
