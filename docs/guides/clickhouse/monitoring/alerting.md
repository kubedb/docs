---
title: ClickHouse Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: ch-monitoring-alerting
    name: Alerting
    parent: ch-monitoring-clickhouse
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ClickHouse Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed ClickHouse instance using the `clickhouse-alerts` Helm chart.

## Before You Begin

* Ensure you have a Kubernetes cluster and that `kubectl` is configured to communicate with it. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Install the KubeDB operator by following the steps [here](/docs/setup/README.md).

* Deploy the database in the `alert-clickhouse` namespace:

  ```bash
  $ kubectl create ns alert-clickhouse
  namespace/alert-clickhouse created
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

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/clickhouse/monitoring/overview.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/clickhouse](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/clickhouse) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

- **KubeDB** deploys ClickHouse with metrics exposed by a [JMX Exporter](https://github.com/prometheus/jmx_exporter) running as a **Java agent inside the `clickhouse` container itself** — not a separate sidecar container. KubeDB uses the JMX agent because the officially recognized ClickHouse exporter image does not yet expose metrics for the KRaft-mode versions KubeDB supports.
- **ServiceMonitor** (named `{clickhouse-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape the JMX agent's HTTP endpoint every 10 seconds.
- **PrometheusRule** is created by the `clickhouse-alerts` chart and contains ClickHouse alert definitions grouped by concern: database health, provisioner, ops-manager, and KubeStash backup/restore.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).
- **Grafana** dashboards for ClickHouse are covered separately — see [Grafana Dashboard](grafana-dashboard.md) rather than duplicated here.

---

## Deploy ClickHouse with Monitoring Enabled

Below is the ClickHouse object we are going to create — a single-node instance with monitoring enabled.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: clickhouse-alert-demo
  namespace: alert-clickhouse
spec:
  version: "24.4.1"
  replicas: 1
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

Here, `spec.monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` resource managed by the Prometheus operator, and `spec.monitor.prometheus.serviceMonitor.labels.release: prometheus` matches the Prometheus `serviceMonitorSelector` so the target is discovered automatically.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/monitoring/clickhouse-alert-demo.yaml
clickhouse.kubedb.com/clickhouse-alert-demo created
```

Wait for the database to go into `Ready` state.

```bash
$ kubectl get clickhouse -n alert-clickhouse clickhouse-alert-demo
NAME                    VERSION   STATUS   AGE
clickhouse-alert-demo   24.4.1    Ready    3m
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n alert-clickhouse --selector="app.kubernetes.io/instance=clickhouse-alert-demo"
NAME                             TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
clickhouse-alert-demo            ClusterIP   10.43.10.20    <none>        8123/TCP,9000/TCP,9009/TCP   3m
clickhouse-alert-demo-pods       ClusterIP   None           <none>        8123/TCP,9000/TCP,9009/TCP   3m
clickhouse-alert-demo-stats      ClusterIP   10.43.10.21    <none>        8001/TCP                     3m
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n alert-clickhouse
NAME                           AGE
clickhouse-alert-demo-stats    3m
```

Verify that the `ServiceMonitor` carries the `release: prometheus` label so Prometheus discovers it.

```bash
$ kubectl get servicemonitor -n alert-clickhouse clickhouse-alert-demo-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Install clickhouse-alerts

The `clickhouse-alerts` chart creates a `PrometheusRule` resource containing all ClickHouse alert definitions.

### Why the Helm release name matters

The chart derives the `PrometheusRule` name and scopes every PromQL expression (via `job="{release-name}-stats"` / `app="{release-name}"`) from the **Helm release name** — so the release name must match the ClickHouse object's name (`clickhouse-alert-demo`) for the rules to be correctly scoped to this instance.

The chart's default label is `release: kube-prometheus-stack`, so we must also override it at install time to match the Prometheus `ruleSelector`.

### Install

```bash
$ helm upgrade -i clickhouse-alert-demo oci://ghcr.io/appscode-charts/clickhouse-alerts \
    -n alert-clickhouse \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus
```

