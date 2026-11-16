---
title: Scale YugabyteDB Horizontally
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-scale-horizontally
    name: Scale Horizontally
    parent: yugabytedb-horizontal-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Scale YugabyteDB Horizontally

This guide scales the tservers of `yb-demo` from 3 to 6 and back to 3.

> Note: YAML files used in this tutorial are stored in [docs/examples/yugabytedb/scaling/horizontal-scaling](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/yugabytedb/scaling/horizontal-scaling) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- Deploy `yb-demo` from the [quickstart](/docs/guides/yugabytedb/quickstart/quickstart.md).
- **The PlacementPolicy must allow more tservers than zones.** The quickstart policy uses `whenUnsatisfiable: DoNotSchedule` on the zone constraint. The PetSet controller then adds a required one-pod-per-zone anti-affinity, and scaling out to 6 tservers over 3 zones left the new pods `Pending`. For the test below the policy was changed to `ScheduleAnyway` on the zone constraint (existing pods keep the old affinity until they are recreated by a [Restart](/docs/guides/yugabytedb/restart/restart.md)):

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  name: spread-across-zones
spec:
  zoneSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: ScheduleAnyway
  nodeSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: ScheduleAnyway
```

(This policy manifest is not in the examples folder because the exact variant used in the test was not saved; it is the quickstart policy with the zone constraint relaxed.)

## Scale out

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-demo-hscale-out
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef: {name: yb-demo}
  horizontalScaling:
    tserver: 6
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/scaling/horizontal-scaling/ops-request-scale-out.yaml
yugabytedbopsrequest.ops.kubedb.com/yb-demo-hscale-out created
```

Measured under load (4 workers of insert plus select): **162 s**, 204,538 successful calls, **0 failed calls**, longest period without success 2.3 s. Conditions: `ScaleTServers`, `WaitBalancerIdle`, `UpdatePlacement`, `Successful`.

## Scale in

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-demo-hscale-in
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef: {name: yb-demo}
  horizontalScaling:
    tserver: 3
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/scaling/horizontal-scaling/ops-request-scale-in.yaml
yugabytedbopsrequest.ops.kubedb.com/yb-demo-hscale-in created
```

Measured under load: **111 s**, 157,520 successful calls, 2 failed calls (one dropped connection 57P01 and one connect failure to the Service during the pod removal), longest period without success 0.28 s. Conditions: `DrainTServers`, `ScaleTServers`, `RemoveTServers`, `UpdatePlacement`, `Successful`. Afterwards the PVCs `data-0-yb-demo-tserver-3` to `-5` were removed, the blacklist was empty, and the master registry showed 3 ALIVE tservers.

## Cleaning up

```bash
$ kubectl delete ybops -n demo yb-demo-hscale-out yb-demo-hscale-in
```
