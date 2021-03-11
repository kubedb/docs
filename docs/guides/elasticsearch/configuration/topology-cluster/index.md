---
title: Configuring Elasticsearch Topology Cluster
menu:
  docs_{{ .version }}:
    identifier: es-configuration-topology-cluster
    name: Topology Cluster
    parent: es-configuration
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Configure Elasticsearch Topology Cluster 

In an Elasticsearch topology cluster, each node is assigned with a dedicated role such as master, data, and ingest. The cluster must have at least one master node, one data node, and one ingest node. In this tutorial, we will see how to configure a topology cluster.

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

Say we want to change the default log directories for our cluster and want to configure disk-based shard allocation. We also want that the log directory name should have node-role in it (ie. demonstrating node-role specific configurations).

If a user may want to provide node-role specific configurations, say configurations that will only be merged to master nodes. To achieve this, users need to add the node role as a prefix to the file name.

- Format: `<node-role>-<file-name>.extension`
- Samples:
  - `data-elasticsearch.yml`: Only applied to data nodes.
  - `master-jvm.options`: Only applied to master nodes.
  - `ingest-log4j2.properties`: Only applied to ingest nodes.
  - `elasticsearch.yml`: Empty node-role means it will be applied to all nodes.

Let's create the `elasticsearch.yml` files with our desire configurations.

**elasticsearch.yml** is for all nodes:

```yaml
node.processors: 2
```

**master-elasticsearch.yml** is for master nodes:

```yaml
path:
  logs: "/usr/share/elasticsearch/data/master-logs-dir"
```

**data-elasticsearch.yml** is for data nodes:

```yaml
path:
  logs: "/usr/share/elasticsearch/data/data-logs-dir"
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

**ingest-elasticsearch.yml** is for ingest nodes:

```yaml
path:
  logs: "/usr/share/elasticsearch/data/ingest-logs-dir"
```

Let's create a k8s secret containing the above configurations where the file name will be the key and the file-content as the value:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: es-custom-config
  namespace: demo
stringData:
  elasticsearch.yml: |-
    node.processors: 2
  master-elasticsearch.yml: |-
    path:
      logs: "/usr/share/elasticsearch/data/master-logs-dir"
  ingest-elasticsearch.yml: |-
    path:
      logs: "/usr/share/elasticsearch/data/ingest-logs-dir"
  data-elasticsearch.yml: |-
    path:
      logs: "/usr/share/elasticsearch/data/data-logs-dir"
    cluster.routing.allocation.disk.threshold_enabled: true
    cluster.routing.allocation.disk.watermark.low: 15gb
    cluster.routing.allocation.disk.watermark.high: 10gb
    cluster.routing.allocation.disk.watermark.flood_stage: 5gb
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/configuration/topology-cluster/yamls/config-secret.yaml
secret/es-custom-config created
```

Now that the config secret is created, it needs to be mention in the [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md) object's yaml:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-topology
  namespace: demo
spec:
  enableSSL: true 
  version: xpack-7.9.1-v1
  configSecret:
    name: es-custom-config # mentioned here!
  storageType: Durable
  terminationPolicy: WipeOut
  topology:
    master:
      replicas: 1
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      replicas: 1
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 100Gi
    ingest:
      replicas: 1
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
```

Now, create the Elasticsearch object by the following command:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/configuration/topology-cluster/yamls/es-topology.yaml 
elasticsearch.kubedb.com/es-topology created
```

Now, wait for the Elasticsearch to become ready:

```bash
$ kubectl get elasticsearch -n demo -w
NAME          VERSION          STATUS         AGE
es-topology   xpack-7.9.1-v1   Provisioning   12s
es-topology   xpack-7.9.1-v1   Provisioning   2m2s
es-topology   xpack-7.9.1-v1   Ready          2m2s
```

## Verify Configuration

Let's connect to the Elasticsearch cluster that we have created and check the node settings to verify whether our configurations are applied or not:

Connect to the Cluster:

```bash
# Port-forward the service to local machine
$ kubectl port-forward -n demo svc/es-topology 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```

Now, our Elasticsearch cluster is accessible at `localhost:9200`.

**Connection information:**

- Address: `localhost:9200`
- Username:

  ```bash
  $ kubectl get secret -n demo es-topology-elastic-cred -o jsonpath='{.data.username}' | base64 -d
  elastic
  ```

- Password:

  ```bash
  $ kubectl get secret -n demo es-topology-elastic-cred -o jsonpath='{.data.password}' | base64 -d
  F2sIde1TbZqOR_gF
  ```

Now, we will query for settings of all nodes in an Elasticsearch cluster,

