---
title: Run PerconaXtraDB with Custom RBAC resources
menu:
  docs_{{ .version }}:
    identifier: px-custom-rbac-quickstart
    name: Custom RBAC
    parent: px-custom-rbac
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Using Custom RBAC resources

KubeDB (version 0.13.0 and higher) supports finer user control over role based access permissions provided to a PerconaXtraDB instance. This tutorial will show you how to use KubeDB to run PerconaXtraDB instance with custom RBAC resources.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/install.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/percona-xtradb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/percona-xtradb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows users to provide custom RBAC resources, namely, `ServiceAccount`, `Role`, and `RoleBinding` for PerconaXtraDB. This is provided via the `.spec.podTemplate.spec.serviceAccountName` field in PerconaXtraDB object. If this field is left empty, the KubeDB operator will create a service account name matching PerconaXtraDB object name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually.

This guide will show you how to create custom `Service Account`, `Role`, and `RoleBinding` for a PerconaXtraDB instance named `px-custom-rbac` to provide the bare minimum access permissions.

## Custom RBAC for PerconaXtraDB

At first, let's create a `Service Account` in `demo` namespace.

```console
$ kubectl create serviceaccount -n demo px-custom-serviceaccount
serviceaccount/px-custom-serviceaccount created
```

It should create a service account.

```console
$ kubectl get serviceaccount -n demo px-custom-serviceaccount -o yaml
```

Output:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: "2020-01-09T09:05:05Z"
  name: px-custom-serviceaccount
  namespace: demo
  resourceVersion: "30521"
  selfLink: /api/v1/namespaces/demo/serviceaccounts/px-custom-serviceaccount
  uid: 06ff3050-bd49-41eb-9be5-ffdd69b174e7
secrets:
- name: px-custom-serviceaccount-token-wt4pm
```

Now, we need to create a role that has necessary access permissions for the PerconaXtraDB instance named `px-custom-rbac`.

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/percona-xtradb/px-custom-role.yaml
role.rbac.authorization.k8s.io/px-custom-role created
```

Below is the YAML for the Role we just created.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: px-custom-role
  namespace: demo
rules:
  - apiGroups:
      - policy
    resourceNames:
      - percona-xtradb-db
    resources:
      - podsecuritypolicies
    verbs:
      - use
```

This permission is required for PerconaXtraDB Pods running on PSP enabled clusters.

Now create a `RoleBinding` to bind this `Role` with the already created service account.

```console
$ kubectl create rolebinding px-custom-rolebinding --role=px-custom-role --serviceaccount=demo:px-custom-serviceaccount --namespace=demo
rolebinding.rbac.authorization.k8s.io/px-custom-rolebinding created
```

It should bind `px-custom-role` and `px-custom-serviceaccount` successfully.

```console
$ kubectl get rolebinding -n demo px-custom-rolebinding -o yaml
```

Output:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  creationTimestamp: "2020-01-09T10:09:26Z"
  name: px-custom-rolebinding
  namespace: demo
  resourceVersion: "38236"
  selfLink: /apis/rbac.authorization.k8s.io/v1/namespaces/demo/rolebindings/px-custom-rolebinding
  uid: eec78f30-2050-4d91-aff2-18df1779dfd5
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: px-custom-role
subjects:
- kind: ServiceAccount
  name: px-custom-serviceaccount
  namespace: demo
```

Now, create a PerconaXtraDB object specifying `.spec.podTemplate.spec.serviceAccountName` field to `px-custom-serviceaccount`.

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/percona-xtradb/px-custom-rbac.yaml
perconaxtradb.kubedb.com/px-custom-rbac created
```

Below is the YAML for the PerconaXtraDB object we just created.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: PerconaXtraDB
metadata:
  name: px-custom-rbac
  namespace: demo
spec:
  version: "5.7"
  replicas: 1
  storageType: Durable
  podTemplate:
    spec:
      serviceAccountName: px-custom-serviceaccount
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
  terminationPolicy: DoNotTerminate
```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, statefulset, services, secret etc. If everything goes well, we should see that a Pod with the name `px-custom-rbac-0` has been created.

