---
title: Vertical Scaling DocumentDB
menu:
  docs_{{ .version }}:
    identifier: guides-documentdb-scaling-vertical-details
    name: Vertical Scaling
    parent: guides-documentdb-scaling-vertical
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scaling of a DocumentDB Cluster

Vertical scaling changes the **CPU and memory** allocated to the containers in a `DocumentDB`
database. A DocumentDB pod runs two containers that can be sized independently:

- `documentdb` — the database engine (MongoDB wire protocol over internal PostgreSQL).
- `documentdb-coordinator` — the Raft member that handles leader election and membership.

A `DocumentDBOpsRequest` of type `VerticalScaling` lets you set new resource requests/limits for
either or both. The operator rolls the change out pod by pod (evicting standbys first, the
primary last) so the cluster stays available.

## Before You Begin

- You need a Kubernetes cluster and the `kubectl` CLI configured to talk to it.
- Install KubeDB following the steps [here](/docs/setup/README.md).
- This tutorial uses a namespace called `demo` (`kubectl create ns demo`).
- Deploy a `DocumentDB` cluster (`documentdb-cls-sample`) and wait for it to become `Ready`.

> Note: YAML files used in this tutorial are stored in [docs/examples/documentdb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/documentdb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Resources before

```bash
kubectl get docdb -n demo documentdb-cls-sample \
    -o jsonpath='{range .spec.podTemplate.spec.containers[*]}{.name}: requests={.resources.requests} limits={.resources.limits}{"\n"}{end}'
```
documentdb: requests={"cpu":"500m","memory":"2Gi"} limits={"memory":"2Gi"}
documentdb-coordinator: requests={"cpu":"200m","memory":"256Mi"} limits={"memory":"256Mi"}

## Create the VerticalScaling OpsRequest

This request bumps the `documentdb` engine and, at the same time, *lowers* the coordinator's CPU
request — both containers are addressed in one OpsRequest:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DocumentDBOpsRequest
metadata:
  name: documentdb-cls-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: documentdb-cls-sample
  verticalScaling:
    documentdb:
      resources:
        requests:
          cpu: 600m
          memory: 2.5Gi
        limits:
          cpu: "1"
          memory: 2.5Gi
    coordinator:
      resources:
        requests:
          cpu: 100m
          memory: 256Mi
```

```bash
kubectl apply -f cluster-vertical-scaling.yaml
```
documentdbopsrequest.ops.kubedb.com/documentdb-cls-vscale created

```bash
kubectl get dcops -n demo documentdb-cls-vscale
```
NAME                    TYPE              STATUS       AGE
documentdb-cls-vscale   VerticalScaling   Successful   3m33s

The status conditions show the PetSet being patched and each pod being evicted and re-checked
for readiness before the next is touched:

```bash
kubectl get dcops -n demo documentdb-cls-vscale \
    -o jsonpath='{range .status.conditions[*]}{.type}={.status} :: {.message}{"\n"}{end}'
```
Running=True :: Vertical Scaling is in progress
UpdatePetSets=True :: Successfully updated petsets resources
EvictPod=True :: evict pod; ConditionStatus:True
CheckPodReady=True :: check pod ready; ConditionStatus:True
CheckReplicaFunc=True :: check replica func; ConditionStatus:True
VerticalScale=True :: VerticalScaleSucceeded
RestartReadReplicas=True :: Successfully Restarted Read Replicas
Successful=True :: Successfully Vertically Scaled Database

## Resources after

Both containers reflect the new sizing (note `2.5Gi` is normalized to its binary equivalent
`2560Mi`, and the `documentdb` container now carries a CPU limit of `1`):

```bash
kubectl get docdb -n demo documentdb-cls-sample \
    -o jsonpath='{range .spec.podTemplate.spec.containers[*]}{.name}: requests={.resources.requests} limits={.resources.limits}{"\n"}{end}'
```
documentdb: requests={"cpu":"600m","memory":"2560Mi"} limits={"cpu":"1","memory":"2560Mi"}
documentdb-coordinator: requests={"cpu":"100m","memory":"256Mi"} limits={"memory":"256Mi"}

The live pod spec matches — the change propagated all the way to the running containers:

```bash
kubectl get pod -n demo documentdb-cls-sample-0 \
    -o jsonpath='{range .spec.containers[*]}{.name}: req={.resources.requests} lim={.resources.limits}{"\n"}{end}'
```
documentdb: req={"cpu":"600m","memory":"2560Mi"} lim={"cpu":"1","memory":"2560Mi"}
documentdb-coordinator: req={"cpu":"100m","memory":"256Mi"} lim={"memory":"256Mi"}

The cluster remains healthy and accepts MongoDB traffic after the rollout:

```bash
PASS=$(kubectl get secret -n demo documentdb-cls-sample-auth -o jsonpath='{.data.password}' | base64 -d)
```

```bash
kubectl exec -n demo documentdb-cls-sample-0 -c documentdb -- \
    mongosh "mongodb://default_user:${PASS}@localhost:10260/?tls=true&tlsAllowInvalidCertificates=true" \
    --quiet --eval 'db.runCommand({ ping: 1 })'
```
{ ok: 1 }

## Standalone

The same `DocumentDBOpsRequest` works for a standalone (`replicas: 1`) instance — point
`spec.databaseRef.name` at the standalone database (`documentdb-sa-sample`) and address the
`documentdb` (and optionally `coordinator`) container under `spec.verticalScaling`.

> [!NOTE]
> On the build used to capture this guide (`pg17-0.109.0`), standalone instances did not finish
> bootstrapping (the standalone PetSet omits the `documentdb-coordinator` sidecar, so the
> internal PostgreSQL is never initialized and the database never reaches `Ready`). Because
> OpsRequests are admitted only against a `Ready` database, the standalone variant could not be
> exercised live; the cluster procedure above applies verbatim once a standalone instance is
> healthy.

## Cleaning Up

```bash
kubectl delete documentdbopsrequest -n demo documentdb-cls-vscale
kubectl delete documentdb -n demo documentdb-cls-sample
kubectl delete ns demo
```

## Next Steps

- [Horizontal scaling](/docs/guides/documentdb/scaling/horizontal-scaling/) of a DocumentDB cluster.
- [Compute autoscaling](/docs/guides/documentdb/autoscaler/compute/) of a DocumentDB cluster.
