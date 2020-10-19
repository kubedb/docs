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

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mysql/cli/mysql-demo.yaml
mysql.kubedb.com/mysql-demo created
```

You can provide namespace as a flag `--namespace`. Provided namespace should match with namespace specified in input file.

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mysql/cli/mysql-demo.yaml --namespace=kube-system
mysql.kubedb.com/mysql-demo created
```

`kubectl create` command also considers `stdin` as input.

```bash
cat mysql-demo.yaml | kubectl create -f -
```

### How to List Objects

`kubectl get` command allows users to list or find any KubeDB object. To list all MySQL objects in `default` namespace, run the following command:

```bash
$ kubectl get mysql
NAME         VERSION   STATUS    AGE
mysql-demo   8.0.21    Running   5m1s
mysql-dev    5.7.31    Running   10m1s
mysql-prod   8.0.20    Running   20m1s
```

To get YAML of an object, use `--output=yaml` flag.

```yaml
$ kubectl get mysql mysql-demo --output=yaml
apiVersion: kubedb.com/v1alpha1
kind: MySQL
metadata:
  creationTimestamp: "2020-08-25T11:21:29Z"
  finalizers:
  - kubedb.com
  ...
  generation: 2
    operation: Update
    time: "2020-08-25T11:22:40Z"
  name: mysql-demo
  namespace: default
  resourceVersion: "8763"
  selfLink: /apis/kubedb.com/v1alpha1/namespaces/default/mysqls/mysql-demo
  uid: daac5549-0a7b-4e25-8773-473dffabf1cd
spec:
  databaseSecret:
    secretName: mysql-demo-auth
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      resources: {}
      serviceAccountName: mysql-demo
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
  terminationPolicy: Delete
  updateStrategy:
    type: RollingUpdate
  version: 8.0.21
status:
  observedGeneration: 2
  phase: Running
```

To get JSON of an object, use `--output=json` flag.

```console
kubectl get mysql mysql-demo --output=json
```

To list all KubeDB objects, use following command:

```bash
$ kubectl get all -o wide
NAME               READY   STATUS    RESTARTS   AGE   IP            NODE          NOMINATED NODE   READINESS GATES
pod/mysql-demo-0   1/1     Running   0          21m   10.244.1.10   kind-worker   <none>           <none>

NAME                     TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE   SELECTOR
service/kubernetes       ClusterIP   10.96.0.1       <none>        443/TCP    64m   <none>
service/mysql-demo       ClusterIP   10.109.208.91   <none>        3306/TCP   21m   kubedb.com/kind=MySQL,kubedb.com/name=mysql-demo
service/mysql-demo-gvr   ClusterIP   None            <none>        3306/TCP   21m   kubedb.com/kind=MySQL,kubedb.com/name=mysql-demo

NAME                          READY   AGE   CONTAINERS   IMAGES
statefulset.apps/mysql-demo   1/1     21m   mysql        kubedb/mysql:8.0.21

NAME                                            TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/mysql-demo   kubedb.com/mysql   8.0.21    20m

NAME                          VERSION   STATUS    AGE
mysql.kubedb.com/mysql-demo   8.0.21    Running   21m
```

Flag `--output=wide` is used to print additional information.

List command supports short names for each object types. You can use it like `kubectl get <short-name>`. Below are the short name for KubeDB objects:

- MySQL: `my`

To print only object name, run the following command:

```console
$ kubectl get all -o name
mysql/mysql-demo
mysql/mysql-dev
mysql/mysql-prod
mysql/mysql-qa
```

### How to Describe Objects

`kubectl dba describe` command allows users to describe any KubeDB object. The following command will describe MySQL database `mysql-demo` with relevant information.

```bash
$ kubectl dba describe my mysql-demo
Name:               mysql-demo
Namespace:          default
CreationTimestamp:  Tue, 25 Aug 2020 17:21:29 +0600
Labels:             <none>
Annotations:        <none>
Replicas:           1  total
Status:             Running
StorageType:        Durable
Volume:
  StorageClass:      standard
  Capacity:          1Gi
  Access Modes:      RWO
Paused:              false
Halted:              false
Termination Policy:  Delete

StatefulSet:          
  Name:               mysql-demo
  CreationTimestamp:  Tue, 25 Aug 2020 17:21:29 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=mysql-demo
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mysql
                        app.kubernetes.io/version=8.0.21
                        kubedb.com/kind=MySQL
                        kubedb.com/name=mysql-demo
  Annotations:        <none>
  Replicas:           824635270088 desired | 1 total
  Pods Status:        1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         mysql-demo
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mysql-demo
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysql
                  app.kubernetes.io/version=8.0.21
                  kubedb.com/kind=MySQL
                  kubedb.com/name=mysql-demo
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.109.208.91
  Port:         db  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.1.10:3306

Service:        
  Name:         mysql-demo-gvr
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mysql-demo
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysql
                  app.kubernetes.io/version=8.0.21
                  kubedb.com/kind=MySQL
                  kubedb.com/name=mysql-demo
  Annotations:    service.alpha.kubernetes.io/tolerate-unready-endpoints=true
  Type:         ClusterIP
  IP:           None
  Port:         db  3306/TCP
  TargetPort:   3306/TCP
  Endpoints:    10.244.1.10:3306

Database Secret:
  Name:         mysql-demo-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mysql-demo
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysql
                  app.kubernetes.io/version=8.0.21
                  kubedb.com/kind=MySQL
                  kubedb.com/name=mysql-demo
  Annotations:  <none>
  Type:         Opaque
  Data:
    password:  16 bytes
    username:  4 bytes

AppBinding:
  Metadata:
    Creation Timestamp:  2020-08-25T11:22:39Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    mysql-demo
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        mysql
      app.kubernetes.io/version:     8.0.21
      kubedb.com/kind:               MySQL
      kubedb.com/name:               mysql-demo
    Name:                            mysql-demo
    Namespace:                       default
  Spec:
    Client Config:
      Service:
        Name:    mysql-demo
        Path:    /
        Port:    3306
        Scheme:  mysql
      URL:       tcp(mysql-demo:3306)/
    Secret:
      Name:   mysql-demo-auth
    Type:     kubedb.com/mysql
    Version:  8.0.21

Events:
  Type    Reason      Age   From            Message
  ----    ------      ----  ----            -------
  Normal  Successful  27m   MySQL operator  Successfully created Service
  Normal  Successful  26m   MySQL operator  Successfully created StatefulSet
  Normal  Successful  26m   MySQL operator  Successfully created MySQL
  Normal  Successful  26m   MySQL operator  Successfully created appbinding
```

`kubectl dba describe` command provides following basic information about a MySQL database.

- StatefulSet
- Storage (Persistent Volume)
- Service
- Secret (If available)
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

To learn about various options of `describe` command, please visit [here](/docs/examples/mysql/cli/kubectl-dba_describe.md).

### How to Edit Objects

`kubectl edit` command allows users to directly edit any KubeDB object. It will open the editor defined by _KUBEDB_EDITOR_, or _EDITOR_ environment variables, or fall back to `nano`.

Lets edit an existing running MySQL object to setup database [Halted](/docs/concepts/databases/mysql.md#spechalted). The following command will open MySQL `mysql-demo` in editor.

```console
$ kubectl edit my -n demo mysql-quickstart

spec:
  ....
  databaseSecret:
    secretName: mysql-quickstart-auth
# add database halted = true to delete StatefulSet services and database other resources
  halted: true
  ....

mysql.kubedb.com/mysql-quickstart edited
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
