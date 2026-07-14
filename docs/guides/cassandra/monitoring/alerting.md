---
title: Cassandra Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: cas-monitoring-alerting
    name: Alerting
    parent: cas-monitoring-cassandra
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Cassandra Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed Cassandra cluster using the `cassandra-alerts` Helm chart.

## Before You Begin

* Ensure you have a Kubernetes cluster and that `kubectl` is configured to communicate with it. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Install the KubeDB operator by following the steps [here](/docs/setup/README.md).

* Deploy the database in the `demo` namespace:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

* This tutorial assumes you already have a **kube-prometheus-stack** running in your cluster, with `Prometheus` configured so that both `serviceMonitorSelector` and `ruleSelector` match the label `release: prometheus`. See the [Grafana Dashboard](grafana-dashboard.md#configuration) guide for how to deploy kube-prometheus-stack if you don't have it yet.

  To verify the selectors:

  ```bash
  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.ruleSelector}'
  {"matchLabels":{"release":"prometheus"}}

  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.serviceMonitorSelector}'
  {"matchLabels":{"release":"prometheus"}}
  ```

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/cassandra/monitoring/overview.md).

* For dashboards and visualisation, see [Grafana Dashboard](grafana-dashboard.md) for Cassandra.

> Note: YAML files used in this tutorial are stored in [docs/examples/cassandra](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/cassandra) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

