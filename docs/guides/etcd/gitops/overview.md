---
title: Etcd GitOps Overview
description: Etcd GitOps Overview
menu:
  docs_{{ .version }}:
    identifier: etcd-gitops-overview
    name: Overview
    parent: etcd-gitops
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# GitOps Overview for Etcd

This guide gives you an overview of how the KubeDB `gitops` operator works with `Etcd` clusters
through the `gitops.kubedb.com/v1alpha1` API.

## Before You Begin

- You should be familiar with the [Etcd](/docs/guides/etcd/concepts/etcd.md) and
  [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md) objects.

## Two objects, one spec

There are two `Etcd` kinds involved, and they are not the same object:

| Kind | API group | Role |
|---|---|---|
| `Etcd` | `gitops.kubedb.com/v1alpha1` | The **desired state** you commit to Git. This is the object your GitOps tool applies. |
| `Etcd` | `kubedb.com/v1alpha2` | The **live database**, created and owned by the gitops operator. |

The wrapper is a thin one: its `spec` is literally the same `EtcdSpec` type as the live `Etcd`
object, so anything you can write in a `kubedb.com/v1alpha2` `Etcd` you can write in the
`gitops.kubedb.com/v1alpha1` one, unchanged. The wrapper adds only a `status.gitops` section on top
of the normal database status.

Because the spec types are identical, the gitops operator runs your object through the same
defaulting and validation webhook the live `Etcd` object uses **before** it creates anything. An
invalid spec is rejected up front rather than producing a broken database.

## Workflow

<figure align="center">
  <img alt="GitOps Flow" src="/docs/images/gitops/gitops.jpg">
  <figcaption align="center">Fig: GitOps process</figcaption>
</figure>

1. **Define a GitOps `Etcd`**: write a custom resource of kind `Etcd` using the
   `gitops.kubedb.com/v1alpha1` API.
2. **Store it in Git**: push the manifest to your Git repository.
3. **Automated deployment**: a GitOps tool such as `ArgoCD` or `FluxCD` watches the repository and
   applies the manifest to your cluster.
4. **Create the database**: the gitops operator validates the spec and creates the corresponding
   `kubedb.com/v1alpha2` `Etcd` object, which the KubeDB provisioner then turns into a real cluster.
5. **Handle updates**: this is the important part — see below.

## Drift becomes an OpsRequest, not a mutation

When you change the committed manifest, the gitops operator does **not** patch the change straight
onto the live `Etcd` object. Mutating a running etcd cluster's spec in place is exactly how you lose
quorum. Instead the operator diffs the wrapper against the live database and, for each meaningful
difference, generates an `EtcdOpsRequest` — the same safe, ordered, quorum-aware machinery you would
use by hand.

| What you changed in Git | Generated `EtcdOpsRequest` type | Generated name |
|---|---|---|
| `spec.podTemplate.spec.containers[etcd].resources`, or `spec.monitor.prometheus.exporter.resources` | `VerticalScaling` | `<name>-verticalscaling-<suffix>` |
| `spec.replicas` | `HorizontalScaling` | `<name>-horizontalscaling-<suffix>` |
| `spec.storage.resources.requests.storage` | `VolumeExpansion` | `<name>-volumeexpansion-<suffix>` |
| `spec.storage.storageClassName` | `StorageMigration` | `<name>-storagemigration-<suffix>` |
| `spec.configuration.tuning` | `Reconfigure` | `<name>-reconfigure-<suffix>` |
| `spec.authSecret` | `RotateAuth` | `<name>-rotate-auth-<suffix>` |
| `spec.tls` | `ReconfigureTLS` | `<name>-reconfiguretls-<suffix>` |
| `spec.version` | `UpdateVersion` | `<name>-versionupdate-<suffix>` |
| `spec.monitor` or `spec.archiver` | `Restart` | `<name>-restart-<suffix>` |

The suffix is a random string, so a fresh OpsRequest is created for every change and the history
stays readable.

A handful of fields are not worth an OpsRequest and are patched onto the live object directly:
`spec.autoOps`, `spec.serviceTemplates`, `spec.halted`, `spec.healthChecker`, `spec.deletionPolicy`
and `spec.allowedSchemas`.

Some things you should know about the etcd-specific behaviour:

- **`spec.replicas` scaling goes through the membership state machine.** The generated
  `HorizontalScaling` OpsRequest is a thin wrapper: it patches the replica count and waits for the
  provisioner to converge. Scale-up adds the new ordinal as a raft *learner*, waits for it to catch
  up with the leader, then promotes it. Scale-down removes the highest-ordinal member from the etcd
  membership *before* the PetSet shrinks. Only one membership mutation happens per reconcile pass.
- **Shrinking a volume is rejected.** If the committed `spec.storage` request is smaller than the
  live one, the operator returns an error instead of generating a `VolumeExpansion`.
- **Custom configuration means `spec.configuration.tuning` only.** etcd's `--config-file` is
  mutually exclusive with the individual flags KubeDB must set for cluster bootstrap, so there is no
  free-form config file to reconcile. Removing the `tuning` block from Git generates a `Reconfigure`
  OpsRequest that resets etcd back to the operator defaults.
- **Autoscaler wins.** If an `EtcdAutoscaler` in the same namespace targets this database with a
  `spec.compute` section, the gitops operator will not generate `VerticalScaling` OpsRequests for it;
  same for `spec.storage` and `VolumeExpansion`. Otherwise the two controllers would fight over the
  same fields.
- **The exporter row is structural.** `spec.monitor.prometheus.exporter.resources` maps onto
  `EtcdVerticalScalingSpec.Exporter`, which exists for parity with other KubeDB databases. etcd
  serves its own Prometheus metrics natively and KubeDB deploys **no exporter sidecar container**, so
  there is nothing for those resources to apply to.

## Status

The wrapper's `status` inlines the live database's status (phase, conditions, and so on) and adds a
`status.gitops` section recording, per observed generation, which OpsRequests were generated and
where the change request stands — `Pending`, `InProgress`, `Current` or `Failed`.

## Next Steps

- Follow the step-by-step guide in
  [GitOps Etcd using the KubeDB GitOps Operator](/docs/guides/etcd/gitops/gitops.md).
- Detail concepts of the [Etcd](/docs/guides/etcd/concepts/etcd.md) object.
- Detail concepts of the [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md) object.
