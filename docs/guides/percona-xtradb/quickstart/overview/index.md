---
title: PerconaXtraDB Quickstart
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-quickstart-overview
    name: Overview
    parent: guides-perconaxtradb-quickstart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PerconaXtraDB QuickStart

This tutorial will show you how to use KubeDB to run a PerconaXtraDB database.

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/percona-xtradb/quickstart/overview/images/perconaxtradb-lifecycle.svg">
</p>

> Note: The yaml files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/percona-xtradb/quickstart/overview/examples).

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

## Find Available PerconaXtraDBVersion

When you have installed KubeDB, it has created `PerconaXtraDBVersion` crd for all supported PerconaXtraDB versions. Check it by using the following command,

```bash
$ kubectl get perconaxtradbversions
NAME     VERSION   DB_IMAGE                                DEPRECATED   AGE
8.0.26   8.0.26    percona/percona-xtradb-cluster:8.0.26                6m1s
8.0.28   8.0.28    percona/percona-xtradb-cluster:8.0.28                6m1s
```

## Create a PerconaXtraDB database

KubeDB implements a `PerconaXtraDB` CRD to define the specification of a PerconaXtraDB database. Below is the `PerconaXtraDB` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: sample-pxc
  namespace: demo
spec:
  version: "8.0.26"
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/quickstart/overview/examples/sample-pxc.yaml
perconaxtradb.kubedb.com/sample-pxc created
```

Here,

- `spec.version` is the name of the PerconaXtraDBVersion CRD where the docker images are specified. In this tutorial, a PerconaXtraDB `8.0.26` database is going to create.
- `spec.storageType` specifies the type of storage that will be used for PerconaXtraDB database. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create PerconaXtraDB database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.terminationPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `PerconaXtraDB` crd or which resources KubeDB should keep or delete when you delete `PerconaXtraDB` crd. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.terminationPolicy` is set to `DoNotTerminate`.

> Note: `spec.storage` section is used to create PVC for database pod. It will create PVC with storage size specified in `storage.resources.requests` field. Don't specify limits here. PVC does not get resized automatically.

KubeDB operator watches for `PerconaXtraDB` objects using Kubernetes api. When a `PerconaXtraDB` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching PerconaXtraDB object name. KubeDB operator will also create a governing service for StatefulSets with the name `kubedb`, if one is not already present.

