---
title: Reconfiguring Etcd
menu:
  docs_{{ .version }}:
    identifier: etcd-reconfigure-overview
    name: Overview
    parent: etcd-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring Etcd

This guide gives an overview of how the KubeDB Ops-manager operator reconfigures a running `Etcd`
cluster.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)

## What "configuration" means for Etcd

Most KubeDB databases are configured through a config file that is mounted into the pod — a
`configSecret` whose contents are merged with the operator's defaults. **Etcd is different.**

etcd's `--config-file` is *mutually exclusive* with individual command line flags: if you hand etcd
a config file, that file has to own every single setting, including the bootstrap flags KubeDB
computes per member (`--name`, `--initial-advertise-peer-urls`, `--listen-peer-urls`,
`--listen-client-urls`, `--advertise-client-urls`, the TLS file paths, and so on). Silently taking
over those would break membership management.

So KubeDB exposes a small, typed set of tuning knobs instead, under `spec.configuration.tuning`,
and renders them straight onto the etcd command line:

| Field                     | etcd flag                     | Type     | Meaning                                                                               |
|---------------------------|-------------------------------|----------|---------------------------------------------------------------------------------------|
| `quotaBackendBytes`       | `--quota-backend-bytes`       | int64    | Maximum size of the backend database. etcd raises a `NOSPACE` alarm and goes read-only once the backend grows past it. |
| `autoCompactionMode`      | `--auto-compaction-mode`      | string   | Either `periodic` or `revision`; selects how `autoCompactionRetention` is interpreted. |
| `autoCompactionRetention` | `--auto-compaction-retention` | string   | A duration (e.g. `1h`) when the mode is `periodic`, a revision count (e.g. `1000`) when the mode is `revision`. |
| `snapshotCount`           | `--snapshot-count`            | uint64   | Number of committed transactions that trigger a snapshot to disk.                      |

Two consequences follow from this design, and they are the two things worth remembering about
reconfiguring etcd:

1. **`spec.configuration.applyConfig` is not supported.** `applyConfig` merges free-form key/value
   pairs into a rendered config file, and there is no such file here — accepting it would silently
   drop your settings. An `EtcdOpsRequest` that sets it is rejected with
   `spec.configuration.applyConfig is not supported for etcd; use spec.configuration.tuning`.
2. **A `Reconfigure` always requires a restart.** The knobs are process command line flags, and
   etcd has no live-reload path for them (there is no equivalent of PostgreSQL's
   `pg_reload_conf()`). The new values only take effect when the member process is started again.

## How the Reconfiguring Etcd Process Works

The reconfiguration process consists of the following steps:

1. At first, a user creates an `Etcd` Custom Resource (CR).

2. The `KubeDB` Provisioner operator watches the `Etcd` CR.

3. When the operator finds an `Etcd` CR, it creates the required `PetSet` and the related resources
   like Secrets, Services, and so on.

4. Then, in order to reconfigure the cluster, the user creates an `EtcdOpsRequest` CR of type
   `Reconfigure`, carrying the desired `spec.configuration.tuning` values.

5. The `KubeDB` Ops-manager operator watches the `EtcdOpsRequest` CR.

6. When it finds an `EtcdOpsRequest` CR, it pauses the `Etcd` object referred to by the request, so
   that the Provisioner operator does not perform any operation on it during reconfiguration.

7. The Ops-manager operator folds the requested knobs into the `Etcd` object's
   `spec.configuration.tuning`. Only the knobs the request actually names are touched, so a partial
   reconfigure never silently resets the others.

8. It then re-renders the `PetSet` from the patched spec, so the etcd container args carry the new
   flags. This is tracked by the `UpdateEtcdPetSet` condition.

9. Finally it rolls the member pods so that they come up with the new flags. This reuses exactly
   the same leader-aware sequence as the [Restart](/docs/guides/etcd/restart/restart.md) ops
   request: followers first, one at a time, quorum verified between each, and the Raft leader last,
   after its leadership has been transferred away.

10. After the members have restarted successfully, the Ops-manager operator resumes the `Etcd`
    object so that the Provisioner operator resumes its usual operations.

> Setting the tuning knobs at **creation time** — that is, on the `Etcd` object itself rather than
> through an ops request — is covered by the Custom Configuration guide and by
> [`spec.configuration`](/docs/guides/etcd/concepts/etcd.md#specconfiguration) in the Etcd concept
> doc. This guide is only about changing them on an already running cluster with an
> `EtcdOpsRequest`.

In the [next](/docs/guides/etcd/reconfigure/reconfigure.md) doc, we are going to show a step by step
guide on reconfiguring an Etcd cluster using the `EtcdOpsRequest` CRD.
