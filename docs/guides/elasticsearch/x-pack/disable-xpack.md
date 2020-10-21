---
title: Disable X-Pack
menu:
  docs_{{ .version }}:
    identifier: es-disable-x-pack
    name: Disable X-Pack
    parent: es-x-pack
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Disable X-Pack Plugin

You data is precious. Definitely, you will not want to leave your production database unprotected. Hence, KubeDB automates Elasticsearch X-Pack configuration. It provides you authentication, authorization and TLS security. However, you can disable X-Pack security. You have to set `spec.disableSecurity` field of Elasticsearch object to `true`.

This tutorial will show you how to disable X-Pack security for Elasticsearch database in KubeDB.

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

## X-Pack enabled ElasticsearchVersion

To deploy with X-Pack, you need to use an `ElasticsearchVersion` where `X-Pack` is used as `authPlugin`.

Here, we are going to use ElasticsearchVersion `7.3.2`.

> To change authPlugin, it is recommended to create another `ElasticsearchVersion` CRD. Then, use that `ElasticsearchVersion` to install an Elasticsearch without authentication, or with other authPlugin.

```bash
$ kubectl get elasticsearchversions 7.3.2 -o yaml
```

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: ElasticsearchVersion
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"catalog.kubedb.com/v1alpha1","kind":"ElasticsearchVersion","metadata":{"annotations":{},"labels":{"app":"kubedb"},"name":"7.3.2"},"spec":{"authPlugin":"X-Pack","db":{"image":"kubedb/elasticsearch:7.3.2"},"exporter":{"image":"kubedb/elasticsearch_exporter:1.0.2"},"initContainer":{"image":"kubedb/busybox","yqImage":"kubedb/yq:2.4.0"},"podSecurityPolicies":{"databasePolicyName":"elasticsearch-db","snapshotterPolicyName":"elasticsearch-snapshot"},"tools":{"image":"kubedb/elasticsearch-tools:7.3.2"},"version":"7.3.2"}}
  creationTimestamp: "2019-09-26T05:46:47Z"
  generation: 1
  labels:
    app: kubedb
  name: 7.3.2
  resourceVersion: "2781140"
  selfLink: /apis/catalog.kubedb.com/v1alpha1/elasticsearchversions/7.3.2
  uid: 07309b1a-e021-11e9-acff-42010a8001f4
spec:
  authPlugin: X-Pack
  db:
    image: kubedb/elasticsearch:7.3.2
  exporter:
    image: kubedb/elasticsearch_exporter:1.0.2
  initContainer:
    image: kubedb/busybox
    yqImage: kubedb/yq:2.4.0
  podSecurityPolicies:
    databasePolicyName: elasticsearch-db
    snapshotterPolicyName: elasticsearch-snapshot
  tools:
    image: kubedb/elasticsearch-tools:7.3.2
  version: 7.3.2
```

## Create Elasticsearch

In order to disable X-Pack, you have to set `spec.disableSecurity` field of `Elasticsearch` object to `true`.

Below is the YAML of `Elasticsearch` object that will be created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-xpack-disabled
  namespace: demo
spec:
  version: "7.3.2"
  disableSecurity: true
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Let's create the Elasticsearch object,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/x-pack/es-xpack-disabled.yaml
elasticsearch.kubedb.com/es-xpack-disabled created
```

Wait for Elasticsearch to be ready,

```bash
$ kubectl get es -n demo es-xpack-disabled
NAME                VERSION   STATUS    AGE
es-xpack-disabled   7.3.2     Running   6m14s
```

## Connect to Elasticsearch Database

As we have disabled X-Pack security, we no longer require *username* and *password* to connect with our Elasticsearch database.

At first, forward port 9200 of `es-xpack-disabled-0` pod. Run following command in a separate terminal,

```bash
$ kubectl port-forward -n demo es-xpack-disabled-0 9200
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
  "cluster_name" : "es-xpack-disabled",
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

Additionally, to query the settings about xpack,

```json
$ curl "localhost:9200/_nodes/_all/settings?pretty"
{
  "_nodes" : {
    "total" : 1,
    "successful" : 1,
    "failed" : 0
  },
  "cluster_name" : "es-xpack-disabled",
  "nodes" : {
    "GpHq4kaERoq8_43zXup_mA" : {
      "name" : "es-xpack-disabled-0",
      "transport_address" : "10.244.1.7:9300",
      "host" : "10.244.1.7",
      "ip" : "10.244.1.7",
      "version" : "7.3.2",
      "build_flavor" : "default",
      "build_type" : "docker",
      "build_hash" : "1c1faf1",
      "roles" : [
        "ingest",
        "master",
        "data"
      ],
      "attributes" : {
        "ml.machine_memory" : "16683249664",
        "xpack.installed" : "true",
        "ml.max_open_jobs" : "20"
      },
      "settings" : {
        "cluster" : {
          "initial_master_nodes" : "es-xpack-disabled-0",
          "name" : "es-xpack-disabled",
          "election" : {
            "strategy" : "supports_voting_only"
          }
        },
        "node" : {
          "name" : "es-xpack-disabled-0",
          "attr" : {
            "xpack" : {
              "installed" : "true"
            },
            "ml" : {
              "machine_memory" : "16683249664",
              "max_open_jobs" : "20"
            }
          },
          "data" : "true",
          "ingest" : "true",
          "master" : "true"
        },
        "path" : {
          "logs" : "/usr/share/elasticsearch/logs",
          "home" : "/usr/share/elasticsearch"
        },
        "discovery" : {
          "seed_hosts" : "es-xpack-disabled-master"
        },
        "client" : {
          "type" : "node"
        },
        "http" : {
          "type" : "security4",
          "type.default" : "netty4"
        },
        "transport" : {
          "type" : "security4",
          "features" : {
            "x-pack" : "true"
          },
          "type.default" : "netty4"
        },
        "network" : {
          "host" : "0.0.0.0"
        }
      }
    }
  }
}
```

Here, `xpack.security.enabled` is set to `false`. As a result, `xpack` security configurations are missing from the node settings.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo es/es-xpack-disabled -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo es/es-xpack-disabled

kubectl delete ns demo
```

## Next Steps

- Learn how to [create TLS certificates](/docs/guides/elasticsearch/x-pack/issue-certificate.md).
- Learn how to generate [x-pack configuration](/docs/guides/elasticsearch/x-pack/configuration.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
