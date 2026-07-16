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

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed Microsoft SQL Server instance using the `mssqlserver-alerts` Helm chart.

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

* Before proceeding, complete the [Configuration](grafana-dashboard.md#configuration) steps to deploy **kube-prometheus-stack** and **Panopticon**.

* This tutorial assumes you already have a **kube-prometheus-stack** running in your cluster, with `Prometheus` configured so that both `serviceMonitorSelector` and `ruleSelector` match the label `release: prometheus`.

  To verify the selectors:

  ```bash
  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.ruleSelector}'
  {"matchLabels":{"release":"prometheus"}}

  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.serviceMonitorSelector}'
  {"matchLabels":{"release":"prometheus"}}
  ```

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/mssqlserver/monitoring/overview.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/mssqlserver](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

- **KubeDB** deploys MSSQLServer with a metrics-exporter sidecar (container `exporter`) that exposes metrics on the `{mssqlserver-name}-stats` service.
- **ServiceMonitor** (named `{mssqlserver-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape the exporter every 10 seconds.
- **PrometheusRule** is created by the `mssqlserver-alerts` chart and contains MSSQLServer alert definitions grouped by concern: database health, provisioner, and ops-manager. A fourth group (`kubeStash`) is declared in the chart but currently fails to render — see the chart-bug callout below.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).
- **Grafana** dashboards for MSSQLServer are covered separately — see [Grafana Dashboard](grafana-dashboard.md) rather than duplicated here.

---

## Deploy MSSQLServer with Monitoring Enabled

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssqlserver-alert-demo
  namespace: alert-mssqlserver
spec:
  version: "2022-cu12"
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

`ACCEPT_EULA=Y` and `MSSQL_PID` are required by the upstream SQL Server image itself, not KubeDB — see [Microsoft's environment variable reference](https://learn.microsoft.com/en-us/sql/linux/sql-server-linux-configure-environment-variables) for valid `MSSQL_PID` values.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/monitoring/mssqlserver-alert-demo.yaml
mssqlserver.kubedb.com/mssqlserver-alert-demo created
```

Wait for the database to go into `Ready` state.

```bash
$ kubectl get mssqlserver -n alert-mssqlserver mssqlserver-alert-demo
NAME                       VERSION      STATUS   AGE
mssqlserver-alert-demo     2022-cu12    Ready    5m
```

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

## Step 1 — Install mssqlserver-alerts

### Why the Helm release name matters

The chart derives the `PrometheusRule` name and scopes every PromQL expression (via `job="{release-name}-stats"` / `app="{release-name}"`) from the **Helm release name** — so the release name must match the MSSQLServer object's name (`mssqlserver-alert-demo`).

### Chart bug: the `kubeStash` group fails to render

> **Chart bug found (v2026.7.14):** `helm template`/`helm install` for `mssqlserver-alerts` fails outright with `error converting YAML to JSON: yaml: line 149: mapping values are not allowed in this context` if the `kubeStash` alert group is left at its default `enabled: warning`. The group's rule template has a real indentation bug — the `labels:`/`annotations:` block for each `kubeStash` rule is emitted one level shallower than required, so keys like `k8s_group`, `k8s_kind`, `app` end up as siblings of `labels:` instead of nested under it, breaking YAML parsing entirely. **Workaround:** disable the group at install time with `--set form.alert.groups.kubeStash.enabled=none` — the rest of the chart (`database`, `provisioner`, `opsManager`) renders correctly once this group is disabled.

### Install

```bash
$ helm upgrade -i mssqlserver-alert-demo oci://ghcr.io/appscode-charts/mssqlserver-alerts \
    -n alert-mssqlserver \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus \
    --set form.alert.groups.kubeStash.enabled=none
```

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-mssqlserver
NAME                       AGE
mssqlserver-alert-demo     30s

$ kubectl get prometheusrule -n alert-mssqlserver mssqlserver-alert-demo \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Confirm Prometheus loaded the rules

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules` and locate the `mssqlserver.database`, `mssqlserver.provisioner`, and `mssqlserver.opsManager` groups.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/mssqlserver/monitoring/mssqlserver-alerting-prom-rules.png" style="padding:10px">
</p>

All three groups should show **OK** (with `kubeStash` absent, per the workaround above).

---

## Verify End-to-End

### 1. Check the Prometheus target is UP

Open `http://localhost:9090/query?g0.expr=up%7Bnamespace%3D%22alert-mssqlserver%22%7D&g0.tab=1`.

<p align="center">
  <img alt="Prometheus up query — mssqlserver-alert-demo-0 UP" src="/docs/images/mssqlserver/monitoring/mssqlserver-alerting-prom-target.png" style="padding:10px">
</p>

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

### 4. Grafana dashboard

See [Grafana Dashboard](grafana-dashboard.md) for how to provision and explore the MSSQLServer dashboards (via the `kubedb-grafana-dashboards` chart, `--set featureGates.MSSQLServer=true`).

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

### 4. Restore MSSQLServer

Stop the loop from step 1.

```bash
$ kubectl get mssqlserver -n alert-mssqlserver mssqlserver-alert-demo -w
NAME                     VERSION     STATUS   AGE
mssqlserver-alert-demo   2022-cu12   Ready    24m
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

### KubeStash Group (currently broken — see chart-bug callout above)

Declared in `values.yaml` but not currently installable at chart v2026.7.14 without disabling it.

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
        enabled: "none"   # required until the chart's indentation bug above is fixed upstream
      database:
        enabled: warning
        rules:
          mssqlserverRestarted:
            enabled: true
            severity: warning
```

```bash
$ helm upgrade mssqlserver-alert-demo oci://ghcr.io/appscode-charts/mssqlserver-alerts \
    -n alert-mssqlserver \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

> Once the upstream chart fixes the `kubeStash` group's indentation, re-check whether `enabled: "none"` is still necessary before upgrading further.

---

## Cleaning up

```bash
$ helm uninstall mssqlserver-alert-demo -n alert-mssqlserver
$ kubectl delete mssqlserver -n alert-mssqlserver mssqlserver-alert-demo
$ kubectl delete issuer -n alert-mssqlserver mssqlserver-ca-issuer
$ kubectl delete secret -n alert-mssqlserver mssqlserver-ca
$ kubectl delete ns alert-mssqlserver
```

## Next Steps

- Monitor your MSSQLServer instance with KubeDB using [Prometheus operator](/docs/guides/mssqlserver/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
