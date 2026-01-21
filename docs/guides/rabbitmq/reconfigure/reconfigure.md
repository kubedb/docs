---
title: Reconfigure RabbitMQ Cluster
menu:
  docs_{{ .version }}:
    identifier: rm-reconfigure-cluster
    name: Reconfigure Configurations
    parent: rm-reconfigure
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure RabbitMQ Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to reconfigure a RabbitMQ cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [RabbitMQ](/docs/guides/rabbitmq/concepts/rabbitmq.md)
  - [RabbitMQOpsRequest](/docs/guides/rabbitmq/concepts/opsrequest.md)
  - [Reconfigure Overview](/docs/guides/rabbitmq/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [examples](/docs/examples/rabbitmq) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

Now, we are going to deploy a  `RabbitMQ` cluster using a supported version by `KubeDB` operator. Then we are going to apply `RabbitMQOpsRequest` to reconfigure its configuration.

### Prepare RabbitMQ Standalone Database

Now, we are going to deploy a `RabbitMQ` cluster with version `3.13.2`.

### Deploy RabbitMQ standalone 

At first, we will create `rabbitmq.conf` file containing required configuration settings.

```ini
$ cat rabbitmq.conf
default_vhost = /customvhost
```
Here, `default_vhost` is set to `/customvhost` instead of the default vhost `/`.

Now, we will create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo rabbit-custom-config --from-file=./rabbitmq.conf
secret/rabbit-custom-config created
```

In this section, we are going to create a RabbitMQ object specifying `spec.configuration` field to apply this custom configuration. Below is the YAML of the `RabbitMQ` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rm-cluster
  namespace: demo
spec:
  version: "3.13.2"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  configuration:
    secretName: rabbit-custom-config
```

Let's create the `RabbitMQ` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/cluster/rabbit-custom-config.yaml
rabbitmq.kubedb.com/rm-cluster created
```

Now, wait until `rm-cluster` has status `Ready`. i.e,

```bash
$ kubectl get rm -n demo
NAME            TYPE                  VERSION   STATUS   AGE
rm-cluster      kubedb.com/v1alpha2   3.13.2    Ready    79m
```

Now, we will check if the database has started with the custom configuration we have provided.

First we need to get the username and password to connect to a RabbitMQ instance,
```bash
$ kubectl get secrets -n demo rm-cluster-admin-cred -o jsonpath='{.data.\username}' | base64 -d
admin

$ kubectl get secrets -n demo rm-cluster-admin-cred  -o jsonpath='{.data.\password}' | base64 -d
m6lXjZugrC4VEpB8
```

Now let's check the configuration we have provided by using rabbitmq's inbuilt cli.

```bash
$ kubectl exec -it -n demo rm-cluster-0 -- bash
Defaulted container "rabbitmq" out of: rabbitmq, rabbitmq-init (init)
rm-cluster-0:/$ rabbitmqctl list_vhosts
Listing vhosts ...
name
/customvhost
```

Provided custom vhost is there and is defaulted.

### Reconfigure using new secret

Now we will update this default vhost to `/newvhost` using Reconfigure Ops-Request.

Now, Let's edit the `rabbitmq.conf` file containing required configuration settings.

```bash
$ echo "default_vhost = /newvhost" > rabbitmq.conf
```

Then, we will create a new secret with this configuration file.

```bash
$ kubectl create secret generic -n demo new-custom-config --from-file=./rabbitmq.conf
secret/new-custom-config created
```

#### Create RabbitMQOpsRequest

Now, we will use this secret to replace the previous secret using a `RabbitMQOpsRequest` CR. The `RabbitMQOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RabbitMQOpsRequest
metadata:
  name: reconfigure-rm-cluster
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: rm-cluster
  configuration:
    configSecret:
      name: new-custom-config
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `mops-reconfigure-standalone` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.configSecret.name` specifies the name of the new secret.
- Have a look [here](/docs/guides/rabbitmq/concepts/opsrequest.md#specconfiguration) on the respective sections to understand the `readinessCriteria`, `timeout` & `apply` fields.

Let's create the `RabbitMQOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/opsrequests/rabbit-reconfigure-with-secret.yaml
rabbitmqopsrequest.ops.kubedb.com/reconfigure-rm-cluster created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Ops-manager operator will update the `configSecret` of `RabbitMQ` object.

Let's wait for `RabbitMQOpsRequest` to be `Successful`.  Run the following command to watch `RabbitMQOpsRequest` CR,

```bash
$ watch kubectl get rabbitmqopsrequest -n demo
Every 2.0s: kubectl get rabbitmqopsrequest -n demo
NAME                          TYPE          STATUS       AGE
reconfigure-rm-cluster        Reconfigure   Successful   3m
```

We can see from the above output that the `RabbitMQOpsRequest` has succeeded. If we describe the `RabbitMQOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe rabbitmqopsrequest -n demo reconfigure-rm-cluster
Name:         reconfigure-rm-cluster
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RabbitMQOpsRequest
Metadata:
  Creation Timestamp:  2024-09-10T11:09:16Z
  Generation:          1
  Resource Version:    70651
  UID:                 5c99031f-6604-48ac-b700-96f896c5d0b3
Spec:
  Apply:  IfReady
  Configuration:
    Config Secret:
      Name:  new-custom-config
  Database Ref:
    Name:   rm-cluster
  Timeout:  5m
  Type:     Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-09-10T11:09:16Z
    Message:               RabbitMQ ops-request has started to reconfigure RabbitMQ nodes
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-09-10T11:09:24Z
    Message:               successfully reconciled the RabbitMQ with new configure
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-09-10T11:09:29Z
    Message:               get pod; ConditionStatus:True; PodName:rm-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--rm-cluster-0
    Last Transition Time:  2024-09-10T11:09:29Z
    Message:               evict pod; ConditionStatus:True; PodName:rm-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--rm-cluster-0
    Last Transition Time:  2024-09-10T11:09:34Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2024-09-10T11:09:49Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-09-10T11:09:50Z
    Message:               Successfully completed reconfigure RabbitMQ
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                 Age    From                         Message
  ----     ------                                                 ----   ----                         -------
  Normal   Starting                                               6m13s  KubeDB Ops-manager Operator  Start processing for RabbitMQOpsRequest: demo/reconfigure-rm-cluster
  Normal   Starting                                               6m13s  KubeDB Ops-manager Operator  Pausing RabbitMQ databse: demo/rm-cluster
  Normal   Successful                                             6m13s  KubeDB Ops-manager Operator  Successfully paused RabbitMQ database: demo/rm-cluster for RabbitMQOpsRequest: reconfigure
  Normal   UpdatePetSets                                          6m5s   KubeDB Ops-manager Operator  successfully reconciled the RabbitMQ with new configure
  Warning  get pod; ConditionStatus:True; PodName:rm-cluster-0    6m     KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:rm-cluster-0
  Warning  evict pod; ConditionStatus:True; PodName:rm-cluster-0  6m     KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:rm-cluster-0
  Warning  running pod; ConditionStatus:False                     5m55s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Normal   RestartNodes                                           5m40s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                               5m40s  KubeDB Ops-manager Operator  Resuming RabbitMQ database: demo/rm-cluster
  Normal   Successful                                             5m39s  KubeDB Ops-manager Operator  Successfully resumed RabbitMQ database: demo/rm-cluster for RabbitMQOpsRequest: reconfigure-rm-cluster
```

Now let's check the configuration we have provided after reconfiguration.

```bash
$ kubectl exec -it -n demo rm-cluster-0 -- bash
Defaulted container "rabbitmq" out of: rabbitmq, rabbitmq-init (init)
rm-cluster-0:/$ rabbitmqctl list_vhosts
Listing vhosts ...
name
/newvhost
/customvhost
```
As we can see from the configuration of running RabbitMQ, `/newvhost` is in the list of vhosts.

### Reconfigure using apply config

Let's say you are in a rush or, don't want to create a secret for updating configuration. You can directly do that using the following manifest.

#### Create RabbitMQOpsRequest

Now, we will use the new configuration in the `configuration.applyConfig` field in the `RabbitMQOpsRequest` CR. The `RabbitMQOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RabbitMQOpsRequest
metadata:
  name: reconfigure-apply
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: rm-cluster
  configuration:
    applyConfig:
      rabbitmq.conf: |
        default_vhost = /newvhost
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `rm-cluster` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.applyConfig` specifies the new configuration that will be merged in the existing secret.

Let's create the `RabbitMQOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/opsrequests/rabbitmq-reconfigure-apply.yaml
rabbitmqopsrequest.ops.kubedb.com/reconfigure-apply created
```

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete rm -n demo rm-cluster
kubectl delete rabbitmqopsrequest -n demo reconfigure-apply reconfigure-rm-cluster
```