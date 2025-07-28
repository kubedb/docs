---
title: Reconfigure Cassandra Topology
menu:
  docs_{{ .version }}:
    identifier: cas-reconfigure-topology
    name: Topology
    parent: cas-reconfigure
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Cassandra Topology Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to reconfigure a Cassandra Topology cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Cassandra](/docs/guides/cassandra/concepts/cassandra.md)
    - [Topology](/docs/guides/cassandra/clustering/topology-cluster/index.md)
    - [CassandraOpsRequest](/docs/guides/cassandra/concepts/cassandraopsrequest.md)
    - [Reconfigure Overview](/docs/guides/cassandra/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/cassandra](/docs/examples/cassandra) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

Now, we are going to deploy a  `Cassandra` Topology cluster using a supported version by `KubeDB` operator. Then we are going to apply `CassandraOpsRequest` to reconfigure its configuration.

### Prepare Cassandra Topology Cluster

Now, we are going to deploy a `Cassandra` topology cluster with version `5.0.3`.

### Deploy Cassandra

At first, we will create a secret with the `cassandra.yaml` file containing required configuration settings.

**cassandra.yaml:**

```properties
read_request_timeout: 6000ms
write_request_timeout: 2500ms
```

Here, `read_request_timeout` is set to `6000ms`, whereas the default value is `5000ms` and `write_request_timeout` is set to `2500ms`, whereas the default value is 2000ms.

Let's create a k8s secret containing the above configuration where the file name will be the key and the file-content as the value:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cas-topology-custom-config
  namespace: demo
stringData:
  cassandra.yaml: |-
    read_request_timeout: 6000ms
    write_request_timeout: 2500ms
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/reconfigure/cassandra-topology-custom-config-secret.yaml
secret/cas-topology-custom-config created
```


In this section, we are going to create a Cassandra object specifying `spec.configSecret` field to apply this custom configuration. Below is the YAML of the `Cassandra` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:
  name: cassandra-prod
  namespace: demo
spec:
  version: 5.0.3
  configSecret:
    name: cas-topology-custom-config
  topology:
    rack:
      - name: r0
        replicas: 2
        podTemplate:
          spec:
            containers:
              - name: cassandra
                resources:
                  limits:
                    memory: 2Gi
                    cpu: 2
                  requests:
                    memory: 1Gi
                    cpu: 1
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
        storageType: Durable
  deletionPolicy: WipeOut
```

Let's create the `Cassandra` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/reconfigure/cassandra-topology.yaml
cassandra.kubedb.com/cassandra-prod created
```

Now, wait until `cassandra-prod` has status `Ready`. i.e,

```bash
$ kubectl get cas -n demo -w
NAME             TYPE                  VERSION   STATUS         AGE
cassandra-prod   kubedb.com/v1alpha2   5.0.3     Provisioning   48s
cassandra-prod   kubedb.com/v1alpha2   5.0.3     Provisioning   81s
.
.
cassandra-prod   kubedb.com/v1alpha2   5.0.3     Ready          105s
```

Now, we will check if the cassandra has started with the custom configuration we have provided.

Exec into the Cassandra pod and execute the following commands to see the configurations:
```bash
$ kubectl exec -it -n demo cassandra-prod-rack-r0-0  -- bash
Defaulted container "cassandra" out of: cassandra, cassandra-init (init), medusa-init (init)
[cassandra@cassandra-prod-rack-r0-0 /]$ cat /etc/cassandra/cassandra.yaml | grep request_timeout
read_request_timeout: 6000ms
range_request_timeout: 10000ms
write_request_timeout: 2500ms
counter_write_request_timeout: 5000ms
truncate_request_timeout: 60000ms
request_timeout: 10000ms
```
Here, we can see that our given configuration is applied to the Cassandra cluster . `read_request_timeout` is set to `6000ms` from the default value `5000ms`.

### Reconfigure using new config secret

Now we will reconfigure this cluster to set `read_request_timeout` to `6500ms`.

Now, update our `cassandra.yaml` file with the new configuration.

**cassandra.yaml:**

```properties
read_request_timeout=6500ms
```

Then, we will create a new secret with this configuration file.

At first, create `cassandra.yaml` file containing required configuration settings.

```bash
$ cat cassandra.yaml
read_request_timeout: 6500ms
```

Now, create the secret with this configuration file.

```bash
$ kubectl create secret generic -n demo  new-cas-topology-custom-config --from-file=./cassandra.yaml
secret/new-cas-topology-custom-config created
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/reconfigure/new-cassandra-topology-custom-config-secret.yaml
secret/new-cas-topology-custom-config created
```

#### Create CassandraOpsRequest

Now, we will use this secret to replace the previous secret using a `CassandraOpsRequest` CR. The `CassandraOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name:  casops-reconfigure-topology
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: cassandra-prod
  configuration:
    configSecret:
      name: new-cas-topology-custom-config
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `cassandra-prod` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configSecret.name` specifies the name of the new secret.

Let's create the `CassandraOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/reconfigure/cassandra-reconfigure-update-topology-ops.yaml
cassandraopsrequest.ops.kubedb.com/casops-reconfigure-topology created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Ops-manager operator will update the `configSecret` of `Cassandra` object.

