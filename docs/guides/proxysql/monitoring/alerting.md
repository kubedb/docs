---
title: ProxySQL Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-monitoring-alerting
    name: Alerting
    parent: guides-proxysql-monitoring
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ProxySQL Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed ProxySQL instance using the `proxysql-alerts` Helm chart.

## Before You Begin

* Ensure you have a Kubernetes cluster and that `kubectl` is configured to communicate with it. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Install the KubeDB operator by following the steps [here](/docs/setup/README.md).

* Deploy the database in the `alert-proxysql` namespace:

  ```bash
  $ kubectl create ns alert-proxysql
  namespace/alert-proxysql created
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

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/proxysql/monitoring/overview/index.md).

* ProxySQL is a proxy layer in front of a MySQL backend, so this tutorial first deploys a 3-member MySQL Group Replication cluster, then a ProxySQL instance pointed at it. Both objects are deployed with monitoring enabled.

> Note: YAML files used in this tutorial are stored in [docs/examples/proxysql](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/proxysql) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

- **KubeDB** deploys ProxySQL with metrics exposed directly by the `proxysql` container itself on port `6070` (an embedded `proxysql_exporter`-style endpoint) — there is no separate exporter sidecar.
- **ServiceMonitor** (named `{proxysql-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape the metrics endpoint every 10 seconds.
- **PrometheusRule** is created by the `proxysql-alerts` chart and contains ProxySQL alert definitions grouped by concern: database health, cluster sync, provisioner, and ops-manager.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).
- **Grafana** dashboards for ProxySQL are covered separately — see [Grafana Dashboard](grafana-dashboard.md) rather than duplicated here.

---

## Deploy the MySQL Backend

ProxySQL routes traffic to a MySQL cluster, so deploy the backend first. Below is the MySQL object we are going to create — a 3-member Group Replication cluster on the `longhorn` StorageClass.

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: my-group-alert
  namespace: alert-proxysql
spec:
  version: "9.1.0"
  replicas: 3
  topology:
    mode: GroupReplication
    group:
      name: "dc002fc3-c412-4d18-b1d4-66c1fbfbbc9b"
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/proxysql/monitoring/my-group-alert.yaml
mysql.kubedb.com/my-group-alert created
```

Wait for the MySQL cluster to go into `Ready` state, and confirm all three PVCs bind on `longhorn`.

```bash
$ kubectl get mysql -n alert-proxysql my-group-alert
NAME             VERSION   STATUS   AGE
my-group-alert   9.1.0     Ready    5m

$ kubectl get pvc -n alert-proxysql
NAME                    STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-my-group-alert-0   Bound    pvc-9d47405f-06ec-4c79-ab93-b1588451896a   1Gi        RWO            longhorn       5m
data-my-group-alert-1   Bound    pvc-a3fec212-fff1-489d-b9d0-c8daf143f8fc   1Gi        RWO            longhorn       2m
data-my-group-alert-2   Bound    pvc-a3fd6857-70bf-4a1b-a20a-2997379d8678   1Gi        RWO            longhorn       2m
```

> **Note:** `storageClassName` is immutable on a PVC. If you need to move an existing MySQL/ProxySQL instance from one StorageClass to another (e.g. `local-path` → `longhorn`), you cannot edit the field in place — delete the database object (and its PVCs, if `deletionPolicy` is not `WipeOut`) and recreate it pointing at the new StorageClass.

## Deploy ProxySQL with Monitoring Enabled

Now deploy ProxySQL, pointing `spec.backend.name` at the MySQL cluster above.

```yaml
apiVersion: kubedb.com/v1
kind: ProxySQL
metadata:
  name: proxysql-alert
  namespace: alert-proxysql
spec:
  version: "2.3.2-debian"
  replicas: 1
  backend:
    name: my-group-alert
  deletionPolicy: WipeOut
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 42004
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

Here,

