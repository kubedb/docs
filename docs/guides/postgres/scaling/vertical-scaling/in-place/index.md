---
title: In-Place Vertical Scaling Postgres
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-scaling-vertical-inplace
    name: In-Place Vertical Scaling
    parent: guides-postgres-scaling-vertical
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# In-Place Vertical Scaling Postgres

This guide will show you how to use `KubeDB-Ops-Manager` to update the CPU and
memory of a running `Postgres` instance **in place**, that is, without recreating
the Pods and without a primary failover.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line
  tool must be configured to communicate with your cluster. If you do not already
  have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- In-place resize requires a Kubernetes cluster with the **in-place pod resize**
  feature available (the `pods/resize` subresource; container-level in-place resize
  is GA in Kubernetes v1.35). On clusters that do not support it, KubeDB
  automatically falls back to the regular restart-based vertical scaling, so the
  request still completes (see [Eligibility and fallback](#eligibility-and-fallback)).

- Install `KubeDB-Provisioner` and `KubeDB-Ops-Manager` in your cluster following
  the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Postgres](/docs/guides/postgres/concepts/postgres.md)
  - [PostgresOpsRequest](/docs/guides/postgres/concepts/opsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/postgres/scaling/vertical-scaling/overview/index.md)

To keep everything isolated, we are going to use a separate namespace called `demo`
throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/postgres/scaling/vertical-scaling/in-place/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/scaling/vertical-scaling/in-place/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## In-Place vs Restart vertical scaling

The default vertical scaling path patches the `PetSet` and then **evicts and
recreates** every Pod (replicas first, then a primary failover, then the primary)
so each Pod starts with the new resources. It is safe but disruptive: every resize
is a full rolling restart plus a failover.

In-place vertical scaling instead asks the kubelet to change the running
container's cgroup limits through the `pods/resize` subresource, so:

- **CPU** changes (in either direction) take effect live, with no Pod restart and
  no failover.
- A **memory increase** grows the cgroup live. Because PostgreSQL sets
  `shared_buffers` only at startup, that GUC is left unchanged; the reloadable,
  memory-derived GUCs (`effective_cache_size`, `work_mem`, `maintenance_work_mem`)
  are re-applied live so the database can take advantage of the extra memory.

You opt in per request with `spec.verticalScaling.mode: InPlace`. The mode
defaults to `Restart`, so existing OpsRequests behave exactly as before.

## Deploy Postgres

Below is the YAML of a 3-replica `Postgres` cluster we are going to create. Using a
cluster (replicas > 1) lets us confirm that in-place scaling keeps the same primary
(no failover). We also enable [auto-tuning](/docs/guides/postgres/configuration/pgtune.md)
(`spec.configuration.tuning`) so that the memory-derived parameters are managed by
KubeDB — this is what lets the in-place memory increase re-apply the reloadable GUCs
live (see [In-Place memory increase](#in-place-memory-increase)).

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg
  namespace: demo
spec:
  version: "13.13"
  replicas: 3
  standbyMode: Hot
  configuration:
    tuning:
      profile: oltp
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

Let's create the `Postgres` cr we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/scaling/vertical-scaling/in-place/yamls/postgres.yaml
postgres.kubedb.com/pg created
```

Wait for the `Postgres` to become `Ready`,

```bash
$ kubectl get postgres -n demo pg
NAME   VERSION   STATUS   AGE
pg     13.13     Ready    4m16s
```

Let's check the `pg-0` Pod's postgres container resources (the postgres container
is the first container, so its index is `0`), and note the Pod's UID, its
`restartCount`, and the role of each Pod, so we can compare after the resize.

```bash
$ kubectl get pod -n demo pg-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}

$ kubectl get pod -n demo pg-0 -o jsonpath='{.metadata.uid}{"\n"}'
6f0a4c9e-2d2b-4d8e-9d4a-2b9c1f5a7e10

$ kubectl get pods -n demo -l app.kubernetes.io/instance=pg -L kubedb.com/role
NAME   READY   STATUS    RESTARTS   AGE     ROLE
pg-0   2/2     Running   0          4m51s   primary
pg-1   2/2     Running   0          3m50s   standby
pg-2   2/2     Running   0          3m46s   standby
```

KubeDB provisions the Pods with a `resizePolicy` of `NotRequired` for CPU and
memory, which is what lets the kubelet resize them without a restart:

```bash
$ kubectl get pod -n demo pg-0 -o json | jq '.spec.containers[0].resizePolicy'
[
  {
    "resourceName": "cpu",
    "restartPolicy": "NotRequired"
  },
  {
    "resourceName": "memory",
    "restartPolicy": "NotRequired"
  }
]
```

## In-Place CPU scaling (no restart, no failover)

In order to update the resources in place, create a `PostgresOpsRequest` with
`spec.verticalScaling.mode: InPlace`. Below is the YAML we are going to apply; it
raises the CPU request and limit to `1` and keeps memory at `1Gi`.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pg-scale-vertical-inplace
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: pg
  verticalScaling:
    mode: InPlace
    postgres:
      resources:
        requests:
          memory: "1Gi"
          cpu: "1"
        limits:
          memory: "1Gi"
          cpu: "1"
```

Here,

- `spec.databaseRef.name` specifies the `pg` `Postgres` database.
- `spec.type` specifies that we are performing `VerticalScaling`.
- `spec.verticalScaling.mode: InPlace` requests the in-place path. (Omitting `mode`,
  or setting it to `Restart`, uses the default restart-based path.)
- `spec.verticalScaling.postgres` is the desired postgres container resources.

Let's create it,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/scaling/vertical-scaling/in-place/yamls/pg-vertical-scaling-inplace.yaml
postgresopsrequest.ops.kubedb.com/pg-scale-vertical-inplace created
```

**Wait for the OpsRequest to succeed:**

```bash
$ watch kubectl get postgresopsrequest -n demo pg-scale-vertical-inplace
NAME                         TYPE              STATUS       AGE
pg-scale-vertical-inplace    VerticalScaling   Successful   1m12s
```

Unlike the restart path, the `KubeDB-Ops-Manager` does not evict any Pod here. It
patches the `PetSet` template, the PetSet controller drives the kubelet resize on
each Pod, and the operator waits for the resize to be actuated. The `describe`
output reflects this — there is no `PauseDatabase`/eviction-driven restart of the
members for the resize:

```bash
$ kubectl get postgresopsrequest -n demo pg-scale-vertical-inplace -o yaml | yq '.status.conditions[].type'
Progressing
UpdatePetSets
VerticalScale
Successful
```

**Verify the Pods were resized in place (not recreated):**

The clearest proof is that the Pod UID and `restartCount` are unchanged, while the
container resources are updated:

```bash
$ kubectl get pod -n demo pg-0 -o jsonpath='{.metadata.uid}{"\n"}'
6f0a4c9e-2d2b-4d8e-9d4a-2b9c1f5a7e10        # same UID as before — the Pod was not recreated

$ kubectl get pods -n demo -l app.kubernetes.io/instance=pg -L kubedb.com/role
NAME   READY   STATUS    RESTARTS   AGE     ROLE
pg-0   2/2     Running   0          9m12s   primary    # still primary — no failover
pg-1   2/2     Running   0          8m11s   standby
pg-2   2/2     Running   0          8m07s   standby

$ kubectl get pod -n demo pg-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "1",
    "memory": "1Gi"
  }
}
```

You can also confirm the kubelet actually applied the change by checking the
allocated/actual resources reported in the Pod status:

```bash
$ kubectl get pod -n demo pg-0 -o json | jq '.status.containerStatuses[0].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "1",
    "memory": "1Gi"
  }
}
```

The `RESTARTS` column stays at `0`, the UID is unchanged, and `pg-0` is still the
primary: the CPU was scaled with no restart and no failover.

## In-Place memory increase

A memory **increase** is also done in place. Create a `PostgresOpsRequest` with
`spec.verticalScaling.mode: InPlace` and the default
`spec.verticalScaling.memoryPolicy: ResizeOnly`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pg-scale-vertical-inplace-mem
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: pg
  verticalScaling:
    mode: InPlace
    memoryPolicy: ResizeOnly
    postgres:
      resources:
        requests:
          memory: "2Gi"
          cpu: "1"
        limits:
          memory: "2Gi"
          cpu: "1"
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/scaling/vertical-scaling/in-place/yamls/pg-vertical-scaling-inplace-memory.yaml
postgresopsrequest.ops.kubedb.com/pg-scale-vertical-inplace-mem created
```

