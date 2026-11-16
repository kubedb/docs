---
title: Monitoring Overview
menu:
  docs_{{ .version }}:
    identifier: dw-monitoring-overview
    name: Monitoring
    parent: dw-monitoring-durableworkflow
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring DurableWorkflow with KubeDB

Every role of Temporal and Cadence exposes Prometheus metrics itself, on port `9090` (the `metrics` port of the catalog), so KubeDB does not add an exporter sidecar. When `spec.monitor` is set, KubeDB creates the Service `<name>-stats`, which selects the pods of the cluster, and either annotates it for the builtin Prometheus scraper or creates a `ServiceMonitor`. The engines tag each series with the role in a `type` label.

For the fields of `spec.monitor` and the general flow, read [Monitoring Elasticsearch with KubeDB](/docs/guides/elasticsearch/monitoring/overview.md): they are the same for every KubeDB database, except that there is no `exporter` to configure here.

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Builtin Prometheus

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: temporal-builtin
  namespace: demo
spec:
  version: temporal-1.32.1
  topology:
    frontend:
      replicas: 2
    history:
      replicas: 2
    matching:
      replicas: 2
    worker:
      replicas: 2
  persistence:
    default:
      name: pg
      namespace: demo
  monitor:
    agent: prometheus.io/builtin
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/monitoring/builtin.yaml
$ kubectl get svc -n demo temporal-builtin-stats
```

Configure the Prometheus scraper as described in [Using builtin Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).

## Prometheus operator

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: temporal-prom
  namespace: demo
spec:
  version: temporal-1.32.1
  topology:
    frontend:
      replicas: 2
    history:
      replicas: 2
    matching:
      replicas: 2
    worker:
      replicas: 2
  persistence:
    default:
      name: pg
      namespace: demo
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  deletionPolicy: WipeOut
```

The `release: prometheus` label must match the `serviceMonitorSelector` of your Prometheus.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/monitoring/prometheus-operator.yaml
$ kubectl get servicemonitor -n demo
```

## What to watch

The metrics are the ones of the engine, so the upstream Grafana dashboards of [Temporal](https://docs.temporal.io/references/cluster-metrics) and [Cadence](https://cadenceworkflow.io/docs/operation-guide/monitoring/) apply. The useful signals are the availability of the frontend, the latency and errors of the persistence calls, the movement of history shard ownership, and the lag of visibility writes. KubeDB does not ship its own dashboards or alert rules for DurableWorkflow yet.

## Cleaning up

```bash
kubectl delete dw -n demo temporal-builtin temporal-prom
```

## Next Steps

- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