- `spec.backend.name: my-group-alert` tells ProxySQL which KubeDB MySQL object to load-balance traffic to.
- `spec.monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` resource managed by the Prometheus operator.
- `spec.monitor.prometheus.serviceMonitor.labels.release: prometheus` adds the `release: prometheus` label to the created `ServiceMonitor`, matching the Prometheus `serviceMonitorSelector` so the target is discovered automatically.

> ProxySQL itself does not provision its own PVC — it has no `spec.storage` field, so there is nothing to migrate to `longhorn` beyond the MySQL backend's PVCs above.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/proxysql/monitoring/proxysql-alert.yaml
proxysql.kubedb.com/proxysql-alert created
```

Now, wait for ProxySQL to go into `Ready` state.

```bash
$ kubectl get proxysql -n alert-proxysql proxysql-alert
NAME             VERSION        STATUS   AGE
proxysql-alert   2.3.2-debian   Ready    40s
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n alert-proxysql --selector="app.kubernetes.io/instance=proxysql-alert"
NAME                    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
proxysql-alert          ClusterIP   10.43.180.207   <none>        6033/TCP            40s
proxysql-alert-pods     ClusterIP   None            <none>        6032/TCP,6033/TCP   40s
proxysql-alert-stats    ClusterIP   10.43.61.242    <none>        6070/TCP            40s
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n alert-proxysql
NAME                    AGE
proxysql-alert-stats    40s
```

Verify that the `ServiceMonitor` carries the `release: prometheus` label so Prometheus discovers it.

```bash
$ kubectl get servicemonitor -n alert-proxysql proxysql-alert-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Install proxysql-alerts

The `proxysql-alerts` chart creates a `PrometheusRule` resource containing ProxySQL alert definitions grouped by concern.

### Why the Helm release name matters

The chart derives the `PrometheusRule` name and scopes every PromQL expression (via `job="{release-name}-stats"` / `app="{release-name}"`) from the **Helm release name** — so the release name must match the ProxySQL object's name (`proxysql-alert`) for the rules to be correctly scoped to this instance.

The chart's default label is `release: kube-prometheus-stack`, so we must also override it at install time to match the Prometheus `ruleSelector`.

### Install

```bash
$ helm upgrade -i proxysql-alert oci://ghcr.io/appscode-charts/proxysql-alerts \
    -n alert-proxysql \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus
```

| Flag | Value | Purpose |
|------|-------|---------|
| `proxysql-alert` (release name) | — | Scopes every PromQL expression to this instance (`job="proxysql-alert-stats"`, `app="proxysql-alert"`) |
| `-n alert-proxysql` | `alert-proxysql` | Installs the `PrometheusRule` in the same namespace as the database |
| `form.alert.labels.release` | `prometheus` | Matches the Prometheus `ruleSelector` so the rules are loaded |

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-proxysql
NAME             AGE
proxysql-alert   30s
```

Confirm the `release: prometheus` label is present.

```bash
$ kubectl get prometheusrule -n alert-proxysql proxysql-alert \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Confirm Prometheus loaded the rules

