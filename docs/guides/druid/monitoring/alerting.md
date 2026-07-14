---
title: Druid Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: guides-druid-monitoring-alerting
    name: Alerting
    parent: guides-druid-monitoring
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Druid Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed Apache Druid instance using the `druid-alerts` Helm chart.

## Before You Begin

* Ensure you have a Kubernetes cluster and that `kubectl` is configured to communicate with it. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Install the KubeDB operator by following the steps [here](/docs/setup/README.md), making sure to enable the `Druid` and `ZooKeeper` feature gates.

* Deploy the database in the `demo` namespace:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

* This tutorial assumes you already have a **kube-prometheus-stack** running in your cluster, with `Prometheus` configured so that both `serviceMonitorSelector` and `ruleSelector` match the label `release: prometheus`. See the [Grafana Dashboard](grafana-dashboard.md) guide for how to deploy kube-prometheus-stack if you don't have it yet.

  To verify the selectors:

  ```bash
  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.ruleSelector}'
  {"matchLabels":{"release":"prometheus"}}

  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.serviceMonitorSelector}'
  {"matchLabels":{"release":"prometheus"}}
  ```

* Druid requires an external **deep storage** backend (for storing segments) and a **metadata storage** database before it can become `Ready`. This tutorial uses a MinIO tenant as S3-compatible deep storage, and lets KubeDB auto-provision a MySQL cluster for metadata storage. See the [Druid Quickstart](/docs/guides/druid/quickstart/guide/index.md#get-external-dependencies-ready) guide for the full walkthrough of setting these up — the short version is repeated in the [Deploy](#deploy-druid-with-monitoring-enabled) section below.

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/druid/monitoring/overview.md).

* For dashboards and visualisation, see [Grafana Dashboard](grafana-dashboard.md) for Druid.

