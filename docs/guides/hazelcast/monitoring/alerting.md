---
title: Hazelcast Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: hz-monitoring-alerting
    name: Alerting
    parent: hz-monitoring-hazelcast
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Hazelcast Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed Hazelcast instance using the `hazelcast-alerts` Helm chart.

## Before You Begin

* Ensure you have a Kubernetes cluster and that `kubectl` is configured to communicate with it. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Install the KubeDB operator by following the steps [here](/docs/setup/README.md).

* Deploy the database in the `alert-hazelcast` namespace:

  ```bash
  $ kubectl create ns alert-hazelcast
  namespace/alert-hazelcast created
  ```

* Hazelcast requires an Enterprise license. Create a secret with your license key before deploying:

  ```bash
  $ kubectl create secret generic hz-license-key -n alert-hazelcast \
      --from-literal=licenseKey='your hazelcast license key'
  secret/hz-license-key created
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

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/hazelcast/monitoring/overview.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/hazelcast](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hazelcast) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

- **KubeDB** deploys Hazelcast with a metrics-exporter sidecar (container `exporter`) that exposes Hazelcast's own JMX-derived metrics (`com_hazelcast_Metrics_*`).
- **ServiceMonitor** (named `{hazelcast-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape the exporter every 10 seconds.
- **PrometheusRule** is created by the `hazelcast-alerts` chart and contains alert definitions grouped by concern: database health (which also embeds KubeDB-operator-sourced `hazelcastDown`/`hazelcastPhaseCritical` alerts) and provisioner.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).
- **Grafana** dashboards for Hazelcast are covered separately — see [Grafana Dashboard](grafana-dashboard.md) rather than duplicated here.

---

## Deploy Hazelcast with Monitoring Enabled

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Hazelcast
metadata:
  name: hazelcast-alert-demo
  namespace: alert-hazelcast
spec:
  version: "5.5.2"
  replicas: 1
  licenseSecret:
    name: hz-license-key
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/monitoring/hazelcast-alert-demo.yaml
hazelcast.kubedb.com/hazelcast-alert-demo created
```

Wait for the database to go into `Ready` state.

```bash
$ kubectl get hazelcast -n alert-hazelcast hazelcast-alert-demo
NAME                    VERSION   STATUS   AGE
hazelcast-alert-demo    5.5.2     Ready    3m
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n alert-hazelcast --selector="app.kubernetes.io/instance=hazelcast-alert-demo"
NAME                              TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
hazelcast-alert-demo              ClusterIP   10.43.10.20    <none>        5701/TCP    3m
hazelcast-alert-demo-pods         ClusterIP   None           <none>        5701/TCP    3m
hazelcast-alert-demo-stats        ClusterIP   10.43.10.21    <none>        8080/TCP    3m
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n alert-hazelcast
NAME                           AGE
hazelcast-alert-demo-stats     3m

$ kubectl get servicemonitor -n alert-hazelcast hazelcast-alert-demo-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Install hazelcast-alerts

### Why the Helm release name matters

The chart derives the `PrometheusRule` name and scopes every PromQL expression from the **Helm release name** — so the release name must match the Hazelcast object's name (`hazelcast-alert-demo`).

### Install

```bash
$ helm upgrade -i hazelcast-alert-demo oci://ghcr.io/appscode-charts/hazelcast-alerts \
    -n alert-hazelcast \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus
```

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-hazelcast
NAME                     AGE
hazelcast-alert-demo     30s

$ kubectl get prometheusrule -n alert-hazelcast hazelcast-alert-demo \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Confirm Prometheus loaded the rules

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules` and locate the `hazelcast.database` and `hazelcast.provisioner` groups.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/hazelcast/monitoring/hazelcast-alerting-prom-rules.png" style="padding:10px">
</p>

Both groups should show **OK**. `hazelcast-alerts` v2026.7.14 has no `opsManager`/`stash`/`kubeStash` groups at all — only `database` and `provisioner`.

> **Note the overlap:** the `database` group's `hazelcastDown` (`for: 30s`) and `hazelcastPhaseCritical` (`for: 3m`) key off the exact same `kubedb_com_hazelcast_status_phase` metric as the `provisioner` group's `KubeDBhazelcastPhaseNotReady`/`KubeDBhazelcastPhaseCritical` (`for: 1m`/`15m`) — a real outage fires **both** pairs of alerts (at different times, since the `for` windows differ), not a bug exactly, but worth knowing so you don't mistake it for two independent problems.

---

## Verify End-to-End

### 1. Check the Prometheus target is UP

Open `http://localhost:9090/query?g0.expr=up%7Bnamespace%3D%22alert-hazelcast%22%7D&g0.tab=1`.

<p align="center">
  <img alt="Prometheus up query — hazelcast-alert-demo-0 UP" src="/docs/images/hazelcast/monitoring/hazelcast-alerting-prom-target.png" style="padding:10px">
