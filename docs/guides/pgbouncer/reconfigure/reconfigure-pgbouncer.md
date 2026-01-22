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

In this section, we are going to create a PgBouncer object specifying `spec.configuration` field to apply this custom configuration. Below is the YAML of the `PgBouncer` CR that we are going to create,

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
$ kubectl exec -it -n demo pb-custom-0  -- /bin/sh
pb-custom-0:/$ cat etc/config/pgbouncer.ini
[databases]
postgres= host=ha-postgres.demo.svc port=5432 dbname=postgres

[pgbouncer]
max_client_conn = 87
default_pool_size = 2
min_pool_size = 1
max_db_connections = 1
logfile = /tmp/pgbouncer.log
listen_port = 5432
ignore_startup_parameters = extra_float_digits
pidfile = /tmp/pgbouncer.pid
listen_addr = *
reserve_pool_size = 5
reserve_pool_timeout = 5
auth_type = scram-sha-256
auth_file =  /var/run/pgbouncer/secret/userlist
admin_users =  pgbouncer
pool_mode = session
max_user_connections = 2
stats_period = 60
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
    configSecret:
      name: new-custom-config
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `pb-csutom` pgbouncer.
- `spec.type` specifies that we are performing `Reconfigure` on our pgbouncer.
- `spec.configuration.configSecret.name` specifies the name of the new secret.
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
  Creation Timestamp:  2024-11-28T10:06:23Z
  Generation:          1
  Resource Version:    86377
  UID:                 f96d088e-a32b-40eb-bd9b-ca15a8370548
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
    Last Transition Time:  2024-11-28T10:06:23Z
    Message:               Controller has started to Progress with Reconfigure of PgBouncerOpsRequest: demo/pbops-reconfigure
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2024-11-28T10:06:26Z
    Message:               paused pgbouncer database
    Observed Generation:   1
    Reason:                Paused
    Status:                True
    Type:                  Paused
    Last Transition Time:  2024-11-28T10:06:36Z
    Message:               Successfully updated PgBouncer
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-11-28T10:06:36Z
    Message:               Successfully updated PgBouncer backend secret
    Observed Generation:   1
    Reason:                UpdateBackendSecret
    Status:                True
    Type:                  UpdateBackendSecret
    Last Transition Time:  2024-11-28T10:06:41Z
    Message:               get pod; ConditionStatus:True; PodName:pb-custom-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pb-custom-0
    Last Transition Time:  2024-11-28T10:07:16Z
    Message:               volume mount check; ConditionStatus:True; PodName:pb-custom-0
    Observed Generation:   1
    Status:                True
    Type:                  VolumeMountCheck--pb-custom-0
    Last Transition Time:  2024-11-28T10:07:21Z
    Message:               reload config; ConditionStatus:True; PodName:pb-custom-0
    Observed Generation:   1
    Status:                True
    Type:                  ReloadConfig--pb-custom-0
    Last Transition Time:  2024-11-28T10:07:21Z
    Message:               Reloading performed successfully in PgBouncer: demo/pb-custom for PgBouncerOpsRequest: pbops-reconfigure
    Observed Generation:   1
    Reason:                ReloadPodsSucceeded
    Status:                True
    Type:                  ReloadPods
    Last Transition Time:  2024-11-28T10:07:21Z
    Message:               Successfully Reconfigured
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-11-28T10:07:21Z
    Message:               Controller has successfully completed  with Reconfigure of PgBouncerOpsRequest: demo/pbops-reconfigure
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                          Age   From                         Message
  ----     ------                                                          ----  ----                         -------
  Normal   Starting                                                        70s   KubeDB Ops-manager Operator  Start processing for PgBouncerOpsRequest: demo/pbops-reconfigure
  Normal   Starting                                                        70s   KubeDB Ops-manager Operator  Pausing PgBouncer databse: demo/pb-custom
  Normal   Successful                                                      70s   KubeDB Ops-manager Operator  Successfully paused PgBouncer database: demo/pb-custom for PgBouncerOpsRequest: pbops-reconfigure
  Warning  get pod; ConditionStatus:True; PodName:pb-custom-0              52s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pb-custom-0
  Warning  volume mount check; ConditionStatus:False; PodName:pb-custom-0  52s   KubeDB Ops-manager Operator  volume mount check; ConditionStatus:False; PodName:pb-custom-0
  Warning  get pod; ConditionStatus:True; PodName:pb-custom-0              47s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pb-custom-0
  Warning  get pod; ConditionStatus:True; PodName:pb-custom-0              42s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pb-custom-0
  Warning  get pod; ConditionStatus:True; PodName:pb-custom-0              37s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pb-custom-0
  Warning  get pod; ConditionStatus:True; PodName:pb-custom-0              32s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pb-custom-0
  Warning  get pod; ConditionStatus:True; PodName:pb-custom-0              27s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pb-custom-0
  Warning  get pod; ConditionStatus:True; PodName:pb-custom-0              22s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pb-custom-0
  Warning  get pod; ConditionStatus:True; PodName:pb-custom-0              17s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pb-custom-0
  Warning  volume mount check; ConditionStatus:True; PodName:pb-custom-0   17s   KubeDB Ops-manager Operator  volume mount check; ConditionStatus:True; PodName:pb-custom-0
  Warning  reload config; ConditionStatus:True; PodName:pb-custom-0        12s   KubeDB Ops-manager Operator  reload config; ConditionStatus:True; PodName:pb-custom-0
  Warning  reload config; ConditionStatus:True; PodName:pb-custom-0        12s   KubeDB Ops-manager Operator  reload config; ConditionStatus:True; PodName:pb-custom-0
  Normal   Successful                                                      12s   KubeDB Ops-manager Operator  Reloading performed successfully in PgBouncer: demo/pb-custom for PgBouncerOpsRequest: pbops-reconfigure
  Normal   Starting                                                        12s   KubeDB Ops-manager Operator  Resuming PgBouncer database: demo/pb-custom
  Normal   Successful                                                      12s   KubeDB Ops-manager Operator  Successfully resumed PgBouncer database: demo/pb-custom
  Normal   Successful                                                      12s   KubeDB Ops-manager Operator  Controller has Successfully Reconfigured PgBouncer databases: demo/pb-custom
