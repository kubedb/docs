---
title: Restart MongoDB
menu:
  docs_{{ .version }}:
    identifier: mg-restart-details
    name: Restart MongoDB
    parent: mg-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart MongoDB

KubeDB supports restarting the MongoDB database via a MongoDBOpsRequest. Restarting is useful if some pods are got stuck in some phase, or they are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/mongodb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy MongoDB

In this section, we are going to deploy a MongoDB database using KubeDB.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mongo
  namespace: demo
spec:
  version: "4.4.26"
  replicaSet:
    name: "replicaset"
  podTemplate:
    spec:
      resources:
        requests:
          cpu: "300m"
          memory: "300Mi"
  replicas: 2
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
  arbiter: {}
  hidden:
    replicas: 2
    storage:
      storageClassName: "standard"
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 2Gi
```

- `spec.replicaSet` represents the configuration for replicaset.
    - `name` denotes the name of mongodb replicaset.
- `spec.replicas` denotes the number of general members in `rs0` mongodb replicaset.
- `spec.podTemplate` denotes specifications of all the 3 general replicaset members.
- `spec.ephemeralStorage` holds the emptyDir volume specifications. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. So, each members will have a pod of this ephemeral storage configuration.
- `spec.arbiter` denotes arbiter-node spec of the deployed MongoDB CRD. 
- `spec.hidden` denotes hidden-node spec of the deployed MongoDB CRD.

Let's create the `MongoDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/restart/mongo.yaml
mongodb.kubedb.com/mongo created
```

## Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: mongo
  readinessCriteria:
    oplogMaxLagSeconds: 10
    objectsCountDiffPercentage: 15
  timeout: 3m
  apply: Always
```

- `spec.type` specifies the Type of the ops Request
- `spec.databaseRef` holds the name of the MongoDB database.  The db should be available in the same namespace as the opsRequest
- The meaning of`spec.readinessCriteria`, `spec.timeout` & `spec.apply` fields will be found [here](/docs/guides/mongodb/concepts/opsrequest.md#specreadinessCriteria)

> Note: The method of restarting the standalone & sharded db is exactly same as above. All you need, is to specify the corresponding MongoDB name in `spec.databaseRef.name` section.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/restart/ops.yaml
mongodbopsrequest.ops.kubedb.com/restart created
```

Now the Ops-manager operator will first restart the general secondary pods, then serially the arbiters, the hidden nodes, & lastly will restart the Primary of the database.

```shell
$ kubectl get mgops -n demo
NAME      TYPE      STATUS       AGE
restart   Restart   Successful   10m

$ kubectl get mgops -n demo -oyaml restart
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"MongoDBOpsRequest","metadata":{"annotations":{},"name":"restart","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"mongo"},"readinessCriteria":{"objectsCountDiffPercentage":15,"oplogMaxLagSeconds":10},"timeout":"3m","type":"Restart"}}
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
    message: MongoDB ops request is restarting the database nodes
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
    message: Successfully restarted all nodes of MongoDB
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
kubectl delete mongodbopsrequest -n demo restart
kubectl delete mongodb -n demo mongo
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Initialize [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Use [kubedb cli](/docs/guides/mongodb/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
