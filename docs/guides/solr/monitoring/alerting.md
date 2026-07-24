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

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed Solr instance using the `solr-alerts` Helm chart.

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

> Note: YAML files used in this tutorial are stored in [docs/examples/solr](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/solr) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

- **KubeDB** deploys Solr with a metrics-exporter sidecar (container `exporter`) that exposes Solr's own metrics (`solr_metrics_*`, `solr_collections_*`).
- **ServiceMonitor** (named `{solr-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape the exporter every 10 seconds.
- **PrometheusRule** is created by the `solr-alerts` chart and contains alert definitions grouped by concern: database health and provisioner.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).
- **Grafana** dashboards for Solr are covered separately — see [Grafana Dashboard](grafana-dashboard.md) rather than duplicated here.

---

## Deploy the ZooKeeper Coordinator

Solr coordinates via a KubeDB `ZooKeeper` cluster, so deploy that first.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: zoo-alert-demo
  namespace: alert-solr
spec:
  version: "3.8.3"
  replicas: 3
  deletionPolicy: WipeOut
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/monitoring/zookeeper-alert-demo.yaml
zookeeper.kubedb.com/zoo-alert-demo created

$ kubectl get zookeeper -n alert-solr zoo-alert-demo
NAME              VERSION   STATUS   AGE
zoo-alert-demo    3.8.3     Ready    3m
```

## Deploy Solr with Monitoring Enabled

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-alert-demo
  namespace: alert-solr
spec:
  version: "9.4.1"
  replicas: 1
  deletionPolicy: WipeOut
  zookeeperRef:
    name: zoo-alert-demo
    namespace: alert-solr
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/monitoring/solr-alert-demo.yaml
solr.kubedb.com/solr-alert-demo created
```

Wait for the database to go into `Ready` state.

```bash
$ kubectl get solr -n alert-solr solr-alert-demo
NAME              VERSION   STATUS   AGE
solr-alert-demo   9.4.1     Ready    3m
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n alert-solr --selector="app.kubernetes.io/instance=solr-alert-demo"
NAME                        TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
solr-alert-demo             ClusterIP   10.43.10.20    <none>        8983/TCP    3m
solr-alert-demo-pods        ClusterIP   None           <none>        8983/TCP    3m
solr-alert-demo-stats       ClusterIP   10.43.10.21    <none>        8080/TCP    3m
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n alert-solr
NAME                    AGE
solr-alert-demo-stats   3m

$ kubectl get servicemonitor -n alert-solr solr-alert-demo-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Install solr-alerts

### Why the Helm release name matters

The chart derives the `PrometheusRule` name and scopes every PromQL expression from the **Helm release name** — so the release name must match the Solr object's name (`solr-alert-demo`).

### Install

```bash
$ helm upgrade -i solr-alert-demo oci://ghcr.io/appscode-charts/solr-alerts \
    -n alert-solr \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus
```

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-solr
NAME                AGE
solr-alert-demo     30s

$ kubectl get prometheusrule -n alert-solr solr-alert-demo \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

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
  <img alt="Prometheus up query — solr-alert-demo-0 UP" src="/docs/images/solr/monitoring/solr-alerting-prom-target.png" style="padding:10px">
</p>

### 2. Confirm the Solr alerts are inactive

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — Solr groups inactive" src="/docs/images/solr/monitoring/solr-alerting-prom-alerts.png" style="padding:10px">
</p>

All rules should show **INACTIVE**.

### 3. Check AlertManager

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager" src="/docs/images/solr/monitoring/solr-alerting-alertmanager.png" style="padding:10px">
</p>

### 4. Grafana dashboard

See [Grafana Dashboard](grafana-dashboard.md) for how to provision and explore the Solr dashboards (via the `kubedb-grafana-dashboards` chart, `--set featureGates.Solr=true`).

---

## Simulating a Firing Alert

This section deliberately triggers `KubeDBSolrPhaseNotReady` by crashing the main Solr JVM process.

### 1. Crash the Solr process

```bash
$ kubectl exec -n alert-solr solr-alert-demo-0 -c solr -- sh -c '
    end=$(( $(date +%s) + 90 ));
    while [ $(date +%s) -lt $end ]; do
      pid=$(pgrep -f "org.apache.solr" | head -1);
      [ -n "$pid" ] && kill -9 "$pid" 2>/dev/null;
      sleep 1;
    done'
```

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — KubeDBSolrPhaseNotReady Firing" src="/docs/images/solr/monitoring/solr-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

`KubeDBSolrPhaseNotReady` (`for: 1m`) should transition to **FIRING** once the KubeDB operator marks the resource `NotReady` and holds it there past the one-minute window. `SolrDownShards`/`SolrRecoveryFailedShards` may also fire if the crash leaves shard replicas in a bad state.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager — KubeDBSolrPhaseNotReady Firing" src="/docs/images/solr/monitoring/solr-alerting-alertmanager-firing.png" style="padding:10px">
</p>

### 4. Restore Solr

Stop the loop from step 1.

```bash
$ kubectl get solr -n alert-solr solr-alert-demo -w
NAME              VERSION   STATUS   AGE
solr-alert-demo   9.4.1     Ready    24m
```

If Solr does not recover on its own within a minute or two, force a clean restart: `kubectl delete pod -n alert-solr solr-alert-demo-0`.

---

## Alert Reference

All alerts are scoped to the `solr-alert-demo` instance in the `alert-solr` namespace via `job="solr-alert-demo-stats"` / `namespace="alert-solr"` (database group), or `app="solr-alert-demo"` / `namespace="alert-solr"` (provisioner group).

### Database Group

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

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBSolrPhaseNotReady` | critical | 1m | KubeDB marked the Solr resource `NotReady`. |
| `KubeDBSolrPhaseCritical` | warning | 1m | Solr is degraded but not fully unavailable. |

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
          solrHighQPS:
            enabled: true
            duration: "2m"
            severity: warning
```

```bash
$ helm upgrade solr-alert-demo oci://ghcr.io/appscode-charts/solr-alerts \
    -n alert-solr \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

```bash
$ helm uninstall solr-alert-demo -n alert-solr
$ kubectl delete solr -n alert-solr solr-alert-demo
$ kubectl delete zookeeper -n alert-solr zoo-alert-demo
$ kubectl delete ns alert-solr
```

## Next Steps

- Monitor your Solr instance with KubeDB using [built-in Prometheus](/docs/guides/solr/monitoring/prometheus-builtin.md).
- Monitor your Solr instance with KubeDB using [Prometheus operator](/docs/guides/solr/monitoring/prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
