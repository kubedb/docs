---
title: YugabyteDB Compute Autoscaling
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-compute-autoscaling-guide
    name: Autoscale Compute
    parent: yugabytedb-compute-autoscaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# YugabyteDB Compute Autoscaling

> **Not verified on a cluster.** The autoscaler for YugabyteDB has not been run on a real cluster. What was checked: the controller has unit tests with fake clients (role spec selection, the VerticalScaling and VolumeExpansion requests it builds, the rules for creating ops requests, and a storage autoscaling run against a fake metrics server) and they passed on the machine that built it; the API webhook has a unit test. The API types live on the branch `yugabytedb-autoscaler` of `apimachinery`, which is not yet merged into the branch `yugabytedb`. Nothing here was observed against a real VPA recommender or real volumes.

You need a running `yb-demo` from the [quickstart](/docs/guides/yugabytedb/quickstart/quickstart.md), the KubeDB autoscaler, and a Kubernetes Vertical Pod Autoscaler recommender in the cluster.

The policy below is set per role. The fields of each role are the common compute autoscaler fields: `trigger` (`"On"` to enable), `minAllowed` and `maxAllowed` (bounds of the recommendation), `controlledResources`, `containerControlledValues`, `resourceDiffPercentage` (how far the recommendation must differ from the current resources before a request is created) and `podLifeTimeThreshold` (a pod must have run this long before its usage is trusted). `compute.nodeTopology` (`name`, `scaleUpDiffPercentage`, `scaleDownDiffPercentage`) is available for node pool aware scaling.

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: YugabyteDBAutoscaler
metadata:
  name: yb-compute-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: yb-demo
  opsRequestOptions:
    apply: IfReady
    timeout: 20m
    maxRetries: 1
  compute:
    tserver:
      trigger: "On"
      podLifeTimeThreshold: 10m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: "1"
        memory: 2Gi
      maxAllowed:
        cpu: "4"
        memory: 8Gi
      controlledResources:
        - cpu
        - memory
      containerControlledValues: RequestsAndLimits
    master:
      trigger: "On"
      podLifeTimeThreshold: 10m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 500m
        memory: 1Gi
      maxAllowed:
        cpu: "2"
        memory: 4Gi
      controlledResources:
        - cpu
        - memory
      containerControlledValues: RequestsAndLimits
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/autoscaler/yugabytedb-compute-autoscaler.yaml
```

This manifest was written from the API types for this guide; it was not applied. The container names are `yb-master` and `yb-tserver`. Because YugabyteDB needs a memory limit on each container, keep `containerControlledValues: RequestsAndLimits` and a `minAllowed.memory` that your workload can run in.

Watch the requests it creates:

```bash
$ kubectl get ybops -n demo
```

## Cleaning up

```bash
$ kubectl delete yugabytedbautoscaler -n demo yb-compute-autoscaler
```