```bash
$ kubectl describe -n demo perconaxtradb sample-pxc
Name:         sample-pxc
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         PerconaXtraDB
Metadata:
  Creation Timestamp:  2022-12-19T09:54:09Z
  Finalizers:
    kubedb.com
  Generation:  4
  ...
  Resource Version:  4309
  UID:               75511bbb-d24f-41a9-9b1c-4bfffd1f5289
Spec:
  Allowed Schemas:
    Namespaces:
      From:  Same
  Auth Secret:
    Name:  sample-pxc-auth
  Auto Ops:
  Coordinator:
    Resources:
  Health Checker:
    Failure Threshold:  1
    Period Seconds:     10
    Timeout Seconds:    10
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Affinity:
        Pod Anti Affinity:
          Preferred During Scheduling Ignored During Execution:
            Pod Affinity Term:
              Label Selector:
                Match Labels:
                  app.kubernetes.io/instance:    sample-pxc
                  app.kubernetes.io/managed-by:  kubedb.com
                  app.kubernetes.io/name:        perconaxtradbs.kubedb.com
              Namespaces:
                demo
              Topology Key:  kubernetes.io/hostname
            Weight:          100
            Pod Affinity Term:
              Label Selector:
                Match Labels:
                  app.kubernetes.io/instance:    sample-pxc
                  app.kubernetes.io/managed-by:  kubedb.com
                  app.kubernetes.io/name:        perconaxtradbs.kubedb.com
              Namespaces:
                demo
              Topology Key:  failure-domain.beta.kubernetes.io/zone
            Weight:          50
      Resources:
        Limits:
          Memory:  1Gi
        Requests:
          Cpu:     500m
          Memory:  1Gi
      Security Context:
        Fs Group:            1001
        Run As Group:        1001
        Run As User:         1001
      Service Account Name:  sample-pxc
  Replicas:                  3
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         1Gi
    Storage Class Name:  standard
  Storage Type:          Durable
  System User Secrets:
    Monitor User Secret:
      Name:  sample-pxc-monitor
    Replication User Secret:
      Name:            sample-pxc-replication
  Termination Policy:  Delete
  Version:             8.0.26
Status:
  Conditions:
    Last Transition Time:  2022-12-19T09:54:09Z
    Message:               The KubeDB operator has started the provisioning of PerconaXtraDB: demo/sample-pxc
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2022-12-19T09:56:53Z
    Message:               All desired replicas are ready.
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2022-12-19T10:00:03Z
    Message:               database sample-pxc/demo is ready
    Observed Generation:   4
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2022-12-19T09:59:13Z
    Message:               database sample-pxc/demo is accepting connection
    Observed Generation:   4
    Reason:                AcceptingConnection
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2022-12-19T10:00:19Z
    Message:               The PerconaXtraDB: demo/sample-pxc is successfully provisioned.
    Observed Generation:   4
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Observed Generation:     4
  Phase:                   Ready
Events:
  Type     Reason        Age    From             Message
  ----     ------        ----   ----             -------
  Normal   PhaseChanged  6m42s  KubeDB Operator  Phase changed from  to Provisioning.
  Normal   Successful    6m42s  KubeDB Operator  Successfully created governing service
  Normal   Successful    6m42s  KubeDB Operator  Successfully created Service
  Normal   Successful    6m32s  KubeDB Operator  Successfully created StatefulSet demo/sample-pxc
  Normal   Successful    6m32s  KubeDB Operator  Successfully created PerconaXtraDB
  Normal   Successful    6m32s  KubeDB Operator  Successfully created appbinding
  Normal   PhaseChanged  51s    KubeDB Operator  Phase changed from NotReady to Provisioning.
  Normal   PhaseChanged  32s    KubeDB Operator  Phase changed from Provisioning to Ready.
  
  
$ kubectl get statefulset -n demo
NAME             READY   AGE
sample-pxc   1/1     27m

$ kubectl get pvc -n demo
NAME                    STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-sample-pxc-0   Bound    pvc-10651900-d975-467f-80ff-9c4755bdf917   1Gi        RWO            standard       27m

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS   REASON   AGE
pvc-10651900-d975-467f-80ff-9c4755bdf917   1Gi        RWO            Delete           Bound    demo/data-sample-pxc-0   standard                27m

$ kubectl get service -n demo
NAME                  TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
sample-pxc        ClusterIP   10.105.207.172   <none>        3306/TCP   28m
sample-pxc-pods   ClusterIP   None             <none>        3306/TCP   28m
```

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created. Run the following command to see PerconaXtraDB object status:

```bash
$ kubectl get perconaxtradb -n demo
NAME         VERSION   STATUS   AGE
sample-pxc   8.0.26    Ready    9m32s
```

## Connect with PerconaXtraDB database

KubeDB operator has created a new Secret called `sample-pxc-auth` for storing the password for `perconaxtradb` superuser. This secret contains a `username` key which contains the *username* for PerconaXtraDB superuser and a `password` key which contains the *password* for PerconaXtraDB superuser.

If you want to use an existing secret please specify that when creating the PerconaXtraDB object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `root` as value of `username`.

Now, we need `username` and `password` to connect to this database from `kubeclt exec` command. In this example, `sample-pxc-auth`  secret holds username and password.

```bash
$ kubectl get secrets -n demo sample-pxc-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo sample-pxc-auth -o jsonpath='{.data.\password}' | base64 -d
w*yOU$b53dTbjsjJ
```

We will exec into the pod `sample-pxc-0` and connet to the database using `username` and `password`.

```bash
$ kubectl exec -it -n demo sample-pxc-0 -- perconaxtradb -u root --password='w*yOU$b53dTbjsjJ'

Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> show databases;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
5 rows in set (0.00 sec)

```

## Database TerminationPolicy

This field is used to regulate the deletion process of the related resources when `PerconaXtraDB` object is deleted. User can set the value of this field according to their needs. The available options and their use case scenario is described below:

**DoNotTerminate:**

When `terminationPolicy` is set to `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.terminationPolicy` is set to `DoNotTerminate`. If you create a database with `terminationPolicy`  `DoNotTerminate` and try to delete it, you will see this:

```bash
$ kubectl delete perconaxtradb sample-pxc -n demo
Error from server (BadRequest): admission webhook "perconaxtradb.validators.kubedb.com" denied the request: perconaxtradb "perconaxtradb-quickstart" can't be halted. To delete, change spec.terminationPolicy
```