Let's wait for `CassandraOpsRequest` to be `Successful`.  Run the following command to watch `CassandraOpsRequest` CR,

```bash
$ kubectl get cassandraopsrequests -n demo 
NAME                          TYPE          STATUS       AGE
casops-reconfigure-topology   Reconfigure   Successful   2m53s
```

We can see from the above output that the `CassandraOpsRequest` has succeeded. If we describe the `CassandraOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$  kubectl describe cassandraopsrequest -n demo casops-reconfigure-topology
Name:         casops-reconfigure-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         CassandraOpsRequest
Metadata:
  Creation Timestamp:  2025-07-22T09:14:45Z
  Generation:          1
  Resource Version:    141080
  UID:                 35eba2d1-6a7f-4288-8529-11c086c85cb9
Spec:
  Apply:  IfReady
  Configuration:
    Config Secret:
      Name:  new-cas-topology-custom-config
  Database Ref:
    Name:   cassandra-prod
  Timeout:  5m
  Type:     Reconfigure
Status:
  Conditions:
    Last Transition Time:  2025-07-22T09:14:45Z
    Message:               Cassandra ops-request has started to reconfigure Cassandra nodes
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2025-07-22T09:14:53Z
    Message:               successfully reconciled the Cassandra with new configure
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-22T09:17:38Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-07-22T09:14:58Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-22T09:14:58Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-22T09:15:03Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-07-22T09:15:38Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-22T09:15:38Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-22T09:17:38Z
    Message:               Successfully completed reconfigure Cassandra
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                             Age    From                         Message
  ----     ------                                                             ----   ----                         -------
  Normal   Starting                                                           3m18s  KubeDB Ops-manager Operator  Start processing for CassandraOpsRequest: demo/casops-reconfigure-topology
  Normal   Starting                                                           3m18s  KubeDB Ops-manager Operator  Pausing Cassandra databse: demo/cassandra-prod
  Normal   Successful                                                         3m18s  KubeDB Ops-manager Operator  Successfully paused Cassandra database: demo/cassandra-prod for CassandraOpsRequest: casops-reconfigure-topology
  Normal   UpdatePetSets                                                      3m10s  KubeDB Ops-manager Operator  successfully reconciled the Cassandra with new configure
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0    3m5s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0  3m5s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  running pod; ConditionStatus:False                                 3m     KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1    2m25s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1  2m25s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0    105s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0  105s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1    65s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1  65s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Normal   RestartNodes                                                       25s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                           25s    KubeDB Ops-manager Operator  Resuming Cassandra database: demo/cassandra-prod
  Normal   Successful                                                         25s    KubeDB Ops-manager Operator  Successfully resumed Cassandra database: demo/cassandra-prod for CassandraOpsRequest: casops-reconfigure-topology
```

Now let's exec one of the instance to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo cassandra-prod-rack-r0-0  -- bash
Defaulted container "cassandra" out of: cassandra, cassandra-init (init), medusa-init (init)
[cassandra@cassandra-prod-rack-r0-0 /]$ cat /etc/cassandra/cassandra.yaml | grep request_timeout
read_request_timeout: 6500ms
range_request_timeout: 10000ms
write_request_timeout: 2500ms
counter_write_request_timeout: 5000ms
truncate_request_timeout: 60000ms
request_timeout: 10000ms
```

As we can see from the configuration of ready cassandra, the value of `read_request_timeout` has been changed from `6000ms` to `6500ms`. So the reconfiguration of the cluster is successful.


### Reconfigure using apply config

Now we will reconfigure this cluster again to set `read_request_timeout` to `5500ms`. This time we won't use a new secret. We will use the `applyConfig` field of the `CassandraOpsRequest`. This will merge the new config in the existing secret.

#### Create CassandraOpsRequest

Now, we will use the new configuration in the `applyConfig` field in the `CassandraOpsRequest` CR. The `CassandraOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name:  casops-reconfigure-apply-topology
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: cassandra-prod
  configuration:
    applyConfig:
      cassandra.yaml: |-
        read_request_timeout=5500ms
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `cassandra-prod` cluster.
- `spec.type` specifies that we are performing `Reconfigure` on cassandra.
- `spec.configuration.applyConfig` specifies the new configuration that will be merged in the existing secret.

