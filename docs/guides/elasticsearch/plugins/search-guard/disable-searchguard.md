---
title: Disable Search Guard
menu:
  docs_{{ .version }}:
    identifier: es-disable-search-guard
    name: Disable Search Guard
    parent: es-search-guard-elasticsearch
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Disable Search Guard Plugin

Databases are precious. Definitely, you will not want to left your production database unprotected. Hence, KubeDB ship with Search Guard plugin integrated with it. It provides you authentication, authorization and TLS security. However, you can disable Search Guard plugin. You have to set `spec.authPlugin` field of Elasticsearch object to `None`.

This tutorial will show you how to disable Search Guard plugin for Elasticsearch database in KubeDB.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created

$ kubectl get ns demo
NAME    STATUS  AGE
demo    Active  5s
```

> Note: YAML files used in this tutorial are stored in [docs/examples/elasticsearch](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/elasticsearch) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create Elasticsearch

In order to disable Search Guard, you have to set `spec.authPlugin` field of Elasticsearch object to `None`. Below is the YAML of Elasticsearch object that will be created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-sg-disabled
  namespace: demo
spec:
  version: searchguard-7.9.3
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Let's create the Elasticsearch object we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/search-guard/es-sg-disabled.yaml
elasticsearch.kubedb.com/es-sg-disabled created
```

Wait for Elasticsearch to be ready,

```bash
$ kubectl get es -n demo es-sg-disabled
NAME             VERSION   STATUS    AGE
es-sg-disabled   6.3-v1    Running   27m
```

## Connect to Elasticsearch Database

As we have disabled Search Guard plugin, we no longer require *username* and *password* to connect with our Elasticsearch database.

At first, forward port 9200 of `es-sg-disabled-0` pod. Run following command in a separate terminal,

```bash
$ kubectl port-forward -n demo es-sg-disabled-0 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```

Now, we can connect with the database at `localhost:9200`.

Let's check health of our Elasticsearch database.

```bash
$ curl "localhost:9200/_cluster/health?pretty"
```

```json
{
  "cluster_name" : "es-sg-disabled",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 1,
  "number_of_data_nodes" : 1,
  "active_primary_shards" : 0,
  "active_shards" : 0,
  "relocating_shards" : 0,
  "initializing_shards" : 0,
  "unassigned_shards" : 0,
  "delayed_unassigned_shards" : 0,
  "number_of_pending_tasks" : 0,
  "number_of_in_flight_fetch" : 0,
  "task_max_waiting_in_queue_millis" : 0,
  "active_shards_percent_as_number" : 100.0
}
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo es/es-sg-disabled -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete -n demo es/es-sg-disabled

$ kubectl delete ns demo
```

## Next Steps

- Learn how to [create TLS certificates](/docs/guides/elasticsearch/plugins/search-guard/issue-certificate.md).
- Learn how to generate [search-guard configuration](/docs/guides/elasticsearch/plugins/search-guard/configuration.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
