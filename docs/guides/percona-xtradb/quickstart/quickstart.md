---
title: PerconaXtraDB Quickstart Guide
menu:
  docs_{{ .version }}:
    identifier: px-quickstart-guide
    name: PerconaXtraDB Quickstart Guide
    parent: px-quickstart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Percona XtraDB QuickStart

This tutorial will show you how to use KubeDB to run a PerconaXtraDB database.

<p align="center">
  <img alt="lifecycle"  src="/docs/images/percona-xtradb/Lifecycle_of_a_PerconaXtraDB.svg">
</p>

> Note: The yaml files used in this tutorial are stored in [docs/examples/percona-xtradb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/percona-xtradb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) is required to run KubeDB. Check the available StorageClass in cluster.

  ```bash
  $ kubectl get storageclasses
  NAME                 PROVISIONER                AGE
  standard (default)   k8s.io/minikube-hostpath   4h
  ```

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Find Available PerconaXtraDBVersion

When you have installed KubeDB, it has created `PerconaXtraDBVersion` objects for all supported PerconaXtraDB versions. Check,

```bash
$ kubectl get pxversion
NAME          VERSION   DB_IMAGE                            DEPRECATED   AGE
5.7           5.7       kubedb/percona:5.7                               14m
8.0.26   5.7       kubedb/percona-xtradb-cluster:5.7                14m
```

## Create a PerconaXtraDB database

KubeDB implements a `PerconaXtraDB` CRD to define the specification of a PerconaXtraDB database. Below is the `PerconaXtraDB` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: demo-quickstart
  namespace: demo
spec:
  version: "8.0.26"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
  terminationPolicy: Delete
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/percona-xtradb/quickstart.yaml
perconaxtradb.kubedb.com/demo-quickstart created
```

Here,

- `.spec.version` is the name of the PerconaXtraDBVersion object where the docker images are specified. In this tutorial, a PerconaXtraDB of version 5.7 is going to be created.
- `.spec.storageType` specifies the type of storage that will be used for PerconaXtraDB database. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create PerconaXtraDB database using `EmptyDir` volume. In this case, you don't have to specify `.spec.storage` field. This is useful for testing purposes.
- `.spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `.spec.terminationPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `PerconaXtraDB` object or which resources KubeDB should keep or delete when you delete `PerconaXtraDB` object. If admission webhook is enabled, It prevents users from deleting the database as long as the `.spec.terminationPolicy` is set to `DoNotTerminate`. Learn details of all `TerminationPolicy` [here](/docs/guides/percona-xtradb/concepts/percona-xtradb.md#specterminationpolicy)

> Note: `.spec.storage` section is used to create PVC for database Pod. It will create PVC with storage size specified in `.spec.storage.resources.requests` field. Don't specify limits here. PVC does not get resized automatically.

KubeDB operator watches for `PerconaXtraDB` objects using Kubernetes api. When a `PerconaXtraDB` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching PerconaXtraDB object name. KubeDB operator will also create a governing service for StatefulSets with the name `<percona-xtradb-object-name>-gvr`, if one is not already present.

```bash
$ kubectl dba describe px -n demo demo-quickstart
Name:         demo-quickstart
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         PerconaXtraDB
Metadata:
  Creation Timestamp:  2019-12-23T13:00:34Z
  Finalizers:
    kubedb.com
  Generation:        2
  Resource Version:  45676
  Self Link:         /apis/kubedb.com/v1alpha2/namespaces/demo/perconaxtradbs/demo-quickstart
  UID:               ef82922b-ce02-4184-8f3d-237a37dfec43
Spec:
  Database Secret:
    Secret Name:  demo-quickstart-auth
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Readiness Probe:
        Exec:
          Command:
            bash
            -c
            export MYSQL_PWD="${MYSQL_ROOT_PASSWORD}"
ping_resp=$(mysqladmin -uroot ping)
if [[ "$ping_resp" != "mysqld is alive" ]]; then
    echo "[ERROR] server is not ready. PING_RESPONSE: $ping_resp"
    exit 1
fi

        Initial Delay Seconds:  30
        Period Seconds:         10
      Resources:
  Replicas:  1
  Service Template:
    Metadata:
    Spec:
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         50Mi
    Storage Class Name:  standard
  Storage Type:          Durable
  Termination Policy:    WipeOut
  Version:  5.7
Status:
  Observed Generation:  2
  Phase:                Running
