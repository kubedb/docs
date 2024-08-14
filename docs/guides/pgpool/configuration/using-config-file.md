---
title: Run Pgpool with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: pp-using-config-file-configuration
    name: Customize Configurations
    parent: pp-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for Pgpool. This tutorial will show you how to use KubeDB to run a Pgpool with custom configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/pgpool](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgpool) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

Pgpool allows configuring via configuration file. The default configuration file for Pgpool deployed by `KubeDB` can be found in `opt/pgpool-II/etc/pgpool.conf`. When `spec.configSecret` is set to pgpool, KubeDB operator will get the secret and after that it will validate the values of the secret and then will keep the validated customizable configurations from the user and merge it with the remaining default config. After all that this secret will be mounted to Pgpool for use it as the configuration file.

> To learn available configuration option of Pgpool see [Configuration Options](https://www.pgpool.net/docs/45/en/html/runtime-config.html).

At first, you have to create a secret with your configuration file contents as the value of this key `pgpool.conf`. Then, you have to specify the name of this secret in `spec.configSecret.name` section while creating Pgpool crd.

## Prepare Postgres
For a Pgpool surely we will need a Postgres server so, prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgpool/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.


## Custom Configuration

At first, create `pgpool.conf` file containing required configuration settings.

```bash
$ cat pgpool.conf
num_init_children = 6
max_pool = 65
child_life_time = 400
```

Now, create the secret with this configuration file.

```bash
$ kubectl create secret generic -n demo pp-configuration --from-file=./pgpool.conf
secret/pp-configuration created
```

Verify the secret has the configuration file.

```bash
$  kubectl get secret -n demo pp-configuration -o yaml
apiVersion: v1
data:
  pgpool.conf: bnVtX2luaXRfY2hpbGRyZW4gPSA2Cm1heF9wb29sID0gNjUKY2hpbGRfbGlmZV90aW1lID0gNDAwCg==
kind: Secret
metadata:
  creationTimestamp: "2024-07-29T12:40:48Z"
  name: pp-configuration
  namespace: demo
  resourceVersion: "32076"
  uid: 80f5324a-9a65-4801-b136-21d2fa001b12
type: Opaque

$ echo bnVtX2luaXRfY2hpbGRyZW4gPSA2Cm1heF9wb29sID0gNjUKY2hpbGRfbGlmZV90aW1lID0gNDAwCg== | base64 -d
num_init_children = 6
max_pool = 65
child_life_time = 400
```

Now, create Pgpool crd specifying `spec.configSecret` field.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pp-custom-config
  namespace: demo
spec:
  version: "4.4.5"
  replicas: 1
  configSecret:
    name: pp-configuration
  postgresRef:
    name: ha-postgres
    namespace: demo
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/configuration/pgpool-config-file.yaml
pgpool.kubedb.com/pp-custom-config created
```

Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `pp-custom-config-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo pp-custom-config-0
NAME                 READY   STATUS    RESTARTS   AGE
pp-custom-config-0   1/1     Running   0          35s
```

Now, we will check if the pgpool has started with the custom configuration we have provided.

Now, you can exec into the pgpool pod and find if the custom configuration is there,

```bash
$ kubectl exec -it -n demo pp-custom-config-0 -- bash
pp-custom-config-0:/$ cat opt/pgpool-II/etc/pgpool.conf
backend_hostname0 = 'ha-postgres.demo.svc'
backend_port0 = 5432
backend_weight0 = 1
backend_flag0 = 'ALWAYS_PRIMARY|DISALLOW_TO_FAILOVER'
backend_hostname1 = 'ha-postgres-standby.demo.svc'
backend_port1 = 5432
backend_weight1 = 1
backend_flag1 = 'DISALLOW_TO_FAILOVER'
listen_addresses = *
log_per_node_statement = on
num_init_children = 6
max_pool = 65
child_life_time = '400'
child_max_connections = 0
connection_life_time = 0
client_idle_limit = 0
connection_cache = on
load_balance_mode = on
log_min_messages = 'warning'
statement_level_load_balance = 'off'
memory_cache_enabled = 'off'
enable_pool_hba = on
port = 9999
socket_dir = '/var/run/pgpool'
pcp_listen_addresses = *
pcp_port = 9595
pcp_socket_dir = '/var/run/pgpool'
sr_check_period = 0
health_check_period = 0
backend_clustering_mode = 'streaming_replication'
ssl = 'off'
failover_on_backend_error = 'off'
memqcache_oiddir = '/tmp/oiddir/'
allow_clear_text_frontend_auth = 'false'
failover_on_backend_error = 'off'
pp-custom-config-0:/$ exit
exit
```

As we can see from the configuration of running pgpool, the value of `num_init_children`, `max_pool` and `child_life_time` has been set to our desired value successfully.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo pp/pp-custom-config
kubectl delete -n demo secret pp-configuration
kubectl delete pg -n demo ha-postgres
kubectl delete ns demo
```

## Next Steps

- Monitor your Pgpool database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/pgpool/monitoring/using-prometheus-operator.md).
- Monitor your Pgpool database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/pgpool/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Pgpool object](/docs/guides/pgpool/concepts/pgpool.md).
- Detail concepts of [PgpoolVersion object](/docs/guides/pgpool/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
