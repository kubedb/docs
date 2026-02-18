---
title: Reconfigure Ignite Cluster
menu:
  docs_{{ .version }}:
    identifier: ig-reconfigure-cluster
    name: Reconfigure Configurations
    parent: ig-reconfigure
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Ignite Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to reconfigure a Ignite cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Ignite](/docs/guides/ignite/concepts/ignite.md)
  - [IgniteOpsRequest](/docs/guides/ignite/concepts/opsrequest.md)
  - [Reconfigure Overview](/docs/guides/ignite/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [examples](/docs/examples/ignite) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

Now, we are going to deploy a  `Ignite` cluster using a supported version by `KubeDB` operator. Then we are going to apply `IgniteOpsRequest` to reconfigure its configuration.

### Prepare Ignite Database

Now, we are going to deploy a `Ignite` cluster with version `2.17.0`.

### Deploy Ignite

At first, we will create `ignite.conf` file containing required configuration settings.

Now, we will create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo ig-custom-config --from-file=./node-configuration.xml
secret/ig-custom-config created
```

In this section, we are going to create a Ignite object specifying `spec.configuration` field to apply this custom configuration. Below is the YAML of the `Ignite` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Ignite
metadata:
  name: ig-cluster
  namespace: demo
spec:
  version: "2.17.0"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  configuration:
    secretName: ig-custom-config
```

Let's create the `Ignite` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ignite/cluster/ig-custom-config.yaml
ignite.kubedb.com/ig-cluster created
```

Now, wait until `ig-cluster` has status `Ready`. i.e,

```bash
$ kubectl get ig -n demo
NAME            TYPE                  VERSION   STATUS   AGE
ig-cluster      kubedb.com/v1alpha2   2.17.0    Ready    79m
```

Now, we will check if the database has started with the custom configuration we have provided.

First we need to get the username and password to connect to a Ignite instance,
```bash
$ kubectl get secrets -n demo ig-cluster-admin-cred -o jsonpath='{.data.\username}' | base64 -d
admin

$ kubectl get secrets -n demo ig-cluster-admin-cred  -o jsonpath='{.data.\password}' | base64 -d
m6lXjZugrC4VEpB8
```

### Reconfigure using new secret

Now, we will create a new secret with this configuration file.

```bash
$ kubectl create secret generic -n demo new-custom-config --from-file=./ignite.conf
secret/new-custom-config created
```

#### Create IgniteOpsRequest

Now, we will use this secret to replace the previous secret using a `IgniteOpsRequest` CR. The `IgniteOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: reconfigure-ig-cluster
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: ig-cluster
  configuration:
    configSecret:
      name: new-custom-config
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `igps-reconfigure` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.configSecret.name` specifies the name of the new secret.
- Have a look [here](/docs/guides/ignite/concepts/opsrequest.md#specconfiguration) on the respective sections to understand the `readinessCriteria`, `timeout` & `apply` fields.

Let's create the `IgniteOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ignite/opsrequests/ig-reconfigure-with-secret.yaml
igniteopsrequest.ops.kubedb.com/reconfigure-ig-cluster created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Ops-manager operator will update the `configSecret` of `Ignite` object.

Let's wait for `IgniteOpsRequest` to be `Successful`.  Run the following command to watch `IgniteOpsRequest` CR,

```bash
$ watch kubectl get igniteopsrequest -n demo
Every 2.0s: kubectl get igniteopsrequest -n demo
NAME                          TYPE          STATUS       AGE
reconfigure-ig-cluster        Reconfigure   Successful   3m
```

We can see from the above output that the `IgniteOpsRequest` has succeeded. If we describe the `IgniteOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe igniteopsrequest -n demo reconfigure-ig-cluster
Name:         reconfigure-ig-cluster
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         IgniteOpsRequest
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
    Name:   ig-cluster
  Timeout:  5m
  Type:     Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-09-10T11:09:16Z
    Message:               Ignite ops-request has started to reconfigure Ignite nodes
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-09-10T11:09:24Z
    Message:               successfully reconciled the Ignite with new configure
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-09-10T11:09:29Z
    Message:               get pod; ConditionStatus:True; PodName:ig-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ig-cluster-0
    Last Transition Time:  2024-09-10T11:09:29Z
    Message:               evict pod; ConditionStatus:True; PodName:ig-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ig-cluster-0
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
    Message:               Successfully completed reconfigure Ignite
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                 Age    From                         Message
  ----     ------                                                 ----   ----                         -------
  Normal   Starting                                               6m13s  KubeDB Ops-manager Operator  Start processing for IgniteOpsRequest: demo/reconfigure-ig-cluster
  Normal   Starting                                               6m13s  KubeDB Ops-manager Operator  Pausing Ignite databse: demo/ig-cluster
  Normal   Successful                                             6m13s  KubeDB Ops-manager Operator  Successfully paused Ignite database: demo/ig-cluster for IgniteOpsRequest: reconfigure
  Normal   UpdatePetSets                                          6m5s   KubeDB Ops-manager Operator  successfully reconciled the Ignite with new configure
  Warning  get pod; ConditionStatus:True; PodName:ig-cluster-0    6m     KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ig-cluster-0
  Warning  evict pod; ConditionStatus:True; PodName:ig-cluster-0  6m     KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ig-cluster-0
  Warning  running pod; ConditionStatus:False                     5m55s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Normal   RestartNodes                                           5m40s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                               5m40s  KubeDB Ops-manager Operator  Resuming Ignite database: demo/ig-cluster
  Normal   Successful                                             5m39s  KubeDB Ops-manager Operator  Successfully resumed Ignite database: demo/ig-cluster for IgniteOpsRequest: reconfigure-ig-cluster
```

### Reconfigure using apply config

Let's say you are in a rush or, don't want to create a secret for updating configuration. You can directly do that using the following manifest.

#### Create IgniteOpsRequest

Now, we will use the new configuration in the `configuration.applyConfig` field in the `IgniteOpsRequest` CR. The `IgniteOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: reconfigure-apply
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: ig-cluster
  configuration:
    applyConfig:
      node-configuration.xml: |
        <?xml version="1.0" encoding="UTF-8"?>
        <beans xmlns="http://www.springframework.org/schema/beans"
        xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
        xsi:schemaLocation="http://www.springframework.org/schema/beans
        http://www.springframework.org/schema/beans/spring-beans-3.0.xsd">
        <!-- Your Ignite Configuration -->
        <bean class="org.apache.ignite.configuration.IgniteConfiguration">
    
        <property name="authenticationEnabled" value="true"/>

        </bean>
        </beans>
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `ig-cluster` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.applyConfig` specifies the new configuration that will be merged in the existing secret.

Let's create the `IgniteOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ignite/opsrequests/ignite-reconfigure-apply.yaml
igniteopsrequest.ops.kubedb.com/reconfigure-apply created
```

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete ig -n demo ig-cluster
kubectl delete igniteopsrequest -n demo reconfigure-apply reconfigure-ig-cluster
```