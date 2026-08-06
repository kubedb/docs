---
title: Solr Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: sl-monitoring-alerting
    name: Alerting
    parent: sl-monitoring-solr
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Solr Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed Solr instance using the `solr-alerts` Helm chart. This chart also bundles a Grafana dashboard that it imports automatically through a post-install Job — no separate dashboard chart is required.

## Before You Begin

* Ensure you have a Kubernetes cluster and that `kubectl` is configured to communicate with it. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Install the KubeDB operator by following the steps [here](/docs/setup/README.md).

* Deploy the database in the `alert-solr` namespace:

  ```bash
  $ kubectl create ns alert-solr
  namespace/alert-solr created
  ```

* Solr requires a reference to a KubeDB `ZooKeeper` cluster for coordination — deploy one first (see below).

* Before proceeding, complete the [Configuration](grafana-dashboard.md#configuration) steps to deploy **kube-prometheus-stack** and **Panopticon**.

* This tutorial assumes you already have a **kube-prometheus-stack** running in your cluster, with `Prometheus` configured so that both `serviceMonitorSelector` and `ruleSelector` match the label `release: prometheus`.

  To verify the selectors:

  ```bash
  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.ruleSelector}'
  {"matchLabels":{"release":"prometheus"}}

  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.serviceMonitorSelector}'
  {"matchLabels":{"release":"prometheus"}}
  ```

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/solr/monitoring/overview.md).

* You will also need a Grafana API key / token with **Editor** permission so the chart's dashboard-import Job can push the dashboard. See [Step 1](#step-1--create-a-grafana-api-key) below.

> Note: YAML files used in this tutorial are stored in [docs/examples/solr](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/solr) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

<p align="center">
  <img alt="Solr Alerting Architecture" src="/docs/images/solr/monitoring/solr-alerting-overview.svg">
</p>