Check that the StatefulSet's Pod is running.

```console
$ kubectl get px -n demo px-custom-rbac
NAME             VERSION   STATUS    AGE
px-custom-rbac   5.7       Running   2m11s

$ kubectl get pod -n demo px-custom-rbac-0
NAME               READY   STATUS    RESTARTS   AGE
px-custom-rbac-0   1/1     Running   0          29m
```

Check the Pod's log to see if the database is ready.

```console
$ kubectl logs -f -n demo px-custom-rbac-0
Initializing database
2020-01-09T10:20:17.263910Z 0 [Warning] TIMESTAMP with implicit DEFAULT value is deprecated. Please use --explicit_defaults_for_timestamp server option (see documentation for more details).
2020-01-09T10:20:18.822906Z 0 [Warning] InnoDB: New log files created, LSN=45790
2020-01-09T10:20:19.045411Z 0 [Warning] InnoDB: Creating foreign key constraint system tables.
2020-01-09T10:20:19.098009Z 0 [Warning] No existing UUID has been found, so we assume that this is the first time that this server has been started. Generating a new UUID: a2e567b1-32c9-11ea-aefc-aa9dfb1ff340.
2020-01-09T10:20:19.101483Z 0 [Warning] Gtid table is not ready to be used. Table 'mysql.gtid_executed' cannot be opened.
2020-01-09T10:20:19.230258Z 0 [Warning] CA certificate ca.pem is self signed.
2020-01-09T10:20:19.299803Z 1 [Warning] root@localhost is created with an empty password ! Please consider switching off the --initialize-insecure option.
Database initialized
MySQL init process in progress...
2020-01-09T10:20:25.387944Z 0 [Warning] TIMESTAMP with implicit DEFAULT value is deprecated. Please use --explicit_defaults_for_timestamp server option (see documentation for more details).
...

MySQL init process done. Ready for start up.

2020-01-09T10:20:31.336569Z 0 [Warning] TIMESTAMP with implicit DEFAULT value is deprecated. Please use --explicit_defaults_for_timestamp server option (see documentation for more details).
2020-01-09T10:20:31.337661Z 0 [Note] mysqld (mysqld 5.7.26-29) starting as process 1 ...
2020-01-09T10:20:31.339936Z 0 [Note] InnoDB: PUNCH HOLE support available
2020-01-09T10:20:31.339955Z 0 [Note] InnoDB: Mutexes and rw_locks use GCC atomic builtins
2020-01-09T10:20:31.339959Z 0 [Note] InnoDB: Uses event mutexes
2020-01-09T10:20:31.339963Z 0 [Note] InnoDB: GCC builtin __atomic_thread_fence() is used for memory barrier
2020-01-09T10:20:31.339967Z 0 [Note] InnoDB: Compressed tables use zlib 1.2.7
2020-01-09T10:20:31.339970Z 0 [Note] InnoDB: Using Linux native AIO
2020-01-09T10:20:31.340152Z 0 [Note] InnoDB: Number of pools: 1
2020-01-09T10:20:31.340238Z 0 [Note] InnoDB: Using CPU crc32 instructions
2020-01-09T10:20:31.341474Z 0 [Note] InnoDB: Initializing buffer pool, total size = 128M, instances = 1, chunk size = 128M
2020-01-09T10:20:31.344542Z 0 [Note] InnoDB: Completed initialization of buffer pool
2020-01-09T10:20:31.346110Z 0 [Note] InnoDB: If the mysqld execution user is authorized, page cleaner thread priority can be changed. See the man page of setpriority().
2020-01-09T10:20:31.365087Z 0 [Note] InnoDB: Crash recovery did not find the parallel doublewrite buffer at /var/lib/mysql/xb_doublewrite
2020-01-09T10:20:31.365513Z 0 [Note] InnoDB: Highest supported file format is Barracuda.
2020-01-09T10:20:31.446986Z 0 [Note] InnoDB: Created parallel doublewrite buffer at /var/lib/mysql/xb_doublewrite, size 3932160 bytes
2020-01-09T10:20:31.459225Z 0 [Note] InnoDB: Creating shared tablespace for temporary tables
2020-01-09T10:20:31.459318Z 0 [Note] InnoDB: Setting file './ibtmp1' size to 12 MB. Physically writing the file full; Please wait ...
2020-01-09T10:20:31.665372Z 0 [Note] InnoDB: File './ibtmp1' size is now 12 MB.
2020-01-09T10:20:31.666040Z 0 [Note] InnoDB: 96 redo rollback segment(s) found. 96 redo rollback segment(s) are active.
2020-01-09T10:20:31.666051Z 0 [Note] InnoDB: 32 non-redo rollback segment(s) are active.
2020-01-09T10:20:31.666399Z 0 [Note] InnoDB: Waiting for purge to start
2020-01-09T10:20:31.716560Z 0 [Note] InnoDB: Percona XtraDB (http://www.percona.com) 5.7.26-29 started; log sequence number 11874994
2020-01-09T10:20:31.717099Z 0 [Note] InnoDB: Loading buffer pool(s) from /var/lib/mysql/ib_buffer_pool
2020-01-09T10:20:31.717713Z 0 [Note] Plugin 'FEDERATED' is disabled.
2020-01-09T10:20:31.724077Z 0 [Note] InnoDB: Buffer pool(s) load completed at 200109 10:20:31
2020-01-09T10:20:31.726495Z 0 [Note] Found ca.pem, server-cert.pem and server-key.pem in data directory. Trying to enable SSL support using them.
2020-01-09T10:20:31.726507Z 0 [Note] Skipping generation of SSL certificates as certificate files are present in data directory.
2020-01-09T10:20:31.726976Z 0 [Warning] CA certificate ca.pem is self signed.
2020-01-09T10:20:31.727004Z 0 [Note] Skipping generation of RSA key pair as key files are present in data directory.
2020-01-09T10:20:31.727333Z 0 [Note] Server hostname (bind-address): '*'; port: 3306
2020-01-09T10:20:31.727696Z 0 [Note] IPv6 is available.
2020-01-09T10:20:31.727705Z 0 [Note]   - '::' resolves to '::';
2020-01-09T10:20:31.727719Z 0 [Note] Server socket created on IP: '::'.
2020-01-09T10:20:31.738417Z 0 [Note] Event Scheduler: Loaded 0 events
2020-01-09T10:20:31.738597Z 0 [Note] mysqld: ready for connections.
Version: '5.7.26-29'  socket: '/var/lib/mysql/mysql.sock'  port: 3306  Percona Server (GPL), Release 29, Revision 11ad961
```

