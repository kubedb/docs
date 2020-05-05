---
title: Run MySQL with Custom RBAC resources
menu:
  docs_{{ .version }}:
    identifier: my-custom-rbac-quickstart
    name: Custom RBAC
    parent: my-custom-rbac
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Using Custom RBAC resources

KubeDB (version 0.13.0 and higher) supports finer user control over role based access permissions provided to a MySQL instance. This tutorial will show you how to use KubeDB to run MySQL instance with custom RBAC resources.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/mysql](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mysql) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows users to provide custom RBAC resources, namely, `ServiceAccount`, `Role`, and `RoleBinding` for MySQL. This is provided via the `spec.podTemplate.spec.serviceAccountName` field in MySQL crd.   If this field is left empty, the KubeDB operator will create a service account name matching MySQL crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually.

This guide will show you how to create custom `Service Account`, `Role`, and `RoleBinding` for a MySQL instance named `quick-postges` to provide the bare minimum access permissions.

## Custom RBAC for MySQL

At first, let's create a `Service Acoount` in `demo` namespace.

```console
$ kubectl create serviceaccount -n demo my-custom-serviceaccount
serviceaccount/my-custom-serviceaccount created
```

It should create a service account.

```yaml
$ kubectl get serviceaccount -n demo my-custom-serviceaccount -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: "2019-05-30T04:23:39Z"
  name: my-custom-serviceaccount
  namespace: demo
  resourceVersion: "21657"
  selfLink: /api/v1/namespaces/demo/serviceaccounts/myserviceaccount
  uid: b2ec2b05-8292-11e9-8d10-080027a8b217
secrets:
- name: myserviceaccount-token-t8zxd
```

Now, we need to create a role that has necessary access permissions for the MySQL instance named `quick-mysql`.

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mysql/custom-rbac/my-custom-role.yaml
role.rbac.authorization.k8s.io/my-custom-role created
```

Below is the YAML for the Role we just created.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: my-custom-role
  namespace: demo
rules:
- apiGroups:
  - policy
  resourceNames:
  - mysql-db
  resources:
  - podsecuritypolicies
  verbs:
  - use
```

This permission is required for MySQL pods running on PSP enabled clusters.

Now create a `RoleBinding` to bind this `Role` with the already created service account.

```console
$ kubectl create rolebinding my-custom-rolebinding --role=my-custom-role --serviceaccount=demo:my-custom-serviceaccount --namespace=demo
rolebinding.rbac.authorization.k8s.io/my-custom-rolebinding created

```

It should bind `my-custom-role` and `my-custom-serviceaccount` successfully.

```yaml
$ kubectl get rolebinding -n demo my-custom-rolebinding -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  creationTimestamp: "kubectl get rolebinding -n demo my-custom-rolebinding -o yaml"
  name: my-custom-rolebinding
  namespace: demo
  resourceVersion: "1405"
  selfLink: /apis/rbac.authorization.k8s.io/v1/namespaces/demo/rolebindings/my-custom-rolebinding
  uid: 123afc02-8297-11e9-8d10-080027a8b217
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: my-custom-role
subjects:
- kind: ServiceAccount
  name: my-custom-serviceaccount
  namespace: demo

```

Now, create a MySQL crd specifying `spec.podTemplate.spec.serviceAccountName` field to `my-custom-serviceaccount`.

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mysql/custom-rbac/my-custom-db.yaml
mysql.kubedb.com/quick-mysql created
```

Below is the YAML for the MySQL crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: quick-mysql
  namespace: demo
spec:
  version: "8.0-v2"
  storageType: Durable
  podTemplate:
    spec:
      serviceAccountName: my-custom-serviceaccount
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: DoNotTerminate

```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, statefulset, services, secret etc. If everything goes well, we should see that a pod with the name `quick-mysql-0` has been created.

Check that the statefulset's pod is running

```console
$ kubectl get pod -n demo quick-mysql-0
NAME                READY     STATUS    RESTARTS   AGE
quick-mysql-0   1/1       Running   0          14m
```

Check the pod's log to see if the database is ready

