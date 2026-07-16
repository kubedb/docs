---
title: MariaDB Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-monitoring-alerting
    name: Alerting
    parent: guides-mariadb-monitoring
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MariaDB Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed MariaDB instance using the `mariadb-alerts` Helm chart.

## Before You Begin

* Ensure you have a Kubernetes cluster and that `kubectl` is configured to communicate with it. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Install the KubeDB operator by following the steps [here](/docs/setup/README.md).

* Deploy the database in the `alert-mariadb` namespace:

  ```bash
  $ kubectl create ns alert-mariadb
  namespace/alert-mariadb created
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

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/mariadb/monitoring/overview/index.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/mariadb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mariadb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

- **KubeDB** deploys MariaDB with a `mysqld_exporter`-compatible sidecar (container `exporter`) that exposes metrics used by both MySQL and MariaDB alert charts (`mysql_*` metric names).
- **ServiceMonitor** (named `{mariadb-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape the exporter every 10 seconds.
- **PrometheusRule** is created by the `mariadb-alerts` chart and contains MariaDB alert definitions grouped by concern: database health, Galera cluster, provisioner, ops-manager, Stash backup/restore, KubeStash backup/restore, and schema manager.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).
- **Grafana** dashboards for MariaDB are covered separately — see [Grafana Dashboard](grafana-dashboard.md) rather than duplicated here.

---

## Deploy MariaDB with Monitoring Enabled

Below is the MariaDB object we are going to create — a single standalone instance with monitoring enabled. (The chart's `cluster` group only produces data if you deploy a Galera cluster via `spec.topology`; a standalone instance simply leaves that group's alert INACTIVE.)

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: mariadb-alert-demo
  namespace: alert-mariadb
spec:
  version: "12.1.2"
  deletionPolicy: WipeOut
  replicas: 3
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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/monitoring/mariadb-alert-demo.yaml
mariadb.kubedb.com/mariadb-alert-demo created
```

Wait for the database to go into `Ready` state.

```bash
$ kubectl get mariadb -n alert-mariadb mariadb-alert-demo
NAME                 VERSION   STATUS   AGE
mariadb-alert-demo    11.5.2    Ready    3m
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n alert-mariadb --selector="app.kubernetes.io/instance=mariadb-alert-demo"
NAME                         TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)             AGE
mariadb-alert-demo           ClusterIP   10.43.10.20    <none>        3306/TCP            3m
mariadb-alert-demo-pods      ClusterIP   None           <none>        3306/TCP            3m
mariadb-alert-demo-stats     ClusterIP   10.43.10.21    <none>        56790/TCP           3m
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n alert-mariadb
NAME                      AGE
mariadb-alert-demo-stats  3m
```

Verify that the `ServiceMonitor` carries the `release: prometheus` label so Prometheus discovers it.

```bash
$ kubectl get servicemonitor -n alert-mariadb mariadb-alert-demo-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Install mariadb-alerts

The `mariadb-alerts` chart creates a `PrometheusRule` resource containing all MariaDB alert definitions.

### Why the Helm release name matters

The chart derives the `PrometheusRule` name and scopes every PromQL expression (via `job="{release-name}-stats"` / `app="{release-name}"`) from the **Helm release name** — so the release name must match the MariaDB object's name (`mariadb-alert-demo`) for the rules to be correctly scoped to this instance.

The chart's default label is `release: kube-prometheus-stack`, so we must also override it at install time to match the Prometheus `ruleSelector`.

### Install

```bash
$ helm upgrade -i mariadb-alert-demo oci://ghcr.io/appscode-charts/mariadb-alerts \
    -n alert-mariadb \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus
```

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-mariadb
NAME                  AGE
mariadb-alert-demo    30s

$ kubectl get prometheusrule -n alert-mariadb mariadb-alert-demo \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Confirm Prometheus loaded the rules