| Flag | Value | Purpose |
|------|-------|---------|
| `clickhouse-alert-demo` (release name) | — | Scopes every PromQL expression to this instance (`job="clickhouse-alert-demo-stats"`, `app="clickhouse-alert-demo"`) |
| `-n alert-clickhouse` | `alert-clickhouse` | Installs the `PrometheusRule` in the same namespace as the database |
| `form.alert.labels.release` | `prometheus` | Matches the Prometheus `ruleSelector` so the rules are loaded |

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-clickhouse
NAME                     AGE
clickhouse-alert-demo    30s
```

Confirm the `release: prometheus` label is present.

```bash
$ kubectl get prometheusrule -n alert-clickhouse clickhouse-alert-demo \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Confirm Prometheus loaded the rules

Port-forward the Prometheus UI.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules` and locate the `clickhouse.database`, `clickhouse.provisioner`, `clickhouse.opsManager`, and `clickhouse.kubeStash` groups.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/clickhouse/monitoring/clickhouse-alerting-prom-rules.png" style="padding:10px">
</p>

All groups should show **OK**, confirming that Prometheus has loaded and is evaluating the ClickHouse alert definitions every 30 seconds.

---

## Verify End-to-End

### 1. Check the Prometheus target is UP

Prometheus discovers more than 20 scrape pools on a shared cluster, so instead of the Target health page, query `up` directly for a reliable view.

Open `http://localhost:9090/query?g0.expr=up%7Bnamespace%3D%22alert-clickhouse%22%7D&g0.tab=1`.

<p align="center">
  <img alt="Prometheus up query — clickhouse-alert-demo-0 UP" src="/docs/images/clickhouse/monitoring/clickhouse-alerting-prom-target.png" style="padding:10px">
</p>

The `clickhouse-alert-demo-0` pod should report `up == 1` via the `clickhouse-alert-demo-stats` service/job, confirming Prometheus is scraping the JMX agent successfully.

### 2. Confirm the ClickHouse alerts are inactive

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — ClickHouse groups inactive" src="/docs/images/clickhouse/monitoring/clickhouse-alerting-prom-alerts.png" style="padding:10px">
</p>

All rules across the `clickhouse.database`, `clickhouse.provisioner`, `clickhouse.opsManager`, and `clickhouse.kubeStash` groups should show **INACTIVE**, confirming the instance is healthy and no thresholds are breached. (`clickhouse.kubeStash` rules stay INACTIVE with no data unless KubeStash backups are configured on this instance.)

### 3. Check AlertManager

Port-forward AlertManager to view any currently firing alerts.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager" src="/docs/images/clickhouse/monitoring/clickhouse-alerting-alertmanager.png" style="padding:10px">
</p>

No alerts should be firing for the `alert-clickhouse` namespace.

### 4. Grafana dashboard

Grafana dashboards for ClickHouse are documented separately rather than duplicated in this alerting guide — see [Grafana Dashboard](grafana-dashboard.md) for how to provision and explore the ClickHouse dashboards (via the `kubedb-grafana-dashboards` chart, `--set featureGates.ClickHouse=true`).

---

## Simulating a Firing Alert

This section deliberately triggers `KubeDBClickHousePhaseNotReady` so you can observe the full alert lifecycle — from firing in Prometheus through to the AlertManager dashboard — and then resolve it.

Since the JMX exporter runs as a Java agent inside the `clickhouse` process itself rather than a separate sidecar, crashing the ClickHouse process takes the metrics endpoint down with it. Find the main ClickHouse process and crash-loop it long enough for the KubeDB operator to mark the resource `NotReady` and hold it there past the alert's `for: 1m` window.

### 1. Crash the ClickHouse process repeatedly

```bash
$ kubectl exec -n alert-clickhouse clickhouse-alert-demo-0 -c clickhouse -- sh -c '
    end=$(( $(date +%s) + 90 ));
    while [ $(date +%s) -lt $end ]; do
      pid=$(pgrep -f clickhouse-server | head -1);
      [ -n "$pid" ] && kill -9 "$pid" 2>/dev/null;
      sleep 1;
    done'
```

> If the container's entrypoint respawns `clickhouse-server` faster than the loop above can catch it, run the loop directly inside a single `exec` session (as above, not from repeated external `kubectl exec` calls) to avoid round-trip latency — see the general technique notes in the pgbouncer/pgpool alerting tutorials for why this matters.

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — KubeDBClickHousePhaseNotReady Firing" src="/docs/images/clickhouse/monitoring/clickhouse-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

`KubeDBClickHousePhaseNotReady` should transition from **INACTIVE** to **FIRING** once `kubedb_com_clickhouse_status_phase{phase="NotReady"}` has read `1` continuously for the full `for: 1m` duration.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager — KubeDBClickHousePhaseNotReady Firing" src="/docs/images/clickhouse/monitoring/clickhouse-alerting-alertmanager-firing.png" style="padding:10px">
</p>

