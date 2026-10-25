---
title: YugabyteDB In-Place Vertical Scaling
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-vertical-scaling-in-place
    name: In-Place (CPU only)
    parent: yugabytedb-vertical-scaling
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# In-Place Vertical Scaling (CPU only)

> **Not verified on a cluster.** This feature has not been run on a real cluster. What was checked: the admission webhook rules of the API (`apimachinery`, branch `yugabytedb`) have unit tests, and they pass. Everything below describes what the API types and those rules say, not an observed run, so expect to find differences when you try it.

`spec.verticalScaling.mode` is `Restart` (the default) or `InPlace`.

- `Restart`: the new resources are written into the PetSets and the pods are rolled one at a time, masters first. The memory settings of YugabyteDB (`--memory_limit_hard_bytes`) are re-rendered into the flagfile.
- `InPlace`: the CPU of the **running pods** is resized without a restart.

## Why only the CPU

The memory limit of YugabyteDB is derived from the container memory by flags when the process starts (`--memory_limit_hard_bytes` and the PostgreSQL memory). A running process does not re-read it, so a new memory value needs a restart. The webhook therefore rejects an `InPlace` request that changes the memory, requests or limits, of any role:

```
spec.verticalScaling.tserver.resources.limits.memory (8Gi) differs from the current value (4Gi): only the CPU can be resized in place, use mode Restart to change the memory
```

The memory in the request must therefore equal the current value, as in the example:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-vertical-in-place
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: yb-demo
  verticalScaling:
    # Only the CPU can change in place. The memory must equal the current value.
    mode: InPlace
    tserver:
      resources:
        requests:
          cpu: "2"
          memory: 4Gi
        limits:
          cpu: "4"
          memory: 4Gi
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/scaling/vertical-scaling/ops-request-in-place.yaml
```

## Other rules

- `resources.limits.memory` must be set for every role in the request, and no request value may exceed its limit.
- At least one of `master`, `tserver` and `readReplica` must be set. `readReplica` needs `spec.topology.readReplica`.
- In-place resize of pod resources is a Kubernetes feature that has to be available on your cluster and container runtime. This page does not describe how KubeDB handles a cluster where it is not available, because that was not verified.

## Cleaning up

```bash
$ kubectl delete ybops -n demo yb-vertical-in-place
```
