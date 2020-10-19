---
title: CLI | KubeDB
menu:
  docs_{{ .version }}:
    identifier: my-cli-cli
    name: Quickstart
    parent: my-cli-mysql
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Manage KubeDB objects using CLIs

## KubeDB CLI

KubeDB comes with its own cli. It is called `kubedb` cli. `kubedb` can be used to manage any KubeDB object. `kubedb` cli also performs various validations to improve ux. To install KubeDB cli on your workstation, follow the steps [here](/docs/setup/README.md).

### How to Create objects

`kubectl create` creates a database CRD object in `default` namespace by default. Following command will create a MySQL object as specified in `mysql.yaml`.

```console
$ kubectl create -f mysql-demo.yaml
mysql.kubedb.com/mysql-demo created
```

You can provide namespace as a flag `--namespace`. Provided namespace should match with namespace specified in input file.

```console
$ kubectl create -f mysql-demo.yaml --namespace=kube-system
mysql.kubedb.com/mysql-demo created
```

`kubectl create` command also considers `stdin` as input.

```console
cat mysql-demo.yaml | kubectl create -f -
```

### How to List Objects

`kubectl get` command allows users to list or find any KubeDB object. To list all MySQL objects in `default` namespace, run the following command:

```console
$ kubectl get mysql
NAME         VERSION   STATUS    AGE
mysql-demo   8.0-v2    Running   2m
mysql-dev    8.0-v2    Running   1m
mysql-prod   8.0-v2    Running   1m
mysql-qa     8.0-v2    Running   1m
```

To get YAML of an object, use `--output=yaml` flag.

```yaml
$ kubectl get mysql mysql-demo --output=yaml
apiVersion: kubedb.com/v1alpha1
kind: MySQL
metadata:
  creationTimestamp: 2018-09-27T13:07:23Z
  finalizers:
  - kubedb.com
  generation: 2
  name: mysql-demo
  namespace: default
  resourceVersion: "19279"
  selfLink: /apis/kubedb.com/v1alpha1/namespaces/default/mysqls/mysql-demo
  uid: 46034ac3-c256-11e8-b2cc-080027d9f35e
spec:
  databaseSecret:
    secretName: mysql-demo-auth
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      resources: {}
  replicas: 1
  serviceTemplate:
    metadata: {}
    spec: {}
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  terminationPolicy: Pause
  updateStrategy:
    type: RollingUpdate
  version: 8.0-v2
status:
  observedGeneration: 2$4213139756412538772
  phase: Running
```

To get JSON of an object, use `--output=json` flag.

```console
kubectl get mysql mysql-demo --output=json
```

To list all KubeDB objects, use following command:

```console
$ kubectl get all -o wide
NAME                          VERSION   STATUS    AGE
mysql.kubedb.com/mysql-demo   8.0-v2    Running   3m
mysql.kubedb.com/mysql-dev    8.0-v2    Running   2m
mysql.kubedb.com/mysql-prod   8.0-v2    Running   2m
mysql.kubedb.com/mysql-qa     8.0-v2    Running   2m

NAME                                DATABASE              BUCKET              STATUS      AGE
snap/mysql-demo-20170605-073557     my/mysql-demo         gs:bucket-name      Succeeded   9m
snap/snapshot-20171212-114700       my/mysql-demo         gs:bucket-name      Succeeded   1h
```

Flag `--output=wide` is used to print additional information.

List command supports short names for each object types. You can use it like `kubectl get <short-name>`. Below are the short name for KubeDB objects:

- MySQL: `my`
- Snapshot: `snap`
- DormantDatabase: `drmn`

You can print labels with objects. The following command will list all Snapshots with their corresponding labels.

```console
$ kubectl get snap --show-labels
NAME                          DATABASE              STATUS      AGE       LABELS
mysql-demo-20170605-073557    my/mysql-demo         Succeeded   11m       kubedb.com/kind=MySQL,kubedb.com/name=mysql-demo
snapshot-20171212-114700      my/mysql-demo         Succeeded   1h        kubedb.com/kind=MySQL,kubedb.com/name=mysql-demo
```

You can also filter list using `--selector` flag.

```console
$ kubectl get snap --selector='kubedb.com/kind=MySQL' --show-labels
NAME                          DATABASE         STATUS      AGE       LABELS
mysql-demo-20171212-073557    my/mysql-demo    Succeeded   14m       kubedb.com/kind=MySQL,kubedb.com/name=mysql-demo
snapshot-20171212-114700      my/mysql-demo    Succeeded   2h        kubedb.com/kind=MySQL,kubedb.com/name=mysql-demo
```

To print only object name, run the following command:

```console
$ kubectl get all -o name
mysql/mysql-demo
mysql/mysql-dev
mysql/mysql-prod
mysql/mysql-qa
snapshot/mysql-demo-20170605-073557
snapshot/snapshot-20170505-114700
```

### How to Describe Objects

`kubectl dba describe` command allows users to describe any KubeDB object. The following command will describe MySQL database `mysql-demo` with relevant information.