```

Now let's exec into the pgbouncer pod and check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo pb-custom-0  -- /bin/sh
pb-custom-0:/$ cat etc/config/pgbouncer.ini
[databases]
postgres= host=ha-postgres.demo.svc port=5432 dbname=postgres

[pgbouncer]
max_db_connections = 1
logfile = /tmp/pgbouncer.log
listen_addr = *
admin_users =  pgbouncer
pool_mode = session
max_client_conn = 87
listen_port = 5432
ignore_startup_parameters = extra_float_digits
auth_file =  /var/run/pgbouncer/secret/userlist
default_pool_size = 2
min_pool_size = 1
max_user_connections = 2
stats_period = 60
auth_type = md5
pidfile = /tmp/pgbouncer.pid
reserve_pool_size = 5
reserve_pool_timeout = 5
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
- `spec.configuration.applyConfig` specifies the new configuration that will be merged in the existing secret.

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
  Creation Timestamp:  2024-11-28T10:11:52Z
  Generation:          1
  Resource Version:    86774
  UID:                 a4b8e8b5-0b82-4391-a8fe-66911aa5bee6
Spec:
  Apply:  IfReady
  Configuration:
    Pgbouncer:
      Apply Config:
        pgbouncer.ini:  [pgbouncer]
auth_type=scram-sha-256
  Database Ref:
    Name:   pb-custom
  Timeout:  5m
  Type:     Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-11-28T10:11:52Z
    Message:               Controller has started to Progress with Reconfigure of PgBouncerOpsRequest: demo/pbops-reconfigure-apply
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2024-11-28T10:11:55Z
    Message:               paused pgbouncer database
    Observed Generation:   1
    Reason:                Paused
    Status:                True
    Type:                  Paused
    Last Transition Time:  2024-11-28T10:11:55Z
    Message:               Successfully updated PgBouncer
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-11-28T10:11:55Z
    Message:               Successfully updated PgBouncer backend secret
    Observed Generation:   1
    Reason:                UpdateBackendSecret
    Status:                True
    Type:                  UpdateBackendSecret
    Last Transition Time:  2024-11-28T10:12:00Z
    Message:               get pod; ConditionStatus:True; PodName:pb-custom-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pb-custom-0
    Last Transition Time:  2024-11-28T10:12:00Z
    Message:               volume mount check; ConditionStatus:True; PodName:pb-custom-0
    Observed Generation:   1
    Status:                True
    Type:                  VolumeMountCheck--pb-custom-0
    Last Transition Time:  2024-11-28T10:12:05Z
    Message:               reload config; ConditionStatus:True; PodName:pb-custom-0
    Observed Generation:   1
    Status:                True
    Type:                  ReloadConfig--pb-custom-0
    Last Transition Time:  2024-11-28T10:12:05Z
    Message:               Reloading performed successfully in PgBouncer: demo/pb-custom for PgBouncerOpsRequest: pbops-reconfigure-apply
    Observed Generation:   1
    Reason:                ReloadPodsSucceeded
    Status:                True
    Type:                  ReloadPods
    Last Transition Time:  2024-11-28T10:12:05Z
    Message:               Successfully Reconfigured
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-11-28T10:12:05Z
    Message:               Controller has successfully completed  with Reconfigure of PgBouncerOpsRequest: demo/pbops-reconfigure-apply
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                         Age   From                         Message
  ----     ------                                                         ----  ----                         -------
  Normal   Starting                                                       54s   KubeDB Ops-manager Operator  Start processing for PgBouncerOpsRequest: demo/pbops-reconfigure-apply
  Normal   Starting                                                       54s   KubeDB Ops-manager Operator  Pausing PgBouncer databse: demo/pb-custom
  Normal   Successful                                                     54s   KubeDB Ops-manager Operator  Successfully paused PgBouncer database: demo/pb-custom for PgBouncerOpsRequest: pbops-reconfigure-apply
  Warning  get pod; ConditionStatus:True; PodName:pb-custom-0             46s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pb-custom-0
  Warning  volume mount check; ConditionStatus:True; PodName:pb-custom-0  46s   KubeDB Ops-manager Operator  volume mount check; ConditionStatus:True; PodName:pb-custom-0
  Warning  reload config; ConditionStatus:True; PodName:pb-custom-0       41s   KubeDB Ops-manager Operator  reload config; ConditionStatus:True; PodName:pb-custom-0
  Warning  reload config; ConditionStatus:True; PodName:pb-custom-0       41s   KubeDB Ops-manager Operator  reload config; ConditionStatus:True; PodName:pb-custom-0
  Normal   Successful                                                     41s   KubeDB Ops-manager Operator  Reloading performed successfully in PgBouncer: demo/pb-custom for PgBouncerOpsRequest: pbops-reconfigure-apply
  Normal   Starting                                                       41s   KubeDB Ops-manager Operator  Resuming PgBouncer database: demo/pb-custom
  Normal   Successful                                                     41s   KubeDB Ops-manager Operator  Successfully resumed PgBouncer database: demo/pb-custom
  Normal   Successful                                                     41s   KubeDB Ops-manager Operator  Controller has Successfully Reconfigured PgBouncer databases: demo/pb-custom
  Normal   Starting                                                       41s   KubeDB Ops-manager Operator  Resuming PgBouncer database: demo/pb-custom
  Normal   Successful                                                     41s   KubeDB Ops-manager Operator  Successfully resumed PgBouncer database: demo/pb-custom
  Normal   Successful                                                     41s   KubeDB Ops-manager Operator  Controller has Successfully Reconfigured PgBouncer databases: demo/pb-custom
```