```console
$ kubectl logs -f -n demo quick-mysql-0
Initializing database
2019-05-31T05:02:35.307699Z 0 [Warning] [MY-011070] [Server] 'Disabling symbolic links using --skip-symbolic-links (or equivalent) is the default. Consider not using this option as it' is deprecated and will be removed in a future release.
2019-05-31T05:02:35.307762Z 0 [System] [MY-013169] [Server] /usr/sbin/mysqld (mysqld 8.0.14) initializing of server in progress as process 29
2019-05-31T05:02:47.346326Z 5 [Warning] [MY-010453] [Server] root@localhost is created with an empty password ! Please consider switching off the --initialize-insecure option.
2019-05-31T05:02:53.777918Z 0 [System] [MY-013170] [Server] /usr/sbin/mysqld (mysqld 8.0.14) initializing of server has completed
Database initialized
MySQL init process in progress...
MySQL init process in progress...
2019-05-31T05:02:56.656884Z 0 [Warning] [MY-011070] [Server] 'Disabling symbolic links using --skip-symbolic-links (or equivalent) is the default. Consider not using this option as it' is deprecated and will be removed in a future release.
2019-05-31T05:02:56.656953Z 0 [System] [MY-010116] [Server] /usr/sbin/mysqld (mysqld 8.0.14) starting as process 80
2019-05-31T05:02:57.876853Z 0 [Warning] [MY-010068] [Server] CA certificate ca.pem is self signed.
2019-05-31T05:02:57.892774Z 0 [Warning] [MY-011810] [Server] Insecure configuration for --pid-file: Location '/var/run/mysqld' in the path is accessible to all OS users. Consider choosing a different directory.
2019-05-31T05:02:57.910391Z 0 [System] [MY-010931] [Server] /usr/sbin/mysqld: ready for connections. Version: '8.0.14'  socket: '/var/run/mysqld/mysqld.sock'  port: 0  MySQL Community Server - GPL.
2019-05-31T05:02:58.045050Z 0 [System] [MY-011323] [Server] X Plugin ready for connections. Socket: '/var/run/mysqld/mysqlx.sock'
Warning: Unable to load '/usr/share/zoneinfo/iso3166.tab' as time zone. Skipping it.
Warning: Unable to load '/usr/share/zoneinfo/leap-seconds.list' as time zone. Skipping it.
Warning: Unable to load '/usr/share/zoneinfo/zone.tab' as time zone. Skipping it.
Warning: Unable to load '/usr/share/zoneinfo/zone1970.tab' as time zone. Skipping it.

2019-05-31T05:03:04.217396Z 0 [System] [MY-010910] [Server] /usr/sbin/mysqld: Shutdown complete (mysqld 8.0.14)  MySQL Community Server - GPL.

MySQL init process done. Ready for start up.
```

Once we see `MySQL init process done. Ready for start up.` in the log, the database is ready.

## Reusing Service Account

An existing service account can be reused in another MySQL instance. No new access permission is required to run the new MySQL instance.

Now, create MySQL crd `minute-mysql` using the existing service account name `my-custom-serviceaccount` in the `spec.podTemplate.spec.serviceAccountName` field.

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mysql/custom-rbac/my-custom-db-two.yaml
mysql.kubedb.com/quick-mysql created
```

Below is the YAML for the MySQL crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: minute-mysql
  namespace: demo
spec:
  version: "8.0-v2"
  storageType: Durable
  podTemplate:
    spec:
      serviceAccountName: my-custom-serviceaccount
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: DoNotTerminate

```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, statefulset, services, secret etc. If everything goes well, we should see that a pod with the name `minute-mysql-0` has been created.

Check that the statefulset's pod is running

```console
$ kubectl get pod -n demo minute-mysql-0
NAME                READY     STATUS    RESTARTS   AGE
minute-mysql-0   1/1       Running   0          14m
```

Check the pod's log to see if the database is ready