```console
$ kubectl dba describe my mysql-demo
Name:               mysql-demo
Namespace:          default
CreationTimestamp:  Thu, 27 Sep 2018 19:07:23 +0600
Labels:             <none>
Annotations:        <none>
Replicas:           1  total
Status:             Running
  StorageType:      Durable
Volume:
  StorageClass:  standard
  Capacity:      1Gi
  Access Modes:  RWO

StatefulSet:
  Name:               mysql-demo
  CreationTimestamp:  Thu, 27 Sep 2018 19:07:25 +0600
  Labels:               kubedb.com/kind=MySQL
                        kubedb.com/name=mysql-demo
  Annotations:        <none>
  Replicas:           824638226772 desired | 1 total
  Pods Status:        1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:
  Name:         mysql-demo
  Labels:         kubedb.com/kind=MySQL
                  kubedb.com/name=mysql-demo
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.102.105.123
  Port:         db  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    172.17.0.5:3306

Database Secret:
  Name:         mysql-demo-auth
  Labels:         kubedb.com/kind=MySQL
                  kubedb.com/name=mysql-demo
  Annotations:  <none>

Type:  Opaque

Data
====
  password:  16 bytes
  user:      4 bytes

No Snapshots.

Events:
  Type    Reason      Age   From            Message
  ----    ------      ----  ----            -------
  Normal  Successful  4m    MySQL operator  Successfully created Service
  Normal  Successful  4m    MySQL operator  Successfully created StatefulSet
  Normal  Successful  4m    MySQL operator  Successfully created MySQL
  Normal  Successful  4m    MySQL operator  Successfully patched StatefulSet
  Normal  Successful  4m    MySQL operator  Successfully patched MySQL
  Normal  Successful  3m    MySQL operator  Successfully patched StatefulSet
  Normal  Successful  3m    MySQL operator  Successfully patched MySQL
```

`kubectl dba describe` command provides following basic information about a MySQL database.

- StatefulSet
- Storage (Persistent Volume)
- Service
- Secret (If available)
- Snapshots (If any)
- Monitoring system (If available)

To hide events on KubeDB object, use flag `--show-events=false`

To describe all MySQL objects in `default` namespace, use following command

```console
kubectl dba describe my
```

To describe all MySQL objects from every namespace, provide `--all-namespaces` flag.

```console
kubectl dba describe my --all-namespaces
```

To describe all KubeDB objects from every namespace, use the following command:

```console
kubectl dba describe all --all-namespaces
```

You can also describe KubeDB objects with matching labels. The following command will describe all MySQL objects with specified labels from every namespace.

```console
kubectl dba describe my --all-namespaces --selector='group=dev'
```

To learn about various options of `describe` command, please visit [here](/docs/reference/kubectl-dba_describe.md).

### How to Edit Objects

`kubectl edit` command allows users to directly edit any KubeDB object. It will open the editor defined by _KUBEDB_EDITOR_, or _EDITOR_ environment variables, or fall back to `nano`.

Let's edit an existing running MySQL object to setup [Scheduled Backup](/docs/guides/mysql/snapshot/scheduled-backup.md). The following command will open MySQL `mysql-demo` in editor.

```console
$ kubectl edit my mysql-demo

# Add following under Spec to configure periodic backups
# backupSchedule:
#   cronExpression: '@every 1m'
#   storageSecretName: my-snap-secret
#   gcs:
#     bucket: bucket-name

mysql "mysql-demo" edited
```

#### Edit Restrictions

Various fields of a KubeDB object can't be edited using `edit` command. The following fields are restricted from updates for all KubeDB objects:

- apiVersion
- kind
- metadata.name
- metadata.namespace

If StatefulSets exists for a MySQL database, following fields can't be modified as well.

- spec.databaseSecret
- spec.init
- spec.storageType
- spec.storage
- spec.podTemplate.spec.nodeSelector

For DormantDatabase, `spec.origin` can't be edited using `kubectl edit`

### How to Delete Objects

`kubectl delete` command will delete an object in `default` namespace by default unless namespace is provided. The following command will delete a MySQL `mysql-dev` in default namespace

```console
$ kubectl delete mysql mysql-dev
mysql.kubedb.com "mysql-dev" deleted
```

You can also use YAML files to delete objects. The following command will delete a mysql using the type and name specified in `mysql.yaml`.

```console
$ kubectl delete -f mysql-demo.yaml
mysql.kubedb.com "mysql-dev" deleted
```

`kubectl delete` command also takes input from `stdin`.

```console
cat mysql-demo.yaml | kubectl delete -f -
```

To delete database with matching labels, use `--selector` flag. The following command will delete mysql with label `mysql.kubedb.com/name=mysql-demo`.

```console
kubectl delete mysql -l mysql.kubedb.com/name=mysql-demo
```

## Using Kubectl

You can use Kubectl with KubeDB objects like any other CRDs. Below are some common examples of using Kubectl with KubeDB objects.

```console
# Create objects
$ kubectl create -f

# List objects
$ kubectl get mysql
$ kubectl get mysql.kubedb.com

# Delete objects
$ kubectl delete mysql <name>
```

## Next Steps

- Learn how to use KubeDB to run a MySQL database [here](/docs/guides/mysql/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