```bash
$ curl -XGET -k -u 'elastic:F2sIde1TbZqOR_gF' "https://localhost:9200/_nodes/_all/settings?pretty"
```

This will return a large JSON with node settings. Here is the prettified JSON response,

```json
{
  "_nodes" : {
    "total" : 3,
    "successful" : 3,
    "failed" : 0
  },
  "cluster_name" : "es-topology",
  "nodes" : {
    "PnvWHS4tTZaNLX8yiUykEg" : {
      "name" : "es-topology-data-0",
      "transport_address" : "10.244.0.37:9300",
      "host" : "10.244.0.37",
      "ip" : "10.244.0.37",
      "version" : "7.9.1",
      "build_flavor" : "default",
      "build_type" : "docker",
      "build_hash" : "083627f112ba94dffc1232e8b42b73492789ef91",
      "roles" : [
        "data",
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
          "name" : "es-topology",
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
          }
        },
        "node" : {
          "name" : "es-topology-data-0",
          "processors" : "2",
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
          "ingest" : "false",
          "master" : "false"
        },
        "path" : {
          "logs" : "/usr/share/elasticsearch/data/data-logs-dir",
          "home" : "/usr/share/elasticsearch"
        },
        "discovery" : {
          "seed_hosts" : "es-topology-master"
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
    "5EeawayWTa6aw9D8pcYlGQ" : {
      "name" : "es-topology-ingest-0",
      "transport_address" : "10.244.0.36:9300",
      "host" : "10.244.0.36",
      "ip" : "10.244.0.36",
      "version" : "7.9.1",
      "build_flavor" : "default",
      "build_type" : "docker",
      "build_hash" : "083627f112ba94dffc1232e8b42b73492789ef91",
      "roles" : [
        "ingest",
        "ml",
        "remote_cluster_client"
      ],
      "attributes" : {
        "ml.machine_memory" : "1073741824",
        "xpack.installed" : "true",
        "transform.node" : "false",
        "ml.max_open_jobs" : "20"
      },
      "settings" : {
        "cluster" : {
          "name" : "es-topology",
          "election" : {
            "strategy" : "supports_voting_only"
          }
        },
        "node" : {
          "name" : "es-topology-ingest-0",
          "processors" : "2",
          "attr" : {
            "transform" : {
              "node" : "false"
            },
            "xpack" : {
              "installed" : "true"
            },
            "ml" : {
              "machine_memory" : "1073741824",
              "max_open_jobs" : "20"
            }
          },
          "data" : "false",
          "ingest" : "true",
          "master" : "false"
        },
        "path" : {
          "logs" : "/usr/share/elasticsearch/data/ingest-logs-dir",
          "home" : "/usr/share/elasticsearch"
        },
        "discovery" : {
          "seed_hosts" : "es-topology-master"
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
    "d2YO9jGNRzuPczGpITuxNA" : {
      "name" : "es-topology-master-0",
      "transport_address" : "10.244.0.38:9300",
      "host" : "10.244.0.38",
      "ip" : "10.244.0.38",
      "version" : "7.9.1",
      "build_flavor" : "default",
      "build_type" : "docker",
      "build_hash" : "083627f112ba94dffc1232e8b42b73492789ef91",
      "roles" : [
        "master",
        "ml",
        "remote_cluster_client"
      ],
      "attributes" : {
        "ml.machine_memory" : "1073741824",
        "ml.max_open_jobs" : "20",
        "xpack.installed" : "true",
        "transform.node" : "false"
      },
      "settings" : {
        "cluster" : {
          "initial_master_nodes" : "es-topology-master-0",
          "name" : "es-topology",
          "election" : {
            "strategy" : "supports_voting_only"
          }
        },
        "node" : {
          "name" : "es-topology-master-0",
          "processors" : "2",
          "attr" : {
            "transform" : {
              "node" : "false"
            },
            "xpack" : {
              "installed" : "true"
            },
            "ml" : {
              "machine_memory" : "1073741824",
              "max_open_jobs" : "20"
            }
          },
          "data" : "false",
          "ingest" : "false",
          "master" : "true"
        },
        "path" : {
          "logs" : "/usr/share/elasticsearch/data/master-logs-dir",
          "home" : "/usr/share/elasticsearch"
        },
        "discovery" : {
          "seed_hosts" : "es-topology-master"
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

Here we can see that our given configuration is merged to the default configurations. The common configuration `node.processors` is merged to all types of nodes. The node role-specific log directories are also configured. The disk-based shard allocation setting merged to data nodes.  

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete elasticsearch -n demo es-topology

$ kubectl delete secret -n demo es-custom-config 

$ kubectl delete namespace demo
```

## Next Steps
