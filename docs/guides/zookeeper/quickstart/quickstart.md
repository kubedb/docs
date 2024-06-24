---
title: ZooKeeper Quickstart
menu:
  docs_{{ .version }}:
    identifier: zk-quickstart-quickstart
    name: Overview
    parent: zk-quickstart-zookeeper
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ZooKeeper QuickStart

This tutorial will show you how to use KubeDB to run a ZooKeeper server.

<p align="center">
  <img alt="lifecycle"  src="/docs/images/zookeeper/zookeeper-lifecycle.png">
</p>

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Please set `global.featureGates.ZooKeeper=true`
to install ZooKeeper CRDs.

- [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) is required to run KubeDB. Check the available StorageClass in cluster.

  ```bash
  $ kubectl get storageclasses
  NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
  standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  20h
  ```

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create namespace demo
  namespace/demo created

  $ kubectl get namespaces
  NAME          STATUS    AGE
  demo          Active    10s
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Available ZooKeeperVersions

When you have installed KubeDB, it has created `ZooKeeperVersions` crd for all supported ZooKeeper versions. Check:

```bash
$ kubectl get zookeeperversions
NAME    VERSION   DB_IMAGE                                  DEPRECATED   AGE
3.7.2   3.7.2     ghcr.io/appscode-images/zookeeper:3.7.2                94s
3.8.3   3.8.3     ghcr.io/appscode-images/zookeeper:3.8.3                94s
3.9.1   3.9.1     ghcr.io/appscode-images/zookeeper:3.9.1                94s
```

## Create a ZooKeeper server

KubeDB implements a `ZooKeeper` CRD to define the specification of a ZooKeeper server. Below is the `ZooKeeper` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: zk-quickstart
  namespace: demo
spec:
  version: "3.9.1"
  adminServerPort: 8080
  replicas: 3
  storage:
    resources:
      requests:
        storage: "1Gi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: "WipeOut"
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/quickstart/zoo.yaml
zookeeper.kubedb.com/zk-quickstart created
```

Here,

- `spec.version` is name of the ZooKeeperVersion crd where the docker images are specified. In this tutorial, a ZooKeeper 3.9.1 database is created.
- `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `ZooKeeper` crd or which resources KubeDB should keep or delete when you delete `ZooKeeper` crd. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

> Note: `spec.storage` section is used to create PVC for database pod. It will create PVC with storage size specified in storage.resources.requests field. Don't specify limits here. PVC does not get resized automatically.

KubeDB operator watches for `ZooKeeper` objects using Kubernetes api. When a `ZooKeeper` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching ZooKeeper object name. KubeDB operator will also create a governing service for StatefulSets with the name `kubedb`, if one is not already present.

```bash
$ kubectl get zk -n demo
NAME            TYPE                  VERSION   STATUS   AGE
zk-quickstart   kubedb.com/v1alpha2   3.9.1     Ready    105s

$ kubectl describe zk -n demo zk-quickstart
Name:         zk-quickstart
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         ZooKeeper
Metadata:
  Creation Timestamp:  2024-05-02T08:25:26Z
  Finalizers:
    kubedb.com
  Generation:        3
  Resource Version:  4219
  UID:               dd69e514-3049-4d08-8b57-92f8246dda35
Spec:
  Admin Server Port:  8080
  Auth Secret:
    Name:  zk-quickstart-auth
  Health Checker:
    Failure Threshold:  3
    Period Seconds:     20
    Timeout Seconds:    10
  Pod Placement Policy:
    Name:  default
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Containers:
        Name:  zookeeper
        Resources:
          Limits:
            Memory:  1Gi
          Requests:
            Cpu:     500m
            Memory:  1Gi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Group:     1000
          Run As Non Root:  true
          Run As User:      1000
          Seccomp Profile:
            Type:  RuntimeDefault
      Init Containers:
        Name:  zookeeper-init
        Resources:
          Limits:
            Memory:  512Mi
          Requests:
            Cpu:     200m
            Memory:  512Mi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Group:     1000
          Run As Non Root:  true
          Run As User:      1000
          Seccomp Profile:
            Type:  RuntimeDefault
      Security Context:
        Fs Group:  1000
  Replicas:        3
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         1Gi
    Storage Class Name:  standard
  Deletion Policy:    WipeOut
  Version:               3.9.1
Status:
  Conditions:
    Last Transition Time:  2024-05-02T08:25:26Z
    Message:               The KubeDB operator has started the provisioning of ZooKeeper: demo/zk-quickstart
    Observed Generation:   1
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2024-05-02T08:25:50Z
    Message:               All replicas are ready for ZooKeeper demo/zk-quickstart
    Observed Generation:   3
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2024-05-02T08:26:10Z
    Message:               The ZooKeeper: demo/zk-quickstart is accepting connection requests.
    Observed Generation:   3
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2024-05-02T08:26:10Z
    Message:               The ZooKeeper: demo/zk-quickstart is ready.
    Observed Generation:   3
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2024-05-02T08:26:13Z
    Message:               ZooKeeper: demo/zk-quickstart is successfully provisioned.
    Observed Generation:   3
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:                    <none>


$ kubectl get petset -n demo
NAME            AGE
zk-quickstart   3m14s


$ kubectl get pvc -n demo
NAME                                 STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
zk-quickstart-data-zk-quickstart-0   Bound    pvc-1e1850b8-4e5c-418c-a722-89df98f28998   1Gi        RWO            standard       3m40s
zk-quickstart-data-zk-quickstart-1   Bound    pvc-e2bb4b02-b138-4589-9e43-bcaf599b6513   1Gi        RWO            standard       3m31s
zk-quickstart-data-zk-quickstart-2   Bound    pvc-988ab6b2-e5ed-4c75-8418-31186bd1d3db   1Gi        RWO            standard       3m25s


$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                     STORAGECLASS   REASON   AGE
pvc-1e1850b8-4e5c-418c-a722-89df98f28998   1Gi        RWO            Delete           Bound    demo/zk-quickstart-data-zk-quickstart-0   standard                3m52s
pvc-988ab6b2-e5ed-4c75-8418-31186bd1d3db   1Gi        RWO            Delete           Bound    demo/zk-quickstart-data-zk-quickstart-2   standard                3m40s
pvc-e2bb4b02-b138-4589-9e43-bcaf599b6513   1Gi        RWO            Delete           Bound    demo/zk-quickstart-data-zk-quickstart-1   standard                3m46s


$ kubectl get service -n demo
NAME                         TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
zk-quickstart                ClusterIP   10.96.26.38    <none>        2181/TCP                     4m15s
zk-quickstart-admin-server   ClusterIP   10.96.49.134   <none>        8080/TCP                     4m15s
zk-quickstart-pods           ClusterIP   None           <none>        2181/TCP,2888/TCP,3888/TCP   4m15s
```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. Run the following command to see the modified ZooKeeper object:

