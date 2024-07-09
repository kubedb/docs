---
title: X-Pack Configuration
menu:
  docs_{{ .version }}:
    identifier: es-configuration-xpack
    name: X-Pack Configuration
    parent: es-x-pack
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# X-Pack Configuration

X-Pack is an Elastic Stack extension that provides security along with other features. In KubeDB, X-Pack authentication can be used with elasticsearch `6.8` and `7.2+`. In this guide, we will show, how to use xpack authentication or disable it.

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

## X-Pack AuthPlugin

In 0.13.0 release, a new field is introduced to `ElasticsearchVersions` crd, named `authPlugin`. In prior this releases, `authPlugin` was part of `Elasticsearch` CRD spec, which is deprecated since 0.13.0-rc.1.

The `spec.authPlugin` is an required field in ElasticsearchVersion CRD, which specifies which plugin to use for authentication. Currently, this field accepts either `X-Pack` or `SearchGuard`.

To see, which authPlugin is used in the target ElasticsearchVersion, run the following command:

```bash
kubectl get elasticsearchversions 7.3.2 -o yaml
```

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: ElasticsearchVersion
metadata:
  name: xpack-8.11.1
spec:
  authPlugin: X-Pack
  db:
    image: kubedb/elasticsearch:7.9.1-xpack
  distribution: ElasticStack
  exporter:
    image: kubedb/elasticsearch_exporter:1.1.0
  initContainer:
    image: kubedb/toybox:0.8.4
    yqImage: kubedb/elasticsearch-init:7.9.1-xpack-v1
  podSecurityPolicies:
    databasePolicyName: elasticsearch-db
  stash:
    addon:
      backupTask:
        name: elasticsearch-backup-7.3.2
      restoreTask:
        name: elasticsearch-restore-7.3.2
  version: 7.9.1
```

## Changing authPlugin

To change authPlugin, it is recommended to create a new ElasticsearchVersion CRD. Then, use that elasticsearchVersion to install an Elasticsearch server with that authPlugin.

## Deploy with X-Pack

To deploy with X-Pack, you need to use an `ElasticsearchVersion` where `X-Pack` is set to `authPlugin`.

Here, we are going to use ElasticsearchVersion `7.3.2`, which is mentioned earlier in this guide.

Now, let's create an Elasticsearch server using the following yaml.

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: config-elasticsearch
  namespace: demo
spec:
  version: xpack-8.11.1
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/x-pack/config-elasticsearch.yaml
elasticsearch.kubedb.com/config-elasticsearch created
```

The deployed elasticsearch object specs, after the mutation is done by kubedb:

```yaml
$ kubectl get elasticsearch -n demo config-elasticsearch -o yaml

apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  creationTimestamp: "2019-09-30T08:34:10Z"
  finalizers:
  - kubedb.com
  generation: 3
  name: config-elasticsearch
  namespace: demo
  resourceVersion: "60830"
  selfLink: /apis/kubedb.com/v1/namespaces/demo/elasticsearches/config-elasticsearch
  uid: 13263dfa-e35d-11e9-85c8-42010a8c002f
spec:
  authSecret:
    name: config-elasticsearch-auth
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      resources: {}
      serviceAccountName: config-elasticsearch
  replicas: 1
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: Halt
  version: xpack-8.11.1
status:
  observedGeneration: 1$4210395375389091791
  phase: Running
```

As we can see, KubeDB has created a secret named `config-elasticsearch-auth`, which contains password for built-in user `elastic` .

## Manually Generated Password

If you want to provide your own password, you need to create a secret that contains two keys: `ADMIN_USERNAME`, `ADMIN_PASSWORD`.

```bash
$ export ADMIN_PASSWORD=admin-password
$ kubectl create secret generic -n demo config-elasticsearch-auth \
                --from-literal=ADMIN_USERNAME=elastic \
                --from-literal=ADMIN_PASSWORD=harderPASSWORD \
secret/config-elasticsearch-auth created
```

> Use this Secret `config-elasticsearch-auth` in `spec.authSecret` field of your Elasticsearch object while creating the elasticsearch for the 1st time. Changing the password after creating, won't work at this time.

## Connect to Elasticsearch Database

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created.

