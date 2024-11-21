---
title: Reconfigure Standalone PgBouncer Database
menu:
  docs_{{ .version }}:
    identifier: pb-reconfigure-pgbouncer
    name: PgBouncer Reconfigure
    parent: pb-reconfigure
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure PgBouncer

This guide will show you how to use `KubeDB` Ops-manager operator to reconfigure a PgBouncer.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [PgBouncer](/docs/guides/pgbouncer/concepts/pgbouncer.md)
  - [PgBouncerOpsRequest](/docs/guides/pgbouncer/concepts/opsrequest.md)
  - [Reconfigure Overview](/docs/guides/pgbouncer/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/pgbouncer](/docs/examples/pgbouncer) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Prepare Postgres
For a PgBouncer surely we will need a Postgres server so, prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgbouncer/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.

Now, we are going to deploy a  `PgBouncer` using a supported version by `KubeDB` operator. Then we are going to apply `PgBouncerOpsRequest` to reconfigure its configuration.

### Prepare PgBouncer

Now, we are going to deploy a `PgBouncer` with version `1.18.0`.

### Deploy PgBouncer 

At first, we will create `pgbouncer.ini` file containing required configuration settings.

```ini
$ cat pgbouncer.ini
[pgbouncer]
auth_type = scram-sha-256
```
Here, `auth_type` is set to `scram-sha-256`, whereas the default value is `md5`.

Now, we will create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo pb-custom-config --from-file=./pgbouncer.ini
secret/pb-custom-config created
```

In this section, we are going to create a PgBouncer object specifying `spec.configSecret` field to apply this custom configuration. Below is the YAML of the `PgBouncer` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  name: pb-custom
  namespace: demo
spec:
  replicas: 1
  version: "1.18.0"
  database:
    syncUsers: true
    databaseName: "postgres"
    databaseRef:
      name: "ha-postgres"
      namespace: demo
  connectionPool:
    poolMode: session
    port: 5432
    reservePoolSize: 5
    maxClientConnections: 87
    defaultPoolSize: 2
    minPoolSize: 1
  deletionPolicy: WipeOut
```