Port-forward the Prometheus UI.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules` and locate the `mariadb.database`, `mariadb.cluster`, `mariadb.provisioner`, `mariadb.opsManager`, `mariadb.stash`, `mariadb.kubeStash`, and `mariadb.schemaManager` groups.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/mariadb/monitoring/mariadb-alerting-prom-rules.png" style="padding:10px">
</p>

All groups should show **OK**, confirming that Prometheus has loaded and is evaluating the MariaDB alert definitions every 30 seconds. Unlike several other `*-alerts` charts in this project, `mariadb-alerts` v2026.7.14 renders every group declared in its `values.yaml` — no missing-group gap found here.

---

## Verify End-to-End

### 1. Check the Prometheus target is UP

Open `http://localhost:9090/query?g0.expr=up%7Bnamespace%3D%22alert-mariadb%22%7D&g0.tab=1`.

<p align="center">
  <img alt="Prometheus up query — mariadb-alert-demo-0 UP" src="/docs/images/mariadb/monitoring/mariadb-alerting-prom-target.png" style="padding:10px">
</p>

The `mariadb-alert-demo-0` pod should report `up == 1` via the `mariadb-alert-demo-stats` service/job.

### 2. Confirm the MariaDB alerts are inactive

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — MariaDB groups inactive" src="/docs/images/mariadb/monitoring/mariadb-alerting-prom-alerts.png" style="padding:10px">
</p>

All rules should show **INACTIVE** on a healthy standalone instance. `GaleraReplicationLatencyTooLong` (the `cluster` group) has no data at all on a standalone instance since it depends on Galera-specific metrics — that's expected, not a bug.

### 3. Check AlertManager

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager" src="/docs/images/mariadb/monitoring/mariadb-alerting-alertmanager.png" style="padding:10px">
</p>

No alerts should be firing for the `alert-mariadb` namespace.

### 4. Grafana dashboard

See [Grafana Dashboard](grafana-dashboard.md) for how to provision and explore the MariaDB dashboards (via the `kubedb-grafana-dashboards` chart, `--set featureGates.MariaDB=true`).

---

## Simulating a Firing Alert

This section deliberately triggers `MariaDBInstanceDown` (instant, `for: 0m`) by crashing the main `mariadb` process, and observes the alert through Prometheus and AlertManager.

### 1. Crash the MariaDB process

```bash
$ kubectl exec -n alert-mariadb mariadb-alert-demo-0 -c mariadb -- sh -c '
    end=$(( $(date +%s) + 30 ));
    while [ $(date +%s) -lt $end ]; do
      pid=$(pgrep -x mariadbd | head -1);
      [ -n "$pid" ] && kill -9 "$pid" 2>/dev/null;
      sleep 1;
    done'
```

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — MariaDBInstanceDown Firing" src="/docs/images/mariadb/monitoring/mariadb-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

`MariaDBInstanceDown` (`mysql_up == 0`) should transition straight to **FIRING** since it has no `for` delay.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager — MariaDBInstanceDown Firing" src="/docs/images/mariadb/monitoring/mariadb-alerting-alertmanager-firing.png" style="padding:10px">
</p>

### 4. Restore MariaDB

Stop the loop from step 1.

```bash
$ kubectl get mariadb -n alert-mariadb mariadb-alert-demo -w
NAME                 VERSION   STATUS   AGE
mariadb-alert-demo   11.5.2    Ready    24m
```

If MariaDB does not recover on its own within a minute or two, force a clean restart: `kubectl delete pod -n alert-mariadb mariadb-alert-demo-0`.

---

## Alert Reference

All alerts are scoped to the `mariadb-alert-demo` instance in the `alert-mariadb` namespace via the PromQL label filters `job="mariadb-alert-demo-stats"` / `namespace="alert-mariadb"` (database/cluster groups), or `app="mariadb-alert-demo"` / `namespace="alert-mariadb"` (provisioner/opsManager/stash/kubeStash/schemaManager groups).