- **KubeDB** deploys Cassandra with the [JMX Exporter](https://github.com/prometheus/jmx_exporter)-based `exporter` sidecar container, which scrapes the Cassandra JVM's JMX metrics and exposes them as Prometheus metrics on port `8080`.
- **Stats Service** (named `{cassandra-name}-stats`) is created automatically by KubeDB and fronts the exporter's metrics endpoint on port `56790`, which is proxied to the exporter's actual listening port `8080` inside the pod.
- **ServiceMonitor** (named `{cassandra-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape the exporter every 10 seconds.
- **PrometheusRule** is created by the `cassandra-alerts` chart and contains all Cassandra alert definitions grouped by concern: database health and provisioner.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).

---

## Deploy Cassandra with Monitoring Enabled

At first, let's deploy a Cassandra cluster with monitoring enabled. Below is the Cassandra object we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:
  name: cas-alert-demo
  namespace: demo
spec:
  version: 5.0.3
  topology:
    rack:
      - name: r0
        replicas: 2
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
        storageType: Durable
  deletionPolicy: WipeOut
  monitor:
    agent: "prometheus.io/operator"
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

Here,

- `spec.topology.rack[].replicas: 2` — Cassandra requires more than one replica per rack for its admission webhook to accept the object, so this demo uses the smallest allowed topology (1 rack, 2 replicas).
- `spec.monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` resource managed by the Prometheus operator.
- `spec.monitor.prometheus.serviceMonitor.labels.release: prometheus` adds the `release: prometheus` label to the created `ServiceMonitor`, matching the Prometheus `serviceMonitorSelector` so the target is discovered automatically.

Let's create the Cassandra resource.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/monitoring/cas-alert-demo.yaml
cassandra.kubedb.com/cas-alert-demo created
```

Now, wait for the database to go into `Ready` state.

```bash
$ kubectl get cassandra -n demo cas-alert-demo
NAME             VERSION   STATUS   AGE
cas-alert-demo   5.0.3     Ready    2m
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=cas-alert-demo"
NAME                          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                               AGE
cas-alert-demo                ClusterIP   10.43.49.40     <none>        9042/TCP,7000/TCP,7199/TCP,7001/TCP   2m
cas-alert-demo-rack-r0-pods   ClusterIP   None            <none>        9042/TCP,7000/TCP,7199/TCP,7001/TCP   2m
cas-alert-demo-stats          ClusterIP   10.43.160.146   <none>        56790/TCP                             2m
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n demo
NAME                   AGE
cas-alert-demo-stats   2m
```

Verify that the `ServiceMonitor` carries the `release: prometheus` label so Prometheus discovers it.

```bash
$ kubectl get servicemonitor -n demo cas-alert-demo-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Install cassandra-alerts

The `cassandra-alerts` chart creates a `PrometheusRule` resource containing all Cassandra alert definitions grouped by concern: database health and provisioner.

### Why the Helm release name matters

The chart derives the PromQL `job`/instance scoping (and the `PrometheusRule` name) from the **Helm release name**, not from a values field — so the release name must match the Cassandra object's name (`cas-alert-demo`) for the rules to be correctly scoped to this instance.

The chart's default label is `release: kube-prometheus-stack`, so we must also override it at install time to match the Prometheus `ruleSelector`.

> **Note:** The chart's default values leave `form.alert.groups.database.rules.cassandraDown.val` unset, which renders an incomplete PromQL expression (`... > `) that the `PrometheusRule` admission webhook rejects. Explicitly set it to `0` at install time — the `CassandraDown` expression counts down instances and fires whenever that count is greater than `0`.

### Install

```bash
$ helm upgrade -i cas-alert-demo oci://ghcr.io/appscode-charts/cassandra-alerts \
    -n demo \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus \
    --set form.alert.groups.database.rules.cassandraDown.val=0
```

| Flag | Value | Purpose |
|------|-------|---------|
| `cas-alert-demo` (release name) | — | Scopes every PromQL expression to this instance (`job="cas-alert-demo-stats"`) |
| `-n demo` | `demo` | Installs the `PrometheusRule` in the same namespace as the database |
| `form.alert.labels.release` | `prometheus` | Matches the Prometheus `ruleSelector` so the rules are loaded |
| `form.alert.groups.database.rules.cassandraDown.val` | `0` | Fixes the chart's missing default so the `CassandraDown` expression is valid PromQL |

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n demo
NAME             AGE
cas-alert-demo   30s
```

Confirm the `release: prometheus` label is present.

```bash
$ kubectl get prometheusrule -n demo cas-alert-demo \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Confirm Prometheus loaded the rules

Port-forward the Prometheus UI and open the **Status → Rule health** page.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules?search=cassandra`.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/cassandra/monitoring/cas-alerting-prom-rules.png" style="padding:10px">
</p>

Both the `cassandra.database.demo.cas-alert-demo.rules` and `cassandra.provisioner.demo.cas-alert-demo.rules` groups are visible with all rules showing **OK**, confirming that Prometheus has loaded and is evaluating the Cassandra alert definitions every 30 seconds.

---

## Verify End-to-End

### 1. Check the exporter is running

The `exporter` sidecar inside the Cassandra pod scrapes JMX and serves metrics at `:8080/metrics`. A metric named `java:lang:runtime:uptime` confirms the exporter can reach the Cassandra JVM.

```bash
$ kubectl exec -n demo cas-alert-demo-rack-r0-0 -c cassandra -- \
    curl -s localhost:8080/metrics | grep 'name="java:lang:runtime:uptime"'
cassandra_stats{cluster="Test Cluster",datacenter="dc1",keyspace="",table="",name="java:lang:runtime:uptime",} 458763.0
```

> The `exporter` container's own image ships neither `curl` nor `wget`, so the check above runs from the `cassandra` container instead — both containers share the pod's network namespace, so `localhost:8080` reaches the exporter's HTTP server either way.

### 2. Check the Prometheus target is UP

Open `http://localhost:9090/targets?search=cas-alert-demo`.

<p align="center">
  <img alt="Prometheus Target UP" src="/docs/images/cassandra/monitoring/cas-alerting-prom-target.png" style="padding:10px">
</p>

The target `serviceMonitor/demo/cas-alert-demo-stats/0` shows **2 / 2 up**, confirming metrics are being scraped from both `cas-alert-demo-rack-r0-0` and `cas-alert-demo-rack-r0-1` in the `demo` namespace.

### 3. Confirm all Cassandra alerts are inactive

Open `http://localhost:9090/alerts?search=cassandra` to see the Cassandra alert groups.

<p align="center">
  <img alt="Prometheus Alerts — All Inactive" src="/docs/images/cassandra/monitoring/cas-alerting-prom-alerts.png" style="padding:10px">
</p>

All 6 rules in the `cassandra.database` group and both rules in the `cassandra.provisioner` group show **INACTIVE**, meaning the cluster is healthy and no thresholds are breached.

### 4. Check AlertManager

Port-forward AlertManager to view any currently firing alerts.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`. With a healthy Cassandra cluster, no alerts for `cas-alert-demo` will be listed here.

---

## Simulating a Firing Alert

The previous section confirmed that all alerts are **INACTIVE** while the cluster is healthy. This section walks through deliberately triggering the `CassandraDown` critical alert so you can observe the full alert lifecycle — from firing in Prometheus through to the AlertManager dashboard — and then resolve it.

### 1. Stop the metrics endpoint

Unlike some other KubeDB charts, `CassandraDown` is **not** driven by a custom exporter-reported gauge — it is built directly on Prometheus's own scrape-health metric, `up{job="cas-alert-demo-stats"}`. That metric only goes to `0` when Prometheus can no longer reach the scrape target's HTTP endpoint at all.

On this build, the exporter's HTTP server (the process actually scraped by Prometheus on port `8080`) runs inside the `exporter` container — killing the `cassandra` container alone leaves the exporter's HTTP endpoint reachable (Prometheus would keep scraping it successfully), so `up` would stay `1` and `CassandraDown` would never fire. To reproduce a real scrape failure, stop the `exporter` container's process instead:

```bash
$ kubectl exec -n demo cas-alert-demo-rack-r0-1 -c exporter -- kill 1
```

Kubernetes restarts the crashed `exporter` container in the background. The restart is quick, so the outage window can be short — if the alert resolves before you finish inspecting it, repeat the `kill 1` command a few times in a row; Kubernetes' crash-loop backoff will keep the container down for a longer stretch on subsequent attempts, giving you a wider window to observe the firing state.

Wait 30–60 seconds for the next Prometheus scrape cycle (configured at 10s) and rule-evaluation cycle (30s) to register the failure.

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts?search=cassandra`.

<p align="center">
  <img alt="Prometheus Alerts — CassandraDown Firing" src="/docs/images/cassandra/monitoring/cas-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

Because `CassandraDown` has `for: 0m` (instant), it moves directly from **INACTIVE** to **FIRING** within one evaluation cycle. Note that `KubeDBCassandraPhaseCritical` also enters **PENDING** shortly after, since the KubeDB operator observes the degraded rack.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager — CassandraDown Firing" src="/docs/images/cassandra/monitoring/cas-alerting-alertmanager-firing.png" style="padding:10px">
</p>

AlertManager shows the `CassandraDown` alert. The alert card displays labels including:

- **alertname**: `CassandraDown`
- **severity**: `critical`
- **app**: `cas-alert-demo`, **app_namespace**: `demo`
- **job**: `cas-alert-demo-stats`

> Note: this chart's alert labels use `app_namespace` rather than a plain `namespace` label — filter or group on `app_namespace` when searching for these alerts in AlertManager.

AlertManager routes this alert to every receiver configured in your `alertmanagerConfig` (Slack, email, PagerDuty, webhook, etc.) based on your routing tree. If no receiver is configured, the alert is visible here but silently dropped.

### 4. Restore Cassandra

Delete the pod so KubeDB recreates it cleanly.

```bash
$ kubectl delete pod -n demo cas-alert-demo-rack-r0-1
```

Once the exporter's `/metrics` endpoint is reachable again, Prometheus marks the alert **INACTIVE** and AlertManager sends a **resolved** notification to all receivers.

---

## Alert Reference

All alerts are scoped to the `cas-alert-demo` instance in the `demo` namespace via the PromQL label filters `job="cas-alert-demo-stats"` and `app_namespace="demo"`.

### Database Group

Fired based on live metrics from the Cassandra JMX exporter.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `CassandraDown` | critical | instant | The Prometheus scrape target for this instance is unreachable — the exporter's metrics endpoint is down or the pod is unreachable. |
| `CassandraServiceRespawn` | critical | instant | Cassandra restarted recently (JVM uptime < 180s). |
| `ConnectionTimeouts` | warning | instant | More than 100 connection timeouts observed in the last minute. |
| `DroppedMessages` | warning | instant | One or more internal Cassandra messages have been dropped — a sign of overload or backpressure. |
| `HighReadLatency` | warning | instant | 99th-percentile coordinator read latency on the health-check table exceeds 7000 (µs). |
| `HighWriteLatency` | warning | instant | 99th-percentile coordinator write latency on the health-check table exceeds 7000 (µs). |

### Provisioner Group

Monitors the KubeDB operator's view of the Cassandra resource phase.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBCassandraPhaseNotReady` | critical | 1m | KubeDB marked the Cassandra resource `NotReady` — the operator cannot reach the cluster. |
| `KubeDBCassandraPhaseCritical` | warning | 5m | The instance is in a degraded/critical phase. |

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
          cassandraDown:
            enabled: true
            duration: "0m"
            val: 0
          cassandraHighReadLatency:
            enabled: true
            duration: "1m"
            val: 15000      # allow up to 15ms 99th-percentile read latency
            severity: warning
      provisioner:
        enabled: "none"     # disable all provisioner alerts
```

```bash
$ helm upgrade cas-alert-demo oci://ghcr.io/appscode-charts/cassandra-alerts \
    -n demo \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

To remove all resources created in this tutorial, run the following commands.

```bash
# Remove the cassandra-alerts release
$ helm uninstall cas-alert-demo -n demo

# Remove the Cassandra instance
$ kubectl delete cassandra -n demo cas-alert-demo

# Delete namespace
$ kubectl delete ns demo
```

## Next Steps

- Monitor your Cassandra cluster with KubeDB using [builtin Prometheus](/docs/guides/cassandra/monitoring/using-builtin-prometheus.md).
- Monitor your Cassandra cluster with KubeDB using [Prometheus operator](/docs/guides/cassandra/monitoring/using-prometheus-operator.md).
- Visualise Cassandra metrics with [Grafana Dashboard](grafana-dashboard.md).
- Learn how to use KubeDB to run a Apache Cassandra cluster [here](/docs/guides/cassandra/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