Port-forward the Prometheus UI.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules` and filter by **proxysql**.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/proxysql/monitoring/proxysql-alerting-prom-rules.png" style="padding:10px">
</p>

Three rule groups are visible — `proxysql.database`, `proxysql.opsManager`, and `proxysql.provisioner` — all showing **OK**, confirming Prometheus has loaded and is evaluating the ProxySQL alert definitions every 30 seconds.

> **Known chart bug (v2026.7.14):** The chart's own `values.yaml` declares nine alerts under the `database` group (`ProxySQLInstanceDown`, `ProxySQLServiceDown`, `ProxySQLTooManyConnections`, `ProxySQLHighThreadsRunning`, `ProxySQLSlowQueries`, `ProxySQLRestarted`, `ProxySQLHighQPS`, `ProxySQLHighIncomingBytes`, `ProxySQLHighOutgoingBytes`) plus a separate `cluster` group holding `ProxySQLCLusterSyncFailure`. Running `helm template`/`helm get manifest` against this same chart version now renders all of that correctly. However, the `PrometheusRule` actually installed live in this cluster contains only **one** rule under the group named `proxysql.database...` — and it is in fact the `cluster` group's `ProxySQLCLusterSyncFailure` rule content, mislabeled under the `database` group name, with the real database-health rules (instance/service down, connections, QPS, threads, slow queries, bytes) entirely absent, and no independent `proxysql.cluster...` group at all. Since the Helm release has never been upgraded (`generation: 1`, single revision) and `managedFields` shows only `helm` ever wrote the object, this indicates the chart's OCI artifact at tag `v2026.7.14` changed content after this release was first installed. If you need the full rule set, re-run the `helm upgrade` command above to reconcile the live object with the current chart content — the value of doing so is that you gain the missing `database`-group health alerts, at the cost of a brief `PrometheusRule` re-apply (no database downtime).

---

## Verify End-to-End

### 1. Check the metrics endpoint

The `proxysql` container serves its own Prometheus metrics at `:6070/metrics` — no exporter sidecar is involved.

```bash
$ kubectl exec -n alert-proxysql proxysql-alert-0 -c proxysql -- \
                                      wget -qO- localhost:6070/metrics | grep proxysql_servers_table_version_total
# HELP proxysql_servers_table_version_total Number of times the "servers_table" have been modified.
# TYPE proxysql_servers_table_version_total counter
proxysql_servers_table_version_total 16.000000
```

### 2. Check the Prometheus target is UP

Prometheus discovers more than 20 scrape pools on a shared cluster, so instead of the Target health page, query `up` directly for a reliable view.

Open `http://localhost:9090/query?g0.expr=up%7Bnamespace%3D%22alert-proxysql%22%7D&g0.tab=1`.

<p align="center">
  <img alt="Prometheus up query — proxysql-alert-0 UP" src="/docs/images/proxysql/monitoring/proxysql-alerting-prom-target.png" style="padding:10px">
</p>

The `proxysql-alert-0` pod reports `up == 1` via the `proxysql-alert-stats` service/job, confirming Prometheus is scraping it successfully.

### 3. Confirm the ProxySQL alerts are inactive

Open `http://localhost:9090/alerts` and filter by **proxysql**.

<p align="center">
  <img alt="Prometheus Alerts — ProxySQL groups inactive" src="/docs/images/proxysql/monitoring/proxysql-alerting-prom-alerts.png" style="padding:10px">
</p>

