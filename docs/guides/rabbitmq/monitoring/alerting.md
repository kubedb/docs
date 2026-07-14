---
title: RabbitMQ Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: rm-monitoring-alerting
    name: Alerting
    parent: rm-monitoring-guides
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RabbitMQ Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed RabbitMQ instance using the `rabbitmq-alerts` Helm chart.

## Before You Begin

* Ensure you have a Kubernetes cluster and that `kubectl` is configured to communicate with it. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Install the KubeDB operator by following the steps [here](/docs/setup/README.md).

* Create a dedicated namespace to deploy the database:

  ```bash
  $ kubectl create ns alert-rabbitmq
  namespace/alert-rabbitmq created
  ```

* This tutorial assumes you already have a **kube-prometheus-stack** running in your cluster, with `Prometheus` configured so that both `serviceMonitorSelector` and `ruleSelector` match the label `release: prometheus`. See the [Grafana Dashboard](grafana-dashboard.md) guide for how to deploy kube-prometheus-stack if you don't have it yet.

  To verify the selectors:

  ```bash
  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.ruleSelector}'
  {"matchLabels":{"release":"prometheus"}}

  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.serviceMonitorSelector}'
  {"matchLabels":{"release":"prometheus"}}
  ```

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/rabbitmq/monitoring/overview.md).

* For dashboards and visualisation, see [Grafana Dashboard](grafana-dashboard.md) for RabbitMQ.

> Note: YAML files used in this tutorial are stored in [docs/examples/rabbitmq](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/rabbitmq) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

