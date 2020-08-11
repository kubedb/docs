---
title: Run TLS Secured Elasticsearch
menu:
  docs_{{ .version }}:
    identifier: es-use-tls-x-pack
    name: Use TLS
    parent: es-x-pack
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Run TLS Secured Elasticsearch

X-Pack provides facility to secure your Elasticsearch cluster with TLS. By default, KubeDB does not enable TLS security. You have to enable it by setting `spec.enableSSL: true`. If TLS is enabled, only HTTPS calls are allowed to database server.

This tutorial will show you how to connect with Elasticsearch cluster using certificate when TLS is enabled.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created

$ kubectl get ns demo
NAME    STATUS  AGE
demo    Active  5s
```

> Note: YAML files used in this tutorial are stored in [docs/examples/elasticsearch](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/elasticsearch) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create Elasticsearch

In order to enable TLS, we have to set `spec.enableSSL` field of Elasticsearch object to `true`. Below is the YAML of Elasticsearch object that will be created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: Elasticsearch
metadata:
  name: ssl-elasticsearch
  namespace: demo
spec:
  version: 7.3.2
  replicas: 2
  enableSSL: true
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Let's create the Elasticsearch object we have shown above,

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/x-pack/ssl-elasticsearch.yaml
elasticsearch.kubedb.com/ssl-elasticsearch created
```

```console
$ kubectl get es -n demo ssl-elasticsearch
NAME                VERSION   STATUS    AGE
ssl-elasticsearch   7.3.2     Running   5m54s
```

## Connect to Elasticsearch Database

As we have enabled TLS for our Elasticsearch cluster, only HTTPS calls are allowed to the Elasticsearch server. So, we need to provide certificate to connect with Elasticsearch. If you do not provide certificate manually through `spec.certificateSecret` field of Elasticsearch object, KubeDB will create a secret `{elasticsearch name}-cert` with necessary certificates.

Let's check the certificates that has been created for Elasticsearch `ssl-elasticsearch` by KubeDB operator.

```console
$ kubectl get secret -n demo ssl-elasticsearch-cert -o yaml
```

```yaml
apiVersion: v1
data:
  client.jks: /u3+7QAAAAIAAAABAAAA...mVv0I52GubpXTAahXDo=
  node.jks: /u3+7QAAAAIAAAABAAAA...pn6opk0qoxabtPTP30c=
  root.jks: /u3+7QAAAAIAAAABAAAA...rjIEWtBA1IMnDcB2JJm5
  root.pem: LS0tLS1CRUdJTiBDRVJU...VElGSUNBVEUtLS0tLQo=
  sgadmin.jks: /u3+7QAAAAIAAAABAAAA...12OXut1U7gYnEyJsBg==
  key_pass: NnRhN3h2
kind: Secret
metadata:
  creationTimestamp: 2018-02-19T09:51:45Z
  labels:
    kubedb.com/kind: Elasticsearch
    kubedb.com/name: ssl-elasticsearch
  name: ssl-elasticsearch-cert
  namespace: demo
  resourceVersion: "754"
  selfLink: /api/v1/namespaces/demo/secrets/ssl-elasticsearch-cert
  uid: 7efdaf31-155a-11e8-a001-42010a8000d5
type: Opaque
```

Here, `root.pem` file is the root CA in `.pem` format. We will require to provide this file while sending REST request to the Elasticsearch server.

Let's forward port 9200 of `ssl-elasticsearch-0` pod. Run following command in a separate terminal,

```console
$ kubectl port-forward -n demo ssl-elasticsearch-0 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```

Now, we can connect with the database at `localhost:9200`.

**Connection information:**

- Address: `localhost:9200`
- Username: Run following command to get *username*

  ```console
  $ kubectl get secrets -n demo ssl-elasticsearch-auth -o jsonpath='{.data.\ADMIN_USERNAME}' | base64 -d
  elastic
  ```

- Password: Run following command to get *password*

  ```console
  $ kubectl get secrets -n demo ssl-elasticsearch-auth -o jsonpath='{.data.\ADMIN_PASSWORD}' | base64 -d
  err5ns7w
  ```

