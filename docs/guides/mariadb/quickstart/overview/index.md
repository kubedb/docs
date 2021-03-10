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

> Note: The yaml files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/quides/mariadb/quickstart/overview/examples).

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
NAME      VERSION   DB_IMAGE                 DEPRECATED   AGE
10.4.17   10.4.17   kubedb/mariadb:10.4.17                15h
10.5.8    10.5.8    kubedb/mariadb:10.5.8                 15h
```

## Create a MariaDB database

KubeDB implements a `MariaDB` CRD to define the specification of a MariaDB database. Below is the `MariaDB` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "10.5.8"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/quickstart/overview/examples/sample-mariadb.yaml
mariadb.kubedb.com/mariadb-quickstart created
```

Here,

- `spec.version` is the name of the MariaDBVersion CRD where the docker images are specified. In this tutorial, a MariaDB `8.0.21` database is going to create.
- `spec.storageType` specifies the type of storage that will be used for MariaDB database. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create MariaDB database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.terminationPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `MariaDB` crd or which resources KubeDB should keep or delete when you delete `MariaDB` crd. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.terminationPolicy` is set to `DoNotTerminate`. Learn details of all `TerminationPolicy` [here](/docs/guides/mariadb/concepts/mariadb.md#specterminationpolicy)

> Note: spec.storage section is used to create PVC for database pod. It will create PVC with storage size specified instorage.resources.requests field. Don't specify limits here. PVC does not get resized automatically.

KubeDB operator watches for `MariaDB` objects using Kubernetes api. When a `MariaDB` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching MariaDB object name. KubeDB operator will also create a governing service for StatefulSets with the name `kubedb`, if one is not already present.

```bash
$ kubectl dba describe my -n demo mariadb-quickstart
Name:               mariadb-quickstart
Namespace:          demo
CreationTimestamp:  Mon, 31 Aug 2020 16:39:47 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1alpha2","kind":"MariaDB","metadata":{"annotations":{},"name":"mariadb-quickstart","namespace":"demo"},"spec":{"storage":{"acces...
Replicas:           1  total
Status:             Running
StorageType:        Durable
Volume:
  StorageClass:      standard
  Capacity:          1Gi
  Access Modes:      RWO
Halted:              false
Halted:              false
Termination Policy:  DoNotTerminate

StatefulSet:          
  Name:               mariadb-quickstart
  CreationTimestamp:  Mon, 31 Aug 2020 16:39:47 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mariadbs.kubedb.com
                        app.kubernetes.io/instance=mariadb-quickstart
  Annotations:        <none>
  Replicas:           824634389080 desired | 1 total
  Pods Status:        1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         mariadb-quickstart
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mariadbs.kubedb.com
                  app.kubernetes.io/instance=mariadb-quickstart
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.103.57.226
  Port:         db  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.2.13:3306

Service:        
  Name:         mariadb-quickstart-gvr
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mariadbs.kubedb.com
                  app.kubernetes.io/instance=mariadb-quickstart
  Annotations:    service.alpha.kubernetes.io/tolerate-unready-endpoints=true
  Type:         ClusterIP
  IP:           None
  Port:         db  3306/TCP
  TargetPort:   3306/TCP
  Endpoints:    10.244.2.13:3306

Database Secret:
  Name:         mariadb-quickstart-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mariadbs.kubedb.com
                  app.kubernetes.io/instance=mariadb-quickstart
  Annotations:  <none>
  Type:         Opaque
  Data:
    password:  16 bytes
    username:  4 bytes

AppBinding:
  Metadata:
    Annotations:
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1alpha2","kind":"MariaDB","metadata":{"annotations":{},"name":"mariadb-quickstart","namespace":"demo"},"spec":{"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","terminationPolicy":"DoNotTerminate","version":"8.0.21"}}

    Creation Timestamp:  2020-08-31T10:40:53Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    mariadb-quickstart
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        mariadb
      app.kubernetes.io/version:     8.0.21
      app.kubernetes.io/name:        mariadbs.kubedb.com
      app.kubernetes.io/instance:               mariadb-quickstart
    Name:                            mariadb-quickstart
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    mariadb-quickstart
        Path:    /
        Port:    3306
        Scheme:  mariadb
      URL:       tcp(mariadb-quickstart:3306)/
    Secret:
      Name:   mariadb-quickstart-auth
    Type:     kubedb.com/mariadb
    Version:  8.0.21

