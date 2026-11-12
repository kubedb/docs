---
title: Milvus Compute Autoscaling
menu:
  docs_{{ .version }}:
    identifier: milvus-autoscaler-compute-guide
    name: Guide
    parent: milvus-autoscaler-compute
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Milvus Compute Autoscaling

This guide will show you how to use the `KubeDB` Autoscaler operator to autoscale the compute resources (CPU/memory) of a Milvus database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [MilvusAutoscaler](/docs/guides/milvus/concepts/milvusautoscaler.md)
  - [Compute Autoscaling Overview](/docs/guides/milvus/autoscaler/compute/overview.md)

- Install the **KubeDB Autoscaler** operator and a **metrics server** in your cluster — the autoscaler's built-in recommender needs metrics to produce recommendations.

  ```bash
  $ kubectl get deploy metrics-server -n kube-system
  NAME             READY   UP-TO-DATE   AVAILABLE   AGE
  metrics-server   1/1     1            1           5m
  ```

- Complete the dependency setup from [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md). It installs MinIO and creates the object-storage secrets (`milvus-storage-config` for `milvus-standalone`, `milvus-cluster-storage-config` for `milvus-cluster`).

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/autoscaler/compute/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/autoscaler/compute/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Compute Autoscaling — Standalone Milvus

Deploy the [standalone quickstart](/docs/guides/milvus/quickstart/standalone.md) manifest and wait until it is `Ready`:

