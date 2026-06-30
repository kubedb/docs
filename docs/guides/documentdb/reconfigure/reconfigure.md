---
title: Reconfigure DocumentDB
menu:
  docs_{{ .version }}:
    identifier: dc-reconfigure-details
    name: Reconfigure DocumentDB
    parent: dc-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure DocumentDB

KubeDB lets you change the runtime configuration of a running `DocumentDB` database without
recreating it, using a `DocumentDBOpsRequest` of type `Reconfigure`. The underlying engine is
PostgreSQL (DocumentDB speaks the MongoDB wire protocol on top of it), so the tunables you pass
are PostgreSQL parameters supplied through a `user.conf` fragment.

Two modes are supported:

- **Apply a custom config** — `spec.configuration.applyConfig` merges the keys you provide into
  the running configuration.
- **Remove the custom config** — `spec.configuration.removeCustomConfig: true` drops any
  previously applied custom configuration and returns the database to its defaults.

## Before You Begin

- You need a Kubernetes cluster and the `kubectl` CLI configured to talk to it.
- Install KubeDB following the steps [here](/docs/setup/README.md).
- This tutorial uses a namespace called `demo` (`kubectl create ns demo`).
- Deploy a `DocumentDB` cluster (`documentdb-cls-sample`) and wait for it to become `Ready`.

> Note: YAML files used in this tutorial are stored in [docs/examples/documentdb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/documentdb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Inspecting the current configuration

`max_connections` is a convenient parameter to watch because it has a visible default of `100`.
You can read it from the internal PostgreSQL engine (port `9712`, backend-only) using the admin
credentials from `<db>-admin-auth`:

```bash
$ ADMINU=$(kubectl get secret -n demo documentdb-cls-sample-admin-auth -o jsonpath='{.data.username}' | base64 -d)
$ ADMINP=$(kubectl get secret -n demo documentdb-cls-sample-admin-auth -o jsonpath='{.data.password}' | base64 -d)
$ kubectl exec -n demo documentdb-cls-sample-0 -c documentdb -- \
    bash -lc "PGPASSWORD='$ADMINP' psql -h localhost -p 9712 -U '$ADMINU' -d postgres -tAc 'show max_connections'"
100
```

## Apply a custom configuration

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DocumentDBOpsRequest
metadata:
  name: documentdb-cls-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: documentdb-cls-sample
  configuration:
    applyConfig:
      user.conf: |
        max_connections=250
```

```bash
$ kubectl apply -f cluster-reconfigure.yaml
documentdbopsrequest.ops.kubedb.com/documentdb-cls-reconfigure created
```

The operator performs a careful, leader-aware rollout: it transfers Raft leadership to the first
pod, pauses the `documentdb-coordinator` so it does not trigger an automatic failover during the
restart, then evicts the pod so it comes back with the new configuration mounted:

```bash
$ kubectl get dcops -n demo documentdb-cls-reconfigure \
    -o jsonpath='{range .status.conditions[*]}{.type}={.status} :: {.message}{"\n"}{end}'
Running=True :: Reconfiguring DocumentDB Database
ReconcileDocumentDBDatabase=True :: Successfully Reconciled DocumentDB Database
TransferLeaderShipToFirstNodeBeforeCoordinatorPaused=True :: Successfully Transferred Leadership to first pod before documentdb-coordinator paused
PausePgCoordinatorBeforeCustomRestart=True :: Successfully Pause DocumentDB-Coordinator Before Custom Restart
EvictPod=True :: evict pod; ConditionStatus:True
CheckPodReady--documentdb-cls-sample-0=False :: check pod ready; ConditionStatus:False; PodName:documentdb-cls-sample-0
```

## Remove a custom configuration

Once a custom configuration has been applied, you remove it and return to the defaults with:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DocumentDBOpsRequest
metadata:
  name: documentdb-cls-reconfigure-remove
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: documentdb-cls-sample
  configuration:
    removeCustomConfig: true
```

```bash
$ kubectl apply -f cluster-reconfigure-remove.yaml
documentdbopsrequest.ops.kubedb.com/documentdb-cls-reconfigure-remove created
```

The operator performs the same leader-aware rolling restart, dropping the custom-config volume
from each pod so `max_connections` returns to its default of `100`.

> [!CAUTION]
> **Known limitation on the `pg17-0.109.0` build used to capture this guide.** The `Reconfigure`
> OpsRequest did not converge: the operator transferred leadership and paused the coordinator
> correctly, but when it recreated the first pod with the new custom-config volume, the
> referenced config `Secret` was never created, so the pod was stuck in `Init` with a
> `FailedMount` and the OpsRequest ended in `Failed`:
>
> ```bash
> $ kubectl get events -n demo --field-selector involvedObject.name=documentdb-cls-sample-0 | grep FailedMount
> Warning  FailedMount  pod/documentdb-cls-sample-0  MountVolume.SetUp failed for volume "custom-config" : secret "documentdb-cls-sample-c037f1" not found
>
> $ kubectl get dcops -n demo documentdb-cls-reconfigure
> NAME                         TYPE          STATUS   AGE
> documentdb-cls-reconfigure   Reconfigure   Failed   11m
> ```
>
> Because `Reconfigure` pauses the coordinator before restarting the primary, a mid-flight
> failure can leave the database `NotReady` with the coordinator waiting to be resumed. Since
> OpsRequests are admitted only while the database is `Ready`, a follow-up OpsRequest cannot
> clear that state — recovery requires recreating the `DocumentDB`. Validate `Reconfigure`
> against a non-production cluster on this build before relying on it. The YAML and rollout
> mechanics above are the intended workflow; provisioning custom configuration up front (see
> [Custom Configuration](/docs/guides/documentdb/configuration/using-config-file/)) is
> unaffected.

## Standalone

The same `DocumentDBOpsRequest` applies to a standalone (`replicas: 1`) instance — point
`spec.databaseRef.name` at `documentdb-sa-sample`. On this build standalone instances did not
finish bootstrapping (see the [Restart](/docs/guides/documentdb/restart/) guide), so the
standalone `Reconfigure` could not be exercised live.

## Cleaning Up

```bash
kubectl delete documentdbopsrequest -n demo documentdb-cls-reconfigure documentdb-cls-reconfigure-remove --ignore-not-found
kubectl delete documentdb -n demo documentdb-cls-sample
kubectl delete ns demo
```

## Next Steps

- Provision a database with [Custom Configuration](/docs/guides/documentdb/configuration/using-config-file/).
- [Restart](/docs/guides/documentdb/restart/) a DocumentDB database.
