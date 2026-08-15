---
title: Vertical Scaling Etcd
menu:
  docs_{{ .version }}:
    identifier: etcd-vertical-scaling-ops
    name: Scale Vertically
    parent: etcd-vertical-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale Etcd

This guide will show you how to use the `KubeDB` Ops-manager operator to update the compute
resources of an `Etcd` cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be
  configured to communicate with your cluster. If you do not already have a cluster, you can create
  one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps
  [here](/docs/setup/README.md). Etcd support is an **alpha** feature, so the operators must be
  installed with the `Etcd` feature gate turned on (`--set featureGates.Etcd=true`).

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/etcd/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this
tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in
> [docs/examples/etcd/scaling/vertical-scaling](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd/scaling/vertical-scaling)
> directory of the [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling

Here, we are going to deploy an `Etcd` cluster using a version supported by the `KubeDB` operator,
and then apply vertical scaling on it.

### Deploy Etcd

Below is the YAML of the `Etcd` CR that we are going to create,

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/scaling/vertical-scaling/etcd.yaml
etcd.kubedb.com/etcd-cluster created
```

Now, wait until `etcd-cluster` has status `Ready`. i.e,

```bash
$ kubectl get etcd -n demo
NAME           VERSION   STATUS   AGE
etcd-cluster   3.6.4     Ready    4m
```

Let's check the resources of the `etcd` container:

```bash
$ kubectl get pod -n demo etcd-cluster-0 -o json | jq '.spec.containers[] | select(.name=="etcd") | .resources'
{
  "limits": {
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

These are the default resources assigned by the KubeDB operator (etcd is latency sensitive but its
working set is bounded by the backend quota rather than by the dataset size, so the default footprint
is deliberately modest).

We are now ready to apply an `EtcdOpsRequest` to update these resources.

### Vertical Scaling

#### Create EtcdOpsRequest

Below is the YAML of the `EtcdOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: etcd-cluster
  verticalScaling:
    mode: Restart
    etcd:
      resources:
        requests:
          cpu: "1"
          memory: "2Gi"
        limits:
          cpu: "1"
          memory: "2Gi"
  timeout: 10m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing the vertical scaling operation on the
  `etcd-cluster` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.verticalScaling.etcd.resources` specifies the desired resources of the `etcd` container after
  scaling.
- `spec.verticalScaling.mode` specifies how the scaling is actuated — `Restart` (the default; the
  member pods are restarted one at a time, leader last) or `InPlace` (the running pods are resized
  through the `pods/resize` subresource with no restart at all). See
  [Vertical Scaling Modes](/docs/guides/etcd/scaling/vertical-scaling/overview.md#vertical-scaling-modes).
- Have a look [here](/docs/guides/etcd/concepts/etcdopsrequest.md) to understand the `timeout` and
  `apply` fields.

> `spec.verticalScaling.etcd` and `spec.verticalScaling.exporter` are the only two containers the API
> accepts, and at least one of them must be set — the validating webhook rejects an empty
> `verticalScaling`. In practice you always want `etcd`: KubeDB does not run an exporter sidecar for
> etcd (see the [note in the overview](/docs/guides/etcd/scaling/vertical-scaling/overview.md#scalable-containers)).

Let's create the `EtcdOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/scaling/vertical-scaling/etcdops-vscale.yaml
etcdopsrequest.ops.kubedb.com/etcd-vscale created
```

#### Verify the Etcd resources are updated successfully

Let's wait for the `EtcdOpsRequest` to become `Successful`. Run the following command to watch the
`EtcdOpsRequest` CR,

```bash
$ watch kubectl get etcdopsrequest -n demo
Every 2.0s: kubectl get etcdopsrequest -n demo
NAME          TYPE              STATUS       AGE
etcd-vscale   VerticalScaling   Successful   3m12s
```

If we describe the `EtcdOpsRequest`, we get an overview of the steps that were followed:

```bash
$ kubectl describe etcdopsrequest -n demo etcd-vscale
Name:         etcd-vscale
Namespace:    demo
API Version:  ops.kubedb.com/v1alpha1
Kind:         EtcdOpsRequest
Metadata:
  Creation Timestamp:  2026-08-15T09:41:02Z
  Generation:          1
  Resource Version:    120451
  UID:                 3ca1e0c4-7a0f-4f1d-9b3a-6f2a4c1c9f01
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   etcd-cluster
  Timeout:  10m
  Type:     VerticalScaling
  Vertical Scaling:
    Etcd:
      Resources:
        Limits:
          Cpu:     1
          Memory:  2Gi
        Requests:
          Cpu:     1
          Memory:  2Gi
    Mode:          Restart
Status:
  Conditions:
    Last Transition Time:  2026-08-15T09:41:02Z
    Message:               Vertical Scaling is in progress
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2026-08-15T09:41:12Z
    Message:               Successfully updated the petset resources
    Observed Generation:   1
    Reason:                UpdateEtcdPetSet
    Status:                True
    Type:                  UpdateEtcdPetSet
    Last Transition Time:  2026-08-15T09:41:12Z
    Message:               Successfully updated the Etcd resources
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2026-08-15T09:41:47Z
    Message:               check pod ready; ConditionStatus:True; PodName:etcd-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--etcd-cluster-1
    Last Transition Time:  2026-08-15T09:41:52Z
    Message:               etcd cluster healthy; ConditionStatus:True; PodName:etcd-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  EtcdClusterHealthy--etcd-cluster-1
    Last Transition Time:  2026-08-15T09:42:27Z
    Message:               check pod ready; ConditionStatus:True; PodName:etcd-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--etcd-cluster-2
    Last Transition Time:  2026-08-15T09:42:32Z
    Message:               etcd cluster healthy; ConditionStatus:True; PodName:etcd-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  EtcdClusterHealthy--etcd-cluster-2
    Last Transition Time:  2026-08-15T09:43:07Z
    Message:               check pod ready; ConditionStatus:True; PodName:etcd-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--etcd-cluster-0
    Last Transition Time:  2026-08-15T09:43:12Z
    Message:               Successfully restarted the etcd members with the new resources
    Observed Generation:   1
    Reason:                VerticalScale
    Status:                True
    Type:                  VerticalScale
    Last Transition Time:  2026-08-15T09:43:14Z
    Message:               Successfully vertically scaled the etcd cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason            Age    From                         Message
  ----    ------            ----   ----                         -------
  Normal  PauseDatabase     3m12s  KubeDB Ops-manager Operator  Pausing Etcd demo/etcd-cluster
  Normal  UpdateEtcdPetSet  3m2s   KubeDB Ops-manager Operator  Successfully updated the petset resources
  Normal  UpdateDatabase    3m2s   KubeDB Ops-manager Operator  Successfully updated the Etcd resources
  Normal  VerticalScale     12s    KubeDB Ops-manager Operator  Successfully restarted the etcd members with the new resources
  Normal  ResumeDatabase    10s    KubeDB Ops-manager Operator  Successfully resumed Etcd demo/etcd-cluster
  Normal  Successful        10s    KubeDB Ops-manager Operator  Successfully vertically scaled the etcd cluster
```

A few things worth pointing out in that condition list:

- The per-pod `CheckPodReady--<pod>` / `EtcdClusterHealthy--<pod>` conditions are written for every
  member the operation touched, not only for the steps that had to wait — they are both the
  sequencing gate and the operator's own record of what is already done, so they survive a restart of
  the operator. The gate itself is: the operator does not evict the next member until the previous
  one is `Ready` **and** the cluster reports a healthy quorum (`healthy members >= N/2+1`) again. The
  matching `GetLeader--<pod>` and `EvictPod--<pod>` conditions are elided from the output above.
- The followers (`etcd-cluster-1`, `etcd-cluster-2` in this run) are restarted before the Raft leader
  (`etcd-cluster-0` here). Which pod is the leader is discovered at runtime, so the order you observe
  depends on where leadership happens to sit when the ops request starts.
- Before the leader is evicted, the operator issues a `MoveLeader` RPC to hand leadership to another
  voting member, and always records a `MoveLeader--<pod>` condition naming the pod it moved
  leadership away from.

Now let's verify from the `PetSet` and the pods that the resources have been updated:

```bash
$ kubectl get petset -n demo etcd-cluster \
  -o jsonpath='{.spec.template.spec.containers[?(@.name=="etcd")].resources}' | jq
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "1",
    "memory": "2Gi"
  }
}

$ kubectl get pod -n demo etcd-cluster-0 -o json | jq '.spec.containers[] | select(.name=="etcd") | .resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "1",
    "memory": "2Gi"
  }
}
```

The new resources are also persisted onto the `Etcd` object itself, so the Provisioner operator
re-renders the same template after the database is resumed:

```bash
$ kubectl get etcd -n demo etcd-cluster \
  -o jsonpath='{.spec.podTemplate.spec.containers[?(@.name=="etcd")].resources}' | jq
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "1",
    "memory": "2Gi"
  }
}
```

### In-Place Vertical Scaling

To resize the member pods **without restarting them**, set `spec.verticalScaling.mode` to `InPlace`.
The operator resizes the running containers through the Kubernetes `pods/resize` subresource; no pod
is evicted, no Raft leadership is moved, and there is no election risk at all.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-vscale-inplace
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: etcd-cluster
  verticalScaling:
    mode: InPlace
    etcd:
      resources:
        requests:
          cpu: "1"
          memory: "3Gi"
        limits:
          cpu: "1"
          memory: "3Gi"
  timeout: 10m
  apply: IfReady
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/scaling/vertical-scaling/etcdops-vscale-inplace.yaml
etcdopsrequest.ops.kubedb.com/etcd-vscale-inplace created
```

