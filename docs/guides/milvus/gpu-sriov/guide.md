---
title: Run Milvus on GPUs and SR-IOV networks
menu:
  docs_{{ .version }}:
    identifier: milvus-gpu-sriov-guide
    name: Guide
    parent: milvus-gpu-sriov
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Milvus on GPUs and SR-IOV networks

KubeDB can schedule Milvus roles onto GPU nodes and attach them to an SR-IOV secondary network through Multus, including RDMA and GPUDirect Storage (GDS) devices. This guide shows the fields that control both, what the operator renders from them, and how to check that a running cluster is using them.

## Before You Begin

- You should be familiar with the [Milvus](/docs/guides/milvus/concepts/milvus.md) CRD and a [Distributed](/docs/guides/milvus/quickstart/distributed.md) deployment.
- For GPUs: the NVIDIA device plugin (or GPU Operator) must advertise a GPU resource, by default `nvidia.com/gpu`.
- For SR-IOV: Multus, the SR-IOV device plugin and a `NetworkAttachmentDefinition` (NAD) in the namespace of the Milvus. Creating these is a cluster-admin task outside KubeDB.

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/gpu-sriov/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/gpu-sriov/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Use a GPU build of Milvus

GPU requests are only accepted on a `MilvusVersion` that declares `spec.db.gpu.supported: true`. The catalog ships these as `-gpu` entries:

```bash
$ kubectl get milvusversion -o custom-columns=NAME:.metadata.name,GPU:.spec.db.gpu.supported,IMAGE:.spec.db.image
NAME         GPU      IMAGE
2.6.15       <none>   ghcr.io/appscode-images/milvus:2.6.15
2.6.15-gpu   true     ghcr.io/appscode-images/milvus:2.6.15-gpu
...
```

The webhook rejects a Milvus that requests a GPU anywhere (`spec.gpu`, a role's or group's `gpu`, or an `nvidia.com/gpu` resource written into a `podTemplate`) unless its version is a GPU build. A `MilvusOpsRequest` of type `UpdateVersion` is held to the same rule, so a GPU database cannot be moved onto a CPU-only image.

## Request GPUs per role

In a Distributed Milvus, each role takes a `gpu` block:

| Field | Default | Meaning |
|---|---|---|
| `resourceName` | `nvidia.com/gpu` | Extended resource to request |
| `count` | `1` | GPUs per pod, set in both requests and limits |
| `nodeSelector` | none | Added to the pod's `nodeSelector` |
| `tolerations` | none | Appended to the pod's tolerations |

QueryNode (search) and DataNode (index build) are the roles that use a GPU. Your own CPU and memory requests in the role's `podTemplate` are kept; only the GPU resource is added.

### Spread one role across several GPU pools

`gpu.nodeSelector` is a conjunction of labels, so it can only point a role at one pool. To run, for example, six QueryNodes with two on L40S nodes and four on H200 nodes, use a PetSet `PlacementPolicy` with node affinity. It assigns pods to each domain by count:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  name: milvus-querynode-gpu-pools
spec:
  affinity:
    nodeAffinity:
      - topologyKey: nvidia.com/gpu.product
        whenUnsatisfiable: DoNotSchedule
        domains:
          - values: ["NVIDIA-L40S"]
            replicas: "2"
          - values: ["NVIDIA-H200"]
            replicas: "4"