### Database Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `MariaDBInstanceDown` | critical | instant | `mysql_up == 0` on this instance. |
| `MariaDBServiceDown` | critical | instant | No replica behind the service is answering. |
| `MariaDBTooManyConnections` | warning | 2m | Connection count is high relative to `max_connections`. |
| `MariaDBHighThreadsRunning` | warning | 2m | Too many threads actively running. |
| `MariaDBSlowQueries` | warning | 2m | Slow-query count is increasing. |
| `MariaDBInnoDBLogWaits` | warning | instant | InnoDB log waits are occurring — I/O may be a bottleneck. |
| `MariaDBRestarted` | warning | instant | Uptime indicates a recent restart. |
| `MariaDBHighQPS` | critical | instant | Query rate is unusually high. |
| `MariaDBHighIncomingBytes` | critical | instant | Inbound network traffic is unusually high. |
| `MariaDBHighOutgoingBytes` | critical | instant | Outbound network traffic is unusually high. |
| `MariaDBTooManyOpenFiles` | warning | 2m | Open file count is high relative to the limit. |
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
| `KubeDBMariaDBPhaseNotReady` | critical | 1m | KubeDB marked the MariaDB resource `NotReady`. |
| `KubeDBMariaDBPhaseCritical` | warning | 15m | MariaDB is degraded but not fully unavailable. |

### OpsManager Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBMariaDBOpsRequestStatusProgressingToLong` | critical | 30m | An ops request has been running for 30+ minutes. |
| `KubeDBMariaDBOpsRequestFailed` | critical | instant | An ops request failed. |

### Stash / KubeStash Groups

Only meaningful once Stash or KubeStash backup/restore is configured.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `MariaDBStashBackupSessionFailed` / `MariaDBKubeStashBackupSessionFailed` | critical | instant | The most recent backup session failed. |
| `MariaDBStashRestoreSessionFailed` / `MariaDBKubeStashRestoreSessionFailed` | critical | instant | The most recent restore session failed. |
| `MariaDBStashNoBackupSessionForTooLong` / `MariaDBKubeStashNoBackupSessionForTooLong` | warning | instant | No recent successful backup. |
| `MariaDBStashRepositoryCorrupted` / `MariaDBKubeStashRepositoryCorrupted` | critical | 5m | Backup repository integrity check failed. |
| `MariaDBStashRepositoryStorageRunningLow` / `MariaDBKubeStashRepositoryStorageRunningLow` | warning | 5m | Backup repository storage usage is high. |
| `MariaDBStashBackupSessionPeriodTooLong` / `MariaDBKubeStashBackupSessionPeriodTooLong` | warning | instant | A backup session is taking unusually long. |
| `MariaDBStashRestoreSessionPeriodTooLong` / `MariaDBKubeStashRestoreSessionPeriodTooLong` | warning | instant | A restore session is taking unusually long. |

### SchemaManager Group

Only meaningful when using `MariaDBDatabase` schema-manager objects.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBMariaDBSchemaPendingForTooLong` | warning | 30m | A `MariaDBDatabase` object stuck `Pending`. |
| `KubeDBMariaDBSchemaInProgressForTooLong` | warning | 30m | A `MariaDBDatabase` object stuck `InProgress`. |
| `KubeDBMariaDBSchemaTerminatingForTooLong` | warning | 30m | A `MariaDBDatabase` object stuck `Terminating`. |
| `KubeDBMariaDBSchemaFailed` | warning | instant | A `MariaDBDatabase` object failed. |
| `KubeDBMariaDBSchemaExpired` | warning | instant | A `MariaDBDatabase` object expired. |

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
          mariadbTooManyConnections:
            enabled: true
            duration: "5m"
            severity: warning
      cluster:
        enabled: "none"    # disable if you don't run Galera
```

```bash
$ helm upgrade mariadb-alert-demo oci://ghcr.io/appscode-charts/mariadb-alerts \
    -n alert-mariadb \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

```bash
$ helm uninstall mariadb-alert-demo -n alert-mariadb
$ kubectl delete mariadb -n alert-mariadb mariadb-alert-demo
$ kubectl delete ns alert-mariadb
```

## Next Steps

- Monitor your MariaDB database with KubeDB using [built-in Prometheus](/docs/guides/mariadb/monitoring/builtin-prometheus/index.md).
- Monitor your MariaDB database with KubeDB using [Prometheus operator](/docs/guides/mariadb/monitoring/prometheus-operator/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