Now let's exec into the pgbouncer pod and check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo pb-custom-0  -- /bin/sh 
pb-custom-0:/$ cat etc/config/pgbouncer.ini
[databases]
postgres= host=ha-postgres.demo.svc port=5432 dbname=postgres

[pgbouncer]
stats_period = 60
pidfile = /tmp/pgbouncer.pid
pool_mode = session
reserve_pool_timeout = 5
max_client_conn = 87
min_pool_size = 1
default_pool_size = 2
listen_addr = *
max_db_connections = 1
max_user_connections = 2
auth_type=scram-sha-256
ignore_startup_parameters = extra_float_digits
admin_users =  pgbouncer
auth_file =  /var/run/pgbouncer/secret/userlist
logfile = /tmp/pgbouncer.log
listen_port = 5432
reserve_pool_size = 5
pb-custom-0:/$ exit
exit
```

As we can see from the configuration of running pgbouncer, the value of `auth_type` has been changed from `md5` to `scram-sha-256`. So the reconfiguration of the pgbouncer using the `applyConfig` field is successful.

### Remove config

This will remove all the custom config previously provided. After this Ops-manager will merge the new given config with the default config and apply this.

- `spec.databaseRef.name` specifies that we are reconfiguring `pb-custom` pgbouncer.
- `spec.type` specifies that we are performing `Reconfigure` on our pgbouncer.
- `spec.configuration.removeCustomConfig` specifies for boolean values to remove previous custom configuration.



## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:
```bash
kubectl delete -n demo pb/pb-custom
kubectl delete pgbounceropsrequest -n demo pbops-reconfigure  pbops-reconfigure-apply
kubectl delete pg -n demo ha-postgres
kubectl delete ns demo
```