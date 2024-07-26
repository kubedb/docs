---
title: MariaDB Quickstart
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-quickstart-overview
    name: Overview
    parent: guides-mariadb-quickstart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MariaDB QuickStart

This tutorial will show you how to use KubeDB to run a MariaDB database.

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/mariadb/quickstart/overview/images/mariadb-lifecycle.png">
</p>

> Note: The yaml files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mariadb/quickstart/overview/examples).

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

```
$ kubectl create ns demo
namespace/demo created
```

## Find Available MariaDBVersion

When you have installed KubeDB, it has created `MariaDBVersion` crd for all supported MariaDB versions. Check it by using the following command,

```bash
$ kubectl get mariadbversions
NAME      VERSION   DB_IMAGE          DEPRECATED   AGE
10.4.32   10.4.32   mariadb:10.4.32                9s
10.5.23    10.5.23    mariadb:10.5.23                 9s
10.6.16    10.6.16    mariadb:10.6.16                 9s
```

## Create a MariaDB database

KubeDB implements a `MariaDB` CRD to define the specification of a MariaDB database. Below is the `MariaDB` object created in this tutorial.

`Note`: If your `KubeDB version` is less or equal to `v2024.6.4`, You have to use `v1alpha2` apiVersion.

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "10.5.23"
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/quickstart/overview/examples/sample-mariadb-v1.yaml
mariadb.kubedb.com/sample-mariadb created
```

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "10.5.23"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: Delete
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/quickstart/overview/examples/sample-mariadb-v1alpha2.yaml
mariadb.kubedb.com/sample-mariadb created
```

Here,

- `spec.version` is the name of the MariaDBVersion CRD where the docker images are specified. In this tutorial, a MariaDB `10.5.23` database is going to create.
- `spec.storageType` specifies the type of storage that will be used for MariaDB database. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create MariaDB database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.terminationPolicy` or `spec.deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `MariaDB` crd or which resources KubeDB should keep or delete when you delete `MariaDB` crd. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.terminationPolicy` is set to `DoNotTerminate`.

> Note: spec.storage section is used to create PVC for database pod. It will create PVC with storage size specified instorage.resources.requests field. Don't specify limits here. PVC does not get resized automatically.

KubeDB operator watches for `MariaDB` objects using Kubernetes api. When a `MariaDB` object is created, KubeDB operator will create a new PetSet and a Service with the matching MariaDB object name. KubeDB operator will also create a governing service for PetSets with the name `kubedb`, if one is not already present.

```bash
$ kubectl describe -n demo mariadb sample-mariadb
Name:         sample-mariadb
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1
Kind:         MariaDB
Metadata:
  Creation Timestamp:  2022-06-06T04:42:27Z
  Finalizers:
    kubedb.com
  Generation:  2
  ...
  Resource Version:  2673
  UID:               2f9c9453-6e78-4521-91ea-34ad2da398bc
Spec:
  Allowed Schemas:
    Namespaces:
      From:  Same
  Auth Secret:
    Name:  sample-mariadb-auth
  Pod Template:
    ...
  Replicas:                  1
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         1Gi
    Storage Class Name:  standard
  Storage Type:          Durable
  Termination Policy:    WipeOut
  Version:               10.5.23
Status:
  Conditions:
    Last Transition Time:  2022-06-06T04:42:27Z
    Message:               The KubeDB operator has started the provisioning of MariaDB: demo/sample-mariadb
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2022-06-06T04:43:37Z
    Message:               database sample-mariadb/demo is ready
    Reason:                AllReplicasReady
    Status:                True
    Type:                  Ready
    Last Transition Time:  2022-06-06T04:43:37Z
    Message:               database sample-mariadb/demo is accepting connection
    Reason:                AcceptingConnection
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2022-06-06T04:43:26Z
    Message:               All desired replicas are ready.
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2022-06-06T04:43:37Z
    Message:               The MariaDB: demo/sample-mariadb is successfully provisioned.
    Observed Generation:   2
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Observed Generation:     2
  Phase:                   Ready
Events:
  Type    Reason      Age    From             Message
  ----    ------      ----   ----             -------
  Normal  Successful  3m49s  KubeDB Operator  Successfully created governing service
  Normal  Successful  3m49s  KubeDB Operator  Successfully created Service
  Normal  Successful  3m49s  KubeDB Operator  Successfully created PetSet demo/sample-mariadb
  Normal  Successful  3m49s  KubeDB Operator  Successfully created MariaDB
  Normal  Successful  3m49s  KubeDB Operator  Successfully created appbinding

  
  
$ kubectl get petset -n demo
NAME             READY   AGE
sample-mariadb   1/1     27m

