---
title: Restart RabbitMQ
menu:
  docs_{{ .version }}:
    identifier: rm-restart-details
    name: Restart RabbitMQ
    parent: rm-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart RabbitMQ

KubeDB supports restarting the RabbitMQ database via a RabbitMQOpsRequest. Restarting is useful if some pods are got stuck in some phase, or they are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/rabbitmq](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/rabbitmq) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy RabbitMQ

In this section, we are going to deploy a RabbitMQ database using KubeDB.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rm
  namespace: demo
spec:
  version: "3.13.2"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `RabbitMQ` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/restart/rm.yaml
rabbitmq.kubedb.com/rm created
```

## Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RabbitMQOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: rm
  timeout: 3m
  apply: Always
```

- `spec.type` specifies the Type of the ops Request
- `spec.databaseRef` holds the name of the RabbitMQ database.  The db should be available in the same namespace as the opsRequest
- The meaning of `spec.timeout` & `spec.apply` fields will be found [here](/docs/guides/rabbitmq/concepts/opsrequest.md#spectimeout)

> Note: The method of restarting the standalone & sharded db is exactly same as above. All you need, is to specify the corresponding RabbitMQ name in `spec.databaseRef.name` section.

Let's create the `RabbitMQOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/restart/ops.yaml
rabbitmqopsrequest.ops.kubedb.com/restart created
```

Now the Ops-manager operator will first restart the general secondary pods, then serially the arbiters, the hidden nodes, & lastly will restart the Primary of the database.

```shell
$ kubectl get mgops -n demo
NAME      TYPE      STATUS       AGE
restart   Restart   Successful   10m

$ kubectl get mgops -n demo -oyaml restart
apiVersion: ops.kubedb.com/v1alpha1
kind: RabbitMQOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"RabbitMQOpsRequest","metadata":{"annotations":{},"name":"restart","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"mongo"},"readinessCriteria":{"objectsCountDiffPercentage":15,"oplogMaxLagSeconds":10},"timeout":"3m","type":"Restart"}}
  creationTimestamp: "2022-10-31T08:54:45Z"
  generation: 1
  name: restart
  namespace: demo
  resourceVersion: "738625"
  uid: 32f6c52f-6114-4e25-b3a1-877223cf7145
spec:
  apply: Always
  databaseRef:
    name: mongo
  readinessCriteria:
    objectsCountDiffPercentage: 15
    oplogMaxLagSeconds: 10
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2022-10-31T08:54:45Z"
    message: RabbitMQ ops request is restarting the database nodes
    observedGeneration: 1
    reason: Restart
    status: "True"
    type: Restart
  - lastTransitionTime: "2022-10-31T08:57:05Z"
    message: Successfully Restarted ReplicaSet nodes
    observedGeneration: 1
    reason: RestartReplicaSet
    status: "True"
    type: RestartReplicaSet
  - lastTransitionTime: "2022-10-31T08:57:05Z"
    message: Successfully restarted all nodes of RabbitMQ
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
kubectl delete rabbitmqopsrequest -n demo restart
kubectl delete rabbitmq -n demo rm
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [RabbitMQ object](/docs/guides/rabbitmq/concepts/rabbitmq.md).
- Monitor your RabbitMQ database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/rabbitmq/monitoring/using-prometheus-operator.md).
- Monitor your RabbitMQ database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/rabbitmq/monitoring/using-builtin-prometheus.md).
- Detail concepts of [RabbitMQ object](/docs/guides/rabbitmq/concepts/rabbitmq.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
