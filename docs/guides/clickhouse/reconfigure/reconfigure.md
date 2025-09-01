---
title: Reconfigure ClickHouse Cluster
menu:
  docs_{{ .version }}:
    identifier: ch-reconfigure-cluster
    name: Reconfigure Configurations
    parent: ch-reconfigure
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure ClickHouse Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to reconfigure a ClickHouse cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [ClickHouse](/docs/guides/clickhouse/concepts/clickhouse.md)
    - [ClickHouseOpsRequest](/docs/guides/clickhouse/concepts/clickhouseopsrequest.md)
    - [Reconfigure Overview](/docs/guides/clickhouse/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/clickhouse](/docs/examples/clickhouse) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

Now, we are going to deploy a  `ClickHouse` cluster using a supported version by `KubeDB` operator. Then we are going to apply `ClickHouseOpsRequest` to reconfigure its configuration.

### Prepare ClickHouse Cluster

Now, we are going to deploy a `ClickHouse` topology cluster with version `24.4.1`.

### Deploy ClickHouse

At first, we will create a secret with the `ch-config.yaml` file containing required configuration settings.

**ch-config.yaml:**

```properties
profiles:
      default:
        max_query_size: 200000
```

Here, `max_query_size` is set to `200000`, whereas the default value is `262144`

Let's create a k8s secret containing the above configuration where the file name will be the key and the file-content as the value:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ch-custom-config
  namespace: demo
type: Opaque
stringData:
  ch-config.yaml: |
    profiles:
      default:
        max_query_size: 200000
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/reconfigure/ch-config-secret.yaml
secret/ch-custom-config created
```


In this section, we are going to create a ClickHouse object specifying `spec.configSecret` field to apply this custom configuration. Below is the YAML of the `ClickHouse` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: clickhouse-prod
  namespace: demo
spec:
  version: 24.4.1
  configSecret:
    name: ch-custom-config
  clusterTopology:
    clickHouseKeeper:
      externallyManaged: false
      spec:
        replicas: 3
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
    cluster:
      - name: appscode-cluster
        shards: 2
        replicas: 2
        podTemplate:
          spec:
            containers:
              - name: clickhouse
                resources:
                  limits:
                    memory: 4Gi
                  requests:
                    cpu: 500m
                    memory: 2Gi
            initContainers:
              - name: clickhouse-init
                resources:
                  limits:
                    memory: 1Gi
                  requests:
                    cpu: 500m
                    memory: 1Gi
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `ClickHouse` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/reconfigure/clickhouse-cluster.yaml
clickhouse.kubedb.com/clickhouse-prod created
```

Now, wait until `clickhouse-prod` has status `Ready`. i.e,

```bash
➤ kubectl get ch -n demo clickhouse-prod -w
NAME              TYPE                  VERSION   STATUS         AGE
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Provisioning   101s
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Provisioning   109s
.
.
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Ready          2m17s
```

Now, we will check if the clickhouse has started with the custom configuration we have provided.

Exec into the ClickHouse pod and execute the following commands to see the configurations:
```bash
➤ kubectl exec -it -n demo clickhouse-prod-appscode-cluster-shard-0-0 -- bash
Defaulted container "clickhouse" out of: clickhouse, clickhouse-init (init)
clickhouse@clickhouse-prod-appscode-cluster-shard-0-0:/$ cat /etc/clickhouse-server/conf.d/ch-config.yaml
profiles:
  default:
    max_query_size: 200000
```
Here, we can see that our given configuration is applied to the ClickHouse cluster . `profiles.default.max_query_size` is set to `200000` from the default value `262144`.

### Reconfigure using new config secret

Now we will reconfigure this cluster to set `max_query_size` to `150000`.

Now, update our `ch-config.yaml` file with the new configuration.

```properties
profiles:
      default:
        max_query_size: 150000
```

Then, we will create a new secret with this configuration file.

At first, create `clickhouse.yaml` file containing required configuration settings.

```bash
$ cat clickhouse.yaml
read_request_timeout: 6500ms
```

Then, we will create a new secret with this configuration file.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: new-ch-custom-config
  namespace: demo
type: Opaque
stringData:
  ch-config.yaml: |
    profiles:
      default:
        max_query_size: 150000

```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/reconfigure/new-ch-config-secret.yaml
secret/new-ch-custom-config created
```

#### Create ClickHouseOpsRequest

Now, we will use this secret to replace the previous secret using a `ClickHouseOpsRequest` CR. The `ClickHouseOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-cluster-reconfigure-with-secret
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: clickhouse-prod
  configuration:
    configSecret:
      name: new-ch-custom-config
  timeout: 10m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `clickhouse-prod` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configSecret.name` specifies the name of the new secret.