Let's create the `PgBouncer` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/reconfigure/pb-custom-config.yaml
pgbouncer.kubedb.com/pb-custom created
```

Now, wait until `pb-custom` has status `Ready`. i.e,

```bash
$ kubectl get pb -n demo
NAME        TYPE                  VERSION   STATUS   AGE
pb-custom   kubedb.com/v1         1.18.0    Ready    112s
```

Now, we will check if the pgbouncer has started with the custom configuration we have provided.

Now, you can exec into the pgbouncer pod and find if the custom configuration is there,

```bash
$ kubectl exec -it -n demo pb-custom-0 -- bash
pb-custom-0:/$ cat opt/pgbouncer/etc/pgbouncer.ini
backend_hostname0 = 'ha-postgres.demo.svc'
backend_port0 = 5432
backend_weight0 = 1
backend_flag0 = 'ALWAYS_PRIMARY|DISALLOW_TO_FAILOVER'
backend_hostname1 = 'ha-postgres-standby.demo.svc'
backend_port1 = 5432
backend_weight1 = 1
backend_flag1 = 'DISALLOW_TO_FAILOVER'
max_pool = 60
enable_pool_hba = on
listen_addresses = *
port = 9999
socket_dir = '/var/run/pgbouncer'
pcp_listen_addresses = *
pcp_port = 9595
pcp_socket_dir = '/var/run/pgbouncer'
log_per_node_statement = on
sr_check_period = 0
health_check_period = 0
backend_clustering_mode = 'streaming_replication'
num_init_children = 5
child_life_time = 300
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
pb-custom-0:/$ exit
exit
```

As we can see from the configuration of running pgbouncer, the value of `auth_type` has been set to `scram-sha-256`.

### Reconfigure using new secret

Now we will reconfigure this pgbouncer to set `auth_type` to `md5`. 

Now, we will edit the `pgbouncer.ini` file containing required configuration settings.

```ini
$ cat pgbouncer.ini
auth_type=md5
```

Then, we will create a new secret with this configuration file.

```bash
$ kubectl create secret generic -n demo new-custom-config --from-file=./pgbouncer.ini
secret/new-custom-config created
```

#### Create PgBouncerOpsRequest

Now, we will use this secret to replace the previous secret using a `PgBouncerOpsRequest` CR. The `PgBouncerOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pbops-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: pb-custom
  configuration:
    pgbouncer:
      configSecret:
        name: new-custom-config
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `pb-csutom` pgbouncer.
- `spec.type` specifies that we are performing `Reconfigure` on our pgbouncer.
- `spec.configuration.pgbouncer.configSecret.name` specifies the name of the new secret.
- Have a look [here](/docs/guides/pgbouncer/concepts/opsrequest.md#spectimeout) on the respective sections to understand the `timeout` & `apply` fields.

Let's create the `PgBouncerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/reconfigure/pbops-reconfigure.yaml
pgbounceropsrequest.ops.kubedb.com/pbops-reconfigure created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Ops-manager operator will update the `configSecret` of `PgBouncer` object.

Let's wait for `PgBouncerOpsRequest` to be `Successful`.  Run the following command to watch `PgBouncerOpsRequest` CR,

```bash
$ watch kubectl get pgbounceropsrequest -n demo
Every 2.0s: kubectl get pgbounceropsrequest -n demo
NAME                TYPE          STATUS       AGE
pbops-reconfigure   Reconfigure   Successful   63s
```

We can see from the above output that the `PgBouncerOpsRequest` has succeeded. If we describe the `PgBouncerOpsRequest` we will get an overview of the steps that were followed to reconfigure the pgbouncer.

```bash
$ kubectl describe pgbounceropsrequest -n demo pbops-reconfigure
Name:         pbops-reconfigure
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgBouncerOpsRequest
Metadata:
  Creation Timestamp:  2024-07-30T05:42:56Z
  Generation:          1
  Resource Version:    95239
  UID:                 54a12624-048c-49a6-b852-6286da587535
Spec:
  Apply:  IfReady
  Configuration:
    Config Secret:
      Name:  new-custom-config
  Database Ref:
    Name:   pb-custom
  Timeout:  5m
  Type:     Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-07-30T05:42:56Z
    Message:               PgBouncer ops-request has started to `Reconfigure` the PgBouncer nodes
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-07-30T05:42:59Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-07-30T05:43:00Z
    Message:               Successfully updated PetSet
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-30T05:43:00Z
    Message:               Successfully updated PgBouncer
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-07-30T05:43:45Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-07-30T05:43:05Z
    Message:               get pod; ConditionStatus:True; PodName:pb-custom-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pb-custom-0
    Last Transition Time:  2024-07-30T05:43:05Z
    Message:               evict pod; ConditionStatus:True; PodName:pb-custom-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pb-custom-0
    Last Transition Time:  2024-07-30T05:43:40Z
    Message:               check pod running; ConditionStatus:True; PodName:pb-custom-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--pb-custom-0
    Last Transition Time:  2024-07-30T05:43:45Z
    Message:               Successfully completed the reconfigure for PgBouncer
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                         Age   From                         Message
  ----     ------                                                         ----  ----                         -------
  Normal   Starting                                                       100s  KubeDB Ops-manager Operator  Start processing for PgBouncerOpsRequest: demo/pbops-reconfigure
  Normal   Starting                                                       100s  KubeDB Ops-manager Operator  Pausing PgBouncer databse: demo/pb-custom
  Normal   Successful                                                     100s  KubeDB Ops-manager Operator  Successfully paused PgBouncer database: demo/pb-custom for PgBouncerOpsRequest: pbops-reconfigure
  Normal   UpdatePetSets                                                  96s   KubeDB Ops-manager Operator  Successfully updated PetSet
  Normal   UpdatePetSets                                                  96s   KubeDB Ops-manager Operator  Successfully updated PetSet
  Normal   UpdateDatabase                                                 96s   KubeDB Ops-manager Operator  Successfully updated PgBouncer
  Warning  get pod; ConditionStatus:True; PodName:pb-custom-0             91s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pb-custom-0
  Warning  evict pod; ConditionStatus:True; PodName:pb-custom-0           91s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pb-custom-0
  Warning  check pod running; ConditionStatus:False; PodName:pb-custom-0  86s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:pb-custom-0
  Warning  check pod running; ConditionStatus:True; PodName:pb-custom-0   56s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:pb-custom-0
  Normal   RestartPods                                                    51s   KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                       51s   KubeDB Ops-manager Operator  Resuming PgBouncer database: demo/pb-custom
  Normal   Successful                                                     51s   KubeDB Ops-manager Operator  Successfully resumed PgBouncer database: demo/pb-custom for PgBouncerOpsRequest: pbops-reconfigure
```

Now let's exec into the pgbouncer pod and check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo pb-custom-0 -- bash
pb-custom-0:/$ cat opt/pgbouncer/etc/pgbouncer.ini
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
num_init_children = 5
max_pool = 50
child_life_time = '300'
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
socket_dir = '/var/run/pgbouncer'
pcp_listen_addresses = *
pcp_port = 9595
pcp_socket_dir = '/var/run/pgbouncer'
sr_check_period = 0
health_check_period = 0
backend_clustering_mode = 'streaming_replication'
ssl = 'off'
failover_on_backend_error = 'off'
memqcache_oiddir = '/tmp/oiddir/'
allow_clear_text_frontend_auth = 'false'
failover_on_backend_error = 'off'
pb-custom-0:/$ exit
exit
```

As we can see from the configuration of running pgbouncer, the value of `auth_type` has been changed from `scram-sha-256` to `md5`. So the reconfiguration of the pgbouncer is successful.


### Reconfigure using apply config

Now we will reconfigure this pgbouncer again to set `auth_type` to `scram-sha-256`. This time we won't use a new secret. We will use the `applyConfig` field of the `PgBouncerOpsRequest`. This will merge the new config in the existing secret.

#### Create PgBouncerOpsRequest

Now, we will use the new configuration in the `data` field in the `PgBouncerOpsRequest` CR. The `PgBouncerOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pbops-reconfigure-apply
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: pb-custom
  configuration:
    pgbouncer:
      applyConfig:
        pgbouncer.ini: |-
          [pgbouncer]
          auth_type=scram-sha-256
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `pb-custom` pgbouncer.
- `spec.type` specifies that we are performing `Reconfigure` on our pgbouncer.
- `spec.configuration.pgbouncer.applyConfig` specifies the new configuration that will be merged in the existing secret.

Let's create the `PgBouncerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/reconfigure/pbops-reconfigure-apply.yaml
pgbounceropsrequest.ops.kubedb.com/pbops-reconfigure-apply created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Ops-manager operator will merge this new config with the existing configuration.

Let's wait for `PgBouncerOpsRequest` to be `Successful`.  Run the following command to watch `PgBouncerOpsRequest` CR,

```bash
$ watch kubectl get pgbounceropsrequest -n demo
Every 2.0s: kubectl get pgbounceropsrequest -n demo
NAME                      TYPE          STATUS       AGE
pbops-reconfigure         Reconfigure   Successful   9m15s
pbops-reconfigure-apply   Reconfigure   Successful   53s
```

We can see from the above output that the `PgBouncerOpsRequest` has succeeded. If we describe the `PgBouncerOpsRequest` we will get an overview of the steps that were followed to reconfigure the pgbouncer.

```bash
$ kubectl describe pgbounceropsrequest -n demo pbops-reconfigure-apply
Name:         pbops-reconfigure-apply
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgBouncerOpsRequest
Metadata:
  Creation Timestamp:  2024-07-30T05:51:18Z
  Generation:          1
  Resource Version:    95874
  UID:                 92b0f18c-a329-4bb7-85d0-ef66f32bf57a
Spec:
  Apply:  IfReady
  Configuration:
    Apply Config:
      pgbouncer.ini:  max_pool = 75
  Database Ref:
    Name:   pb-custom
  Timeout:  5m
  Type:     Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-07-30T05:51:18Z
    Message:               PgBouncer ops-request has started to `Reconfigure` the PgBouncer nodes
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-07-30T05:51:21Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-07-30T05:51:21Z
    Message:               Successfully updated PetSet
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-30T05:51:22Z
    Message:               Successfully updated PgBouncer
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-07-30T05:52:07Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-07-30T05:51:27Z
    Message:               get pod; ConditionStatus:True; PodName:pb-custom-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pb-custom-0
    Last Transition Time:  2024-07-30T05:51:27Z
    Message:               evict pod; ConditionStatus:True; PodName:pb-custom-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pb-custom-0
    Last Transition Time:  2024-07-30T05:52:02Z
    Message:               check pod running; ConditionStatus:True; PodName:pb-custom-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--pb-custom-0
    Last Transition Time:  2024-07-30T05:52:07Z
    Message:               Successfully completed the reconfigure for PgBouncer
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                         Age   From                         Message
  ----     ------                                                         ----  ----                         -------
  Normal   Starting                                                       77s   KubeDB Ops-manager Operator  Start processing for PgBouncerOpsRequest: demo/pbops-reconfigure-apply
  Normal   Starting                                                       77s   KubeDB Ops-manager Operator  Pausing PgBouncer databse: demo/pb-custom
  Normal   Successful                                                     77s   KubeDB Ops-manager Operator  Successfully paused PgBouncer database: demo/pb-custom for PgBouncerOpsRequest: pbops-reconfigure-apply
  Normal   UpdatePetSets                                                  74s   KubeDB Ops-manager Operator  Successfully updated PetSet
  Normal   UpdatePetSets                                                  73s   KubeDB Ops-manager Operator  Successfully updated PetSet
  Normal   UpdateDatabase                                                 73s   KubeDB Ops-manager Operator  Successfully updated PgBouncer
  Warning  get pod; ConditionStatus:True; PodName:pb-custom-0             68s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pb-custom-0
  Warning  evict pod; ConditionStatus:True; PodName:pb-custom-0           68s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pb-custom-0
  Warning  check pod running; ConditionStatus:False; PodName:pb-custom-0  63s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:pb-custom-0
  Warning  check pod running; ConditionStatus:True; PodName:pb-custom-0   33s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:pb-custom-0
  Normal   RestartPods                                                    28s   KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                       28s   KubeDB Ops-manager Operator  Resuming PgBouncer database: demo/pb-custom
  Normal   Successful                                                     28s   KubeDB Ops-manager Operator  Successfully resumed PgBouncer database: demo/pb-custom for PgBouncerOpsRequest: pbops-reconfigure-apply
```

Now let's exec into the pgbouncer pod and check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo pb-custom-0 -- bash 
pb-custom-0:/$ cat opt/pgbouncer/etc/pgbouncer.ini
memory_cache_enabled = 'off'
num_init_children = 5
pcp_socket_dir = '/var/run/pgbouncer'
port = '9999'
enable_pool_hba = on
log_min_messages = 'warning'
pcp_port = '9595'
sr_check_period = 0
ssl = 'off'
backend_weight1 = 1
load_balance_mode = on
backend_weight0 = 1
backend_port0 = '5432'
connection_cache = on
backend_hostname1 = 'ha-postgres-standby.demo.svc'
health_check_period = 0
memqcache_oiddir = '/tmp/oiddir/'
statement_level_load_balance = 'off'
allow_clear_text_frontend_auth = 'false'
log_per_node_statement = on
backend_hostname0 = 'ha-postgres.demo.svc'
backend_flag1 = 'DISALLOW_TO_FAILOVER'
listen_addresses = *
failover_on_backend_error = 'off'
pcp_listen_addresses = *
child_max_connections = 0
socket_dir = '/var/run/pgbouncer'
backend_flag0 = 'ALWAYS_PRIMARY|DISALLOW_TO_FAILOVER'
backend_port1 = '5432'
backend_clustering_mode = 'streaming_replication'
connection_life_time = 0
child_life_time = '300'
max_pool = 75
client_idle_limit = 0
failover_on_backend_error = 'off'
pb-custom-0:/$ exit
exit
```

As we can see from the configuration of running pgbouncer, the value of `auth_type` has been changed from `md5` to `scram-sha-256`. So the reconfiguration of the pgbouncer using the `applyConfig` field is successful.


### Remove config

Now we will reconfigure this pgbouncer to remove the custom config provided and get it back to the default config. We will use the `removeCustomConfig` field of the `PgBouncerOpsRequest`. This will remove all the custom config provided and get the pgbouncer back to the default config.

#### Create PgBouncerOpsRequest

Now, we will use the `removeCustomConfig` field in the `PgBouncerOpsRequest` CR. The `PgBouncerOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pbops-reconfigure-remove
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: pb-custom
  configuration:
    pgbouncer:
      removeCustomConfig: true
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `pb-custom` pgbouncer.
- `spec.type` specifies that we are performing `Reconfigure` on our pgbouncer.
- `spec.configuration.pgbouncer.removeCustomConfig` specifies for boolean values to remove custom configuration.

Let's create the `PgBouncerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/reconfigure/pbops-reconfigure-remove.yaml
pgbounceropsrequest.ops.kubedb.com/pbops-reconfigure-remove created
```

#### Verify if the configuration is removed

If everything goes well, `KubeDB` Ops-manager operator will remove the custom configuration and move back to the default configuration.

Let's wait for `PgBouncerOpsRequest` to be `Successful`.  Run the following command to watch `PgBouncerOpsRequest` CR,

```bash
$ watch kubectl get pgbounceropsrequest -n demo
Every 2.0s: kubectl get pgbounceropsrequest -n demo
kubectl get pgbounceropsrequest -n demo
NAME                       TYPE          STATUS       AGE
pbops-reconfigure          Reconfigure   Successful   71m
pbops-reconfigure-apply    Reconfigure   Successful   63m
pbops-reconfigure-remove   Reconfigure   Successful   57s
```

We can see from the above output that the `PgBouncerOpsRequest` has succeeded. If we describe the `PgBouncerOpsRequest` we will get an overview of the steps that were followed to reconfigure the pgbouncer.

```bash
$ kubectl describe pgbounceropsrequest -n demo pbops-reconfigure-remove
Name:         pbops-reconfigure-remove
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgBouncerOpsRequest
Metadata:
  Creation Timestamp:  2024-07-30T06:53:27Z
  Generation:          1
  Resource Version:    99827
  UID:                 24c9cba3-5e85-40dc-96f3-373d2dd7a8ba
Spec:
  Apply:  IfReady
  Configuration:
    Remove Custom Config:  true
  Database Ref:
    Name:   pb-custom
  Timeout:  5m
  Type:     Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-07-30T06:53:28Z
    Message:               PgBouncer ops-request has started to `Reconfigure` the PgBouncer nodes
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-07-30T06:53:31Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-07-30T06:53:31Z
    Message:               Successfully updated PetSet
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-30T06:53:32Z
    Message:               Successfully updated PgBouncer
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-07-30T06:54:17Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-07-30T06:53:37Z
    Message:               get pod; ConditionStatus:True; PodName:pb-custom-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pb-custom-0
    Last Transition Time:  2024-07-30T06:53:37Z
    Message:               evict pod; ConditionStatus:True; PodName:pb-custom-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pb-custom-0
    Last Transition Time:  2024-07-30T06:54:12Z
    Message:               check pod running; ConditionStatus:True; PodName:pb-custom-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--pb-custom-0
    Last Transition Time:  2024-07-30T06:54:17Z
    Message:               Successfully completed the reconfigure for PgBouncer
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                         Age   From                         Message
  ----     ------                                                         ----  ----                         -------
  Normal   Starting                                                       74s   KubeDB Ops-manager Operator  Start processing for PgBouncerOpsRequest: demo/pbops-reconfigure-remove
  Normal   Starting                                                       74s   KubeDB Ops-manager Operator  Pausing PgBouncer databse: demo/pb-custom
  Normal   Successful                                                     74s   KubeDB Ops-manager Operator  Successfully paused PgBouncer database: demo/pb-custom for PgBouncerOpsRequest: pbops-reconfigure-remove
  Normal   UpdatePetSets                                                  71s   KubeDB Ops-manager Operator  Successfully updated PetSet
  Normal   UpdatePetSets                                                  70s   KubeDB Ops-manager Operator  Successfully updated PetSet
  Normal   UpdateDatabase                                                 70s   KubeDB Ops-manager Operator  Successfully updated PgBouncer
  Warning  get pod; ConditionStatus:True; PodName:pb-custom-0             65s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pb-custom-0
  Warning  evict pod; ConditionStatus:True; PodName:pb-custom-0           65s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pb-custom-0
  Warning  check pod running; ConditionStatus:False; PodName:pb-custom-0  60s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:pb-custom-0
  Warning  check pod running; ConditionStatus:True; PodName:pb-custom-0   30s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:pb-custom-0
  Normal   RestartPods                                                    25s   KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                       25s   KubeDB Ops-manager Operator  Resuming PgBouncer database: demo/pb-custom
  Normal   Successful                                                     25s   KubeDB Ops-manager Operator  Successfully resumed PgBouncer database: demo/pb-custom for PgBouncerOpsRequest: pbops-reconfigure-remove
```

Now let's exec into the pgbouncer pod and check the configuration.

```bash
$ kubectl exec -it -n demo pb-custom-0 -- bash 
pb-custom-0:/$ cat opt/pgbouncer/etc/pgbouncer.ini
backend_hostname0 = 'ha-postgres.demo.svc'
backend_port0 = 5432
backend_weight0 = 1
backend_flag0 = 'ALWAYS_PRIMARY|DISALLOW_TO_FAILOVER'
backend_hostname1 = 'ha-postgres-standby.demo.svc'
backend_port1 = 5432
backend_weight1 = 1
backend_flag1 = 'DISALLOW_TO_FAILOVER'
enable_pool_hba = on
listen_addresses = *
port = 9999
socket_dir = '/var/run/pgbouncer'
pcp_listen_addresses = *
pcp_port = 9595
pcp_socket_dir = '/var/run/pgbouncer'
log_per_node_statement = on
sr_check_period = 0
health_check_period = 0
backend_clustering_mode = 'streaming_replication'
num_init_children = 5
max_pool = 15
child_life_time = 300
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
pb-custom-0:/$ exit
exit
```

As we can see from the configuration of running pgbouncer, the value of `auth_type` has been changed from `scram-sha-256` to `md5`. So the reconfiguration of the pgbouncer using the `removeCustomConfig` field is successful.


## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:
```bash
kubectl delete -n demo pb/pb-custom
kubectl delete pgbounceropsrequest -n demo pbops-reconfigure  pbops-reconfigure-apply pbops-reconfigure-remove
kubectl delete pg -n demo ha-postgres
kubectl delete ns demo
```