Events:
  Type    Reason      Age   From            Message
  ----    ------      ----  ----            -------
  Normal  Successful  3m    MariaDB operator  Successfully created Service
  Normal  Successful  2m    MariaDB operator  Successfully created StatefulSet
  Normal  Successful  2m    MariaDB operator  Successfully created MariaDB
  Normal  Successful  2m    MariaDB operator  Successfully created appbinding


$ kubectl get statefulset -n demo
NAME               READY   AGE
mariadb-quickstart   1/1     2m22s

$ kubectl get pvc -n demo
NAME                      STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-mariadb-quickstart-0   Bound     pvc-652e02c7-0d7f-11e8-9091-08002751ae8c   1Gi        RWO            standard       10m

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS    CLAIM                          STORAGECLASS   REASON    AGE
pvc-652e02c7-0d7f-11e8-9091-08002751ae8c   1Gi        RWO            Delete           Bound     demo/data-mariadb-quickstart-0   standard                 11m

$ kubectl get service -n demo
NAME                    TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE
myadmin                 LoadBalancer   10.104.142.213   <pending>     80:31529/TCP   6h2m
mariadb-quickstart        ClusterIP      10.109.217.165   <none>        3306/TCP       5m56s
mariadb-quickstart-gvr    ClusterIP      None             <none>        3306/TCP       5m56s
```

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created. Run the following command to see the modified MariaDB object:

```yaml
$ kubectl get my -n demo mariadb-quickstart -o yaml
 $ kubectl get my -n demo mariadb-quickstart -o yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"MariaDB","metadata":{"annotations":{},"name":"mariadb-quickstart","namespace":"demo"},"spec":{"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","terminationPolicy":"DoNotTerminate","version":"8.0.21"}}
  creationTimestamp: "2020-08-27T12:19:42Z"
  finalizers:
  - kubedb.com
  ...
  name: mariadb-quickstart
  namespace: demo
  resourceVersion: "70812"
  selfLink: /apis/kubedb.com/v1alpha2/namespaces/demo/mariadbs/mariadb-quickstart
  uid: 837ac85a-134a-457e-b126-f4681d92f117
spec:
  authSecret:
    name: mariadb-quickstart-auth
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      resources: {}
      serviceAccountName: mariadb-quickstart
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
  version: 8.0.21
status:
  observedGeneration: 2
  phase: Running
```

## Connect with MariaDB database

KubeDB operator has created a new Secret called `mariadb-quickstart-auth` *(format: {mariadb-object-name}-auth)* for storing the password for `mariadb` superuser. This secret contains a `username` key which contains the *username* for MariaDB superuser and a `password` key which contains the *password* for MariaDB superuser.

If you want to use an existing secret please specify that when creating the MariaDB object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `root` as value of `username`. For more details see [here](/docs/guides/mariadb/concepts/mariadb.md#specdatabasesecret).

Now, you can connect to this database from the phpMyAdmin dashboard using the database pod IP and and `mariadb` user password.

```bash
$ kubectl get pods mariadb-quickstart-0 -n demo -o yaml | grep podIP
  podIP: 10.244.2.13