- **KubeDB** deploys RabbitMQ with the built-in `rabbitmq_prometheus` plugin enabled, which serves metrics directly from the `rabbitmq` container on port `15692` — unlike some other databases, RabbitMQ needs no separate exporter sidecar.
- **ServiceMonitor** (named `{rabbitmq-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape the metrics endpoint every 10 seconds.
- **KubeDB operator (panopticon)** also exposes the CR's own status as a metric, `kubedb_com_rabbitmq_status_phase`. The `RabbitMQDown` and provisioner-group alerts key off this metric instead of the database's own stats endpoint, so they fire purely based on what KubeDB itself observes about the resource — even if the metrics scrape target is otherwise healthy.
- **PrometheusRule** is created by the `rabbitmq-alerts` chart and contains RabbitMQ alert definitions grouped by concern: database health and provisioner.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).

---

## Deploy RabbitMQ with Monitoring Enabled

At first, let's deploy a RabbitMQ database with monitoring enabled. Below is the RabbitMQ object we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rmq-alert-demo
  namespace: alert-rabbitmq
spec:
  version: "4.0.4"
  replicas: 1
  deletionPolicy: WipeOut
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        interval: 10s
        labels:
          release: prometheus
```

Here,

- `spec.monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` resource managed by the Prometheus operator.
- `spec.monitor.prometheus.serviceMonitor.labels.release: prometheus` adds the `release: prometheus` label to the created `ServiceMonitor`, matching the Prometheus `serviceMonitorSelector` so the target is discovered automatically.

Let's create the RabbitMQ resource.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/monitoring/rmq-alert-demo.yaml
rabbitmq.kubedb.com/rmq-alert-demo created
```

Now, wait for the database to go into `Ready` state.

```bash
$ kubectl get rabbitmq -n alert-rabbitmq rmq-alert-demo
NAME             VERSION   STATUS   AGE
rmq-alert-demo   4.0.4     Ready    49s
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n alert-rabbitmq --selector="app.kubernetes.io/instance=rmq-alert-demo"
NAME                         TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                                           AGE
rmq-alert-demo               ClusterIP   10.43.174.187   <none>        5672/TCP,1883/TCP,61613/TCP,15675/TCP,15674/TCP   59s
rmq-alert-demo-dashboard     ClusterIP   10.43.236.58    <none>        15672/TCP                                         59s
rmq-alert-demo-pods          ClusterIP   None            <none>        4369/TCP,25672/TCP                                59s
rmq-alert-demo-stats         ClusterIP   10.43.128.229   <none>        15692/TCP                                         59s
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n alert-rabbitmq
NAME                    AGE
rmq-alert-demo-stats    55s
```

Verify that the `ServiceMonitor` carries the `release: prometheus` label so Prometheus discovers it.

```bash
$ kubectl get servicemonitor -n alert-rabbitmq rmq-alert-demo-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Install rabbitmq-alerts

The `rabbitmq-alerts` chart creates a `PrometheusRule` resource containing RabbitMQ alert definitions grouped by concern: database health and provisioner.

### Why the Helm release name matters

The chart derives the PromQL `job`/instance scoping (and the `PrometheusRule` name) from the **Helm release name**, not from a values field — so the release name must match the RabbitMQ object's name (`rmq-alert-demo`) for the rules to be correctly scoped to this instance.

The chart's default label is `release: kube-prometheus-stack`, so we must also override it at install time to match the Prometheus `ruleSelector`.

### Install

```bash
$ helm upgrade -i rmq-alert-demo oci://ghcr.io/appscode-charts/rabbitmq-alerts \
    -n alert-rabbitmq \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus
```

| Flag | Value | Purpose |
|------|-------|---------|
| `rmq-alert-demo` (release name) | — | Scopes every PromQL expression to this instance (`job="rmq-alert-demo-stats"`, `app="rmq-alert-demo"`) |
| `-n alert-rabbitmq` | `alert-rabbitmq` | Installs the `PrometheusRule` in the same namespace as the database |
| `form.alert.labels.release` | `prometheus` | Matches the Prometheus `ruleSelector` so the rules are loaded |

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-rabbitmq
NAME             AGE
rmq-alert-demo   30s
```

Confirm the `release: prometheus` label is present.

```bash
$ kubectl get prometheusrule -n alert-rabbitmq rmq-alert-demo \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Confirm Prometheus loaded the rules

Port-forward the Prometheus UI and open the **Status → Rule health** page.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules?search=rabbitmq`.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/rabbitmq/monitoring/rmq-alerting-prom-rules.png" style="padding:10px">
</p>

The `rabbitmq.database.alert-rabbitmq.rmq-alert-demo.rules` group is visible with all rules showing **OK**, confirming that Prometheus has loaded and is evaluating the RabbitMQ alert definitions every 30 seconds.

---

## Verify End-to-End

### 1. Check the metrics endpoint is reachable

Because RabbitMQ's Prometheus plugin runs inside the `rabbitmq` container itself, there is no separate `exporter` container to check — query the plugin's own endpoint directly.

```bash
$ kubectl exec -n alert-rabbitmq rmq-alert-demo-0 -c rabbitmq -- \
    wget -qO- http://127.0.0.1:15692/metrics | grep rabbitmq_identity_info
rabbitmq_identity_info{rabbitmq_node="rabbit@rmq-alert-demo-0.rmq-alert-demo-pods.alert-rabbitmq",rabbitmq_cluster="rmq-alert-demo",rabbitmq_cluster_permanent_id="rabbitmq-cluster-id-HcbvDHSZhF_lBd8K4iWRkQ"} 1
```

### 2. Check the Prometheus target is UP

Open `http://localhost:9090/targets?search=rmq-alert-demo`.

<p align="center">
  <img alt="Prometheus Target UP" src="/docs/images/rabbitmq/monitoring/rmq-alerting-prom-target.png" style="padding:10px">
</p>

The target `serviceMonitor/alert-rabbitmq/rmq-alert-demo-stats/0` shows **UP**, confirming metrics are being scraped from `rmq-alert-demo-0` in the `alert-rabbitmq` namespace.

### 3. Confirm the RabbitMQ alerts

Open `http://localhost:9090/alerts?search=rabbitmq` to see the RabbitMQ alert groups.

<p align="center">
  <img alt="Prometheus Alerts" src="/docs/images/rabbitmq/monitoring/rmq-alerting-prom-alerts.png" style="padding:10px">
</p>

9 of the 11 rules in the `rabbitmq.database` group show **INACTIVE**, meaning the cluster is healthy and no thresholds are breached. `DiskUsageHigh` and `DiskAlmostFull` show **PENDING** here — this demo cluster uses the `local-path` storage class, whose PVCs are just directories on the node's root filesystem, so `kubelet_volume_stats_used_bytes` reflects the **node's actual disk usage** rather than a small, isolated volume. On a node with real dedicated storage for the PVC, these two rules will normally sit **INACTIVE** just like the rest.

### 4. Check AlertManager

Port-forward AlertManager to view any currently firing alerts.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`. **PENDING** rules have not yet fired — only alerts that cross into **FIRING** are forwarded to AlertManager — so with a healthy RabbitMQ instance no alerts for `rmq-alert-demo` are listed here yet.

---

## Simulating a Firing Alert

The previous section confirmed that the RabbitMQ alerts are healthy. This section walks through deliberately triggering the `RabbitMQDown` critical alert so you can observe the full alert lifecycle — from firing in Prometheus through to the AlertManager dashboard — and then resolve it.

### 1. Crash the RabbitMQ process

Kill the RabbitMQ process inside the pod. Unlike Memcached, a RabbitMQ pod has only a single `rabbitmq` container — the Prometheus plugin runs inside the same process being killed, so a lone `kill 1` restarts fast enough that the container becomes `Ready` again before KubeDB's health check can even observe the outage. Repeat the kill a few times over ~30–45 seconds to hold the pod in a crash loop long enough for the KubeDB operator to mark the RabbitMQ resource `NotReady` for the full evaluation window.

```bash
$ for i in $(seq 1 8); do
    kubectl exec -n alert-rabbitmq rmq-alert-demo-0 -c rabbitmq -- kill 1
    sleep 6
  done
```

Watch the CR phase move from `Ready` → `Critical` → `NotReady`:

```bash
$ kubectl get rabbitmq -n alert-rabbitmq rmq-alert-demo -o jsonpath='{.status.phase}'
NotReady
```

`RabbitMQDown` and `RabbitMQPhaseCritical` key off `kubedb_com_rabbitmq_status_phase` (a metric emitted by the KubeDB operator itself), so what matters is the CR's `status.phase`, not the exporter's own scrape health. Wait 30–60 seconds for the next rule-evaluation cycle (30 s) to register the failure once the phase settles on `NotReady`.

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts?search=rabbitmq`.

<p align="center">
  <img alt="Prometheus Alerts — RabbitMQDown Firing" src="/docs/images/rabbitmq/monitoring/rmq-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

`RabbitMQDown` moves from **INACTIVE** to **FIRING** once its `for: 30s` window elapses with the phase held at `NotReady`. The provisioner-group `KubeDBRabbitMQPhaseNotReady` alert (`for: 1m`) fires for the same reason shortly after.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093/#/alerts?filter=%7Bnamespace%3D%22alert-rabbitmq%22%7D`.

<p align="center">
  <img alt="AlertManager — RabbitMQDown Firing" src="/docs/images/rabbitmq/monitoring/rmq-alerting-alertmanager-firing.png" style="padding:10px">
</p>

AlertManager shows both `RabbitMQDown` and `KubeDBRabbitMQPhaseNotReady`. The alert cards display:

- **Severity**: `critical`
- **rabbitmq**: `rmq-alert-demo` in the `alert-rabbitmq` namespace
- **phase**: `NotReady`
- **Started**: timestamp when the alert first fired

Note that the `instance`/`pod`/`job` labels on these two alerts point at the KubeDB operator's **panopticon** component (e.g. `job="panopticon"`), not at the RabbitMQ pod itself — because these alerts are derived from the operator's own status metric rather than from the database's stats endpoint.

AlertManager routes this alert to every receiver configured in your `alertmanagerConfig` (Slack, email, PagerDuty, webhook, etc.) based on your routing tree. If no receiver is configured, the alert is visible here but silently dropped.

### 4. Restore RabbitMQ

Delete the pod so KubeDB recreates it cleanly.

```bash
$ kubectl delete pod -n alert-rabbitmq rmq-alert-demo-0
pod "rmq-alert-demo-0" deleted
```

Once `status.phase` returns to `Ready`, Prometheus marks both alerts **INACTIVE** again and AlertManager sends a **resolved** notification to all receivers.

---

## Alert Reference

All alerts are scoped to the `rmq-alert-demo` instance in the `alert-rabbitmq` namespace via the PromQL label filters `job="rmq-alert-demo-stats"` / `app="rmq-alert-demo"` and `namespace="alert-rabbitmq"`.

### Database Group

Fired based on live metrics from RabbitMQ's `rabbitmq_prometheus` plugin, plus the two persistent-volume rules and the two KubeDB status-phase rules.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `RabbitmqFileDescriptorsNearLimit` | warning | 30s | More than 80% of the node's file descriptor limit is in use — at 100%, new connections will be refused and disk writes may fail. |
| `RabbitmqQueueIsGrowing` | warning | 30s | A queue's message count has been steadily increasing over the last 10 minutes — consumers may not be keeping up with publishers. |
| `RabbitmqUnroutableMessages` | warning | 30s | Messages published to an exchange could not be routed to any queue in the last 5 minutes — check your exchange/queue bindings. |
| `RabbitmqTCPSocketsNearLimit` | warning | 30s | More than 80% of the node's TCP socket limit is in use — at 100%, new connections will be refused. |
| `RabbitmqLowDiskWatermarkPredicted` | warning | 30s | Based on the last 24h trend, free disk space is predicted to drop below the configured watermark within 24 hours, which would block all publishers cluster-wide. |
| `RabbitmqInsufficientEstablishedErlangDistributionLinks` | warning | 30s | Fewer Erlang distribution links than expected for a full-mesh cluster are established — indicates partial inter-node connectivity issues. |
| `RabbitmqHighConnectionChurn` | warning | 30s | More than 10% of total connections were opened/closed per second over the last 5 minutes — client connections are short-lived instead of long-lived. |
| `RabbitMQPhaseCritical` | warning | 3m | KubeDB reports the database in Critical phase — one or more nodes are down, but read/write is not yet hampered. |
| `RabbitMQDown` | critical | 30s | KubeDB reports the database in NotReady phase — the cluster is not accepting connections and read/write is failing. |
| `DiskUsageHigh` | warning | 1m | The RabbitMQ data volume (PVC) is more than 80% full. |
| `DiskAlmostFull` | critical | 1m | The RabbitMQ data volume (PVC) is more than 95% full. |

### Provisioner Group

Monitors the KubeDB operator's view of the RabbitMQ resource phase.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBRabbitMQPhaseNotReady` | critical | 1m | KubeDB marked the RabbitMQ resource `NotReady` — operator cannot reach the database. |
| `KubeDBRabbitMQPhaseCritical` | warning | 15m | The instance is in a degraded/critical phase. |

### OpsManager Group

The chart's `values.yaml` also ships an `opsManager` group (`opsRequestOnProgress`, `opsRequestStatusProgressingToLong`, `opsRequestFailed`), intended to track `RabbitMQOpsRequest` lifecycle during upgrades, scaling, and reconfiguration — the same pattern used by other `*-alerts` charts.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `opsRequestOnProgress` | info | instant | An ops request is currently in progress. |
| `opsRequestStatusProgressingToLong` | critical | 30m | An ops request has been running for 30+ minutes — likely stuck. |
| `opsRequestFailed` | critical | instant | An ops request failed — check the `RabbitMQOpsRequest` object for the error. |

> **Verified note:** in chart version `v2026.7.14`, `templates/alert.yaml` only renders the `database` and `provisioner` groups into the `PrometheusRule` — `kubectl get prometheusrule -n alert-rabbitmq rmq-alert-demo -o yaml` shows no `opsManager` rules, even though `opsManager.enabled` is `warning` by default in `values.yaml`. If you rely on ops-request alerting, verify with `kubectl get prometheusrule -n <namespace> <release> -o yaml` after installing, and check for a newer chart version if you need these rules.

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
          rabbitmqHighConnectionChurn:
            enabled: true
            duration: "2m"
            severity: warning
          diskUsageHigh:
            enabled: true
            val: 90        # fire at 90% disk usage instead of the default 80%
            duration: "5m"
            severity: warning
      provisioner:
        enabled: "none"    # disable all provisioner alerts
```

```bash
$ helm upgrade rmq-alert-demo oci://ghcr.io/appscode-charts/rabbitmq-alerts \
    -n alert-rabbitmq \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

To remove all resources created in this tutorial, run the following commands.

```bash
# Remove the rabbitmq-alerts release
$ helm uninstall rmq-alert-demo -n alert-rabbitmq

# Remove the RabbitMQ instance
$ kubectl delete rabbitmq -n alert-rabbitmq rmq-alert-demo

# Delete namespace
$ kubectl delete ns alert-rabbitmq
```

## Next Steps

- Monitor your RabbitMQ database with KubeDB using [builtin Prometheus](/docs/guides/rabbitmq/monitoring/using-builtin-prometheus.md).
- Monitor your RabbitMQ database with KubeDB using [Prometheus operator](/docs/guides/rabbitmq/monitoring/using-prometheus-operator.md).
- Visualise RabbitMQ metrics with [Grafana Dashboard](grafana-dashboard.md).
- Detail concepts of [RabbitMQ object](/docs/guides/rabbitmq/concepts/rabbitmq.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
