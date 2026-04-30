---
title: Monitor Neo4j using Builtin Prometheus Discovery
menu:
  docs_{{ .version }}:
    identifier: neo4j-using-builtin-prometheus-monitoring
    name: Builtin Prometheus
    parent: neo4j-monitoring
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Neo4j with Builtin Prometheus

This tutorial will show you how to monitor Neo4j database using builtin [Prometheus](https://github.com/prometheus/prometheus) scraper.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/neo4j/monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy respective monitoring resources. We are going to deploy the database in `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

## Deploy Neo4j with Monitoring Enabled

At first, let's deploy a Neo4j database with monitoring enabled.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: builtin-prom-neo4j
  namespace: demo
spec:
  version: "2025.11.2"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
  monitor:
    agent: prometheus.io/builtin
```

Here, `spec.monitor.agent: prometheus.io/builtin` specifies that we are going to monitor this server using builtin Prometheus scraper.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/monitoring/builtin-prom-neo4j.yaml
neo4j.kubedb.com/builtin-prom-neo4j created
```

Now, wait for the database to go into `Ready` state.

```bash
$ kubectl get neo4j -n demo builtin-prom-neo4j
NAME                 VERSION   STATUS   AGE
builtin-prom-neo4j   2025.11.2 Ready    2m
```

KubeDB will create a separate stats service with name `{Neo4j crd name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=builtin-prom-neo4j"
NAME                       TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                         AGE
builtin-prom-neo4j         ClusterIP   10.96.100.30   <none>        7474/TCP,7687/TCP,6362/TCP     2m
builtin-prom-neo4j-stats   ClusterIP   10.96.100.31   <none>        2004/TCP                        90s
```

Let's describe the stats service:

```bash
$ kubectl describe svc -n demo builtin-prom-neo4j-stats
Name:              builtin-prom-neo4j-stats
Namespace:         demo
Labels:            app.kubernetes.io/name=neo4js.kubedb.com
                   app.kubernetes.io/instance=builtin-prom-neo4j
Annotations:       monitoring.appscode.com/agent: prometheus.io/builtin
                   prometheus.io/path: /metrics
                   prometheus.io/port: 2004
                   prometheus.io/scrape: true
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo neo4j/builtin-prom-neo4j -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo neo4j/builtin-prom-neo4j

kubectl delete ns demo
kubectl delete ns monitoring
```