AlertManager should show the `KubeDBClickHousePhaseNotReady` alert with **Severity: critical**.

### 4. Restore ClickHouse

Stop the loop from step 1 and give the operator a few reconcile cycles to mark the resource `Ready` again.

```bash
$ kubectl get clickhouse -n alert-clickhouse clickhouse-alert-demo -w
NAME                    VERSION   STATUS   AGE
clickhouse-alert-demo   24.4.1    Ready    24m
```

If the pod does not recover cleanly, force a clean restart with `kubectl delete pod -n alert-clickhouse clickhouse-alert-demo-0`.

---

## Alert Reference

All alerts are scoped to the `clickhouse-alert-demo` instance in the `alert-clickhouse` namespace via the PromQL label filters `job="clickhouse-alert-demo-stats"` / `namespace="alert-clickhouse"` (database group), or `app="clickhouse-alert-demo"` / `namespace="alert-clickhouse"` (provisioner/opsManager/kubeStash groups).

### Database Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `ClickhouseInstanceDown` | critical | 1m | The JMX agent's scrape target reports `up == 0`. |
| `ClickhouseTooManyConnections` | critical | 5m | Too many concurrent TCP connections. |
| `ClickhouseTooManyActiveQueries` | warning | 1m | Too many queries running concurrently. |
| `ClickhouseReplicationPartFetchFailed` | warning | 5m | Replicated part fetches are failing. |
| `ClickhouseBrokenPartsDetected` | critical | instant | Too many unexpected data parts detected. |
| `ClickhouseDataPartCorrupted` | warning | 5m | A data part failed a parse/assertion check. |
| `DiskUsageHigh` | warning | 5m | Persistent volume usage exceeds the configured threshold. |
| `DiskAlmostFull` | critical | 5m | Persistent volume usage is critically high. |

### Provisioner Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBClickHousePhaseNotReady` | critical | 1m | KubeDB marked the ClickHouse resource `NotReady`. |
| `KubeDBClickHousePhaseCritical` | warning | 15m | ClickHouse is degraded but not fully unavailable. |

### OpsManager Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBClickHouseOpsRequestStatusProgressingToLong` | critical | 30m | An ops request has been running for 30+ minutes. |
| `KubeDBClickHouseOpsRequestFailed` | critical | instant | An ops request failed. |

### KubeStash Group

Only meaningful once KubeStash backup/restore is configured for this instance.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `ClickHouseKubeStashBackupSessionFailed` | critical | instant | The most recent backup session failed. |
| `ClickHouseKubeStashRestoreSessionFailed` | critical | instant | The most recent restore session failed. |
| `ClickHouseKubeStashNoBackupSessionForTooLong` | warning | instant | No successful backup recorded recently. |
| `ClickHouseKubeStashRepositoryCorrupted` | critical | 5m | Backup repository integrity check failed. |
| `ClickHouseKubeStashRepositoryStorageRunningLow` | warning | 5m | Backup repository storage usage is high. |
| `ClickHouseKubeStashBackupSessionPeriodTooLong` | warning | instant | A backup session is taking unusually long. |
| `ClickHouseKubeStashRestoreSessionPeriodTooLong` | warning | instant | A restore session is taking unusually long. |

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
          clickhouseTooManyConnections:
            enabled: true
            duration: "10m"
            severity: warning
      kubeStash:
        enabled: "none"    # disable if you don't use KubeStash
```

```bash
$ helm upgrade clickhouse-alert-demo oci://ghcr.io/appscode-charts/clickhouse-alerts \
    -n alert-clickhouse \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

To remove all resources created in this tutorial, run the following commands.

```bash
# Remove the clickhouse-alerts release (PrometheusRule)
$ helm uninstall clickhouse-alert-demo -n alert-clickhouse

# Remove the ClickHouse instance
$ kubectl delete clickhouse -n alert-clickhouse clickhouse-alert-demo

# Delete namespace
$ kubectl delete ns alert-clickhouse
```

## Next Steps

- Monitor your ClickHouse instance with KubeDB using [built-in Prometheus](/docs/guides/clickhouse/monitoring/using-builtin-prometheus.md).
- Monitor your ClickHouse instance with KubeDB using [Prometheus operator](/docs/guides/clickhouse/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
