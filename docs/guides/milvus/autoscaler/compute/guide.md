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

- Install the **KubeDB Autoscaler** operator and a **metrics server** in your cluster — the VPA recommender needs metrics to produce recommendations.

  ```bash
  $ kubectl get deploy metrics-server -n kube-system
  NAME             READY   UP-TO-DATE   AVAILABLE   AGE
  metrics-server   1/1     1            1           5m
  ```

- An object-storage secret named `my-release-minio` must exist in the `demo` namespace.

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/autoscaler/compute/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/autoscaler/compute/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Compute Autoscaling — Standalone Milvus

Deploy a standalone Milvus and wait until it is `Ready`. Then create a `MilvusAutoscaler` targeting the standalone `node`:

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
      podLifeTimeThreshold: 1m
      minAllowed:
        cpu: 100m
        memory: 256Mi
      maxAllowed:
        cpu: 1000m
        memory: 2Gi
      resourceDiffPercentage: 10
      controlledResources: ["cpu", "memory"]
  opsRequestOptions:
    apply: IfReady
    timeout: 5m
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/autoscaler/compute/yamls/compute-standalone.yaml
milvusautoscaler.autoscaling.kubedb.com/milvus-standalone-compute-autoscaler created
```

The autoscaler creates a `VerticalPodAutoscaler` (VPA) object. Once the VPA recommender produces a recommendation that differs from the current resources by more than `resourceDiffPercentage`, the autoscaler creates a `VerticalScaling` `MilvusOpsRequest`.

```bash
$ kubectl get milvusautoscaler -n demo
NAME                                   AGE
milvus-standalone-compute-autoscaler   59s
```

The autoscaler runs a VPA recommender (fed by the metrics server) and records the recommendation in its status. Once enough samples are collected, the `RecommendationProvided` condition becomes `True` and a target resource set is published:

```bash
$ kubectl get milvusautoscaler milvus-standalone-compute-autoscaler -n demo -o jsonpath='{.status}' | jq .
{
  "vpas": [
    {
      "conditions": [
        { "type": "RecommendationProvided", "status": "True", "lastTransitionTime": "2026-06-30T18:40:17Z" }
      ],
      "recommendation": {
        "containerRecommendations": [
          {
            "containerName": "milvus",
            "lowerBound": { "cpu": "100m", "memory": "256Mi" },
            "target":     { "cpu": "143m", "memory": "256Mi" },
            "uncappedTarget": { "cpu": "143m", "memory": "262144k" }
          }
        ]
      },
      "ref": { "containerName": "milvus", "vpaObjectName": "milvus-standalone" }
    }
  ]
}
```

Here the recommended `target` is `cpu: 143m` / `memory: 256Mi` (the standalone idles well below its `500m` request). Because the recommendation differs from the current request by more than `resourceDiffPercentage` (10%) and stays within `minAllowed`/`maxAllowed`, the autoscaler creates a `VerticalScaling` `MilvusOpsRequest`. This is recorded in the autoscaler status as a `CreateOpsRequest` condition:

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

For a distributed Milvus, the autoscaler is keyed by role (`proxy`, `mixcoord`, `datanode`, `querynode`, `streamingnode`). The sample below enables compute autoscaling for all five roles:

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
      podLifeTimeThreshold: 1m
      minAllowed:
        cpu: 100m
        memory: 256Mi
      maxAllowed:
        cpu: 1000m
        memory: 2Gi
      resourceDiffPercentage: 10
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

The behavior is identical to standalone, except a VPA object and resource recommendation are produced **per role**. When any role's recommendation differs from its current resources by more than `resourceDiffPercentage`, the autoscaler creates a `VerticalScaling` `MilvusOpsRequest` scoped to that role (see the [vertical scaling guide](/docs/guides/milvus/scaling/vertical-scaling/guide.md) for what that ops request looks like).

## Cleaning up

```bash
$ kubectl delete milvusautoscaler -n demo --all
$ kubectl delete milvus.kubedb.com -n demo milvus-standalone
$ kubectl delete ns demo
```

## Next Steps

- Learn about [storage autoscaling](/docs/guides/milvus/autoscaler/storage/guide.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
