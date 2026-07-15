---
title: MySQL Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-monitoring-alerting
    name: Alerting
    parent: guides-mysql-monitoring
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MySQL Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed MySQL instance using the `mysql-alerts` Helm chart.

## Before You Begin

* Ensure you have a Kubernetes cluster and that `kubectl` is configured to communicate with it. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Install the KubeDB operator by following the steps [here](/docs/setup/README.md).

* Deploy the database in the `alert-mysql` namespace:

  ```bash
  $ kubectl create ns alert-mysql
  namespace/alert-mysql created
  ```

* This tutorial assumes you already have a **kube-prometheus-stack** running in your cluster, with `Prometheus` configured so that both `serviceMonitorSelector` and `ruleSelector` match the label `release: prometheus`. See the [Grafana Dashboard](grafana-dashboard.md#configuration) guide for how to deploy kube-prometheus-stack if you don't have it yet.

  To verify the selectors:

  ```bash
  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.ruleSelector}'
  {"matchLabels":{"release":"prometheus"}}

  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.serviceMonitorSelector}'
  {"matchLabels":{"release":"prometheus"}}
  ```

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/mysql/monitoring/overview/index.md).

* For dashboards and visualisation, see [Grafana Dashboard](grafana-dashboard.md) for MySQL.

> Note: YAML files used in this tutorial are stored in [docs/examples/mysql](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mysql) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