$ kubectl get pvc -n demo
NAME                    STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-sample-mariadb-0   Bound    pvc-10651900-d975-467f-80ff-9c4755bdf917   1Gi        RWO            standard       27m

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS   REASON   AGE
pvc-10651900-d975-467f-80ff-9c4755bdf917   1Gi        RWO            Delete           Bound    demo/data-sample-mariadb-0   standard                27m

$ kubectl get service -n demo
NAME                  TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
sample-mariadb        ClusterIP   10.105.207.172   <none>        3306/TCP   28m
sample-mariadb-pods   ClusterIP   None             <none>        3306/TCP   28m
```

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created. Run the following command to see the modified MariaDB object:

```yaml
$ kubectl get mariadb -n demo sample-mariadb -o yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1","kind":"MariaDB","metadata":{"annotations":{},"name":"sample-mariadb","namespace":"demo"},"spec":{"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","deletionPolicy":"WipeOut","version":"10.5.23"}}
  creationTimestamp: "2021-03-10T04:31:09Z"
  finalizers:
  - kubedb.com
  generation: 2
  ...
  name: sample-mariadb
  namespace: demo
  resourceVersion: "7952"
  selfLink: /apis/kubedb.com/v1/namespaces/demo/mariadbs/sample-mariadb
  uid: 412a4739-ac65-4b5a-a943-5e148f3222b1
spec:
  authSecret:
    name: sample-mariadb-auth
  ...
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
  version: 10.5.23
status:
  observedGeneration: 2
  phase: Ready
```

## Connect with MariaDB database

KubeDB operator has created a new Secret called `mariadb-quickstart-auth` *(format: {mariadb-object-name}-auth)* for storing the password for `mariadb` superuser. This secret contains a `username` key which contains the *username* for MariaDB superuser and a `password` key which contains the *password* for MariaDB superuser.

If you want to use an existing secret please specify that when creating the MariaDB object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `root` as value of `username`.

Now, we need `username` and `password` to connect to this database from `kubeclt exec` command. In this example, `sample-mariadb-auth`  secret holds username and password.

```bash
$ kubectl get secrets -n demo sample-mariadb-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo sample-mariadb-auth -o jsonpath='{.data.\password}' | base64 -d
w*yOU$b53dTbjsjJ
```

We will exec into the pod `sample-mariadb-0` and conncet to the database using `username` and `password`.

```bash
$ kubectl exec -it -n demo sample-mariadb-0 -- mariadb -u root --password='w*yOU$b53dTbjsjJ'
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 335
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> show databases;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| mysql              |
| performance_schema |
+--------------------+
3 rows in set (0.001 sec)

```

## Database DeletionPolicy

This field is used to regulate the deletion process of the related resources when `MariaDB` object is deleted. User can set the value of this field according to their needs. The available options and their use case scenario is described below:

**DoNotTerminate:**

When `deletionPolicy` is set to `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`. If you create a database with `deletionPolicy`  `DoNotTerminate` and try to delete it, you will see this:

```bash
$ kubectl delete mariadb sample-mariadb -n demo
Error from server (BadRequest): admission webhook "mariadb.validators.kubedb.com" denied the request: mariadb "mariadb-quickstart" can't be halted. To delete, change spec.deletionPolicy
```

Now, run `kubectl edit mariadb sample-mariadb -n demo` to set `spec.deletionPolicy` to `Halt` (which deletes the mariadb object and keeps PVC, snapshots, Secrets intact) or remove this field (which default to `Delete`). Then you will be able to delete/halt the database.


**Halt:**

Suppose you want to reuse your database volume and credential to deploy your database in future using the same configurations. But, right now you just want to delete the database except the database volumes and credentials. In this scenario, you must set the `MariaDB` object `deletionPolicy` to `Halt`.

When the `DeletionPolicy` is set to `Halt` and the MariaDB object is deleted, the KubeDB operator will delete the PetSet and its pods but leaves the `PVCs`, `secrets` and database backup data(`snapshots`) intact. You can set the `deletionPolicy` to `Halt` in existing database using `edit` command for testing.

At first, run `kubectl edit mariadb sample-mariadb -n demo` to set `spec.deletionPolicy` to `Halt`. Then delete the mariadb object,

```bash
$ kubectl delete mariadb sample-mariadb -n demo
mariadb.kubedb.com "sample-mariadb" deleted
```

Now, run the following command to get all mariadb resources in `demo` namespaces,

```bash
$ kubectl get sts,svc,secret,pvc -n demo
NAME                         TYPE                                  DATA   AGE
secret/default-token-w2pgw   kubernetes.io/service-account-token   3      31m
secret/sample-mariadb-auth   kubernetes.io/basic-auth              2      39s