Events:
  Type    Reason      Age   From                    Message
  ----    ------      ----  ----                    -------
  Normal  Successful  17m   PerconaXtraDB operator  Successfully created Service
  Normal  Successful  15m   PerconaXtraDB operator  Successfully created StatefulSet demo/demo-quickstart
  Normal  Successful  15m   PerconaXtraDB operator  Successfully created PerconaXtraDB
  Normal  Successful  15m   PerconaXtraDB operator  Successfully created appbinding
  Normal  Successful  15m   PerconaXtraDB operator  Successfully patched StatefulSet demo/demo-quickstart
  Normal  Successful  15m   PerconaXtraDB operator  Successfully patched PerconaXtraDB

$ kubectl get statefulset -n demo
NAME              READY   AGE
demo-quickstart   1/1     18m

$ kubectl get pvc -n demo
NAME                     STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-demo-quickstart-0   Bound    pvc-b1a12fb2-cc4f-45d1-9d3c-921181d0d8cc   50Mi       RWO            standard       23m

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                         STORAGECLASS   REASON   AGE
pvc-b1a12fb2-cc4f-45d1-9d3c-921181d0d8cc   50Mi       RWO            Delete           Bound    demo/data-demo-quickstart-0   standard                23m

$ kubectl get service -n demo
NAME                  TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
demo-quickstart       ClusterIP   10.96.88.71   <none>        3306/TCP   19m
demo-quickstart-gvr   ClusterIP   None          <none>        3306/TCP   19m
```

KubeDB operator sets the `.status.phase` to `"Running"` once the database is successfully created. Run the following command to see the modified `PerconaXtraDB` object:

```bash
$ kubectl get px -n demo demo-quickstart -o yaml
```

And the output is as follows:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  creationTimestamp: "2019-12-23T13:00:34Z"
  finalizers:
  - kubedb.com
  generation: 2
  name: demo-quickstart
  namespace: demo
  resourceVersion: "45676"
  selfLink: /apis/kubedb.com/v1alpha2/namespaces/demo/perconaxtradbs/demo-quickstart
  uid: ef82922b-ce02-4184-8f3d-237a37dfec43
spec:
  authSecret:
    name: demo-quickstart-auth
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      readinessProbe:
        exec:
          command:
          - bash
          - -c
          - |
            export MYSQL_PWD="${MYSQL_ROOT_PASSWORD}"
            ping_resp=$(mysqladmin -uroot ping)
            if [[ "$ping_resp" != "mysqld is alive" ]]; then
                echo "[ERROR] server is not ready. PING_RESPONSE: $ping_resp"
                exit 1
            fi
        initialDelaySeconds: 30
        periodSeconds: 10
      resources: {}
  replicas: 1
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
    storageClassName: standard
  storageType: Durable
  terminationPolicy: Delete
  version: "8.0.26"
status:
  observedGeneration: 2
  phase: Running
```

## Connect with PerconaXtraDB database

KubeDB operator has created a new Secret called `demo-quickstart-auth` *(format: {percona-xtradb-object-name}-auth)* for storing the password for `mysql` superuser. This secret contains a `username` key which contains the **"username"** for `mysql` superuser and a `password` key which contains the **"password"** for the superuser.

If you want to use an existing secret please specify that when creating the PerconaXtraDB object using `.spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys (`username` and `password`) in `.data` section and also make sure of using `root` as value of `username` key. For more details see [here](/docs/guides/percona-xtradb/concepts/percona-xtradb.md#specdatabasesecret).

Now, you can connect to this database using the database pod IP and and `root` user password.

```bash
$ kubectl get pods demo-quickstart-0 -n demo -o yaml | grep "podIP"
  podIP: 10.244.2.6

$ kubectl get secrets -n demo demo-quickstart-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo demo-quickstart-auth -o jsonpath='{.data.\password}' | base64 -d
y9vpf8LSa8SLqiYC
```

You can connect to the database Pod `demo-quickstart-0`. In that case you just need to specify the host name of the corresponding Pod (either PodIP or the fully-qualified-domain-name for that Pod using the governing service named <percona-xtradb-object-name>-gvr) by --host flag.

```bash
# connect to the server
$ kubectl exec -it -n demo demo-quickstart-0 -- mysql -u root --password=y9vpf8LSa8SLqiYC --host=demo-quickstart-0.demo-quickstart-gvr.demo -e "select 1;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---+
| 1 |
+---+
| 1 |
+---+

# create a database
$ kubectl exec -it -n demo demo-quickstart-0 -- mysql -u root --password=y9vpf8LSa8SLqiYC --host=demo-quickstart-0.demo-quickstart-gvr.demo -e "CREATE DATABASE playground;"
mysql: [Warning] Using a password on the command line interface can be insecure.

