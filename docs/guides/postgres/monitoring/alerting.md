---
title: PostgreSQL Alerting with Prometheus
description: Complete guide to setting up PostgreSQL alerts using postgres-alerts and kubedb-grafana-dashboards Helm charts
menu:
  docs_{{ .version }}:
    identifier: pg-monitoring-alerting
    name: Alerting
    parent: pg-monitoring-postgres
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PostgreSQL Alerting with Prometheus

This guide walks through installing Prometheus-based alerts for a KubeDB-managed PostgreSQL instance and explains how each alert works end-to-end.

## Architecture Overview

```
PostgreSQL Pod (pg-grafana-demo)
  └── postgres_exporter sidecar  ──scrape──► Prometheus (kube-prometheus-stack)
                                                  │
                                     PrometheusRule│(postgres-alerts chart)
                                                  │
                                              AlertManager ──► notifications
                                                  │
                                              Grafana
                                           (kubedb-grafana-dashboards)
```

- **KubeDB** deploys PostgreSQL with a built-in `postgres_exporter` sidecar on port `56790`.
- **ServiceMonitor** (`pg-grafana-demo-stats`) tells Prometheus to scrape the exporter every 10s.
- **PrometheusRule** (created by `postgres-alerts` chart) defines the alert conditions.
- **Prometheus Operator** evaluates rules and fires alerts to AlertManager.
- **AlertManager** routes fired alerts to configured receivers (email, Slack, PagerDuty, etc.).
- **Grafana** visualises the metrics using dashboards from `kubedb-grafana-dashboards`.

## Prerequisites

| Component | Details |
|-----------|---------|
| Cluster | k3s single-node (`bonusree`), kubeconfig at `/home/banusree/all_db.yaml` |
| KubeDB | `v2026.4.27` in `kubedb` namespace |
| PostgreSQL instance | `pg-grafana-demo` (version 13.13) in `demo` namespace |
| kube-prometheus-stack | `v86.2.3` in `monitoring` namespace |
| Prometheus | `prometheus-kube-prometheus-prometheus`, ruleSelector label: `release: prometheus` |
| AlertManager | `prometheus-kube-prometheus-alertmanager` |
| Grafana | `prometheus-grafana` (Grafana 13.0.2) with sidecar dashboard loading |

### Verify the PostgreSQL instance has monitoring enabled

```bash
kubectl --kubeconfig=/home/banusree/all_db.yaml -n demo get postgres pg-grafana-demo -o yaml | grep -A15 "monitor:"
```

Expected output:
```yaml
monitor:
  agent: prometheus.io/operator
  prometheus:
    exporter:
      port: 56790
    serviceMonitor:
      interval: 10s
      labels:
        release: prometheus
```

The `release: prometheus` label on the ServiceMonitor makes Prometheus discover and scrape this target.

---

## Step 1 — Install postgres-alerts