Once we see `MySQL init process done. Ready for start up.` in the log, the database is ready.

## Reusing Service Account

An existing service account can be reused in another PerconaXtraDB instance. No new access permission is required to run the new PerconaXtraDB instance.

Now, create another PerconaXtraDB object `minute-mysql` using the existing service account name `px-custom-serviceaccount` in the `.spec.podTemplate.spec.serviceAccountName` field.

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/percona-xtradb/px-custom-rbac-two.yaml
perconaxtradb.kubedb.com/px-custom-rbac-two created
```

Below is the YAML for the PerconaXtraDB object we just created.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: PerconaXtraDB
metadata:
  name: px-custom-rbac-two
  namespace: demo
spec:
  version: "5.7"
  replicas: 1
  storageType: Durable
  podTemplate:
    spec:
      serviceAccountName: px-custom-serviceaccount
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
  terminationPolicy: DoNotTerminate
```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, statefulset, services, secret etc. If everything goes well, we should see that a Pod with the name `px-custom-rbac-two-0` has been created.

Check that the StatefulSet's Pod is running

```console
$ kubectl get px -n demo px-custom-rbac-two
NAME                 VERSION   STATUS    AGE
px-custom-rbac-two   5.7       Running   2m29s

$ kubectl get pod -n demo px-custom-rbac-two-0
NAME                   READY   STATUS    RESTARTS   AGE
px-custom-rbac-two-0   1/1     Running   0          3m7s
```

