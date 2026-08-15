---
title: Etcd Monitoring Overview
description: Etcd Monitoring Overview
menu:
  docs_{{ .version }}:
    identifier: etcd-monitoring-overview
    name: Overview
    parent: etcd-monitoring-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Etcd with KubeDB

KubeDB has native support for monitoring via [Prometheus](https://prometheus.io/). You can use the builtin [Prometheus](https://github.com/prometheus/prometheus) scraper or the [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) to monitor a KubeDB-managed etcd cluster.

## etcd is its own exporter

This is the one thing to understand before you read any further, because it differs from almost every other database KubeDB manages.

For PostgreSQL, MySQL, MongoDB and friends, KubeDB injects a **Prometheus exporter sidecar** into the database pod. The sidecar connects to the database, translates its statistics into Prometheus text format, and serves them on its own port.

**etcd needs none of that.** etcd is written in Go and instrumented with the Prometheus client library directly, so it already speaks Prometheus natively. There is **no exporter sidecar container** in an etcd member pod — the pod contains a single operator-managed `etcd` container and nothing else. When you enable `spec.monitor`, KubeDB does not add anything to the pod at all; it only creates the Kubernetes objects that let a Prometheus server find the metrics etcd is already serving.

Practically, that means:

- `kubectl get pod` shows `1/1`, not `2/2`, on a monitored etcd member.
- The `spec.monitor.prometheus.exporter` sub-fields that configure a sidecar — `args`, `env`, `resources`, `securityContext` — have **no effect** for etcd, because there is no container for them to configure. Only `spec.monitor.prometheus.exporter.port` is meaningful: it is the port number that the stats Service publishes.
- There is no exporter container to size, so a `VerticalScaling` `EtcdOpsRequest` has nothing to do with `spec.verticalScaling.exporter` either. That field exists in the API for structural parity with the other databases.

## Where the metrics come from

Every etcd member is started with a dedicated metrics listener:

```
--listen-metrics-urls=http://0.0.0.0:2381
```

This is a separate socket from the client API (`2379`) and the Raft peer channel (`2380`). It serves two paths:

| Path       | Purpose                                                                                                   |
|------------|-----------------------------------------------------------------------------------------------------------|
| `/metrics` | The Prometheus exposition endpoint. This is what gets scraped.                                             |
| `/health`  | etcd's own health handler. KubeDB points the pod's readiness probe at it.                                  |

The container port is declared as `metrics` (`2381`) alongside `client` (`2379`) and `peer` (`2380`).

Keeping metrics on their own listener is deliberate. etcd also serves `/metrics` on the client port, but once TLS is enabled KubeDB renders `--client-cert-auth=true`, and a client that cannot present a certificate — kubelet running a readiness probe, or a Prometheus server that has not been given one — would be rejected there. The dedicated listener stays plain HTTP, so probes and scraping keep working unchanged after you turn TLS on. Note the consequence: **enabling `spec.tls` does not encrypt the metrics endpoint.** If your Prometheus server and your etcd cluster are in different trust domains, use a NetworkPolicy or a service mesh to control access to port `2381`.

> KubeDB deliberately does **not** issue a `metrics-exporter` certificate for etcd. The alias is accepted by the `Etcd` schema, and defaulting even stamps an entry for it into `spec.tls.certificates`, but nothing ever creates or mounts that Secret — etcd's `--listen-metrics-urls` has no TLS flags of its own, so the listener is plain HTTP by design. The stats Service and the generated `ServiceMonitor` are therefore configured with `scheme: http`.

## The stats Service

When `spec.monitor.agent` is a Prometheus agent, KubeDB creates a second Service next to the regular client Service, named `<etcd-crd-name>-stats`:

- It selects the same pods as the client Service.
- It publishes a single port named `metrics`, whose port number is `spec.monitor.prometheus.exporter.port` (default `2381`), targeting etcd's own metrics port on the pod.
- It is labelled with the standard KubeDB offshoot labels plus `kubedb.com/role: stats`, which is what a `ServiceMonitor` selects on.

Everything else depends on which agent you chose.

## Configure Monitoring

| Field                                              | Type       | Uses                                                                                                                                    |
|----------------------------------------------------|------------|------------------------------------------------------------------------------------------------------------------------------------------|
| `spec.monitor.agent`                               | `Required` | The monitoring agent. Either `prometheus.io/builtin` or `prometheus.io/operator`.                                                          |
| `spec.monitor.prometheus.exporter.port`            | `Optional` | Port number published by the stats Service. Defaults to `2381`. (For etcd this is only the Service port — there is no exporter process.)  |
| `spec.monitor.prometheus.serviceMonitor.labels`    | `Optional` | Labels put on the generated `ServiceMonitor`, so your `Prometheus` object's `serviceMonitorSelector` matches it.                           |
| `spec.monitor.prometheus.serviceMonitor.interval`  | `Optional` | Scrape interval on the generated `ServiceMonitor`.                                                                                        |

`spec.monitor.prometheus.exporter.args`, `.env`, `.resources` and `.securityContext` are accepted by the schema (the type is shared across all KubeDB databases) but have no effect for etcd — see above.

### `prometheus.io/builtin`

KubeDB annotates the stats Service so that a plain Prometheus server doing Kubernetes service-endpoint discovery picks it up:

```
prometheus.io/scrape: "true"
prometheus.io/scheme: http
prometheus.io/path: /metrics
prometheus.io/port: "2381"
monitoring.appscode.com/agent: prometheus.io/builtin
```

No CRD is involved. See [Monitor Etcd using builtin Prometheus](/docs/guides/etcd/monitoring/using-builtin-prometheus.md).

### `prometheus.io/operator`

KubeDB additionally creates a `ServiceMonitor` named `<etcd-crd-name>-stats`, **in the same namespace as the `Etcd` object**, selecting the stats Service. See [Monitor Etcd using Prometheus operator](/docs/guides/etcd/monitoring/using-prometheus-operator.md).

## Sample Configuration

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-monitoring
  namespace: demo
spec:
  version: 3.6.4
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  deletionPolicy: WipeOut
```

## What you get to look at

Because the metrics come from etcd itself rather than from a third-party exporter, you get upstream etcd's own metric names — the same ones the etcd documentation and community dashboards use. A few that are worth an alert:

| Metric                                    | Why it matters                                                                                                                              |
|-------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------|
| `etcd_server_has_leader`                  | `1` if this member sees a leader, `0` if it does not. A `0` anywhere means that member cannot serve writes; `0` everywhere means quorum is lost. |
| `etcd_server_leader_changes_seen_total`   | Cumulative leader elections. A rising counter means instability — usually disk or network latency.                                             |
| `etcd_mvcc_db_total_size_in_bytes`        | Current backend database size. Compare it against the quota below.                                                                            |
| `etcd_server_quota_backend_bytes`         | The configured `--quota-backend-bytes`. When the size metric approaches this, etcd is heading for a `NOSPACE` alarm and read-only mode.        |
| `etcd_disk_wal_fsync_duration_seconds`    | WAL fsync latency histogram. etcd is unusually sensitive to slow disks; this is the first thing to check when elections start happening.       |
| `etcd_disk_backend_commit_duration_seconds` | Backend commit latency histogram.                                                                                                           |
| `etcd_network_peer_round_trip_time_seconds` | Raft peer RTT histogram, labelled `To`. Useful for spotting one bad member.                                                                 |

The backend-size-versus-quota pair is the one most worth wiring an alert to, because hitting the quota puts the cluster into a read-only alarm state that needs manual intervention. See [custom configuration](/docs/guides/etcd/custom-configuration/using-config.md) for how to set the quota and the auto-compaction policy, and the `Compact`/`Defragment` [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md) types for reclaiming space.

## Next Steps

- Monitor your etcd cluster with KubeDB using [builtin Prometheus](/docs/guides/etcd/monitoring/using-builtin-prometheus.md).
- Monitor your etcd cluster with KubeDB using the [Prometheus operator](/docs/guides/etcd/monitoring/using-prometheus-operator.md).
- Detail concepts of [Etcd object](/docs/guides/etcd/concepts/etcd.md).