- **KubeDB** deploys Solr with a metrics-exporter sidecar (container `exporter`) that exposes Solr's own metrics (`solr_metrics_*`, `solr_collections_*`).
- **ServiceMonitor** (named `{solr-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape the exporter every 10 seconds.
- **PrometheusRule** is created by the `solr-alerts` chart and contains alert definitions grouped by concern: database health and provisioner.
- **Dashboard-import Job** — when `grafana.enabled` is `true`, the chart also creates a one-shot `Job` that `POST`s a bundled dashboard JSON straight to your Grafana instance's `/api/dashboards/import` endpoint.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).

---

## Deploy the ZooKeeper Coordinator

Solr coordinates via a KubeDB `ZooKeeper` cluster, so deploy that first.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: zoo
  namespace: alert-solr
spec:
  version: 3.8.3
  replicas: 3
  deletionPolicy: WipeOut
  adminServerPort: 8080
  storage:
    resources:
      requests:
        storage: "100Mi"
    storageClassName: local-path
    accessModes:
      - ReadWriteOnce
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/monitoring/zookeeper-alert-demo.yaml
zookeeper.kubedb.com/zoo created

$ kubectl get zookeeper -n alert-solr zoo
NAME   VERSION   STATUS   AGE
zoo    3.8.3     Ready    3m
```

## Deploy Solr with Monitoring Enabled

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-alert
  namespace: alert-solr
spec:
  version: 9.8.0
  replicas: 3
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  solrModules:
  - s3-repository
  - gcs-repository
  - prometheus-exporter
  zookeeperRef:
    name: zoo
    namespace: alert-solr
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: local-path
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/monitoring/solr-alert-demo.yaml
solr.kubedb.com/solr-alert created
```

Wait for the database to go into `Ready` state.

```bash
$ kubectl get solr -n alert-solr solr-alert
NAME         VERSION   STATUS   AGE
solr-alert   9.8.0     Ready    5m
```

KubeDB brings up 3 pods, one per Solr node:

```bash
$ kubectl get pods -n alert-solr
NAME           READY   STATUS    RESTARTS   AGE
solr-alert-0   1/1     Running   0          5m
solr-alert-1   1/1     Running   0          5m
solr-alert-2   1/1     Running   0          5m
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n alert-solr --selector="app.kubernetes.io/instance=solr-alert"
NAME               TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
solr-alert         ClusterIP   10.43.219.25    <none>        8983/TCP   5m
solr-alert-pods    ClusterIP   None            <none>        8983/TCP   5m
solr-alert-stats   ClusterIP   10.43.163.252   <none>        9854/TCP   5m
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n alert-solr
NAME               AGE
solr-alert-stats   5m
```

Verify that the `ServiceMonitor` carries the `release: prometheus` label so Prometheus discovers it.

```bash
$ kubectl get servicemonitor -n alert-solr solr-alert-stats \
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
      -d '{"name":"solr-alerts-demo","role":"Editor"}'
  # Note the returned "key"

  # Stop the port-forward
  $ kill %1
  ```

Either way, you end up with a bearer token to use as `grafana.apikey` below.

## Step 2 — Install solr-alerts

### Why the Helm release name matters

The chart derives the `PrometheusRule` name and scopes every PromQL expression from the **Helm release name** — so the release name must match the Solr object's name (`solr-alert`).

### Install

```bash
$ helm upgrade -i solr-alert appscode/solr-alerts \
    -n alert-solr \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus \
    --set grafana.enabled=true \
    --set grafana.url="http://prometheus-grafana.monitoring.svc:80" \
    --set grafana.apikey="<token-from-above>" \
    --set grafana.jobName=solr-alert-stats
```

| Flag | Value | Purpose |
|------|-------|---------|
| `grafana.url` | in-cluster Grafana URL | The dashboard-import Job runs **inside the cluster**, so this must be a cluster-internal address, not `localhost` |
| `grafana.apikey` | token from Step 1 | Authenticates the dashboard-import `POST` request |
| `grafana.jobName` | `solr-alert-stats` | **Required** — the chart's default (`kubedb-databases`) doesn't match any real Prometheus job, so most of the dashboard's panels show "No data" unless you override it to your instance's actual stats-service name |

> To install **alerts only, without the dashboard**, omit the `grafana.*` flags (or set `--set grafana.enabled=false`).

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-solr
NAME                AGE
solr-alert     30s

$ kubectl get prometheusrule -n alert-solr solr-alert \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Verify the dashboard-import Job

```bash
$ kubectl get job -n alert-solr
NAME                  STATUS     COMPLETIONS   AGE
solr-alert-post-job   Complete   1/1           17s

$ kubectl logs -n alert-solr job/solr-alert-post-job
{"pluginId":"","title":"kubedb.com / Solr / alert-solr / solr-alert","imported":true, ...}
```

A `"imported":true` response confirms the dashboard `kubedb.com / Solr / alert-solr / solr-alert` now exists in Grafana.

### Confirm Prometheus loaded the rules

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules` and locate the `solr.database` and `solr.provisioner` groups.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/solr/monitoring/solr-alerting-prom-rules.png" style="padding:10px">
</p>

Both groups should show **OK**. `solr-alerts` v2026.7.14 has no `opsManager`/`stash`/`kubeStash` groups — only `database` and `provisioner`. Note there is no plain `SolrDown` alert; the closest equivalent is `SolrDownShards` (shard-level) and the provisioner group's `KubeDBSolrPhaseNotReady`.

---

## Verify End-to-End

### 1. Check the Prometheus target is UP

Open `http://localhost:9090/query?g0.expr=up%7Bnamespace%3D%22alert-solr%22%7D&g0.tab=1`.

<p align="center">
  <img alt="Prometheus up query — solr-alert-0 UP" src="/docs/images/solr/monitoring/solr-alerting-prom-target.png" style="padding:10px">
</p>

All three pods — `solr-alert-0`, `solr-alert-1`, `solr-alert-2` — report `up == 1`, confirming Prometheus is scraping every Solr node in the `alert-solr` namespace.

### 2. Confirm the Solr alerts are inactive

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — Solr groups inactive" src="/docs/images/solr/monitoring/solr-alerting-prom-alerts.png" style="padding:10px">
</p>

All 9 rules in the `solr.database` group and both rules in the `solr.provisioner` group show **INACTIVE**, confirming the cluster is healthy and no thresholds are breached.

### 3. Check AlertManager

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager" src="/docs/images/solr/monitoring/solr-alerting-alertmanager.png" style="padding:10px">
</p>

---

## Simulating a Firing Alert

This section deliberately triggers `KubeDBSolrPhaseNotReady` by repeatedly crashing the main Solr JVM process on one node. A single `kill` restarts fast enough that the container becomes `Ready` again before KubeDB's health check can observe the outage, so the crash needs to be sustained over a longer window than the `for` duration of the alert.

### 1. Crash the Solr process repeatedly

```bash
$ end=$(( $(date +%s) + 150 ))
  while [ $(date +%s) -lt $end ]; do
    kubectl exec -n alert-solr solr-alert-0 -c solr -- sh -c 'pid=$(pgrep -f "org.apache.solr" | head -1); [ -n "$pid" ] && kill -9 "$pid"' >/dev/null 2>&1
    sleep 5
  done
```

Watch the CR phase move to `NotReady`:

```bash
$ kubectl get solr -n alert-solr solr-alert -o jsonpath='{.status.phase}'
NotReady
```

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — SolrDownShards Firing, KubeDBSolrPhaseNotReady Pending" src="/docs/images/solr/monitoring/solr-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

`SolrDownShards` (database group, `for: 30s`) reaches **FIRING** first, since the crashed node's shard replica goes unreachable almost immediately. The provisioner-group `KubeDBSolrPhaseNotReady` (`for: 1m`) takes longer — in this screenshot it's still **PENDING**, and transitions to **FIRING** once the KubeDB operator holds the resource at `NotReady` past the full one-minute window (confirmed via `kubectl get solr ... -o jsonpath='{.status.phase}'` above, and in the AlertManager screenshot next).

### 3. Check the AlertManager dashboard

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager — KubeDBSolrPhaseNotReady Firing" src="/docs/images/solr/monitoring/solr-alerting-alertmanager-firing.png" style="padding:10px">
</p>

AlertManager shows the `KubeDBSolrPhaseNotReady` alert, confirming it did reach **FIRING** shortly after the previous screenshot. The alert card displays labels including:

- **alertname**: `KubeDBSolrPhaseNotReady`
- **severity**: `critical`
- **app**: `solr-alert`, **app_namespace**: `alert-solr`
- **phase**: `NotReady`
- **k8s_kind**: `Solr`

Note that the `instance`/`pod`/`job` labels point at the KubeDB operator's **panopticon** component (`job="panopticon"`), not at the Solr pod itself — because this alert is derived from the operator's own status metric rather than from the Solr exporter.

AlertManager routes this alert to every receiver configured in your `alertmanagerConfig` (Slack, email, PagerDuty, webhook, etc.) based on your routing tree. If no receiver is configured, the alert is visible here but silently dropped.

### 4. Restore Solr

Stop the loop from step 1.

```bash
$ kubectl get solr -n alert-solr solr-alert -w
NAME         VERSION   STATUS   AGE
solr-alert   9.8.0     Ready    24m
```

If Solr does not recover on its own within a minute or two, force a clean restart: `kubectl delete pod -n alert-solr solr-alert-0`.

---

## Alert Reference

All alerts are scoped to the `solr-alert` instance in the `alert-solr` namespace via `job="solr-alert-stats"` / `namespace="alert-solr"` (database group), or `app="solr-alert"` / `namespace="alert-solr"` (provisioner group).

### Database Group

Fired based on live metrics from the Solr exporter sidecar and node/kubelet metrics.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `SolrDownShards` | critical | 30s | One or more collection shards have no active replica. |
| `SolrRecoveryFailedShards` | critical | 30s | A shard replica is stuck in recovery-failed state. |
| `SolrHighThreadRunning` | warning | 30s | JVM thread count is high. |
| `SolrHighPoolSize` | warning | 30s | JVM memory pool usage is high. |
| `SolrHighQPS` | warning | 30s | Query rate is unusually high for a collection. |
| `SolrHighHeapSize` | warning | 30s | JVM heap usage is high. |
| `SolrHighBufferSize` | warning | 30s | JVM direct buffer usage is high. |
| `DiskUsageHigh` | warning | 1m | Persistent volume usage exceeds 80%. |
| `DiskAlmostFull` | critical | 1m | Persistent volume usage exceeds 95%. |

### Provisioner Group

Monitors the KubeDB operator's view of the Solr resource phase (sourced from Panopticon, not the Solr metrics endpoint).

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBSolrPhaseNotReady` | critical | 1m | KubeDB marked the Solr resource `NotReady`. |
| `KubeDBSolrPhaseCritical` | warning | 1m | Solr is degraded but not fully unavailable. |

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
          solrHighQPS:
            enabled: true
            duration: "2m"
            severity: warning
```

```bash
$ helm upgrade solr-alert appscode/solr-alerts \
    -n alert-solr \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

To remove all resources created in this tutorial, run the following commands.

```bash
# Remove the solr-alerts release (PrometheusRule + dashboard-import Job)
$ helm uninstall solr-alert -n alert-solr

# Remove the imported Grafana dashboard (it is not removed by helm uninstall)
$ curl -s -X DELETE -H "Authorization: Bearer <grafana-token>" \
    http://localhost:3000/api/dashboards/uid/<uid>

$ kubectl delete solr -n alert-solr solr-alert
$ kubectl delete zookeeper -n alert-solr zoo
$ kubectl delete ns alert-solr
```

## Next Steps

- Monitor your Solr instance with KubeDB using [built-in Prometheus](/docs/guides/solr/monitoring/prometheus-builtin.md).
- Monitor your Solr instance with KubeDB using [Prometheus operator](/docs/guides/solr/monitoring/prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