- **KubeDB** deploys MySQL with a built-in `mysqld_exporter` sidecar (container name `exporter`) that exposes metrics on port `56790`.
- **ServiceMonitor** (named `{mysql-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape the exporter every 10 seconds.
- **PrometheusRule** is created by the `mysql-alerts` chart and contains all MySQL alert definitions grouped by concern: database health, group replication, provisioner, ops-manager, and backups (Stash and KubeStash).
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).

---

## Deploy MySQL with Monitoring Enabled

At first, let's deploy a MySQL database with monitoring enabled. Below is the MySQL object we are going to create. Note that `spec.storage.storageClassName` is set to `longhorn` so the database's data volume is backed by Longhorn-replicated block storage rather than node-local storage.

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-alert-alert-mysql
  namespace: alert-mysql
spec:
  version: "8.4.8"
  deletionPolicy: WipeOut
  storage:
    storageClassName: "longhorn"
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
```

Here,

- `spec.storage.storageClassName: "longhorn"` provisions the data volume from the `longhorn` `StorageClass` instead of the default `local-path`, giving the volume replicated, node-independent storage.
- `spec.monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` resource managed by the Prometheus operator.
- `spec.monitor.prometheus.serviceMonitor.labels.release: prometheus` adds the `release: prometheus` label to the created `ServiceMonitor`, matching the Prometheus `serviceMonitorSelector` so the target is discovered automatically.

Let's create the namespace and the MySQL resource.

```bash
$ kubectl create ns alert-mysql
namespace/alert-mysql created

$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mysql/monitoring/mysql-alert-alert-mysql.yaml
mysql.kubedb.com/mysql-alert-alert-mysql created
```

Now, wait for the database to go into `Ready` state.

```bash
$ kubectl get mysql -n alert-mysql mysql-alert-alert-mysql
NAME                      VERSION   STATUS   AGE
mysql-alert-alert-mysql   8.4.8     Ready    17m
```

Confirm the data volume is actually `Bound` on the `longhorn` `StorageClass`.

```bash
$ kubectl get pvc -n alert-mysql
NAME                             STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-mysql-alert-alert-mysql-0   Bound    pvc-dee12333-ab1f-40df-958d-40edc156b30c   1Gi        RWO            longhorn       17m
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n alert-mysql --selector="app.kubernetes.io/instance=mysql-alert-alert-mysql"
NAME                            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
mysql-alert-alert-mysql         ClusterIP   10.43.149.251   <none>        3306/TCP    17m
mysql-alert-alert-mysql-pods    ClusterIP   None            <none>        3306/TCP    17m
mysql-alert-alert-mysql-stats   ClusterIP   10.43.91.232    <none>        56790/TCP   17m
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n alert-mysql
NAME                            AGE
mysql-alert-alert-mysql-stats   17m
```

Verify that the `ServiceMonitor` carries the `release: prometheus` label so Prometheus discovers it.

```bash
$ kubectl get servicemonitor -n alert-mysql mysql-alert-alert-mysql-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Install mysql-alerts

The `mysql-alerts` chart creates a `PrometheusRule` resource containing all MySQL alert definitions grouped by concern: database health, group replication, provisioner, ops-manager, and backups (Stash / KubeStash).

### Why the Helm release name matters

The chart derives the PromQL `job`/instance scoping (and the `PrometheusRule` name) from the **Helm release name**, not from a values field — so the release name must match the MySQL object's name (`mysql-alert-alert-mysql`) for the rules to be correctly scoped to this instance.

The chart's default label is `release: kube-prometheus-stack`, so we must also override it at install time to match the Prometheus `ruleSelector`.

### Install

```bash
$ helm upgrade -i mysql-alert-alert-mysql oci://ghcr.io/appscode-charts/mysql-alerts \
    -n alert-mysql \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus
```

| Flag | Value | Purpose |
|------|-------|---------|
| `mysql-alert-alert-mysql` (release name) | — | Scopes every PromQL expression to this instance (`job="mysql-alert-alert-mysql-stats"`) |
| `-n alert-mysql` | `alert-mysql` | Installs the `PrometheusRule` in the same namespace as the database |
| `form.alert.labels.release` | `prometheus` | Matches the Prometheus `ruleSelector` so the rules are loaded |

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-mysql
NAME                      AGE
mysql-alert-alert-mysql   30s
```

Confirm the `release: prometheus` label is present.

```bash
$ kubectl get prometheusrule -n alert-mysql mysql-alert-alert-mysql \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Confirm Prometheus loaded the rules

Port-forward the Prometheus UI.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules` and search for **mysql-alert-alert-mysql**.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/mysql/monitoring/mysql-alerting-prom-rules.png" style="padding:10px">
</p>

The `mysql.database.alert-mysql.mysql-alert-alert-mysql.rules` group (and the accompanying `mysql.group`, `mysql.provisioner`, `mysql.opsManager`, `mysql.stash`, `mysql.kubeStash`, and `mysql.schemaManager` groups) are visible with all rules showing **OK**, confirming that Prometheus has loaded and is evaluating the MySQL alert definitions every 30 seconds.

> **Chart note:** unlike some other `*-alerts` charts, every alert group declared in `mysql-alerts`' `values.yaml` (`database`, `group`, `provisioner`, `opsManager`, `stash`, `kubeStash`, `schemaManager`) is actually rendered into the `PrometheusRule` at v2026.7.14 — there is no group silently missing from the template.

---

## Verify End-to-End

### 1. Check the exporter is running

The `exporter` sidecar inside the MySQL pod serves metrics at `:56790/metrics`. A value of `mysql_up 1` confirms the exporter can reach MySQL.

```bash
$ kubectl exec -n alert-mysql mysql-alert-alert-mysql-0 -c exporter -- \
    wget -qO- localhost:56790/metrics | grep mysql_up
mysql_up 1
```

### 2. Check the Prometheus target is UP

Prometheus discovers more than 20 scrape pools on a shared cluster, so instead of the Target health page, query `up` directly for a reliable view.

Open `http://localhost:9090/query?g0.expr=up%7Bnamespace%3D%22alert-mysql%22%7D&g0.tab=1`.

<p align="center">
  <img alt="Prometheus up query — target UP" src="/docs/images/mysql/monitoring/mysql-alerting-prom-target.png" style="padding:10px">
</p>

The target reports `up == 1` for `mysql-alert-alert-mysql-0` in the `alert-mysql` namespace, confirming Prometheus is scraping the exporter on the longhorn-backed pod.

### 3. Confirm all MySQL alerts are inactive

Open `http://localhost:9090/alerts` and locate the `mysql-alert-alert-mysql` groups.

<p align="center">
  <img alt="Prometheus Alerts — All Inactive" src="/docs/images/mysql/monitoring/mysql-alerting-prom-alerts.png" style="padding:10px">
</p>

All 13 rules in the `mysql.database` group show **INACTIVE (13)**, meaning the database is healthy and no thresholds are breached. This also confirms `DiskUsageHigh`/`DiskAlmostFull` are inactive — this chart's disk-usage PromQL correctly divides against `kubelet_volume_stats_capacity_bytes`, so (unlike some other `*-alerts` charts) it does not falsely fire on a healthy volume regardless of storage backend.

### 4. Check AlertManager

Port-forward AlertManager to view any currently firing alerts.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`. With a healthy MySQL instance, no alerts for `mysql-alert-alert-mysql` will be listed here.

<p align="center">
  <img alt="AlertManager" src="/docs/images/mysql/monitoring/mysql-alerting-alertmanager.png" style="padding:10px">
</p>

### 5. Visualise metrics with Grafana

MySQL metrics can be visualised on the pre-built KubeDB Grafana dashboards. See [Grafana Dashboard](grafana-dashboard.md) for MySQL for how to install and explore them — they are not duplicated here.

---

## Simulating a Firing Alert

The previous section confirmed that all alerts are **INACTIVE** while the database is healthy. This section walks through deliberately triggering the `MySQLInstanceDown` critical alert so you can observe the full alert lifecycle — from firing in Prometheus through to the AlertManager dashboard — and then resolve it.

The `exporter` sidecar runs as a **separate container** from `mysql`, so it keeps running even after the main `mysql` container crashes and gets restarted by Kubernetes. `MySQLInstanceDown`/`MySQLServiceDown` fire as soon as a single scrape observes `mysql_up == 0` (`for: 0m`), but Kubernetes tends to restart a crashed container within a few seconds, so a single `kill` can recover before the next 10-second scrape catches it. Repeatedly killing the `mysql` process for a short window makes the outage reliably visible to Prometheus. The provisioner alert `KubeDBMySQLPhaseNotReady` has `for: 1m`, so keep the loop running for at least a minute to also observe that alert fire.

### 1. Crash the MySQL process repeatedly

```bash
$ while true; do
    kubectl exec -n alert-mysql mysql-alert-alert-mysql-0 -c mysql -- sh -c "kill 1" >/dev/null 2>&1
    sleep 3
  done
```

Let this loop run for a couple of minutes (leave it running while you check the next steps), then stop it once you've captured the firing state.

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — MySQLInstanceDown Firing" src="/docs/images/mysql/monitoring/mysql-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

`MySQLInstanceDown` and `MySQLServiceDown` transition from **INACTIVE** to **FIRING** once the exporter reports `mysql_up == 0` on a scrape — since both have `for: 0m`, they fire on the very next evaluation cycle after the metric goes stale/zero.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager — MySQLInstanceDown Firing" src="/docs/images/mysql/monitoring/mysql-alerting-alertmanager-firing.png" style="padding:10px">
</p>

AlertManager shows both `MySQLInstanceDown` (sourced from the exporter's `mysql_up` metric) and `KubeDBMySQLPhaseNotReady` (sourced from the KubeDB operator's own view of the resource, exported via Panopticon) once the crash loop has persisted past `KubeDBMySQLPhaseNotReady`'s `for: 1m` window. The alert cards display:

- **Severity**: `critical`
- **pod** / **mysql**: `mysql-alert-alert-mysql-0` / `mysql-alert-alert-mysql` in the `alert-mysql` namespace
- **job**: `mysql-alert-alert-mysql-stats`
- **Started**: timestamp when the alert first fired

AlertManager routes this alert to every receiver configured in your `alertmanagerConfig` (Slack, email, PagerDuty, webhook, etc.) based on your routing tree. If no receiver is configured, the alert is visible here but silently dropped.

### 4. Restore MySQL

Stop the loop from step 1. The container recovers on its own — Kubernetes just needs a few uninterrupted seconds without a fresh `kill` to let `mysqld` finish starting up.

```bash
$ kubectl get pods -n alert-mysql -w
NAME                        READY   STATUS    RESTARTS   AGE
mysql-alert-alert-mysql-0   2/2     Running   19         42m
```

Once `mysql_up` returns to `1` continuously and the resource phase returns to `Ready`, Prometheus marks the alerts **INACTIVE** again and AlertManager sends a **resolved** notification to all receivers.

---

## Alert Reference

All alerts are scoped to the `mysql-alert-alert-mysql` instance in the `alert-mysql` namespace via the PromQL label filters `job="mysql-alert-alert-mysql-stats"` and `namespace="alert-mysql"` (database/group groups), or `app="mysql-alert-alert-mysql"` and `namespace="alert-mysql"` (provisioner/opsManager/stash/kubeStash/schemaManager groups).

### Database Group

Fired based on live metrics from `mysqld_exporter`.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `MySQLInstanceDown` | critical | instant | Exporter cannot reach MySQL — instance is down or `mysqld` crashed. |
| `MySQLServiceDown` | critical | instant | No pod behind the stats service reports `mysql_up == 1` — the service has no healthy backend. |
| `MySQLTooManyConnections` | warning | 2m | More than 80% of `max_connections` are in use — connection pool nearing exhaustion. |
| `MySQLHighThreadsRunning` | warning | 2m | More than 60% of `max_connections` worth of threads are actively running — the server is under heavy load. |
| `MySQLSlowQueries` | warning | 2m | New slow queries have been logged in the last minute. |
| `MySQLInnoDBLogWaits` | warning | instant | InnoDB log writes are stalling (>10 waits in 15m) — I/O may be a bottleneck. |
| `MySQLRestarted` | warning | instant | MySQL uptime is under 60 seconds — the server restarted recently. |
| `MySQLHighQPS` | critical | instant | Queries per second exceed 1000 — unusually high query load. |
| `MySQLHighIncomingBytes` | critical | instant | Incoming network traffic exceeds 1MB/s. |
| `MySQLHighOutgoingBytes` | critical | instant | Outgoing network traffic exceeds 1MB/s. |
| `MySQLTooManyOpenFiles` | warning | 2m | More than 80% of `open_files_limit` are in use. |
| `DiskUsageHigh` | warning | 1m | Persistent volume usage exceeds 80% — plan for expansion. |
| `DiskAlmostFull` | critical | 1m | Persistent volume usage exceeds 95% — MySQL may become read-only or crash. |

### Group Replication Group

Fired based on MySQL Group Replication performance-schema metrics; only meaningful for clustered/group-replication deployments.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `MySQLHighReplicationDelay` | warning | 5m | Group Replication apply time on a member exceeds 0.5s. |
| `MySQLHighReplicationTransportTime` | warning | 5m | Group Replication transport time exceeds 0.5s — network lag between members. |
| `MySQLHighReplicationApplyTime` | warning | 5m | Group Replication apply time exceeds 0.5s. |
| `MySQLReplicationHighTransactionTime` | warning | 5m | Transaction time on the local relay queue exceeds 0.5s — the member is falling behind. |

> **Chart bug (v2026.7.14):** `MySQLHighReplicationDelay` and `MySQLHighReplicationApplyTime` render with the **identical** PromQL expression (`mysql_perf_schema_replication_group_worker_apply_time_seconds`) — they will always fire and clear together. `MySQLHighReplicationDelay` was almost certainly intended to alert on a different metric (e.g. an actual replication-lag/delay gauge). Neither alert can be exercised on this tutorial's single-node MySQL instance regardless, since Group Replication metrics only populate on a clustered deployment.

### Provisioner Group

Monitors the KubeDB operator's view of the MySQL resource phase (sourced from Panopticon, not the MySQL metrics endpoint).

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBMySQLPhaseNotReady` | critical | 1m | KubeDB marked the MySQL resource `NotReady` — operator cannot reach a healthy instance. |
| `KubeDBMySQLPhaseCritical` | warning | 15m | The MySQL resource is in a degraded/critical phase. |

### OpsManager Group

Tracks `MySQLOpsRequest` lifecycle during upgrades, scaling, reconfiguration, and certificate rotations.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBMySQLOpsRequestStatusProgressingToLong` | critical | 30m | A `MySQLOpsRequest` has been in progress for 30+ minutes — likely stuck. |
| `KubeDBMySQLOpsRequestFailed` | critical | instant | A `MySQLOpsRequest` failed — check the `MySQLOpsRequest` object for the error. |

### Stash Group

Tracks Stash-driven backup/restore health for this instance. (You do not need to configure backups to see these rules; they are included in the `PrometheusRule` regardless.)

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `MySQLStashBackupSessionFailed` | critical | instant | A Stash backup session failed. |
| `MySQLStashRestoreSessionFailed` | critical | instant | A Stash restore session failed. |
| `MySQLStashNoBackupSessionForTooLong` | warning | instant | No successful backup session for more than 18000s (5 hours). |
| `MySQLStashRepositoryCorrupted` | critical | 5m | The Stash backup repository integrity check failed — repository is corrupted. |
| `MySQLStashRepositoryStorageRunningLow` | warning | 5m | Backup repository storage size has exceeded 10GB. |
| `MySQLStashBackupSessionPeriodTooLong` | warning | instant | A backup session took more than 1800s (30 minutes) to complete. |
| `MySQLStashRestoreSessionPeriodTooLong` | warning | instant | A restore session took more than 1800s (30 minutes) to complete. |

### KubeStash Group

Tracks KubeStash-driven backup/restore health for this instance. Same semantics as the Stash group above, sourced from KubeStash metrics instead.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `MySQLKubeStashBackupSessionFailed` | critical | instant | A KubeStash backup session failed. |
| `MySQLKubeStashRestoreSessionFailed` | critical | instant | A KubeStash restore session failed. |
| `MySQLKubeStashNoBackupSessionForTooLong` | warning | instant | No successful backup session for more than 18000s (5 hours). |
| `MySQLKubeStashRepositoryCorrupted` | critical | 5m | The KubeStash backup repository integrity check failed — repository is corrupted. |
| `MySQLKubeStashRepositoryStorageRunningLow` | warning | 5m | Backup repository storage size has exceeded 10GB. |
| `MySQLKubeStashBackupSessionPeriodTooLong` | warning | instant | A backup session took more than 1800s (30 minutes) to complete. |
| `MySQLKubeStashRestoreSessionPeriodTooLong` | warning | instant | A restore session took more than 1800s (30 minutes) to complete. |

### SchemaManager Group

Monitors `MySQLDatabase` schema lifecycle objects managed by KubeDB Schema Manager.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBMySQLSchemaPendingForTooLong` | warning | 30m | Schema object stuck in `Pending` — may be waiting on a dependency. |
| `KubeDBMySQLSchemaInProgressForTooLong` | warning | 30m | Schema migration running for 30+ minutes — may be stuck. |
| `KubeDBMySQLSchemaTerminatingForTooLong` | warning | 30m | Schema deletion stuck — a finalizer may be blocking it. |
| `KubeDBMySQLSchemaFailed` | warning | instant | Schema operation failed. |
| `KubeDBMySQLSchemaExpired` | warning | instant | A schema with a TTL has expired and been revoked. |

---

## Customising Alerts

To override thresholds or disable specific alert groups, create a custom values file and upgrade the chart.

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
          mysqlTooManyConnections:
            enabled: true
            duration: "5m"
            val: 90        # fire at 90% instead of the default 80%
            severity: warning
      opsManager:
        enabled: "none"    # disable all ops-manager alerts
```

```bash
$ helm upgrade mysql-alert-alert-mysql oci://ghcr.io/appscode-charts/mysql-alerts \
    -n alert-mysql \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

To remove all resources created in this tutorial, run the following commands.

```bash
# Remove the mysql-alerts release
$ helm uninstall mysql-alert-alert-mysql -n alert-mysql

# Remove the MySQL instance
$ kubectl delete mysql -n alert-mysql mysql-alert-alert-mysql

# Delete namespace
$ kubectl delete ns alert-mysql
```

## Next Steps

- Monitor your MySQL database with KubeDB using [builtin Prometheus](/docs/guides/mysql/monitoring/builtin-prometheus/index.md).
- Monitor your MySQL database with KubeDB using [Prometheus operator](/docs/guides/mysql/monitoring/prometheus-operator/index.md).
- Visualise MySQL metrics with [Grafana Dashboard](grafana-dashboard.md).
- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
