---
title: Ignite Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: ig-monitoring-alerting
    name: Alerting
    parent: ig-monitoring-ignite
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Ignite Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed Ignite instance using the `ignite-alerts` Helm chart. This chart also bundles a Grafana dashboard that it imports automatically through a post-install Job — no separate dashboard chart is required.

## Before You Begin

* Ensure you have a Kubernetes cluster and that `kubectl` is configured to communicate with it. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Install the KubeDB operator by following the steps [here](/docs/setup/README.md).

* Deploy the database in the `alert-ignite` namespace:

  ```bash
  $ kubectl create ns alert-ignite
  namespace/alert-ignite created
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

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/ignite/monitoring/overview.md).

* You will also need a Grafana API key / token with **Editor** permission so the chart's dashboard-import Job can push the dashboard. See [Step 1](#step-1--create-a-grafana-api-key) below.

> Note: YAML files used in this tutorial are stored in [docs/examples/ignite](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/ignite) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

<p align="center">
  <img alt="Ignite Alerting Architecture" src="/docs/images/ignite/monitoring/ignite-alerting-overview.svg">
</p>

- **KubeDB** deploys Ignite with a metrics-exporter sidecar (container `exporter`) that exposes Ignite's own JMX-derived metrics (`sys_*`, `io_*`, `cluster_*`, `ignite_*`).
- **ServiceMonitor** (named `{ignite-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape the exporter every 10 seconds.
- **PrometheusRule** is created by the `ignite-alerts` chart and contains alert definitions grouped by concern: database health (which also embeds KubeDB-operator-sourced `IgniteDown`/`IgnitePhaseCritical` alerts) and provisioner.
- **Dashboard-import Job** — when `grafana.enabled` is `true`, the chart also creates a one-shot `Job` that `POST`s a bundled dashboard JSON straight to your Grafana instance's `/api/dashboards/import` endpoint.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).

---

## Deploy Ignite with Monitoring Enabled

At first, let's deploy an Ignite instance with monitoring enabled. Below is the Ignite object we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Ignite
metadata:
  name: ignite-alert-demo
  namespace: alert-ignite
spec:
  replicas: 1
  version: "2.17.0"
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

Here,

- `spec.replicas: 1` deploys a single-node Ignite instance.
- `spec.monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` resource managed by the Prometheus operator.
- `spec.monitor.prometheus.serviceMonitor.labels.release: prometheus` adds the `release: prometheus` label to the created `ServiceMonitor`, matching the Prometheus `serviceMonitorSelector` so the target is discovered automatically.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ignite/monitoring/ignite-alert-demo.yaml
ignite.kubedb.com/ignite-alert-demo created
```

Now, wait for the database to go into `Ready` state.

```bash
$ kubectl get ignite -n alert-ignite ignite-alert-demo
NAME                VERSION   STATUS   AGE
ignite-alert-demo   2.17.0    Ready    3m
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n alert-ignite --selector="app.kubernetes.io/instance=ignite-alert-demo"
NAME                          TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
ignite-alert-demo             ClusterIP   10.43.10.20    <none>        10800/TCP   3m
ignite-alert-demo-pods        ClusterIP   None           <none>        10800/TCP   3m
ignite-alert-demo-stats       ClusterIP   10.43.10.21    <none>        8080/TCP    3m
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n alert-ignite
NAME                       AGE
ignite-alert-demo-stats    3m
```

Verify that the `ServiceMonitor` carries the `release: prometheus` label so Prometheus discovers it.

```bash
$ kubectl get servicemonitor -n alert-ignite ignite-alert-demo-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Create a Grafana API Key

The chart's dashboard-import Job authenticates to Grafana with a bearer token, so create one first.

* **Grafana 9+**: **Administration → Service accounts → Add service account** → role **Editor** → **Add token**. Copy the token.
* **Grafana 8.x and earlier** (no Service Accounts UI, e.g. the bundled `kube-prometheus-stack` Grafana 7.5.5): use the legacy **API Keys** endpoint instead:

  ```bash
  # Port-forward Grafana
  $ kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80

  # Retrieve the admin password
  $ kubectl get secret -n monitoring prometheus-grafana \
      -o jsonpath='{.data.admin-password}' | base64 -d && echo

  # Create an API key with Editor role
  $ curl -s -X POST -H "Content-Type: application/json" \
      -u admin:<grafana_password> \
      http://localhost:3000/api/auth/keys \
      -d '{"name":"ignite-alerts-demo","role":"Editor"}'
  # Note the returned "key"

  # Stop the port-forward
  $ kill %1
  ```

Either way, you end up with a bearer token to use as `grafana.apikey` below.

## Step 2 — Install ignite-alerts

### Why the Helm release name matters

The chart derives the `PrometheusRule` name and scopes every PromQL expression from the **Helm release name** — so the release name must match the Ignite object's name (`ignite-alert-demo`).

### Install

```bash
$ helm upgrade -i ignite-alert-demo oci://ghcr.io/appscode-charts/ignite-alerts \
    -n alert-ignite \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus \
    --set grafana.enabled=true \
    --set grafana.url="http://prometheus-grafana.monitoring.svc:80" \
    --set grafana.apikey="<token-from-above>" \
    --set grafana.jobName=ignite-alert-demo-stats
```

| Flag | Value | Purpose |
|------|-------|---------|
| `grafana.url` | in-cluster Grafana URL | The dashboard-import Job runs **inside the cluster**, so this must be a cluster-internal address, not `localhost` |
| `grafana.apikey` | token from Step 1 | Authenticates the dashboard-import `POST` request |
| `grafana.jobName` | `ignite-alert-demo-stats` | **Required** — the chart's default (`kubedb-databases`) doesn't match any real Prometheus job, so most of the dashboard's panels show "No data" unless you override it to your instance's actual stats-service name |

> To install **alerts only, without the dashboard**, omit the `grafana.*` flags (or set `--set grafana.enabled=false`).

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-ignite
NAME                  AGE
ignite-alert-demo     30s

$ kubectl get prometheusrule -n alert-ignite ignite-alert-demo \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Verify the dashboard-import Job

```bash
$ kubectl get job -n alert-ignite
NAME                         STATUS     COMPLETIONS   AGE
ignite-alert-demo-post-job   Complete   1/1           17s

$ kubectl logs -n alert-ignite job/ignite-alert-demo-post-job
{"pluginId":"","title":"kubedb.com / Ignite / alert-ignite / ignite-alert-demo","imported":true, ...}
```

A `"imported":true` response confirms the dashboard `kubedb.com / Ignite / alert-ignite / ignite-alert-demo` now exists in Grafana.

### Confirm Prometheus loaded the rules

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules` and locate the `ignite.database` and `ignite.provisioner` groups.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/ignite/monitoring/ignite-alerting-prom-rules.png" style="padding:10px">
</p>

Both groups should show **OK**. `ignite-alerts` v2026.7.14 has no `opsManager`/`stash`/`kubeStash` groups — only `database` and `provisioner`.

> **Note the overlap:** the `database` group's `IgniteDown` (`for: 30s`) and `IgnitePhaseCritical` (`for: 1m`) key off the same `kubedb_com_ignite_status_phase` metric as the `provisioner` group's `KubeDBIgnitePhaseNotReady`/`KubeDBIgnitePhaseCritical` (`for: 1m`/`15m`) — expect both pairs to eventually fire together during a real outage, at different times.

---

## Verify End-to-End

### 1. Check the Prometheus target is UP

Open `http://localhost:9090/query?g0.expr=up%7Bnamespace%3D%22alert-ignite%22%7D&g0.tab=1`.

<p align="center">
  <img alt="Prometheus up query — ignite-alert-demo-0 UP" src="/docs/images/ignite/monitoring/ignite-alerting-prom-target.png" style="padding:10px">
</p>

### 2. Confirm the Ignite alerts are inactive

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — Ignite groups inactive" src="/docs/images/ignite/monitoring/ignite-alerting-prom-alerts.png" style="padding:10px">
</p>

All rules should show **INACTIVE**. `IgniteClusterNoBaselineNode` will only have data once the cluster's baseline topology is activated (a normal part of Ignite persistence setup, not KubeDB-specific).

### 3. Check AlertManager

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager" src="/docs/images/ignite/monitoring/ignite-alerting-alertmanager.png" style="padding:10px">
</p>

---

## Simulating a Firing Alert

This section deliberately triggers `IgniteDown` (`for: 30s`, the fastest down-signal) by crashing the main Ignite JVM process.

### 1. Crash the Ignite process

```bash
$ kubectl exec -n alert-ignite ignite-alert-demo-0 -c ignite -- sh -c '
    end=$(( $(date +%s) + 60 ));
    while [ $(date +%s) -lt $end ]; do
      pid=$(pgrep -f "org.apache.ignite" | head -1);
      [ -n "$pid" ] && kill -9 "$pid" 2>/dev/null;
      sleep 1;
    done'
```

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — IgniteDown Firing" src="/docs/images/ignite/monitoring/ignite-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

`IgniteDown` (`kubedb_com_ignite_status_phase{phase!="Ready"} == 1`, `for: 30s`) should transition to **FIRING** first.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager — IgniteDown Firing" src="/docs/images/ignite/monitoring/ignite-alerting-alertmanager-firing.png" style="padding:10px">
</p>

### 4. Restore Ignite

Stop the loop from step 1.

```bash
$ kubectl get ignite -n alert-ignite ignite-alert-demo -w
NAME                VERSION   STATUS   AGE
ignite-alert-demo   2.17.0    Ready    24m
```

If Ignite does not recover on its own within a minute or two, force a clean restart: `kubectl delete pod -n alert-ignite ignite-alert-demo-0`.

---

## Alert Reference

All alerts are scoped to the `ignite-alert-demo` instance in the `alert-ignite` namespace via `job="ignite-alert-demo-stats"` / `namespace="alert-ignite"` (database group), or `app="ignite-alert-demo"` / `namespace="alert-ignite"` (provisioner group and the two operator-phase alerts embedded in the database group).

### Database Group

Fired based on live metrics from the Ignite exporter sidecar and node/kubelet metrics, plus two KubeDB-operator-sourced phase alerts (`IgniteDown`, `IgnitePhaseCritical`).

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `IgniteDown` | critical | 30s | KubeDB operator view: resource not `Ready`. Fastest down-signal available. |
| `IgnitePhaseCritical` | warning | 1m | KubeDB operator view: resource `Critical` (duplicates the provisioner group's own version at a different `for`). |
| `IgniteClusterNoBaselineNode` | warning | 1m | The cluster has no baseline topology node registered. |
| `IgniteRestarted` | warning | 1m | Uptime indicates a recent restart. |
| `IgniteHighCPULoad` | warning | 1m | System CPU load exceeds 80%. |
| `IgniteHighHeapMemoryUsed` | warning | 1m | JVM heap usage is high. |
| `IgniteHighDataregionOffHeapUsed` | warning | 1m | Off-heap data region usage is high. |
| `IgniteJVMPausesTotalDuration` | warning | 1m | Long JVM GC pauses detected. |
| `DiskUsageHigh` | warning | 1m | Persistent volume usage exceeds 80%. |
| `DiskAlmostFull` | critical | 1m | Persistent volume usage exceeds 95%. |

### Provisioner Group

Monitors the KubeDB operator's view of the Ignite resource phase.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBIgnitePhaseNotReady` | critical | 1m | KubeDB marked the Ignite resource `NotReady`. |
| `KubeDBIgnitePhaseCritical` | warning | 15m | Ignite is degraded but not fully unavailable. |

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
          igniteHighCPULoad:
            enabled: true
            duration: "5m"
            severity: warning
```

```bash
$ helm upgrade ignite-alert-demo oci://ghcr.io/appscode-charts/ignite-alerts \
    -n alert-ignite \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

To remove all resources created in this tutorial, run the following commands.

```bash
# Remove the ignite-alerts release (PrometheusRule + dashboard-import Job)
$ helm uninstall ignite-alert-demo -n alert-ignite

# Remove the imported Grafana dashboard (it is not removed by helm uninstall)
$ curl -s -X DELETE -H "Authorization: Bearer <grafana-token>" \
    http://localhost:3000/api/dashboards/uid/<uid>

$ kubectl delete ignite -n alert-ignite ignite-alert-demo
$ kubectl delete ns alert-ignite
```

## Next Steps

- Monitor your Ignite instance with KubeDB using [built-in Prometheus](/docs/guides/ignite/monitoring/using-builtin-prometheus.md).
- Monitor your Ignite instance with KubeDB using [Prometheus operator](/docs/guides/ignite/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