```

Refer to it from the role's `podTemplate.spec.podPlacementPolicy` and leave `gpu.nodeSelector` unset for that role. `PlacementPolicy` is cluster-scoped.

Alternatively, split the role into `groups`, one per pool, each with its own `gpu.nodeSelector`. Every group is a separate PetSet, but Milvus does not see them: all QueryNodes still join the same Milvus resource group. When `groups` is set, the role-level `replicas`, `podTemplate`, `network` and `gpu` are ignored.

## Attach roles to an SR-IOV network

Each role takes a `network.sriov` block:

| Field | Required | Meaning |
|---|---|---|
| `attachmentRef` | yes | Name of the NAD, in the same namespace as the Milvus |
| `resourceName` | yes | SR-IOV VF resource advertised by the device plugin |
| `rdmaResourceName` | no | RDMA resource, if the VF is used for RDMA |
| `interface` | no | Interface name inside the pod, default `net1` |
| `gds.attachmentRef`, `gds.resourceName` | no | A second attachment for GPUDirect Storage, default interface `net2` |

For each role with `network.sriov`, the operator:

1. Adds the Multus annotation `k8s.v1.cni.cncf.io/networks` to the pods, and one unit of each named resource to the `milvus` container.
2. Adds the init container `milvus-sriov-user-config-init`. It waits up to 30 seconds for an IPv4 address on the interface and writes `MILVUS_CONF_<SECTION>_IP` exports. The `milvus` container sources them before it starts, so the role registers its SR-IOV address, not the pod IP, in the meta etcd. MixCoord advertises under `rootCoord`, `queryCoord`, `dataCoord` and `indexCoord`; StreamingNode under `streamingNode` and `queryNode`.
3. Adds a readiness probe that fails when the interface loses carrier or its address, or when Milvus stops answering `/healthz` on that address.

Set `network.sriov` on all five roles. Every role dials the others by their registered address, so a role left on the pod network may be unable to reach the rest. The webhook warns when only some roles have it.

Do not add your own advertise-IP init container to the `podTemplate`. Init containers in the `podTemplate` are kept and would run next to the operator's.

### Security context and RDMA

By default the `milvus` container runs as non-root with all capabilities dropped. RDMA verbs commonly need `IPC_LOCK`. Set it in the role's `podTemplate`; the defaults only fill fields you leave empty:

```yaml
podTemplate:
  spec:
    containers:
      - name: milvus
        securityContext:
          capabilities:
            add: ["IPC_LOCK"]
```

On OpenShift, set `spec.restrictedSCC: true` to let the SCC assign the user and group instead of the catalog's `runAsUser`.

## Deploy

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/gpu-sriov/yamls/placement-policy.yaml
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/gpu-sriov/yamls/milvus-gpu-sriov.yaml
```

The example uses YAML anchors (`&sriov`, `*sriov`) to repeat the same `network` block on several roles.

## Verify

The Milvus reports two conditions for SR-IOV:

| Condition | True when |
|---|---|
| `SRIOVNetworkAttached` | Multus reports an address on the configured interface for every SR-IOV pod |
| `SRIOVAddressAdvertised` | Every SR-IOV pod's session in the meta etcd carries its SR-IOV address |

```bash
$ kubectl get milvus -n demo milvus-gpu -o jsonpath='{range .status.conditions[*]}{.type}={.status} {.reason}{"\n"}{end}' | grep SRIOV
SRIOVNetworkAttached=True AllSRIOVNetworksAttached
SRIOVAddressAdvertised=True AllSessionsOnSRIOVNetwork
```

Check placement and GPU allocation:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=milvus-gpu -o wide
$ kubectl get pod -n demo milvus-gpu-querynode-0 -o jsonpath='{.spec.containers[?(@.name=="milvus")].resources}'
```

## Troubleshooting

- **Pod stuck in `Init`**: the init container found no IPv4 address on the interface within 30 seconds. Check the NAD name, the VF resource and the pod's `k8s.v1.cni.cncf.io/network-status` annotation.
- **`SRIOVAddressAdvertised` is False with reason `SessionsNotOnSRIOVNetwork`**: the message names the pods whose etcd session still shows another address. Restart them with a `Restart` ops request after fixing the network.
- **Admission error about `spec.db.gpu.supported`**: the version is not a GPU build. Use the `-gpu` entry.

## Cleaning up

```bash
$ kubectl delete milvus -n demo milvus-gpu
$ kubectl delete placementpolicy milvus-querynode-gpu-pools
```