$ kubectl get secrets -n demo mariadb-quickstart-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mariadb-quickstart-auth -o jsonpath='{.data.\password}' | base64 -d
l0yKjI1E7IMohsGR
```

---
>Note: In MariaDB: `8.0.14` connection to phpMyAdmin may give error as it is using `caching_sha2_password` and `sha256_password` authentication plugins over `mariadb_native_password`. If the error happens do the following for work around. But, It's not recommended to change authentication plugins. See [here](https://stackoverflow.com/questions/49948350/phpmyadmin-on-mariadb-8-0) for alternative solutions.

```bash
kubectl exec -it -n demo mariadb-quickstart-0 -- mariadb -u root --password=l0yKjI1E7IMohsGR -e "ALTER USER root IDENTIFIED WITH mariadb_native_password BY 'l0yKjI1E7IMohsGR';"
```
---

Now, open your browser and go to the following URL: _http://{node-ip}:{myadmin-svc-nodeport}_. To log into the phpMyAdmin, use host __`mariadb-quickstart.demo`__ or __`10.244.2.13`__ , username __`root`__ and password __`l0yKjI1E7IMohsGR`__.

## Database TerminationPolicy

This field is used to regulate the deletion process of the related resources when `MariaDB` object is deleted. User can set the value of this field according to their needs. The available options and their use case scenario is described below:

**DoNotTerminate:**

When `terminationPolicy` is set to `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.terminationPolicy` is set to `DoNotTerminate`. You can see this below:

```bash
$ kubectl delete my mariadb-quickstart -n demo
Error from server (BadRequest): admission webhook "mariadb.validators.kubedb.com" denied the request: mariadb "mariadb-quickstart" can't be halted. To delete, change spec.terminationPolicy
```

Now, run `kubectl edit my mariadb-quickstart -n demo` to set `spec.terminationPolicy` to `Halt` (which deletes the mariadb object and keeps PVC, snapshots, Secrets intact) or remove this field (which default to `Delete`). Then you will be able to delete/halt the database.

Learn details of all `TerminationPolicy` [here](/docs/guides/mariadb/concepts/mariadb.md#specterminationpolicy).

**Halt:**

Suppose you want to reuse your database volume and credential to deploy your database in future using the same configurations. But, right now you just want to delete the database except the database volumes and credentials. In this scenario, you must set the `MariaDB` object `terminationPolicy` to `Halt`.

When the [TerminationPolicy](/docs/guides/mariadb/concepts/mariadb.md#specterminationpolicy) is set to `halt` and the MariaDB object is deleted, the KubeDB operator will delete the StatefulSet and its pods but leaves the `PVCs`, `secrets` and database backup data(`snapshots`) intact. You can set the `terminationPolicy` to `halt` in existing database using `edit` command for testing.

At first, run `kubectl edit my mariadb-quickstart -n demo` to set `spec.terminationPolicy` to `Halt`. Then delete the mariadb object,

```bash
$ kubectl delete my mariadb-quickstart -n demo
mariadb.kubedb.com "mariadb-quickstart" deleted
```

Now, run the following command to get all mariadb resources in `demo` namespaces,

```bash
$ kubectl get sts,svc,secret,pvc -n demo
NAME                           TYPE                                  DATA   AGE
secret/default-token-lgbjm     kubernetes.io/service-account-token   3      23h
secret/mariadb-quickstart-auth   Opaque                                2      20h

