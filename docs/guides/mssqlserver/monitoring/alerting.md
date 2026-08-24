---
title: MSSQLServer Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: ms-monitoring-alerting
    name: Alerting
    parent: ms-monitoring
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MSSQLServer Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed Microsoft SQL Server instance using the `mssqlserver-alerts` Helm chart. This chart also bundles a Grafana dashboard that it imports automatically through a post-install Job — no separate dashboard chart is required.

## Before You Begin

* Ensure you have a Kubernetes cluster and that `kubectl` is configured to communicate with it. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Install the KubeDB operator by following the steps [here](/docs/setup/README.md).

* Install [cert-manager](https://cert-manager.io/docs/installation/) — MSSQLServer requires TLS, issued via a cert-manager `Issuer`.

* Deploy the database in the `alert-mssqlserver` namespace:

  ```bash
  $ kubectl create ns alert-mssqlserver
  namespace/alert-mssqlserver created
  ```

* Create a self-signed CA and an `Issuer` for MSSQLServer to use:

  ```bash
  $ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=MSSQLServer/O=kubedb"

  $ kubectl create secret tls mssqlserver-ca --cert=ca.crt --key=ca.key --namespace=alert-mssqlserver
  secret/mssqlserver-ca created

  $ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/monitoring/mssqlserver-ca-issuer.yaml
  issuer.cert-manager.io/mssqlserver-ca-issuer created
  ```

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/mssqlserver/monitoring/overview.md).

* You will also need a Grafana API key / token with **Editor** permission so the chart's dashboard-import Job can push the dashboard. See [Step 1](#step-1--create-a-grafana-api-key) below.

