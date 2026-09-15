---
title: GPU and SR-IOV Networking for Milvus
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

# GPU and SR-IOV Networking for Milvus

KubeDB Milvus supports two related, independently useful capabilities:

- **GPU device scheduling** (`spec.gpu`), for Milvus's GPU-accelerated index types: `GPU_CAGRA`, `GPU_IVF_FLAT`, `GPU_IVF_PQ`, `GPU_BRUTE_FORCE`. Works the same way in `Standalone` and `Distributed` mode.
- **SR-IOV networking via Multus CNI** (`spec.network.sriov`), for nodes where a fast NIC is exposed to pods only through SR-IOV virtual functions, not the default cluster network. This is `Distributed`-only in its full effect, and has a topology requirement described below that is easy to get wrong.

## Before You Begin

- You should be familiar with the [Milvus](/docs/guides/milvus/concepts/milvus.md) concept, in particular `spec.gpu`/`spec.network` and `spec.topology`.
- Complete the dependency setup from [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md).
- This guide covers what KubeDB configures on your behalf. It does **not** cover installing GPU or SR-IOV cluster infrastructure — that is a cluster-admin task, done once per cluster, independent of any particular `Milvus` object:
  - the [NVIDIA GPU Operator](https://github.com/NVIDIA/gpu-operator) (or an equivalent device plugin) advertising `nvidia.com/gpu` on your GPU nodes;
  - the [SR-IOV Network Device Plugin](https://github.com/k8snetworkplumbingwg/sriov-network-device-plugin) advertising the SR-IOV virtual function as an extended resource (e.g. `intel.com/sriov_net_A`);
  - [Multus CNI](https://github.com/k8snetworkplumbingwg/multus-cni) installed as the cluster's meta-plugin;
  - a `NetworkAttachmentDefinition` naming the SR-IOV network, **with an IPAM plugin configured** (e.g. [whereabouts](https://github.com/k8snetworkplumbingwg/whereabouts)) — without one, the attached interface never gets an address at all.

## GPU: `spec.gpu`

`spec.gpu` requests a GPU device on the `milvus` container and the accompanying node-scheduling hints:

```yaml
gpu:
  resourceName: nvidia.com/gpu   # optional; this is the default. Override for vGPU/MIG resource names.
  count: 1                       # optional; defaults to 1.
  nodeSelector:
    nvidia.com/gpu.present: "true"
  tolerations:
  - key: nvidia.com/gpu
    operator: Exists
    effect: NoSchedule
```

- In **Standalone** mode, set this at the top level: `spec.gpu`.
- In **Distributed** mode, set this per role: `spec.topology.distributed.<role>.gpu`. `queryNode` (loads segments into GPU memory and runs search) and `dataNode` (builds GPU-accelerated indexes) are the roles that use it; the other three roles (`mixCoord`, `streamingNode`, `proxy`) don't need it.

### The `MilvusVersion` must declare GPU support

GPU support is a property of the Milvus **image** — whether it was built with GPU-accelerated `knowhere`/CUDA — not something every `MilvusVersion` in the catalog has. The referenced `MilvusVersion` must set:

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: MilvusVersion
spec:
  db:
    gpu:
      supported: true
```

If it doesn't, the admission webhook **rejects** the `Milvus` object outright — whether the GPU request came from the typed `spec.gpu` field or from a hand-written `nvidia.com/gpu` resource request/limit in `spec.podTemplate`. This also runs on updates, so changing `spec.version` to a `MilvusVersion` without GPU support, while a role still requests one, is rejected too.

> Check with your KubeDB distribution for which catalog `MilvusVersion` entries (if any) currently declare `gpu.supported: true` — a plain, non-GPU image cannot serve a GPU request no matter what the `Milvus` object says.

## SR-IOV networking: `spec.network.sriov`

```yaml
network:
  sriov:
    attachmentRef: milvus-sriov-net       # a cluster-admin-authored NetworkAttachmentDefinition, same namespace
    resourceName: intel.com/sriov_net_A   # must match the SR-IOV device plugin's advertised resource
    rdmaResourceName: rdma/rdma_shared_device_a   # optional; only if you specifically need GPUDirect RDMA
    interface: net1                        # optional; defaults to net1 (Multus's own convention)
    gds:                                    # optional second SR-IOV attachment, for GPU Direct Storage
      attachmentRef: milvus-gds-net
      resourceName: intel.com/sriov_net_B
      interface: net2                       # optional; defaults to net2
```

Setting this on a **Distributed** role does three things:

1. Stamps the `k8s.v1.cni.cncf.io/networks` annotation on the pod, so Multus attaches the named `NetworkAttachmentDefinition`(s) as secondary interface(s) — one entry for the primary attachment, plus a second for `gds` if set.
2. Requests the matching extended resource(s) (`resourceName`, `rdmaResourceName`, and `gds.resourceName` if set) so the pod is scheduled onto a node with an available virtual function for each.
3. Adds an init container that patches this component's advertised address to the *primary* attachment's IP — **this step only happens in Distributed mode.** `gds` is for data movement to object storage, not for the component's own advertised address, so it's never involved in step 3.

### Why step 3 matters: Milvus's own address auto-detection doesn't know about your SR-IOV interface

Every Milvus component registers its own `<component>.ip:port` in etcd, and that's the address every *other* component actually dials it on. Left unset, Milvus auto-detects this address by scanning every interface on the pod — with no way to prefer a specific one. Once a second (SR-IOV) interface is attached, this auto-detection keeps advertising the *primary* pod IP, same as before — the SR-IOV NIC sits attached but unused. KubeDB's init container fixes this two ways, belt-and-braces:

- it writes a small `user.yaml` config overlay Milvus already knows how to merge on top of `milvus.yaml` — no other config file is touched;
- it also writes `MILVUS_CONF_<SECTION>_IP` environment variables to a small script, sourced by the main container's entrypoint before `milvus` itself starts.

Both do the same job; running both is a defense-in-depth choice, since environment-variable override behavior has been observed to vary across Milvus 2.x releases. The discovery script retries the interface lookup for up to 30 seconds (in case Multus is still attaching the interface when the init container starts) before failing outright — the pod is left in `Init:Error` rather than silently starting up on the wrong network.

This is also why the mechanism is **Distributed-only**: `Standalone` is a single process with no other Milvus component to register an address *for*. Setting `spec.network.sriov` on a `Standalone` `Milvus` still stamps the annotation and resource request (useful if you want the pod scheduled with a fast NIC for some other reason), but does not run the advertise-IP init container.

### You almost always need this on all five Distributed roles, not just the GPU-bearing ones

`mixCoord`, `dataNode`, `queryNode`, `streamingNode`, and `proxy` all dial each other directly, by whichever address a component has advertised — never through a Kubernetes Service. If `queryNode`'s advertised address moves to an SR-IOV-only IP but `mixCoord`/`streamingNode`/`proxy` are left on the default network, they may no longer be able to reach it at all — particularly if the SR-IOV network is a physically isolated fabric with no route back to the primary cluster network, which is the common case. The admission webhook **warns** (not rejects, since it can't see your actual network topology) if `spec.network.sriov` is set on some but not all five Distributed roles.

**The recommended default topology**, when SR-IOV is in play at all: set `spec.network.sriov` on all five roles, and set `spec.gpu` on `queryNode`/`dataNode` only.

## Heterogeneous GPU pools: `groups`

If one role needs more than one GPU/network class at once — e.g. `queryNode` split across a ConnectX-6 pool and a ConnectX-7 pool, each with its own `NetworkAttachmentDefinition`, node selector, and resource name — use `groups` instead of the role-level fields:

```yaml
querynode:
  groups:
  - name: cx6
    replicas: 2
    gpu:
      nodeSelector:
        nic: connectx-6
    network:
      sriov:
        attachmentRef: milvus-sriov-cx6
        resourceName: mellanox.com/cx6_vf
  - name: cx7
    replicas: 2
    gpu:
      nodeSelector:
        nic: connectx-7
    network:
      sriov:
        attachmentRef: milvus-sriov-cx7
        resourceName: mellanox.com/cx7_vf
```

This creates two PetSets, `<db>-querynode-cx6` and `<db>-querynode-cx7`, each independently scheduled and reconciled — instead of the usual single `<db>-querynode` PetSet. `groups` is empty by default; every example elsewhere in this guide, without `groups` set, is unaffected.

To scale one group without touching the rest, target it directly in a `MilvusOpsRequest` instead of the whole-role fields:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: milvus-gpu-cluster-scale-cx7
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: milvus-gpu-cluster
  horizontalScaling:
    topology:
      groups:
      - nodeType: querynode
        group: cx7
        replicas: 4
```

`verticalScaling.groups` works the same way, with `resources` in place of `replicas`. A `streamingnode` group may also carry its own `storageType`/`storage` (see [Milvus CRD](/docs/guides/milvus/concepts/milvus.md)), falling back to the role-level default when unset.

## Full example: Distributed Milvus with GPU + SR-IOV

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-gpu-cluster
  namespace: demo
spec:
  version: "2.6.11"        # must reference a MilvusVersion with spec.db.gpu.supported: true
  objectStorage:
    configSecret:
      name: milvus-storage-config
  topology:
    mode: Distributed
    distributed:
      mixcoord:
        replicas: 2
        network:
          sriov:
            attachmentRef: milvus-sriov-net
            resourceName: intel.com/sriov_net_A
      datanode:
        replicas: 2
        gpu:
          count: 1
        network:
          sriov:
            attachmentRef: milvus-sriov-net
            resourceName: intel.com/sriov_net_A
      proxy:
        replicas: 2
        network:
          sriov:
            attachmentRef: milvus-sriov-net
            resourceName: intel.com/sriov_net_A
      querynode:
        replicas: 2
        gpu:
          count: 1
          nodeSelector:
            nvidia.com/gpu.present: "true"
          tolerations:
          - key: nvidia.com/gpu
            operator: Exists
            effect: NoSchedule
        network:
          sriov:
            attachmentRef: milvus-sriov-net
            resourceName: intel.com/sriov_net_A
      streamingnode:
        replicas: 3
        storageType: Durable
        storage:
          accessModes:
            - ReadWriteOnce
          storageClassName: local-path
          resources:
            requests:
              storage: 10Gi
        network:
          sriov:
            attachmentRef: milvus-sriov-net
            resourceName: intel.com/sriov_net_A
  configuration:
    inline: |
      gpu:
        initMemSize: 2048
        maxMemSize: 4096
  deletionPolicy: WipeOut
```

Notes on this example:

- `attachmentRef: milvus-sriov-net` refers to a `NetworkAttachmentDefinition` your cluster admin has already created in the `demo` namespace, with an IPAM plugin configured. KubeDB does not create or validate this object.
- `spec.configuration.inline`'s `gpu.initMemSize`/`gpu.maxMemSize` size Milvus's own GPU memory pool — a Milvus config concern, independent of the Kubernetes-level GPU/SR-IOV scheduling above.
- Only `queryNode` and `dataNode` carry `gpu`; all five roles carry `network.sriov`.

## Checking what KubeDB did

```bash
$ kubectl get milvus -n demo milvus-gpu-cluster -o yaml
$ kubectl get pods -n demo -l app.kubernetes.io/instance=milvus-gpu-cluster -o wide
$ kubectl describe pod -n demo <querynode-pod> # confirm the nvidia.com/gpu and SR-IOV resource requests, and the k8s.v1.cni.cncf.io/networks annotation
$ kubectl get pod -n demo <querynode-pod> -o jsonpath='{.metadata.annotations.k8s\.v1\.cni\.cncf\.io/network-status}'
$ kubectl exec -n demo <querynode-pod> -c milvus -- cat /milvus/configs/user.yaml
$ kubectl exec -n demo <querynode-pod> -c milvus -- cat /milvus/sriov-env.sh
$ kubectl get milvus -n demo milvus-gpu-cluster -o jsonpath='{.status.conditions[?(@.type=="SRIOVNetworkAttached")]}'
```

The `user.yaml`/`sriov-env.sh` commands should show the patched `queryNode.ip` / `MILVUS_CONF_QUERYNODE_IP` (or the equivalent key for whichever role you're checking), matching the pod's SR-IOV secondary IP from the `network-status` output above. The `SRIOVNetworkAttached` condition is a cluster-wide summary of the same check, computed by the operator itself: it reads every SR-IOV-enabled pod's own `network-status` annotation and reports `False`, with the offending pods named in its message, if any pod's expected interface never showed up with an IP. This is observational only — it never blocks or delays `Milvus.status.phase`.

## Known limitations

- The init container's IP-discovery approach and its base image are a reasonable default, not a fully validated, one-size-fits-all script — if your cluster's SR-IOV/IPAM setup behaves differently, this may need adjustment. Please report issues.
- KubeDB validates whether the referenced `MilvusVersion` has GPU support at all; it cannot see, and does not validate, per-collection index-type choices (`GPU_CAGRA` vs. `GPU_IVF_FLAT` vs. a plain CPU index) — those are chosen by the Milvus SDK client at collection-creation time.
- KubeDB does not verify that the referenced `NetworkAttachmentDefinition` exists, or that the SR-IOV/GPU device plugins are installed — a missing prerequisite surfaces as a normal pod-scheduling failure.
- `SRIOVNetworkAttached` confirms the CNI-level attachment (Multus's own `network-status`), not that Milvus's *internally registered* etcd address matches it — those should agree, since the same IP feeds both, but the condition doesn't independently query Milvus's etcd session state to prove it.
- If a pod is restarted and lands on a different SR-IOV IP, the init container picks up the new address immediately (it reruns fresh on every pod start) — but a *peer* Milvus process that already cached the old address may take a moment to notice, independent of anything Kubernetes or KubeDB can observe or control.

## Related Concepts

- [Milvus CRD](/docs/guides/milvus/concepts/milvus.md)
- [Monitoring Milvus](/docs/guides/milvus/monitoring/using-prometheus-operator.md)
