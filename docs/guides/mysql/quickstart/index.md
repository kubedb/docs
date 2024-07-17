---
title: MySQL Quickstart
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-quickstart
    name: Quickstart
    parent: guides-mysql
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MySQL QuickStart

This tutorial will show you how to use KubeDB to run a MySQL database.

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/mysql/quickstart/images/mysql-lifecycle.png">
</p>

> Note: The yaml files used in this tutorial are stored in [docs/guides/mysql/quickstart/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/quickstart/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) is required to run KubeDB. Check the available StorageClass in cluster.

  ```bash
  $ kubectl get storageclasses
  NAME                 PROVISIONER             RECLAIMPOLICY     VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
  standard (default)   rancher.io/local-path   Delete            WaitForFirstConsumer   false                  6h22m
  ```

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Find Available MySQLVersion

When you have installed KubeDB, it has created `MySQLVersion` crd for all supported MySQL versions. Check it by using the following command,

```bash
$ kubectl get mysqlversions
NAME            VERSION   DISTRIBUTION   DB_IMAGE                    DEPRECATED   AGE
5.7.35-v1       5.7.35    Official       mysql:5.7.35                             9s
5.7.44          5.7.44    Official       mysql:5.7.44                             9s
8.0.17          8.0.17    Official       mysql:8.0.17                             9s
8.0.35          8.0.35    Official       mysql:8.0.35                             9s
8.0.31-innodb   8.0.35    MySQL          mysql/mysql-server:8.0.35                9s
8.0.35          8.0.35    Official       mysql:8.0.35                             9s
8.0.3-v4        8.0.3     Official       mysql:8.0.3                              9s

```

## Create a MySQL database

KubeDB implements a `MySQL` CRD to define the specification of a MySQL database. Below is the `MySQL` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-quickstart
  namespace: demo