The `postgres-alerts` chart (from [opnpulse/alerts](https://github.com/opnpulse/alerts)) creates a `PrometheusRule` resource containing all PostgreSQL alert definitions.

### Why the `release: prometheus` label matters

The Prometheus instance is configured with:

```yaml
ruleSelector:
  matchLabels:
    release: prometheus
```

This means only `PrometheusRule` resources carrying the label `release: prometheus` are loaded. The chart default is `release: kube-prometheus-stack`, so we override it.

### Install command

```bash
helm upgrade -i postgres-alerts oci://ghcr.io/appscode-charts/postgres-alerts \
  -n demo \
  --create-namespace \
  --version=v2026.2.24 \
  --set metadata.release.name=pg-grafana-demo \
  --set metadata.release.namespace=demo \
  --set form.alert.labels.release=prometheus
```

| Flag | Value | Reason |
|------|-------|--------|
| `-n demo` | `demo` | Same namespace as the PostgreSQL instance |
| `metadata.release.name` | `pg-grafana-demo` | Scopes PromQL filters to this instance |
| `metadata.release.namespace` | `demo` | Scopes PromQL filters to this namespace |
| `form.alert.labels.release` | `prometheus` | Matches the Prometheus `ruleSelector` |

### Verify

```bash
kubectl --kubeconfig=/home/banusree/all_db.yaml -n demo get prometheusrule
# NAME              AGE
# postgres-alerts   <age>

kubectl --kubeconfig=/home/banusree/all_db.yaml -n demo get prometheusrule postgres-alerts \
  -o jsonpath='{.metadata.labels}'
# {"release":"prometheus", ...}
```

---

## Step 2 — Install kubedb-grafana-dashboards

The `kubedb-grafana-dashboards` chart (from [kubedb/installer](https://github.com/kubedb/installer)) creates `GrafanaDashboard` CRDs containing pre-built PostgreSQL dashboards. These are reconciled by the openviz Grafana operator into Grafana.

### Create a Grafana service account token

The dashboard import job authenticates to Grafana with a Bearer token.

```bash
# Port-forward Grafana locally
kubectl --kubeconfig=/home/banusree/all_db.yaml -n monitoring \
  port-forward svc/prometheus-grafana 3000:80 &

# Create a Grafana service account with Admin role
curl -s -X POST -H "Content-Type: application/json" \
  -u admin:<grafana-admin-password> \
  http://localhost:3000/api/serviceaccounts \
  -d '{"name":"kubedb-dashboards","role":"Admin"}'
# Note the returned "id"

# Create a token for the service account
curl -s -X POST -H "Content-Type: application/json" \
  -u admin:<grafana-admin-password> \
  http://localhost:3000/api/serviceaccounts/<id>/tokens \
  -d '{"name":"kubedb-token","secondsToLive":0}'
# Note the returned "key"

kill %1  # stop port-forward
```

### Install command

Only Postgres dashboards are enabled to stay within Helm's 1MB secret size limit.

```bash
helm repo add appscode https://charts.appscode.com/stable/
helm repo update appscode

helm upgrade -i kubedb-grafana-dashboards appscode/kubedb-grafana-dashboards \
  -n kubeops \
  --create-namespace \
  --version=v2026.6.18-rc.2 \
  --set featureGates.Postgres=true \
  --set featureGates.Cassandra=false \
  --set featureGates.ClickHouse=false \
  --set featureGates.DB2=false \
  --set featureGates.Druid=false \
  --set featureGates.Elasticsearch=false \
  --set featureGates.HanaDB=false \
  --set featureGates.Hazelcast=false \
  --set featureGates.Ignite=false \
  --set featureGates.Kafka=false \
  --set featureGates.MariaDB=false \
  --set featureGates.Memcached=false \
  --set featureGates.Milvus=false \
  --set featureGates.MongoDB=false \
  --set featureGates.MSSQLServer=false \
  --set featureGates.MySQL=false \
  --set featureGates.Neo4j=false \
  --set featureGates.Oracle=false \
  --set featureGates.PerconaXtraDB=false \
  --set featureGates.PgBouncer=false \
  --set featureGates.Pgpool=false \
  --set featureGates.ProxySQL=false \
  --set featureGates.Qdrant=false \
  --set featureGates.RabbitMQ=false \
  --set featureGates.Redis=false \
  --set featureGates.Singlestore=false \
  --set featureGates.Solr=false \
  --set featureGates.Weaviate=false \
  --set featureGates.ZooKeeper=false \
  --set grafana.url="http://prometheus-grafana.monitoring.svc:80" \
  --set grafana.apikey="<token-from-above>"
```

### Verify dashboards

```bash
kubectl --kubeconfig=/home/banusree/all_db.yaml -n kubeops get grafanadashboards
# NAME                       TITLE                          STATUS   AGE
# kubedb-postgres-database   KubeDB / Postgres / Database            <age>
# kubedb-postgres-pod        KubeDB / Postgres / Pod                 <age>
# kubedb-postgres-summary    KubeDB / Postgres / Summary             <age>
```

---

## Alert Reference

All alerts are scoped to the `pg-grafana-demo` instance in the `demo` namespace via PromQL filters:  
`job="postgres-alerts-stats", namespace="demo"`.

### Database Group

These alerts fire based on metrics from `postgres_exporter`.

| Alert | Severity | Expression Summary | For | What It Means |
|-------|----------|--------------------|-----|---------------|
| `PostgresqlDown` | critical | `pg_up == 0` | instant | The exporter cannot reach PostgreSQL — the instance is down or the exporter crashed. |
| `PostgresqlSplitBrain` | critical | `count(pg_replication_is_replica == 0) > 1` | instant | More than one node is acting as primary — data divergence risk in a replica set. |
| `PostgresqlTooManyLocksAcquired` | critical | `locks / (max_locks_per_tx * max_connections) > threshold` | 2m | Lock table is nearly full; transactions may start failing with "out of shared memory". |
| `PostgresReplicationSlotLagHigh` | warning | `pg_replication_slots_pg_wal_lsn_diff > 800MB` | 1m | A replication slot consumer is falling behind; WAL files are accumulating on disk. |
| `PostgresReplicationSlotLagCritical` | critical | `pg_replication_slots_pg_wal_lsn_diff > 1.2GB` | 1m | Slot lag is critical — disk exhaustion or slot invalidation is imminent. |
| `PostgresqlRestarted` | critical | `time() - pg_postmaster_start_time_seconds < 60` | instant | PostgreSQL restarted within the last minute — unexpected restart detected. |
| `PostgresqlExporterError` | warning | `pg_exporter_last_scrape_error == 1` | 5m | The exporter itself has errors — metrics may be missing or stale. |
| `PostgresqlHighRollbackRate` | warning | `rate(xact_rollback) / rate(xact_commit) > threshold` | instant | High proportion of transactions are rolling back — indicates application errors or lock contention. |
| `PostgresTooManyConnections` | warning | `sum(pg_stat_activity_count) / max_connections > 95%` | 2m | Connection pool is nearly exhausted; new connections may be refused. |
| `DiskUsageHigh` | warning | `kubelet_volume_stats_used_bytes / capacity > 80%` | 1m | PVC used space exceeds 80% — time to plan for expansion. |
| `DiskAlmostFull` | critical | `kubelet_volume_stats_used_bytes / capacity > 95%` | 1m | PVC almost full — PostgreSQL may become read-only or crash. |

### Provisioner Group

These alerts monitor the KubeDB operator's view of the PostgreSQL resource phase.

| Alert | Severity | Expression Summary | For | What It Means |
|-------|----------|--------------------|-----|---------------|
| `KubeDBPostgreSQLPhaseNotReady` | critical | `kubedb_com_postgres_status_phase == "NotReady"` | 1m | KubeDB has marked the Postgres resource `NotReady` — operator cannot reach the database. |
| `KubeDBPostgreSQLPhaseCritical` | warning | `kubedb_com_postgres_status_phase == "Critical"` | 15m | One or more replicas are down; the cluster is degraded but the primary is still up. |

### OpsManager Group

These alerts track `PostgresOpsRequest` lifecycle — used during upgrades, scaling, reconfiguration, and certificate rotations.

| Alert | Severity | Expression Summary | For | What It Means |
|-------|----------|--------------------|-----|---------------|
| `KubeDBPostgreSQLOpsRequestStatusProgressingToLong` | critical | `ops_request_status == "Progressing"` | 30m | An ops request has been running for 30+ minutes — stuck or failed mid-way. |
| `KubeDBPostgreSQLOpsRequestFailed` | critical | `ops_request_status == "Failed"` | instant | An ops request failed — check the OpsRequest object for the error. |

### Backup & Restore Groups (Stash / KubeStash)

Two parallel sets of alerts cover both the legacy Stash and current KubeStash backup frameworks.

| Alert | Severity | What It Means |
|-------|----------|---------------|
| `PostgreSQL*BackupSessionFailed` | critical | A scheduled backup run failed — check BackupSession logs. |
| `PostgreSQL*RestoreSessionFailed` | critical | A restore operation failed. |
| `PostgreSQL*NoBackupSessionForTooLong` | critical | No successful backup in the expected window — backup may be misconfigured or stuck. |
| `PostgreSQL*RepositoryCorrupted` | critical | The backup repository integrity check failed — backups may be unrestorable. |
| `PostgreSQL*RepositoryStorageRunningLow` | warning | Backup storage has less than 10GB free. |
| `PostgreSQL*BackupSessionPeriodTooLong` | warning | A backup took longer than 30 minutes — may indicate slow storage or large database size. |
| `PostgreSQL*RestoreSessionPeriodTooLong` | warning | A restore took longer than 30 minutes. |

### SchemaManager Group

These alerts monitor `PostgresDatabase` schema lifecycle objects managed by KubeDB Schema Manager.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBPostgreSQLSchemaPendingForTooLong` | warning | 30m | A schema object has been in `Pending` state for 30+ minutes — may indicate waiting for a dependency. |
| `KubeDBPostgreSQLSchemaInProgressForTooLong` | warning | 30m | Schema migration is running for 30+ minutes — may be stuck. |
| `KubeDBPostgreSQLSchemaTerminatingForTooLong` | warning | 30m | Schema deletion is stuck — a finalizer may be blocking it. |
| `KubeDBPostgreSQLSchemaFailed` | warning | instant | Schema operation failed. |
| `KubeDBPostgreSQLSchemaExpired` | warning | instant | A schema with a TTL has expired and been revoked. |

---

## How Alerting Works End-to-End

### 1. Metrics Collection

KubeDB injects `postgres_exporter` as a sidecar into every PostgreSQL pod. It connects to the database and exposes metrics at `:56790/metrics`.

```bash
# Verify the exporter is running inside the pod
kubectl --kubeconfig=/home/banusree/all_db.yaml -n demo exec pg-grafana-demo-0 \
  -c postgres-exporter -- wget -qO- localhost:56790/metrics | grep pg_up
# pg_up{...} 1
```

### 2. ServiceMonitor Discovery

The `pg-grafana-demo-stats` ServiceMonitor tells Prometheus to scrape the stats service every 10 seconds. The ServiceMonitor carries label `release: prometheus` matching the Prometheus `serviceMonitorSelector`, so Prometheus picks it up automatically.

```bash
kubectl --kubeconfig=/home/banusree/all_db.yaml -n demo get servicemonitor pg-grafana-demo-stats -o yaml
```

The scrape job is named `postgres-alerts-stats` — this is the `job` label used in all PromQL expressions.

### 3. PrometheusRule Evaluation

Prometheus loads the `postgres-alerts` PrometheusRule because it has `release: prometheus` matching the `ruleSelector`. Every 30 seconds (global scrape interval), Prometheus evaluates each rule expression:

```
pg_up{job="postgres-alerts-stats", namespace="demo"} == 0
```

If the expression evaluates to a non-empty result for longer than the `for` duration, the alert transitions from `pending` → `firing`.

### 4. AlertManager Routing

Firing alerts are sent to AlertManager at `prometheus-kube-prometheus-alertmanager:9093`. AlertManager groups, inhibits, and silences alerts according to its configuration, then routes them to receivers (default: no receiver — configure via `alertmanagerConfig` or Helm values of kube-prometheus-stack).

To check currently firing alerts:
```bash
kubectl --kubeconfig=/home/banusree/all_db.yaml -n monitoring \
  port-forward svc/prometheus-kube-prometheus-alertmanager 9093:9093 &
# Open http://localhost:9093 in browser
```

To check Prometheus rule status:
```bash
kubectl --kubeconfig=/home/banusree/all_db.yaml -n monitoring \
  port-forward svc/prometheus-kube-prometheus-prometheus 9090:9090 &
# Open http://localhost:9090/alerts in browser
```

### 5. Grafana Visualisation

The three `GrafanaDashboard` CRDs contain pre-built dashboard JSON for PostgreSQL:

| Dashboard | What It Shows |
|-----------|---------------|
| `KubeDB / Postgres / Summary` | High-level health, connection count, uptime, phase |
| `KubeDB / Postgres / Database` | Query rates, transaction rates, rollbacks, lock counts, replication lag |
| `KubeDB / Postgres / Pod` | Pod-level CPU, memory, disk I/O, network |

Access Grafana:
```bash
kubectl --kubeconfig=/home/banusree/all_db.yaml -n monitoring \
  port-forward svc/prometheus-grafana 3000:80 &
# Open http://localhost:3000 — admin / <password from prometheus-grafana secret>
```

---

## Installed Helm Releases

```bash
helm --kubeconfig=/home/banusree/all_db.yaml list --all-namespaces | grep -E "postgres-alerts|kubedb-grafana"
```

| Release | Namespace | Chart | Version |
|---------|-----------|-------|---------|
| `postgres-alerts` | `demo` | `postgres-alerts` | `v2026.2.24` |
| `kubedb-grafana-dashboards` | `kubeops` | `kubedb-grafana-dashboards` | `v2026.6.18-rc.2` |

## Uninstall

```bash
# Remove alerts
helm --kubeconfig=/home/banusree/all_db.yaml uninstall postgres-alerts -n demo

# Remove Grafana dashboards
helm --kubeconfig=/home/banusree/all_db.yaml uninstall kubedb-grafana-dashboards -n kubeops
```

## Customising Alerts

To override thresholds or disable specific alert groups, create a `values.yaml` and upgrade:

```yaml
# custom-alerts.yaml
form:
  alert:
    labels:
      release: prometheus
    groups:
      database:
        enabled: warning
        rules:
          postgresqlTooManyConnections:
            enabled: true
            duration: "5m"
            val: 90   # trigger at 90% instead of 95%
            severity: warning
      stash:
        enabled: "none"   # disable all stash alerts
```

```bash
helm upgrade postgres-alerts oci://ghcr.io/appscode-charts/postgres-alerts \
  -n demo \
  --version=v2026.2.24 \
  --set metadata.release.name=pg-grafana-demo \
  --set metadata.release.namespace=demo \
  -f custom-alerts.yaml
```

## Sources

- Alert chart: https://github.com/opnpulse/alerts/tree/master/charts/postgres-alerts
- Grafana dashboards chart: https://github.com/kubedb/installer/tree/master/charts/kubedb-grafana-dashboards