# create a table
$ kubectl exec -it -n demo demo-quickstart-0 -- mysql -u root --password=y9vpf8LSa8SLqiYC --host=demo-quickstart-0.demo-quickstart-gvr.demo -e "CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));"
mysql: [Warning] Using a password on the command line interface can be insecure.

# insert a row
$ kubectl exec -it -n demo demo-quickstart-0 -- mysql -u root --password=y9vpf8LSa8SLqiYC --host=demo-quickstart-0.demo-quickstart-gvr.demo -e "INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 2, 'blue');"
mysql: [Warning] Using a password on the command line interface can be insecure.

# read
$ kubectl exec -it -n demo demo-quickstart-0 -- mysql -u root --password=y9vpf8LSa8SLqiYC --host=demo-quickstart-0.demo-quickstart-gvr.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  3 | slide |     2 | blue  |
+----+-------+-------+-------+
```

## DoNotTerminate Property

When `terminationPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidatingWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, it prevents users from deleting the database as long as the `.spec.terminationPolicy` is set to `DoNotTerminate`. You can see this below:

```bash
$ kubectl delete px demo-quickstart -n demo
Error from server (BadRequest): admission webhook "perconaxtradb.validators.kubedb.com" denied the request: percona-xtradb "demo/demo-quickstart" can't be halted. To delete, change spec.terminationPolicy
```

Now, run `$ kubectl edit px demo-quickstart -n demo` to set `.spec.terminationPolicy` to `Halt` (for which KubeDB operator creates `dormantdatabase` when the PerconaXtraDB object is deleted and it keeps PVC, Secrets intact) or remove this field (which default to `Halt`). Then you will be able to delete/halt the database.