spec:
  version: "8.0.35"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: Delete
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/quickstart/yamls/quickstart.yaml
mysql.kubedb.com/mysql-quickstart created
```

Here,

- `spec.version` is the name of the MySQLVersion CRD where the docker images are specified. In this tutorial, a MySQL `8.0.35` database is going to be created.
- `spec.storageType` specifies the type of storage that will be used for MySQL database. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create MySQL database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `MySQL` crd or which resources KubeDB should keep or delete when you delete `MySQL` crd. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`. Learn details of all `DeletionPolicy` [here](/docs/guides/mysql/concepts/database/index.md#specdeletionpolicy)

> Note: spec.storage section is used to create PVC for database pod. It will create PVC with storage size specified instorage.resources.requests field. Don't specify limits here. PVC does not get resized automatically.

KubeDB operator watches for `MySQL` objects using Kubernetes api. When a `MySQL` object is created, KubeDB operator will create a new PetSet and a Service with the matching MySQL object name. KubeDB operator will also create a governing service for PetSets with the name `kubedb`, if one is not already present.

```bash
$ kubectl dba describe my -n demo mysql-quickstart
Name:               mysql-quickstart
Namespace:          demo
CreationTimestamp:  Fri, 03 Jun 2022 12:50:40 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1","kind":"MySQL","metadata":{"annotations":{},"name":"mysql-quickstart","namespace":"demo"},"spec":{"storage":{"acces...
Replicas:           1  total
Status:             Ready
StorageType:        Durable
Volume:
  StorageClass:      standard
  Capacity:          1Gi
  Access Modes:      RWO
Paused:              false
Halted:              false
Termination Policy:  DoNotTerminate

PetSet:          
  Name:               mysql-quickstart
  CreationTimestamp:  Fri, 03 Jun 2022 12:50:40 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=mysql-quickstart
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:        <none>
  Replicas:           824646358808 desired | 1 total
  Pods Status:        1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         mysql-quickstart
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mysql-quickstart
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.150.194
  Port:         primary  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.30:3306

Service:        
  Name:         mysql-quickstart-pods
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mysql-quickstart
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           None
  Port:         db  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.30:3306

Auth Secret:
  Name:         mysql-quickstart-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mysql-quickstart
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         kubernetes.io/basic-auth
  Data:
    password:  16 bytes
    username:  4 bytes

AppBinding:
  Metadata:
    Annotations:
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1","kind":"MySQL","metadata":{"annotations":{},"name":"mysql-quickstart","namespace":"demo"},"spec":{"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","deletionPolicy":"DoNotTerminate","version":"8.0.35"}}

    Creation Timestamp:  2022-06-03T06:50:40Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    mysql-quickstart
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        mysqls.kubedb.com
    Name:                            mysql-quickstart
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    mysql-quickstart
        Path:    /
        Port:    3306
        Scheme:  mysql
      URL:       tcp(mysql-quickstart.demo.svc:3306)/
    Parameters:
      API Version:  appcatalog.appscode.com/v1alpha1
      Kind:         StashAddon
      Stash:
        Addon:
          Backup Task:
            Name:  mysql-backup-8.0.21
            Params:
              Name:   args
              Value:  --all-databases --set-gtid-purged=OFF
          Restore Task:
            Name:  mysql-restore-8.0.21
    Secret:
      Name:   mysql-quickstart-auth
    Type:     kubedb.com/mysql
    Version:  8.0.35

Events:
  Type     Reason      Age   From             Message
  ----     ------      ----  ----             -------
  Normal   Successful  32s   KubeDB Operator  Successfully created governing service
  Normal   Successful  32s   KubeDB Operator  Successfully created service for primary/standalone
  Normal   Successful  32s   KubeDB Operator  Successfully created database auth secret
  Normal   Successful  32s   KubeDB Operator  Successfully created PetSet
  Normal   Successful  32s   KubeDB Operator  Successfully created MySQL
  Normal   Successful  32s   KubeDB Operator  Successfully created appbinding



$ kubectl get petset -n demo
NAME               READY   AGE
mysql-quickstart   1/1     3m19s

$ kubectl get pvc -n demo
NAME                      STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-mysql-quickstart-0   Bound    pvc-ab44ce95-2300-47d7-8f25-3cd7bc5b0091   1Gi        RWO            standard       3m50s


$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS   REASON   AGE
pvc-ab44ce95-2300-47d7-8f25-3cd7bc5b0091   1Gi        RWO            Delete           Bound    demo/data-mysql-quickstart-0   standard                4m19s

kubectl get service -n demo
NAME                    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
mysql-quickstart        ClusterIP   10.96.150.194   <none>        3306/TCP   5m13s
mysql-quickstart-pods   ClusterIP   None            <none>        3306/TCP   5m13s
```

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created. Run the following command to see the modified MySQL object:

```yaml
$ kubectl get my -n demo mysql-quickstart -o yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1","kind":"MySQL","metadata":{"annotations":{},"name":"mysql-quickstart","namespace":"demo"},"spec":{"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","deletionPolicy":"DoNotTerminate","version":"8.0.35"}}
  creationTimestamp: "2022-06-03T06:50:40Z"
  finalizers:
  - kubedb.com
spec:
  allowedReadReplicas:
    namespaces:
      from: Same
  allowedSchemas:
    namespaces:
      from: Same
  authSecret:
    name: mysql-quickstart-auth
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/instance: mysql-quickstart
                  app.kubernetes.io/managed-by: kubedb.com
                  app.kubernetes.io/name: mysqls.kubedb.com
              namespaces:
              - demo
              topologyKey: kubernetes.io/hostname
            weight: 100
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/instance: mysql-quickstart
                  app.kubernetes.io/managed-by: kubedb.com
                  app.kubernetes.io/name: mysqls.kubedb.com
              namespaces:
              - demo
              topologyKey: failure-domain.beta.kubernetes.io/zone
            weight: 50
      resources:
        limits:
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      serviceAccountName: mysql-quickstart
  replicas: 1
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: Delete
  useAddressType: DNS
  version: 8.0.35
status:
  conditions:
  - lastTransitionTime: "2022-06-03T06:50:40Z"
    message: 'The KubeDB operator has started the provisioning of MySQL: demo/mysql-quickstart'
    reason: DatabaseProvisioningStartedSuccessfully
    status: "True"
    type: ProvisioningStarted
  - lastTransitionTime: "2022-06-03T06:50:46Z"
    message: All desired replicas are ready.
    reason: AllReplicasReady
    status: "True"
    type: ReplicaReady
  - lastTransitionTime: "2022-06-03T06:51:05Z"
    message: database demo/mysql-quickstart is accepting connection
    reason: AcceptingConnection
    status: "True"
    type: AcceptingConnection
  - lastTransitionTime: "2022-06-03T06:51:05Z"
    message: database demo/mysql-quickstart is ready
    reason: AllReplicasReady
    status: "True"
    type: Ready
  - lastTransitionTime: "2022-06-03T06:51:05Z"
    message: 'The MySQL: demo/mysql-quickstart is successfully provisioned.'
    observedGeneration: 2
    reason: DatabaseSuccessfullyProvisioned
    status: "True"
    type: Provisioned
  observedGeneration: 2
  phase: Ready

```

## Connect with MySQL database

KubeDB operator has created a new Secret called `mysql-quickstart-auth` *(format: {mysql-object-name}-auth)* for storing the password for `mysql` superuser. This secret contains a `username` key which contains the *username* for MySQL superuser and a `password` key which contains the *password* for MySQL superuser.

If you want to use an existing secret please specify that when creating the MySQL object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `root` as value of `username`. For more details see [here](/docs/guides/mysql/concepts/database/index.md#specdatabasesecret).

Now, we need `username` and `password` to connect to this database from `kubectl exec` command. In this example  `mysql-quickstart-auth` secret holds username and password

```bash
$ kubectl get pods mysql-quickstart-0 -n demo -o yaml | grep podIP
  podIP: 10.244.0.30

$ kubectl get secrets -n demo mysql-quickstart-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mysql-quickstart-auth -o jsonpath='{.data.\password}' | base64 -d
H(Y.s)pg&cX1Ds3J
```
we will exec into the pod `mysql-quickstart-0` and connect to the database using username and password

```bash
$ kubectl exec -it -n demo mysql-quickstart-0 -- bash

root@mysql-quickstart-0:/# mysql -uroot -p"H(Y.s)pg&cX1Ds3J"

Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 351
Server version: 8.0.35 MySQL Community Server - GPL

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

mysql> show databases;
+--------------------+
| Database           |
+--------------------+
| hello              |
| information_schema |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
5 rows in set (0.00 sec)

```
you can also connect with database management tools like [phpmyadmin](https://hub.docker.com/_/phpmyadmin), [dbgate](https://hub.docker.com/r/dbgate/dbgate).

__connecting with `phpmyadmin`__

lets create a deployment of `phpmyadmin`

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/quickstart/yamls/phpmyadmin.yaml

deployment/myadmin created
service/myadmin created

$ kubectl get pods -n demo --watch
NAME                       READY   STATUS    RESTARTS   AGE
myadmin-85d86cf5b5-f4mq4   1/1     Running   0          8s
mysql-quickstart-0         1/1     Running   0          12m


$ kubectl get svc -n demo
NAME                    TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
myadmin                 LoadBalancer   10.96.108.199   <pending>     80:32634/TCP   51s
mysql-quickstart        ClusterIP      10.96.150.194   <none>        3306/TCP       13m
mysql-quickstart-pods   ClusterIP      None            <none>        3306/TCP       13m


```
Lets, open your browser and go to the following URL: _http://{node-ip}:{myadmin-svc-nodeport}_. For kind cluster, you can get this URL by running the following command:

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
According to this example, the URL will be [ http://172.18.0.4:30158]( http://172.18.0.4:30158).You can also use the external-ip of the service.Also port forward your service to connect.


>Note: In MySQL: `8.0.14` connection to phpMyAdmin may give error as it is using `caching_sha2_password` and `sha256_password` authentication plugins over `mysql_native_password`. If the error happens do the following for work around. But, It's not recommended to change authentication plugins. See [here](https://stackoverflow.com/questions/49948350/phpmyadmin-on-mysql-8-0) for alternative solutions. You can use mysql_native_password try `kubectl exec -it -n demo mysql-quickstart-0 -- mysql -u root --password='H(Y.s)pg&cX1Ds3J' -e "ALTER USER root IDENTIFIED WITH mysql_native_password BY 'H(Y.s)pg&cX1Ds3J';"`
---
To log into the phpMyAdmin, use host __`mysql-quickstart.demo`__ or __`10.244.0.30`__ , username __`root`__ and password __`H(Y.s)pg&cX1Ds3J`__.

__connecting with `dbgate`__


```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/quickstart/yamls/dbgate.yaml

deployment/dbgate created
service/dbgate created

$ kubectl get pods -n demo --watch
NAME                       READY   STATUS    RESTARTS   AGE
demo                 dbgate-77d7fd4889-bfhb9                         1/1     Running   0          17m

-85d86cf5b5-f4mq4   1/1     Running   0          8s
mysql-quickstart-0         1/1     Running   0          12m


$ kubectl get svc -n demo
NAME                    TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
dbgate                  LoadBalancer   10.96.226.216   <pending>     3000:32475/TCP               51s
mysql-quickstart        ClusterIP      10.96.150.194   <none>        3306/TCP       13m
mysql-quickstart-pods   ClusterIP      None            <none>        3306/TCP       13m

```

Lets, open your browser and go to the following URL: _http://{node-ip}:{dbgate-svc-nodeport}_. For kind cluster, you can get this URL by running the following command:

```bash
$ kubectl get svc -n demo dbgate -o json | jq '.spec.ports[].nodePort'
32475

$ kubectl get node -o json | jq '.items[].status.addresses[].address'
"172.18.0.3"
"kind-control-plane"
"172.18.0.4"
"kind-worker"
"172.18.0.2"
"kind-worker2"

# expected url will be:
url: http://172.18.0.4:32475
```
According to this example, the URL will be [ http://172.18.0.4:30158]( http://172.18.0.4:30158).You can also use the external-ip of the service.Also port forward your service to connect.

You can connect multiple different database using db gate. To log into  MySQL select the MYSQL driver and use server __`mysql-quickstart.demo`__ or __`10.244.0.30`__ , username __`root`__ and password __`H(Y.s)pg&cX1Ds3J`__.

## Database DeletionPolicy

This field is used to regulate the deletion process of the related resources when `MySQL` object is deleted. User can set the value of this field according to their needs. The available options and their use case scenario is described below:

**DoNotTerminate:**

When `deletionPolicy` is set to `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`. You can see this below:

```bash
$ kubectl delete my mysql-quickstart -n demo
Error from server (BadRequest): admission webhook "mysql.validators.kubedb.com" denied the request: mysql "mysql-quickstart" can't be halted. To delete, change spec.deletionPolicy
```

Now, run `kubectl edit my mysql-quickstart -n demo` to set `spec.deletionPolicy` to `Halt` (which deletes the mysql object and keeps PVC, snapshots, Secrets intact) or remove this field (which default to `Delete`). Then you will be able to delete/halt the database.

Learn details of all `DeletionPolicy` [here](/docs/guides/mysql/concepts/database/index.md#specdeletionpolicy).

**Halt:**

Suppose you want to reuse your database volume and credential to deploy your database in future using the same configurations. But, right now you just want to delete the database except the database volumes and credentials. In this scenario, you must set the `MySQL` object `deletionPolicy` to `Halt`.

When the [DeletionPolicy](/docs/guides/mysql/concepts/database/index.md#specdeletionpolicy) is set to `halt` and the MySQL object is deleted, the KubeDB operator will delete the PetSet and its pods but leaves the `PVCs`, `secrets` and database backup data(`snapshots`) intact. You can set the `deletionPolicy` to `halt` in existing database using `edit` command for testing.

At first, run `kubectl edit my mysql-quickstart -n demo` to set `spec.deletionPolicy` to `Halt`. Then delete the mysql object,

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

From the above output, you can see that all mysql resources(`PetSet`, `Service`, etc.) are deleted except `PVC` and `Secret`. You can recreate your mysql again using this resources.

>You can also set the `deletionPolicy` to `Halt`(deprecated). It's behavior same as `halt` and right now `Halt` is replaced by `Halt`.

**Delete:**

If you want to delete the existing database along with the volumes used, but want to restore the database from previously taken `snapshots` and `secrets` then you might want to set the `MySQL` object `deletionPolicy` to `Delete`. In this setting, `PetSet` and the volumes will be deleted. If you decide to restore the database, you can do so using the snapshots and the credentials.

When the [DeletionPolicy](/docs/guides/mysql/concepts/database/index.md#specdeletionpolicy) is set to `Delete` and the MySQL object is deleted, the KubeDB operator will delete the PetSet and its pods along with PVCs but leaves the `secret` and database backup data(`snapshots`) intact.

Suppose, we have a database with `deletionPolicy` set to `Delete`. Now, are going to delete the database using the following command:

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

From the above output, you can see that all mysql resources(`PetSet`, `Service`, `PVCs` etc.) are deleted except `Secret`. You can initialize your mysql using `snapshots`(if previously taken) and `secret`.

>If you don't set the deletionPolicy then the kubeDB set the DeletionPolicy to Delete by-default.

**WipeOut:**

You can totally delete the `MySQL` database and relevant resources without any tracking by setting `deletionPolicy` to `WipeOut`. KubeDB operator will delete all relevant resources of this `MySQL` database (i.e, `PVCs`, `Secrets`, `Snapshots`) when the `deletionPolicy` is set to `WipeOut`.

Suppose, we have a database with `deletionPolicy` set to `WipeOut`. Now, are going to delete the database using the following command:

```yaml
$ kubectl delete my mysql-quickstart -n demo
mysql.kubedb.com "mysql-quickstart" deleted
```

Now, run the following command to get all mysql resources in `demo` namespaces,

```bash
$ kubectl get sts,svc,secret,pvc -n demo
No resources found in demo namespace.
```

From the above output, you can see that all mysql resources are deleted. there is no option to recreate/reinitialize your database if `deletionPolicy` is set to `Delete`.

>Be careful when you set the `deletionPolicy` to `Delete`. Because there is no option to trace the database resources if once deleted the database.

## Database Halted

If you want to delete MySQL resources(`PetSet`,`Service`, etc.) without deleting the `MySQL` object, `PVCs` and `Secret` you have to set the `spec.halted` to `true`. KubeDB operator will be able to delete the MySQL related resources except the `MySQL` object, `PVCs` and `Secret`.

Suppose we have a database running `mysql-quickstart` in our cluster. Now, we are going to set `spec.halted` to `true` in `MySQL`  object by running `kubectl edit -n demo mysql-quickstart` command.

Run the following command to get MySQL resources,

```bash
$ kubectl get my,sts,secret,svc,pvc -n demo
NAME                                VERSION   STATUS   AGE
mysql.kubedb.com/mysql-quickstart   8.0.35    Halted   22m

NAME                           TYPE                                  DATA   AGE
secret/default-token-lgbjm     kubernetes.io/service-account-token   3      27h
secret/mysql-quickstart-auth   Opaque                                2      22m

NAME                                            STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-mysql-quickstart-0   Bound    pvc-7ab0ebb0-bb2e-45c1-9af1-4f175672605b   1Gi        RWO            standard       22m
```

From the above output , you can see that `MySQL` object, `PVCs`, `Secret` are still alive. Then you can recreate your `MySQL` with same configuration.

>When you set `spec.halted` to `true` in `MySQL` object then the `deletionPolicy` is also set to `Halt` by KubeDB operator.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mysql/mysql-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mysql/mysql-quickstart

kubectl delete ns demo
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for production environment. You can follow these tips to avoid them.

1. **Use `storageType: Ephemeral`**. Databases are precious. You might not want to lose your data in your production environment if database pod fail. So, we recommend to use `spec.storageType: Durable` and provide storage spec in `spec.storage` section. For testing purpose, you can just use `spec.storageType: Ephemeral`. KubeDB will use [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) for storage. You will not require to provide `spec.storage` section.
2. **Use `deletionPolicy: WipeOut`**. It is nice to be able to delete everything created by KubeDB for a particular MySQL crd when you delete the crd. For more details about termination policy, please visit [here](/docs/guides/mysql/concepts/database/index.md#specdeletionpolicy).

## Next Steps

- Initialize [MySQL with Script](/docs/guides/mysql/initialization/index.md).
- Monitor your MySQL database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mysql/monitoring/prometheus-operator/index.md).
- Monitor your MySQL database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mysql/monitoring/builtin-prometheus/index.md).
- Use [private Docker registry](/docs/guides/mysql/private-registry/index.md) to deploy MySQL with KubeDB.
- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Detail concepts of [MySQLVersion object](/docs/guides/mysql/concepts/catalog/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
