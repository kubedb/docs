---
title: Reprovision MongoDB
menu:
  docs_{{ .version }}:
    identifier: mg-reprovision-details
    name: Reprovision MongoDB
    parent: mg-reprovision
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reprovision MongoDB

KubeDB supports reprovisioning the MongoDB database via a MongoDBOpsRequest. Reprovisioning is useful if you want, for some reason, to deploy a new MongoDB with the same specifications. This tutorial will show you how to use that.

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
apiVersion: kubedb.com/v1
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
- `spec.ephemeralStorage` holds the emptyDir volume specifications. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. So, each members will have a pod of this ephemeral storage configuration.
- `spec.arbiter` denotes arbiter-node spec of the deployed MongoDB CRD.
- `spec.hidden` denotes hidden-node spec of the deployed MongoDB CRD.

Let's create the `MongoDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reprovision/mongo.yaml
mongodb.kubedb.com/mongo created
```

## Apply Reprovision opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: repro
  namespace: demo
spec:
  type: Reprovision
  databaseRef:
    name: mongo
  apply: Always
```

- `spec.type` specifies the Type of the ops Request
- `spec.databaseRef` holds the name of the MongoDB database.  The db should be available in the same namespace as the opsRequest
- `spec.apply` is set to Always to denote that, we want reprovisioning even if the db was not Ready.

> Note: The method of reprovisioning the standalone & sharded db is exactly same as above. All you need, is to specify the corresponding MongoDB name in `spec.databaseRef.name` section.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reprovision/ops.yaml
mongodbopsrequest.ops.kubedb.com/repro created
```

Now the Ops-manager operator will
1) Pause the DB
2) Delete all petsets
3) Remove `Provisioned` condition from db
4) Reconcile the db for start
5) Wait for DB to be Ready. 

```shell
$ kubectl get mgops -n demo
NAME    TYPE          STATUS       AGE
repro   Reprovision   Successful   2m


$ kubectl get mgops -n demo -oyaml repro
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"MongoDBOpsRequest","metadata":{"annotations":{},"name":"repro","namespace":"demo"},"spec":{"databaseRef":{"name":"mongo"},"type":"Reprovision"}}
  creationTimestamp: "2022-10-31T09:50:35Z"
  generation: 1
  name: repro
  namespace: demo
  resourceVersion: "743676"
  uid: b3444d38-bef3-4043-925f-551fe6c86123
spec:
  apply: Always
  databaseRef:
    name: mongo
  type: Reprovision
status:
  conditions:
  - lastTransitionTime: "2022-10-31T09:50:35Z"
    message: MongoDB ops request is reprovisioning the database
    observedGeneration: 1
    reason: Reprovision
    status: "True"
    type: Reprovision
  - lastTransitionTime: "2022-10-31T09:50:45Z"
    message: Successfully Deleted All the PetSets
    observedGeneration: 1
    reason: DeletePetSets
    status: "True"
    type: DeletePetSets
  - lastTransitionTime: "2022-10-31T09:52:05Z"
    message: Database Phase is Ready
    observedGeneration: 1
    reason: DatabaseReady
    status: "True"
    type: DatabaseReady
  - lastTransitionTime: "2022-10-31T09:52:05Z"
    message: Successfully Reprovisioned the database
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
kubectl delete mongodbopsrequest -n demo repro
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