Now, run `kubectl edit perconaxtradb sample-pxc -n demo` to set `spec.terminationPolicy` to `Halt` (which deletes the perconaxtradb object and keeps PVC, snapshots, Secrets intact) or remove this field (which default to `Delete`). Then you will be able to delete/halt the database.


**Halt:**

Suppose you want to reuse your database volume and credential to deploy your database in future using the same configurations. But, right now you just want to delete the database except the database volumes and credentials. In this scenario, you must set the `PerconaXtraDB` object `terminationPolicy` to `Halt`.

When the `TerminationPolicy` is set to `Halt` and the PerconaXtraDB object is deleted, the KubeDB operator will delete the StatefulSet and its pods but leaves the `PVCs`, `secrets` and database backup data(`snapshots`) intact. You can set the `terminationPolicy` to `Halt` in existing database using `edit` command for testing.

At first, run `kubectl edit perconaxtradb sample-pxc -n demo` to set `spec.terminationPolicy` to `Halt`. Then delete the perconaxtradb object,

```bash
$ kubectl delete perconaxtradb sample-pxc -n demo
perconaxtradb.kubedb.com "sample-pxc" deleted
```

Now, run the following command to get all perconaxtradb resources in `demo` namespaces,

```bash
$ kubectl get sts,svc,secret,pvc -n demo
NAME                         TYPE                                  DATA   AGE
secret/default-token-w2pgw   kubernetes.io/service-account-token   3      31m
secret/sample-pxc-auth   kubernetes.io/basic-auth              2      39s

NAME                                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-sample-pxc-0   Bound    pvc-7502c222-2b02-4363-9027-91ab0e7b76dc   1Gi        RWO            standard       39s
```

From the above output, you can see that all perconaxtradb resources(`StatefulSet`, `Service`, etc.) are deleted except `PVC` and `Secret`. You can recreate your perconaxtradb again using this resources.

**Delete:**

If you want to delete the existing database along with the volumes used, but want to restore the database from previously taken `snapshots` and `secrets` then you might want to set the `PerconaXtraDB` object `terminationPolicy` to `Delete`. In this setting, `StatefulSet` and the volumes will be deleted. If you decide to restore the database, you can do so using the snapshots and the credentials.

When the `TerminationPolicy` is set to `Delete` and the PerconaXtraDB object is deleted, the KubeDB operator will delete the StatefulSet and its pods along with PVCs but leaves the `secret` and database backup data(`snapshots`) intact.

Suppose, we have a database with `terminationPolicy` set to `Delete`. Now, are going to delete the database using the following command:

```bash
$ kubectl delete perconaxtradb sample-pxc -n demo
perconaxtradb.kubedb.com "sample-pxc" deleted
```

Now, run the following command to get all perconaxtradb resources in `demo` namespaces,

```bash
$ kubectl get sts,svc,secret,pvc -n demo
NAME                          READY   AGE
statefulset.apps/sample-pxc   3/3     3m46s

NAME                      TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/sample-pxc        ClusterIP   10.96.128.19   <none>        3306/TCP   4m5s
service/sample-pxc-pods   ClusterIP   None           <none>        3306/TCP   4m5s

NAME                            TYPE                                  DATA   AGE
secret/default-token-r556j      kubernetes.io/service-account-token   3      20m
secret/sample-pxc-auth          kubernetes.io/basic-auth              2      20m
secret/sample-pxc-monitor       kubernetes.io/basic-auth              2      20m
secret/sample-pxc-replication   kubernetes.io/basic-auth              2      20m
secret/sample-pxc-token-p25ww   kubernetes.io/service-account-token   3      4m5s

NAME                                      STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-sample-pxc-0   Bound    pvc-11f7b634-689e-457e-ba41-157a51090475   1Gi        RWO            standard       3m46s
persistentvolumeclaim/data-sample-pxc-1   Bound    pvc-84dce4b5-35df-4a06-bfea-b0530d83ebb0   1Gi        RWO            standard       3m46s
persistentvolumeclaim/data-sample-pxc-2   Bound    pvc-85a35a7c-dfb8-4ca2-96a6-21c9e0b892db   1Gi        RWO            standard       3m46s
```