The condition list looks the same up to `UpdateDatabase`; the actuation step is then reported through
`RequestInPlaceResize--<pod>` / `CheckInPlaceResizeSettled--<pod>` conditions and finishes with:

```yaml
    Message:               Successfully resized the etcd members in place
    Reason:                VerticalScale
    Status:                True
    Type:                  VerticalScale
```

If the kubelet reports the resize as `Infeasible` for a pod (its Node cannot fit the new request
without a reschedule), the operator logs that and falls back to the `Restart` behavior **for that pod
only**, using the same leader-last ordering. So a partially infeasible `InPlace` request still
converges — some pods are resized live, the rest are rolled.

> **Note:** `InPlace` requires the Kubernetes `InPlacePodVerticalScaling` feature gate (on by default
> from Kubernetes v1.33). On older clusters use `Restart`.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete etcdopsrequest -n demo etcd-vscale etcd-vscale-inplace
kubectl delete etcd -n demo etcd-cluster
kubectl delete ns demo
```

## Next Steps

- Scale the cluster [horizontally](/docs/guides/etcd/scaling/horizontal-scaling/horizontal-scaling.md).
- [Expand the volume](/docs/guides/etcd/volume-expansion/volume-expansion.md) of an Etcd cluster.
- Detail concepts of the [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md) object.