```console
$ kubectl logs -f -n demo minute-mysql-0
Initializing database
2019-05-31T05:09:12.165236Z 0 [Warning] [MY-011070] [Server] 'Disabling symbolic links using --skip-symbolic-links (or equivalent) is the default. Consider not using this option as it' is deprecated and will be removed in a future release.
2019-05-31T05:09:12.165298Z 0 [System] [MY-013169] [Server] /usr/sbin/mysqld (mysqld 8.0.14) initializing of server in progress as process 28
2019-05-31T05:09:24.903995Z 5 [Warning] [MY-010453] [Server] root@localhost is created with an empty password ! Please consider switching off the --initialize-insecure option.
2019-05-31T05:09:30.857155Z 0 [System] [MY-013170] [Server] /usr/sbin/mysqld (mysqld 8.0.14) initializing of server has completed
Database initialized
MySQL init process in progress...
MySQL init process in progress...
2019-05-31T05:09:33.931254Z 0 [Warning] [MY-011070] [Server] 'Disabling symbolic links using --skip-symbolic-links (or equivalent) is the default. Consider not using this option as it' is deprecated and will be removed in a future release.
2019-05-31T05:09:33.931315Z 0 [System] [MY-010116] [Server] /usr/sbin/mysqld (mysqld 8.0.14) starting as process 79
2019-05-31T05:09:34.819349Z 0 [Warning] [MY-010068] [Server] CA certificate ca.pem is self signed.
2019-05-31T05:09:34.834673Z 0 [Warning] [MY-011810] [Server] Insecure configuration for --pid-file: Location '/var/run/mysqld' in the path is accessible to all OS users. Consider choosing a different directory.
2019-05-31T05:09:34.850188Z 0 [System] [MY-010931] [Server] /usr/sbin/mysqld: ready for connections. Version: '8.0.14'  socket: '/var/run/mysqld/mysqld.sock'  port: 0  MySQL Community Server - GPL.
2019-05-31T05:09:35.064435Z 0 [System] [MY-011323] [Server] X Plugin ready for connections. Socket: '/var/run/mysqld/mysqlx.sock'
Warning: Unable to load '/usr/share/zoneinfo/iso3166.tab' as time zone. Skipping it.
Warning: Unable to load '/usr/share/zoneinfo/leap-seconds.list' as time zone. Skipping it.
Warning: Unable to load '/usr/share/zoneinfo/zone.tab' as time zone. Skipping it.
Warning: Unable to load '/usr/share/zoneinfo/zone1970.tab' as time zone. Skipping it.

2019-05-31T05:09:41.236940Z 0 [System] [MY-010910] [Server] /usr/sbin/mysqld: Shutdown complete (mysqld 8.0.14)  MySQL Community Server - GPL.

MySQL init process done. Ready for start up.

```

`MySQL init process done. Ready for start up.` in the log signifies that the database is running successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
kubectl patch -n demo my/quick-mysql -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo my/quick-mysql

kubectl patch -n demo my/minute-mysql -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo my/minute-mysql

kubectl delete -n demo role my-custom-role
kubectl delete -n demo rolebinding my-custom-rolebinding

kubectl delete sa -n demo my-custom-serviceaccount

kubectl delete ns demo
```

If you would like to uninstall the KubeDB operator, please follow the steps [here](/docs/setup/operator/uninstall.md).

## Next Steps

- [Quickstart MySQL](/docs/guides/mysql/quickstart/quickstart.md) with KubeDB Operator.
- [Snapshot and Restore](/docs/guides/mysql/snapshot/backup-and-restore.md) process of MySQL instances using KubeDB.
- Take [Scheduled Snapshot](/docs/guides/mysql/snapshot/scheduled-backup.md) of MySQL instances using KubeDB.
- Initialize [MySQL with Script](/docs/guides/mysql/initialization/using-script.md).
- Initialize [MySQL with Snapshot](/docs/guides/mysql/initialization/using-snapshot.md).
- Monitor your MySQL instance with KubeDB using [out-of-the-box CoreOS Prometheus Operator](/docs/guides/mysql/monitoring/using-coreos-prometheus-operator.md).
- Monitor your MySQL instance with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mysql/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mysql/private-registry/using-private-registry.md) to deploy MySQL with KubeDB.
- Use [kubedb cli](/docs/guides/mysql/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [MySQL object](/docs/concepts/databases/mysql.md).
- Detail concepts of [Snapshot object](/docs/concepts/snapshot.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

