---
title: Monitoring PostgreSQL Remote Replicas
menu:
  docs_{{ .version }}:
    identifier: pg-remote-replica-monitoring
    name: Monitoring
    parent: pg-remote-replica
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring PostgreSQL Remote Replicas

A remote replica is disaster-recovery infrastructure: the questions its monitoring must
answer are *is the replica streaming*, *can it see its source*, *how much data is at risk
if the source data center is lost right now* (RPO), and *did self-healing fire*. This
guide wires those up on the **replica-side cluster** and installs a Grafana dashboard
built around exactly those questions.

## What serves the metrics

Every remote replica pod runs a `pg-coordinator` sidecar in remote-replica mode. Besides
recovery and role-label management, it serves DR metrics on the `raft-metrics` port
(23790), fed by its own monitor and lag-monitor loops — Prometheus scrapes never touch
the source database:

| Metric | Meaning |
|---|---|
| `pg_coordinator_remote_replica_lag_bytes` | WAL bytes the source has written beyond what this pod has replayed — the data at risk (RPO). Absent until the first measurement |
| `pg_coordinator_remote_replica_streaming` | 1 when the source confirms this pod in `pg_stat_replication`; mirrors the pod's `standby` role label |
| `pg_coordinator_remote_replica_source_reachable` | 1 when the last source query succeeded — a DR replica that cannot see its source is not protecting anything |
| `pg_coordinator_remote_replica_last_lag_check_timestamp_seconds` | when the lag was last measured; the monitor backs off to 300s while in sync, so up to ~5 min of age is normal |
| `pg_coordinator_remote_replica_recovery_total{action,result}` | pg_rewind / pg_basebackup self-healing attempts; all four series exported from start, so any step above 0 is a real event |

The standard `postgres_exporter` (port 56790, added by `spec.monitor`) contributes
`pg_replication_is_replica`, `pg_replication_lag_seconds`, `pg_stat_activity_count`, etc.

## Prerequisites

On the replica-side cluster:

- [kube-prometheus-stack](https://artifacthub.io/packages/helm/prometheus-community/kube-prometheus-stack).
  The dashboard below is developed against **Grafana 7.5.x** (`--set grafana.image.tag=7.5.5`).
- [Panopticon](https://appscode.com/products/panopticon/) with your KubeDB license — it
  exports `kubedb_com_postgres_info`, which drives the dashboard's `app` variable.
- The `kubedb-metrics` chart (MetricsConfigurations for Panopticon).

## Step 1: enable monitoring on the remote replica

Add `spec.monitor` to the remote replica Postgres (the `release: prometheus` label must
match your kube-prometheus-stack release name — Prometheus only selects ServiceMonitors
carrying it):

```yaml
spec:
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 30s
```

```bash
kubectl patch pg pg-london -n demo --type merge -p '{
  "spec": {"monitor": {"agent": "prometheus.io/operator",
    "prometheus": {"serviceMonitor": {"labels": {"release": "prometheus"}, "interval": "30s"}}}}}'
```

This adds the exporter container to the pod template and creates the `<db>-stats`
Service + ServiceMonitor exposing **both** metric ports (`metrics`/56790 and
`raft-metrics`/23790).

> **The pod must be restarted once**: remote replica PetSets use the `OnDelete` update
> strategy, so the exporter container only appears after a pod delete. Streaming resumes
> automatically after the restart.

```bash
kubectl delete pod pg-london-0 -n demo
kubectl wait pg pg-london -n demo --for=jsonpath='{.status.phase}'=Ready --timeout=300s
kubectl get pod pg-london-0 -n demo -o jsonpath='{range .spec.containers[*]}{.name} {end}'
# postgres pg-coordinator exporter
```

## Step 2: if the cluster runs NetworkPolicies, allow the scrape

KubeDB can deploy NetworkPolicies that restrict ingress to database pods to their own
namespace (plus the operator). Prometheus lives in another namespace, so **both scrape
targets stay down** until you allow it. The symptom: `up{job="<db>-stats"} == 0` while
`wget 127.0.0.1:56790/metrics` inside the pod works fine.

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-prometheus-scrape
  namespace: demo
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/component: database
      app.kubernetes.io/managed-by: kubedb.com
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: monitoring
    ports:
    - port: 56790
      protocol: TCP
    - port: 23790
      protocol: TCP
```

Skip this step if `kubectl get networkpolicy -n demo` shows nothing — whether KubeDB
creates these policies is an install-time choice (`networkPolicy.enabled` in the chart
values). On a cluster **without** them, do **not** apply this policy: it would become the
only policy selecting the database pods and deny all other ingress — operator health
checks fail and the CR sticks in `Provisioning`.

## Step 3: verify the scrape

```bash
PROM=prometheus-prometheus-kube-prometheus-prometheus-0
kubectl exec -n monitoring $PROM -c prometheus -- promtool query instant \
  http://localhost:9090 'up{namespace="demo",pod=~"pg-london-.*"}'
# both endpoints (metrics and raft-metrics) must be 1

kubectl exec -n monitoring $PROM -c prometheus -- promtool query instant \
  http://localhost:9090 'pg_coordinator_remote_replica_streaming{namespace="demo"}'
# 1 per streaming pod
```

## Step 4: install the dashboard

The **KubeDB / Postgres / Remote Replica** dashboard
([opnpulse/dashboards, postgres folder](https://github.com/opnpulse/dashboards/tree/master/postgres))
has three rows: *DR Protection Status* (streaming, source reachable, RPO in bytes, lag
data age, recoveries in 24h), *Replication Lag* (byte lag from the coordinator; apply
lag in seconds from the exporter — the latter also grows while the source is idle, read
them together), and *Self-Healing & Replica Health*.

Provision it as a ConfigMap so it survives Grafana restarts (kube-prometheus-stack's
Grafana has no persistence — dashboards imported through the UI or API are lost on pod
restart; the sidecar re-provisions labeled ConfigMaps):

```bash
kubectl create configmap pg-remote-replica-dashboard -n monitoring \
  --from-file=postgres_remote_replica_dashboard.json
kubectl label configmap pg-remote-replica-dashboard -n monitoring grafana_dashboard=1
```

Then open Grafana and select your namespace and database in the `namespace` / `app`
variables:

```bash
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
```

## Reading the dashboard

| Symptom | Likely meaning |
|---|---|
| Streaming red, Source Reachable green | a replica pod is down, diverged, or mid-recovery; watch Recovery Actions |
| Source Reachable red | the clusters are partitioned or the source is down — severity-1 even though the replica looks healthy locally |
| RPO climbing, Streaming green | WAL arrives but replay cannot keep up (or replay is paused) |
| Apply Lag climbing, byte lag 0 | the source is idle; not an incident |
| Recoveries ≥ 1 | self-healing fired (pg_rewind or re-seed) — read the coordinator logs of the affected pod |

## Next Steps

- [Remote Replica overview](/docs/guides/postgres/remote-replica/remotereplica.md)
- [Cross-Cluster DR with Bidirectional Failover](/docs/guides/postgres/remote-replica/advanced-setup.md)
- [Migration from Self-Managed PostgreSQL](/docs/guides/postgres/remote-replica/migration.md)