```bash
$ kubectl get es -n demo config-elasticsearch -o wide
NAME                   VERSION   STATUS    AGE
config-elasticsearch   7.3.2     Running   2m8s
```

To connect to the elasticsearch node, we are going to use port forward to the elasticsearch pod. Run following command on a separate terminal,

```bash
$ kubectl port-forward -n demo config-elasticsearch-0 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```

**Connection information:**

- Address: `localhost:9200`
- Username: Run following command to get *username*

  ```bash
  $ kubectl get secrets -n demo config-elasticsearch-auth -o jsonpath='{.data.\ADMIN_USERNAME}' | base64 -d
    elastic
  ```

- Password: Run following command to get *password*

  ```bash
  $ kubectl get secrets -n demo config-elasticsearch-auth -o jsonpath='{.data.\ADMIN_PASSWORD}' | base64 -d
    ruobj2eo
  ```

Firstly, try to connect to this database without providing any authentication. You will face the following error:

```bash
$ curl "localhost:9200/_cluster/health?pretty"
```

```json
{
  "error" : {
    "root_cause" : [
      {
        "type" : "security_exception",
        "reason" : "missing authentication credentials for REST request [/_cluster/health?pretty]",
        "header" : {
          "WWW-Authenticate" : "Basic realm=\"security\" charset=\"UTF-8\""
        }
      }
    ],
    "type" : "security_exception",
    "reason" : "missing authentication credentials for REST request [/_cluster/health?pretty]",
    "header" : {
      "WWW-Authenticate" : "Basic realm=\"security\" charset=\"UTF-8\""
    }
  },
  "status" : 401
}
```

Now, provide the authentication,

```json
$ curl --user elastic:ruobj2eo "localhost:9200/_cluster/health?pretty"
{
  "cluster_name" : "config-elasticsearch",
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
$ curl --user "elastic:ruobj2eo" "localhost:9200/_nodes/_all/settings?pretty"
{
  "_nodes": {
    "total": 1,
    "successful": 1,
    "failed": 0
  },
  "cluster_name": "config-elasticsearch",
  "nodes": {
    "LxLZBdU6SLemcv6mF1p2vw": {
      "name": "config-elasticsearch-0",
      "transport_address": "10.8.0.112:9300",
      "host": "10.8.0.112",
      "ip": "10.8.0.112",
      "version": "7.3.2",
      "build_flavor": "default",
      "build_type": "docker",
      "build_hash": "508c38a",
      "roles": [
        "master",
        "data",
        "ingest"
      ],
      "attributes": {
        "ml.machine_memory": "7841255424",
        "xpack.installed": "true",
        "ml.max_open_jobs": "20"
      },
      "settings": {
        "cluster": {
          "initial_master_nodes": "config-elasticsearch-0",
          "name": "config-elasticsearch"
        },
        "node": {
          "name": "config-elasticsearch-0",
          "attr": {
            "xpack": {
              "installed": "true"
            },
            "ml": {
              "machine_memory": "7841255424",
              "max_open_jobs": "20"
            }
          },
          "data": "true",
          "ingest": "true",
          "master": "true"
        },
        "path": {
          "logs": "/usr/share/elasticsearch/logs",
          "home": "/usr/share/elasticsearch"
        },
        "discovery": {
          "seed_hosts": "config-elasticsearch-master"
        },
        "client": {
          "type": "node"
        },
        "http": {
          "type": "security4",
          "type.default": "netty4"
        },
        "transport": {
          "type": "security4",
          "features": {
            "x-pack": "true"
          },
          "type.default": "netty4"
        },
        "xpack": {
          "security": {
            "http": {
              "ssl": {
                "enabled": "false"
              }
            },
            "enabled": "true",
            "transport": {
              "ssl": {
                "enabled": "true"
              }
            }
          }
        },
        "network": {
          "host": "0.0.0.0"
        }
      }
    }
  }
}
```

As you can see, `xpack.security.enabled` is set to true.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo es/config-elasticsearch -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo es/config-elasticsearch

kubectl delete ns demo
```

## Next Steps

- Learn how to use [ssl enabled](/docs/guides/elasticsearch/plugins/x-pack/use-tls.md) elasticsearch cluster with xpack.