> Note: YAML files used in this tutorial are stored in [docs/guides/druid/monitoring/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/druid/monitoring/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

- **KubeDB** deploys Druid nodes (router, broker, coordinator, historical, middleManager, overlord) with a **JMX Exporter** Java agent running inside each Druid container, exposing Prometheus metrics over HTTP — unlike most other KubeDB databases which use a separate sidecar exporter container.
- **ServiceMonitor** (named `{druid-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape every Druid node's exporter every 10 seconds.
- **PrometheusRule** is created by the `druid-alerts` chart and contains all Druid alert definitions grouped by concern: database health and provisioner.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).

<figure align="center">
  <img alt="Monitoring process of Druid using Prometheus Operator" src="/docs/guides/druid/monitoring/images/druid-monitoring.png">
</figure>

---

## Deploy Druid with Monitoring Enabled

### Prepare Deep Storage and Metadata Storage

Druid cannot start without a deep storage backend for segments. We install a MinIO tenant to provide S3-compatible storage in the same namespace as the database (the cluster-wide `minio-operator` is assumed to already be installed — see [Druid Quickstart](/docs/guides/druid/quickstart/guide/index.md#get-external-dependencies-ready) if it isn't):

```bash
$ helm repo add minio https://operator.min.io/
$ helm repo update minio

$ helm upgrade --install --namespace demo druid-minio minio/tenant \
    --set tenant.pools[0].servers=1 \
    --set tenant.pools[0].volumesPerServer=1 \
    --set tenant.pools[0].size=1Gi \
    --set tenant.certificate.requestAutoCert=false \
    --set tenant.buckets[0].name="druid" \
    --set tenant.pools[0].name="default"
```

Once the tenant pod is `Running`, note the headless service it creates (typically `myminio-hl`) and create the `deep-storage-config` Secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: deep-storage-config
  namespace: demo
stringData:
  druid.storage.type: "s3"
  druid.storage.bucket: "druid"
  druid.storage.baseKey: "druid/segments"
  druid.s3.accessKey: "minio"
  druid.s3.secretKey: "minio123"
  druid.s3.protocol: "http"
  druid.s3.enablePathStyleAccess: "true"
  druid.s3.endpoint.signingRegion: "us-east-1"
  druid.s3.endpoint.url: "http://myminio-hl.demo.svc.cluster.local:9000/"
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/monitoring/yamls/deep-storage-config.yaml
secret/deep-storage-config created
```

KubeDB doesn't require you to set up metadata storage yourself — if `spec.metadataStorage` is left unset, the operator automatically provisions a dedicated MySQL cluster (3 replicas, for group-replication quorum) and a ZooKeeper ensemble for cluster coordination. This means the Druid object stays in the `Provisioning` phase for several minutes on first deploy while these dependencies come up — this is expected, not a stuck deployment.

### Deploy

Below is the Druid object we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-alert-demo
  namespace: demo
spec:
  version: 28.0.1
  deepStorage:
    type: s3
    configSecret:
      name: deep-storage-config
  topology:
    routers:
      replicas: 1
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  deletionPolicy: WipeOut
```

Here,

- `spec.monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` resource managed by the Prometheus operator.
- `spec.monitor.prometheus.serviceMonitor.labels.release: prometheus` adds the `release: prometheus` label to the created `ServiceMonitor`, matching the Prometheus `serviceMonitorSelector` so the target is discovered automatically.

Let's create the Druid resource.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/monitoring/yamls/druid-alert-demo.yaml
druid.kubedb.com/druid-alert-demo created
```

Now, wait for the database to go into `Ready` state. Because of the MySQL and ZooKeeper dependencies mentioned above, this can take **5 minutes or more** on a freshly created namespace.

```bash
$ kubectl get druid -n demo druid-alert-demo -w
NAME               VERSION   STATUS         AGE
druid-alert-demo   28.0.1    Provisioning   30s
druid-alert-demo   28.0.1    Provisioning   4m50s
druid-alert-demo   28.0.1    Ready          5m20s
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=druid-alert-demo"
NAME                       TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
druid-alert-demo-routers   ClusterIP   10.43.100.12    <none>        8888/TCP            5m
druid-alert-demo-stats     ClusterIP   10.43.211.90    <none>        8888/TCP,9255/TCP   5m
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n demo
NAME                     AGE
druid-alert-demo-stats   5m
```

Verify that the `ServiceMonitor` carries the `release: prometheus` label so Prometheus discovers it.

```bash
$ kubectl get servicemonitor -n demo druid-alert-demo-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Install druid-alerts

The `druid-alerts` chart creates a `PrometheusRule` resource containing all Druid alert definitions grouped by concern: database health and provisioner.

### Why the Helm release name matters

The chart derives the PromQL `job`/instance scoping (and the `PrometheusRule` name) from the **Helm release name**, not from a values field — so the release name must match the Druid object's name (`druid-alert-demo`) for the rules to be correctly scoped to this instance.

The chart's default label is `release: kube-prometheus-stack`, so we must also override it at install time to match the Prometheus `ruleSelector`.

### Install

```bash
$ helm upgrade -i druid-alert-demo oci://ghcr.io/appscode-charts/druid-alerts \
    -n demo \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus
```

| Flag | Value | Purpose |
|------|-------|---------|
| `druid-alert-demo` (release name) | — | Scopes every PromQL expression to this instance (`service="druid-alert-demo-stats"`) |
| `-n demo` | `demo` | Installs the `PrometheusRule` in the same namespace as the database |
| `form.alert.labels.release` | `prometheus` | Matches the Prometheus `ruleSelector` so the rules are loaded |

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n demo
NAME               AGE
druid-alert-demo   30s
```

Confirm the `release: prometheus` label is present.

```bash
$ kubectl get prometheusrule -n demo druid-alert-demo \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Confirm Prometheus loaded the rules

Port-forward the Prometheus UI and open the **Status → Rule health** page.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules?search=druid`.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/guides/druid/monitoring/images/druid-alerting-prom-rules.png" style="padding:10px">
</p>

The `druid.database.demo.druid-alert-demo.rules` and `druid.provisioner.demo.druid-alert-demo.rules` groups are visible with all rules showing **OK**, confirming that Prometheus has loaded and is evaluating the Druid alert definitions every 30 seconds.

---

## Verify End-to-End

### 1. Check the exporter is running

Every Druid node runs the JMX exporter Java agent, serving metrics on the stats service's `9255` port. A value of `druid_service_heartbeat 1` confirms the node is up and being scraped.

```bash
$ kubectl exec -n demo druid-alert-demo-routers-0 -c druid -- \
    curl -s localhost:9255/metrics | grep druid_service_heartbeat
druid_service_heartbeat{service="druid-alert-demo-stats",...,} 1.0
```

### 2. Check the Prometheus target is UP

Open `http://localhost:9090/targets?search=druid-alert-demo`.

<p align="center">
  <img alt="Prometheus Target UP" src="/docs/guides/druid/monitoring/images/druid-alerting-prom-target.png" style="padding:10px">
</p>

The target(s) for `serviceMonitor/demo/druid-alert-demo-stats` show **UP**, confirming metrics are being scraped from every Druid node in the `demo` namespace.

### 3. Confirm all Druid alerts are inactive

Open `http://localhost:9090/alerts?search=druid` to see the Druid alert groups.

<p align="center">
  <img alt="Prometheus Alerts — All Inactive" src="/docs/guides/druid/monitoring/images/druid-alerting-prom-alerts.png" style="padding:10px">
</p>

All rules in the `druid.database` and `druid.provisioner` groups show **INACTIVE**, meaning the cluster is healthy and no thresholds are breached.

### 4. Check AlertManager

Port-forward AlertManager to view any currently firing alerts.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`. With a healthy Druid instance, no alerts for `druid-alert-demo` will be listed here.

---

## Simulating a Firing Alert

The previous section confirmed that all alerts are **INACTIVE** while the database is healthy. This section walks through deliberately triggering the `DruidDown` critical alert so you can observe the full alert lifecycle — from firing in Prometheus through to the AlertManager dashboard — and then resolve it.

### 1. Stop a Druid node process

Druid runs several node types as separate pods (router, broker, coordinator, historical, middleManager, overlord). Pick one — here we use the router — and kill the main Druid JVM process (pid 1) inside its main container. This crashes the main container so the JMX exporter agent running inside it goes down with it, and the pod reports `druid_service_heartbeat 0` on the next scrape once the container restarts into a crash state, while Kubernetes handles the restart in the background.

```bash
$ kubectl exec -n demo druid-alert-demo-routers-0 -c druid -- kill 1
```

Wait 30–90 seconds for the next Prometheus scrape cycle (configured at 10s) and rule-evaluation cycle (30s) to register the failure — `DruidDown` has `for: 1m`, so it needs one full minute of continuous failure before it transitions to firing.

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts?search=druid`.

<p align="center">
  <img alt="Prometheus Alerts — DruidDown Firing" src="/docs/guides/druid/monitoring/images/druid-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

Because `DruidDown` has `for: 1m`, it moves from **INACTIVE** to **PENDING** and then to **FIRING** once the condition holds continuously for a full minute.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager — DruidDown Firing" src="/docs/guides/druid/monitoring/images/druid-alerting-alertmanager-firing.png" style="padding:10px">
</p>

AlertManager shows the `DruidDown` alert. The alert card displays:

- **Severity**: `critical`
- **Instance/pod**: `druid-alert-demo-routers-0` in the `demo` namespace
- **service**: `druid-alert-demo-stats`
- **Started**: timestamp when the alert first fired

AlertManager routes this alert to every receiver configured in your `alertmanagerConfig` (Slack, email, PagerDuty, webhook, etc.) based on your routing tree. If no receiver is configured, the alert is visible here but silently dropped.

### 4. Restore Druid

Delete the pod so KubeDB recreates it cleanly.

```bash
$ kubectl delete pod -n demo druid-alert-demo-routers-0
```

Once `druid_service_heartbeat` returns to `1`, Prometheus marks the alert **INACTIVE** again and AlertManager sends a **resolved** notification to all receivers.

---

## Alert Reference

All alerts are scoped to the `druid-alert-demo` instance in the `demo` namespace via the PromQL label filters `service="druid-alert-demo-stats"` and `namespace="demo"`.

### Database Group

Fired based on live JMX-exporter metrics from the Druid nodes.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `DruidDown` | critical | 1m | One of the Druid services is down for more than the configured duration. |
| `ZKDisconnected` | critical | 1m | Druid lost connection to ZooKeeper. |
| `HighQueryTime` | warning | 1m | A query took more than 1 second to complete on a historical node. |
| `HighQueryWaitTime` | warning | 1m | Druid spent more than 1 second waiting for a segment to be scanned. |
| `HighSegmentScanPending` | warning | 1m | More than 2 segments are queued waiting to be scanned. |
| `HighSegmentUsage` | critical | 1m | More than 95% of space is used by served segments. |
| `HighJVMPoolUsage` | warning | 30s | More than 95% of a JVM memory pool is being used. |
| `HighJVMMemoryUsage` | critical | 30s | More than 95% of JVM memory is being used. |

### Provisioner Group

Monitors the KubeDB operator's view of the Druid resource phase.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBDruidPhaseNotReady` | critical | 1m | KubeDB marked the Druid resource `NotReady` — operator cannot reach the cluster. |
| `KubeDBDruidPhaseCritical` | warning | 15m | The instance is in a degraded/critical phase. |

### OpsManager Group

Tracks `DruidOpsRequest` lifecycle during upgrades, scaling, and reconfiguration. These rules are defined in the chart's `values.yaml` under `form.alert.groups.opsManager`, following the same convention used by the ops-manager group in other `*-alerts` charts (e.g. `memcached-alerts`).

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `opsRequestOnProgress` | info | instant | An ops request is currently in progress. |
| `opsRequestStatusProgressingToLong` | critical | 30m | An ops request has been running for 30+ minutes — likely stuck. |
| `opsRequestFailed` | critical | instant | An ops request failed — check the `OpsRequest` object for the error. |

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
          highSegmentUsage:
            enabled: true
            duration: "5m"
            val: 90        # fire at 90% segment usage instead of the default 95%
            severity: warning
      provisioner:
        enabled: "none"    # disable all provisioner alerts
```

```bash
$ helm upgrade druid-alert-demo oci://ghcr.io/appscode-charts/druid-alerts \
    -n demo \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

To remove all resources created in this tutorial, run the following commands.

```bash
# Remove the druid-alerts release
$ helm uninstall druid-alert-demo -n demo

# Remove the Druid instance
$ kubectl delete druid -n demo druid-alert-demo

# Remove the MinIO tenant used for deep storage
$ helm uninstall druid-minio -n demo

# Delete namespace
$ kubectl delete ns demo
```

## Next Steps

- Monitor your Druid database with KubeDB using [builtin Prometheus](/docs/guides/druid/monitoring/using-builtin-prometheus.md).
- Monitor your Druid database with KubeDB using [Prometheus operator](/docs/guides/druid/monitoring/using-prometheus-operator.md).
- Visualise Druid metrics with [Grafana Dashboard](grafana-dashboard.md).
- Detail concepts of [DruidVersion object](/docs/guides/druid/concepts/druidversion.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
