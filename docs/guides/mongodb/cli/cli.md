---
title: CLI | KubeDB
menu:
  docs_{{ .version }}:
    identifier: mg-cli-cli
    name: Quickstart
    parent: mg-cli-mongodb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Manage KubeDB objects using CLIs

## KubeDB CLI

KubeDB comes with its own cli. It is called `kubedb` cli. `kubedb` can be used to manage any KubeDB object. `kubedb` cli also performs various validations to improve ux. To install KubeDB cli on your workstation, follow the steps [here](/docs/setup/README.md).

### How to Create objects

`kubectl create` creates a database CRD object in `default` namespace by default. Following command will create a MongoDB object as specified in `mongodb.yaml`.

```bash
$ kubectl create -f mongodb-demo.yaml
mongodb.kubedb.com/mongodb-demo created
```

You can provide namespace as a flag `--namespace`. Provided namespace should match with namespace specified in input file.

```bash
$ kubectl create -f mongodb-demo.yaml --namespace=kube-system
mongodb.kubedb.com/mongodb-demo
```

`kubectl create` command also considers `stdin` as input.

```bash
cat mongodb-demo.yaml | kubectl create -f -
```

### How to List Objects

`kubectl get` command allows users to list or find any KubeDB object. To list all MongoDB objects in `default` namespace, run the following command:

```bash
$ kubectl get mongodb
NAME           VERSION   STATUS    AGE
mongodb-demo   3.4-v3    Ready     13m
mongodb-dev    3.4-v3    Ready     11m
mongodb-prod   3.4-v3    Ready     11m
mongodb-qa     3.4-v3    Ready     10m
```

To get YAML of an object, use `--output=yaml` flag.

```yaml
$ kubectl get mongodb mongodb-demo --output=yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  creationTimestamp: "2019-02-06T10:31:04Z"
  finalizers:
  - kubedb.com
  generation: 2
  name: mongodb-demo
  namespace: demo
  resourceVersion: "94703"
  selfLink: /apis/kubedb.com/v1/namespaces/default/mongodbs/mongodb-demo
  uid: 4eaaba0e-29fa-11e9-aebf-080027875192
spec:
  authSecret:
    name: mongodb-demo-auth
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      livenessProbe:
        exec:
          command:
          - mongo
          - --eval
          - db.adminCommand('ping')
        failureThreshold: 3
        periodSeconds: 10
        successThreshold: 1
        timeoutSeconds: 5
      readinessProbe:
        exec:
          command:
          - mongo
          - --eval
          - db.adminCommand('ping')
        failureThreshold: 3
        periodSeconds: 10
        successThreshold: 1
        timeoutSeconds: 1
      resources: {}
  replicas: 1
  storage:
    accessModes:
    - ReadWriteOnce
    dataSource: null
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: Halt
  version: 3.4-v3
status:
  observedGeneration: 2$4213139756412538772
  phase: Ready
```

To get JSON of an object, use `--output=json` flag.

```bash
kubectl get mongodb mongodb-demo --output=json
```

To list all KubeDB objects, use following command:

```bash
$ kubectl get kubedb -o wide
NAME                VERSION     STATUS  AGE
mg/mongodb-demo     3.4         Ready   3h
mg/mongodb-dev      3.4         Ready   3h
mg/mongodb-prod     3.4         Ready   3h
mg/mongodb-qa       3.4         Ready   3h

NAME                                DATABASE                BUCKET              STATUS      AGE
snap/mongodb-demo-20170605-073557   mg/mongodb-demo         gs:bucket-name      Succeeded   9m
snap/snapshot-20171212-114700       mg/mongodb-demo         gs:bucket-name      Succeeded   1h
```

Flag `--output=wide` is used to print additional information.

List command supports short names for each object types. You can use it like `kubectl get <short-name>`. Below are the short name for KubeDB objects:

- MongoDB: `mg`
- Snapshot: `snap`
- DormantDatabase: `drmn`

You can print labels with objects. The following command will list all Snapshots with their corresponding labels.

```bash
$ kubectl get snap --show-labels
NAME                            DATABASE                STATUS      AGE       LABELS
mongodb-demo-20170605-073557    mg/mongodb-demo         Succeeded   11m       app.kubernetes.io/name=mongodbs.kubedb.com,app.kubernetes.io/instance=mongodb-demo
snapshot-20171212-114700        mg/mongodb-demo         Succeeded   1h        app.kubernetes.io/name=mongodbs.kubedb.com,app.kubernetes.io/instance=mongodb-demo
```

You can also filter list using `--selector` flag.

```bash
$ kubectl get snap --selector='app.kubernetes.io/name=mongodbs.kubedb.com' --show-labels
NAME                            DATABASE           STATUS      AGE       LABELS
mongodb-demo-20171212-073557    mg/mongodb-demo    Succeeded   14m       app.kubernetes.io/name=mongodbs.kubedb.com,app.kubernetes.io/instance=mongodb-demo
snapshot-20171212-114700        mg/mongodb-demo    Succeeded   2h        app.kubernetes.io/name=mongodbs.kubedb.com,app.kubernetes.io/instance=mongodb-demo
```

To print only object name, run the following command:

```bash
$ kubectl get all -o name
mongodb/mongodb-demo
mongodb/mongodb-dev
mongodb/mongodb-prod
mongodb/mongodb-qa
snapshot/mongodb-demo-20170605-073557
snapshot/snapshot-20170505-114700
```

