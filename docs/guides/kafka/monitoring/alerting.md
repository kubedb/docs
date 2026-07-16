---
title: Kafka Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: kf-monitoring-alerting
    name: Alerting
    parent: kf-monitoring-kafka
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Kafka Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed Kafka cluster using the `kafka-alerts` Helm chart.

## Before You Begin

* Ensure you have a Kubernetes cluster and that `kubectl` is configured to communicate with it. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Install the KubeDB operator by following the steps [here](/docs/setup/README.md).

* Deploy the database in the `alert-kafka` namespace:

  ```bash
  $ kubectl create ns alert-kafka
  namespace/alert-kafka created
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

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/kafka/monitoring/overview.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/kafka](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/kafka) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

- **KubeDB** deploys Kafka with metrics exposed by a [JMX Exporter](https://github.com/prometheus/jmx_exporter) running as a **Java agent inside the `kafka` container itself** — not a separate sidecar container. KubeDB uses the JMX agent because the officially recognized Kafka exporter image does not yet expose metrics for the KRaft-mode versions KubeDB supports.
- **ServiceMonitor** (named `{kafka-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape the JMX agent's HTTP endpoint every 10 seconds.
- **PrometheusRule** is created by the `kafka-alerts` chart and contains alert definitions grouped by concern: database health (which also embeds KubeDB-operator-sourced `KafkaDown`/`KafkaPhaseCritical` alerts) and provisioner.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).
- **Grafana** dashboards for Kafka are covered separately — see [Grafana Dashboard](grafana-dashboard.md) rather than duplicated here.

---

## Deploy Kafka with Monitoring Enabled

Below is a single-broker Kafka object for this tutorial (a production cluster would use `spec.topology` for separate broker/controller roles).

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-alert-demo
  namespace: alert-kafka
spec:
  replicas: 1
  version: "3.9.0"
  storageType: Durable
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/monitoring/kafka-alert-demo.yaml
kafka.kubedb.com/kafka-alert-demo created
```

Wait for the cluster to go into `Ready` state.

```bash
$ kubectl get kafka -n alert-kafka kafka-alert-demo
NAME                VERSION   STATUS   AGE
kafka-alert-demo    3.9.0     Ready    3m
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n alert-kafka --selector="app.kubernetes.io/instance=kafka-alert-demo"
NAME                        TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
kafka-alert-demo            ClusterIP   10.43.10.20    <none>        9092/TCP    3m
kafka-alert-demo-pods       ClusterIP   None           <none>        9092/TCP    3m
kafka-alert-demo-stats      ClusterIP   10.43.10.21    <none>        56790/TCP   3m
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n alert-kafka
NAME                     AGE
kafka-alert-demo-stats   3m

$ kubectl get servicemonitor -n alert-kafka kafka-alert-demo-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Install kafka-alerts

### Why the Helm release name matters

The chart derives the `PrometheusRule` name and scopes every PromQL expression from the **Helm release name** — so the release name must match the Kafka object's name (`kafka-alert-demo`).

### Install

```bash
$ helm upgrade -i kafka-alert-demo oci://ghcr.io/appscode-charts/kafka-alerts \
    -n alert-kafka \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus
```

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-kafka
NAME                 AGE
kafka-alert-demo     30s

$ kubectl get prometheusrule -n alert-kafka kafka-alert-demo \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Confirm Prometheus loaded the rules

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules` and locate the `kafka.database` and `kafka.provisioner` groups.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/kafka/monitoring/kafka-alerting-prom-rules.png" style="padding:10px">
</p>

Both groups should show **OK**. `kafka-alerts` v2026.7.14 has no `opsManager`/`stash`/`kubeStash` groups — only `database` and `provisioner`.

> **Note the overlap:** the `database` group's `KafkaDown` (`for: 30s`) and `KafkaPhaseCritical` (`for: 3m`) key off the same `kubedb_com_kafka_status_phase` metric as the `provisioner` group's `KubeDBKafkaPhaseNotReady`/`KubeDBKafkaPhaseCritical` (`for: 1m`/`15m`) — expect both pairs to eventually fire together during a real outage, at different times.

---

## Verify End-to-End

### 1. Check the Prometheus target is UP

Open `http://localhost:9090/query?g0.expr=up%7Bnamespace%3D%22alert-kafka%22%7D&g0.tab=1`.

<p align="center">
  <img alt="Prometheus up query — kafka-alert-demo-0 UP" src="/docs/images/kafka/monitoring/kafka-alerting-prom-target.png" style="padding:10px">
</p>

### 2. Confirm the Kafka alerts are inactive

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — Kafka groups inactive" src="/docs/images/kafka/monitoring/kafka-alerting-prom-alerts.png" style="padding:10px">
</p>

All rules should show **INACTIVE**. `KafkaTopicCount` and the replication-related alerts (`KafkaUnderReplicatedPartitions`, `KafkaUnderMinIsrPartitionCount`, `KafkaISRExpandRate`/`KafkaISRShrinkRate`) are naturally quiet on a single-broker cluster with no topics yet.