NAME                                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-sample-mariadb-0   Bound    pvc-7502c222-2b02-4363-9027-91ab0e7b76dc   1Gi        RWO            standard       39s
```

From the above output, you can see that all mariadb resources(`PetSet`, `Service`, etc.) are deleted except `PVC` and `Secret`. You can recreate your mariadb again using this resources.

**Delete:**

If you want to delete the existing database along with the volumes used, but want to restore the database from previously taken `snapshots` and `secrets` then you might want to set the `MariaDB` object `deletionPolicy` to `Delete`. In this setting, `PetSet` and the volumes will be deleted. If you decide to restore the database, you can do so using the snapshots and the credentials.

When the `DeletionPolicy` is set to `Delete` and the MariaDB object is deleted, the KubeDB operator will delete the PetSet and its pods along with PVCs but leaves the `secret` and database backup data(`snapshots`) intact.

Suppose, we have a database with `deletionPolicy` set to `Delete`. Now, are going to delete the database using the following command:

```bash
$ kubectl delete mariadb sample-mariadb -n demo
mariadb.kubedb.com "sample-mariadb" deleted
```

Now, run the following command to get all mariadb resources in `demo` namespaces,

```bash
$ kubectl get sts,svc,secret,pvc -n demo
NAME                         TYPE                                  DATA   AGE
secret/default-token-w2pgw   kubernetes.io/service-account-token   3      31m
secret/sample-mariadb-auth   kubernetes.io/basic-auth              2      39s
```

From the above output, you can see that all mariadb resources(`PetSet`, `Service`, `PVCs` etc.) are deleted except `Secret`. You can initialize your mariadb using `snapshots`(if previously taken) and `secret`.

>If you don't set the deletionPolicy then the kubeDB set the DeletionPolicy to Delete by-default.

**WipeOut:**

You can totally delete the `MariaDB` database and relevant resources without any tracking by setting `deletionPolicy` to `WipeOut`. KubeDB operator will delete all relevant resources of this `MariaDB` database (i.e, `PVCs`, `Secrets`, `Snapshots`) when the `deletionPolicy` is set to `WipeOut`.

Suppose, we have a database with `deletionPolicy` set to `WipeOut`. Now, are going to delete the database using the following command:

```yaml
$ kubectl delete mariadb sample-mariadb -n demo
mariadb.kubedb.com "sample-mariadb" deleted
```

Now, run the following command to get all mariadb resources in `demo` namespaces,

```bash
$ kubectl get sts,svc,secret,pvc -n demo
No resources found in demo namespace.
```

From the above output, you can see that all mariadb resources are deleted. there is no option to recreate/reinitialize your database if `deletionPolicy` is set to `Delete`.

>Be careful when you set the `deletionPolicy` to `Delete`. Because there is no option to trace the database resources if once deleted the database.

## Database Halted

If you want to delete MariaDB resources(`PetSet`,`Service`, etc.) without deleting the `MariaDB` object, `PVCs` and `Secret` you have to set the `spec.halted` to `true`. KubeDB operator will be able to delete the MariaDB related resources except the `MariaDB` object, `PVCs` and `Secret`.

Suppose we have a database running `mariadb-quickstart` in our cluster. Now, we are going to set `spec.halted` to `true` in `MariaDB`  object by running `kubectl edit -n demo mariadb-quickstart` command.

Run the following command to get MariaDB resources,

```bash
$ kubectl get mariadb,sts,secret,svc,pvc -n demo
NAME                                VERSION   STATUS   AGE
mariadb.kubedb.com/mariadb-quickstart   10.5.23    Halted   22m

NAME                           TYPE                                  DATA   AGE
secret/default-token-lgbjm     kubernetes.io/service-account-token   3      27h
secret/mariadb-quickstart-auth   Opaque                                2      22m

NAME                                            STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-mariadb-quickstart-0   Bound    pvc-7ab0ebb0-bb2e-45c1-9af1-4f175672605b   1Gi        RWO            standard       22m
```

From the above output , you can see that `MariaDB` object, `PVCs`, `Secret` are still alive. Then you can recreate your `MariaDB` with same configuration.

>When you set `spec.halted` to `true` in `MariaDB` object then the `deletionPolicy` is also set to `Halt` by KubeDB operator.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo mariadb/sample-mariadb

kubectl delete ns demo
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for production environment. You can follow these tips to avoid them.

1. **Use `storageType: Ephemeral`**. Databases are precious. You might not want to lose your data in your production environment if database pod fail. So, we recommend to use `spec.storageType: Durable` and provide storage spec in `spec.storage` section. For testing purpose, you can just use `spec.storageType: Ephemeral`. KubeDB will use [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) for storage. You will not require to provide `spec.storage` section.
2. **Use `deletionPolicy: WipeOut`**. It is nice to be able to delete everything created by KubeDB for a particular MariaDB crd when you delete the crd.

## Next Steps

- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
