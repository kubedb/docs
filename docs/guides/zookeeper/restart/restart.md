---
title: Restart ZooKeeper
menu:
  docs_{{ .version }}:
    identifier: zk-restart-details
    name: Restart Cluster
    parent: zk-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart ZooKeeper

KubeDB supports restarting the ZooKeeper database via a ZooKeeperOpsRequest. Restarting is useful if some pods are got stuck in some phase, or they are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/zookeeper](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/zookeeper) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy ZooKeeper

In this section, we are going to deploy a ZooKeeper database using KubeDB.

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
  storage:
    resources:
      requests:
        storage: "1Gi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: "WipeOut"

```

Let's create the `ZooKeeper` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/restart/zookeeper.yaml
zookeeper.kubedb.com/zk-quickstart created
```

## Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: zk-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: zk-quickstart
  timeout: 3m
  apply: Always
```

- `spec.type` specifies the Type of the ops Request
- `spec.databaseRef` holds the name of the ZooKeeper database.  The db should be available in the same namespace as the opsRequest
- The meaning of `spec.timeout` & `spec.apply` fields will be found [here](/docs/guides/zookeeper/concepts/opsrequest.md#spectimeout)

> Note: The method of restarting the standalone & clustered zookeeper is exactly same as above. All you need, is to specify the corresponding ZooKeeper name in `spec.databaseRef.name` section.

Let's create the `ZooKeeperOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/restart/ops.yaml
zookeeperopsrequest.ops.kubedb.com/zk-restart created
```

Now the Ops-manager operator will restart the pods sequentially by their cardinal suffix.

```shell
$ kubectl get zookeeperopsrequest -n demo
NAME      TYPE      STATUS       AGE
zk-restart   Restart   Successful   10m

$ kubectl get zookeeperopsrequest -n demo -oyaml zk-restart
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"ZooKeeperOpsRequest","metadata":{"annotations":{},"name":"zk-restart","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"zk-quickstart"},"timeout":"3m","type":"Restart"}}
  creationTimestamp: "2024-10-22T09:22:57Z"
  generation: 1
  name: zk-restart
  namespace: demo
  resourceVersion: "1072309"
  uid: 6091d9fa-1c2b-4734-bdd1-1ace91460bea
spec:
  apply: Always
  databaseRef:
    name: zk-quickstart
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2024-10-22T09:22:57Z"
    message: ZooKeeper ops-request has started to restart zookeeper nodes
    observedGeneration: 1
    reason: Restart
    status: "True"
    type: Restart
  - lastTransitionTime: "2024-10-22T09:25:45Z"
    message: Successfully Restarted ZooKeeper nodes
    observedGeneration: 1
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - lastTransitionTime: "2024-10-22T09:23:05Z"
    message: get pod; ConditionStatus:True; PodName:zk-quickstart-0
    observedGeneration: 1
    status: "True"
    type: GetPod--zk-quickstart-0
  - lastTransitionTime: "2024-10-22T09:23:05Z"
    message: evict pod; ConditionStatus:True; PodName:zk-quickstart-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--zk-quickstart-0
  - lastTransitionTime: "2024-10-22T09:23:10Z"
    message: running pod; ConditionStatus:False
    observedGeneration: 1
    status: "False"
    type: RunningPod
  - lastTransitionTime: "2024-10-22T09:23:45Z"
    message: get pod; ConditionStatus:True; PodName:zk-quickstart-1
    observedGeneration: 1
    status: "True"
    type: GetPod--zk-quickstart-1
  - lastTransitionTime: "2024-10-22T09:23:45Z"
    message: evict pod; ConditionStatus:True; PodName:zk-quickstart-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--zk-quickstart-1
  - lastTransitionTime: "2024-10-22T09:24:25Z"
    message: get pod; ConditionStatus:True; PodName:zk-quickstart-2
    observedGeneration: 1
    status: "True"
    type: GetPod--zk-quickstart-2
  - lastTransitionTime: "2024-10-22T09:24:25Z"
    message: evict pod; ConditionStatus:True; PodName:zk-quickstart-2
    observedGeneration: 1
    status: "True"
    type: EvictPod--zk-quickstart-2
  - lastTransitionTime: "2024-10-22T09:25:05Z"
    message: get pod; ConditionStatus:True; PodName:zk-quickstart-3
    observedGeneration: 1
    status: "True"
    type: GetPod--zk-quickstart-3
  - lastTransitionTime: "2024-10-22T09:25:05Z"
    message: evict pod; ConditionStatus:True; PodName:zk-quickstart-3
    observedGeneration: 1
    status: "True"
    type: EvictPod--zk-quickstart-3
  - lastTransitionTime: "2024-10-22T09:25:45Z"
    message: Controller has successfully restart the ZooKeeper replicas
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful

```


## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete zookeeperopsrequest -n demo zk-restart
kubectl delete zookeeper -n demo zk-quickstart
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [ZooKeeper object](/docs/guides/zookeeper/concepts/zookeeper.md).
- Detail concepts of [ZooKeeper object](/docs/guides/zookeeper/concepts/zookeeper.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
