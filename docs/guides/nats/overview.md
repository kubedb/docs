---
title: NATS Overview
menu:
  docs_{{ .version }}:
    identifier: nats-overview
    name: Overview
    parent: nats-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# NATS with KubeDB

[NATS](https://nats.io) is a messaging system with core publish/subscribe and request/reply, and **JetStream**, a persistence layer with streams, consumers, key-value and object stores. KubeDB runs NATS as a managed database: you declare a `NATS` object and KubeDB creates the PetSet, Services, credentials, certificates and monitoring, and runs day-2 operations without dropping clients.

## What KubeDB creates

A `NATS` object is always **one NATS cluster**. All pods are identical `nats-server` processes, so there is a single PetSet and no per-role topology. Each pod runs these containers:

| Container | Purpose |
|---|---|
| `nats` | the server |
| `reloader` | signals the server to reload its configuration when the config Secret changes |
| `exporter` | `prometheus-nats-exporter`, only when monitoring is enabled |

| Port | Purpose | Exposed on |
|---|---|---|
| 4222 | clients | the client Service `<name>` |
| 6222 | cluster routes | the headless Service `<name>-pods` only |
| 8222 | HTTP monitor, probes | the headless Service; limited by a NetworkPolicy |
| 7777 | Prometheus exporter | the stats Service `<name>-stats` |
| 7222 | gateway (multi-DC) | the headless Service only |
| 7422 | leaf nodes | the headless Service and, when `leafNodes.accept` is true, the client Service |

## Custom resources

| Kind | API group | Purpose |
|---|---|---|
| `NATS` | `kubedb.com/v1alpha2` | the cluster: replicas, JetStream storage, auth, TLS, configuration, monitoring |
| `NATSVersion` | `catalog.kubedb.com/v1alpha1` | server, reloader and exporter images, upgrade constraints, lame-duck defaults |
| `NATSOpsRequest` | `ops.kubedb.com/v1alpha1` | day-2 operations |
| `NATSAutoscaler` | `autoscaling.kubedb.com/v1alpha1` | compute and storage autoscaling |

The resource plural of `NATS` is `natses` (a plural equal to the singular is not allowed). The short name is `nats`, so `kubectl get nats` works.

## JetStream and replica counts

- With JetStream enabled (the default) use `replicas: 1` or an odd number `>= 3`, because streams and the JetStream meta group are Raft groups.
- With `jetStream.enabled: false` the cluster holds no state, needs no PVC, and any `replicas >= 1` is valid.
- `jetStream.storageType: Ephemeral` uses an `emptyDir` and is allowed only with one replica.

## Guides

- [Quickstart](/docs/guides/nats/quickstart.md)
- [Clustering and configuration](/docs/guides/nats/clustering.md)
- [TLS](/docs/guides/nats/tls.md)
- [Authentication modes](/docs/guides/nats/auth-modes.md)
- [Monitoring](/docs/guides/nats/monitoring.md)
- [Multi-DC: gateways, leaf nodes and distributed mode](/docs/guides/nats/multi-dc.md)
- [Ops requests](/docs/guides/nats/ops-requests.md)
- [Backup and restore](/docs/guides/nats/backup.md)