All 5 currently-loaded rules across the three groups show **INACTIVE**, confirming the cluster is healthy and no thresholds are breached (see the chart-bug callout above for why only 5 of the chart's defined rules are loaded).

### 4. Check AlertManager

Port-forward AlertManager to view any currently firing alerts.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager" src="/docs/images/proxysql/monitoring/proxysql-alerting-alertmanager.png" style="padding:10px">
</p>

No alerts are firing for the `alert-proxysql` namespace.

### 5. Grafana dashboard

Grafana dashboards for ProxySQL are documented separately rather than duplicated in this alerting guide — see [Grafana Dashboard](grafana-dashboard.md) for how to provision and explore the ProxySQL dashboards (via the `kubedb-grafana-dashboards` chart, `--set featureGates.ProxySQL=true`).

---

## Simulating a Firing Alert

The previous section showed that all currently-loaded ProxySQL alerts are **INACTIVE** while the instance is healthy. This section deliberately triggers the `KubeDBProxySQLPhaseNotReady` critical alert so you can observe the full alert lifecycle — from firing in Prometheus through to the AlertManager dashboard — and then resolve it.

ProxySQL runs as a **single container per pod** — there is no separate exporter sidecar. The container has neither `ps` nor `pgrep`, only `bash`/`sh`, so identify the actual `proxysql` process via `/proc` and kill it directly (rather than the container's PID 1, which is `tini`). Because `KubeDBProxySQLPhaseNotReady` requires the condition to persist for `for: 1m`, a single `kill` is not enough — keep the process crashing long enough for the KubeDB operator to mark the resource `NotReady` and hold it there past the one-minute window.

### 1. Crash the ProxySQL process repeatedly

```bash
$ while true; do
    kubectl exec -n alert-proxysql proxysql-alert-0 -c proxysql -- bash -c '
      for p in /proc/[0-9]*; do
        pid=$(basename "$p")
        cmd=$(tr "\0" " " < "$p/cmdline" 2>/dev/null)
        case "$cmd" in
          proxysql\ -c*) kill -9 "$pid" ;;
        esac
      done
    ' >/dev/null 2>&1
    sleep 3
  done
```

Let this loop run for a couple of minutes (leave it running while you check the next steps), then stop it once you've captured the firing state.

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts` filtered by **proxysql**.

<p align="center">
  <img alt="Prometheus Alerts — KubeDBProxySQLPhaseNotReady Firing" src="/docs/images/proxysql/monitoring/proxysql-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

`KubeDBProxySQLPhaseNotReady` transitions from **INACTIVE** to **FIRING** once `kubedb_com_proxysql_status_phase{phase="NotReady"}` has read `1` continuously for the full `for: 1m` duration — this metric comes from the KubeDB operator's own view of the resource (exported via Panopticon), not from the ProxySQL metrics endpoint itself.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager — KubeDBProxySQLPhaseNotReady Firing" src="/docs/images/proxysql/monitoring/proxysql-alerting-alertmanager-firing.png" style="padding:10px">
</p>

AlertManager shows the `KubeDBProxySQLPhaseNotReady` alert. The alert card displays:

- **Severity**: `critical`
- **proxysql**: `proxysql-alert` in the `alert-proxysql` namespace
- **phase**: `NotReady`
- **Started**: timestamp when the alert first fired

AlertManager routes this alert to every receiver configured in your `alertmanagerConfig` (Slack, email, PagerDuty, webhook, etc.) based on your routing tree. If no receiver is configured, the alert is visible here but silently dropped.

### 4. Restore ProxySQL

Stop the loop from step 1.

> **Note:** Unlike some other KubeDB databases, the ProxySQL image's entrypoint script does **not** automatically respawn the `proxysql` process after it is repeatedly `kill -9`'d — the wrapper script exits instead of retrying, and no liveness probe recovers it, so the pod can remain `Running` (`READY 1/1`) while the daemon inside is actually dead. If ProxySQL does not return to `Ready` on its own within a minute or two of stopping the loop, force a clean restart:
>
> ```bash
> $ kubectl delete pod -n alert-proxysql proxysql-alert-0
> pod "proxysql-alert-0" deleted
> ```
>
> The PetSet controller recreates the pod immediately.

```bash
$ kubectl get proxysql -n alert-proxysql proxysql-alert -w
NAME             VERSION        STATUS   AGE
proxysql-alert   2.3.2-debian   Ready    24m
```

Once the phase returns to `Ready`, Prometheus marks the alert **INACTIVE** again and AlertManager sends a **resolved** notification to all receivers.

---

## Alert Reference

All alerts are scoped to the `proxysql-alert` instance in the `alert-proxysql` namespace via the PromQL label filters `job="proxysql-alert-stats"` / `namespace="alert-proxysql"` (database/cluster groups), or `app="proxysql-alert"` / `namespace="alert-proxysql"` (provisioner/opsManager groups).

The tables below list every alert **defined by the chart's `values.yaml`**. As documented in the chart-bug callout above, only `ProxySQLCLusterSyncFailure`, `KubeDBProxySQLPhaseNotReady`, `KubeDBProxySQLPhaseCritical`, `KubeDBProxySQLOpsRequestStatusProgressingToLong`, and `KubeDBProxySQLOpsRequestFailed` are actually loaded in this cluster's live `PrometheusRule`; the rest of the `Database Group` table describes what `helm template` renders for this chart version but is **not currently active** until the release is upgraded/reconciled.

### Database Group

Fired based on live metrics from the ProxySQL container's built-in metrics endpoint.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `ProxySQLInstanceDown` | critical | instant | `proxysql_uptime_seconds_total` reads `0` — the ProxySQL process is down. |
| `ProxySQLServiceDown` | critical | instant | The summed uptime across the service is `0` — no ProxySQL replica is answering. |
| `ProxySQLTooManyConnections` | warning | 2m | Client connections exceed 80% of `proxysql_mysql_max_connections`. |
| `ProxySQLHighThreadsRunning` | warning | 2m | More than 60 worker threads are running — the proxy may be saturated. |
| `ProxySQLSlowQueries` | warning | 2m | `proxysql_slow_queries_total` increased in the last minute. |
| `ProxySQLRestarted` | warning | instant | `proxysql_uptime_seconds_total` is below 60s — the process restarted recently. |
| `ProxySQLHighQPS` | critical | instant | Query rate exceeds 1000 QPS. |
| `ProxySQLHighIncomingBytes` | critical | instant | Frontend-received byte rate exceeds 1 MB/s. |
| `ProxySQLHighOutgoingBytes` | critical | instant | Frontend-sent byte rate exceeds 1 MB/s. |

### Cluster Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `ProxySQLCLusterSyncFailure` | warning | 5m | `proxysql_cluster_syn_conflict_total` rate exceeds `0.1/s` — ProxySQL cluster nodes are failing to sync config. **This is the only rule currently loaded live in this cluster, under a group mislabeled `proxysql.database...` — see the chart-bug callout above.** |

### Provisioner Group

Monitors the KubeDB operator's view of the ProxySQL resource phase (sourced from Panopticon, not the ProxySQL metrics endpoint). **Loaded live.**

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBProxySQLPhaseNotReady` | critical | 1m | KubeDB marked the ProxySQL resource `NotReady` — operator cannot reach a healthy instance. |
| `KubeDBProxySQLPhaseCritical` | warning | 15m | ProxySQL is degraded but not fully unavailable. |

### OpsManager Group

Tracks `ProxySQLOpsRequest` lifecycle during upgrades, scaling, and reconfiguration. **Loaded live** (except `opsRequestOnProgress`, whose `info` severity is filtered out by the chart's `enabled: warning` group gate).

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBProxySQLOpsRequestStatusProgressingToLong` | critical | 30m | An ops request has been running for 30+ minutes — likely stuck. |
| `KubeDBProxySQLOpsRequestFailed` | critical | instant | An ops request failed — check the `ProxySQLOpsRequest` object for the error. |

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
          proxysqlTooManyConnections:
            enabled: true
            duration: "5m"
            val: 90        # fire at 90% instead of the default 80%
            severity: warning
      cluster:
        enabled: "none"    # disable the cluster-sync alert
```

```bash
$ helm upgrade proxysql-alert oci://ghcr.io/appscode-charts/proxysql-alerts \
    -n alert-proxysql \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

> Since the currently-installed release predates the chart's present content (see the chart-bug callout above), any `helm upgrade` — customised or not — will also pull in the full, correctly-split `database`/`cluster` rule groups.

---

## Cleaning up

To remove all resources created in this tutorial, run the following commands.

```bash
# Remove the proxysql-alerts release (PrometheusRule)
$ helm uninstall proxysql-alert -n alert-proxysql

# Remove the ProxySQL instance
$ kubectl delete proxysql -n alert-proxysql proxysql-alert

# Remove the MySQL backend
$ kubectl delete mysql -n alert-proxysql my-group-alert

# Delete namespace
$ kubectl delete ns alert-proxysql
```

## Next Steps

- Monitor your ProxySQL instance with KubeDB using [built-in Prometheus](/docs/guides/proxysql/monitoring/builtin-prometheus/index.md).
- Monitor your ProxySQL instance with KubeDB using [Prometheus operator](/docs/guides/proxysql/monitoring/prometheus-operator/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