With `memoryPolicy: ResizeOnly` (the default for `InPlace`):

- the memory cgroup is grown live (no restart);
- `shared_buffers` is **left unchanged** — it is a restart-only PostgreSQL
  parameter, and KubeDB intentionally does not change it here so a later restart
  will not surprise-jump it;
- the reloadable, memory-derived GUCs (`effective_cache_size`, `work_mem`,
  `maintenance_work_mem`) are re-applied live with `ALTER SYSTEM SET ...; SELECT
  pg_reload_conf();`, computed from the new memory. This applies only when KubeDB
  manages the tuning, i.e. `spec.configuration.tuning` is set on the `Postgres`
  (we enabled it above).

You can see the difference — `shared_buffers` keeps the value it started with at
`1Gi`, while `effective_cache_size` now tracks the `2Gi`:

```bash
$ kubectl exec -it -n demo pg-0 -c postgres -- psql -U postgres -c \
  "SHOW shared_buffers; SHOW effective_cache_size;"
 shared_buffers
----------------
 256MB             # unchanged — restart-only

 effective_cache_size
----------------------
 1536MB            # re-applied live (75% of the new 2Gi)
```

If you want `shared_buffers` itself to track the new memory, use
`spec.verticalScaling.memoryPolicy: Retune` instead — that requires a restart and
KubeDB runs it on the restart path (see below).

