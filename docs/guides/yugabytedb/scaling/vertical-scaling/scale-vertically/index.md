---
title: Scale YugabyteDB Vertically
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-scale-vertically
    name: Scale Vertically
    parent: yugabytedb-vertical-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Scale YugabyteDB Vertically

This guide changes the CPU and memory of the masters and tservers of the `yb-demo` universe.

> Note: YAML files used in this tutorial are stored in [docs/examples/yugabytedb/scaling/vertical-scaling](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/yugabytedb/scaling/vertical-scaling) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

Deploy `yb-demo` from the [quickstart](/docs/guides/yugabytedb/quickstart/quickstart.md) (masters with 1Gi, tservers with 2Gi of memory).

## Apply the VerticalScaling OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-demo-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef: {name: yb-demo}
  verticalScaling:
    master:
      resources:
        requests: {cpu: 500m, memory: 1280Mi}
        limits: {memory: 1280Mi}
    tserver:
      resources:
        requests: {cpu: "1", memory: 2560Mi}
        limits: {memory: 2560Mi}
  rolling: {masterWait: 20s, tserverWait: 20s}
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/scaling/vertical-scaling/ops-request.yaml
yugabytedbopsrequest.ops.kubedb.com/yb-demo-vscale created

$ kubectl get ybops -n demo yb-demo-vscale -w
```

The conditions are `VerticalScaling`, `UpdateResources`, `RollMasters`, `RollTServers` and `Successful`.

## What was measured

Under the same load as in [Restart](/docs/guides/yugabytedb/restart/restart.md):

| Measure | Result |
|---------|--------|
| Duration | 482 s |
| Successful calls | 613,680 |
| Failed calls | 8, all dropped connections to the restarted pod (57P01), no other failure |
| Longest period without any success | 3.3 s |

After the operation, `memory_limit_hard_bytes` was recomputed on the tserver pod: 2560Mi x 0.7 = 1879048192 bytes, as expected. "Zero downtime" means no failed calls except the connections dropped to the restarted pod.

## Cleaning up

```bash
$ kubectl delete ybops -n demo yb-demo-vscale
```
