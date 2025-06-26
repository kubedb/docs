---
title: ClickHouse Quickstart
menu:
  docs_{{ .version }}:
    identifier: guides-clickhouse-quickstart
    name: Quickstart
    parent: guides-clickhouse
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ClickHouse QuickStart

This tutorial will show you how to use KubeDB to run a ClickHouse database.

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/clickhouse/quickstart/images/clickhouse-lifecycle.png">
</p>

> Note: The yaml files used in this tutorial are stored in [docs/guides/clickhouse/quickstart/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/clickhouse/quickstart/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

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

## Find Available ClickHouseVersion

When you have installed KubeDB, it has created `ClickHouseVersion` crd for all supported ClickHouse versions. Check it by using the following command,

```bash
$ kubectl get clickhouseversions
NAME     VERSION   DB_IMAGE                              DEPRECATED   AGE
24.4.1   24.4.1    clickhouse/clickhouse-server:24.4.1                3h21m
```

## Create a ClickHouse database

KubeDB implements a `ClickHouse` CRD to define the specification of a ClickHouse database. Below is the `ClickHouse` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: clickhouse-quickstart
  namespace: demo
spec:
  version: 24.4.1
  replicas: 1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/clickhouse/quickstart/yamls/quickstart-v1alpha2.yaml
clickhouse.kubedb.com/clickhouse-quickstart created
```

Here,

- `spec.version` is the name of the ClickHouseVersion CRD where the docker images are specified. In this tutorial, a ClickHouse `8.0.35` database is going to be created.
- `spec.storageType` specifies the type of storage that will be used for ClickHouse database. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create ClickHouse database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.terminationPolicy` or `spec.deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `ClickHouse` crd or which resources KubeDB should keep or delete when you delete `ClickHouse` crd. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.terminationPolicy` is set to `DoNotTerminate`. Learn details of all `TerminationPolicy` [here](/docs/guides/mysql/concepts/database/index.md#specterminationpolicy)

> Note: spec.storage section is used to create PVC for database pod. It will create PVC with storage size specified instorage.resources.requests field. Don't specify limits here. PVC does not get resized automatically.

KubeDB operator watches for `ClickHouse` objects using Kubernetes api. When a `ClickHouse` object is created, KubeDB operator will create a new PetSet and a Service with the matching ClickHouse object name. KubeDB operator will also create a governing service for PetSets with the name `kubedb`, if one is not already present.

```bash
$ kubectl dba describe my -n demo clickhouse-quickstart
Name:               clickhouse-quickstart
Namespace:          demo
CreationTimestamp:  Fri, 03 Jun 2022 12:50:40 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1","kind":"ClickHouse","metadata":{"annotations":{},"name":"clickhouse-quickstart","namespace":"demo"},"spec":{"storage":{"acces...
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
  Name:               clickhouse-quickstart
  CreationTimestamp:  Fri, 03 Jun 2022 12:50:40 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=clickhouse-quickstart
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=clickhouses.kubedb.com
  Annotations:        <none>
  Replicas:           824646358808 desired | 1 total
  Pods Status:        1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         clickhouse-quickstart
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=clickhouse-quickstart
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=clickhouses.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.150.194
  Port:         primary  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.30:3306

Service:        
  Name:         clickhouse-quickstart-pods
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=clickhouse-quickstart
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=clickhouses.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           None
  Port:         db  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.30:3306

Auth Secret:
  Name:         clickhouse-quickstart-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=clickhouse-quickstart
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=clickhouses.kubedb.com
  Annotations:  <none>
  Type:         kubernetes.io/basic-auth
  Data:
    password:  16 bytes
    username:  4 bytes

AppBinding:
  Metadata:
    Annotations:
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1","kind":"ClickHouse","metadata":{"annotations":{},"name":"clickhouse-quickstart","namespace":"demo"},"spec":{"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","deletionPolicy":"DoNotTerminate","version":"8.0.35"}}

    Creation Timestamp:  2022-06-03T06:50:40Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    clickhouse-quickstart
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        clickhouses.kubedb.com
    Name:                            clickhouse-quickstart
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    clickhouse-quickstart
        Path:    /
        Port:    3306
        Scheme:  clickhouse
      URL:       tcp(clickhouse-quickstart.demo.svc:3306)/
    Parameters:
      API Version:  appcatalog.appscode.com/v1alpha1
      Kind:         StashAddon
      Stash:
        Addon:
          Backup Task:
            Name:  clickhouse-backup-8.0.21
            Params:
              Name:   args
              Value:  --all-databases --set-gtid-purged=OFF
          Restore Task:
            Name:  clickhouse-restore-8.0.21
    Secret:
      Name:   clickhouse-quickstart-auth
    Type:     kubedb.com/clickhouse
    Version:  8.0.35

Events:
  Type     Reason      Age   From             Message
  ----     ------      ----  ----             -------
  Normal   Successful  32s   KubeDB Operator  Successfully created governing service
  Normal   Successful  32s   KubeDB Operator  Successfully created service for primary/standalone
  Normal   Successful  32s   KubeDB Operator  Successfully created database auth secret
  Normal   Successful  32s   KubeDB Operator  Successfully created PetSet
  Normal   Successful  32s   KubeDB Operator  Successfully created ClickHouse
  Normal   Successful  32s   KubeDB Operator  Successfully created appbinding



$ kubectl get petset -n demo
NAME               READY   AGE
clickhouse-quickstart   1/1     3m19s

$ kubectl get pvc -n demo
NAME                      STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-clickhouse-quickstart-0   Bound    pvc-ab44ce95-2300-47d7-8f25-3cd7bc5b0091   1Gi        RWO            standard       3m50s


$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS   REASON   AGE
pvc-ab44ce95-2300-47d7-8f25-3cd7bc5b0091   1Gi        RWO            Delete           Bound    demo/data-clickhouse-quickstart-0   standard                4m19s

kubectl get service -n demo
NAME                    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
clickhouse-quickstart        ClusterIP   10.96.150.194   <none>        3306/TCP   5m13s
clickhouse-quickstart-pods   ClusterIP   None            <none>        3306/TCP   5m13s
```

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created. Run the following command to see the modified ClickHouse object:

```yaml
$ kubectl get my -n demo clickhouse-quickstart -o yaml
apiVersion: kubedb.com/v1
kind: ClickHouse
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1","kind":"ClickHouse","metadata":{"annotations":{},"name":"clickhouse-quickstart","namespace":"demo"},"spec":{"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","deletionPolicy":"DoNotTerminate","version":"8.0.35"}}
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
    name: clickhouse-quickstart-auth
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      resources:
        limits:
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      serviceAccountName: clickhouse-quickstart
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
    message: 'The KubeDB operator has started the provisioning of ClickHouse: demo/clickhouse-quickstart'
    reason: DatabaseProvisioningStartedSuccessfully
    status: "True"
    type: ProvisioningStarted
  - lastTransitionTime: "2022-06-03T06:50:46Z"
    message: All desired replicas are ready.
    reason: AllReplicasReady
    status: "True"
    type: ReplicaReady
  - lastTransitionTime: "2022-06-03T06:51:05Z"
    message: database demo/clickhouse-quickstart is accepting connection
    reason: AcceptingConnection
    status: "True"
    type: AcceptingConnection
  - lastTransitionTime: "2022-06-03T06:51:05Z"
    message: database demo/clickhouse-quickstart is ready
    reason: AllReplicasReady
    status: "True"
    type: Ready
  - lastTransitionTime: "2022-06-03T06:51:05Z"
    message: 'The ClickHouse: demo/clickhouse-quickstart is successfully provisioned.'
    observedGeneration: 2
    reason: DatabaseSuccessfullyProvisioned
    status: "True"
    type: Provisioned
  observedGeneration: 2
  phase: Ready

```

## Connect with ClickHouse database

KubeDB operator has created a new Secret called `clickhouse-quickstart-auth` *(format: {clickhouse-object-name}-auth)* for storing the password for `clickhouse` superuser. This secret contains a `username` key which contains the *username* for ClickHouse superuser and a `password` key which contains the *password* for ClickHouse superuser.

If you want to use an existing secret please specify that when creating the ClickHouse object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `root` as value of `username`. For more details see [here](/docs/guides/mysql/concepts/database/index.md#specdatabasesecret).

Now, we need `username` and `password` to connect to this database from `kubectl exec` command. In this example  `clickhouse-quickstart-auth` secret holds username and password

```bash
$ kubectl get pods clickhouse-quickstart-0 -n demo -o yaml | grep podIP
  podIP: 10.244.0.14

$ kubectl get secrets -n demo clickhouse-quickstart-auth -o jsonpath='{.data.\username}' | base64 -d
admin

$ kubectl get secrets -n demo clickhouse-quickstart-auth -o jsonpath='{.data.\password}' | base64 -d
e6S2JnXBxSe39hxg
```
we will exec into the pod `clickhouse-quickstart-0` and connect to the database using username and password

```bash
$ kubectl exec -it -n demo clickhouse-quickstart-0 -- bash
Defaulted container "clickhouse" out of: clickhouse, clickhouse-init (init)
clickhouse@clickhouse-quickstart-0:/$ clickhouse-client -uadmin --password="e6S2JnXBxSe39hxg"

ClickHouse client version 24.4.1.2088 (official build).
Connecting to localhost:9000 as user admin.
Connected to ClickHouse server version 24.4.1.

Warnings:
 * Delay accounting is not enabled, OSIOWaitMicroseconds will not be gathered. You can enable it using `echo 1 > /proc/sys/kernel/task_delayacct` or by using sysctl.
 * Effective user of the process (clickhouse) does not match the owner of the data (root).

clickhouse-quickstart-0.clickhouse-quickstart-pods.demo.svc.cluster.local :) show databases

SHOW DATABASES

Query id: 12353f2c-d6d1-4dbc-a4cf-aa2f6a5e0ce4

   ┌─name───────────────┐
1. │ INFORMATION_SCHEMA │
2. │ default            │
3. │ information_schema │
4. │ kubedb_system      │
5. │ system             │
   └────────────────────┘

5 rows in set. Elapsed: 0.004 sec.

```
## Database DeletionPolicy

This field is used to regulate the deletion process of the related resources when `ClickHouse` object is deleted. User can set the value of this field according to their needs. The available options and their use case scenario is described below:

**DoNotTerminate:**

When `deletionPolicy` is set to `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`. You can see this below:

```bash
$ kubectl delete my clickhouse-quickstart -n demo
Error from server (BadRequest): admission webhook "clickhouse.validators.kubedb.com" denied the request: clickhouse "clickhouse-quickstart" can't be halted. To delete, change spec.deletionPolicy
```

Now, run `kubectl edit my clickhouse-quickstart -n demo` to set `spec.deletionPolicy` to `Halt` (which deletes the clickhouse object and keeps PVC, snapshots, Secrets intact) or remove this field (which default to `Delete`). Then you will be able to delete/halt the database.

Learn details of all `DeletionPolicy` [here](/docs/guides/mysql/concepts/database/index.md#specdeletionpolicy).

**Halt:**

Suppose you want to reuse your database volume and credential to deploy your database in future using the same configurations. But, right now you just want to delete the database except the database volumes and credentials. In this scenario, you must set the `ClickHouse` object `deletionPolicy` to `Halt`.

When the [DeletionPolicy](/docs/guides/clickhouse/concepts/database/index.md#specdeletionpolicy) is set to `halt` and the ClickHouse object is deleted, the KubeDB operator will delete the PetSet and its pods but leaves the `PVCs`, `secrets` and database backup data(`snapshots`) intact. You can set the `deletionPolicy` to `halt` in existing database using `edit` command for testing.

At first, run `kubectl edit ch clickhouse-quickstart -n demo` to set `spec.deletionPolicy` to `Halt`. Then delete the clickhouse object,

```bash
$ kubectl delete ch clickhouse-quickstart -n demo
clickhouse.kubedb.com "clickhouse-quickstart" deleted
```

Now, run the following command to get all clickhouse resources in `demo` namespaces,

```bash
$ kubectl get sts,svc,secret,pvc -n demo
NAME                           TYPE                                  DATA   AGE
secret/default-token-lgbjm     kubernetes.io/service-account-token   3      23h
secret/clickhouse-quickstart-auth   Opaque                                2      20h

NAME                                            STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-clickhouse-quickstart-0   Bound    pvc-716f627c-9aa2-47b6-aa64-a547aab6f55c   1Gi        RWO            standard       20h
```

From the above output, you can see that all clickhouse resources(`PetSet`, `Service`, etc.) are deleted except `PVC` and `Secret`. You can recreate your clickhouse again using this resources.

>You can also set the `deletionPolicy` to `Halt`(deprecated). It's behavior same as `halt` and right now `Halt` is replaced by `Halt`.

**Delete:**

If you want to delete the existing database along with the volumes used, but want to restore the database from previously taken `snapshots` and `secrets` then you might want to set the `ClickHouse` object `deletionPolicy` to `Delete`. In this setting, `PetSet` and the volumes will be deleted. If you decide to restore the database, you can do so using the snapshots and the credentials.

When the [DeletionPolicy](/docs/guides/clickhouse/concepts/database/index.md#specdeletionpolicy) is set to `Delete` and the ClickHouse object is deleted, the KubeDB operator will delete the PetSet and its pods along with PVCs but leaves the `secret` and database backup data(`snapshots`) intact.

Suppose, we have a database with `deletionPolicy` set to `Delete`. Now, are going to delete the database using the following command:

```bash
$ kubectl delete my clickhouse-quickstart -n demo
clickhouse.kubedb.com "clickhouse-quickstart" deleted
```

Now, run the following command to get all clickhouse resources in `demo` namespaces,

```bash
$ kubectl get sts,svc,secret,pvc -n demo
NAME                           TYPE                                  DATA   AGE
secret/default-token-lgbjm     kubernetes.io/service-account-token   3      24h
secret/clickhouse-quickstart-auth   Opaque
```

From the above output, you can see that all clickhouse resources(`PetSet`, `Service`, `PVCs` etc.) are deleted except `Secret`. You can initialize your clickhouse using `snapshots`(if previously taken) and `secret`.

>If you don't set the deletionPolicy then the kubeDB set the DeletionPolicy to Delete by-default.

**WipeOut:**

You can totally delete the `ClickHouse` database and relevant resources without any tracking by setting `deletionPolicy` to `WipeOut`. KubeDB operator will delete all relevant resources of this `ClickHouse` database (i.e, `PVCs`, `Secrets`, `Snapshots`) when the `deletionPolicy` is set to `WipeOut`.

Suppose, we have a database with `deletionPolicy` set to `WipeOut`. Now, are going to delete the database using the following command:

```yaml
$ kubectl delete my clickhouse-quickstart -n demo
clickhouse.kubedb.com "clickhouse-quickstart" deleted
```

Now, run the following command to get all clickhouse resources in `demo` namespaces,

```bash
$ kubectl get sts,svc,secret,pvc -n demo
No resources found in demo namespace.
```

From the above output, you can see that all clickhouse resources are deleted. there is no option to recreate/reinitialize your database if `deletionPolicy` is set to `Delete`.

>Be careful when you set the `deletionPolicy` to `Delete`. Because there is no option to trace the database resources if once deleted the database.

## Database Halted

If you want to delete ClickHouse resources(`PetSet`,`Service`, etc.) without deleting the `ClickHouse` object, `PVCs` and `Secret` you have to set the `spec.halted` to `true`. KubeDB operator will be able to delete the ClickHouse related resources except the `ClickHouse` object, `PVCs` and `Secret`.

Suppose we have a database running `clickhouse-quickstart` in our cluster. Now, we are going to set `spec.halted` to `true` in `ClickHouse`  object by running `kubectl edit -n demo clickhouse-quickstart` command.

Run the following command to get ClickHouse resources,

```bash
$ kubectl get my,sts,secret,svc,pvc -n demo
NAME                                VERSION   STATUS   AGE
clickhouse.kubedb.com/clickhouse-quickstart   8.0.35    Halted   22m

NAME                           TYPE                                  DATA   AGE
secret/default-token-lgbjm     kubernetes.io/service-account-token   3      27h
secret/clickhouse-quickstart-auth   Opaque                                2      22m

NAME                                            STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-clickhouse-quickstart-0   Bound    pvc-7ab0ebb0-bb2e-45c1-9af1-4f175672605b   1Gi        RWO            standard       22m
```

From the above output , you can see that `ClickHouse` object, `PVCs`, `Secret` are still alive. Then you can recreate your `ClickHouse` with same configuration.

>When you set `spec.halted` to `true` in `ClickHouse` object then the `deletionPolicy` is also set to `Halt` by KubeDB operator.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo clickhouse/clickhouse-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo clickhouse/clickhouse-quickstart

kubectl delete ns demo
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for production environment. You can follow these tips to avoid them.

1. **Use `storageType: Ephemeral`**. Databases are precious. You might not want to lose your data in your production environment if database pod fail. So, we recommend to use `spec.storageType: Durable` and provide storage spec in `spec.storage` section. For testing purpose, you can just use `spec.storageType: Ephemeral`. KubeDB will use [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) for storage. You will not require to provide `spec.storage` section.
2. **Use `deletionPolicy: WipeOut`**. It is nice to be able to delete everything created by KubeDB for a particular ClickHouse crd when you delete the crd. For more details about termination policy, please visit [here](/docs/guides/clickhouse/concepts/database/index.md#specdeletionpolicy).

## Next Steps