- Root CA: Run following command to get `root.pem` file

  ```console
  $ kubectl get secrets -n demo ssl-elasticsearch-cert -o jsonpath='{.data.\root\.pem}' | base64 --decode > root.pem
  ```

Now, let's check health of our Elasticsearch database.

```console
$ curl --user "elastic:err5ns7w" "https://localhost:9200/_cluster/health?pretty" --cacert root.pem
```

```json
{
  "cluster_name" : "ssl-elasticsearch",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 2,
  "number_of_data_nodes" : 2,
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
$ curl --user "elastic:err5ns7w"  "https://localhost:9200/_nodes/_all/settings?pretty" --cacert root.pem
{
  "_nodes" : {
    "total" : 2,
    "successful" : 2,
    "failed" : 0
  },
  "cluster_name" : "ssl-elasticsearch",
  "nodes" : {
    "RUZU2vafThaLJwt6AJgNUQ" : {
      "name" : "ssl-elasticsearch-0",
      "transport_address" : "10.4.1.109:9300",
      "host" : "10.4.1.109",
      "ip" : "10.4.1.109",
      "version" : "7.3.2",
      "build_flavor" : "default",
      "build_type" : "docker",
      "build_hash" : "508c38a",
      "roles" : [
        "master",
        "data",
        "ingest"
      ],
      "attributes" : {
        "ml.machine_memory" : "7841263616",
        "xpack.installed" : "true",
        "ml.max_open_jobs" : "20"
      },
      "settings" : {
        "cluster" : {
          "initial_master_nodes" : "ssl-elasticsearch-0,ssl-elasticsearch-1",
          "name" : "ssl-elasticsearch"
        },
        "node" : {
          "name" : "ssl-elasticsearch-0",
          "attr" : {
            "xpack" : {
              "installed" : "true"
            },
            "ml" : {
              "machine_memory" : "7841263616",
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
          "seed_hosts" : "ssl-elasticsearch-master"
        },
        "client" : {
          "type" : "node"
        },
        "http" : {
          "compression" : "false",
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
        "xpack" : {
          "security" : {
            "http" : {
              "ssl" : {
                "enabled" : "true"
              }
            },
            "enabled" : "true",
            "transport" : {
              "ssl" : {
                "enabled" : "true"
              }
            }
          }
        },
        "network" : {
          "host" : "0.0.0.0"
        }
      }
    },
    "I9aircHnRsqFqVgLkia3_A" : {
      "name" : "ssl-elasticsearch-1",
      "transport_address" : "10.4.0.174:9300",
      "host" : "10.4.0.174",
      "ip" : "10.4.0.174",
      "version" : "7.3.2",
      "build_flavor" : "default",
      "build_type" : "docker",
      "build_hash" : "508c38a",
      "roles" : [
        "master",
        "data",
        "ingest"
      ],
      "attributes" : {
        "ml.machine_memory" : "7841263616",
        "ml.max_open_jobs" : "20",
        "xpack.installed" : "true"
      },
      "settings" : {
        "cluster" : {
          "initial_master_nodes" : "ssl-elasticsearch-0,ssl-elasticsearch-1",
          "name" : "ssl-elasticsearch"
        },
        "node" : {
          "name" : "ssl-elasticsearch-1",
          "attr" : {
            "xpack" : {
              "installed" : "true"
            },
            "ml" : {
              "machine_memory" : "7841263616",
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
          "seed_hosts" : "ssl-elasticsearch-master"
        },
        "client" : {
          "type" : "node"
        },
        "http" : {
          "compression" : "false",
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
        "xpack" : {
          "security" : {
            "http" : {
              "ssl" : {
                "enabled" : "true"
              }
            },
            "enabled" : "true",
            "transport" : {
              "ssl" : {
                "enabled" : "true"
              }
            }
          }
        },
        "network" : {
          "host" : "0.0.0.0"
        }
      }
    }
  }
}
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
kubectl patch -n demo es/ssl-elasticsearch -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo es/ssl-elasticsearch

kubectl delete ns demo
```

## Next Steps

- Learn how to [create TLS certificates](/docs/guides/elasticsearch/x-pack/issue-certificate.md).
- Learn how to generate [x-pack configuration](/docs/guides/elasticsearch/x-pack/configuration.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
