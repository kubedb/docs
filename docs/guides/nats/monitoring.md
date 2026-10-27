---
title: Monitoring NATS
menu:
  docs_{{ .version }}:
    identifier: nats-monitoring
    name: Monitoring
    parent: nats-guides
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring NATS

KubeDB adds a `prometheus-nats-exporter` sidecar when `spec.monitor` is set. It reads the local monitor port (`8222`) with `-connz -routez -subz -varz` and, when JetStream is on, `-jsz=all`, and serves `/metrics` on port `7777` of the `<name>-stats` Service.

> Note: The yaml files used in this tutorial are stored in [docs/examples/nats/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/nats/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured. If you do not have one, create it with [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).
- Install KubeDB following the steps [here](/docs/setup/README.md), with `--set global.featureGates.NATS=true` so the NATS CRDs are installed.
- This tutorial uses the `demo` namespace:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Prometheus operator

Install the [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator), then:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: NATS
metadata:
  name: nats-monitored
  namespace: demo
spec:
  version: "2.15.1"
  replicas: 3
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/nats/monitoring/prom-operator-nats.yaml
$ kubectl get servicemonitor -n demo
$ kubectl get svc -n demo nats-monitored-stats
```

The `serviceMonitor.labels` must match your Prometheus `serviceMonitorSelector`.

## Builtin Prometheus

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/nats/monitoring/builtin-prom-nats.yaml
```

The stats Service carries the `prometheus.io/scrape`, `prometheus.io/path` and `prometheus.io/port` annotations for your Prometheus configuration. See [using builtin Prometheus](/docs/guides/rabbitmq/monitoring/using-builtin-prometheus.md) for a scrape configuration that applies to any KubeDB database.

## Alerts worth setting

- `slow_consumers > 0`
- the JetStream meta group has no leader
- a peer reported `outdated` for more than a few minutes
- JetStream storage above 80% of `max_file_store`
- missing routes for a peer (`/routez` grouped by remote name)
- pods in lame-duck mode for longer than the configured duration

Stream replica lag per peer is not an exporter metric. KubeDB reads `/jsz` in its health checker and reports it as the conditions `JetStreamQuorum` and `UnderReplicatedStreams` on the `NATS` status. Phase `Critical` means some pods are not Ready but the meta group still has quorum.

Port `8222` has no authentication and exposes connection and subscription details. KubeDB creates a NetworkPolicy that allows ingress to it only from the operator and Prometheus pods.

## Cleanup

```bash
$ kubectl delete -n demo nats/nats-monitored nats/nats-builtin-prom
```
