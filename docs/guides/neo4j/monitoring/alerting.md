---
title: Neo4j Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: neo4j-monitoring-alerting
    name: Alerting
    parent: neo4j-monitoring
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Neo4j Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed Neo4j cluster using the `neo4j-alerts` Helm chart. Unlike most other `*-alerts` charts, `neo4j-alerts` also bundles a Grafana dashboard that it imports automatically through a post-install Job — no separate dashboard chart is required.

## Before You Begin

* Ensure you have a Kubernetes cluster and that `kubectl` is configured to communicate with it. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Install the KubeDB operator by following the steps [here](/docs/setup/README.md).

* Deploy the database in the `alert-neo4j` namespace:

  ```bash
  $ kubectl create ns alert-neo4j
  namespace/alert-neo4j created
  ```

* This tutorial assumes you already have a **kube-prometheus-stack** running in your cluster, with `Prometheus` configured so that both `serviceMonitorSelector` and `ruleSelector` match the label `release: prometheus`.

  To verify the selectors:

  ```bash
  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.ruleSelector}'
  {"matchLabels":{"release":"prometheus"}}

  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.serviceMonitorSelector}'
  {"matchLabels":{"release":"prometheus"}}
  ```

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/neo4j/monitoring/overview.md).

* You will also need a Grafana API key / token with **Editor** permission so the chart's dashboard-import Job can push the dashboard. See [Step 2](#create-a-grafana-api-key) below.

> Note: YAML files used in this tutorial are stored in [docs/examples/neo4j](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/neo4j) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

