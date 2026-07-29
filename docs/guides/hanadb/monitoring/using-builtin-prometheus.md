---
title: HanaDB Builtin Prometheus
menu:
  docs_{{ .version }}:
    identifier: guides-hanadb-monitoring-builtin
    name: Builtin Prometheus
    parent: guides-hanadb-monitoring
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring HanaDB with Builtin Prometheus

This guide deploys a HanaDB with the builtin Prometheus agent so a Prometheus server that scrapes by pod
annotation can collect its metrics.

> Note: The YAML files used in this tutorial are stored in [docs/examples/hanadb/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/monitoring) folder in the GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- Install the KubeDB Provisioner and Ops-manager operators following the steps [here](/docs/setup/README.md).
- Create a namespace:

```bash
$ kubectl create ns demo
namespace/demo created
```

## Deploy a HanaDB with Builtin Monitoring

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hanadb-builtin-prometheus
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: local-path
    accessModes: ["ReadWriteOnce"]
    resources:
      requests:
        storage: 64Gi
  deletionPolicy: WipeOut
  monitor:
    agent: prometheus.io/builtin
    prometheus:
      exporter:
        port: 9668
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/monitoring/builtin-prometheus.yaml
hanadb.kubedb.com/hanadb-builtin-prometheus created
```

Wait until the database is `Ready`.

## Verify the Metrics Endpoint

KubeDB adds an `exporter` container and a `<db>-stats` Service:

```bash
$ kubectl get pod -n demo hanadb-builtin-prometheus-0 -o jsonpath='{range .spec.containers[*]}{.name}{"\n"}{end}'
hanadb
exporter

$ kubectl get svc -n demo -l app.kubernetes.io/instance=hanadb-builtin-prometheus
NAME                              TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)               AGE
hanadb-builtin-prometheus         ClusterIP   10.43.27.56     <none>        39017/TCP             17m
hanadb-builtin-prometheus-pods    ClusterIP   None            <none>        39001/TCP,39017/TCP   17m
hanadb-builtin-prometheus-stats   ClusterIP   10.43.169.153   <none>        9668/TCP              17m
```

The stats Service carries the `prometheus.io/scrape`, `prometheus.io/port`, and `prometheus.io/path`
annotations a builtin Prometheus uses to discover the target:

```bash
$ kubectl get svc -n demo hanadb-builtin-prometheus-stats -o jsonpath='{.metadata.annotations}' | jq
{
  "monitoring.appscode.com/agent": "prometheus.io/builtin",
  "prometheus.io/path": "/metrics",
  "prometheus.io/port": "9668",
  "prometheus.io/scheme": "http",
  "prometheus.io/scrape": "true"
}
```

Scrape the metrics to confirm the exporter is serving (the `exporter` container is distroless, so curl
the stats Service from a throwaway pod):

```bash
$ kubectl run hdb-metrics-check -n demo --rm -i --restart=Never --image=curlimages/curl:8.10.1 -- \
  curl -s http://hanadb-builtin-prometheus-stats.demo.svc:9668/metrics | grep -E '^hanadb_' | head
hanadb_column_tables_used_memory_mb{database_name="SYSTEMDB",host="hanadb-builtin-prometheus-0",insnr="90",sid="HXE"} 6.0
hanadb_schema_used_memory_mb{database_name="SYSTEMDB",host="hanadb-builtin-prometheus-0",insnr="90",schema_name="_SYS_REPO",sid="HXE"} 1.0
hanadb_schema_used_memory_mb{database_name="SYSTEMDB",host="hanadb-builtin-prometheus-0",insnr="90",schema_name="_SYS_DI",sid="HXE"} 1.0
```

## Cleaning Up

```bash
$ kubectl delete hanadb.kubedb.com -n demo hanadb-builtin-prometheus
$ kubectl delete ns demo
```

## Next Steps

- [Monitor with the Prometheus Operator](/docs/guides/hanadb/monitoring/using-prometheus-operator.md).
