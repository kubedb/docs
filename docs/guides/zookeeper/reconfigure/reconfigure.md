---
title: Reconfigure ZooKeeper Ensemble
menu:
  docs_{{ .version }}:
    identifier: zk-ensemble-reconfigure
    name: Reconfigure Configurations
    parent: zk-reconfigure
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure ZooKeeper Ensemble

This guide will show you how to use `KubeDB` Ops-manager operator to reconfigure a ZooKeeper ensemble.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [ZooKeeper](/docs/guides/zookeeper/concepts/zookeeper.md)
  - [ZooKeeperOpsRequest](/docs/guides/zookeeper/concepts/opsrequest.md)
  - [Reconfigure Overview](/docs/guides/zookeeper/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [examples](/docs/examples/zookeeper) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

Now, we are going to deploy a  `ZooKeeper` cluster using a supported version by `KubeDB` operator. Then we are going to apply `ZooKeeperOpsRequest` to reconfigure its configuration.

### Prepare ZooKeeper Ensemble

Now, we are going to deploy a `ZooKeeper` cluster with version `3.8.3`.

### Deploy ZooKeeper Ensemble

At first, we will create `secret` named zk-configuration containing required configuration settings.

```yaml
apiVersion: v1
stringData:
  zoo.cfg: |
    maxClientCnxns=70
kind: Secret
metadata:
  name: zk-configuration
  namespace: demo
```
Here, `maxClientCnxns` is set to `70`, whereas the default value is `60`.

Now, we will apply the secret with custom configuration.
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/reconfiguration/secret.yaml
secret/zk-configuration created
```

In this section, we are going to create a ZooKeeper object specifying `spec.configSecret` field to apply this custom configuration. Below is the YAML of the `ZooKeeper` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: zk-quickstart
  namespace: demo
spec:
  version: "3.8.3"
  adminServerPort: 8080
  replicas: 3
  configSecret:
    name: zk-configuration
  storage:
    resources:
      requests:
        storage: "1Gi"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: "WipeOut"
```

Let's create the `ZooKeeper` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/reconfiguration/sample-zk-configuration.yaml
zookeeper.kubedb.com/zk-quickstart created
```

Now, wait until `zk-quickstart` has status `Ready`. i.e,

```bash
$ kubectl get zk -n demo
NAME               VERSION     STATUS    AGE
zk-quickstart      3.8.3      Ready     23s
```

Now, we will check if the database has started with the custom configuration we have provided.

Now, you can exec into the zookeeper pod and find if the custom configuration is there,

```bash
$ Defaulted container "zookeeper" out of: zookeeper, zookeeper-init (init)
zookeeper@zk-quickstart-0:/apache-zookeeper-3.8.3-bin$ echo conf | nc localhost 2181
clientPort=2181
secureClientPort=-1
dataDir=/data/version-2
dataDirSize=134218330
dataLogDir=/data/version-2
dataLogSize=134218330
tickTime=2000
maxClientCnxns=70
minSessionTimeout=4000
maxSessionTimeout=40000
clientPortListenBacklog=-1
serverId=1
initLimit=10
syncLimit=2
electionAlg=3
electionPort=3888
quorumPort=2888
peerType=0
membership: 
server.1=zk-quickstart-0.zk-quickstart-pods.demo.svc.cluster.local:2888:3888:participant;0.0.0.0:2181
server.2=zk-quickstart-1.zk-quickstart-pods.demo.svc.cluster.local:2888:3888:participant;0.0.0.0:2181
server.3=zk-quickstart-2.zk-quickstart-pods.demo.svc.cluster.local:2888:3888:participant;0.0.0.0:2181
version=100000011zookeeper@zk-quickstart-0:/apache-zookeeper-3.8.3-bin$ exit
exit
```

As we can see from the configuration of running zookeeper, the value of `maxClientCnxns` has been set to `70`.

### Reconfigure using new secret

Now we will reconfigure this database to set `maxClientCnxns` to `100`.

At first, we will create `secret` named new-configuration containing required configuration settings.

```yaml
apiVersion: v1
stringData:
  zoo.cfg: |
    maxClientCnxns=100
kind: Secret
metadata:
  name: zk-new-configuration
  namespace: demo
```
Here, `maxClientCnxns` is set to `100`.

Now, we will apply the secret with custom configuration.
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/reconfiguration/new-secret.yaml
secret/zk-new-configuration created
```

#### Create ZooKeeperOpsRequest

Now, we will use this secret to replace the previous secret using a `ZooKeeperOpsRequest` CR. The `ZooKeeperOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: zk-reconfig
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: zk-quickstart
  configuration:
    configSecret:
      name: zk-new-configuration
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `zk-quickstart` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.configSecret.name` specifies the name of the new secret.

Let's create the `ZooKeeperOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/reconfiguration/zkops-reconfiguration.yaml
zookeeperopsrequest.ops.kubedb.com/zk-reconfig created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Ops-manager operator will update the `configSecret` of `ZooKeeper` object.

Let's wait for `ZooKeeperOpsRequest` to be `Successful`.  Run the following command to watch `ZooKeeperOpsRequest` CR,

```bash
$ watch kubectl get zookeeperopsrequest -n demo
Every 2.0s: kubectl get zookeeperopsrequest -n demo
NAME               TYPE          STATUS       AGE
zk-reconfig        Reconfigure   Successful   1m
```

We can see from the above output that the `ZooKeeperOpsRequest` has succeeded. If we describe the `ZooKeeperOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe zookeeperopsrequest -n demo zk-reconfig
Name:         zk-reconfig
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ZooKeeperOpsRequest
Metadata:
  Creation Timestamp:  2024-10-30T08:27:00Z
  Generation:          1
  Resource Version:    1548116
  UID:                 4f3daa11-c41b-4079-a8d8-1040931284ef
Spec:
  Apply:  IfReady
  Configuration:
    Config Secret:
      Name:  zk-new-configuration
  Database Ref:
    Name:  zk-quickstart
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-10-30T08:27:00Z
    Message:               ZooKeeper ops-request has started to reconfigure ZooKeeper nodes
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-10-30T08:27:08Z
    Message:               successfully reconciled the ZooKeeper with new configure
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-30T08:29:18Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-10-30T08:27:13Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-0
    Last Transition Time:  2024-10-30T08:27:13Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-0
    Last Transition Time:  2024-10-30T08:27:18Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2024-10-30T08:27:58Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-1
    Last Transition Time:  2024-10-30T08:27:58Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-1
    Last Transition Time:  2024-10-30T08:28:38Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-2
    Last Transition Time:  2024-10-30T08:28:38Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-2
    Last Transition Time:  2024-10-30T08:29:18Z
    Message:               Successfully completed reconfigure ZooKeeper
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:                    <none>
```

Now need to check the new configuration we have provided.

Now, wait until `zk-quickstart` has status `Ready`. i.e,

```bash
$ kubectl get zk -n demo
NAME            VERSION     STATUS    AGE
zk-quickstart   3.8.3      Ready     20s
```

Now let’s exec into the zookeeper pod and check the new configuration we have provided.

```bash
$ Defaulted container "zookeeper" out of: zookeeper, zookeeper-init (init)
zookeeper@zk-quickstart-0:/apache-zookeeper-3.8.3-bin$ echo conf | nc localhost 2181
clientPort=2181
secureClientPort=-1
dataDir=/data/version-2
dataDirSize=134218330
dataLogDir=/data/version-2
dataLogSize=134218330
tickTime=2000
maxClientCnxns=100
minSessionTimeout=4000
maxSessionTimeout=40000
clientPortListenBacklog=-1
serverId=1
initLimit=10
syncLimit=2
electionAlg=3
electionPort=3888
quorumPort=2888
peerType=0
membership: 
server.1=zk-quickstart-0.zk-quickstart-pods.demo.svc.cluster.local:2888:3888:participant;0.0.0.0:2181
server.2=zk-quickstart-1.zk-quickstart-pods.demo.svc.cluster.local:2888:3888:participant;0.0.0.0:2181
server.3=zk-quickstart-2.zk-quickstart-pods.demo.svc.cluster.local:2888:3888:participant;0.0.0.0:2181
version=100000011zookeeper@zk-quickstart-0:/apache-zookeeper-3.8.3-bin$ exit
exit
```

As we can see from the configuration of running zookeeper, the value of `maxClientCnxns` has been changed from `70` to `100`. So the reconfiguration of the zookeeper is successful.

### Reconfigure using apply config

Now we will reconfigure this database again to set `maxClientCnxns` to `90`. This time we won't use a new secret. We will use the `applyConfig` field of the `ZooKeeperOpsRequest`. This will merge the new config in the existing secret.

#### Create ZooKeeperOpsRequest

Now, we will use the new configuration in the `data` field in the `ZooKeeperOpsRequest` CR. The `ZooKeeperOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: zk-reconfig-apply
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: zk-quickstart
  configuration:
    applyConfig:
      zoo.cfg: |
        maxClientCnxns=90
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `zk-quickstart` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.applyConfig` specifies the new configuration that will be merged in the existing secret.

Let's create the `ZooKeeperOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/reconfiguration/zkops-apply-reconfiguration.yaml
zookeeperopsrequest.ops.kubedb.com/zk-reconfig-apply created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Ops-manager operator will merge this new config with the existing configuration.

Let's wait for `ZooKeeperOpsRequest` to be `Successful`.  Run the following command to watch `ZooKeeperOpsRequest` CR,

```bash
$ watch kubectl get zookeeperopsrequest -n demo
NAME                 TYPE          STATUS       AGE
zk-reconfig-apply    Reconfigure   Successful   38s
```

We can see from the above output that the `ZooKeeperOpsRequest` has succeeded. If we describe the `ZooKeeperOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe zookeeperopsrequest -n demo zk-reconfig-apply
Name:         zk-reconfig-apply
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ZooKeeperOpsRequest
Metadata:
  Creation Timestamp:  2024-10-30T08:27:00Z
  Generation:          1
  Resource Version:    1548116
  UID:                 4f3daa11-c41b-4079-a8d8-1040931284ef
Spec:
  Apply:  IfReady
  Configuration:
    Config Secret:
      Name:  zk-new-configuration
  Database Ref:
    Name:  zk-quickstart
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-10-30T08:27:00Z
    Message:               ZooKeeper ops-request has started to reconfigure ZooKeeper nodes
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-10-30T08:27:08Z
    Message:               successfully reconciled the ZooKeeper with new configure
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-30T08:29:18Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-10-30T08:27:13Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-0
    Last Transition Time:  2024-10-30T08:27:13Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-0
    Last Transition Time:  2024-10-30T08:27:18Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2024-10-30T08:27:58Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-1
    Last Transition Time:  2024-10-30T08:27:58Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-1
    Last Transition Time:  2024-10-30T08:28:38Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-2
    Last Transition Time:  2024-10-30T08:28:38Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-2
    Last Transition Time:  2024-10-30T08:29:18Z
    Message:               Successfully completed reconfigure ZooKeeper
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:                    <none>
```

Now need to check the new configuration we have provided.

Now, wait until `zk-quickstart` has status `Ready`. i.e,

```bash
$ kubectl get zk -n demo
NAME            VERSION     STATUS    AGE
zk-quickstart   3.8.3      Ready     20s
```

Now let’s exec into the zookeeper pod and check the new configuration we have provided.

```bash
$ Defaulted container "zookeeper" out of: zookeeper, zookeeper-init (init)
zookeeper@zk-quickstart-0:/apache-zookeeper-3.8.3-bin$ echo conf | nc localhost 2181
clientPort=2181
secureClientPort=-1
dataDir=/data/version-2
dataDirSize=134218330
dataLogDir=/data/version-2
dataLogSize=134218330
tickTime=2000
maxClientCnxns=90
minSessionTimeout=4000
maxSessionTimeout=40000
clientPortListenBacklog=-1
serverId=1
initLimit=10
syncLimit=2
electionAlg=3
electionPort=3888
quorumPort=2888
peerType=0
membership: 
server.1=zk-quickstart-0.zk-quickstart-pods.demo.svc.cluster.local:2888:3888:participant;0.0.0.0:2181
server.2=zk-quickstart-1.zk-quickstart-pods.demo.svc.cluster.local:2888:3888:participant;0.0.0.0:2181
server.3=zk-quickstart-2.zk-quickstart-pods.demo.svc.cluster.local:2888:3888:participant;0.0.0.0:2181
version=100000011zookeeper@zk-quickstart-0:/apache-zookeeper-3.8.3-bin$ exit
exit
```

As we can see from the configuration of running zookeeper, the value of `maxClientCnxns` has been changed from `100` to `90`. So, the reconfiguration of the database using the `applyConfig` field is successful.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete zk -n demo zk-quickstart
kubectl delete zookeeperopsrequest -n demo zk-reconfig zk-reconfig-apply
```