Learn details of all `TerminationPolicy` [here](/docs/guides/percona-xtradb/concepts/percona-xtradb.md#specterminationpolicy).

## Halt Database

When [TerminationPolicy](/docs/guides/percona-xtradb/concepts/percona-xtradb.md#specterminationpolicy) is set to `Halt`, it will halt the PerconaXtraDB database instead of deleting it. Here, If you delete the PerconaXtraDB object, KubeDB operator will delete the StatefulSet and its Pods but leaves the PVCs and Secret unchanged. In KubeDB parlance, we say that `demo-quickstart` PerconaXtraDB database has entered into the dormant state. This is represented by KubeDB operator by creating a matching DormantDatabase object.

```bash
$ kubectl delete px demo-quickstart -n demo
perconaxtradb.kubedb.com "demo-quickstart" deleted

$ kubectl get drmn -n demo demo-quickstart
NAME              STATUS   AGE
demo-quickstart   Halted   4m21s

$ kubectl get secret -n demo
NAME                   TYPE                                  DATA   AGE
default-token-9m5pd    kubernetes.io/service-account-token   3      50m
demo-quickstart-auth   Opaque                                2      50m

$ kubectl get pvc -n demo
NAME                     STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-demo-quickstart-0   Bound    pvc-b1a12fb2-cc4f-45d1-9d3c-921181d0d8cc   50Mi       RWO            standard       51m

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                         STORAGECLASS   REASON   AGE
pvc-b1a12fb2-cc4f-45d1-9d3c-921181d0d8cc   50Mi       RWO            Delete           Bound    demo/data-demo-quickstart-0   standard                50m
```

```yaml
$ kubectl get drmn -n demo demo-quickstart -o yaml
apiVersion: kubedb.com/v1alpha2
kind: DormantDatabase
metadata:
  creationTimestamp: "2019-12-23T13:40:51Z"
  finalizers:
  - kubedb.com
  generation: 1
  labels:
    app.kubernetes.io/name: perconaxtradbs.kubedb.com
  name: demo-quickstart
  namespace: demo
  resourceVersion: "50190"
  selfLink: /apis/kubedb.com/v1alpha2/namespaces/demo/dormantdatabases/demo-quickstart
  uid: 13be40cb-2abd-43d2-bffa-42a7ea7a3f5e
spec:
  origin:
    metadata:
      creationTimestamp: "2019-12-23T13:00:34Z"
      name: demo-quickstart
      namespace: demo
    spec:
      perconaxtradb:
        authSecret:
          name: demo-quickstart-auth
        podTemplate:
          controller: {}
          metadata: {}
          spec:
            readinessProbe:
              exec:
                command:
                - bash
                - -c
                - |
                  export MYSQL_PWD="${MYSQL_ROOT_PASSWORD}"
                  ping_resp=$(mysqladmin -uroot ping)
                  if [[ "$ping_resp" != "mysqld is alive" ]]; then
                      echo "[ERROR] server is not ready. PING_RESPONSE: $ping_resp"
                      exit 1
                  fi
              initialDelaySeconds: 30
              periodSeconds: 10
            resources: {}
        replicas: 1
        storage:
          accessModes:
          - ReadWriteOnce
          resources:
            requests:
              storage: 50Mi
          storageClassName: standard
        storageType: Durable
        terminationPolicy: Halt
        version: "8.0.26"
status:
  observedGeneration: 1
  pausingTime: "2019-12-23T13:40:55Z"
  phase: Halted
```

Here,

- `.spec.origin` is the spec of the original spec of the original PerconaXtraDB object.
- `.status.phase` points to the current database state `Halted`.

## Resume Dormant Database

To resume the database from the dormant state, create same `PerconaXtraDB` object with same Spec.

In this tutorial, the dormant database can be resumed by creating original `PerconaXtraDB` object.

The below command will resume the DormantDatabase `demo-quickstart` that was created before.

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/percona-xtradb/quickstart.yaml
perconaxtradb.kubedb.com/demo-quickstart created
```

Now, if you exec into the database, you can see that the data are intact.

## WipeOut DormantDatabase

You can wipe out a DormantDatabase while deleting the objet by setting `.spec.wipeOut` to true. KubeDB operator will delete any relevant resources of this `PerconaXtraDB` database (i.e, PVCs, Secrets, etc.).

```yaml
$ kubectl delete px demo-quickstart -n demo
kubectl delete px demo-quickstart -n demo

$ kubectl edit drmn -n demo demo-quickstart
apiVersion: kubedb.com/v1alpha2
kind: DormantDatabase
metadata:
  name: demo-quickstart
  namespace: demo
  ...
spec:
  wipeOut: true
  ...
status:
  phase: Halted
  ...
```

If `.spec.wipeOut` is not set to true while deleting the `DormantDatabase` object, then only this object will be deleted and KubeDB operator won't delete related Secrets, PVCs.

## Delete DormantDatabase

As it is already discussed above, `DormantDatabase` can be deleted with or without wiping out the resources. To delete the `DormantDatabase`,

```bash
$ kubectl delete drmn demo-quickstart -n demo
dormantdatabase.kubedb.com "demo-quickstart" deleted
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo px/demo-quickstart -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete -n demo px/demo-quickstart

$ kubectl patch -n demo drmn/demo-quickstart -p '{"spec":{"wipeOut":true}}' --type="merge"
$ kubectl delete -n demo drmn/demo-quickstart

$ kubectl delete ns demo
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for production environment. You can follow these tips to avoid them.

1. **Use `storageType: Ephemeral`**. Databases are precious. You might not want to lose your data in your production environment if database Pod fail. So, we recommend to use `.spec.storageType: Durable` and provide storage spec in `.spec.storage` section. For testing purpose, you can just use `.spec.storageType: Ephemeral`. KubeDB will use [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) for storage. You will not require to provide `.spec.storage` section.
2. **Use `terminationPolicy: WipeOut`**. It is nice to be able to resume database from previous one. So, we create `DormantDatabase` and preserve all your `PVCs`, `Secrets`, etc. If you don't want to resume database, you can just use `.spec.terminationPolicy: WipeOut`. It will not create `DormantDatabase` and it will delete everything created by KubeDB for a particular PerconaXtraDB object when you delete the object. For more details about termination policy, please visit [here](/docs/guides/percona-xtradb/concepts/percona-xtradb.md#specterminationpolicy).

## Next Steps

- How to run [PerconaXtraDB Cluster](/docs/guides/percona-xtradb/clustering/percona-xtradb-cluster.md).
- Initialize [PerconaXtraDB with Script](/docs/guides/percona-xtradb/initialization/using-script.md).
- Monitor your PerconaXtraDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/percona-xtradb/monitoring/using-prometheus-operator.md).
- Monitor your PerconaXtraDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/percona-xtradb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/percona-xtradb/private-registry/using-private-registry.md) to deploy PerconaXtraDB with KubeDB.
- How to use [custom configuration](/docs/guides/percona-xtradb/configuration/using-config-file.md).
- How to use [custom rbac resource](/docs/guides/percona-xtradb/custom-rbac/using-custom-rbac.md) for PerconaXtraDB.
- Use Stash to [Backup PerconaXtraDB](/docs/guides/percona-xtradb/backup/overview/index.md).
- Detail concepts of [PerconaXtraDB object](/docs/guides/percona-xtradb/concepts/percona-xtradb.md).
- Detail concepts of [PerconaXtraDBVersion object](/docs/guides/percona-xtradb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