Check the Pod's log to see if the database is ready

```console
$ kubectl logs -f -n demo px-custom-rbac-two-0
Initializing database
2020-01-09T11:17:53.650798Z 0 [Warning] TIMESTAMP with implicit DEFAULT value is deprecated. Please use --explicit_defaults_for_timestamp server option (see documentation for more details).
2020-01-09T11:17:54.644803Z 0 [Warning] InnoDB: New log files created, LSN=45790
2020-01-09T11:17:54.796314Z 0 [Warning] InnoDB: Creating foreign key constraint system tables.
2020-01-09T11:17:54.830674Z 0 [Warning] No existing UUID has been found, so we assume that this is the first time that this server has been started. Generating a new UUID: aeac5ce3-32d1-11ea-9c28-8641016787d1.
2020-01-09T11:17:54.838176Z 0 [Warning] Gtid table is not ready to be used. Table 'mysql.gtid_executed' cannot be opened.
2020-01-09T11:17:55.032258Z 0 [Warning] CA certificate ca.pem is self signed.
2020-01-09T11:17:55.099944Z 1 [Warning] root@localhost is created with an empty password ! Please consider switching off the --initialize-insecure option.
Database initialized
MySQL init process in progress...
2020-01-09T11:18:00.671870Z 0 [Warning] TIMESTAMP with implicit DEFAULT value is deprecated. Please use --explicit_defaults_for_timestamp server option (see documentation for more details).
...

MySQL init process done. Ready for start up.

2020-01-09T11:18:05.897577Z 0 [Warning] TIMESTAMP with implicit DEFAULT value is deprecated. Please use --explicit_defaults_for_timestamp server option (see documentation for more details).
2020-01-09T11:18:05.898678Z 0 [Note] mysqld (mysqld 5.7.26-29) starting as process 1 ...
2020-01-09T11:18:05.901031Z 0 [Note] InnoDB: PUNCH HOLE support available
2020-01-09T11:18:05.901051Z 0 [Note] InnoDB: Mutexes and rw_locks use GCC atomic builtins
2020-01-09T11:18:05.901055Z 0 [Note] InnoDB: Uses event mutexes
2020-01-09T11:18:05.901058Z 0 [Note] InnoDB: GCC builtin __atomic_thread_fence() is used for memory barrier
2020-01-09T11:18:05.901062Z 0 [Note] InnoDB: Compressed tables use zlib 1.2.7
2020-01-09T11:18:05.901065Z 0 [Note] InnoDB: Using Linux native AIO
2020-01-09T11:18:05.901255Z 0 [Note] InnoDB: Number of pools: 1
2020-01-09T11:18:05.901350Z 0 [Note] InnoDB: Using CPU crc32 instructions
2020-01-09T11:18:05.902544Z 0 [Note] InnoDB: Initializing buffer pool, total size = 128M, instances = 1, chunk size = 128M
2020-01-09T11:18:05.905640Z 0 [Note] InnoDB: Completed initialization of buffer pool
2020-01-09T11:18:05.906981Z 0 [Note] InnoDB: If the mysqld execution user is authorized, page cleaner thread priority can be changed. See the man page of setpriority().
2020-01-09T11:18:05.925827Z 0 [Note] InnoDB: Crash recovery did not find the parallel doublewrite buffer at /var/lib/mysql/xb_doublewrite
2020-01-09T11:18:05.926491Z 0 [Note] InnoDB: Highest supported file format is Barracuda.
2020-01-09T11:18:05.970628Z 0 [Note] InnoDB: Created parallel doublewrite buffer at /var/lib/mysql/xb_doublewrite, size 3932160 bytes
2020-01-09T11:18:05.985063Z 0 [Note] InnoDB: Creating shared tablespace for temporary tables
2020-01-09T11:18:05.985136Z 0 [Note] InnoDB: Setting file './ibtmp1' size to 12 MB. Physically writing the file full; Please wait ...
2020-01-09T11:18:06.079208Z 0 [Note] InnoDB: File './ibtmp1' size is now 12 MB.
2020-01-09T11:18:06.080463Z 0 [Note] InnoDB: 96 redo rollback segment(s) found. 96 redo rollback segment(s) are active.
2020-01-09T11:18:06.080482Z 0 [Note] InnoDB: 32 non-redo rollback segment(s) are active.
2020-01-09T11:18:06.081247Z 0 [Note] InnoDB: Percona XtraDB (http://www.percona.com) 5.7.26-29 started; log sequence number 11875357
2020-01-09T11:18:06.081485Z 0 [Note] InnoDB: Loading buffer pool(s) from /var/lib/mysql/ib_buffer_pool
2020-01-09T11:18:06.081829Z 0 [Note] Plugin 'FEDERATED' is disabled.
2020-01-09T11:18:06.087193Z 0 [Note] InnoDB: Buffer pool(s) load completed at 200109 11:18:06
2020-01-09T11:18:06.089617Z 0 [Note] Found ca.pem, server-cert.pem and server-key.pem in data directory. Trying to enable SSL support using them.
2020-01-09T11:18:06.089632Z 0 [Note] Skipping generation of SSL certificates as certificate files are present in data directory.
2020-01-09T11:18:06.090088Z 0 [Warning] CA certificate ca.pem is self signed.
2020-01-09T11:18:06.090117Z 0 [Note] Skipping generation of RSA key pair as key files are present in data directory.
2020-01-09T11:18:06.090421Z 0 [Note] Server hostname (bind-address): '*'; port: 3306
2020-01-09T11:18:06.090446Z 0 [Note] IPv6 is available.
2020-01-09T11:18:06.090454Z 0 [Note]   - '::' resolves to '::';
2020-01-09T11:18:06.090467Z 0 [Note] Server socket created on IP: '::'.
2020-01-09T11:18:06.110254Z 0 [Note] Event Scheduler: Loaded 0 events
2020-01-09T11:18:06.110715Z 0 [Note] mysqld: ready for connections.
Version: '5.7.26-29'  socket: '/var/lib/mysql/mysql.sock'  port: 3306  Percona Server (GPL), Release 29, Revision 11ad961
```

