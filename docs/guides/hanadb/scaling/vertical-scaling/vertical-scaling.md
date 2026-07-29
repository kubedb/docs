---
title: Vertical Scaling HanaDB
menu:
  docs_{{ .version }}:
    identifier: guides-hanadb-scaling-vertical-scaling
    name: Vertical Scaling
    parent: guides-hanadb-scaling-vertical
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale HanaDB

This guide shows how to change the CPU and memory of a HanaDB database using a `HanaDBOpsRequest` of type
`VerticalScaling`. KubeDB updates the PetSet resources and performs a rolling restart (primary last for a
cluster).

> Note: The YAML files used in this tutorial are stored in [docs/examples/hanadb/scaling](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/scaling) folder in the GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Vertical Scaling Modes

KubeDB actuates vertical scaling in one of two modes, selected through the `spec.verticalScaling.mode`
field of the `HanaDBOpsRequest`:

- **`Restart`** (default): The operator patches the `PetSet` with the new resources and restarts the
  Pods (one at a time, honoring the database's failover rules) so they come back with the updated CPU
  and Memory. This works on every Kubernetes cluster.
- **`InPlace`**: The operator resizes the running containers in place using the Kubernetes
  [in-place Pod resize](https://kubernetes.io/docs/tasks/configure-pod-container/resize-container-resources/)
  (`pods/resize` subresource) — no Pod restart, so scaling happens without downtime or failover. If a
  Node cannot accommodate the new resources (the resize is reported `Infeasible`), the operator
  automatically falls back to the `Restart` behavior for that Pod.

If `spec.verticalScaling.mode` is omitted, it defaults to `Restart`.

> **Note:** `InPlace` mode relies on the Kubernetes `InPlacePodVerticalScaling` feature gate, which is
> enabled by default from Kubernetes v1.33. On older clusters, or when the feature gate is disabled,
> use `Restart` mode.

## Before You Begin

- Install the KubeDB Provisioner and Ops-manager operators following the steps [here](/docs/setup/README.md).
- Deploy a `hanadb-cluster` System Replication database (see [Restart](/docs/guides/hanadb/restart/restart.md#deploy-a-hanadb-system-replication-cluster)) in namespace `demo`.

## Check Resources Before Scaling

```bash
$ kubectl get pod -n demo hanadb-cluster-0 -o json | jq '.spec.containers[] | select(.name=="hanadb") | .resources'
{
  "limits": {
    "cpu": "4",
    "memory": "14Gi"
  },
  "requests": {
    "cpu": "1500m",
    "memory": "8Gi"
  }
}
```

## Create a VerticalScaling HanaDBOpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hdbops-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: hanadb-cluster
  verticalScaling:
    hanadb:
      resources:
        requests:
          cpu: "2100m"
          memory: "8448Mi"
        limits:
          cpu: "4"
          memory: "14Gi"
  timeout: 30m
  apply: IfReady
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/scaling/system-replication-vertical-scaling.yaml
hanadbopsrequest.ops.kubedb.com/hdbops-vscale created
```

Here,

- `spec.verticalScaling.hanadb.resources` are the desired resources for the `hanadb` container. You can
  also scale the `coordinator` and `exporter` containers via `spec.verticalScaling.coordinator` and
  `spec.verticalScaling.exporter`.
- `spec.verticalScaling.mode` specifies how the scaling is actuated — `Restart` (default, restarts the Pods) or `InPlace` (resizes the running Pods without a restart, falling back to restart if a Node can't fit the new resources). See [Vertical Scaling Modes](#vertical-scaling-modes).

## Verify the Scaling

```bash
$ kubectl get hdbops -n demo hdbops-vscale
NAME            TYPE              STATUS       AGE
hdbops-vscale   VerticalScaling   Successful   4m22s
```

```bash
$ kubectl describe hdbops -n demo hdbops-vscale
...
Status:
  Conditions:
    Message:  HanaDB ops-request has started to vertically scaling the HanaDB nodes
    Reason:   VerticalScaling
    Status:   True
    Type:     VerticalScaling
    Message:  Successfully updated PetSets Resources
    Reason:   UpdatePetSets
    Status:   True
    Type:     UpdatePetSets
    Message:  Successfully Restarted Pods With Resources
    Reason:   RestartPods
    Status:   True
    Type:     RestartPods
    Message:  Successfully completed the vertical scaling for HanaDB
    Reason:   Successful
    Status:   True
    Type:     Successful
  Phase:      Successful
```

Confirm the new resources are in effect on the PetSet and pods:

```bash
$ kubectl get petset -n demo hanadb-cluster \
  -o jsonpath='{range .spec.template.spec.containers[?(@.name=="hanadb")]}{.name}{": "}{.resources}{"\n"}{end}'
hanadb: {"limits":{"cpu":"4","memory":"14Gi"},"requests":{"cpu":"2100m","memory":"8448Mi"}}

$ kubectl get pod -n demo hanadb-cluster-0 -o json | jq '.spec.containers[] | select(.name=="hanadb") | .resources'
{
  "limits": {
    "cpu": "4",
    "memory": "14Gi"
  },
  "requests": {
    "cpu": "2100m",
    "memory": "8448Mi"
  }
}
```

### In-Place Vertical Scaling

To resize the Pods **without a restart**, set `spec.verticalScaling.mode` to `InPlace` in the
`HanaDBOpsRequest`. The operator resizes the running containers via the Kubernetes `pods/resize`
subresource and only restarts a Pod if its Node cannot accommodate the new resources.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hdbops-vscale-inplace
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: hanadb-cluster
  verticalScaling:
    mode: InPlace
    hanadb:
      resources:
        requests:
          cpu: "2100m"
          memory: "8448Mi"
        limits:
          cpu: "4"
          memory: "14Gi"
  timeout: 30m
  apply: IfReady
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/scaling/system-replication-vertical-scaling-inplace.yaml
hanadbopsrequest.ops.kubedb.com/hdbops-vscale-inplace created
```

Apply it the same way as above; the resources update in place with no Pod restart.

## Cleaning Up

```bash
$ kubectl delete hdbops -n demo hdbops-vscale
$ kubectl delete hanadb.kubedb.com -n demo hanadb-cluster
$ kubectl delete ns demo
```

## Next Steps

- [Expand the volume](/docs/guides/hanadb/volume-expansion/volume-expansion.md) of a HanaDB.
- Review the [HanaDBOpsRequest CRD](/docs/guides/hanadb/concepts/opsrequest.md).
