---
title: Configuring Elasticsearch Combined Cluster
menu:
  docs_{{ .version }}:
    identifier: es-configuration-combined-cluster
    name: Combined Cluster
    parent: es-configuration
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Configure Elasticsearch Combined Cluster 

In Elasticsearch combined cluster, every node can perform as master, data, and ingest nodes simultaneously. In this tutorial, we will see how to configure a combined cluster.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install the KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create namespace demo
namespace/demo created

$ kubectl get namespace
NAME                 STATUS   AGE
demo                 Active   9s
```

> Note: YAML files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/elasticsearch/configuration/combined-cluster/yamls
) in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Available StorageClass

We will have to provide `StorageClass` in Elasticsearch CR specification. Check available `StorageClass` in your cluster using the following command,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  1h
```

Here, we have `standard` StorageClass in our cluster from [Local Path Provisioner](https://github.com/rancher/local-path-provisioner).

## Use Custom Configuration

Say we want to change the default log directory for our cluster and want to configure disk-based shard allocation. Let's create the `elasticsearch.yml` file with our desire configurations.

**elasticsearch.yml:**

```yaml
path:
  logs: "/usr/share/elasticsearch/data/new-logs-dir"
# For 100gb node space:
# Enable disk-based shard allocation
cluster.routing.allocation.disk.threshold_enabled: true
# prevent Elasticsearch from allocating shards to the node if less than the 15gb of space is available
cluster.routing.allocation.disk.watermark.low: 15gb
# relocate shards away from the node if the node has less than 10gb of free space
cluster.routing.allocation.disk.watermark.high: 10gb
# enforce a read-only index block if the node has less than 5gb of free space
cluster.routing.allocation.disk.watermark.flood_stage: 5gb
```

Let's create a k8s secret containing the above configuration where the file name will be the key and the file-content as the value:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: es-custom-config
  namespace: demo
stringData:
  elasticsearch.yml: |-
    path:
      logs: "/usr/share/elasticsearch/data/new-logs-dir"
    cluster.routing.allocation.disk.threshold_enabled: true
    cluster.routing.allocation.disk.watermark.low: 15gb
    cluster.routing.allocation.disk.watermark.high: 10gb
    cluster.routing.allocation.disk.watermark.flood_stage: 5gb
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/configuration/combined-cluster/yamls/config-secret.yaml
secret/es-custom-config created
```

Now that the config secret is created, it needs to be mention in the [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md) object's yaml:


```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-multinode
  namespace: demo
spec:
  version: xpack-8.11.1
  enableSSL: true
  replicas: 3
  configSecret:
    name: es-custom-config # mentioned here!
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 100Gi
  deletionPolicy: WipeOut
```

Now, create the Elasticsearch object by the following command:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/configuration/combined-cluster/yamls/es-combined.yaml
elasticsearch.kubedb.com/es-multinode created
```

Now, wait for the Elasticsearch to become ready:

```bash
$ kubectl get es -n demo -w
NAME           VERSION          STATUS         AGE
es-multinode   xpack-8.11.1   Provisioning   18s
es-multinode   xpack-8.11.1   Provisioning   2m5s
es-multinode   xpack-8.11.1   Ready          2m5s
```

## Verify Configuration

Let's connect to the Elasticsearch cluster that we have created and check the node settings to verify whether our configurations are applied or not:

Connect to the Cluster:

```bash
# Port-forward the service to local machine
$ kubectl port-forward -n demo svc/es-multinode 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```

Now, our Elasticsearch cluster is accessible at `localhost:9200`.

**Connection information:**

- Address: `localhost:9200`
- Username:

  ```bash
  $ kubectl get secret -n demo es-multinode-elastic-cred -o jsonpath='{.data.username}' | base64 -d
  elastic
  ```

- Password:

  ```bash
  $ kubectl get secret -n demo es-multinode-elastic-cred -o jsonpath='{.data.password}' | base64 -d
  ehG7*7SJZ0o9PA05
  ```

Now, we will query for settings of all nodes in an Elasticsearch cluster,

```bash
$ curl -XGET -k -u 'elastic:ehG7*7SJZ0o9PA05' "https://localhost:9200/_nodes/_all/settings?pretty"

```

This will return a large JSON with node settings. Here is the prettified JSON response,

```json
{
  "_nodes" : {
    "total" : 3,
    "successful" : 3,
    "failed" : 0
  },
  "cluster_name" : "es-multinode",
  "nodes" : {
    "_xWvqAU4QJeMaV4MayTgeg" : {
      "name" : "es-multinode-0",
      "transport_address" : "10.244.0.25:9300",
      "host" : "10.244.0.25",
      "ip" : "10.244.0.25",
      "version" : "7.9.1",
      "build_flavor" : "default",
      "build_type" : "docker",
      "build_hash" : "083627f112ba94dffc1232e8b42b73492789ef91",
      "roles" : [
        "data",
        "ingest",
        "master",
        "ml",
        "remote_cluster_client",
        "transform"
      ],
      "attributes" : {
        "ml.machine_memory" : "1073741824",
        "xpack.installed" : "true",
        "transform.node" : "true",
        "ml.max_open_jobs" : "20"
      },
      "settings" : {
        "cluster" : {
          "name" : "es-multinode",
          "routing" : {
            "allocation" : {
              "disk" : {
                "threshold_enabled" : "true",
                "watermark" : {
                  "low" : "15gb",
                  "flood_stage" : "5gb",
                  "high" : "10gb"
                }
              }
            }
          },
          "election" : {
            "strategy" : "supports_voting_only"
          },
          "initial_master_nodes" : "es-multinode-0,es-multinode-1,es-multinode-2"
        },
        "node" : {
          "name" : "es-multinode-0",
          "attr" : {
            "transform" : {
              "node" : "true"
            },
            "xpack" : {
              "installed" : "true"
            },
            "ml" : {
              "machine_memory" : "1073741824",
              "max_open_jobs" : "20"
            }
          },
          "data" : "true",
          "ingest" : "true",
          "master" : "true"
        },
        "path" : {
          "logs" : "/usr/share/elasticsearch/data/new-logs-dir",
          "home" : "/usr/share/elasticsearch"
        },
        "discovery" : {
          "seed_hosts" : "es-multinode-master"
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
    "0q1IcSSARwu9HrQmtvjDGA" : {
      "name" : "es-multinode-1",
      "transport_address" : "10.244.0.27:9300",
      "host" : "10.244.0.27",
      "ip" : "10.244.0.27",
      "version" : "7.9.1",
      "build_flavor" : "default",
      "build_type" : "docker",
      "build_hash" : "083627f112ba94dffc1232e8b42b73492789ef91",
      "roles" : [
        "data",
        "ingest",
        "master",
        "ml",
        "remote_cluster_client",
        "transform"
      ],
      "attributes" : {
        "ml.machine_memory" : "1073741824",
        "ml.max_open_jobs" : "20",
        "xpack.installed" : "true",
        "transform.node" : "true"
      },
      "settings" : {
        "cluster" : {
          "name" : "es-multinode",
          "routing" : {
            "allocation" : {
              "disk" : {
                "threshold_enabled" : "true",
                "watermark" : {
                  "low" : "15gb",
                  "flood_stage" : "5gb",
                  "high" : "10gb"
                }
              }
            }
          },
          "election" : {
            "strategy" : "supports_voting_only"
          },
          "initial_master_nodes" : "es-multinode-0,es-multinode-1,es-multinode-2"
        },
        "node" : {
          "name" : "es-multinode-1",
          "attr" : {
            "transform" : {
              "node" : "true"
            },
            "xpack" : {
              "installed" : "true"
            },
            "ml" : {
              "machine_memory" : "1073741824",
              "max_open_jobs" : "20"
            }
          },
          "data" : "true",
          "ingest" : "true",
          "master" : "true"
        },
        "path" : {
          "logs" : "/usr/share/elasticsearch/data/new-logs-dir",
          "home" : "/usr/share/elasticsearch"
        },
        "discovery" : {
          "seed_hosts" : "es-multinode-master"
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
    "ITvdnOcERwuG0qBmBJLaww" : {
      "name" : "es-multinode-2",
      "transport_address" : "10.244.0.29:9300",
      "host" : "10.244.0.29",
      "ip" : "10.244.0.29",
      "version" : "7.9.1",
      "build_flavor" : "default",
      "build_type" : "docker",
      "build_hash" : "083627f112ba94dffc1232e8b42b73492789ef91",
      "roles" : [
        "data",
        "ingest",
        "master",
        "ml",
        "remote_cluster_client",
        "transform"
      ],
      "attributes" : {
        "ml.machine_memory" : "1073741824",
        "ml.max_open_jobs" : "20",
        "xpack.installed" : "true",
        "transform.node" : "true"
      },
      "settings" : {
        "cluster" : {
          "name" : "es-multinode",
          "routing" : {
            "allocation" : {
              "disk" : {
                "threshold_enabled" : "true",
                "watermark" : {
                  "low" : "15gb",
                  "flood_stage" : "5gb",
                  "high" : "10gb"
                }
              }
            }
          },
          "election" : {
            "strategy" : "supports_voting_only"
          },
          "initial_master_nodes" : "es-multinode-0,es-multinode-1,es-multinode-2"
        },
        "node" : {
          "name" : "es-multinode-2",
          "attr" : {
            "transform" : {
              "node" : "true"
            },
            "xpack" : {
              "installed" : "true"
            },
            "ml" : {
              "machine_memory" : "1073741824",
              "max_open_jobs" : "20"
            }
          },
          "data" : "true",
          "ingest" : "true",
          "master" : "true"
        },
        "path" : {
          "logs" : "/usr/share/elasticsearch/data/new-logs-dir",
          "home" : "/usr/share/elasticsearch"
        },
        "discovery" : {
          "seed_hosts" : "es-multinode-master"
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

Here we can see that our given configuration is merged to the default configurations.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete elasticsearch -n demo es-multinode 

$ kubectl delete secret -n demo es-custom-config 

$ kubectl delete namespace demo
```

## Next Steps
