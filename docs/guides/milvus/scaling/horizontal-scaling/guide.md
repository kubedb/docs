---
title: Horizontal Scaling Milvus
menu:
  docs_{{ .version }}:
    identifier: milvus-scaling-horizontal-scaling-guide
    name: Guide
    parent: milvus-scaling-horizontal-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scaling Milvus

This guide will show you how to use the `KubeDB` Ops-manager operator to horizontally scale the roles of a distributed Milvus database.

> **Horizontal scaling is distributed-only.** A `Standalone` Milvus is a single all-in-one workload (one PetSet, one replica) and cannot be horizontally scaled. To scale out, deploy Milvus in `Distributed` mode.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/milvus/scaling/horizontal-scaling/overview.md)

- An object-storage secret named `my-release-minio` must exist in the `demo` namespace.

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/scaling/horizontal-scaling/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/scaling/horizontal-scaling/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy a Distributed Milvus

Deploy the distributed database (`milvus-cluster`) and wait until it is `Ready` (see the [distributed quickstart](/docs/guides/milvus/quickstart/distributed.md)). By default each role runs a single replica.

```bash
$ kubectl get petset milvus-cluster-proxy milvus-cluster-streamingnode -n demo -o custom-columns=NAME:.metadata.name,REPLICAS:.spec.replicas
NAME                           REPLICAS
milvus-cluster-proxy           1
milvus-cluster-streamingnode   1
```

## Apply the HorizontalScaling OpsRequest

The sample changes only `proxy` and `streamingnode`. (The provided sample requests `1` for each; because the default is already `1`, this walkthrough requests `2` to demonstrate a scale-up.)

`horizontal-scaling-distributed.yaml`

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: milvus-hscale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: milvus-cluster
  horizontalScaling:
    topology:
      proxy: 2
      streamingnode: 2
```

Here, `spec.horizontalScaling.topology` carries the desired replica count per role. The API also accepts `mixcoord`, `querynode` and `dataNode`; this sample only scales `proxy` and `streamingnode`, but the other roles are scaled the same way.

```bash
$ kubectl apply -f horizontal-scaling-distributed.yaml
milvusopsrequest.ops.kubedb.com/milvus-hscale-up created
```

## Watch Progress and Verify

```bash
$ kubectl get milvusopsrequest milvus-hscale-up -n demo
NAME               TYPE                STATUS       AGE
milvus-hscale-up   HorizontalScaling   Successful   57s
```

```bash
$ kubectl describe milvusopsrequest milvus-hscale-up -n demo
...
Status:
  Conditions:
    Message:  Milvus ops-request has started to horizontally scale the Milvus nodes
    Reason:   HorizontalScaling
    Type:     HorizontalScaling
    Message:  Successfully Scaled Up proxy
    Reason:   ScaleUpProxy
    Type:     ScaleUpProxy
    Message:  pod readyproxy; ConditionStatus:True; PodName:milvus-cluster-proxy-1
    Type:     PodReadyproxy--milvus-cluster-proxy-1
    Message:  Successfully Scaled Up streamingnode
    Reason:   ScaleUpStreamingNode
    Type:     ScaleUpStreamingNode
  Phase:      Successful
```

Both roles now run two replicas:

```bash
$ kubectl get petset milvus-cluster-proxy milvus-cluster-streamingnode -n demo -o custom-columns=NAME:.metadata.name,REPLICAS:.spec.replicas
NAME                           REPLICAS
milvus-cluster-proxy           2
milvus-cluster-streamingnode   2

$ kubectl get pods -n demo -l app.kubernetes.io/instance=milvus-cluster | grep -E 'proxy|streamingnode'
milvus-cluster-proxy-0           1/1     Running   0          70s
milvus-cluster-proxy-1           1/1     Running   0          39s
milvus-cluster-streamingnode-0   1/1     Running   0          119s
milvus-cluster-streamingnode-1   1/1     Running   0          18s
```

> Scaling **down** works the same way — set lower replica counts in `spec.horizontalScaling.topology`.

## Cleaning up

```bash
$ kubectl delete milvusopsrequest -n demo milvus-hscale-up
$ kubectl delete milvus.kubedb.com -n demo milvus-cluster
$ kubectl delete ns demo
```

## Next Steps

- Learn about [vertical scaling](/docs/guides/milvus/scaling/vertical-scaling/guide.md) of a Milvus database.
- Detail concepts of [Milvus object](/docs/guides/milvus/concepts/milvus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