NAME                                            STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-mariadb-quickstart-0   Bound    pvc-716f627c-9aa2-47b6-aa64-a547aab6f55c   1Gi        RWO            standard       20h
```

From the above output, you can see that all mariadb resources(`StatefulSet`, `Service`, etc.) are deleted except `PVC` and `Secret`. You can recreate your mariadb again using this resources.

>You can also set the `terminationPolicy` to `Halt`(deprecated). It's behavior same as `halt` and right now `Halt` is replaced by `Halt`.

**Delete:**

If you want to delete the existing database along with the volumes used, but want to restore the database from previously taken `snapshots` and `secrets` then you might want to set the `MariaDB` object `terminationPolicy` to `Delete`. In this setting, `StatefulSet` and the volumes will be deleted. If you decide to restore the database, you can do so using the snapshots and the credentials.

When the [TerminationPolicy](/docs/guides/mariadb/concepts/mariadb.md#specterminationpolicy) is set to `Delete` and the MariaDB object is deleted, the KubeDB operator will delete the StatefulSet and its pods along with PVCs but leaves the `secret` and database backup data(`snapshots`) intact.

Suppose, we have a database with `terminationPolicy` set to `Delete`. Now, are going to delete the database using the following command:

```bash
$ kubectl delete my mariadb-quickstart -n demo
mariadb.kubedb.com "mariadb-quickstart" deleted
```

Now, run the following command to get all mariadb resources in `demo` namespaces,

```bash
$ kubectl get sts,svc,secret,pvc -n demo
NAME                           TYPE                                  DATA   AGE
secret/default-token-lgbjm     kubernetes.io/service-account-token   3      24h
secret/mariadb-quickstart-auth   Opaque
```

From the above output, you can see that all mariadb resources(`StatefulSet`, `Service`, `PVCs` etc.) are deleted except `Secret`. You can initialize your mariadb using `snapshots`(if previously taken) and `secret`.

>If you don't set the terminationPolicy then the kubeDB set the TerminationPolicy to Delete by-default.

**WipeOut:**

You can totally delete the `MariaDB` database and relevant resources without any tracking by setting `terminationPolicy` to `WipeOut`. KubeDB operator will delete all relevant resources of this `MariaDB` database (i.e, `PVCs`, `Secrets`, `Snapshots`) when the `terminationPolicy` is set to `WipeOut`.

Suppose, we have a database with `terminationPolicy` set to `WipeOut`. Now, are going to delete the database using the following command:

```yaml
$ kubectl delete my mariadb-quickstart -n demo
mariadb.kubedb.com "mariadb-quickstart" deleted
```

Now, run the following command to get all mariadb resources in `demo` namespaces,

```bash
$ kubectl get sts,svc,secret,pvc -n demo
No resources found in demo namespace.
```

From the above output, you can see that all mariadb resources are deleted. there is no option to recreate/reinitialize your database if `terminationPolicy` is set to `Delete`.

>Be careful when you set the `terminationPolicy` to `Delete`. Because there is no option to trace the database resources if once deleted the database.

## Database Halted

If you want to delete MariaDB resources(`StatefulSet`,`Service`, etc.) without deleting the `MariaDB` object, `PVCs` and `Secret` you have to set the `spec.halted` to `true`. KubeDB operator will be able to delete the MariaDB related resources except the `MariaDB` object, `PVCs` and `Secret`.

Suppose we have a database running `mariadb-quickstart` in our cluster. Now, we are going to set `spec.halted` to `true` in `MariaDB`  object by running `kubectl edit -n demo mariadb-quickstart` command.

Run the following command to get MariaDB resources,

```bash
$ kubectl get my,sts,secret,svc,pvc -n demo
NAME                                VERSION   STATUS   AGE
mariadb.kubedb.com/mariadb-quickstart   8.0.21    Halted   22m

NAME                           TYPE                                  DATA   AGE
secret/default-token-lgbjm     kubernetes.io/service-account-token   3      27h
secret/mariadb-quickstart-auth   Opaque                                2      22m

NAME                                            STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-mariadb-quickstart-0   Bound    pvc-7ab0ebb0-bb2e-45c1-9af1-4f175672605b   1Gi        RWO            standard       22m
```

From the above output , you can see that `MariaDB` object, `PVCs`, `Secret` are still alive. Then you can recreate your `MariaDB` with same configuration.

>When you set `spec.halted` to `true` in `MariaDB` object then the `terminationPolicy` is also set to `Halt` by KubeDB operator.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mariadb/mariadb-quickstart -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mariadb/mariadb-quickstart

kubectl delete ns demo
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for production environment. You can follow these tips to avoid them.

1. **Use `storageType: Ephemeral`**. Databases are precious. You might not want to lose your data in your production environment if database pod fail. So, we recommend to use `spec.storageType: Durable` and provide storage spec in `spec.storage` section. For testing purpose, you can just use `spec.storageType: Ephemeral`. KubeDB will use [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) for storage. You will not require to provide `spec.storage` section.
2. **Use `terminationPolicy: WipeOut`**. It is nice to be able to delete everything created by KubeDB for a particular MariaDB crd when you delete the crd. For more details about termination policy, please visit [here](/docs/guides/mariadb/concepts/mariadb.md#specterminationpolicy).

## Next Steps

- Initialize [MariaDB with Script](/docs/guides/mariadb/initialization/using-script.md).
- Monitor your MariaDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mariadb/monitoring/using-prometheus-operator.md).
- Monitor your MariaDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mariadb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mariadb/private-registry/using-private-registry.md) to deploy MariaDB with KubeDB.
- Detail concepts of [MariaDB object](/docs/guides/mariadb/concepts/mariadb.md).
- Detail concepts of [MariaDBVersion object](/docs/guides/mariadb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