### How to Describe Objects

`kubectl dba describe` command allows users to describe any KubeDB object. The following command will describe MongoDB database `mongodb-demo` with relevant information.

```bash
$ kubectl dba describe mg mongodb-demo
Name:               mongodb-demo
Namespace:          default
CreationTimestamp:  Wed, 06 Feb 2019 16:31:04 +0600
Labels:             <none>
Annotations:        <none>
Replicas:           1  total
Status:             Ready
  StorageType:      Durable
Volume:
  StorageClass:  standard
  Capacity:      1Gi
  Access Modes:  RWO

PetSet:
  Name:               mongodb-demo
  CreationTimestamp:  Wed, 06 Feb 2019 16:31:05 +0600
  Labels:               app.kubernetes.io/name=mongodbs.kubedb.com
                        app.kubernetes.io/instance=mongodb-demo
  Annotations:        <none>
  Replicas:           824639727120 desired | 1 total
  Pods Status:        1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:
  Name:         mongodb-demo
  Labels:         app.kubernetes.io/name=mongodbs.kubedb.com
                  app.kubernetes.io/instance=mongodb-demo
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.245.200
  Port:         db  27017/TCP
  TargetPort:   db/TCP
  Endpoints:    172.17.0.8:27017

Service:
  Name:         mongodb-demo-gvr
  Labels:         app.kubernetes.io/name=mongodbs.kubedb.com
                  app.kubernetes.io/instance=mongodb-demo
  Annotations:    service.alpha.kubernetes.io/tolerate-unready-endpoints=true
  Type:         ClusterIP
  IP:           None
  Port:         db  27017/TCP
  TargetPort:   27017/TCP
  Endpoints:    172.17.0.8:27017

Database Secret:
  Name:         mongodb-demo-auth
  Labels:         app.kubernetes.io/name=mongodbs.kubedb.com
                  app.kubernetes.io/instance=mongodb-demo
  Annotations:  <none>

Type:  Opaque

Data
====
  password:  16 bytes
  username:  4 bytes

No Snapshots.

Events:
  Type    Reason      Age   From             Message
  ----    ------      ----  ----             -------
  Normal  Successful  2m    KubeDB operator  Successfully created Service
  Normal  Successful  2m    KubeDB operator  Successfully created PetSet
  Normal  Successful  2m    KubeDB operator  Successfully created MongoDB
  Normal  Successful  2m    KubeDB operator  Successfully created appbinding
  Normal  Successful  2m    KubeDB operator  Successfully patched PetSet
  Normal  Successful  2m    KubeDB operator  Successfully patched MongoDB
```

`kubectl dba describe` command provides following basic information about a MongoDB database.

- PetSet
- Storage (Persistent Volume)
- Service
- Secret (If available)
- Snapshots (If any)
- Monitoring system (If available)

To hide events on KubeDB object, use flag `--show-events=false`

To describe all MongoDB objects in `default` namespace, use following command

```bash
kubectl dba describe mg
```

To describe all MongoDB objects from every namespace, provide `--all-namespaces` flag.

```bash
kubectl dba describe mg --all-namespaces
```

To describe all KubeDB objects from every namespace, use the following command:

```bash
kubectl dba describe all --all-namespaces
```

You can also describe KubeDb objects with matching labels. The following command will describe all MongoDB objects with specified labels from every namespace.

```bash
kubectl dba describe mg --all-namespaces --selector='group=dev'
```

To learn about various options of `describe` command, please visit [here](/docs/reference/cli/kubectl-dba_describe.md).

#### Edit Restrictions

Various fields of a KubeDb object can't be edited using `edit` command. The following fields are restricted from updates for all KubeDB objects:

- apiVersion
- kind
- metadata.name
- metadata.namespace

If PetSets exists for a MongoDB database, following fields can't be modified as well.

- spec.ReplicaSet
- spec.authSecret
- spec.init
- spec.storageType
- spec.storage
- spec.podTemplate.spec.nodeSelector

For DormantDatabase, `spec.origin` can't be edited using `kubectl edit`

### How to Delete Objects

`kubectl delete` command will delete an object in `default` namespace by default unless namespace is provided. The following command will delete a MongoDB `mongodb-dev` in default namespace

```bash
$ kubectl delete mongodb mongodb-dev
mongodb.kubedb.com "mongodb-dev" deleted
```

You can also use YAML files to delete objects. The following command will delete a mongodb using the type and name specified in `mongodb.yaml`.

```bash
$ kubectl delete -f mongodb-demo.yaml
mongodb.kubedb.com "mongodb-dev" deleted
```

`kubectl delete` command also takes input from `stdin`.

```bash
cat mongodb-demo.yaml | kubectl delete -f -
```

To delete database with matching labels, use `--selector` flag. The following command will delete mongodb with label `mongodb.app.kubernetes.io/instance=mongodb-demo`.

```bash
kubectl delete mongodb -l mongodb.app.kubernetes.io/instance=mongodb-demo
```

## Using Kubectl

You can use Kubectl with KubeDB objects like any other CRDs. Below are some common examples of using Kubectl with KubeDB objects.

```bash
# Create objects
$ kubectl create -f

# List objects
$ kubectl get mongodb
$ kubectl get mongodb.kubedb.com

# Delete objects
$ kubectl delete mongodb <name>
```

## Next Steps

- Learn how to use KubeDB to run a MongoDB database [here](/docs/guides/mongodb/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