> Note: YAML files used in this tutorial are stored in [docs/examples/mssqlserver](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Configuration

> Step 1 (`kube-prometheus-stack`) is required to follow this tutorial. Step 2 (Panopticon) is required for the **Provisioner Group** alerts below (`KubeDBMSSQLServerPhase...`) — skip it only if you just want the exporter-based **Database Group** alerts. If you have already completed the step(s) you need in another guide, skip ahead.

### Step 1: Deploy kube-prometheus-stack

`kube-prometheus-stack` installs Prometheus, Prometheus Operator, Alertmanager, and Grafana together. This is the recommended way to get the full monitoring stack on Kubernetes.

Add the prometheus-community Helm repo and install:

```bash
$ helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
$ helm repo update

$ helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace \
  --set grafana.image.tag=7.5.5
```

Wait for all pods to be ready:

```bash
$ kubectl get pods -n monitoring
NAME                                                   READY   STATUS    RESTARTS   AGE
alertmanager-prometheus-kube-prometheus-alertmanager-0 2/2     Running   0          2m
prometheus-grafana-xxxx                                3/3     Running   0          2m
prometheus-kube-prometheus-operator-xxxx               1/1     Running   0          2m
prometheus-kube-prometheus-prometheus-0                2/2     Running   0          2m
prometheus-kube-state-metrics-xxxx                     1/1     Running   0          2m
```

Find the `serviceMonitorSelector`/`ruleSelector` labels that Prometheus uses to pick up `ServiceMonitor`/`PrometheusRule` objects — this is the `release: prometheus` label used throughout this tutorial.

```bash
$ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.ruleSelector}'
{"matchLabels":{"release":"prometheus"}}

$ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.serviceMonitorSelector}'
{"matchLabels":{"release":"prometheus"}}
```

### Step 2: Install Panopticon (required for the Provisioner Group alerts)

Panopticon is the Appscode operator that exports the KubeDB operator's own view of every resource — `kubedb_com_mssqlserver_status_phase` and related metrics. It's what powers the **Provisioner Group** alerts below (`KubeDBMSSQLServerPhaseNotReady`/`KubeDBMSSQLServerPhaseCritical`). Skip this step if you only need the exporter-based **Database Group** alerts.

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update

$ helm upgrade --install panopticon appscode/panopticon \
  --version v2026.4.30 \
  --namespace kubeops --create-namespace \
  --set monitoring.enabled=true \
  --set monitoring.agent=prometheus.io/operator \
  --set monitoring.serviceMonitor.labels.release=prometheus \
  --set-file license=/path/to/kubedb-license.txt \
  --wait --timeout 5m0s
```

Verify Panopticon is running:

```bash
$ kubectl get pods -n kubeops
NAME                          READY   STATUS    RESTARTS   AGE
panopticon-xxxx               1/1     Running   0          1m
```

## Overview

<p align="center">
  <img alt="MSSQLServer Alerting Architecture" src="/docs/images/mssqlserver/monitoring/mssqlserver-alerting-overview.svg">
</p>

- **KubeDB** deploys MSSQLServer with a metrics-exporter sidecar (container `exporter`) that exposes metrics on the `{mssqlserver-name}-stats` service.
- **ServiceMonitor** (named `{mssqlserver-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape the exporter every 10 seconds.
- **PrometheusRule** is created by the `mssqlserver-alerts` chart and contains MSSQLServer alert definitions grouped by concern: database health, provisioner, and ops-manager. A fourth group (`kubeStash`) covers backup/restore alerts and is disabled for this tutorial.
- **Dashboard-import Job** — when `grafana.enabled` is `true`, the chart also creates a one-shot `Job` that `POST`s a bundled dashboard JSON straight to your Grafana instance's `/api/dashboards/import` endpoint.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).

---

## Deploy MSSQLServer with Monitoring Enabled

At first, let's deploy an MSSQLServer instance with monitoring enabled. Below is the MSSQLServer object we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssqlserver-alert-demo
  namespace: alert-mssqlserver
spec:
  version: "2025-cu0"
  replicas: 1
  storageType: Durable
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation # Change to a licensed edition for production use
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

- `spec.tls.issuerRef` points at the `mssqlserver-ca-issuer` Issuer created above, so the SQL Server endpoint is served over TLS.
- `spec.monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` resource managed by the Prometheus operator.
- `spec.monitor.prometheus.serviceMonitor.labels.release: prometheus` adds the `release: prometheus` label to the created `ServiceMonitor`, matching the Prometheus `serviceMonitorSelector` so the target is discovered automatically.
- `ACCEPT_EULA=Y` and `MSSQL_PID` are required by the upstream SQL Server image itself, not KubeDB — see [Microsoft's environment variable reference](https://learn.microsoft.com/en-us/sql/linux/sql-server-linux-configure-environment-variables) for valid `MSSQL_PID` values.

Let's create the MSSQLServer resource.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/monitoring/mssqlserver-alert-demo.yaml
mssqlserver.kubedb.com/mssqlserver-alert-demo created
```

Wait for the database to go into `Ready` state.

```bash
$ kubectl get mssqlserver -n alert-mssqlserver mssqlserver-alert-demo
NAME                     VERSION    STATUS   AGE
mssqlserver-alert-demo   2025-cu0   Ready    5m
```

> First boot takes several minutes — SQL Server doesn't write to its error log or create system databases (`master.mdf`, `msdbdata.mdf`, ...) until it's actually ready to initialize, so `kubectl get pods` showing `2/2 Running` doesn't by itself mean the instance is `Ready` yet. Give it 3-5 minutes on a cold node before assuming something's stuck.

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n alert-mssqlserver --selector="app.kubernetes.io/instance=mssqlserver-alert-demo"
NAME                             TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
mssqlserver-alert-demo           ClusterIP   10.43.10.20    <none>        1433/TCP    5m
mssqlserver-alert-demo-pods      ClusterIP   None           <none>        1433/TCP    5m
mssqlserver-alert-demo-stats     ClusterIP   10.43.10.21    <none>        56790/TCP   5m
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n alert-mssqlserver
NAME                           AGE
mssqlserver-alert-demo-stats   5m

$ kubectl get servicemonitor -n alert-mssqlserver mssqlserver-alert-demo-stats \
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
  $ kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80&

  # Retrieve the admin password
  $ kubectl get secret -n monitoring prometheus-grafana \
      -o jsonpath='{.data.admin-password}' | base64 -d && echo

  # Create an API key with Editor role
  $ curl -s -X POST -H "Content-Type: application/json" \
      -u admin:<grafana_password> \
      http://localhost:3000/api/auth/keys \
      -d '{"name":"mssqlserver-alerts-demo","role":"Editor"}'
  # Note the returned "key"

  # Stop the port-forward
  $ kill %1
  ```

Either way, you end up with a bearer token to use as `grafana.apikey` below.

## Step 2 — Install mssqlserver-alerts

### Why the Helm release name matters

The chart derives the `PrometheusRule` name and scopes every PromQL expression (via `job="{release-name}-stats"` / `app="{release-name}"`) from the **Helm release name** — so the release name must match the MSSQLServer object's name (`mssqlserver-alert-demo`).

The chart's default label is `release: kube-prometheus-stack`, so we must also override it at install time to match the Prometheus `ruleSelector` — installing without `--set form.alert.labels.release=prometheus` produces a `PrometheusRule` that Prometheus never loads, until the label is corrected with a `helm upgrade`.

### Install

This tutorial covers the `database`, `provisioner`, and `opsManager` alert groups; the `kubeStash` group is disabled below.

```bash
$ helm upgrade -i mssqlserver-alert-demo appscode/mssqlserver-alerts \
    -n alert-mssqlserver \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus \
    --set form.alert.groups.kubeStash.enabled=none \
    --set grafana.enabled=true \
    --set grafana.url="http://prometheus-grafana.monitoring.svc:80" \
    --set grafana.apikey="<token-from-above>" \
    --set grafana.jobName=mssqlserver-alert-demo-stats \
    --set form.alert.appSuffix=ms-grafana-demo
```

| Flag | Value | Purpose |
|------|-------|---------|
| `grafana.url` | in-cluster Grafana URL | The dashboard-import Job runs **inside the cluster**, so this must be a cluster-internal address, not `localhost` |
| `grafana.apikey` | token from Step 1 | Authenticates the dashboard-import `POST` request |
| `grafana.jobName` | `mssqlserver-alert-demo-stats` | **Required** — the chart's default (`kubedb-databases`) doesn't match any real Prometheus job, so most of the dashboard's panels show "No data" unless you override it to your instance's actual stats-service name |

> To install **alerts only, without the dashboard**, omit the `grafana.*` flags (or set `--set grafana.enabled=false`).

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-mssqlserver
NAME                       AGE
mssqlserver-alert-demo     30s

$ kubectl get prometheusrule -n alert-mssqlserver mssqlserver-alert-demo \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Verify the dashboard-import Job

```bash
$ kubectl get job -n alert-mssqlserver
NAME                              STATUS     COMPLETIONS   AGE
mssqlserver-alert-demo-post-job   Complete   1/1           17s

$ kubectl logs -n alert-mssqlserver job/mssqlserver-alert-demo-post-job
{"pluginId":"","title":"kubedb.com / MSSQLServer / alert-mssqlserver / mssqlserver-alert-demo","imported":true, ...}
```

A `"imported":true` response confirms the dashboard `kubedb.com / MSSQLServer / alert-mssqlserver / mssqlserver-alert-demo` now exists in Grafana.

### Confirm Prometheus loaded the rules

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules` and locate the `mssqlserver.database`, `mssqlserver.provisioner`, and `mssqlserver.opsManager` groups.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/mssqlserver/monitoring/mssqlserver-alerting-prom-rules.png" style="padding:10px">
</p>

All three groups should show **OK** (`kubeStash` is absent since it was disabled at install time).

---

## Verify End-to-End

### 1. Check the Prometheus target is UP

Open `http://localhost:9090/query?g0.expr=up%7Bnamespace%3D%22alert-mssqlserver%22%7D&g0.tab=1`.

<p align="center">
  <img alt="Prometheus up query — mssqlserver-alert-demo-0 UP" src="/docs/images/mssqlserver/monitoring/mssqlserver-alerting-prom-target.png" style="padding:10px">
</p>

Both series report `up == 1` — the exporter's own target and its `target="mssql_database"` sub-target — confirming Prometheus is scraping `mssqlserver-alert-demo-0` successfully.

### 2. Confirm the MSSQLServer alerts are inactive

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — MSSQLServer groups inactive" src="/docs/images/mssqlserver/monitoring/mssqlserver-alerting-prom-alerts.png" style="padding:10px">
</p>

All rules should show **INACTIVE**.

### 3. Check AlertManager

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager" src="/docs/images/mssqlserver/monitoring/mssqlserver-alerting-alertmanager.png" style="padding:10px">
</p>

This view isn't filtered to a namespace, so alerts from other databases on a shared cluster may also appear here — use the filter box (`app_namespace="alert-mssqlserver"`) to confirm none are specific to this instance.

---

## Simulating a Firing Alert

This section deliberately triggers `MSSQLServerInstanceDown` (instant, `for: 0m`) by crashing the main `sqlservr` process.

### 1. Crash the MSSQLServer process

```bash
$ kubectl exec -n alert-mssqlserver mssqlserver-alert-demo-0 -c mssql -- sh -c '
    end=$(( $(date +%s) + 30 ));
    while [ $(date +%s) -lt $end ]; do
      pid=$(pgrep -x sqlservr | head -1);
      [ -n "$pid" ] && kill -9 "$pid" 2>/dev/null;
      sleep 1;
    done'
```

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — MSSQLServerInstanceDown Firing" src="/docs/images/mssqlserver/monitoring/mssqlserver-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

`MSSQLServerInstanceDown` (`up == 0`) should transition straight to **FIRING**.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager — MSSQLServerInstanceDown Firing" src="/docs/images/mssqlserver/monitoring/mssqlserver-alerting-alertmanager-firing.png" style="padding:10px">
</p>

AlertManager shows the `MSSQLServerInstanceDown` alert. The alert card displays labels including:

- **alertname**: `MSSQLServerInstanceDown`
- **severity**: `critical`
- **app**: `mssqlserver-alert-demo`, **app_namespace**: `alert-mssqlserver`
- **k8s_kind**: `MSSQLServer`
- **job**: `mssqlserver-alert-demo-stats`

> Note: this chart's alert labels use `app_namespace` rather than a plain `namespace` label — filter or group on `app_namespace` when searching for these alerts in AlertManager.

### 4. Restore MSSQLServer

Stop the loop from step 1.

```bash
$ kubectl get mssqlserver -n alert-mssqlserver mssqlserver-alert-demo -w
NAME                     VERSION    STATUS   AGE
mssqlserver-alert-demo   2025-cu0   Ready    24m
```

If MSSQLServer does not recover on its own within a minute or two, force a clean restart: `kubectl delete pod -n alert-mssqlserver mssqlserver-alert-demo-0`.

---

## Alert Reference

All alerts are scoped to the `mssqlserver-alert-demo` instance in the `alert-mssqlserver` namespace via the PromQL label filters `job="mssqlserver-alert-demo-stats"` / `namespace="alert-mssqlserver"` (database group), or `app="mssqlserver-alert-demo"` / `namespace="alert-mssqlserver"` (provisioner/opsManager groups).

### Database Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `MSSQLServerInstanceDown` | critical | instant | `up == 0` on this instance. |
| `MSSQLServerServiceDown` | critical | instant | No replica behind the service is answering. |
| `MSSQLServerRestarted` | critical | instant | Uptime indicates a recent restart. |
| `DiskUsageHigh` | warning | 1m | Persistent volume usage exceeds 80%. |
| `DiskAlmostFull` | critical | 1m | Persistent volume usage exceeds 95%. |

### Provisioner Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBMSSQLServerPhaseNotReady` | critical | 1m | KubeDB marked the MSSQLServer resource `NotReady`. |
| `KubeDBMSSQLServerPhaseCritical` | warning | 15m | MSSQLServer is degraded but not fully unavailable. |

### OpsManager Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBMSSQLServerOpsRequestStatusProgressingToLong` | critical | 30m | An ops request has been running for 30+ minutes. |
| `KubeDBMSSQLServerOpsRequestFailed` | critical | instant | An ops request failed. |

### KubeStash Group (disabled in this tutorial)

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `MSSQLServerKubeStashBackupSessionFailed` | critical | instant | Most recent backup session failed. |
| `MSSQLServerKubeStashRestoreSessionFailed` | critical | instant | Most recent restore session failed. |
| `MSSQLServerKubeStashNoBackupSessionForTooLong` | warning | instant | No recent successful backup. |
| `MSSQLServerKubeStashRepositoryCorrupted` | critical | 5m | Backup repository integrity check failed. |
| `MSSQLServerKubeStashRepositoryStorageRunningLow` | warning | 5m | Backup repository storage usage is high. |

---

## Customising Alerts

```yaml
# custom-alerts.yaml
form:
  alert:
    labels:
      release: prometheus
    groups:
      kubeStash:
        enabled: "none"
      database:
        enabled: warning
        rules:
          mssqlserverRestarted:
            enabled: true
            severity: warning
```

```bash
$ helm upgrade mssqlserver-alert-demo appscode/mssqlserver-alerts \
    -n alert-mssqlserver \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

To remove all resources created in this tutorial, run the following commands.

```bash
# Remove the mssqlserver-alerts release (PrometheusRule + dashboard-import Job)
$ helm uninstall mssqlserver-alert-demo -n alert-mssqlserver

# Remove the imported Grafana dashboard (it is not removed by helm uninstall)
$ curl -s -X DELETE -H "Authorization: Bearer <grafana-token>" \
    http://localhost:3000/api/dashboards/uid/<uid>

$ kubectl delete mssqlserver -n alert-mssqlserver mssqlserver-alert-demo
$ kubectl delete issuer -n alert-mssqlserver mssqlserver-ca-issuer
$ kubectl delete secret -n alert-mssqlserver mssqlserver-ca
$ kubectl delete ns alert-mssqlserver

# Uninstall monitoring stack (optional — skip if other tutorials on this cluster still need them)
$ helm uninstall panopticon -n kubeops
$ helm uninstall prometheus -n monitoring
```

## Next Steps

- Monitor your MSSQLServer instance with KubeDB using [Prometheus operator](/docs/guides/mssqlserver/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