Let's create the `CassandraOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/reconfigure/cassandra-reconfigure-apply-topology.yaml
cassandraopsrequest.ops.kubedb.com/casops-reconfigure-apply-topology created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Ops-manager operator will merge this new config with the existing configuration.

Let's wait for `CassandraOpsRequest` to be `Successful`.  Run the following command to watch `CassandraOpsRequest` CR,

```bash
$ kubectl get cassandraopsrequests -n demo casops-reconfigure-apply-topology 
NAME                               TYPE          STATUS       AGE
casops-reconfigure-apply-topology   Reconfigure   Successful   55s
```

We can see from the above output that the `CassandraOpsRequest` has succeeded. If we describe the `CassandraOpsRequest` we will get an overview of the steps that were followed to reconfigure the cluster.



```bash
$ kubectl describe cassandraopsrequest -n demo casops-reconfigure-apply-topology 
Name:         casops-reconfigure-apply-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         CassandraOpsRequest
Metadata:
  Creation Timestamp:  2025-07-22T09:38:59Z
  Generation:          1
  Resource Version:    144161
  UID:                 9b7144c6-4b6a-4095-b87e-5a0630e29dae
Spec:
  Apply:  IfReady
  Configuration:
    Apply Config:
      cassandra.yaml:  read_request_timeout: 5500ms
  Database Ref:
    Name:   cassandra-prod
  Timeout:  5m
  Type:     Reconfigure
Status:
  Conditions:
    Last Transition Time:  2025-07-22T09:40:24Z
    Message:               Cassandra ops-request has started to reconfigure Cassandra nodes
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2025-07-22T09:40:27Z
    Message:               Successfully prepared user provided custom config secret
    Observed Generation:   1
    Reason:                PrepareCustomConfig
    Status:                True
    Type:                  PrepareCustomConfig
    Last Transition Time:  2025-07-22T09:40:32Z
    Message:               successfully reconciled the Cassandra with new configure
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-22T09:43:17Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-07-22T09:40:37Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-22T09:40:37Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-22T09:40:42Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-07-22T09:41:17Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-22T09:41:17Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-22T09:43:18Z
    Message:               Successfully completed reconfigure Cassandra
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                             Age    From                         Message
  ----     ------                                                             ----   ----                         -------
  Normal   Starting                                                           3m31s  KubeDB Ops-manager Operator  Start processing for CassandraOpsRequest: demo/casops-reconfigure-apply-topology
  Normal   Starting                                                           3m31s  KubeDB Ops-manager Operator  Pausing Cassandra databse: demo/cassandra-prod
  Normal   Successful                                                         3m31s  KubeDB Ops-manager Operator  Successfully paused Cassandra database: demo/cassandra-prod for CassandraOpsRequest: casops-reconfigure-apply-topology
  Normal   UpdatePetSets                                                      3m23s  KubeDB Ops-manager Operator  successfully reconciled the Cassandra with new configure
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0    3m18s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0  3m18s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  running pod; ConditionStatus:False                                 3m13s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1    2m38s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1  2m38s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0    118s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0  118s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1    78s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1  78s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Normal   RestartNodes                                                       37s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                           37s    KubeDB Ops-manager Operator  Resuming Cassandra database: demo/cassandra-prod
  Normal   Successful                                                         37s    KubeDB Ops-manager Operator  Successfully resumed Cassandra database: demo/cassandra-prod for CassandraOpsRequest: casops-reconfigure-apply-topology
```

Now let's exec into one of the instance to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo cassandra-prod-rack-r0-0  -- bash
Defaulted container "cassandra" out of: cassandra, cassandra-init (init), medusa-init (init)
[cassandra@cassandra-prod-rack-r0-0 /]$ cat /etc/cassandra/cassandra.yaml | grep request_timeout
read_request_timeout: 5500ms
range_request_timeout: 10000ms
write_request_timeout: 2500ms
counter_write_request_timeout: 5000ms
truncate_request_timeout: 60000ms
request_timeout: 10000ms```

As we can see from the configuration of ready cassandra, the value of `read_request_timeout` has been changed from `125` to `150`. So the reconfiguration of the database using the `applyConfig` field is successful.


## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete cas -n demo cassandra-dev
kubectl delete cassandraopsrequest -n demo casops-reconfigure-apply-topology casops-reconfigure-topology
kubectl delete secret -n demo cas-topology-custom-config new-cas-topology-custom-config
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Cassandra object](/docs/guides/cassandra/concepts/cassandra.md).
- Different Cassandra topology clustering modes [here](/docs/guides/cassandra/clustering/_index.md).
- Monitor your Cassandra database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/cassandra/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Cassandra database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/cassandra/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