Let's create the `ClickHouseOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/reconfigure/ch-reconfigure-ops-with-secret.yaml
clickhouseopsrequest.ops.kubedb.com/chops-cluster-reconfigure-with-secret created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Ops-manager operator will update the `configSecret` of `ClickHouse` object.

Let's wait for `ClickHouseOpsRequest` to be `Successful`.  Run the following command to watch `ClickHouseOpsRequest` CR,

```bash
➤ kubectl get clickhouseopsrequests -n demo 
NAME                                    TYPE          STATUS       AGE
chops-cluster-reconfigure-with-secret   Reconfigure   Successful   48m
```

We can see from the above output that the `ClickHouseOpsRequest` has succeeded. If we describe the `ClickHouseOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
➤ kubectl describe chops -n demo chops-cluster-reconfigure-with-secret 
Name:         chops-cluster-reconfigure-with-secret
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ClickHouseOpsRequest
Metadata:
  Creation Timestamp:  2025-08-22T10:39:47Z
  Generation:          1
  Resource Version:    458839
  UID:                 54b3c1b1-976b-4129-9b08-dcb11426b990
Spec:
  Apply:  IfReady
  Configuration:
    Config Secret:
      Name:  new-ch-custom-config
  Database Ref:
    Name:   clickhouse-prod
  Timeout:  10m
  Type:     Reconfigure
Status:
  Conditions:
    Last Transition Time:  2025-08-22T10:39:47Z
    Message:               ClickHouse ops-request has started to reconfigure ClickHouse nodes
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2025-08-22T10:39:55Z
    Message:               successfully reconciled the ClickHouse with new configure
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-22T10:39:55Z
    Message:               reconcile; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  Reconcile
    Last Transition Time:  2025-08-22T10:42:20Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-08-22T10:40:00Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-08-22T10:40:00Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-08-22T10:40:05Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-22T10:40:40Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-08-22T10:40:40Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-08-22T10:41:20Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-08-22T10:41:20Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-08-22T10:42:00Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-08-22T10:42:00Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-08-22T10:42:20Z
    Message:               Successfully completed reconfigure ClickHouse
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                               Age   From                         Message
  ----     ------                                                                               ----  ----                         -------
  Normal   Starting                                                                             49m   KubeDB Ops-manager Operator  Start processing for ClickHouseOpsRequest: demo/chops-cluster-reconfigure-with-secret
  Normal   Starting                                                                             49m   KubeDB Ops-manager Operator  Pausing ClickHouse databse: demo/clickhouse-prod
  Normal   Successful                                                                           49m   KubeDB Ops-manager Operator  Successfully paused ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-cluster-reconfigure-with-secret
  Warning  reconcile; ConditionStatus:True                                                      49m   KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                                                      49m   KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                                                      49m   KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Normal   UpdatePetSets                                                                        49m   KubeDB Ops-manager Operator  successfully reconciled the ClickHouse with new configure
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0    48m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0  48m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  running pod; ConditionStatus:False                                                   48m   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1    48m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1  48m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0    47m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0  47m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1    46m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1  46m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Normal   RestartNodes                                                                         46m   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                             46m   KubeDB Ops-manager Operator  Resuming ClickHouse database: demo/clickhouse-prod
  Normal   Successful                                                                           46m   KubeDB Ops-manager Operator  Successfully resumed ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-cluster-reconfigure-with-secret
```

Now let's exec one of the instance to check the new configuration we have provided.

```bash
➤ kubectl exec -it -n demo clickhouse-prod-appscode-cluster-shard-0-0 -- bash
Defaulted container "clickhouse" out of: clickhouse, clickhouse-init (init)
clickhouse@clickhouse-prod-appscode-cluster-shard-0-0:/$ cat /etc/clickhouse-server/conf.d/ch-config.yaml 
profiles:
  default:
    max_query_size: "150000"
```

As we can see from the configuration of ready clickhouse, the value of `max_query_size` has been changed from `200000` to `150000`. So the reconfiguration of the cluster is successful.


### Reconfigure using apply config

Now we will reconfigure this cluster again to set `180000` to ``. This time we won't use a new secret. We will use the `applyConfig` field of the `ClickHouseOpsRequest`. This will merge the new config in the existing secret.

#### Create ClickHouseOpsRequest

Now, we will use the new configuration in the `applyConfig` field in the `ClickHouseOpsRequest` CR. The `ClickHouseOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-cluster-reconfigure-with-config
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: clickhouse-prod
  configuration:
    applyConfig:
      ch-config.yaml: |
        profiles:
          default:
            max_query_size: 180000
  timeout: 10m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `clickhouse-prod` cluster.
- `spec.type` specifies that we are performing `Reconfigure` on clickhouse.
- `spec.configuration.applyConfig` specifies the new configuration that will be merged in the existing secret.

Let's create the `ClickHouseOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/reconfigure/ch-reconfigure-ops-with-apply-config.yaml
clickhouseopsrequest.ops.kubedb.com/chops-cluster-reconfiugre-with-config created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Ops-manager operator will merge this new config with the existing configuration.

