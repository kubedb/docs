---
title: Vertical Scaling Milvus
menu:
  docs_{{ .version }}:
    identifier: milvus-scaling-vertical-scaling-guide
    name: Guide
    parent: milvus-scaling-vertical-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scaling Milvus

This guide will show you how to use the `KubeDB` Ops-manager operator to update the resources (CPU/memory) of a Milvus database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/milvus/scaling/vertical-scaling/overview.md)

- Complete the dependency setup from [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md). It installs MinIO, creates the `my-release-minio` secret, and installs the etcd operator required by Milvus.

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/scaling/vertical-scaling/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/scaling/vertical-scaling/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Vertical Scaling Standalone Milvus

Deploy a standalone Milvus and wait until it is `Ready`. By default the standalone workload requests `cpu: 500m` / `memory: 1Gi`:

```bash
$ kubectl get petset milvus-standalone -n demo -o jsonpath='{.spec.template.spec.containers[0].resources}'
{"limits":{"memory":"1Gi"},"requests":{"cpu":"500m","memory":"1Gi"}}
```

### Apply the VerticalScaling OpsRequest

`vertical-scaling-standalone.yaml`

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: vertical-scaling-standalone
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: milvus-standalone
  verticalScaling:
    node:
      resources:
        requests:
          memory: "2Gi"
          cpu: "1"
        limits:
          memory: "2Gi"
          cpu: "1"
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.verticalScaling.node` carries the new resources for the **standalone** workload (use the `node` key for standalone).
- `spec.verticalScaling.mode` specifies how the scaling is actuated — `Restart` (default, restarts the Pods) or `InPlace` (resizes the running Pods without a restart, falling back to restart if a Node can't fit the new resources). See [Vertical Scaling Modes](/docs/guides/milvus/scaling/vertical-scaling/overview.md#vertical-scaling-modes).

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/scaling/vertical-scaling/yamls/vertical-scaling-standalone.yaml
milvusopsrequest.ops.kubedb.com/vertical-scaling-standalone created
```

### Watch Progress

```bash
$ kubectl get milvusopsrequest vertical-scaling-standalone -n demo
NAME                          TYPE              STATUS       AGE
vertical-scaling-standalone   VerticalScaling   Successful   56s
```

```bash
$ kubectl describe milvusopsrequest vertical-scaling-standalone -n demo
...
Status:
  Conditions:
    Message:  Milvus ops-request has started to vertically scale the Milvus nodes
    Reason:   VerticalScaling
    Type:     VerticalScaling
    Message:  Successfully updated PetSets Resources
    Reason:   UpdatePetSets
    Type:     UpdatePetSets
    Message:  check pod running; ConditionStatus:True; PodName:milvus-standalone-0
    Type:     CheckPodRunning--milvus-standalone-0
    Message:  Successfully Restarted Pods With Resources
    Reason:   RestartPods
    Type:     RestartPods
    Message:  Successfully completed the vertical scaling for Milvus
    Reason:   Successful
    Type:     Successful
  Phase:      Successful
```

### Verify the New Resources

Both the `Milvus` spec and the PetSet pod template now carry the new resources:

```bash
$ kubectl get milvuses.kubedb.com milvus-standalone -n demo -o jsonpath='{.spec.podTemplate.spec.containers[0].resources}'
{"limits":{"cpu":"1","memory":"2Gi"},"requests":{"cpu":"1","memory":"2Gi"}}

$ kubectl get petset milvus-standalone -n demo -o jsonpath='{.spec.template.spec.containers[0].resources}'
{"limits":{"cpu":"1","memory":"2Gi"},"requests":{"cpu":"1","memory":"2Gi"}}
```

### In-Place Vertical Scaling

To resize the Pods **without a restart**, set `spec.verticalScaling.mode` to `InPlace` in the
`MilvusOpsRequest`. The operator resizes the running containers via the Kubernetes `pods/resize`
subresource and only restarts a Pod if its Node cannot accommodate the new resources.

`vertical-scaling-standalone-inplace.yaml`

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: vertical-scaling-standalone-inplace
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: milvus-standalone
  verticalScaling:
    mode: InPlace
    node:
      resources:
        requests:
          memory: "2Gi"
          cpu: "1"
        limits:
          memory: "2Gi"
          cpu: "1"
  timeout: 5m
  apply: IfReady
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/scaling/vertical-scaling/yamls/vertical-scaling-standalone-inplace.yaml
milvusopsrequest.ops.kubedb.com/vertical-scaling-standalone-inplace created
```

Apply it the same way as above; the resources update in place with no Pod restart.

## Vertical Scaling Distributed Milvus

For a distributed Milvus, set the resources per role under `spec.verticalScaling`. You can scale several roles in a single OpsRequest. The sample below scales `mixcoord` and `proxy`:

`vertical-scaling-distributed.yaml`

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: vertical-scaling
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: milvus-cluster
  verticalScaling:
    mixcoord:
      resources:
        requests:
          memory: "2Gi"
          cpu: "1"
        limits:
          memory: "2Gi"
          cpu: "1"
    proxy:
      resources:
        requests:
          memory: "2Gi"
          cpu: "1"
        limits:
          memory: "2Gi"
          cpu: "1"
  timeout: 5m
  apply: IfReady
```

The same approach applies to `datanode`, `querynode` and `streamingnode`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/scaling/vertical-scaling/yamls/vertical-scaling-distributed.yaml
milvusopsrequest.ops.kubedb.com/vertical-scaling created

$ kubectl get milvusopsrequest vertical-scaling -n demo
NAME               TYPE              STATUS       AGE
vertical-scaling   VerticalScaling   Successful   36s
```

Both the `mixcoord` and `proxy` PetSets now carry the new resources (other roles are unchanged):

```bash
$ kubectl get petset milvus-cluster-mixcoord -n demo -o jsonpath='{.spec.template.spec.containers[0].resources}'
{"limits":{"cpu":"1","memory":"2Gi"},"requests":{"cpu":"1","memory":"2Gi"}}

$ kubectl get petset milvus-cluster-proxy -n demo -o jsonpath='{.spec.template.spec.containers[0].resources}'
{"limits":{"cpu":"1","memory":"2Gi"},"requests":{"cpu":"1","memory":"2Gi"}}
```

## Cleaning up

```bash
$ kubectl delete milvusopsrequest -n demo vertical-scaling-standalone
$ kubectl delete milvus.kubedb.com -n demo milvus-standalone
$ kubectl delete ns demo
```

## Next Steps

- Learn about [horizontal scaling](/docs/guides/milvus/scaling/horizontal-scaling/guide.md) of a distributed Milvus.
- Detail concepts of [Milvus object](/docs/guides/milvus/concepts/milvus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
