---
title: PerconaXtraDB Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-monitoring-alerting
    name: Alerting
    parent: guides-perconaxtradb-monitoring
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PerconaXtraDB Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed PerconaXtraDB instance using the `perconaxtradb-alerts` Helm chart.

## Before You Begin

* Ensure you have a Kubernetes cluster and that `kubectl` is configured to communicate with it. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Install the KubeDB operator by following the steps [here](/docs/setup/README.md).

* Deploy the database in the `alert-perconaxtradb` namespace:

  ```bash
  $ kubectl create ns alert-perconaxtradb
  namespace/alert-perconaxtradb created
  ```

* Before proceeding, complete the [Configuration](grafana-dashboard.md#configuration) steps to deploy **kube-prometheus-stack** and **Panopticon**.

* This tutorial assumes you already have a **kube-prometheus-stack** running in your cluster, with `Prometheus` configured so that both `serviceMonitorSelector` and `ruleSelector` match the label `release: prometheus`.

  To verify the selectors:

  ```bash
  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.ruleSelector}'
  {"matchLabels":{"release":"prometheus"}}

  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.serviceMonitorSelector}'
  {"matchLabels":{"release":"prometheus"}}
  ```

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/percona-xtradb/monitoring/overview/index.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/percona-xtradb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/percona-xtradb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

- **KubeDB** deploys PerconaXtraDB with a `mysqld_exporter`-compatible sidecar (container `exporter`) that exposes metrics (`mysql_*`), the same exporter family used by MySQL/MariaDB.
- **ServiceMonitor** (named `{perconaxtradb-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape the exporter every 10 seconds.
- **PrometheusRule** is created by the `perconaxtradb-alerts` chart and contains alert definitions grouped by concern: database health, Galera cluster, provisioner, ops-manager, Stash backup/restore, and schema manager.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).
- **Grafana** dashboards for PerconaXtraDB are covered separately — see [Grafana Dashboard](grafana-dashboard.md) rather than duplicated here.

---

## Deploy PerconaXtraDB with Monitoring Enabled

Below is the PerconaXtraDB object we are going to create — a single standalone instance with monitoring enabled. (The chart's `cluster` group only produces data if you deploy a Galera cluster via `spec.topology`; a standalone instance simply leaves that alert INACTIVE.)

```yaml
apiVersion: kubedb.com/v1
kind: PerconaXtraDB
metadata:
  name: perconaxtradb-alert-demo
  namespace: alert-perconaxtradb
spec:
  version: "8.0.40"
  storageType: Durable
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/percona-xtradb/monitoring/perconaxtradb-alert-demo.yaml
perconaxtradb.kubedb.com/perconaxtradb-alert-demo created
```

Wait for the database to go into `Ready` state.

```bash
$ kubectl get perconaxtradb -n alert-perconaxtradb perconaxtradb-alert-demo
NAME                       VERSION   STATUS   AGE
perconaxtradb-alert-demo   8.0.40    Ready    3m
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n alert-perconaxtradb --selector="app.kubernetes.io/instance=perconaxtradb-alert-demo"
NAME                                   TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
perconaxtradb-alert-demo               ClusterIP   10.43.10.20    <none>        3306/TCP    3m
perconaxtradb-alert-demo-pods          ClusterIP   None           <none>        3306/TCP    3m
perconaxtradb-alert-demo-stats         ClusterIP   10.43.10.21    <none>        56790/TCP   3m
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n alert-perconaxtradb
NAME                             AGE
perconaxtradb-alert-demo-stats   3m

$ kubectl get servicemonitor -n alert-perconaxtradb perconaxtradb-alert-demo-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Install perconaxtradb-alerts

### Why the Helm release name matters

The chart derives the `PrometheusRule` name and scopes every PromQL expression (via `job="{release-name}-stats"` / `app="{release-name}"`) from the **Helm release name** — so the release name must match the PerconaXtraDB object's name (`perconaxtradb-alert-demo`). Note the chart/repo name has no hyphen (`perconaxtradb-alerts`), even though the KubeDB guides directory uses `percona-xtradb`.

### Install

```bash
$ helm upgrade -i perconaxtradb-alert-demo oci://ghcr.io/appscode-charts/perconaxtradb-alerts \
    -n alert-perconaxtradb \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus
```

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-perconaxtradb
NAME                       AGE
perconaxtradb-alert-demo   30s

$ kubectl get prometheusrule -n alert-perconaxtradb perconaxtradb-alert-demo \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Confirm Prometheus loaded the rules

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules` and locate the `perconaxtradb.database`, `perconaxtradb.cluster`, `perconaxtradb.provisioner`, `perconaxtradb.opsManager`, `perconaxtradb.stash`, and `perconaxtradb.schemaManager` groups.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/percona-xtradb/monitoring/perconaxtradb-alerting-prom-rules.png" style="padding:10px">
</p>

All groups should show **OK**. Unlike MariaDB's chart, `perconaxtradb-alerts` v2026.7.14 has a `stash` group but **no** `kubeStash` group — every group it does declare in `values.yaml` renders correctly.

---

## Verify End-to-End

### 1. Check the Prometheus target is UP

Open `http://localhost:9090/query?g0.expr=up%7Bnamespace%3D%22alert-perconaxtradb%22%7D&g0.tab=1`.

<p align="center">
  <img alt="Prometheus up query — perconaxtradb-alert-demo-0 UP" src="/docs/images/percona-xtradb/monitoring/perconaxtradb-alerting-prom-target.png" style="padding:10px">
</p>

### 2. Confirm the PerconaXtraDB alerts are inactive

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — PerconaXtraDB groups inactive" src="/docs/images/percona-xtradb/monitoring/perconaxtradb-alerting-prom-alerts.png" style="padding:10px">
</p>

All rules should show **INACTIVE**. `GaleraReplicationLatencyTooLong` has no data on a standalone instance.

### 3. Check AlertManager

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager" src="/docs/images/percona-xtradb/monitoring/perconaxtradb-alerting-alertmanager.png" style="padding:10px">
</p>

### 4. Grafana dashboard

See [Grafana Dashboard](grafana-dashboard.md) for how to provision and explore the PerconaXtraDB dashboards (via the `kubedb-grafana-dashboards` chart, `--set featureGates.PerconaXtraDB=true`).

---

## Simulating a Firing Alert

This section deliberately triggers `PerconaXtraDBInstanceDown` (instant, `for: 0m`) by crashing the main database process.

### 1. Crash the PerconaXtraDB process

```bash
$ kubectl exec -n alert-perconaxtradb perconaxtradb-alert-demo-0 -c perconaxtradb -- sh -c '
    end=$(( $(date +%s) + 30 ));
    while [ $(date +%s) -lt $end ]; do
      pid=$(pgrep -x mysqld | head -1);
      [ -n "$pid" ] && kill -9 "$pid" 2>/dev/null;
      sleep 1;
    done'
```

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — PerconaXtraDBInstanceDown Firing" src="/docs/images/percona-xtradb/monitoring/perconaxtradb-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

`PerconaXtraDBInstanceDown` (`mysql_up == 0`) should transition straight to **FIRING**.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager — PerconaXtraDBInstanceDown Firing" src="/docs/images/percona-xtradb/monitoring/perconaxtradb-alerting-alertmanager-firing.png" style="padding:10px">
</p>

### 4. Restore PerconaXtraDB

Stop the loop from step 1.

```bash
$ kubectl get perconaxtradb -n alert-perconaxtradb perconaxtradb-alert-demo -w
NAME                       VERSION   STATUS   AGE
perconaxtradb-alert-demo   8.0.40    Ready    24m
```

If PerconaXtraDB does not recover on its own within a minute or two, force a clean restart: `kubectl delete pod -n alert-perconaxtradb perconaxtradb-alert-demo-0`.

---

## Alert Reference

All alerts are scoped to the `perconaxtradb-alert-demo` instance in the `alert-perconaxtradb` namespace via the PromQL label filters `job="perconaxtradb-alert-demo-stats"` / `namespace="alert-perconaxtradb"` (database/cluster groups), or `app="perconaxtradb-alert-demo"` / `namespace="alert-perconaxtradb"` (provisioner/opsManager/stash/schemaManager groups).

### Database Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `PerconaXtraDBInstanceDown` | critical | instant | `mysql_up == 0` on this instance. |
| `PerconaXtraDBServiceDown` | critical | instant | No replica behind the service is answering. |
| `PerconaXtraDBTooManyConnections` | warning | 2m | Connection count is high relative to `max_connections`. |
| `PerconaXtraDBHighThreadsRunning` | warning | 2m | Too many threads actively running. |
| `PerconaXtraDBSlowQueries` | warning | 2m | Slow-query count is increasing. |
| `PerconaXtraDBInnoDBLogWaits` | warning | instant | InnoDB log waits are occurring. |
| `PerconaXtraDBRestarted` | warning | instant | Uptime indicates a recent restart. |
| `PerconaXtraDBHighQPS` | critical | instant | Query rate is unusually high. |
| `PerconaXtraDBHighIncomingBytes` | critical | instant | Inbound network traffic is unusually high. |
| `PerconaXtraDBHighOutgoingBytes` | critical | instant | Outbound network traffic is unusually high. |
| `PerconaXtraDBTooManyOpenFiles` | warning | 2m | Open file count is high relative to the limit. |
| `DiskUsageHigh` | warning | 1m | Persistent volume usage exceeds 80%. |
| `DiskAlmostFull` | critical | 1m | Persistent volume usage exceeds 95%. |

### Cluster Group

Only produces data when `spec.topology` (Galera) is configured.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `GaleraReplicationLatencyTooLong` | warning | 5m | Galera replication latency is high. |

### Provisioner Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBPerconaXtraDBPhaseNotReady` | critical | 1m | KubeDB marked the PerconaXtraDB resource `NotReady`. |
| `KubeDBPerconaXtraDBPhaseCritical` | warning | 15m | PerconaXtraDB is degraded but not fully unavailable. |

### OpsManager Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBPerconaXtraDBOpsRequestStatusProgressingToLong` | critical | 30m | An ops request has been running for 30+ minutes. |
| `KubeDBPerconaXtraDBOpsRequestFailed` | critical | instant | An ops request failed. |

### Stash Group

Only meaningful once Stash backup/restore is configured.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `PerconaXtraDBStashBackupSessionFailed` | critical | instant | Most recent backup session failed. |
| `PerconaXtraDBStashRestoreSessionFailed` | critical | instant | Most recent restore session failed. |
| `PerconaXtraDBStashNoBackupSessionForTooLong` | warning | instant | No recent successful backup. |
| `PerconaXtraDBStashRepositoryCorrupted` | critical | 5m | Backup repository integrity check failed. |
| `PerconaXtraDBStashRepositoryStorageRunningLow` | warning | 5m | Backup repository storage usage is high. |
| `PerconaXtraDBStashBackupSessionPeriodTooLong` | warning | instant | A backup session is taking unusually long. |
| `PerconaXtraDBStashRestoreSessionPeriodTooLong` | warning | instant | A restore session is taking unusually long. |

### SchemaManager Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBPerconaXtraDBSchemaPendingForTooLong` | warning | 30m | A `PerconaXtraDBDatabase` object stuck `Pending`. |
| `KubeDBPerconaXtraDBSchemaInProgressForTooLong` | warning | 30m | A `PerconaXtraDBDatabase` object stuck `InProgress`. |
| `KubeDBPerconaXtraDBSchemaTerminatingForTooLong` | warning | 30m | A `PerconaXtraDBDatabase` object stuck `Terminating`. |
| `KubeDBPerconaXtraDBSchemaFailed` | warning | instant | A `PerconaXtraDBDatabase` object failed. |
| `KubeDBPerconaXtraDBSchemaExpired` | warning | instant | A `PerconaXtraDBDatabase` object expired. |

---

## Customising Alerts

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
          perconaxtradbTooManyConnections:
            enabled: true
            duration: "5m"
            severity: warning
      cluster:
        enabled: "none"    # disable if you don't run Galera
```

```bash
$ helm upgrade perconaxtradb-alert-demo oci://ghcr.io/appscode-charts/perconaxtradb-alerts \
    -n alert-perconaxtradb \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

```bash
$ helm uninstall perconaxtradb-alert-demo -n alert-perconaxtradb
$ kubectl delete perconaxtradb -n alert-perconaxtradb perconaxtradb-alert-demo
$ kubectl delete ns alert-perconaxtradb
```

## Next Steps

- Monitor your PerconaXtraDB instance with KubeDB using [built-in Prometheus](/docs/guides/percona-xtradb/monitoring/builtin-prometheus/index.md).
- Monitor your PerconaXtraDB instance with KubeDB using [Prometheus operator](/docs/guides/percona-xtradb/monitoring/prometheus-operator/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
