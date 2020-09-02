---
title: MySQL Quickstart
menu:
  docs_{{ .version }}:
    identifier: my-quickstart-quickstart
    name: Overview
    parent: my-quickstart-mysql
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# MySQL QuickStart

This tutorial will show you how to use KubeDB to run a MySQL database.

<p align="center">
  <img alt="lifecycle"  src="/docs/images/mysql/mysql-lifecycle.png">
</p>

> Note: The yaml files used in this tutorial are stored in [docs/examples/mysql](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mysql) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) is required to run KubeDB. Check the available StorageClass in cluster.

  ```bash
  $ kubectl get storageclasses
  NAME                 PROVISIONER             RECLAIMPOLICY     VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
  standard (default)   rancher.io/local-path   Delete            WaitForFirstConsumer   false                  6h22m
  ```

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. This tutorial will also use a [phpMyAdmin](https://hub.docker.com/r/phpmyadmin/phpmyadmin/) deployment to connect and test MySQL database, once it is running. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created

  $ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mysql/quickstart/demo-1.yaml
  deployment.extensions/myadmin created
  service/myadmin created

  $ kubectl get pods -n demo --watch
  NAME                      READY     STATUS              RESTARTS   AGE
  myadmin-c4db4df95-8lk74   0/1       ContainerCreating   0          27s
  myadmin-c4db4df95-8lk74   1/1       Running             0          1m

  $ kubectl get svc -n demo
  NAME      TYPE           CLUSTER-IP     EXTERNAL-IP   PORT(S)        AGE
  myadmin   LoadBalancer   10.105.73.16   <pending>     80:30158/TCP   23m
  ```

  Then, open your browser and go to the following URL: _http://{node-ip}:{myadmin-svc-nodeport}_. For kind cluster, you can get this URL by running the following command:

  ```bash
  $ kubectl get svc -n demo myadmin -o json | jq '.spec.ports[].nodePort'
  30158
  
  $ kubectl get node -o json | jq '.items[].status.addresses[].address'
  "172.18.0.3"
  "kind-control-plane"
  "172.18.0.4"
  "kind-worker"
  "172.18.0.2"
  "kind-worker2"
  
  # expected url will be:
  url: http://172.18.0.4:30158
  ```

According to the above example, this URL will be [ http://172.18.0.4:30158]( http://172.18.0.4:30158). The login informations to phpMyAdmin _(host, username and password)_ will be retrieved later in this tutorial.

## Find Available MySQLVersion

When you have installed KubeDB, it has created `MySQLVersion` crd for all supported MySQL versions. Check it by using the following command,

```bash
$ kubectl get mysqlversions
NAME        VERSION   DB_IMAGE                 DEPRECATED   AGE
5           5         kubedb/mysql:5           true         5h36m
5-v1        5         kubedb/mysql:5-v1        true         5h36m
5.7         5.7       kubedb/mysql:5.7         true         5h36m
5.7-v1      5.7       kubedb/mysql:5.7-v1      true         5h36m
5.7-v2      5.7.25    kubedb/mysql:5.7-v2      true         5h36m
5.7-v3      5.7.25    kubedb/mysql:5.7.25      true         5h36m
5.7-v4      5.7.29    kubedb/mysql:5.7.29      true         5h36m
5.7.25      5.7.25    kubedb/mysql:5.7.25      true         5h36m
5.7.25-v1   5.7.25    kubedb/mysql:5.7.25-v1                5h36m
5.7.29      5.7.29    kubedb/mysql:5.7.29                   5h36m
5.7.31      5.7.31    kubedb/mysql:5.7.31                   5h36m
8           8         kubedb/mysql:8           true         5h36m
8-v1        8         kubedb/mysql:8-v1        true         5h36m
8.0         8.0       kubedb/mysql:8.0         true         5h36m
8.0-v1      8.0.3     kubedb/mysql:8.0-v1      true         5h36m
8.0-v2      8.0.14    kubedb/mysql:8.0-v2      true         5h36m
8.0-v3      8.0.20    kubedb/mysql:8.0.20      true         5h36m
8.0.14      8.0.14    kubedb/mysql:8.0.14      true         5h36m
8.0.14-v1   8.0.14    kubedb/mysql:8.0.14-v1                5h36m
8.0.20      8.0.20    kubedb/mysql:8.0.20                   5h36m
8.0.21      8.0.21    kubedb/mysql:8.0.21                   5h36m
8.0.3       8.0.3     kubedb/mysql:8.0.3       true         5h36m
8.0.3-v1    8.0.3     kubedb/mysql:8.0.3-v1
```

## Create a MySQL database

KubeDB implements a `MySQL` CRD to define the specification of a MySQL database. Below is the `MySQL` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: mysql-quickstart
  namespace: demo
spec:
  version: "8.0.21"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: DoNotTerminate
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mysql/quickstart/demo-2.yaml
mysql.kubedb.com/mysql-quickstart created
```

Here,

- `spec.version` is the name of the MySQLVersion CRD where the docker images are specified. In this tutorial, a MySQL `8.0.21` database is going to create.
- `spec.storageType` specifies the type of storage that will be used for MySQL database. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create MySQL database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.terminationPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `MySQL` crd or which resources KubeDB should keep or delete when you delete `MySQL` crd. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.terminationPolicy` is set to `DoNotTerminate`. Learn details of all `TerminationPolicy` [here](/docs/concepts/databases/mysql.md#specterminationpolicy)

> Note: spec.storage section is used to create PVC for database pod. It will create PVC with storage size specified instorage.resources.requests field. Don't specify limits here. PVC does not get resized automatically.

KubeDB operator watches for `MySQL` objects using Kubernetes api. When a `MySQL` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching MySQL object name. KubeDB operator will also create a governing service for StatefulSets with the name `kubedb`, if one is not already present.

```console
$ kubectl dba describe my -n demo mysql-quickstart
Name:               mysql-quickstart
Namespace:          demo
CreationTimestamp:  Mon, 31 Aug 2020 16:39:47 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1alpha1","kind":"MySQL","metadata":{"annotations":{},"name":"mysql-quickstart","namespace":"demo"},"spec":{"storage":{"acces...
Replicas:           1  total
Status:             Running
StorageType:        Durable
Volume:
  StorageClass:      standard
  Capacity:          1Gi
  Access Modes:      RWO
Paused:              false
Halted:              false
Termination Policy:  DoNotTerminate

StatefulSet:          
  Name:               mysql-quickstart
  CreationTimestamp:  Mon, 31 Aug 2020 16:39:47 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=mysql-quickstart
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mysql
                        app.kubernetes.io/version=8.0.21
                        kubedb.com/kind=MySQL
                        kubedb.com/name=mysql-quickstart
  Annotations:        <none>
  Replicas:           824634389080 desired | 1 total
  Pods Status:        1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         mysql-quickstart
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mysql-quickstart
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysql
                  app.kubernetes.io/version=8.0.21
                  kubedb.com/kind=MySQL
                  kubedb.com/name=mysql-quickstart
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.103.57.226
  Port:         db  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.2.13:3306

Service:        
  Name:         mysql-quickstart-gvr
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mysql-quickstart
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysql
                  app.kubernetes.io/version=8.0.21
                  kubedb.com/kind=MySQL
                  kubedb.com/name=mysql-quickstart
  Annotations:    service.alpha.kubernetes.io/tolerate-unready-endpoints=true
  Type:         ClusterIP
  IP:           None
  Port:         db  3306/TCP
  TargetPort:   3306/TCP
  Endpoints:    10.244.2.13:3306

Database Secret:
  Name:         mysql-quickstart-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mysql-quickstart
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysql
                  app.kubernetes.io/version=8.0.21
                  kubedb.com/kind=MySQL
                  kubedb.com/name=mysql-quickstart
  Annotations:  <none>
  Type:         Opaque
  Data:
    password:  16 bytes
    username:  4 bytes

AppBinding:
  Metadata:
    Annotations:
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1alpha1","kind":"MySQL","metadata":{"annotations":{},"name":"mysql-quickstart","namespace":"demo"},"spec":{"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","terminationPolicy":"DoNotTerminate","version":"8.0.21"}}

    Creation Timestamp:  2020-08-31T10:40:53Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    mysql-quickstart
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        mysql
      app.kubernetes.io/version:     8.0.21
      kubedb.com/kind:               MySQL
      kubedb.com/name:               mysql-quickstart
    Name:                            mysql-quickstart
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    mysql-quickstart
        Path:    /
        Port:    3306
        Scheme:  mysql
      URL:       tcp(mysql-quickstart:3306)/
    Secret:
      Name:   mysql-quickstart-auth
    Type:     kubedb.com/mysql
    Version:  8.0.21

Events:
  Type    Reason      Age   From            Message
  ----    ------      ----  ----            -------
  Normal  Successful  3m    MySQL operator  Successfully created Service
  Normal  Successful  2m    MySQL operator  Successfully created StatefulSet
  Normal  Successful  2m    MySQL operator  Successfully created MySQL
  Normal  Successful  2m    MySQL operator  Successfully created appbinding


$ kubectl get statefulset -n demo
NAME               READY   AGE
mysql-quickstart   1/1     2m22s

$ kubectl get pvc -n demo
NAME                      STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-mysql-quickstart-0   Bound     pvc-652e02c7-0d7f-11e8-9091-08002751ae8c   1Gi        RWO            standard       10m

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS    CLAIM                          STORAGECLASS   REASON    AGE
pvc-652e02c7-0d7f-11e8-9091-08002751ae8c   1Gi        RWO            Delete           Bound     demo/data-mysql-quickstart-0   standard                 11m

$ kubectl get service -n demo
NAME                    TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE
myadmin                 LoadBalancer   10.104.142.213   <pending>     80:31529/TCP   6h2m
mysql-quickstart        ClusterIP      10.109.217.165   <none>        3306/TCP       5m56s
mysql-quickstart-gvr    ClusterIP      None             <none>        3306/TCP       5m56s
```

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created. Run the following command to see the modified MySQL object:

```yaml
$ kubectl get my -n demo mysql-quickstart -o yaml
 $ kubectl get my -n demo mysql-quickstart -o yaml
apiVersion: kubedb.com/v1alpha1
kind: MySQL
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha1","kind":"MySQL","metadata":{"annotations":{},"name":"mysql-quickstart","namespace":"demo"},"spec":{"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","terminationPolicy":"DoNotTerminate","version":"8.0.21"}}
  creationTimestamp: "2020-08-27T12:19:42Z"
  finalizers:
  - kubedb.com
  ...
  name: mysql-quickstart
  namespace: demo
  resourceVersion: "70812"
  selfLink: /apis/kubedb.com/v1alpha1/namespaces/demo/mysqls/mysql-quickstart
  uid: 837ac85a-134a-457e-b126-f4681d92f117
spec:
  databaseSecret:
    secretName: mysql-quickstart-auth
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      resources: {}
      serviceAccountName: mysql-quickstart
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
  terminationPolicy: DoNotTerminate
  updateStrategy:
    type: RollingUpdate
  version: 8.0.21
status:
  observedGeneration: 2
  phase: Running
```

## Connect with MySQL database

KubeDB operator has created a new Secret called `mysql-quickstart-auth` *(format: {mysql-object-name}-auth)* for storing the password for `mysql` superuser. This secret contains a `username` key which contains the *username* for MySQL superuser and a `password` key which contains the *password* for MySQL superuser.

If you want to use an existing secret please specify that when creating the MySQL object using `spec.databaseSecret.secretName`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `root` as value of `username`. For more details see [here](/docs/concepts/databases/mysql.md#specdatabasesecret).

Now, you can connect to this database from the phpMyAdmin dashboard using the database pod IP and and `mysql` user password.

```bash
$ kubectl get pods mysql-quickstart-0 -n demo -o yaml | grep podIP
  podIP: 10.244.2.13

$ kubectl get secrets -n demo mysql-quickstart-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mysql-quickstart-auth -o jsonpath='{.data.\password}' | base64 -d
l0yKjI1E7IMohsGR
```

---
>Note: In MySQL: `8.0.14` connection to phpMyAdmin may give error as it is using `caching_sha2_password` and `sha256_password` authentication plugins over `mysql_native_password`. If the error happens do the following for work around. But, It's not recommended to change authentication plugins. See [here](https://stackoverflow.com/questions/49948350/phpmyadmin-on-mysql-8-0) for alternative solutions.

```bash
kubectl exec -it -n demo mysql-quickstart-0 -- mysql -u root --password=l0yKjI1E7IMohsGR -e "ALTER USER root IDENTIFIED WITH mysql_native_password BY 'l0yKjI1E7IMohsGR';"
```
---

Now, open your browser and go to the following URL: _http://{node-ip}:{myadmin-svc-nodeport}_. To log into the phpMyAdmin, use host __`mysql-quickstart.demo`__ or __`10.244.2.13`__ , username __`root`__ and password __`l0yKjI1E7IMohsGR`__.

## Database TerminationPolicy

This field is used to regulate the deletion process of the related resources when `MySQL` object is deleted. User can set the value of this field according to their needs. The available options and their use case scenario is described below:

**DoNotTerminate:**

When `terminationPolicy` is set to `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.terminationPolicy` is set to `DoNotTerminate`. You can see this below:

```bash
$ kubectl delete my mysql-quickstart -n demo
Error from server (BadRequest): admission webhook "mysql.validators.kubedb.com" denied the request: mysql "mysql-quickstart" can't be paused. To delete, change spec.terminationPolicy
```

Now, run `kubectl edit my mysql-quickstart -n demo` to set `spec.terminationPolicy` to `Halt` (which deletes the mysql object and keeps PVC, snapshots, Secrets intact) or remove this field (which default to `Delete`). Then you will be able to delete/pause the database.

Learn details of all `TerminationPolicy` [here](/docs/concepts/databases/mysql.md#specterminationpolicy).

**Halt:**

Suppose you want to reuse your database volume and credential to deploy your database in future using the same configurations. But, right now you just want to delete the database except the database volumes and credentials. In this scenario, you must set the `MySQL` object `terminationPolicy` to `Halt`.

When the [TerminationPolicy](/docs/concepts/databases/mysql.md#specterminationpolicy) is set to `halt` and the MySQL object is deleted, the KubeDB operator will delete the StatefulSet and its pods but leaves the `PVCs`, `secrets` and database backup data(`snapshots`) intact. You can set the `terminationPolicy` to `halt` in existing database using `edit` command for testing.

At first, run `kubectl edit my mysql-quickstart -n demo` to set `spec.terminationPolicy` to `Halt`. Then delete the mysql object,

```bash
$ kubectl delete my mysql-quickstart -n demo
mysql.kubedb.com "mysql-quickstart" deleted
```

Now, run the following command to get all mysql resources in `demo` namespaces,

```bash
$ kubectl get sts,svc,secret,pvc -n demo
NAME                           TYPE                                  DATA   AGE
secret/default-token-lgbjm     kubernetes.io/service-account-token   3      23h
secret/mysql-quickstart-auth   Opaque                                2      20h

NAME                                            STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-mysql-quickstart-0   Bound    pvc-716f627c-9aa2-47b6-aa64-a547aab6f55c   1Gi        RWO            standard       20h
```

From the above output, you can see that all mysql resources(`StatefulSet`, `Service`, etc.) are deleted except `PVC` and `Secret`. You can recreate your mysql again using this resources.

>You can also set the `terminationPolicy` to `Pause`(deprecated). It's behavior same as `halt` and right now `Pause` is replaced by `Halt`.

**Delete:**

If you want to delete the existing database along with the volumes used, but want to restore the database from previously taken `snapshots` and `secrets` then you might want to set the `MySQL` object `terminationPolicy` to `Delete`. In this setting, `StatefulSet` and the volumes will be deleted. If you decide to restore the database, you can do so using the snapshots and the credentials.

When the [TerminationPolicy](/docs/concepts/databases/mysql.md#specterminationpolicy) is set to `Delete` and the MySQL object is deleted, the KubeDB operator will delete the StatefulSet and its pods along with PVCs but leaves the `secret` and database backup data(`snapshot`) intact.

Suppose, we have a database with `terminationPolicy` set to `Delete`. Now, are going to delete the database using the following command:

```bash
$ kubectl delete my mysql-quickstart -n demo
mysql.kubedb.com "mysql-quickstart" deleted
```

Now, run the following command to get all mysql resources in `demo` namespaces,

```bash
$ kubectl get sts,svc,secret,pvc -n demo
NAME                           TYPE                                  DATA   AGE
secret/default-token-lgbjm     kubernetes.io/service-account-token   3      24h
secret/mysql-quickstart-auth   Opaque
```

From the above output, you can see that all mysql resources(`StatefulSet`, `Service`, `PVCs` etc.) are deleted except `Secret`. You can initialize your mysql using `snapshots`(if previously taken) and `secret`.

>If you don't set the terminationPolicy then the kubeDB set the TerminationPolicy to Delete by-default.

**WipeOut:**

You can totally delete the `MySQL` database and relevant resources without any tracking by setting `terminationPolicy` to `WipeOut`. KubeDB operator will delete all relevant resources of this `MySQL` database (i.e, `PVCs`, `Secrets`, `Snapshots`) when the `terminationPolicy` is set to `WipeOut`.

Suppose, we have a database with `terminationPolicy` set to `WipeOut`. Now, are going to delete the database using the following command:

```yaml
$ kubectl delete my mysql-quickstart -n demo
mysql.kubedb.com "mysql-quickstart" deleted
```

Now, run the following command to get all mysql resources in `demo` namespaces,

```bash
$ kubectl get sts,svc,secret,pvc -n demo
No resources found in demo namespace.
```

From the above output, you can see that all mysql resources are deleted. there is no option to recreate/reinitialize your database if `terminationPolicy` is set to `Delete`.

>Be careful when you set the `terminationPolicy` to `Delete`. Because there is no option to trace the database resources if once deleted the database.

## Database Halted

If you want to delete MySQL resources(`StatefulSet`,`Service`, etc.) without deleting the `MySQL` object, `PVCs` and `Secret` you have to set the `spec.halted` to `true`. KubeDB operator will be able to delete the MySQL related resources except the `MySQL` object, `PVCs` and `Secret`.

Suppose we have a database running `mysql-quickstart` in our cluster. Now, we are going to set `spec.halted` to `true` in `MySQL`  object by running `kubectl edit -n demo mysql-quickstart` command.

Run the following command to get MySQL resources,

```bash
$ kubectl get my,sts,secret,svc,pvc -n demo
NAME                                VERSION   STATUS   AGE
mysql.kubedb.com/mysql-quickstart   8.0.21    Halted   22m

NAME                           TYPE                                  DATA   AGE
secret/default-token-lgbjm     kubernetes.io/service-account-token   3      27h
secret/mysql-quickstart-auth   Opaque                                2      22m

NAME                                            STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-mysql-quickstart-0   Bound    pvc-7ab0ebb0-bb2e-45c1-9af1-4f175672605b   1Gi        RWO            standard       22m
```

From the above output , you can see that `MySQL` object, `PVCs`, `Secret` are still alive. Then you can recreate your `MySQL` with same configuration.

>When you set `spec.halted` to `true` in `MySQL` object then the `terminationPolicy` is also set to `Halt` by KubeDB operator.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mysql/mysql-quickstart -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mysql/mysql-quickstart

kubectl delete ns demo
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for production environment. You can follow these tips to avoid them.

1. **Use `storageType: Ephemeral`**. Databases are precious. You might not want to lose your data in your production environment if database pod fail. So, we recommend to use `spec.storageType: Durable` and provide storage spec in `spec.storage` section. For testing purpose, you can just use `spec.storageType: Ephemeral`. KubeDB will use [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) for storage. You will not require to provide `spec.storage` section.
2. **Use `terminationPolicy: WipeOut`**. It is nice to be able to delete everything created by KubeDB for a particular MySQL crd when you delete the crd. For more details about termination policy, please visit [here](/docs/concepts/databases/mysql.md#specterminationpolicy).

## Next Steps

- [Snapshot and Restore](/docs/guides/mysql/snapshot/backup-and-restore.md) process of MySQL databases using KubeDB.
- Take [Scheduled Snapshot](/docs/guides/mysql/snapshot/scheduled-backup.md) of MySQL databases using KubeDB.
- Initialize [MySQL with Script](/docs/guides/mysql/initialization/using-script.md).
- Initialize [MySQL with Snapshot](/docs/guides/mysql/initialization/using-snapshot.md).
- Monitor your MySQL database with KubeDB using [out-of-the-box CoreOS Prometheus Operator](/docs/guides/mysql/monitoring/using-coreos-prometheus-operator.md).
- Monitor your MySQL database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mysql/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mysql/private-registry/using-private-registry.md) to deploy MySQL with KubeDB.
- Detail concepts of [MySQL object](/docs/concepts/databases/mysql.md).
- Detail concepts of [MySQLVersion object](/docs/concepts/catalog/mysql.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