From the above output, you can see that all perconaxtradb resources(`StatefulSet`, `Service`, `PVCs` etc.) are deleted except `Secret`.

>If you don't set the terminationPolicy then the kubeDB set the TerminationPolicy to Delete by-default.

**WipeOut:**

You can totally delete the `PerconaXtraDB` database and relevant resources without any tracking by setting `terminationPolicy` to `WipeOut`. KubeDB operator will delete all relevant resources of this `PerconaXtraDB` database (i.e, `PVCs`, `Secrets`, `Snapshots`) when the `terminationPolicy` is set to `WipeOut`.

Suppose, we have a database with `terminationPolicy` set to `WipeOut`. Now, are going to delete the database using the following command:

```yaml
$ kubectl delete perconaxtradb sample-pxc -n demo
perconaxtradb.kubedb.com "sample-pxc" deleted
```

Now, run the following command to get all perconaxtradb resources in `demo` namespaces,

```bash
$ kubectl get sts,svc,secret,pvc -n demo
No resources found in demo namespace.
```

From the above output, you can see that all perconaxtradb resources are deleted. there is no option to recreate/reinitialize your database if `terminationPolicy` is set to `Delete`.

>Be careful when you set the `terminationPolicy` to `Delete`. Because there is no option to trace the database resources if once deleted the database.

## Database Halted

If you want to delete PerconaXtraDB resources(`StatefulSet`,`Service`, etc.) without deleting the `PerconaXtraDB` object, `PVCs` and `Secret` you have to set the `spec.halted` to `true`. KubeDB operator will be able to delete the PerconaXtraDB related resources except the `PerconaXtraDB` object, `PVCs` and `Secret`.

Suppose we have a database running `perconaxtradb-quickstart` in our cluster. Now, we are going to set `spec.halted` to `true` in `PerconaXtraDB`  object by running `kubectl edit -n demo perconaxtradb-quickstart` command.

Run the following command to get PerconaXtraDB resources,

```bash
$ kubectl get perconaxtradb,sts,secret,svc,pvc -n demo
NAME                                VERSION   STATUS   AGE
perconaxtradb.kubedb.com/sample-pxc   8.0.26    Halted   22m

NAME                      TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/sample-pxc        ClusterIP   10.96.128.19   <none>        3306/TCP   4m5s
service/sample-pxc-pods   ClusterIP   None           <none>        3306/TCP   4m5s

NAME                            TYPE                                  DATA   AGE
secret/default-token-r556j      kubernetes.io/service-account-token   3      20m
secret/sample-pxc-auth          kubernetes.io/basic-auth              2      20m
secret/sample-pxc-monitor       kubernetes.io/basic-auth              2      20m
secret/sample-pxc-replication   kubernetes.io/basic-auth              2      20m
secret/sample-pxc-token-p25ww   kubernetes.io/service-account-token   3      4m5s

NAME                                      STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-sample-pxc-0   Bound    pvc-11f7b634-689e-457e-ba41-157a51090475   1Gi        RWO            standard       3m46s
persistentvolumeclaim/data-sample-pxc-1   Bound    pvc-84dce4b5-35df-4a06-bfea-b0530d83ebb0   1Gi        RWO            standard       3m46s
persistentvolumeclaim/data-sample-pxc-2   Bound    pvc-85a35a7c-dfb8-4ca2-96a6-21c9e0b892db   1Gi        RWO            standard       3m46s
```

From the above output , you can see that `PerconaXtraDB` object, `PVCs`, `Secret` are still alive. Then you can recreate your `PerconaXtraDB` with same configuration.

>When you set `spec.halted` to `true` in `PerconaXtraDB` object then the `terminationPolicy` is also set to `Halt` by KubeDB operator.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo perconaxtradb/sample-pxc

kubectl delete ns demo
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for production environment. You can follow these tips to avoid them.

1. **Use `storageType: Ephemeral`**. Databases are precious. You might not want to lose your data in your production environment if database pod fail. So, we recommend to use `spec.storageType: Durable` and provide storage spec in `spec.storage` section. For testing purpose, you can just use `spec.storageType: Ephemeral`. KubeDB will use [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) for storage. You will not require to provide `spec.storage` section.
2. **Use `terminationPolicy: WipeOut`**. It is nice to be able to delete everything created by KubeDB for a particular PerconaXtraDB crd when you delete the crd.

## Next Steps

- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