> If you deleted a Milvus of the same name in an earlier guide, delete its object-storage prefix first, as shown in [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md#delete-a-milvus-instances-object-storage-data). Otherwise the new database can hang in `Provisioning`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/quickstart/yamls/standalone.yaml
milvus.kubedb.com/milvus-standalone created

$ kubectl get milvuses.kubedb.com -n demo milvus-standalone -w
```

Then create a `MilvusAutoscaler` targeting the standalone `node`:

`compute-standalone.yaml`

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MilvusAutoscaler
metadata:
  name: milvus-standalone-compute-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: milvus-standalone
  compute:
    node:
      trigger: "On"
      podLifeTimeThreshold: 10m
      minAllowed:
        cpu: 100m
        memory: 1Gi
      maxAllowed:
        cpu: 1000m
        memory: 2Gi
      resourceDiffPercentage: 20
      controlledResources: ["cpu", "memory"]
  opsRequestOptions:
    apply: IfReady
    timeout: 5m
```

Here,

- `minAllowed.memory: 1Gi` keeps the autoscaler from shrinking Milvus below what it needs. The generated OpsRequest sets the memory limit equal to the request, so a lower floor (for example `256Mi`) leaves no headroom and Milvus is OOM-killed under real load.
- `podLifeTimeThreshold: 10m` and `resourceDiffPercentage: 20` damp the autoscaler. Every OpsRequest it creates restarts the pods, and a freshly started Milvus briefly uses much more CPU. With a short `podLifeTimeThreshold` and a small `resourceDiffPercentage`, that startup spike immediately triggers another re-scale (and another restart). Waiting 10 minutes after a pod starts and ignoring differences below 20% prevents this.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/autoscaler/compute/yamls/compute-standalone.yaml
milvusautoscaler.autoscaling.kubedb.com/milvus-standalone-compute-autoscaler created
```

The autoscaler runs its own built-in recommender (it does not create a `VerticalPodAutoscaler` object). Once the recommender produces a recommendation that differs from the current resources by more than `resourceDiffPercentage`, the autoscaler creates a `VerticalScaling` `MilvusOpsRequest`.

> **Each autoscaler-created OpsRequest restarts the pods** of the role it scales (it uses `Restart` mode). For a standalone Milvus, that means a short outage every time the autoscaler re-scales.

```bash
$ kubectl get milvusautoscaler -n demo
NAME                                   AGE
milvus-standalone-compute-autoscaler   59s
```

The recommender (fed by the metrics server) records its recommendation in the autoscaler status. Once enough samples are collected, the `RecommendationProvided` condition becomes `True` and a target resource set is published. The status has this shape (values and the large `checkpoints` array elided):

```bash
$ kubectl get milvusautoscaler milvus-standalone-compute-autoscaler -n demo -o jsonpath='{.status}' | jq .
{
  "checkpoints": [ ... ],
  "conditions": [ ... ],
  "vpas": [
    {
      "conditions": [
        { "type": "RecommendationProvided", "status": "True", ... }
      ],
      "recommendation": {
        "containerRecommendations": [
          {
            "containerName": "milvus",
            "lowerBound": { "cpu": ..., "memory": ... },
            "target": { "cpu": ..., "memory": ... },
            "uncappedTarget": { "cpu": ..., "memory": ... },
            ...
          }
        ]
      },
      "vpaName": "milvus-standalone"
    }
  ]
}
```

To see just the recommended target:

```bash
$ kubectl get milvusautoscaler milvus-standalone-compute-autoscaler -n demo -o jsonpath='{.status.vpas[0].recommendation.containerRecommendations[0].target}'
```

An idle standalone uses far less CPU than its `500m` request, so the CPU target is lower, while the memory target is held at `minAllowed` (`1Gi`). When the recommendation differs from the current request by more than `resourceDiffPercentage` (20%) and stays within `minAllowed`/`maxAllowed`, the autoscaler creates a `VerticalScaling` `MilvusOpsRequest`. This is recorded in the autoscaler status as a `CreateOpsRequest` condition:

```bash
$ kubectl get milvusautoscaler milvus-standalone-compute-autoscaler -n demo \
    -o jsonpath='{.status.conditions[?(@.type=="CreateOpsRequest")].message}'
Successfully created MilvusOpsRequest demo/mvops-milvus-standalone-xqwkhv

$ kubectl get milvusopsrequest -n demo
NAME                             TYPE              STATUS        AGE
mvops-milvus-standalone-xqwkhv   VerticalScaling   Progressing   2s
```

The Ops-manager then applies the vertical scaling exactly as in the [vertical scaling guide](/docs/guides/milvus/scaling/vertical-scaling/guide.md), right-sizing the pod to the recommended resources.

## Compute Autoscaling — Distributed Milvus

For a distributed Milvus, the autoscaler is keyed by role (`proxy`, `mixcoord`, `datanode`, `querynode`, `streamingnode`). The sample below enables compute autoscaling for all five roles.

Deploy the [distributed quickstart](/docs/guides/milvus/quickstart/distributed.md) manifest and wait until it is `Ready`:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/quickstart/yamls/distributed.yaml
milvus.kubedb.com/milvus-cluster created

$ kubectl get milvuses.kubedb.com -n demo milvus-cluster -w
```

`compute-distributed.yaml`

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MilvusAutoscaler
metadata:
  name: milvus-compute-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: milvus-cluster
  compute:
    proxy:
      trigger: "On"
      podLifeTimeThreshold: 10m
      minAllowed:
        cpu: 100m
        memory: 1Gi
      maxAllowed:
        cpu: 1000m
        memory: 2Gi
      resourceDiffPercentage: 20
      controlledResources: ["cpu", "memory"]
    mixcoord: { ... }
    datanode: { ... }
    querynode: { ... }
    streamingnode: { ... }
  opsRequestOptions:
    apply: IfReady
    timeout: 5m
```

(Each role block is identical to the `proxy` block above; see the full file in the `yamls` folder.)

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/autoscaler/compute/yamls/compute-distributed.yaml
milvusautoscaler.autoscaling.kubedb.com/milvus-compute-autoscaler created
```

The behavior is identical to standalone, except a resource recommendation is produced **per role** (one `status.vpas[]` entry per role, for example `vpaName: milvus-cluster-proxy`). When any role's recommendation differs from its current resources by more than `resourceDiffPercentage`, the autoscaler creates a `VerticalScaling` `MilvusOpsRequest` scoped to that role (see the [vertical scaling guide](/docs/guides/milvus/scaling/vertical-scaling/guide.md) for what that ops request looks like). The OpsRequests run one at a time, and each restarts the pods of its role.

## Cleaning up

The OpsRequests created by a `MilvusAutoscaler` are owned by it, so deleting the autoscaler also deletes them, **including one that is still running**. An OpsRequest deleted mid-flight can leave the database paused or with pods that don't match its spec. Before deleting the autoscaler, wait until none of its OpsRequests is `Progressing`:

```bash
$ kubectl get milvusopsrequest -n demo
```

Then clean up:

```bash
$ kubectl delete milvusautoscaler -n demo --all
$ kubectl patch -n demo milvus.kubedb.com milvus-standalone -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl patch -n demo milvus.kubedb.com milvus-cluster -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete milvus.kubedb.com -n demo milvus-standalone milvus-cluster
$ kubectl delete ns demo
```

## Next Steps

- Learn about [storage autoscaling](/docs/guides/milvus/autoscaler/storage/guide.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
