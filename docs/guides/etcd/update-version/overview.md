---
title: Updating Etcd Overview
menu:
  docs_{{ .version }}:
    identifier: etcd-update-version-overview
    name: Overview
    parent: etcd-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Overview of Etcd Version Update

This guide gives an overview of how the KubeDB Ops-manager operator updates the version of an `Etcd`
cluster.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdVersion](/docs/guides/etcd/concepts/etcdversion.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)

## The Upgrade Path Is Validated — Read This First

Unlike most KubeDB databases, an etcd version update is **not** a free jump between any two catalog
entries. etcd only supports a narrow upgrade path, and KubeDB enforces it before touching anything:

| Rule                        | Why                                                                                                                                        |
|-----------------------------|--------------------------------------------------------------------------------------------------------------------------------------------|
| **Same major version only** | A major version change is not a rolling upgrade at all.                                                                                     |
| **No downgrades**           | An etcd downgrade needs `etcdutl downgrade enable` plus a storage-schema migration, which is out of scope for this ops type.                 |
| **At most one minor step**  | etcd only guarantees a rolling upgrade across a single minor version (`3.4 → 3.5 → 3.6`). A member of version *N* joining a cluster whose cluster version is *N-2* refuses to start, and there is no supported in-place path that skips a minor. |
| **Patch releases are free** | Any patch move inside a minor (`3.6.1 → 3.6.4`) is always allowed.                                                                           |
| **Target must not be deprecated** | The `EtcdVersion` object being targeted must not have `spec.deprecated: true`.                                                        |

So `3.5.21 → 3.6.4` is fine, and so is `3.6.1 → 3.6.4`. `3.5.21 → 3.7.x` is rejected with a message
telling you to go through `3.6.x` first, and `3.6.4 → 3.5.21` is rejected outright.

These checks run in the Ops-manager operator when the request starts executing, **not** in the
admission webhook — the webhook only requires that `spec.updateVersion.targetVersion` is non-empty.
An invalid path therefore produces an `EtcdOpsRequest` that is accepted and then fails, with the
reason in its events, rather than a rejected `kubectl apply`.

## How the Update Version Process Works

The updating process consists of the following steps:

1. At first, a user creates an `Etcd` Custom Resource (CR).

2. The `KubeDB` Provisioner operator watches the `Etcd` CR and creates a `PetSet` and the other
   Kubernetes resources.

3. In order to update the version, the user creates an `EtcdOpsRequest` CR of type `UpdateVersion`
   with the desired `spec.updateVersion.targetVersion`.

4. The `KubeDB` Ops-manager operator watches the `EtcdOpsRequest` CR. It reads both the current and
   the target `EtcdVersion` objects from the catalog, rejects a deprecated target, and validates the
   upgrade path described above.

5. It then pauses the referenced `Etcd` object, so the Provisioner operator does not reconcile it
   during the update.

6. It patches the `PetSet` pod template with the target version's `etcd` container image, resolved to
   an immutable digest, so every pod recreated from that point on comes up on the new version. (It
   also re-images an init container named `etcd-init` if one is present, but the provisioner never
   creates one, so that only applies if you added it yourself.)

7. It persists `spec.version` on the `Etcd` object, so the Provisioner operator re-renders the same
   template once the database is resumed.

8. It rolls the members onto the new image, **one at a time and leader-last** (see below).

9. Finally, it resumes the `Etcd` object and marks the `EtcdOpsRequest` `Successful`.

## The Rolling Restart Is Leader-Aware

Step 8 is the same restart machinery used by the `Restart` and `VerticalScaling` (mode `Restart`) ops
types, and it is specific to etcd in one important way:

- The operator resolves the current Raft leader through etcd's `Status()` RPC.
- Every **follower** is evicted first, one at a time. After each eviction it waits for the pod to be
  `Ready` **and** for the cluster to report a healthy quorum (`healthy members >= N/2+1`) before
  moving on. Without that gate, the second eviction in a 3-member cluster would take quorum down.
- The **leader is restarted last**, and only after its leadership has been transferred to another
  voting member with `MoveLeader`. A deliberate transfer is near instantaneous; evicting the leader
  outright triggers an election, during which the whole cluster is unavailable for writes for an
  election timeout.
- A single-member cluster has no one to hand leadership to, so it is simply restarted — and it *is*
  unavailable while it restarts.

Because members are rolled one at a time, the cluster runs in a **mixed-version** state for the
duration of the update. That is explicitly supported by etcd for exactly one minor step — which is
precisely the restriction the upgrade-path validation enforces.

In the [next](/docs/guides/etcd/update-version/update-version.md) doc, we are going to show a
step-by-step guide on updating the version of an Etcd cluster using the `EtcdOpsRequest` CRD.