- **KubeDB** deploys Neo4j with metrics exposed directly by the `neo4j` container itself on port `2004` (Neo4j's built-in Prometheus metrics endpoint) — there is no separate exporter sidecar.
- **ServiceMonitor** (named `{neo4j-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape the metrics endpoint every 10 seconds.
- **PrometheusRule** is created by the `neo4j-alerts` chart and contains all Neo4j alert definitions grouped by concern: database health/resource usage and provisioner.
- **Dashboard-import Job** — when `grafana.enabled` is `true` (the default), the chart also creates a one-shot `Job` that `POST`s a bundled dashboard JSON straight to your Grafana instance's `/api/dashboards/import` endpoint.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).

---

## Deploy Neo4j with Monitoring Enabled

At first, let's deploy a 3-node Neo4j cluster with monitoring enabled. Below is the Neo4j object we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-alert-demo
  namespace: alert-neo4j
spec:
  replicas: 3
  version: "2025.12.1"
  deletionPolicy: WipeOut
  storage:
    storageClassName: "local-path"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        interval: 10s
        labels:
          release: prometheus
```

Here,

- `spec.replicas: 3` creates a 3-member Neo4j cluster. Neo4j Enterprise clustering requires a minimum of 3 core members.
- `spec.monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` resource managed by the Prometheus operator.
- `spec.monitor.prometheus.serviceMonitor.labels.release: prometheus` adds the `release: prometheus` label to the created `ServiceMonitor`, matching the Prometheus `serviceMonitorSelector` so the target is discovered automatically.

Let's create the namespace and the Neo4j resource.

```bash
$ kubectl create ns alert-neo4j
namespace/alert-neo4j created

$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/monitoring/neo4j-alert-demo.yaml
neo4j.kubedb.com/neo4j-alert-demo created
```

Now, wait for the database to go into `Ready` state.

```bash
$ kubectl get neo4j -n alert-neo4j neo4j-alert-demo
NAME               VERSION     STATUS   AGE
neo4j-alert-demo   2025.12.1   Ready    3m
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n alert-neo4j --selector="app.kubernetes.io/instance=neo4j-alert-demo"
NAME                     TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                                                 AGE
neo4j-alert-demo         ClusterIP   10.43.30.142    <none>        6362/TCP,7687/TCP,7474/TCP                              3m
neo4j-alert-demo-0       ClusterIP   None            <none>        6362/TCP,7687/TCP,7474/TCP,7688/TCP,7000/TCP,6000/TCP   3m
neo4j-alert-demo-1       ClusterIP   None            <none>        6362/TCP,7687/TCP,7474/TCP,7688/TCP,7000/TCP,6000/TCP   3m
neo4j-alert-demo-2       ClusterIP   None            <none>        6362/TCP,7687/TCP,7474/TCP,7688/TCP,7000/TCP,6000/TCP   3m
neo4j-alert-demo-stats   ClusterIP   10.43.63.217    <none>        2004/TCP                                                3m
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n alert-neo4j
NAME                     AGE
neo4j-alert-demo-stats   3m
```

Verify that the `ServiceMonitor` carries the `release: prometheus` label so Prometheus discovers it.

```bash
$ kubectl get servicemonitor -n alert-neo4j neo4j-alert-demo-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Create a Grafana API Key

The chart's dashboard-import Job authenticates to Grafana with a bearer token, so create one first.

* **Grafana 9+**: **Administration → Service accounts → Add service account** → role **Editor** → **Add token**. Copy the token.
* **Grafana 8.x and earlier** (no Service Accounts UI, e.g. the bundled `kube-prometheus-stack` Grafana 7.5.5 used while verifying this tutorial): use the legacy **API Keys** endpoint instead:

  ```bash
  # Port-forward Grafana
  $ kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80

  # Retrieve the admin password
  $ kubectl get secret -n monitoring prometheus-grafana \
      -o jsonpath='{.data.admin-password}' | base64 -d && echo

  # Create a legacy API key with Editor role
  $ curl -s -X POST -H "Content-Type: application/json" \
      -u admin:<grafana-admin-password> \
      http://localhost:3000/api/auth/keys \
      -d '{"name":"neo4j-alerts-demo-key","role":"Editor"}'
  # Note the returned "key"

  $ kill %1
  ```

Either way, you end up with a bearer token to use as `grafana.apikey` below.

## Step 2 — Install neo4j-alerts

The `neo4j-alerts` chart creates a `PrometheusRule` resource containing all Neo4j alert definitions, **and** (by default) a `Job` that imports a pre-built Grafana dashboard.

### Why the Helm release name matters

The chart derives the PromQL `pod`/`container` scoping (via `.Release.Name`/`.Release.Namespace`) and the `PrometheusRule` name from the **Helm release name**, not from a values field — so the release name must match the Neo4j object's name (`neo4j-alert-demo`) for the rules to be correctly scoped to this instance.

The chart's default label is `release: kube-prometheus-stack`, so we must also override it at install time to match the Prometheus `ruleSelector`.

### Install

```bash
$ helm upgrade -i neo4j-alert-demo appscode/neo4j-alerts \
    -n alert-neo4j \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus \
    --set grafana.enabled=true \
    --set grafana.url="http://prometheus-grafana.monitoring.svc:80" \
    --set grafana.apikey="<grafana-token-from-step-1>"
```

| Flag | Value | Purpose |
|------|-------|---------|
| `neo4j-alert-demo` (release name) | — | Scopes every PromQL expression to this instance (`pod=~"neo4j-alert-demo-.+$"`) |
| `-n alert-neo4j` | `alert-neo4j` | Installs the `PrometheusRule` in the same namespace as the database |
| `form.alert.labels.release` | `prometheus` | Matches the Prometheus `ruleSelector` so the rules are loaded |
| `grafana.url` | in-cluster Grafana URL | The dashboard-import Job runs **inside the cluster**, so this must be a cluster-internal address, not `localhost` |
| `grafana.apikey` | token from Step 1 | Authenticates the dashboard-import `POST` request |

> To install **alerts only, without the dashboard**, set `--set grafana.enabled=false`.

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-neo4j
NAME               AGE
neo4j-alert-demo   30s
```

Confirm the `release: prometheus` label is present.

```bash
$ kubectl get prometheusrule -n alert-neo4j neo4j-alert-demo \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Verify the dashboard-import Job

```bash
$ kubectl get job -n alert-neo4j
NAME                        STATUS     COMPLETIONS   AGE
neo4j-alert-demo-post-job   Complete   1/1           17s

$ kubectl logs -n alert-neo4j job/neo4j-alert-demo-post-job
{"pluginId":"","title":"kubedb.com / Neo4j / alert-neo4j / neo4j-alert-demo","imported":true, ...}
```

A `"imported":true` response confirms the dashboard `kubedb.com / Neo4j / alert-neo4j / neo4j-alert-demo` now exists in Grafana.

### Confirm Prometheus loaded the rules

Port-forward the Prometheus UI.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules` and locate the `neo4j.database` and `neo4j.provisioner` groups.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/neo4j/monitoring/neo4j-alerting-prom-rules.png" style="padding:10px">
</p>

Both groups are visible with all 10 rules showing **OK**, confirming that Prometheus has loaded and is evaluating the Neo4j alert definitions every 30 seconds.

---

## Verify End-to-End

### 1. Check the metrics endpoint

The `neo4j` container serves its own Prometheus metrics at `:2004/metrics` — no exporter sidecar is involved.

```bash
$ kubectl exec -n alert-neo4j neo4j-alert-demo-0 -c neo4j -- \
    wget -qO- localhost:2004/metrics | grep neo4j_dbms_page_cache_hit_ratio
# HELP neo4j_dbms_page_cache_hit_ratio Generated from Dropwizard metric import ...
# TYPE neo4j_dbms_page_cache_hit_ratio gauge
neo4j_dbms_page_cache_hit_ratio 1.0
```

### 2. Check the Prometheus target is UP

Prometheus discovers more than 20 scrape pools on a shared cluster, so instead of the Target health page, query `up` directly for a reliable view.

Open `http://localhost:9090/query?g0.expr=up%7Bnamespace%3D%22alert-neo4j%22%7D&g0.tab=1`.

<p align="center">
  <img alt="Prometheus up query — all 3 pods UP" src="/docs/images/neo4j/monitoring/neo4j-alerting-prom-target.png" style="padding:10px">
</p>

All three `neo4j-alert-demo-{0,1,2}` pods report `up == 1`, confirming Prometheus is scraping every pod in the cluster.

### 3. Confirm the Neo4j alerts are inactive

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — Neo4j groups inactive" src="/docs/images/neo4j/monitoring/neo4j-alerting-prom-alerts.png" style="padding:10px">
</p>

> **Known chart bug (v2026.7.14):** `DiskUsageHigh` and `DiskAlmostFull` may show **FIRING** here even on a healthy cluster with plenty of free space. Their PromQL expression divides `kubelet_volume_stats_used_bytes` by `(kubelet_volume_stats_used_bytes + kube_pod_spec_volumes_persistentvolumeclaims_info)` instead of the PVC's actual capacity metric (`kubelet_volume_stats_capacity_bytes`), which mathematically evaluates to ~100% regardless of real usage. This was confirmed by comparing against `df -h` inside the pod (actual usage: 83%) and against the correct formula used by other `*-alerts` charts (e.g. `postgres-alerts`). The Grafana dashboard's own "Neo4j High Disk Usage" panel is unaffected and shows the correct percentage — only these two `PrometheusRule` expressions are wrong. Until fixed upstream, treat `DiskUsageHigh`/`DiskAlmostFull` from this chart as unreliable, or override them with a correct expression via `form.alert.groups.database.rules.diskUsageHigh` / `diskAlmostFull` in a custom values file.

### 4. Check AlertManager

Port-forward AlertManager to view any currently firing alerts.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager" src="/docs/images/neo4j/monitoring/neo4j-alerting-alertmanager.png" style="padding:10px">
</p>

Because of the disk-alert bug above, you'll see `DiskUsageHigh`/`DiskAlmostFull` here too (one per pod) even though the cluster is healthy — this is expected until the chart is fixed.

### 5. Explore the Grafana dashboard

Port-forward Grafana and log in.

```bash
$ kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
```

Open `http://localhost:3000` and navigate to the dashboard `kubedb.com / Neo4j / alert-neo4j / neo4j-alert-demo` that the Job imported in Step 2.

<p align="center">
  <img alt="Grafana — Neo4j Alerts Dashboard" src="/docs/images/neo4j/monitoring/neo4j-alerting-grafana-dashboard.png" style="padding:10px">
</p>

The dashboard mirrors the alert groups: **Neo4j Phase & Availability** (Down / Critical Phase), **Neo4j Resource Usage** (CPU, memory, disk), and **Neo4j Page Cache Metrics** (usage ratio, hit ratio, page faults). Note that **Neo4j High CPU Usage** shows "No data" on clusters without cAdvisor's `container_cpu_usage_seconds_total` metric exposed (common on some k3s setups) — this is an environment limitation, not a chart bug.

---

## Simulating a Firing Alert

The previous section showed that all genuine health alerts (everything except the buggy disk rules) are **INACTIVE** while the cluster is healthy. This section deliberately triggers the `KubeDBNeo4jPhaseNotReady` critical alert so you can observe the full alert lifecycle — from firing in Prometheus through to the AlertManager dashboard — and then resolve it.

Neo4j runs as a **single container per pod** — there is no separate exporter sidecar to keep reporting metrics once the database process dies, and Kubernetes will restart a killed process within seconds. Because `KubeDBNeo4jPhaseNotReady` requires the condition to persist for `for: 1m`, a single `kill` is not enough: you need to keep the pods crashing long enough for the KubeDB operator to mark the resource `NotReady` and hold it there past the one-minute window.

### 1. Crash the Neo4j process repeatedly

```bash
$ while true; do
    for i in 0 1 2; do
      kubectl exec -n alert-neo4j neo4j-alert-demo-$i -c neo4j -- kill 1 >/dev/null 2>&1
    done
    sleep 3
  done
```

Let this loop run for a couple of minutes (leave it running while you check the next steps), then stop it once you've captured the firing state.

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — KubeDBNeo4jPhaseNotReady Firing" src="/docs/images/neo4j/monitoring/neo4j-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

`KubeDBNeo4jPhaseNotReady` transitions from **INACTIVE** to **FIRING** once `kubedb_com_neo4j_status_phase{phase="NotReady"}` has read `1` continuously for the full `for: 1m` duration — this metric comes from the KubeDB operator's own view of the resource (exported via Panopticon), not from the Neo4j metrics endpoint itself.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager — KubeDBNeo4jPhaseNotReady Firing" src="/docs/images/neo4j/monitoring/neo4j-alerting-alertmanager-firing.png" style="padding:10px">
</p>

AlertManager shows the `KubeDBNeo4jPhaseNotReady` alert. The alert card displays:

- **Severity**: `critical`
- **neo4j**: `neo4j-alert-demo` in the `alert-neo4j` namespace
- **phase**: `NotReady`
- **Started**: timestamp when the alert first fired

AlertManager routes this alert to every receiver configured in your `alertmanagerConfig` (Slack, email, PagerDuty, webhook, etc.) based on your routing tree. If no receiver is configured, the alert is visible here but silently dropped.

### 4. Restore Neo4j

Stop the loop from step 1. The pods recover on their own — KubeDB just needs a few uninterrupted scrape/reconcile cycles to mark the resource `Ready` again.

```bash
$ kubectl get neo4j -n alert-neo4j neo4j-alert-demo -w
NAME               VERSION     STATUS   AGE
neo4j-alert-demo   2025.12.1   Ready    24m
```

Once the phase returns to `Ready`, Prometheus marks the alert **INACTIVE** again and AlertManager sends a **resolved** notification to all receivers.

---

## Alert Reference

All alerts are scoped to the `neo4j-alert-demo` instance in the `alert-neo4j` namespace via the PromQL label filters `pod=~"neo4j-alert-demo-.+$"` and `namespace=~"alert-neo4j"` (database group), or `app="neo4j-alert-demo"` and `namespace="alert-neo4j"` (provisioner group).

### Database Group

Fired based on live metrics from the Neo4j container's built-in metrics endpoint and node/kubelet metrics.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `Neo4jHighCPUUsage` | warning | 1m | Average pod CPU usage exceeds 80%. Requires cAdvisor `container_cpu_usage_seconds_total` — shows no data if unavailable. |
| `Neo4jHighMemoryUsage` | warning | 1m | Average pod memory usage exceeds 80% of the configured limit. |
| `Neo4jPageCacheUsageRatioHigh` | warning | 5m | More than 85% of the allocated page cache is in use — consider allocating more page cache. |
| `Neo4jPageCacheHitRatioLow` | warning | 5m | Page cache hit ratio has dropped below 98% — the database is going to disk too often. |
| `Neo4jPageFaultsHigh` | warning | 5m | More than 5000 page faults in the last 5 minutes — may indicate more page cache is required. |
| `Neo4jPageFaultFailuresHigh` | critical | 5m | Any failed page faults in the last 5 minutes — indicates potential disk I/O issues or corruption. |
| `DiskUsageHigh` | warning | 1m | Persistent volume usage exceeds 80%. **Known chart bug** — see [Verify End-to-End, step 3](#3-confirm-the-neo4j-alerts-are-inactive); currently fires regardless of actual usage. |
| `DiskAlmostFull` | critical | 1m | Persistent volume usage exceeds 95%. **Known chart bug** — same as above. |

### Provisioner Group

Monitors the KubeDB operator's view of the Neo4j resource phase (sourced from Panopticon, not the Neo4j metrics endpoint).

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBNeo4jPhaseNotReady` | critical | 1m | KubeDB marked the Neo4j resource `NotReady` — operator cannot reach a healthy cluster majority. |
| `KubeDBNeo4jPhaseCritical` | warning | 15m | One or more Neo4j core members are down; the cluster is degraded but not fully unavailable. |

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
          neo4jHighMemoryUsage:
            enabled: true
            duration: "5m"
            val: 90        # fire at 90% instead of the default 80%
            severity: warning
      provisioner:
        enabled: "none"    # disable all provisioner alerts
```

```bash
$ helm upgrade neo4j-alert-demo appscode/neo4j-alerts \
    -n alert-neo4j \
    --version=v2026.7.14 \
    --set grafana.enabled=false \
    -f custom-alerts.yaml
```

> Note: `-f` values files don't merge `grafana.url`/`grafana.apikey` automatically — re-pass them (or set `grafana.enabled=false`) on every `helm upgrade`, otherwise the dashboard-import Job re-runs with an empty URL/token and fails.

---

## Cleaning up

To remove all resources created in this tutorial, run the following commands.

```bash
# Remove the neo4j-alerts release (PrometheusRule + dashboard-import Job)
$ helm uninstall neo4j-alert-demo -n alert-neo4j

# Remove the imported Grafana dashboard (it is not removed by helm uninstall)
$ curl -s -X DELETE -H "Authorization: Bearer <grafana-token>" \
    http://localhost:3000/api/dashboards/uid/lhSzgLYDk

# Remove the Neo4j instance
$ kubectl delete neo4j -n alert-neo4j neo4j-alert-demo

# Delete namespace
$ kubectl delete ns alert-neo4j
```

## Next Steps

- Monitor your Neo4j database with KubeDB using [built-in Prometheus](/docs/guides/neo4j/monitoring/using-builtin-prometheus.md).
- Monitor your Neo4j database with KubeDB using [Prometheus operator](/docs/guides/neo4j/monitoring/using-prometheus-operator.md).
- Use [private Docker registry](/docs/guides/neo4j/private-registry/using-private-registry.md) to deploy Neo4j with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
