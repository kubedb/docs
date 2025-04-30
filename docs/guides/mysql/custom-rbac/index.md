---
title: Run MySQL with Custom RBAC resources
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-custom-rbac
    name: Custom RBAC
    parent: guides-mysql
    weight: 31
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom RBAC resources

KubeDB (version 0.13.0 and higher) supports finer user control over role based access permissions provided to a MySQL instance. This tutorial will show you how to use KubeDB to run MySQL instance with custom RBAC resources.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/guides/mysql/custom-rbac/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/custom-rbac/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows users to provide custom RBAC resources, namely, `ServiceAccount`, `Role`, and `RoleBinding` for MySQL. This is provided via the `spec.podTemplate.spec.serviceAccountName` field in MySQL crd.   If this field is left empty, the KubeDB operator will create a service account name matching MySQL crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually.

This guide will show you how to create custom `Service Account`, `Role`, and `RoleBinding` for a MySQL instance named `quick-postges` to provide the bare minimum access permissions.

## Custom RBAC for MySQL

At first, let's create a `Service Acoount` in `demo` namespace.

```bash
$ kubectl create serviceaccount -n demo my-custom-serviceaccount
serviceaccount/my-custom-serviceaccount created
```

It should create a service account.

```yaml
$ kubectl get serviceaccount -n demo my-custom-serviceaccount -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: "2022-06-28T13:43:26Z"
  name: my-custom-serviceaccount
  namespace: demo
  resourceVersion: "1604181"
  uid: bcc79af3-549e-4037-aece-beffab65a6ef
secrets:
- name: my-custom-serviceaccount-token-bvlb5

```

Now, we need to create a role that has necessary access permissions for the MySQL instance named `quick-mysql`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/custom-rbac/yamls/my-custom-role.yaml
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

```bash
$ kubectl create rolebinding my-custom-rolebinding --role=my-custom-role --serviceaccount=demo:my-custom-serviceaccount --namespace=demo
rolebinding.rbac.authorization.k8s.io/my-custom-rolebinding created

```

It should bind `my-custom-role` and `my-custom-serviceaccount` successfully.

```yaml
$ kubectl get rolebinding -n demo my-custom-rolebinding -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  creationTimestamp: "2022-06-28T13:45:58Z"
  name: my-custom-rolebinding
  namespace: demo
  resourceVersion: "1604463"
  uid: c1242a62-a206-45bf-a757-46e0e20484ca
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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/custom-rbac/yamls/my-custom-db.yaml
mysql.kubedb.com/quick-mysql created
```

Below is the YAML for the MySQL crd we just created.

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: quick-mysql
  namespace: demo
spec:
  version: "9.1.0"
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
  deletionPolicy: DoNotTerminate
```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, PetSet, services, secret etc. If everything goes well, we should see that a pod with the name `quick-mysql-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo quick-mysql-0
NAME            READY   STATUS    RESTARTS   AGE
quick-mysql-0   1/1     Running   0          2m44s
```

Check the pod's log to see if the database is ready

```bash
$ kubectl logs -f -n demo quick-mysql-0
...
2022-06-28 13:46:46+00:00 [Note] [Entrypoint]: Entrypoint script for MySQL Server 9.1.0-1debian10 started.
2022-06-28 13:46:46+00:00 [Note] [Entrypoint]: Switching to dedicated user 'mysql'
2022-06-28 13:46:46+00:00 [Note] [Entrypoint]: Entrypoint script for MySQL Server 9.1.0-1debian10 started.

...
2022-06-28T13:47:02.915445Z 0 [System] [MY-011323] [Server] X Plugin ready for connections. Bind-address: '::' port: 33060, socket: /var/run/mysqld/mysqlx.sock
2022-06-28T13:47:02.915504Z 0 [System] [MY-010931] [Server] /usr/sbin/mysqld: ready for connections. Version: '9.1.0'  socket: '/var/run/mysqld/mysqld.sock'  port: 3306  MySQL Community Server - GPL.

```

Once we see `MySQL init process done. Ready for start up.` in the log, the database is ready.

## Reusing Service Account

An existing service account can be reused in another MySQL instance. No new access permission is required to run the new MySQL instance.

Now, create MySQL crd `minute-mysql` using the existing service account name `my-custom-serviceaccount` in the `spec.podTemplate.spec.serviceAccountName` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/custom-rbac/yamls/my-custom-db-two.yaml
mysql.kubedb.com/quick-mysql created
```

Below is the YAML for the MySQL crd we just created.

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: minute-mysql
  namespace: demo
spec:
  version: "9.1.0"
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
  deletionPolicy: DoNotTerminate

```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, petset, services, secret etc. If everything goes well, we should see that a pod with the name `minute-mysql-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo minute-mysql-0
NAME             READY     STATUS    RESTARTS   AGE
minute-mysql-0   1/1       Running   0          14m
```

Check the pod's log to see if the database is ready

```bash
...
2022-06-28 13:48:53+00:00 [Note] [Entrypoint]: Entrypoint script for MySQL Server 9.1.0-1debian10 started.
2022-06-28 13:48:53+00:00 [Note] [Entrypoint]: Switching to dedicated user 'mysql'
2022-06-28 13:48:53+00:00 [Note] [Entrypoint]: Entrypoint script for MySQL Server 9.1.0-1debian10 started.
2022-06-28 13:48:53+00:00 [Note] [Entrypoint]: Initializing database files
2022-06-28T13:48:53.986191Z 0 [System] [MY-013169] [Server] /usr/sbin/mysqld (mysqld 9.1.0) initializing of server in progress as process 43
...
2022-06-28T13:49:11.543893Z 0 [System] [MY-011323] [Server] X Plugin ready for connections. Bind-address: '::' port: 33060, socket: /var/run/mysqld/mysqlx.sock
2022-06-28T13:49:11.543917Z 0 [System] [MY-010931] [Server] /usr/sbin/mysqld: ready for connections. Version: '9.1.0'  socket: '/var/run/mysqld/mysqld.sock'  port: 3306  MySQL Community Server - GPL.



```

`MySQL init process done. Ready for start up.` in the log signifies that the database is running successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo my/quick-mysql -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo my/quick-mysql

kubectl patch -n demo my/minute-mysql -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo my/minute-mysql

kubectl delete -n demo role my-custom-role
kubectl delete -n demo rolebinding my-custom-rolebinding

kubectl delete sa -n demo my-custom-serviceaccount

kubectl delete ns demo
```

If you would like to uninstall the KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- [Quickstart MySQL](/docs/guides/mysql/quickstart/index.md) with KubeDB Operator.
- Initialize [MySQL with Script](/docs/guides/mysql/initialization/index.md).
- Monitor your MySQL database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mysql/monitoring/prometheus-operator/index.md).
- Monitor your MySQL database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mysql/monitoring/builtin-prometheus/index.md).
- Use [private Docker registry](/docs/guides/mysql/private-registry/index.md) to deploy MySQL with KubeDB.
- Use [kubedb cli](/docs/guides/mysql/cli/index.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