`MySQL init process done. Ready for start up.` in the log signifies that the database is running successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
kubectl patch -n demo px/px-custom-rbac-two -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo px/px-custom-rbac-two

kubectl patch -n demo px/px-custom-rbac -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo px/px-custom-rbac

kubectl delete -n demo role px-custom-role
kubectl delete -n demo rolebinding px-custom-rolebinding

kubectl delete sa -n demo px-custom-serviceaccount

kubectl delete ns demo
```

If you would like to uninstall the KubeDB operator, please follow the steps [here](/docs/setup/uninstall.md).

## Next Steps

- Initialize [PerconaXtraDB with Script](/docs/guides/percona-xtradb/initialization/using-script.md).
- Monitor your PerconaXtraDB database with KubeDB using [out-of-the-box CoreOS Prometheus Operator](/docs/guides/percona-xtradb/monitoring/using-coreos-prometheus-operator.md).
- Monitor your PerconaXtraDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/percona-xtradb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/percona-xtradb/private-registry/using-private-registry.md) to deploy PerconaXtraDB with KubeDB.
- How to use [custom configuration](/docs/guides/percona-xtradb/configuration/using-custom-config.md).
- Use Stash to [Backup PerconaXtraDB](/docs/guides/percona-xtradb/snapshot/stash.md).
- Detail concepts of [PerconaXtraDB object](/docs/concepts/databases/percona-xtradb.md).
- Detail concepts of [PerconaXtraDBVersion object](/docs/concepts/catalog/percona-xtradb.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