</p>

### 2. Confirm the Hazelcast alerts are inactive

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — Hazelcast groups inactive" src="/docs/images/hazelcast/monitoring/hazelcast-alerting-prom-alerts.png" style="padding:10px">
</p>

All rules should show **INACTIVE**.

### 3. Check AlertManager

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager" src="/docs/images/hazelcast/monitoring/hazelcast-alerting-alertmanager.png" style="padding:10px">
</p>

### 4. Grafana dashboard

See [Grafana Dashboard](grafana-dashboard.md) for how to provision and explore the Hazelcast dashboards (via the `kubedb-grafana-dashboards` chart, `--set featureGates.Hazelcast=true`).

---

## Simulating a Firing Alert

This section deliberately triggers `hazelcastDown` (`for: 30s`, the fastest down-signal) by crashing the main Hazelcast JVM process.

### 1. Crash the Hazelcast process

```bash
$ kubectl exec -n alert-hazelcast hazelcast-alert-demo-0 -c hazelcast -- sh -c '
    end=$(( $(date +%s) + 60 ));
    while [ $(date +%s) -lt $end ]; do
      pid=$(pgrep -f "java.*hazelcast" | head -1);
      [ -n "$pid" ] && kill -9 "$pid" 2>/dev/null;
      sleep 1;
    done'
```

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — hazelcastDown Firing" src="/docs/images/hazelcast/monitoring/hazelcast-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

`hazelcastDown` (`kubedb_com_hazelcast_status_phase{phase!="Ready"} == 1`, `for: 30s`) should transition to **FIRING** first; if the crash loop runs long enough, `KubeDBhazelcastPhaseNotReady` (`for: 1m`, provisioner group) fires shortly after.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager — hazelcastDown Firing" src="/docs/images/hazelcast/monitoring/hazelcast-alerting-alertmanager-firing.png" style="padding:10px">
</p>

### 4. Restore Hazelcast

Stop the loop from step 1.

```bash
$ kubectl get hazelcast -n alert-hazelcast hazelcast-alert-demo -w
NAME                    VERSION   STATUS   AGE
hazelcast-alert-demo    5.5.2     Ready    24m
```

If Hazelcast does not recover on its own within a minute or two, force a clean restart: `kubectl delete pod -n alert-hazelcast hazelcast-alert-demo-0`.

---

## Alert Reference

All alerts are scoped to the `hazelcast-alert-demo` instance in the `alert-hazelcast` namespace, mostly via `namespace`/`service` label filters matching `$app-stats` (database group), or `app="hazelcast-alert-demo"` / `namespace="alert-hazelcast"` (provisioner group and the two operator-phase alerts embedded in the database group).

### Database Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `hazelcastPartitionCountExceed` | warning | 30s | Active partition count is unusually high. |
| `hazelcastHighHeapPercentage` | warning | 30s | JVM heap usage is high. |
| `hazelcastHighMemoryUsage` | warning | 30s | Hazelcast memory usage is high. |
| `hazelcastHighPhysicalMemoryUsage` | warning | 30s | Physical memory usage is high relative to total. |
| `hazelcastHighLatency` | warning | 30s | Get-operation latency is elevated. |
| `hazelcastSystemCPULoadExceed` | warning | 30s | System CPU load is high. |
| `hazelcastPhaseCritical` | warning | 3m | KubeDB operator view: resource `Critical` (duplicates the provisioner group's own version at a different `for`). |
| `hazelcastDown` | critical | 30s | KubeDB operator view: resource not `Ready`. Fastest down-signal available. |
| `DiskUsageHigh` | warning | 1m | Persistent volume usage exceeds 80%. |
| `DiskAlmostFull` | critical | 1m | Persistent volume usage exceeds 95%. |

### Provisioner Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBhazelcastPhaseNotReady` | critical | 1m | KubeDB marked the Hazelcast resource `NotReady`. |
| `KubeDBhazelcastPhaseCritical` | warning | 15m | Hazelcast is degraded but not fully unavailable. |

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
          hazelcastHighHeapPercentage:
            enabled: true
            duration: "2m"
            severity: warning
```

```bash
$ helm upgrade hazelcast-alert-demo oci://ghcr.io/appscode-charts/hazelcast-alerts \
    -n alert-hazelcast \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

```bash
$ helm uninstall hazelcast-alert-demo -n alert-hazelcast
$ kubectl delete hazelcast -n alert-hazelcast hazelcast-alert-demo
$ kubectl delete secret -n alert-hazelcast hz-license-key
$ kubectl delete ns alert-hazelcast
```

## Next Steps

- Monitor your Hazelcast instance with KubeDB using [built-in Prometheus](/docs/guides/hazelcast/monitoring/prometheus-builtin.md).
- Monitor your Hazelcast instance with KubeDB using [Prometheus operator](/docs/guides/hazelcast/monitoring/prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