Let's wait for `ClickHouseOpsRequest` to be `Successful`.  Run the following command to watch `ClickHouseOpsRequest` CR,

```bash
➤ kubectl get chops -n demo chops-cluster-reconfiugre-with-config 
NAME                                    TYPE          STATUS       AGE
chops-cluster-reconfiugre-with-config   Reconfigure   Successful   12m
```

We can see from the above output that the `ClickHouseOpsRequest` has succeeded. If we describe the `ClickHouseOpsRequest` we will get an overview of the steps that were followed to reconfigure the cluster.



```bash
➤ kubectl describe chops -n demo chops-cluster-reconfiugre-with-config 
Name:         chops-cluster-reconfiugre-with-config
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ClickHouseOpsRequest
Metadata:
  Creation Timestamp:  2025-08-22T11:34:23Z
  Generation:          1
  Resource Version:    466167
  UID:                 2270c10b-490c-43db-9cc9-92171b9513bb
Spec:
  Apply:  IfReady
  Configuration:
    Apply Config:
      config.yaml:  profiles:
  default:
    max_query_size: 180000

  Database Ref:
    Name:   clickhouse-prod
  Timeout:  10m
  Type:     Reconfigure
Status:
  Conditions:
    Last Transition Time:  2025-08-22T11:34:23Z
    Message:               ClickHouse ops-request has started to reconfigure ClickHouse nodes
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2025-08-22T11:34:26Z
    Message:               Successfully prepared user provided custom config secret
    Observed Generation:   1
    Reason:                PrepareCustomConfig
    Status:                True
    Type:                  PrepareCustomConfig
    Last Transition Time:  2025-08-22T11:34:31Z
    Message:               successfully reconciled the ClickHouse with new configure
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-22T11:34:31Z
    Message:               reconcile; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  Reconcile
    Last Transition Time:  2025-08-22T11:36:56Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-08-22T11:34:36Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-08-22T11:34:36Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-08-22T11:34:41Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-22T11:34:56Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-08-22T11:34:56Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-08-22T11:35:36Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-08-22T11:35:36Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-08-22T11:36:21Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-08-22T11:36:21Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-08-22T11:36:56Z
    Message:               Successfully completed reconfigure ClickHouse
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                               Age   From                         Message
  ----     ------                                                                               ----  ----                         -------
  Normal   Starting                                                                             13m   KubeDB Ops-manager Operator  Start processing for ClickHouseOpsRequest: demo/chops-cluster-reconfiugre-with-config
  Normal   Starting                                                                             13m   KubeDB Ops-manager Operator  Pausing ClickHouse databse: demo/clickhouse-prod
  Normal   Successful                                                                           13m   KubeDB Ops-manager Operator  Successfully paused ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-cluster-reconfiugre-with-config
  Warning  reconcile; ConditionStatus:True                                                      13m   KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                                                      13m   KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                                                      13m   KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Normal   UpdatePetSets                                                                        13m   KubeDB Ops-manager Operator  successfully reconciled the ClickHouse with new configure
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0    13m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0  13m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  running pod; ConditionStatus:False                                                   13m   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1    12m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1  12m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0    12m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0  12m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1    11m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1  11m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Normal   RestartNodes                                                                         10m   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                             10m   KubeDB Ops-manager Operator  Resuming ClickHouse database: demo/clickhouse-prod
  Normal   Successful                                                                           10m   KubeDB Ops-manager Operator  Successfully resumed ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-cluster-reconfiugre-with-config
```

Now let's exec into one of the instance to check the new configuration we have provided.

```bash
➤ kubectl exec -it -n demo clickhouse-prod-appscode-cluster-shard-0-0 -- bash
Defaulted container "clickhouse" out of: clickhouse, clickhouse-init (init)
clickhouse@clickhouse-prod-appscode-cluster-shard-0-0:/$ cat /etc/clickhouse-server/conf.d/ch-config.yaml 
profiles:
    default:
        max_query_size: 180000
```

As we can see from the configuration of ready clickhouse, the value of `max_query_size` has been changed from `150000` to `180000`. So the reconfiguration of the database using the `applyConfig` field is successful.


## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete ch -n demo clickhouse-prod
kubectl delete clickhouseopsrequest -n demo chops-cluster-reconfigure-with-config chops-cluster-reconfigure-with-secret
kubectl delete secret -n demo ch-custom-config new-ch-custom-config
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [ClickHouse object](/docs/guides/clickhouse/concepts/clickhouse.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