### 3. Check AlertManager

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager" src="/docs/images/kafka/monitoring/kafka-alerting-alertmanager.png" style="padding:10px">
</p>

### 4. Grafana dashboard

See [Grafana Dashboard](grafana-dashboard.md) for how to provision and explore the Kafka dashboards (via the `kubedb-grafana-dashboards` chart, `--set featureGates.Kafka=true`).

---

## Simulating a Firing Alert

This section deliberately triggers `KafkaDown` (`for: 30s`, the fastest down-signal) by crashing the main Kafka JVM process.

### 1. Crash the Kafka process

```bash
$ kubectl exec -n alert-kafka kafka-alert-demo-0 -c kafka -- sh -c '
    end=$(( $(date +%s) + 60 ));
    while [ $(date +%s) -lt $end ]; do
      pid=$(pgrep -f "kafka.Kafka" | head -1);
      [ -n "$pid" ] && kill -9 "$pid" 2>/dev/null;
      sleep 1;
    done'
```

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — KafkaDown Firing" src="/docs/images/kafka/monitoring/kafka-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

`KafkaDown` (`kubedb_com_kafka_status_phase{phase!="Ready"} == 1`, `for: 30s`) should transition to **FIRING** first.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager — KafkaDown Firing" src="/docs/images/kafka/monitoring/kafka-alerting-alertmanager-firing.png" style="padding:10px">
</p>

### 4. Restore Kafka

Stop the loop from step 1.

```bash
$ kubectl get kafka -n alert-kafka kafka-alert-demo -w
NAME               VERSION   STATUS   AGE
kafka-alert-demo   3.9.0     Ready    24m
```

If Kafka does not recover on its own within a minute or two, force a clean restart: `kubectl delete pod -n alert-kafka kafka-alert-demo-0`.

---

## Alert Reference

All alerts are scoped to the `kafka-alert-demo` instance in the `alert-kafka` namespace via `job="kafka-alert-demo-stats"` / `namespace="alert-kafka"` (database group), or `app="kafka-alert-demo"` / `namespace="alert-kafka"` (provisioner group and the two operator-phase alerts embedded in the database group).

### Database Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KafkaUnderReplicatedPartitions` | warning | 10s | Partitions have fewer in-sync replicas than expected. |
| `KafkaAbnormalControllerState` | warning | 10s | More or fewer than one active controller in the cluster. |
| `KafkaOfflinePartitions` | warning | 10s | One or more partitions have no leader. |
| `KafkaUnderMinIsrPartitionCount` | warning | 10s | Partitions below the minimum in-sync-replica count. |
| `KafkaOfflineLogDirectoryCount` | warning | 10s | A log directory has gone offline (likely a disk issue). |
| `KafkaISRExpandRate` | warning | 1m | ISR set is expanding frequently — sign of flakiness. |
| `KafkaISRShrinkRate` | warning | 1m | ISR set is shrinking frequently — sign of flakiness. |
| `KafkaBrokerCount` | critical | 1m | Broker count has dropped. |
| `KafkaNetworkProcessorIdlePercent` | critical | 1m | Network processor threads are saturated. |
| `KafkaRequestHandlerIdlePercent` | critical | 1m | Request handler threads are saturated. |
| `KafkaReplicaFetcherManagerMaxLag` | critical | 1m | Replica fetcher lag is high. |
| `KafkaTopicCount` | warning | 1m | Topic count changed unexpectedly. |
| `KafkaPhaseCritical` | warning | 3m | KubeDB operator view: resource `Critical` (duplicates the provisioner group's own version at a different `for`). |
| `KafkaDown` | critical | 30s | KubeDB operator view: resource not `Ready`. Fastest down-signal available. |
| `DiskUsageHigh` | warning | 1m | Persistent volume usage exceeds 80%. |
| `DiskAlmostFull` | critical | 1m | Persistent volume usage exceeds 95%. |

### Provisioner Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBKafkaPhaseNotReady` | critical | 1m | KubeDB marked the Kafka resource `NotReady`. |
| `KubeDBKafkaPhaseCritical` | warning | 15m | Kafka is degraded but not fully unavailable. |

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
          kafkaBrokerCount:
            enabled: true
            duration: "2m"
            severity: critical
```

```bash
$ helm upgrade kafka-alert-demo oci://ghcr.io/appscode-charts/kafka-alerts \
    -n alert-kafka \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

```bash
$ helm uninstall kafka-alert-demo -n alert-kafka
$ kubectl delete kafka -n alert-kafka kafka-alert-demo
$ kubectl delete ns alert-kafka
```

## Next Steps

- Monitor your Kafka cluster with KubeDB using [built-in Prometheus](/docs/guides/kafka/monitoring/using-builtin-prometheus.md).
- Monitor your Kafka cluster with KubeDB using [Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