## Eligibility and fallback

In-place is **per request**: KubeDB runs the request in place only if **every**
requested change is eligible, otherwise the whole request falls back to the
restart-based path (and the OpsRequest still completes). A change is in-place
eligible when it is:

- a **CPU** change in any direction, or
- a **memory increase** with `memoryPolicy: ResizeOnly`.

The request falls back to the restart path when any of the following is true:

| Situation | Why |
| --- | --- |
| `spec.verticalScaling.memoryPolicy: Retune` | `shared_buffers` is restart-only, so retuning it needs a restart. |
| A **memory decrease** | A live shrink can be rejected by the kubelet or risk an OOM kill, so KubeDB does it via restart. |
| The cluster does not support in-place resize, or the kubelet reports the resize `Infeasible` (e.g. the node cannot fit the larger request) | KubeDB recreates the Pod so the scheduler can place it on a node with room. |

When a request that asked for `InPlace` falls back, KubeDB records an
`InPlaceResizeEligible=false` condition on the OpsRequest with the reason, and then
runs the regular vertical scaling. This covers the arbiter, the read replicas, and
the primary group — each resizes in place when eligible.

## Autoscaler

The KubeDB Postgres compute autoscaler can emit these requests for you. When the
autoscaler is configured to use in-place mode, the `PostgresOpsRequest`s it creates
carry `mode: InPlace` (with `memoryPolicy: ResizeOnly`), so routine CPU autoscaling
becomes non-disruptive. See the
[compute autoscaling guide](/docs/guides/postgres/autoscaler/compute/cluster.md).

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete postgresopsrequest -n demo pg-scale-vertical-inplace pg-scale-vertical-inplace-mem
kubectl delete postgres -n demo pg
kubectl delete ns demo
```