```bash
$ kubectl get zk -n demo zk-quickstart -o yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"ZooKeeper","metadata":{"annotations":{},"name":"zk-quickstart","namespace":"demo"},"spec":{"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"deletionPolicy":"WipeOut","version":"3.9.1"}}
  creationTimestamp: "2024-05-02T08:25:26Z"
  finalizers:
  - kubedb.com
  generation: 3
  name: zk-quickstart
  namespace: demo
  resourceVersion: "4219"
  uid: dd69e514-3049-4d08-8b57-92f8246dda35
spec:
  adminServerPort: 8080
  authSecret:
    name: zk-quickstart-auth
  healthChecker:
    failureThreshold: 3
    periodSeconds: 20
    timeoutSeconds: 10
  podPlacementPolicy:
    name: default
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      containers:
      - name: zookeeper
        resources:
          limits:
            memory: 1Gi
          requests:
            cpu: 500m
            memory: 1Gi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          runAsGroup: 1000
          runAsNonRoot: true
          runAsUser: 1000
          seccompProfile:
            type: RuntimeDefault
      initContainers:
      - name: zookeeper-init
        resources:
          limits:
            memory: 512Mi
          requests:
            cpu: 200m
            memory: 512Mi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          runAsGroup: 1000
          runAsNonRoot: true
          runAsUser: 1000
          seccompProfile:
            type: RuntimeDefault
      securityContext:
        fsGroup: 1000
  replicas: 3
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  deletionPolicy: WipeOut
  version: 3.9.1
status:
  conditions:
  - lastTransitionTime: "2024-05-02T08:25:26Z"
    message: 'The KubeDB operator has started the provisioning of ZooKeeper: demo/zk-quickstart'
    observedGeneration: 1
    reason: DatabaseProvisioningStartedSuccessfully
    status: "True"
    type: ProvisioningStarted
  - lastTransitionTime: "2024-05-02T08:25:50Z"
    message: All replicas are ready for ZooKeeper demo/zk-quickstart
    observedGeneration: 3
    reason: AllReplicasReady
    status: "True"
    type: ReplicaReady
  - lastTransitionTime: "2024-05-02T08:26:10Z"
    message: 'The ZooKeeper: demo/zk-quickstart is accepting connection requests.'
    observedGeneration: 3
    reason: DatabaseAcceptingConnectionRequest
    status: "True"
    type: AcceptingConnection
  - lastTransitionTime: "2024-05-02T08:26:10Z"
    message: 'The ZooKeeper: demo/zk-quickstart is ready.'
    observedGeneration: 3
    reason: ReadinessCheckSucceeded
    status: "True"
    type: Ready
  - lastTransitionTime: "2024-05-02T08:26:13Z"
    message: 'ZooKeeper: demo/zk-quickstart is successfully provisioned.'
    observedGeneration: 3
    reason: DatabaseSuccessfullyProvisioned
    status: "True"
    type: Provisioned
  phase: Ready
```

Now, you can connect to this database using created service. In this tutorial, we are connecting to the ZooKeeper server from inside of pod.

```bash
$ kubectl exec -it -n demo zk-quickstart-0 -- sh

$ echo ruok | nc localhost 2181
imok

$ zkCli.sh create /hello-dir hello-messege
Connecting to localhost:2181
...
Connection Log Messeges
...
Created /hello-dir

$ zkCli.sh get /hello-dir
Connecting to localhost:2181
...
Connection Log Messeges
...
hello-messege
```

## DoNotTerminate Property

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`. You can see this below:

```bash
$ kubectl delete zk zk-quickstart -n demo
Error from server (BadRequest): admission webhook "zookeeper.validators.kubedb.com" denied the request: zookeeper "zookeeper-quickstart" can't be deleted. To delete, change spec.deletionPolicy
```

Now, run `kubectl edit zk zookeeper-quickstart -n demo` to set `spec.deletionPolicy` to `Halt` . Then you will be able to delete/halt the database.


## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash

$ kubectl patch -n demo zk/zk-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
zookeeper.kubedb.com/zk-quickstart patched

$ kubectl delete -n demo zk/zk-quickstart
zookeeper.kubedb.com "zk-quickstart" deleted

$ kubectl delete ns demo
namespace "demo" deleted
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for production environment. You can follow these tips to avoid them.

**Use `deletionPolicy: WipeOut`**. It is nice to be able to resume database from previous one.So, we preserve all your `PVCs`, auth `Secrets`. If you don't want to resume database, you can just use `spec.deletionPolicy: WipeOut`. It will delete everything created by KubeDB for a particular ZooKeeper crd when you delete the crd. 

## Next Steps

- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
