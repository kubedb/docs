---
title: Run Pgpool with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: pp-using-init-configuration
    name: Customize Init Config
    parent: pp-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Init Configuration

KubeDB supports providing custom configuration for Pgpool while initializing the Pgpool. This tutorial will show you how to use KubeDB to run a Pgpool with init configuration.

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

Pgpool allows configuring via configuration file. The default configuration file for Pgpool deployed by `KubeDB` can be found in `opt/pgpool-II/etc/pgpool.conf`. When `spec.configuration` 
is set to pgpool, KubeDB operator will get the secret and after that it will validate the values of the secret and then will keep the validated customizable configurations from the user
and merge it with the remaining default config. After all that this secret will be mounted to Pgpool for use it as the configuration file.

So, if you do not want to use a configuration file and secret for custom configuration you can use this `spec.configuration.inline` field to provide any Pgpool config wih key value pair. The KubeDB operator will validate these configs provided and will merge with the default configs and make a configuration secret to mount to the Pgpool.

> To learn available configuration option of Pgpool see [Configuration Options](https://www.pgpool.net/docs/45/en/html/runtime-config.html).

## Prepare Postgres
For a Pgpool surely we will need a Postgres server so, prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgpool/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.


## Init Configuration
Now, create Pgpool crd specifying `spec.configuration.inline` field.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pp-init-config
  namespace: demo
spec:
  version: "4.4.5"
  replicas: 1
  postgresRef:
    name: ha-postgres
    namespace: demo
  configuration:
    inline:
      pgpool.conf: |
        num_init_children=6
        max_pooL=65
        child_life_time=400
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/configuration/pgpool-init-config.yaml
pgpool.kubedb.com/pp-init-config created
```

Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `pp-custom-config-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo pp-init-config-0
NAME               READY   STATUS    RESTARTS   AGE
pp-init-config-0   1/1     Running   0          2m31s
```
Now check the config secret KubeDB operator has created and check out `pgpool.conf`.
```bash
$  kubectl get secret -n demo pp-init-config-config -o yaml
apiVersion: v1
data:
  pgpool.conf: YmFja2VuZF9ob3N0bmFtZTAgPSAnaGEtcG9zdGdyZXMuZGVtby5zdmMnCmJhY2tlbmRfcG9ydDAgPSA1NDMyCmJhY2tlbmRfd2VpZ2h0MCA9IDEKYmFja2VuZF9mbGFnMCA9ICdBTFdBWVNfUFJJTUFSWXxESVNBTExPV19UT19GQUlMT1ZFUicKYmFja2VuZF9ob3N0bmFtZTEgPSAnaGEtcG9zdGdyZXMtc3RhbmRieS5kZW1vLnN2YycKYmFja2VuZF9wb3J0MSA9IDU0MzIKYmFja2VuZF93ZWlnaHQxID0gMQpiYWNrZW5kX2ZsYWcxID0gJ0RJU0FMTE9XX1RPX0ZBSUxPVkVSJwpudW1faW5pdF9jaGlsZHJlbiA9IDYKbWF4X3Bvb2wgPSA2NQpjaGlsZF9saWZlX3RpbWUgPSA0MDAKZW5hYmxlX3Bvb2xfaGJhID0gb24KbGlzdGVuX2FkZHJlc3NlcyA9ICoKcG9ydCA9IDk5OTkKc29ja2V0X2RpciA9ICcvdmFyL3J1bi9wZ3Bvb2wnCnBjcF9saXN0ZW5fYWRkcmVzc2VzID0gKgpwY3BfcG9ydCA9IDk1OTUKcGNwX3NvY2tldF9kaXIgPSAnL3Zhci9ydW4vcGdwb29sJwpsb2dfcGVyX25vZGVfc3RhdGVtZW50ID0gb24Kc3JfY2hlY2tfcGVyaW9kID0gMApoZWFsdGhfY2hlY2tfcGVyaW9kID0gMApiYWNrZW5kX2NsdXN0ZXJpbmdfbW9kZSA9ICdzdHJlYW1pbmdfcmVwbGljYXRpb24nCmNoaWxkX21heF9jb25uZWN0aW9ucyA9IDAKY29ubmVjdGlvbl9saWZlX3RpbWUgPSAwCmNsaWVudF9pZGxlX2xpbWl0ID0gMApjb25uZWN0aW9uX2NhY2hlID0gb24KbG9hZF9iYWxhbmNlX21vZGUgPSBvbgpzc2wgPSAnb2ZmJwpmYWlsb3Zlcl9vbl9iYWNrZW5kX2Vycm9yID0gJ29mZicKbG9nX21pbl9tZXNzYWdlcyA9ICd3YXJuaW5nJwpzdGF0ZW1lbnRfbGV2ZWxfbG9hZF9iYWxhbmNlID0gJ29mZicKbWVtb3J5X2NhY2hlX2VuYWJsZWQgPSAnb2ZmJwptZW1xY2FjaGVfb2lkZGlyID0gJy90bXAvb2lkZGlyLycKYWxsb3dfY2xlYXJfdGV4dF9mcm9udGVuZF9hdXRoID0gJ2ZhbHNlJwo=
  pool_hba.conf: I1RZUEUgICAgICBEQVRBQkFTRSAgICAgICAgVVNFUiAgICAgICAgICAgIEFERFJFU1MgICAgICAgICAgICAgICAgIE1FVEhPRAojICJsb2NhbCIgaXMgZm9yIFVuaXggZG9tYWluIHNvY2tldCBjb25uZWN0aW9ucyBvbmx5CmxvY2FsICAgICAgYWxsICAgICAgICAgICAgIGFsbCAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICB0cnVzdAojIElQdjQgbG9jYWwgY29ubmVjdGlvbnM6Cmhvc3QgICAgICAgICBhbGwgICAgICAgICAgICAgYWxsICAgICAgICAgICAgIDEyNy4wLjAuMS8zMiAgICAgICAgICAgIHRydXN0CiMgSVB2NiBsb2NhbCBjb25uZWN0aW9uczoKaG9zdCAgICAgICAgIGFsbCAgICAgICAgICAgICBhbGwgICAgICAgICAgICAgOjoxLzEyOCAgICAgICAgICAgICAgICAgdHJ1c3QKbG9jYWwgICAgICAgIHBvc3RncmVzICAgICAgICBhbGwgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgdHJ1c3QKaG9zdCAgICAgICAgIHBvc3RncmVzICAgICAgICBhbGwgICAgICAgICAgICAgMTI3LjAuMC4xLzMyICAgICAgICAgICAgbWQ1Cmhvc3QgICAgICAgICBwb3N0Z3JlcyAgICAgICAgYWxsICAgICAgICAgICAgIDo6MS8xMjggICAgICAgICAgICAgICAgIG1kNQpob3N0ICAgICAgICAgYWxsICAgICAgICAgICAgIGFsbCAgICAgICAgICAgICAwLjAuMC4wLzAgICAgICAgICAgICAgICBtZDUKaG9zdCAgICAgICAgIHBvc3RncmVzICAgICAgICBwb3N0Z3JlcyAgICAgICAgMC4wLjAuMC8wICAgICAgICAgICAgICAgbWQ1Cmhvc3QgICAgICAgICBhbGwgICAgICAgICAgICAgYWxsICAgICAgICAgICAgIDo6LzAgICAgICAgICAgICAgICAgICAgIG1kNQpob3N0ICAgICAgICAgcG9zdGdyZXMgICAgICAgIHBvc3RncmVzICAgICAgICA6Oi8wICAgICAgICAgICAgICAgICAgICBtZDUK
kind: Secret
metadata:
  creationTimestamp: "2024-07-30T04:42:48Z"
  labels:
    app.kubernetes.io/component: connection-pooler
    app.kubernetes.io/instance: pp-init-config
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: pgpools.kubedb.com
  name: pp-init-config-config
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1alpha2
    blockOwnerDeletion: true
    controller: true
    kind: Pgpool
    name: pp-init-config
    uid: 7ec27da6-8d80-47d4-bc18-f53c8f9333a0
  resourceVersion: "91228"
  uid: d27154e6-b843-4c1d-b2af-79a80af38ca0
type: Opaque

$ echo YmFja2VuZF9ob3N0bmFtZTAgPSAnaGEtcG9zdGdyZXMuZGVtby5zdmMnCmJhY2tlbmRfcG9ydDAgPSA1NDMyCmJhY2tlbmRfd2VpZ2h0MCA9IDEKYmFja2VuZF9mbGFnMCA9ICdBTFdBWVNfUFJJTUFSWXxESVNBTExPV19UT19GQUlMT1ZFUicKYmFja2VuZF9ob3N0bmFtZTEgPSAnaGEtcG9zdGdyZXMtc3RhbmRieS5kZW1vLnN2YycKYmFja2VuZF9wb3J0MSA9IDU0MzIKYmFja2VuZF93ZWlnaHQxID0gMQpiYWNrZW5kX2ZsYWcxID0gJ0RJU0FMTE9XX1RPX0ZBSUxPVkVSJwpudW1faW5pdF9jaGlsZHJlbiA9IDYKbWF4X3Bvb2wgPSA2NQpjaGlsZF9saWZlX3RpbWUgPSA0MDAKZW5hYmxlX3Bvb2xfaGJhID0gb24KbGlzdGVuX2FkZHJlc3NlcyA9ICoKcG9ydCA9IDk5OTkKc29ja2V0X2RpciA9ICcvdmFyL3J1bi9wZ3Bvb2wnCnBjcF9saXN0ZW5fYWRkcmVzc2VzID0gKgpwY3BfcG9ydCA9IDk1OTUKcGNwX3NvY2tldF9kaXIgPSAnL3Zhci9ydW4vcGdwb29sJwpsb2dfcGVyX25vZGVfc3RhdGVtZW50ID0gb24Kc3JfY2hlY2tfcGVyaW9kID0gMApoZWFsdGhfY2hlY2tfcGVyaW9kID0gMApiYWNrZW5kX2NsdXN0ZXJpbmdfbW9kZSA9ICdzdHJlYW1pbmdfcmVwbGljYXRpb24nCmNoaWxkX21heF9jb25uZWN0aW9ucyA9IDAKY29ubmVjdGlvbl9saWZlX3RpbWUgPSAwCmNsaWVudF9pZGxlX2xpbWl0ID0gMApjb25uZWN0aW9uX2NhY2hlID0gb24KbG9hZF9iYWxhbmNlX21vZGUgPSBvbgpzc2wgPSAnb2ZmJwpmYWlsb3Zlcl9vbl9iYWNrZW5kX2Vycm9yID0gJ29mZicKbG9nX21pbl9tZXNzYWdlcyA9ICd3YXJuaW5nJwpzdGF0ZW1lbnRfbGV2ZWxfbG9hZF9iYWxhbmNlID0gJ29mZicKbWVtb3J5X2NhY2hlX2VuYWJsZWQgPSAnb2ZmJwptZW1xY2FjaGVfb2lkZGlyID0gJy90bXAvb2lkZGlyLycKYWxsb3dfY2xlYXJfdGV4dF9mcm9udGVuZF9hdXRoID0gJ2ZhbHNlJwo= | base64 -d
backend_hostname0 = 'ha-postgres.demo.svc'
backend_port0 = 5432
backend_weight0 = 1
backend_flag0 = 'ALWAYS_PRIMARY|DISALLOW_TO_FAILOVER'
backend_hostname1 = 'ha-postgres-standby.demo.svc'
backend_port1 = 5432
backend_weight1 = 1
backend_flag1 = 'DISALLOW_TO_FAILOVER'
num_init_children = 6
max_pool = 65
child_life_time = 400
enable_pool_hba = on
listen_addresses = *
port = 9999
socket_dir = '/var/run/pgpool'
pcp_listen_addresses = *
pcp_port = 9595
pcp_socket_dir = '/var/run/pgpool'
log_per_node_statement = on
sr_check_period = 0
health_check_period = 0
backend_clustering_mode = 'streaming_replication'
child_max_connections = 0
connection_life_time = 0
client_idle_limit = 0
connection_cache = on
load_balance_mode = on
ssl = 'off'
failover_on_backend_error = 'off'
log_min_messages = 'warning'
statement_level_load_balance = 'off'
memory_cache_enabled = 'off'
memqcache_oiddir = '/tmp/oiddir/'
allow_clear_text_frontend_auth = 'false'
```
Now, we will check if the pgpool has started with the init configuration we have provided.

Now, you can exec into the pgpool pod and find if the custom configuration is there,

```bash
$ kubectl exec -it -n demo pp-init-config-0 -- bash
pp-init-config-0:/$ cat opt/pgpool-II/etc/pgpool.conf
backend_hostname0 = 'ha-postgres.demo.svc'
backend_port0 = 5432
backend_weight0 = 1
backend_flag0 = 'ALWAYS_PRIMARY|DISALLOW_TO_FAILOVER'
backend_hostname1 = 'ha-postgres-standby.demo.svc'
backend_port1 = 5432
backend_weight1 = 1
backend_flag1 = 'DISALLOW_TO_FAILOVER'
num_init_children = 6
max_pool = 65
child_life_time = 400
enable_pool_hba = on
listen_addresses = *
port = 9999
socket_dir = '/var/run/pgpool'
pcp_listen_addresses = *
pcp_port = 9595
pcp_socket_dir = '/var/run/pgpool'
log_per_node_statement = on
sr_check_period = 0
health_check_period = 0
backend_clustering_mode = 'streaming_replication'
child_max_connections = 0
connection_life_time = 0
client_idle_limit = 0
connection_cache = on
load_balance_mode = on
ssl = 'off'
failover_on_backend_error = 'off'
log_min_messages = 'warning'
statement_level_load_balance = 'off'
memory_cache_enabled = 'off'
memqcache_oiddir = '/tmp/oiddir/'
allow_clear_text_frontend_auth = 'false'
failover_on_backend_error = 'off'
pp-init-config-0:/$ exit
exit
```

As we can see from the configuration of running pgpool, the value of `num_init_children`, `max_pool` and `child_life_time` has been set to our desired value successfully.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo pp/pp-init-config
kubectl delete pg -n demo ha-postgres
kubectl delete ns demo
```

## Next Steps

- Monitor your Pgpool database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/pgpool/monitoring/using-prometheus-operator.md).
- Monitor your Pgpool database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/pgpool/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Pgpool object](/docs/guides/pgpool/concepts/pgpool.md).
- Detail concepts of [PgpoolVersion object](/docs/guides/pgpool/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
